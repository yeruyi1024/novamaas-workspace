package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	// systemTaskRunnerIdleInterval is the fallback poll interval used to pick up
	// tasks created on other nodes and mark expired leases failed.
	systemTaskRunnerIdleInterval = 15 * time.Second
	systemTaskLockTTL            = 60 * time.Second
	logCleanupBatchSize          = 100

	// systemTaskSchedulerInterval throttles how often the scheduler/stale-lock
	// pass runs, independent of how often the runner wakes to claim tasks.
	systemTaskSchedulerInterval = 15 * time.Second
	systemTaskStaleLockInterval = 30 * time.Second
)

// SystemTaskHandler executes a claimed task of a specific type. Run owns the
// task lifecycle from claim to terminal state: it MUST call
// model.FinishSystemTask (succeeded/failed) before returning and MUST honor
// ctx cancellation, which the runner triggers if the per-type lock is lost.
type SystemTaskHandler interface {
	Type() string
	Run(ctx context.Context, task *model.SystemTask, runnerID string)
}

// ScheduledSystemTaskHandler is a SystemTaskHandler that the scheduler also
// creates periodically when enabled and the configured interval has elapsed
// since the last run.
type ScheduledSystemTaskHandler interface {
	SystemTaskHandler
	Enabled() bool
	Interval() time.Duration
	NewPayload() any
}

var (
	systemTaskHandlersMu sync.RWMutex
	systemTaskHandlers   = map[string]SystemTaskHandler{}
)

// RegisterSystemTaskHandler registers a handler keyed by its Type(). It must be
// called before StartSystemTaskRunner (or any time, since the runner snapshots
// the registry every pass). Re-registering a type replaces the previous handler.
func RegisterSystemTaskHandler(h SystemTaskHandler) {
	if h == nil {
		return
	}
	systemTaskHandlersMu.Lock()
	defer systemTaskHandlersMu.Unlock()
	systemTaskHandlers[h.Type()] = h
}

func registeredSystemTaskHandlers() []SystemTaskHandler {
	systemTaskHandlersMu.RLock()
	defer systemTaskHandlersMu.RUnlock()
	handlers := make([]SystemTaskHandler, 0, len(systemTaskHandlers))
	for _, h := range systemTaskHandlers {
		handlers = append(handlers, h)
	}
	return handlers
}

// logCleanupHandler wraps the existing on-demand log cleanup task as a
// registered (non-scheduled) handler. It is created via StartLogCleanupTask.
type logCleanupHandler struct{}

func (logCleanupHandler) Type() string { return model.SystemTaskTypeLogCleanup }

func (logCleanupHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	runLogCleanupTask(ctx, task, runnerID)
}

func init() {
	RegisterSystemTaskHandler(logCleanupHandler{})
	RegisterSystemTaskHandler(requestBodyArchiveHandler{})
}

type requestBodyArchiveHandler struct{}

func (requestBodyArchiveHandler) Type() string { return model.SystemTaskTypeRequestBodyArchive }

func (requestBodyArchiveHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	runRequestBodyArchiveTask(ctx, task, runnerID)
}

type LogCleanupPayload struct {
	TargetTimestamp int64 `json:"target_timestamp"`
	BatchSize       int   `json:"batch_size"`
}

type LogCleanupState struct {
	Total     int64 `json:"total"`
	Processed int64 `json:"processed"`
	Progress  int   `json:"progress"`
	Remaining int64 `json:"remaining"`
}

type LogCleanupResult struct {
	DeletedCount              int64 `json:"deleted_count"`
	RequestBodiesDeletedCount int64 `json:"request_bodies_deleted_count"`
}

type RequestBodyArchivePayload struct {
	BatchSize int `json:"batch_size"`
}

type RequestBodyArchiveState struct {
	Initialized        bool                              `json:"initialized"`
	Phase              string                            `json:"phase"`
	Total              int64                             `json:"total"`
	Processed          int64                             `json:"processed"`
	Progress           int                               `json:"progress"`
	TaskUpperID        int64                             `json:"task_upper_id"`
	TaskCursor         int64                             `json:"task_cursor"`
	LogUpperCursor     model.LogRequestBodyArchiveCursor `json:"log_upper_cursor"`
	LogCursor          model.LogRequestBodyArchiveCursor `json:"log_cursor"`
	ArchivedCount      int64                             `json:"archived_count"`
	TaskRowsCleaned    int64                             `json:"task_rows_cleaned"`
	LogRowsCleaned     int64                             `json:"log_rows_cleaned"`
	InvalidRowsSkipped int64                             `json:"invalid_rows_skipped"`
}

