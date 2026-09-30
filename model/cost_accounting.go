package model

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	CostAccountingSignedRefundVersion        = 2
	CostAccountingLegacyTurnoverBasisVersion = 3
	CostAccountingOriginalPriceBasisVersion  = 4
	CostAccountingSnapshotVersion            = CostAccountingOriginalPriceBasisVersion
	CostSnapshotSourceRealtime               = "realtime"
	CostSnapshotSourceBackfill               = "backfill"
)

var costDiscountPattern = regexp.MustCompile(`^(?:0(?:\.\d{1,6})?|1(?:\.0{1,6})?)$`)

// CostAccountingSnapshot is immutable accounting evidence captured alongside
// a usage log. Revenue, cost basis, and cost use the same sign: consumption is
// positive and refunds are negative. Later corrections are append-only
// adjustments.
type CostAccountingSnapshot struct {
	ID                int64  `json:"id" gorm:"primaryKey"`
	EventKey          string `json:"event_key" gorm:"type:varchar(191);uniqueIndex"`
	SourceLogID       int64  `json:"source_log_id" gorm:"index"`
	RequestID         string `json:"request_id" gorm:"type:varchar(128);index"`
	UpstreamRequestID string `json:"upstream_request_id" gorm:"type:varchar(128);index"`
	UserID            int    `json:"user_id" gorm:"index:idx_cost_snapshot_user_time,priority:1"`
	Username          string `json:"username" gorm:"type:varchar(191);index"`
	TokenName         string `json:"token_name" gorm:"type:varchar(191);index"`
	ModelName         string `json:"model_name" gorm:"type:varchar(255);index"`
	ChannelID         int    `json:"channel_id" gorm:"index"`
	GroupName         string `json:"group" gorm:"column:group_name;type:varchar(64);index"`
	LogType           int    `json:"log_type" gorm:"index"`
	OccurredAt        int64  `json:"occurred_at" gorm:"bigint;index:idx_cost_snapshot_user_time,priority:2;index"`
	RevenueQuota      int64  `json:"revenue_quota" gorm:"bigint"`
	CostBasisQuota    string `json:"cost_basis_quota" gorm:"type:varchar(64)"`
	CostDiscount      string `json:"cost_discount" gorm:"type:varchar(16)"`
	CostQuota         int64  `json:"cost_quota" gorm:"bigint"`
	SnapshotVersion   int    `json:"snapshot_version"`
	Source            string `json:"source" gorm:"type:varchar(16);index"`
	BatchID           string `json:"batch_id,omitempty" gorm:"type:varchar(64);index"`
	CreatedAt         int64  `json:"created_at" gorm:"bigint"`
}

// CostAccountingAdjustment never overwrites a snapshot. NewCostQuota is
// recorded as an audit convenience; DeltaCostQuota is the value aggregated.
type CostAccountingAdjustment struct {
	ID             int64  `json:"id" gorm:"primaryKey"`
	SnapshotID     int64  `json:"snapshot_id" gorm:"index"`
	DeltaCostQuota int64  `json:"delta_cost_quota" gorm:"bigint"`
	NewCostQuota   int64  `json:"new_cost_quota" gorm:"bigint"`
	CostDiscount   string `json:"cost_discount,omitempty" gorm:"type:varchar(16)"`
	Reason         string `json:"reason" gorm:"type:varchar(500)"`
	ActorID        int    `json:"actor_id" gorm:"index"`
	BatchID        string `json:"batch_id" gorm:"type:varchar(64);index"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint"`
}

type CostAccountingAdjustmentTarget struct {
	SnapshotID   int64
	NewCostQuota int64
	CostDiscount string
}

type CostAccountingInput struct {
	EventKey       string
	CostBasisQuota string
	CostDiscount   string
	CostQuota      int64
	Source         string
	BatchID        string
}

type CostAccountingFilter struct {
	StartTimestamp    int64
	EndTimestamp      int64
	UserID            int
	Username          string
	TokenName         string
	ModelName         string
	ChannelID         int
	channelIDs        []int
	Group             string
	LogType           int
	RequestID         string
	UpstreamRequestID string
}

