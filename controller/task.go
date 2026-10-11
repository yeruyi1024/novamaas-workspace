package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay"
	"github.com/QuantumNous/new-api/service/authz"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func GetAllTask(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)

	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	// 解析其他查询参数
	queryParams := model.SyncTaskQueryParams{
		Platform:       constant.TaskPlatform(c.Query("platform")),
		TaskID:         c.Query("task_id"),
		Status:         c.Query("status"),
		Action:         c.Query("action"),
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ChannelID:      c.Query("channel_id"),
		Username:       c.Query("username"),
		ModelName:      c.Query("model_name"),
	}

	type listResult struct {
		items []*model.Task
		err   error
	}
	type countResult struct {
		total int64
		err   error
	}
	listCh := make(chan listResult, 1)
	countCh := make(chan countResult, 1)
	go func() {
		items, queryErr := model.TaskGetAllTasks(pageInfo.GetStartIdx(), pageInfo.GetPageSize(), queryParams)
		listCh <- listResult{items: items, err: queryErr}
	}()
	go func() {
		total, queryErr := model.TaskCountAllTasks(queryParams)
		countCh <- countResult{total: total, err: queryErr}
	}()
	listed := <-listCh
	counted := <-countCh
	if listed.err != nil {
		common.ApiError(c, listed.err)
		return
	}
	if counted.err != nil {
		common.ApiError(c, counted.err)
		return
	}
	items := listed.items
	total := counted.total
	pageInfo.SetTotal(int(total))
	dtos, err := tasksToDto(items, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetItems(dtos)
	common.ApiSuccess(c, pageInfo)
}

func GetUserTask(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)

	userId := c.GetInt("id")

	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	queryParams := model.SyncTaskQueryParams{
		Platform:       constant.TaskPlatform(c.Query("platform")),
		TaskID:         c.Query("task_id"),
		Status:         c.Query("status"),
		Action:         c.Query("action"),
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ModelName:      c.Query("model_name"),
	}

	type listResult struct {
		items []*model.Task
		err   error
	}
	type countResult struct {
		total int64
		err   error
	}
	listCh := make(chan listResult, 1)
	countCh := make(chan countResult, 1)
	go func() {
		items, queryErr := model.TaskGetAllUserTask(userId, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), queryParams)
		listCh <- listResult{items: items, err: queryErr}
	}()
	go func() {
		total, queryErr := model.TaskCountAllUserTask(userId, queryParams)
		countCh <- countResult{total: total, err: queryErr}
	}()
	listed := <-listCh
	counted := <-countCh
	if listed.err != nil {
		common.ApiError(c, listed.err)
		return
	}
	if counted.err != nil {
		common.ApiError(c, counted.err)
		return
	}
	items := listed.items
	total := counted.total
	pageInfo.SetTotal(int(total))
	dtos, err := tasksToDto(items, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetItems(dtos)
	common.ApiSuccess(c, pageInfo)
}

func GetTaskRequestBody(c *gin.Context) {
	taskID := c.Param("task_id")
	snapshots, exists, err := model.GetArchivedRequestSnapshots(taskID, "")
	if err != nil {
		respondTaskRequestBody(c, nil, false, err)
		return
	}
	if exists {
		if authorizeRequestSnapshots(c, snapshots) {
			common.ApiSuccess(c, snapshots.Original)
		}
		return
	}

	// Compatibility fallback while the one-time archive task has not yet
	// migrated a legacy payload out of tasks.properties.
	task, exists, err := model.GetByTaskIdForAdmin(taskID)
	respondTaskRequestBody(c, task, exists, err)
}

func GetTaskRequestSnapshots(c *gin.Context) {
	respondTaskRequestSnapshots(c, c.Param("task_id"), "")
}