type RequestBodyArchiveResult struct {
	ScannedRows        int64 `json:"scanned_rows"`
	ArchivedCount      int64 `json:"archived_count"`
	TaskRowsCleaned    int64 `json:"task_rows_cleaned"`
	LogRowsCleaned     int64 `json:"log_rows_cleaned"`
	InvalidRowsSkipped int64 `json:"invalid_rows_skipped"`
}

var (
	systemTaskRunnerOnce sync.Once
	// systemTaskWakeup signals the runner to check for runnable tasks
	// immediately instead of waiting for the idle poll. Buffered so a signal
	// raised while the runner is busy is not lost and is handled on the next loop.
	systemTaskWakeup = make(chan struct{}, 1)
)

// notifySystemTaskRunner wakes the runner without blocking. If a wakeup is
// already pending it is a no-op, which is fine since one pass drains all work.
func notifySystemTaskRunner() {
	select {
	case systemTaskWakeup <- struct{}{}:
	default:
	}
}

func StartSystemTaskRunner() {
	systemTaskRunnerOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}

		runnerID := fmt.Sprintf("%s-%s", common.NodeName, common.GetRandomString(8))
		gopool.Go(func() {
			logger.LogInfo(context.Background(), fmt.Sprintf("system task runner started: runner=%s idle_interval=%s", runnerID, systemTaskRunnerIdleInterval))

			ticker := time.NewTicker(systemTaskRunnerIdleInterval)
			defer ticker.Stop()

			var lastScheduler time.Time
			var lastStaleLockCleanup time.Time
			runPass := func() {
				// The scheduler/stale-lock pass is throttled independently of the
				// claim pass: wakeups (e.g. a manual log cleanup) should claim
				// immediately without re-running the scheduler every time.
				now := time.Now()
				if now.Sub(lastStaleLockCleanup) >= systemTaskStaleLockInterval {
					lastStaleLockCleanup = now
					if err := model.ExpireStaleSystemTaskLocks(common.GetTimestamp()); err != nil {
						logger.LogWarn(context.Background(), fmt.Sprintf("system task stale lock cleanup failed: %v", err))
					}
				}
				if now.Sub(lastScheduler) >= systemTaskSchedulerInterval {
					lastScheduler = now
					runSystemTaskScheduler()
				}
				runSystemTaskClaimPass(runnerID)
			}

			runPass()
			for {
				select {
				case <-ticker.C:
				case <-systemTaskWakeup:
				}
				runPass()
			}
		})
	})
}

func StartLogCleanupTask(targetTimestamp int64) (*model.SystemTask, error) {
	if targetTimestamp <= 0 {
		return nil, errors.New("target timestamp is required")
	}

	activeTask, err := model.GetActiveSystemTask(model.SystemTaskTypeLogCleanup)
	if err != nil {
		return nil, err
	}
	if activeTask != nil {
		return activeTask, nil
	}

	payload := LogCleanupPayload{
		TargetTimestamp: targetTimestamp,
		BatchSize:       logCleanupBatchSize,
	}
	state := LogCleanupState{}
	task, err := model.CreateSystemTask(model.SystemTaskTypeLogCleanup, payload, state)
	if err != nil {
		activeTask, activeErr := model.GetActiveSystemTask(model.SystemTaskTypeLogCleanup)
		if activeErr == nil && activeTask != nil {
			return activeTask, nil
		}
		return nil, err
	}
	notifySystemTaskRunner()
	return task, nil
}

func StartRequestBodyArchiveTask() (*model.SystemTask, error) {
	activeTask, err := model.GetActiveSystemTask(model.SystemTaskTypeRequestBodyArchive)
	if err != nil {
		return nil, err
	}
	if activeTask != nil {
		return activeTask, nil
	}

	task, err := model.CreateSystemTask(
		model.SystemTaskTypeRequestBodyArchive,
		RequestBodyArchivePayload{BatchSize: logCleanupBatchSize},
		RequestBodyArchiveState{},
	)
	if err != nil {
		activeTask, activeErr := model.GetActiveSystemTask(model.SystemTaskTypeRequestBodyArchive)
		if activeErr == nil && activeTask != nil {
			return activeTask, nil
		}
		return nil, err
	}
	notifySystemTaskRunner()
	return task, nil
}