type CostAccountingTotals struct {
	ConsumptionQuota   int64 `json:"consumption_quota"`
	RefundQuota        int64 `json:"refund_quota"`
	RevenueQuota       int64 `json:"revenue_quota"`
	CostQuota          int64 `json:"cost_quota"`
	ProfitQuota        int64 `json:"profit_quota"`
	Records            int64 `json:"records"`
	LinkedRecords      int64 `json:"linked_records"`
	DefaultedRecords   int64 `json:"defaulted_records"`
	DefaultedCostQuota int64 `json:"defaulted_cost_quota"`
	UsageRecords       int64 `json:"usage_records"`
	MissingCostRecords int64 `json:"missing_cost_records"`
	AccountingComplete bool  `json:"accounting_complete"`
}

type CostAccountingBucket struct {
	Bucket       int64 `json:"bucket" gorm:"column:bucket"`
	RevenueQuota int64 `json:"revenue_quota"`
	CostQuota    int64 `json:"cost_quota"`
	ProfitQuota  int64 `json:"profit_quota"`
	Records      int64 `json:"records"`
}

type CostAccountingSnapshotView struct {
	CostAccountingSnapshot `gorm:"embedded"`
	AdjustmentQuota        int64 `json:"adjustment_quota"`
	EffectiveCostQuota     int64 `json:"effective_cost_quota"`
	ProfitQuota            int64 `json:"profit_quota"`
}

type costSnapshotFallbackKey struct {
	RequestID    string
	LogType      int
	OccurredAt   int64
	RevenueQuota int64
	UserID       int
	ChannelID    int
	ModelName    string
	GroupName    string
}

func NormalizeCostDiscount(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if !costDiscountPattern.MatchString(value) {
		return "", errors.New("cost discount must be between 0 and 1 with at most 6 decimal places")
	}
	discount, err := decimal.NewFromString(value)
	if err != nil || discount.IsNegative() || discount.GreaterThan(decimal.NewFromInt(1)) {
		return "", errors.New("cost discount must be a decimal between 0 and 1")
	}
	return discount.StringFixed(6), nil
}

func CostSnapshotEventKey(log *Log) string {
	if log.Id > 0 {
		return fmt.Sprintf("log:%d:%d", log.Id, log.Type)
	}
	return fmt.Sprintf("log:%s:%d:%d:%d:%s", log.RequestId, log.Type, log.CreatedAt, log.Quota, log.ModelName)
}

func BuildCostAccountingSnapshot(log *Log, input *CostAccountingInput) (*CostAccountingSnapshot, error) {
	if log == nil || input == nil || (log.Type != LogTypeConsume && log.Type != LogTypeRefund) {
		return nil, nil
	}
	if log.Quota < 0 || log.Quota > common.MaxQuota {
		return nil, errors.New("revenue quota is outside the supported single-request range")
	}
	discount, err := NormalizeCostDiscount(input.CostDiscount)
	if err != nil {
		return nil, err
	}
	if input.CostQuota < -int64(common.MaxQuota) || input.CostQuota > int64(common.MaxQuota) {
		return nil, errors.New("cost quota is outside the supported single-request range")
	}
	costBasis, err := decimal.NewFromString(input.CostBasisQuota)
	if err != nil {
		return nil, errors.New("cost basis quota must be a valid decimal")
	}
	if log.Type == LogTypeConsume && (costBasis.IsNegative() || input.CostQuota < 0) {
		return nil, errors.New("consumption cost accounting values cannot be negative")
	}
	if log.Type == LogTypeRefund && (costBasis.IsPositive() || input.CostQuota > 0) {
		return nil, errors.New("refund cost accounting values cannot be positive")
	}
	revenue := int64(log.Quota)
	if log.Type == LogTypeRefund {
		revenue = -revenue
	}
	if discount == "" && (costBasis.Cmp(decimal.NewFromInt(revenue)) != 0 || input.CostQuota != revenue) {
		return nil, errors.New("unconfigured cost discount must use revenue as cost")
	}
	source := input.Source
	if source == "" {
		source = CostSnapshotSourceRealtime
	}
	eventKey := strings.TrimSpace(input.EventKey)
	if eventKey == "" {
		eventKey = CostSnapshotEventKey(log)
	}
	if len(eventKey) > 191 {
		return nil, errors.New("cost accounting event key is too long")
	}
	snapshot := &CostAccountingSnapshot{
		EventKey:          eventKey,
		SourceLogID:       int64(log.Id),
		RequestID:         log.RequestId,
		UpstreamRequestID: log.UpstreamRequestId,
		UserID:            log.UserId,
		Username:          log.Username,
		TokenName:         log.TokenName,
		ModelName:         log.ModelName,
		ChannelID:         log.ChannelId,
		GroupName:         log.Group,
		LogType:           log.Type,
		OccurredAt:        log.CreatedAt,
		RevenueQuota:      revenue,
		CostBasisQuota:    costBasis.StringFixed(6),
		CostDiscount:      discount,
		CostQuota:         input.CostQuota,
		SnapshotVersion:   CostAccountingSnapshotVersion,
		Source:            source,
		BatchID:           input.BatchID,
		CreatedAt:         common.GetTimestamp(),
	}
	return snapshot, nil
}