func GetTaskPollHistory(c *gin.Context) {
	taskID := c.Param("task_id")
	if _, exists, err := model.GetByTaskIdForAdmin(taskID); err != nil {
		common.ApiError(c, err)
		return
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "task not found"})
		return
	}
	beforeID, err := strconv.ParseInt(c.DefaultQuery("before_id", "0"), 10, 64)
	if err != nil || beforeID < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid before_id"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "limit must be between 1 and 100"})
		return
	}
	entries, err := model.ListTaskPollHistory(c.Request.Context(), taskID, beforeID, limit+1)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	hasMore := len(entries) > limit
	if hasMore {
		entries = entries[:limit]
	}
	var nextBeforeID int64
	if hasMore {
		nextBeforeID = entries[len(entries)-1].ID
	}
	common.ApiSuccess(c, gin.H{"items": entries, "next_before_id": nextBeforeID})
}

func GetLogRequestBody(c *gin.Context) {
	taskID := c.Query("task_id")
	requestID := c.Query("request_id")
	if taskID == "" && requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Task id or request id is required"})
		return
	}
	snapshots, exists, err := model.GetArchivedRequestSnapshots(taskID, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to query log request body"})
		return
	}
	if !exists {
		if taskID != "" {
			task, taskExists, taskErr := model.GetByTaskIdForAdmin(taskID)
			respondTaskRequestBody(c, task, taskExists, taskErr)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Log request body not found"})
		return
	}
	if authorizeRequestSnapshots(c, snapshots) {
		common.ApiSuccess(c, snapshots.Original)
	}
}

func GetLogRequestSnapshots(c *gin.Context) {
	taskID := c.Query("task_id")
	requestID := c.Query("request_id")
	if taskID == "" && requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Task id or request id is required"})
		return
	}
	respondTaskRequestSnapshots(c, taskID, requestID)
}

func respondTaskRequestSnapshots(c *gin.Context, taskID string, requestID string) {
	snapshots, exists, err := model.GetArchivedRequestSnapshots(taskID, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to query request snapshots"})
		return
	}
	if len(snapshots.Original) == 0 && taskID != "" {
		task, taskExists, taskErr := model.GetByTaskIdForAdmin(taskID)
		if taskErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to query request snapshots"})
			return
		}
		if taskExists && task != nil && len(task.Properties.RequestBody) > 0 {
			snapshots.Original = task.Properties.RequestBody
			exists = true
		}
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Request snapshots not found"})
		return
	}
	if authorizeRequestSnapshots(c, snapshots) {
		common.ApiSuccess(c, snapshots)
	}
}

func authorizeRequestSnapshots(c *gin.Context, snapshots model.TaskRequestSnapshots) bool {
	if snapshots.Source != model.RequestSnapshotSourceImageStudio {
		return true
	}
	if !authz.Can(c.GetInt("id"), c.GetInt("role"), authz.ImageStudioRead) || !authz.Can(c.GetInt("id"), c.GetInt("role"), authz.ImageStudioContentRead) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "image review content permission is required"})
		return false
	}
	recordManageAudit(c, "image_studio.content_read", map[string]interface{}{"request_id": c.Query("request_id"), "task_id": c.Param("task_id")})
	return true
}

func respondTaskRequestBody(c *gin.Context, task *model.Task, exists bool, err error) {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to query task request body"})
		return
	}
	if !exists || task == nil || len(task.Properties.RequestBody) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Task request body not found"})
		return
	}
	common.ApiSuccess(c, task.Properties.RequestBody)
}

func tasksToDto(tasks []*model.Task, fillUser bool) ([]*dto.TaskDto, error) {
	var usernames map[int]string
	if fillUser {
		userIds := types.NewSet[int]()
		for _, task := range tasks {
			userIds.Add(task.UserId)
		}
		var err error
		usernames, err = model.GetUsernamesByIDs(userIds.Items())
		if err != nil {
			return nil, err
		}
	}
	result := make([]*dto.TaskDto, len(tasks))
	for i, task := range tasks {
		if fillUser {
			if username, ok := usernames[task.UserId]; ok {
				task.Username = username
			}
		}
		result[i] = relay.TaskModel2Dto(task)
	}
	return result, nil
}