// EnqueueSystemTask creates an on-demand task of the given type. The returned
// bool is true only when a new pending row was created; false means an active
// task of the same type already exists and was returned.
func EnqueueSystemTask(taskType string, payload any) (*model.SystemTask, bool, error) {
	activeTask, err := model.GetActiveSystemTask(taskType)
	if err != nil {
		return nil, false, err
	}
	if activeTask != nil {
		return activeTask, false, nil
	}

	task, err := model.CreateSystemTask(taskType, payload, nil)
	if err != nil {
		activeTask, activeErr := model.GetActiveSystemTask(taskType)
		if activeErr == nil && activeTask != nil {
			return activeTask, false, nil
		}
		return nil, false, err
	}
	notifySystemTaskRunner()
	return task, true, nil
}

// runSystemTaskClaimPass tries to claim one pending task per registered type
// and dispatches each claimed task in its own goroutine so a long-running
// handler (e.g. channel test) never blocks another type (e.g. log cleanup).
func runSystemTaskClaimPass(runnerID string) {
	handlers := registeredSystemTaskHandlers()
	taskTypes := make([]string, 0, len(handlers))
	for _, handler := range handlers {
		taskTypes = append(taskTypes, handler.Type())
	}
	pendingTasks, err := model.FindEarliestPendingSystemTasks(taskTypes)
	if err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("system task runner query failed: %v", err))
		return
	}
	for _, handler := range handlers {
		task := pendingTasks[handler.Type()]
		if task == nil {
			continue
		}
		claimedTask, claimed, err := model.ClaimSystemTask(task.ID, handler.Type(), runnerID, systemTaskLockUntil())
		if err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("system task claim failed: %v", err))
			continue
		}
		if !claimed {
			continue
		}
		dispatchHandler := handler
		dispatchTask := claimedTask
		gopool.Go(func() {
			runWithLeaseHeartbeat(dispatchTask, runnerID, func(ctx context.Context) {
				dispatchHandler.Run(ctx, dispatchTask, runnerID)
			})
		})
	}
}

// runSystemTaskScheduler creates a new task row for each enabled scheduled
// handler whose interval has elapsed since its last run and that has no active
// row. The task active_key unique index deduplicates concurrent creation while
// the per-type lock guarantees only one runner executes the task.
func runSystemTaskScheduler() {
	now := common.GetTimestamp()
	handlers := registeredSystemTaskHandlers()
	scheduledHandlers := make([]ScheduledSystemTaskHandler, 0, len(handlers))
	taskTypes := make([]string, 0, len(handlers))
	for _, handler := range handlers {
		scheduled, ok := handler.(ScheduledSystemTaskHandler)
		if !ok || !scheduled.Enabled() {
			continue
		}
		scheduledHandlers = append(scheduledHandlers, scheduled)
		taskTypes = append(taskTypes, scheduled.Type())
	}
	latestTasks, err := model.GetLatestSystemTasks(taskTypes)
	if err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("system task scheduler query failed: %v", err))
		return
	}
	for _, scheduled := range scheduledHandlers {
		latest := latestTasks[scheduled.Type()]
		if latest != nil {
			if latest.Status == model.SystemTaskStatusPending || latest.Status == model.SystemTaskStatusRunning {
				continue // an active row already exists
			}
			if now-latest.UpdatedAt < int64(scheduled.Interval().Seconds()) {
				continue // not due yet
			}
		}
		if _, err := model.CreateSystemTask(scheduled.Type(), scheduled.NewPayload(), nil); err != nil {
			activeTask, activeErr := model.GetActiveSystemTask(scheduled.Type())
			if activeErr == nil && activeTask != nil {
				continue
			}
			if activeErr != nil {
				logger.LogWarn(context.Background(), fmt.Sprintf("system task scheduler active lookup failed: type=%s err=%v", scheduled.Type(), activeErr))
			}
			logger.LogWarn(context.Background(), fmt.Sprintf("system task scheduler create failed: type=%s err=%v", scheduled.Type(), err))
			continue
		}
	}
}