func RecordCostAccountingSnapshot(log *Log, input *CostAccountingInput) error {
	snapshot, err := BuildCostAccountingSnapshot(log, input)
	if err != nil || snapshot == nil {
		return err
	}
	return DB.Clauses(clause.OnConflict{DoNothing: true}).Create(snapshot).Error
}

func InsertCostAccountingSnapshots(snapshots []CostAccountingSnapshot) (int64, error) {
	if len(snapshots) == 0 {
		return 0, nil
	}
	result := DB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&snapshots, 500)
	return result.RowsAffected, result.Error
}

// ExistingCostSnapshotLogIndexes identifies logs that already have immutable
// cost evidence, even when realtime task billing used a task-scoped event key.
func ExistingCostSnapshotLogIndexes(logs []Log) (map[int]struct{}, error) {
	result := make(map[int]struct{})
	if len(logs) == 0 {
		return result, nil
	}
	eventKeys := make([]string, 0, len(logs))
	sourceLogIDs := make([]int64, 0, len(logs))
	requestIDs := make([]string, 0, len(logs))
	for i := range logs {
		eventKeys = append(eventKeys, CostSnapshotEventKey(&logs[i]))
		if logs[i].Id > 0 {
			sourceLogIDs = append(sourceLogIDs, int64(logs[i].Id))
		}
		if logs[i].RequestId != "" {
			requestIDs = append(requestIDs, logs[i].RequestId)
		}
	}

	query := DB.Model(&CostAccountingSnapshot{}).Where("event_key IN ?", eventKeys)
	if len(sourceLogIDs) > 0 {
		query = query.Or("source_log_id IN ?", sourceLogIDs)
	}
	if len(requestIDs) > 0 {
		query = query.Or("request_id IN ?", requestIDs)
	}
	var snapshots []CostAccountingSnapshot
	if err := query.Find(&snapshots).Error; err != nil {
		return nil, err
	}

	existingEventKeys := make(map[string]struct{}, len(snapshots))
	existingSourceLogIDs := make(map[int64]struct{}, len(snapshots))
	existingFallbackKeys := make(map[costSnapshotFallbackKey]struct{}, len(snapshots))
	for i := range snapshots {
		existingEventKeys[snapshots[i].EventKey] = struct{}{}
		if snapshots[i].SourceLogID > 0 {
			existingSourceLogIDs[snapshots[i].SourceLogID] = struct{}{}
		}
		if snapshots[i].RequestID != "" {
			existingFallbackKeys[costSnapshotFallbackKey{
				RequestID:    snapshots[i].RequestID,
				LogType:      snapshots[i].LogType,
				OccurredAt:   snapshots[i].OccurredAt,
				RevenueQuota: snapshots[i].RevenueQuota,
				UserID:       snapshots[i].UserID,
				ChannelID:    snapshots[i].ChannelID,
				ModelName:    snapshots[i].ModelName,
				GroupName:    snapshots[i].GroupName,
			}] = struct{}{}
		}
	}
	for i := range logs {
		if _, ok := existingEventKeys[CostSnapshotEventKey(&logs[i])]; ok {
			result[i] = struct{}{}
			continue
		}
		if logs[i].Id > 0 {
			if _, ok := existingSourceLogIDs[int64(logs[i].Id)]; ok {
				result[i] = struct{}{}
				continue
			}
		}
		revenueQuota := int64(logs[i].Quota)
		if logs[i].Type == LogTypeRefund {
			revenueQuota = -revenueQuota
		}
		if logs[i].RequestId != "" {
			fallbackKey := costSnapshotFallbackKey{
				RequestID:    logs[i].RequestId,
				LogType:      logs[i].Type,
				OccurredAt:   logs[i].CreatedAt,
				RevenueQuota: revenueQuota,
				UserID:       logs[i].UserId,
				ChannelID:    logs[i].ChannelId,
				ModelName:    logs[i].ModelName,
				GroupName:    logs[i].Group,
			}
			if _, ok := existingFallbackKeys[fallbackKey]; ok {
				result[i] = struct{}{}
			}
		}
	}
	return result, nil
}