// runWithLeaseHeartbeat renews the per-type lock on a background ticker while
// fn runs. The TTL is a crash-detection window, not a task time limit: an
// arbitrarily long handler stays alive as long as the heartbeat succeeds.
func runWithLeaseHeartbeat(task *model.SystemTask, runnerID string, fn func(ctx context.Context)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interval := systemTaskLockTTL / 3
	if interval <= 0 {
		interval = systemTaskLockTTL
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := model.RenewSystemTaskLock(task.TaskID, runnerID, systemTaskLockUntil()); err != nil {
					cancel()
					return
				}
			}
		}
	}()

	fn(ctx)
	close(done)
}

func runLogCleanupTask(ctx context.Context, task *model.SystemTask, runnerID string) {
	payload := LogCleanupPayload{}
	if err := task.DecodePayload(&payload); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if payload.TargetTimestamp <= 0 {
		failSystemTask(task, runnerID, errors.New("target timestamp is required"))
		return
	}
	if payload.BatchSize <= 0 {
		payload.BatchSize = logCleanupBatchSize
	}

	state := LogCleanupState{}
	if err := task.DecodeState(&state); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}

	for {
		remaining, err := model.CountOldLog(ctx, payload.TargetTimestamp)
		if err != nil {
			failSystemTask(task, runnerID, err)
			return
		}
		syncLogCleanupStateFromRemaining(&state, remaining)
		if err := model.UpdateSystemTaskState(task.TaskID, runnerID, state); err != nil {
			logSystemTaskLockError(ctx, task, err)
			return
		}
		if state.Remaining == 0 {
			break
		}

		// Track whether this pass deleted anything so a fresh recount that still
		// reports remaining rows resumes immediately instead of waiting for the
		// lock to expire. If a whole pass deletes nothing while rows remain, the
		// rows cannot be removed and we fail instead of busy-looping.
		progressed := false
		for state.Remaining > 0 {
			rowsAffected, err := model.DeleteOldLogBatch(ctx, payload.TargetTimestamp, payload.BatchSize)
			if err != nil {
				failSystemTask(task, runnerID, err)
				return
			}
			if rowsAffected == 0 {
				break
			}
			progressed = true

			state.Processed += rowsAffected
			if state.Total < state.Processed {
				state.Total = state.Processed
			}
			if state.Remaining > rowsAffected {
				state.Remaining -= rowsAffected
			} else {
				state.Remaining = 0
			}
			state.Progress = logCleanupProgress(state.Processed, state.Total)

			if err := model.UpdateSystemTaskState(task.TaskID, runnerID, state); err != nil {
				logSystemTaskLockError(ctx, task, err)
				return
			}
		}

		if !progressed {
			failSystemTask(task, runnerID, errors.New("no log rows were deleted"))
			return
		}
	}

	state.Remaining = 0
	state.Progress = 100
	if state.Total < state.Processed {
		state.Total = state.Processed
	}
	if err := model.UpdateSystemTaskState(task.TaskID, runnerID, state); err != nil {
		logSystemTaskLockError(ctx, task, err)
		return
	}

	requestBodiesDeleted, err := model.DeleteOrphanTaskRequestBodiesBefore(ctx, payload.TargetTimestamp)
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	result := LogCleanupResult{
		DeletedCount: state.Processed, RequestBodiesDeletedCount: requestBodiesDeleted,
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, result, ""); err != nil {
		logSystemTaskLockError(ctx, task, err)
	}
}