func ListLogsForCostBackfill(startTimestamp, endTimestamp int64, channelID, offset, limit int) ([]Log, error) {
	if startTimestamp <= 0 || endTimestamp < startTimestamp || offset < 0 || limit <= 0 || limit > 5000 {
		return nil, errors.New("invalid cost accounting backfill range")
	}
	query := LOG_DB.Model(&Log{}).
		Where("created_at >= ? AND created_at <= ?", startTimestamp, endTimestamp).
		Where("type IN ?", []int{LogTypeConsume, LogTypeRefund})
	if channelID > 0 {
		query = query.Where("channel_id = ?", channelID)
	}
	logs := make([]Log, 0, limit)
	err := query.Order("created_at asc").Order("request_id asc").Order("id asc").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, err
}

func GetLogsForCostAccounting(logIDs []int64) (map[int64]Log, error) {
	logsByID := make(map[int64]Log, len(logIDs))
	if len(logIDs) == 0 {
		return logsByID, nil
	}
	logs := make([]Log, 0, len(logIDs))
	if err := LOG_DB.Where("id IN ?", logIDs).Find(&logs).Error; err != nil {
		return nil, err
	}
	for i := range logs {
		logsByID[int64(logs[i].Id)] = logs[i]
	}
	return logsByID, nil
}

func applyCostAccountingTextFilter(tx *gorm.DB, column, value string) (*gorm.DB, error) {
	if value == "" {
		return tx, nil
	}
	if !strings.Contains(value, "%") {
		return tx.Where(column+" = ?", value), nil
	}
	pattern, err := sanitizeLikePattern(value)
	if err != nil {
		return nil, err
	}
	return tx.Where(column+" LIKE ? ESCAPE '!'", pattern), nil
}

func applyCostAccountingFilter(tx *gorm.DB, filter CostAccountingFilter) (*gorm.DB, error) {
	if filter.StartTimestamp != 0 {
		tx = tx.Where("snapshots.occurred_at >= ?", filter.StartTimestamp)
	}
	if filter.EndTimestamp != 0 {
		tx = tx.Where("snapshots.occurred_at <= ?", filter.EndTimestamp)
	}
	if filter.UserID != 0 {
		tx = tx.Where("snapshots.user_id = ?", filter.UserID)
	}
	var err error
	if tx, err = applyCostAccountingTextFilter(tx, "snapshots.username", filter.Username); err != nil {
		return nil, err
	}
	if filter.TokenName != "" {
		tx = tx.Where("snapshots.token_name = ?", filter.TokenName)
	}
	if tx, err = applyCostAccountingTextFilter(tx, "snapshots.model_name", filter.ModelName); err != nil {
		return nil, err
	}
	if filter.ChannelID != 0 {
		tx = tx.Where("snapshots.channel_id = ?", filter.ChannelID)
	}
	if len(filter.channelIDs) > 0 {
		tx = tx.Where("snapshots.channel_id IN ?", filter.channelIDs)
	}
	if filter.Group != "" {
		tx = tx.Where("snapshots.group_name = ?", filter.Group)
	}
	if filter.LogType != LogTypeUnknown {
		tx = tx.Where("snapshots.log_type = ?", filter.LogType)
	}
	if filter.RequestID != "" {
		tx = tx.Where("snapshots.request_id = ?", filter.RequestID)
	}
	if filter.UpstreamRequestID != "" {
		tx = tx.Where("snapshots.upstream_request_id = ?", filter.UpstreamRequestID)
	}
	return tx, nil
}