func runRequestBodyArchiveTask(ctx context.Context, task *model.SystemTask, runnerID string) {
	payload := RequestBodyArchivePayload{}
	if err := task.DecodePayload(&payload); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if payload.BatchSize <= 0 {
		payload.BatchSize = logCleanupBatchSize
	}

	state := RequestBodyArchiveState{}
	if err := task.DecodeState(&state); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if !state.Initialized {
		if err := initializeRequestBodyArchiveState(ctx, &state); err != nil {
			failSystemTask(task, runnerID, err)
			return
		}
		if err := model.UpdateSystemTaskState(task.TaskID, runnerID, state); err != nil {
			logSystemTaskLockError(ctx, task, err)
			return
		}
	}

	for state.Phase == "tasks" {
		if err := ctx.Err(); err != nil {
			failSystemTask(task, runnerID, err)
			return
		}
		rows, err := model.FindLegacyTaskRequestBodyBatch(ctx, state.TaskCursor, state.TaskUpperID, payload.BatchSize)
		if err != nil {
			failSystemTask(task, runnerID, err)
			return
		}
		if len(rows) == 0 {
			state.Phase = "logs"
			if err := model.UpdateSystemTaskState(task.TaskID, runnerID, state); err != nil {
				logSystemTaskLockError(ctx, task, err)
				return
			}
			break
		}
		for _, row := range rows {
			archived, archiveErr := model.ArchiveLegacyTaskRequestBody(ctx, row)
			if archiveErr != nil {
				failSystemTask(task, runnerID, archiveErr)
				return
			}
			if archived {
				state.ArchivedCount++
				state.TaskRowsCleaned++
			}
			state.TaskCursor = row.ID
			state.Processed++
		}
		state.Progress = requestBodyArchiveProgress(state.Processed, state.Total)
		if err := model.UpdateSystemTaskState(task.TaskID, runnerID, state); err != nil {
			logSystemTaskLockError(ctx, task, err)
			return
		}
	}

	for state.Phase == "logs" {
		if err := ctx.Err(); err != nil {
			failSystemTask(task, runnerID, err)
			return
		}
		rows, err := model.FindLogRequestBodyArchiveBatch(ctx, state.LogCursor, state.LogUpperCursor, payload.BatchSize)
		if err != nil {
			failSystemTask(task, runnerID, err)
			return
		}
		if len(rows) == 0 {
			state.Phase = "complete"
			break
		}

		updates := make([]model.LogRequestBodyArchiveUpdate, 0, len(rows))
		for _, row := range rows {
			state.LogCursor = model.LogRequestBodyArchiveCursor{
				ID: row.ID, CreatedAt: row.CreatedAt, RequestID: row.RequestID, OtherSHA256: row.OtherSHA256,
			}
			state.Processed++
			if row.Other == "" {
				continue
			}
			other, parseErr := common.StrToMap(row.Other)
			if parseErr != nil || other == nil {
				state.InvalidRowsSkipped++
				continue
			}
			requestBodyValue, hasRequestBody := other["request_body"]
			if !hasRequestBody {
				continue
			}
			requestBody, normalizeErr := normalizeArchivedRequestBody(requestBodyValue)
			if normalizeErr != nil {
				state.InvalidRowsSkipped++
				continue
			}
			taskID, _ := other["task_id"].(string)
			if taskID == "" && row.RequestID == "" {
				state.InvalidRowsSkipped++
				continue
			}
			if saveErr := model.SaveTaskRequestBodyAt(ctx, taskID, row.RequestID, requestBody, row.CreatedAt); saveErr != nil {
				failSystemTask(task, runnerID, saveErr)
				return
			}
			delete(other, "request_body")
			other["request_body_available"] = true
			if taskID == "" {
				other["request_body_ref"] = "request:" + row.RequestID
			}
			updates = append(updates, model.LogRequestBodyArchiveUpdate{
				ID: row.ID, CreatedAt: row.CreatedAt, RequestID: row.RequestID,
				OriginalOther: row.Other, Other: common.MapToJsonStr(other),
			})
			state.ArchivedCount++
			state.LogRowsCleaned++
		}
		if err := model.UpdateArchivedLogRequestBodies(ctx, updates); err != nil {
			failSystemTask(task, runnerID, err)
			return
		}
		state.Progress = requestBodyArchiveProgress(state.Processed, state.Total)
		if err := model.UpdateSystemTaskState(task.TaskID, runnerID, state); err != nil {
			logSystemTaskLockError(ctx, task, err)
			return
		}
	}

	state.Phase = "complete"
	state.Progress = 100
	if err := model.UpdateSystemTaskState(task.TaskID, runnerID, state); err != nil {
		logSystemTaskLockError(ctx, task, err)
		return
	}
	result := RequestBodyArchiveResult{
		ScannedRows: state.Processed, ArchivedCount: state.ArchivedCount,
		TaskRowsCleaned: state.TaskRowsCleaned, LogRowsCleaned: state.LogRowsCleaned,
		InvalidRowsSkipped: state.InvalidRowsSkipped,
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, result, ""); err != nil {
		logSystemTaskLockError(ctx, task, err)
	}
}