func costAccountingQuery(filter CostAccountingFilter) (*gorm.DB, error) {
	adjustments := DB.Model(&CostAccountingAdjustment{}).
		Select("snapshot_id, SUM(delta_cost_quota) AS adjustment_quota").
		Group("snapshot_id")
	// The first snapshot is the immutable evidence for a source log. Historical
	// corrections belong in adjustments; a later duplicate snapshot must never
	// inflate revenue or cost totals.
	query := DB.Table("cost_accounting_snapshots AS snapshots").
		Joins("LEFT JOIN (?) AS adjustments ON adjustments.snapshot_id = snapshots.id", adjustments).
		Where(`snapshots.source_log_id <= 0 OR NOT EXISTS (
			SELECT 1 FROM cost_accounting_snapshots AS earlier
			WHERE earlier.id < snapshots.id
			AND earlier.source_log_id = snapshots.source_log_id
		)`)
	return applyCostAccountingFilter(query, filter)
}

func sumCostAccountingSnapshots(filter CostAccountingFilter) (CostAccountingTotals, error) {
	var totals CostAccountingTotals
	query, err := costAccountingQuery(filter)
	if err != nil {
		return totals, err
	}
	err = query.Select(`
		COALESCE(SUM(CASE WHEN snapshots.log_type = 2 THEN snapshots.revenue_quota ELSE 0 END), 0) AS consumption_quota,
		COALESCE(SUM(CASE WHEN snapshots.log_type = 6 THEN -snapshots.revenue_quota ELSE 0 END), 0) AS refund_quota,
		COALESCE(SUM(snapshots.revenue_quota), 0) AS revenue_quota,
		COALESCE(SUM(snapshots.cost_quota + COALESCE(adjustments.adjustment_quota, 0)), 0) AS cost_quota,
		COUNT(*) AS records,
		COUNT(DISTINCT CASE WHEN snapshots.source_log_id > 0 THEN snapshots.source_log_id END) AS linked_records`).Scan(&totals).Error
	totals.ProfitQuota = totals.RevenueQuota - totals.CostQuota
	return totals, err
}

func unconfiguredCostChannelIDs(filter CostAccountingFilter) ([]int, error) {
	query := DB.Model(&Channel{}).Where("cost_discount IS NULL OR TRIM(cost_discount) = ?", "")
	if filter.ChannelID != 0 {
		query = query.Where("id = ?", filter.ChannelID)
	}
	if len(filter.channelIDs) > 0 {
		query = query.Where("id IN ?", filter.channelIDs)
	}
	var channelIDs []int
	return channelIDs, query.Pluck("id", &channelIDs).Error
}

func SumCostAccounting(filter CostAccountingFilter) (CostAccountingTotals, error) {
	totals, err := sumCostAccountingSnapshots(filter)
	if err != nil {
		return totals, err
	}
	channelIDs, err := unconfiguredCostChannelIDs(filter)
	if err != nil || len(channelIDs) == 0 {
		return totals, err
	}
	fallbackFilter := filter
	fallbackFilter.ChannelID = 0
	fallbackFilter.channelIDs = channelIDs
	statistics, err := sumLogFinancialStatistics(fallbackFilter)
	if err != nil {
		return totals, err
	}
	snapshots, err := sumCostAccountingSnapshots(fallbackFilter)
	if err != nil {
		return totals, err
	}
	if statistics.Records <= snapshots.LinkedRecords {
		return totals, nil
	}
	totals.DefaultedRecords = statistics.Records - snapshots.LinkedRecords
	totals.DefaultedCostQuota = statistics.RevenueQuota - snapshots.RevenueQuota
	totals.ConsumptionQuota += statistics.Quota - snapshots.ConsumptionQuota
	totals.RefundQuota += statistics.RefundQuota - snapshots.RefundQuota
	totals.RevenueQuota += statistics.RevenueQuota - snapshots.RevenueQuota
	totals.CostQuota += totals.DefaultedCostQuota
	totals.ProfitQuota = totals.RevenueQuota - totals.CostQuota
	return totals, nil
}

func ReconcileCostAccountingTotals(statistics LogStatistics, totals CostAccountingTotals) CostAccountingTotals {
	revenueQuota := statistics.Quota - statistics.RefundQuota
	totals.ConsumptionQuota = statistics.Quota
	totals.RefundQuota = statistics.RefundQuota
	totals.RevenueQuota = revenueQuota
	totals.ProfitQuota = revenueQuota - totals.CostQuota
	totals.UsageRecords = statistics.Records
	accountedRecords := totals.LinkedRecords + totals.DefaultedRecords
	if statistics.Records > accountedRecords {
		totals.MissingCostRecords = statistics.Records - accountedRecords
	}
	totals.AccountingComplete = statistics.Records == accountedRecords && totals.Records == totals.LinkedRecords
	return totals
}

func sumCostAccountingSnapshotBuckets(filter CostAccountingFilter, bucketSeconds, offsetSeconds int64) ([]CostAccountingBucket, error) {
	if bucketSeconds <= 0 {
		return nil, errors.New("invalid accounting bucket size")
	}
	bucketExpr := "CAST((snapshots.occurred_at + ?) / ? AS BIGINT) * ? - ?"
	if common.UsingMainDatabase(common.DatabaseTypeMySQL) {
		bucketExpr = "((snapshots.occurred_at + ?) DIV ?) * ? - ?"
	}
	buckets := make([]CostAccountingBucket, 0)
	query, err := costAccountingQuery(filter)
	if err != nil {
		return nil, err
	}
	err = query.
		Select(bucketExpr+` AS bucket,
			COALESCE(SUM(snapshots.revenue_quota), 0) AS revenue_quota,
			COALESCE(SUM(snapshots.cost_quota + COALESCE(adjustments.adjustment_quota, 0)), 0) AS cost_quota,
			COUNT(*) AS records`, offsetSeconds, bucketSeconds, bucketSeconds, offsetSeconds).
		Group("bucket").Order("bucket asc").Scan(&buckets).Error
	for i := range buckets {
		buckets[i].ProfitQuota = buckets[i].RevenueQuota - buckets[i].CostQuota
	}
	return buckets, err
}

func SumCostAccountingBuckets(filter CostAccountingFilter, bucketSeconds, offsetSeconds int64) ([]CostAccountingBucket, error) {
	buckets, err := sumCostAccountingSnapshotBuckets(filter, bucketSeconds, offsetSeconds)
	if err != nil {
		return nil, err
	}
	channelIDs, err := unconfiguredCostChannelIDs(filter)
	if err != nil || len(channelIDs) == 0 {
		return buckets, err
	}
	fallbackFilter := filter
	fallbackFilter.ChannelID = 0
	fallbackFilter.channelIDs = channelIDs
	logBuckets, err := sumLogAccountingBuckets(fallbackFilter, bucketSeconds, offsetSeconds)
	if err != nil {
		return nil, err
	}
	snapshotBuckets, err := sumCostAccountingSnapshotBuckets(fallbackFilter, bucketSeconds, offsetSeconds)
	if err != nil {
		return nil, err
	}
	snapshotRevenueByBucket := make(map[int64]int64, len(snapshotBuckets))
	for _, bucket := range snapshotBuckets {
		snapshotRevenueByBucket[bucket.Bucket] = bucket.RevenueQuota
	}
	bucketIndex := make(map[int64]int, len(buckets))
	for index := range buckets {
		bucketIndex[buckets[index].Bucket] = index
	}
	for _, logBucket := range logBuckets {
		fallbackCost := logBucket.RevenueQuota - snapshotRevenueByBucket[logBucket.Bucket]
		if fallbackCost == 0 {
			continue
		}
		index, ok := bucketIndex[logBucket.Bucket]
		if !ok {
			buckets = append(buckets, CostAccountingBucket{Bucket: logBucket.Bucket})
			index = len(buckets) - 1
			bucketIndex[logBucket.Bucket] = index
		}
		buckets[index].RevenueQuota += fallbackCost
		buckets[index].CostQuota += fallbackCost
	}
	for index := range buckets {
		buckets[index].ProfitQuota = buckets[index].RevenueQuota - buckets[index].CostQuota
	}
	sort.Slice(buckets, func(i, j int) bool { return buckets[i].Bucket < buckets[j].Bucket })
	return buckets, nil
}