func initializeRequestBodyArchiveState(ctx context.Context, state *RequestBodyArchiveState) error {
	taskUpperID, err := model.GetTaskRequestBodyArchiveUpperID(ctx)
	if err != nil {
		return err
	}
	logUpperCursor, err := model.GetLogRequestBodyArchiveUpperCursor(ctx)
	if err != nil {
		return err
	}
	taskTotal, err := model.CountTasksThroughID(ctx, taskUpperID)
	if err != nil {
		return err
	}
	logTotal, err := model.CountLogsThroughCursor(ctx, logUpperCursor)
	if err != nil {
		return err
	}
	state.Initialized = true
	state.Phase = "tasks"
	state.Total = taskTotal + logTotal
	state.TaskUpperID = taskUpperID
	state.LogUpperCursor = logUpperCursor
	state.Progress = requestBodyArchiveProgress(state.Processed, state.Total)
	return nil
}

func normalizeArchivedRequestBody(value any) ([]byte, error) {
	if text, ok := value.(string); ok {
		var decoded any
		if err := common.Unmarshal([]byte(text), &decoded); err == nil {
			return []byte(text), nil
		}
	}
	return common.Marshal(value)
}

func requestBodyArchiveProgress(processed int64, total int64) int {
	if total <= 0 || processed >= total {
		return 100
	}
	if processed <= 0 {
		return 0
	}
	return int(processed * 100 / total)
}

func syncLogCleanupStateFromRemaining(state *LogCleanupState, remaining int64) {
	if state.Total <= 0 {
		state.Total = remaining
		state.Processed = 0
	} else {
		processedFromRemaining := state.Total - remaining
		if processedFromRemaining > state.Processed {
			state.Processed = processedFromRemaining
		}
	}
	if state.Processed < 0 {
		state.Processed = 0
	}
	state.Remaining = remaining
	state.Progress = logCleanupProgress(state.Processed, state.Total)
}

func logCleanupProgress(processed int64, total int64) int {
	if total <= 0 {
		return 100
	}
	if processed <= 0 {
		return 0
	}
	if processed >= total {
		return 100
	}
	return int(processed * 100 / total)
}

func systemTaskLockUntil() int64 {
	return common.GetTimestamp() + int64(systemTaskLockTTL.Seconds())
}

// SystemTaskProgress is the state shape used by handlers that report percentage
// progress (channel test, model update). The frontend reads the progress field
// (0-100) to render a per-task progress indicator.
type SystemTaskProgress struct {
	Total     int `json:"total"`
	Processed int `json:"processed"`
	Progress  int `json:"progress"`
}

// NewSystemTaskProgressReporter returns a throttled progress callback bound to a
// running task. Handlers call it with (processed, total) as they iterate work;
// it persists a {processed,total,progress} state at most once every ~2s, always
// emitting the first update and the final 100%.
// Lock-loss errors are ignored: the lease heartbeat cancels the handler ctx on
// loss, so progress writes are best-effort and never abort the run themselves.
// The returned func is single-goroutine only (call it from the handler loop).
func NewSystemTaskProgressReporter(task *model.SystemTask, runnerID string) func(processed, total int) {
	const minWriteInterval = 2 * time.Second
	var (
		lastWriteAt  time.Time
		lastProgress = -1
	)
	return func(processed, total int) {
		progress := 100
		if total > 0 {
			progress = processed * 100 / total
		}
		if progress < 0 {
			progress = 0
		} else if progress > 100 {
			progress = 100
		}

		if progress < 100 {
			if progress == lastProgress {
				return
			}
			if !lastWriteAt.IsZero() && time.Since(lastWriteAt) < minWriteInterval {
				return
			}
		}
		lastProgress = progress
		lastWriteAt = time.Now()

		state := SystemTaskProgress{Total: total, Processed: processed, Progress: progress}
		_ = model.UpdateSystemTaskState(task.TaskID, runnerID, state)
	}
}

func failSystemTask(task *model.SystemTask, runnerID string, err error) {
	logger.LogWarn(context.Background(), fmt.Sprintf("system task %s failed: %v", task.TaskID, err))
	if finishErr := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusFailed, nil, err.Error()); finishErr != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("system task %s failed to save failure state: %v", task.TaskID, finishErr))
	}
}

func logSystemTaskLockError(ctx context.Context, task *model.SystemTask, err error) {
	if errors.Is(err, model.ErrSystemTaskLockLost) {
		logger.LogWarn(ctx, fmt.Sprintf("system task %s lock lost", task.TaskID))
		return
	}
	logger.LogWarn(ctx, fmt.Sprintf("system task %s update failed: %v", task.TaskID, err))
}