func ListCostAccountingSnapshots(filter CostAccountingFilter, offset, limit int) ([]CostAccountingSnapshotView, int64, error) {
	if offset < 0 || limit <= 0 || limit > 1000 {
		return nil, 0, errors.New("invalid cost accounting snapshot page")
	}
	var total int64
	countQuery, err := costAccountingQuery(filter)
	if err != nil {
		return nil, 0, err
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	views := make([]CostAccountingSnapshotView, 0, limit)
	query, err := costAccountingQuery(filter)
	if err != nil {
		return nil, 0, err
	}
	err = query.
		Select(`snapshots.*,
			COALESCE(adjustments.adjustment_quota, 0) AS adjustment_quota,
			snapshots.cost_quota + COALESCE(adjustments.adjustment_quota, 0) AS effective_cost_quota`).
		Order("snapshots.occurred_at DESC").Order("snapshots.id DESC").Offset(offset).Limit(limit).Scan(&views).Error
	for i := range views {
		views[i].ProfitQuota = views[i].RevenueQuota - views[i].EffectiveCostQuota
	}
	return views, total, err
}

// AttachLogAccounting adds the effective cost of each linked snapshot to a
// page of usage logs. It must only be called for a viewer with financial
// accounting permission, before ClickHouse or self views replace log IDs.
func AttachLogAccounting(logs []*Log) error {
	logIDs := make([]int64, 0, len(logs))
	requestIDs := make([]string, 0, len(logs))
	for _, log := range logs {
		if log.Type != LogTypeConsume && log.Type != LogTypeRefund {
			continue
		}
		revenue := int64(log.Quota)
		if log.Type == LogTypeRefund {
			revenue = -revenue
		}
		log.RevenueQuota = &revenue
		if log.Id > 0 {
			logIDs = append(logIDs, int64(log.Id))
		}
		if log.RequestId != "" {
			requestIDs = append(requestIDs, log.RequestId)
		}
	}
	if len(logIDs) == 0 && len(requestIDs) == 0 {
		return nil
	}

	query, err := costAccountingQuery(CostAccountingFilter{})
	if err != nil {
		return err
	}
	if len(logIDs) > 0 && len(requestIDs) > 0 {
		query = query.Where("snapshots.source_log_id IN ? OR snapshots.request_id IN ?", logIDs, requestIDs)
	} else if len(logIDs) > 0 {
		query = query.Where("snapshots.source_log_id IN ?", logIDs)
	} else {
		query = query.Where("snapshots.request_id IN ?", requestIDs)
	}
	views := make([]CostAccountingSnapshotView, 0, len(logs))
	if err := query.Select(`snapshots.*,
		snapshots.cost_quota + COALESCE(adjustments.adjustment_quota, 0) AS effective_cost_quota`).
		Scan(&views).Error; err != nil {
		return err
	}

	byLogID := make(map[int64][]CostAccountingSnapshotView, len(views))
	byRequestID := make(map[string][]CostAccountingSnapshotView, len(views))
	for _, view := range views {
		if view.SourceLogID > 0 {
			byLogID[view.SourceLogID] = append(byLogID[view.SourceLogID], view)
		}
		if view.RequestID != "" {
			byRequestID[view.RequestID] = append(byRequestID[view.RequestID], view)
		}
	}
	usedSnapshots := make(map[int64]bool, len(views))
	for _, log := range logs {
		if log.RevenueQuota == nil {
			continue
		}
		candidates := append(byLogID[int64(log.Id)], byRequestID[log.RequestId]...)
		for _, view := range candidates {
			if usedSnapshots[view.ID] || !matchesLogAccountingSnapshot(log, view) {
				continue
			}
			cost := view.EffectiveCostQuota
			profit := *log.RevenueQuota - cost
			log.CostQuota = &cost
			log.ProfitQuota = &profit
			usedSnapshots[view.ID] = true
			break
		}
	}
	return nil
}

func matchesLogAccountingSnapshot(log *Log, view CostAccountingSnapshotView) bool {
	return view.LogType == log.Type &&
		view.UserID == log.UserId &&
		view.OccurredAt == log.CreatedAt &&
		view.RevenueQuota == *log.RevenueQuota &&
		view.ChannelID == log.ChannelId &&
		view.ModelName == log.ModelName &&
		(view.RequestID == "" || view.RequestID == log.RequestId)
}

func AdjustCostAccountingSnapshots(targets []CostAccountingAdjustmentTarget, reason string, actorID int, batchID string) ([]CostAccountingAdjustment, error) {
	reason = strings.TrimSpace(reason)
	batchID = strings.TrimSpace(batchID)
	if len(targets) == 0 {
		return nil, nil
	}
	if reason == "" || len(reason) > 500 || len(batchID) > 64 {
		return nil, errors.New("invalid cost accounting adjustment metadata")
	}
	targets = append([]CostAccountingAdjustmentTarget(nil), targets...)
	sort.Slice(targets, func(i, j int) bool { return targets[i].SnapshotID < targets[j].SnapshotID })
	for i, target := range targets {
		if target.SnapshotID <= 0 || target.NewCostQuota < -int64(common.MaxQuota) || target.NewCostQuota > int64(common.MaxQuota) ||
			(i > 0 && target.SnapshotID == targets[i-1].SnapshotID) {
			return nil, errors.New("invalid cost accounting adjustment target")
		}
		if target.CostDiscount != "" {
			discount, err := NormalizeCostDiscount(target.CostDiscount)
			if err != nil {
				return nil, err
			}
			targets[i].CostDiscount = discount
		}
	}
	results := make([]CostAccountingAdjustment, 0, len(targets))
	err := DB.Transaction(func(tx *gorm.DB) error {
		createdAt := common.GetTimestamp()
		snapshotIDs := make([]int64, len(targets))
		for i := range targets {
			snapshotIDs[i] = targets[i].SnapshotID
		}
		snapshots := make([]CostAccountingSnapshot, 0, len(targets))
		if err := lockForUpdate(tx).Where("id IN ?", snapshotIDs).Find(&snapshots).Error; err != nil {
			return err
		}
		if len(snapshots) != len(targets) {
			return errors.New("cost accounting adjustment target not found")
		}
		snapshotsByID := make(map[int64]CostAccountingSnapshot, len(snapshots))
		for _, snapshot := range snapshots {
			snapshotsByID[snapshot.ID] = snapshot
		}
		var adjustmentRows []struct {
			SnapshotID      int64 `gorm:"column:snapshot_id"`
			AdjustmentQuota int64 `gorm:"column:adjustment_quota"`
		}
		if err := tx.Model(&CostAccountingAdjustment{}).
			Select("snapshot_id, COALESCE(SUM(delta_cost_quota), 0) AS adjustment_quota").
			Where("snapshot_id IN ?", snapshotIDs).
			Group("snapshot_id").
			Scan(&adjustmentRows).Error; err != nil {
			return err
		}
		adjustmentsBySnapshotID := make(map[int64]int64, len(adjustmentRows))
		for _, row := range adjustmentRows {
			adjustmentsBySnapshotID[row.SnapshotID] = row.AdjustmentQuota
		}
		for _, target := range targets {
			snapshot := snapshotsByID[target.SnapshotID]
			if snapshot.LogType == LogTypeConsume && target.NewCostQuota < 0 {
				return errors.New("consumption cost adjustment cannot be negative")
			}
			if snapshot.LogType == LogTypeRefund && target.NewCostQuota > 0 {
				return errors.New("refund cost adjustment cannot be positive")
			}
			currentCost := snapshot.CostQuota + adjustmentsBySnapshotID[target.SnapshotID]
			adjustment := CostAccountingAdjustment{
				SnapshotID:     target.SnapshotID,
				DeltaCostQuota: target.NewCostQuota - currentCost,
				NewCostQuota:   target.NewCostQuota,
				CostDiscount:   target.CostDiscount,
				Reason:         reason,
				ActorID:        actorID,
				BatchID:        batchID,
				CreatedAt:      createdAt,
			}
			if adjustment.DeltaCostQuota != 0 {
				results = append(results, adjustment)
			}
		}
		if len(results) == 0 {
			return nil
		}
		return tx.CreateInBatches(&results, 500).Error
	})
	return results, err
}
