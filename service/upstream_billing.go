package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"sort"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

var ErrUpstreamBillingIncomplete = errors.New("upstream accounting evidence is incomplete")
var ErrUpstreamBillingMonthOpen = errors.New("upstream accounting month is not closed for 24 hours")

type UpstreamBillingRow struct {
	BillingRow
	model.UpstreamBillingUsage
	ChargeCount           int64 `json:"charge_count"`
	RefundCount           int64 `json:"refund_count"`
	MissingCostRecords    int64 `json:"missing_cost_records"`
	MissingUsageRecords   int64 `json:"missing_usage_records"`
	EstimatedUsageRecords int64 `json:"estimated_usage_records"`
	PendingTasks          int64 `json:"pending_tasks"`
}

type UpstreamBillingModelRow struct {
	ModelName         string `json:"model_name"`
	UpstreamModelName string `json:"upstream_model_name,omitempty"`
	BillingGroup      string `json:"billing_group,omitempty"`
	ActiveDays        int64  `json:"active_days"`
	FirstPosted       int64  `json:"first_posted_at"`
	LastPosted        int64  `json:"last_posted_at"`
	UpstreamBillingRow
}

type UpstreamBillingDayModelRow struct {
	Date string `json:"date"`
	UpstreamBillingModelRow
}

type UpstreamBillingCoverage struct {
	Complete              bool                               `json:"complete"`
	MissingCostRecords    int64                              `json:"missing_cost_records"`
	MissingUsageRecords   int64                              `json:"missing_usage_records"`
	EstimatedUsageRecords int64                              `json:"estimated_usage_records"`
	PendingTasks          int64                              `json:"pending_tasks"`
	MissingChannels       int64                              `json:"missing_channels"`
	ExcludedChannels      []model.UpstreamBillingChannelView `json:"excluded_channels"`
}

type UpstreamBillingReport struct {
	SchemaVersion       int                             `json:"schema_version"`
	AccountingBasis     string                          `json:"accounting_basis"`
	Timezone            string                          `json:"timezone"`
	Month               string                          `json:"month,omitempty"`
	StartDate           string                          `json:"start_date"`
	EndDate             string                          `json:"end_date"`
	StartAt             int64                           `json:"start_at"`
	EndAt               int64                           `json:"end_at"`
	Group               model.UpstreamBillingGroupView  `json:"group"`
	Currency            BillingCurrency                 `json:"currency"`
	Total               UpstreamBillingRow              `json:"total"`
	Days                []UpstreamBillingRow            `json:"days"`
	Models              []UpstreamBillingModelRow       `json:"models"`
	DayModels           []UpstreamBillingDayModelRow    `json:"day_models"`
	ChannelUsers        []UpstreamBillingChannelUserRow `json:"channel_users,omitempty"`
	Coverage            UpstreamBillingCoverage         `json:"coverage"`
	SourceSHA256        string                          `json:"source_sha256"`
	UpdatedAt           int64                           `json:"updated_at"`
	RoundingDifference  string                          `json:"rounding_difference"`
	Issuer              string                          `json:"issuer"`
	PDFLogoPNG          []byte                          `json:"pdf_logo_png,omitempty"`
	PDFOperatingLogoPNG []byte                          `json:"pdf_operating_logo_png,omitempty"`
	OperatingName       string                          `json:"operating_name,omitempty"`
	PDFFooter           string                          `json:"pdf_footer,omitempty"`
}

type UpstreamBillingSnapshot = UpstreamBillingReport

type UpstreamBillingUsageDetails struct {
	Currency   BillingCurrency                  `json:"currency"`
	Records    []UpstreamBillingUsageRecordView `json:"records"`
	NextCursor int64                            `json:"next_cursor,omitempty"`
}

type UpstreamBillingUsageRecordView struct {
	model.UpstreamBillingStatementRecord
	Amount string `json:"amount"`
}

type upstreamBillingModelKey struct{ model, upstream, group string }

type upstreamBillingAggregate struct {
	charge, refund, count, chargeCount, refundCount int64
	usage                                           model.UpstreamBillingUsage
	missingCost, missingUsage, estimated, pending   int64
	first, last                                     int64
	days                                            map[string]bool
}

func (total *upstreamBillingAggregate) Add(record model.UpstreamBillingStatementRecord) error {
	if record.CostQuota > int64(common.MaxQuota) || record.CostQuota < -int64(common.MaxQuota) {
		return model.ErrBillingEvidenceIntegrity
	}
	// A missing supplier price is evidence to retain, not a zero-price fact.
	// Only known costs contribute to the subtotal frozen into the statement.
	if !record.CostMissing && record.CostQuota >= 0 {
		if record.CostQuota > int64(common.MaxWalletQuota)-total.charge {
			return model.ErrBillingEvidenceIntegrity
		}
		total.charge += record.CostQuota
	} else if !record.CostMissing {
		if -record.CostQuota > int64(common.MaxWalletQuota)-total.refund {
			return model.ErrBillingEvidenceIntegrity
		}
		total.refund -= record.CostQuota
	}
	if total.count == math.MaxInt64 {
		return model.ErrBillingEvidenceIntegrity
	}
	total.count++
	if record.LogType == model.LogTypeRefund {
		total.refundCount++
	} else {
		total.chargeCount++
	}
	values := []*int64{&total.usage.Requests, &total.usage.InputTokens, &total.usage.OutputTokens, &total.usage.CacheReadTokens, &total.usage.CacheWriteTokens, &total.usage.CacheWrite5mTokens, &total.usage.CacheWrite1hTokens, &total.usage.VideoTokens, &total.usage.VideoSeconds}
	increments := []int64{record.Requests, record.InputTokens, record.OutputTokens, record.CacheReadTokens, record.CacheWriteTokens, record.CacheWrite5mTokens, record.CacheWrite1hTokens, record.VideoTokens, record.VideoSeconds}
	for i, value := range values {
		if increments[i] < 0 || increments[i] > int64(common.MaxWalletQuota)-*value {
			return model.ErrBillingEvidenceIntegrity
		}
		*value += increments[i]
	}
	if record.CostMissing {
		total.missingCost++
	}
	if record.UsageMissing {
		total.missingUsage++
	}
	if record.UsageEstimated {
		total.estimated++
	}
	if record.TaskPending {
		total.pending++
	}
	if total.first == 0 || record.OccurredAt < total.first {
		total.first = record.OccurredAt
	}
	if record.OccurredAt > total.last {
		total.last = record.OccurredAt
	}
	if total.days == nil {
		total.days = make(map[string]bool)
	}
	total.days[time.Unix(record.OccurredAt, 0).In(billingLocation).Format("2006-01-02")] = true
	return nil
}

func (total *upstreamBillingAggregate) Row(label string, currency BillingCurrency) (UpstreamBillingRow, error) {
	row, err := billingRow(label, total.charge, total.refund, total.count, currency)
	return UpstreamBillingRow{BillingRow: row, UpstreamBillingUsage: total.usage, ChargeCount: total.chargeCount, RefundCount: total.refundCount, MissingCostRecords: total.missingCost, MissingUsageRecords: total.missingUsage, EstimatedUsageRecords: total.estimated, PendingTasks: total.pending}, err
}

func upstreamBillingPeriod(month, startDate, endDate string) (int64, int64, error) {
	if month != "" {
		if startDate != "" || endDate != "" {
			return 0, 0, errors.New("choose a month or a date range")
		}
		return model.BillingMonthBounds(month)
	}
	start, err := time.ParseInLocation("2006-01-02", startDate, billingLocation)
	if err != nil {
		return 0, 0, errors.New("invalid upstream billing start date")
	}
	end, err := time.ParseInLocation("2006-01-02", endDate, billingLocation)
	if err != nil {
		return 0, 0, errors.New("invalid upstream billing end date")
	}
	end = end.AddDate(0, 0, 1)
	if start.Unix() < 0 || !end.After(start) || end.Unix()-start.Unix() > 366*86400 {
		return 0, 0, errors.New("upstream billing date range must cover 1 to 366 days")
	}
	return start.Unix(), end.Unix(), nil
}

func GetUpstreamBillingReport(ctx context.Context, groupID int, month, startDate, endDate string) (*UpstreamBillingReport, error) {
	report, err := buildUpstreamBillingReport(ctx, groupID, month, startDate, endDate, "")
	return report, err
}

func buildUpstreamBillingReport(ctx context.Context, groupID int, month, startDate, endDate, statementID string) (*UpstreamBillingReport, error) {
	start, end, err := upstreamBillingPeriod(month, startDate, endDate)
	if err != nil {
		return nil, err
	}
	queryCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	group, err := model.GetUpstreamBillingGroup(queryCtx, groupID)
	if err != nil {
		return nil, err
	}
	currency, err := currentBillingCurrency(statementID != "")
	if err != nil {
		return nil, err
	}
	report := &UpstreamBillingReport{SchemaVersion: 2, AccountingBasis: "recorded_supplier_cost", Timezone: "Asia/Shanghai", Month: month, StartAt: start, EndAt: end, Group: *group, Currency: currency, UpdatedAt: common.GetTimestamp(), Coverage: UpstreamBillingCoverage{ExcludedChannels: make([]model.UpstreamBillingChannelView, 0)}, Days: make([]UpstreamBillingRow, 0), Models: make([]UpstreamBillingModelRow, 0), DayModels: make([]UpstreamBillingDayModelRow, 0)}
	report.StartDate, report.EndDate = time.Unix(start, 0).In(billingLocation).Format("2006-01-02"), time.Unix(end-1, 0).In(billingLocation).Format("2006-01-02")
	ids := make([]int, 0, len(group.Channels))
	for _, channel := range group.Channels {
		if channel.Missing {
			report.Coverage.MissingChannels++
		}
		if channel.Eligible {
			ids = append(ids, channel.ID)
		} else {
			report.Coverage.ExcludedChannels = append(report.Coverage.ExcludedChannels, channel)
		}
	}
	all := upstreamBillingAggregate{}
	allocation := upstreamBillingAllocation{}
	byDay := make(map[string]*upstreamBillingAggregate)
	byModel := make(map[upstreamBillingModelKey]*upstreamBillingAggregate)
	type dayModelKey struct {
		day   string
		model upstreamBillingModelKey
	}
	byDayModel := make(map[dayModelKey]*upstreamBillingAggregate)
	records := make([]model.UpstreamBillingStatementRecord, 0, 1000)
	digest := sha256.New()
	var sequence int64
	if len(ids) > 0 {
		err = model.VisitUpstreamBillingSourceRecords(queryCtx, ids, start, end, func(record model.UpstreamBillingStatementRecord) error {
			sequence++
			record.Sequence = sequence
			body, err := common.Marshal(record)
			if err != nil {
				return err
			}
			_, _ = digest.Write(append(body, '\n'))
			if statementID != "" {
				record.StatementID = statementID
				record.CreatedAt = common.GetTimestamp()
				records = append(records, record)
				if len(records) == 1000 {
					if err := model.DB.WithContext(queryCtx).CreateInBatches(&records, 500).Error; err != nil {
						return err
					}
					records = records[:0]
				}
			}
			if err := all.Add(record); err != nil {
				return err
			}
			if err := allocation.Add(record); err != nil {
				return err
			}
			day := time.Unix(record.OccurredAt, 0).In(billingLocation).Format("2006-01-02")
			if byDay[day] == nil {
				byDay[day] = &upstreamBillingAggregate{}
			}
			if err := byDay[day].Add(record); err != nil {
				return err
			}
			key := upstreamBillingModelKey{record.ModelName, record.UpstreamModelName, record.BillingGroup}
			if byModel[key] == nil {
				byModel[key] = &upstreamBillingAggregate{}
			}
			if err := byModel[key].Add(record); err != nil {
				return err
			}
			detail := dayModelKey{day, key}
			if byDayModel[detail] == nil {
				byDayModel[detail] = &upstreamBillingAggregate{}
			}
			return byDayModel[detail].Add(record)
		})
		if err != nil {
			return nil, err
		}
	}
	if len(records) > 0 {
		if err := model.DB.WithContext(queryCtx).CreateInBatches(&records, 500).Error; err != nil {
			return nil, err
		}
	}
	report.SourceSHA256 = hex.EncodeToString(digest.Sum(nil))
	report.ChannelUsers, err = allocation.Rows(queryCtx, *group, currency)
	if err != nil {
		return nil, err
	}
	report.Total, err = all.Row("total", currency)
	if err != nil {
		return nil, err
	}
	for at := time.Unix(start, 0).In(billingLocation); at.Unix() < end; at = at.AddDate(0, 0, 1) {
		label := at.Format("2006-01-02")
		total := byDay[label]
		if total == nil {
			total = &upstreamBillingAggregate{}
		}
		row, err := total.Row(label, currency)
		if err != nil {
			return nil, err
		}
		if at.Unix() > report.UpdatedAt {
			row.State = "future"
		} else if at.AddDate(0, 0, 1).Unix() > report.UpdatedAt {
			row.State = "in_progress"
		} else {
			row.State = "complete"
		}
		report.Days = append(report.Days, row)
	}
	for key, total := range byModel {
		row, err := total.Row(key.model, currency)
		if err != nil {
			return nil, err
		}
		report.Models = append(report.Models, UpstreamBillingModelRow{ModelName: key.model, UpstreamModelName: key.upstream, BillingGroup: key.group, UpstreamBillingRow: row, ActiveDays: int64(len(total.days)), FirstPosted: total.first, LastPosted: total.last})
	}
	for key, total := range byDayModel {
		row, err := total.Row(key.model.model, currency)
		if err != nil {
			return nil, err
		}
		report.DayModels = append(report.DayModels, UpstreamBillingDayModelRow{Date: key.day, UpstreamBillingModelRow: UpstreamBillingModelRow{ModelName: key.model.model, UpstreamModelName: key.model.upstream, BillingGroup: key.model.group, UpstreamBillingRow: row, ActiveDays: 1, FirstPosted: total.first, LastPosted: total.last}})
	}
	sort.Slice(report.Models, func(i, j int) bool { return upstreamBillingModelLess(report.Models[i], report.Models[j]) })
	sort.Slice(report.DayModels, func(i, j int) bool {
		if report.DayModels[i].Date != report.DayModels[j].Date {
			return report.DayModels[i].Date < report.DayModels[j].Date
		}
		return upstreamBillingModelLess(report.DayModels[i].UpstreamBillingModelRow, report.DayModels[j].UpstreamBillingModelRow)
	})
	report.Coverage.MissingCostRecords, report.Coverage.MissingUsageRecords, report.Coverage.EstimatedUsageRecords, report.Coverage.PendingTasks = all.missingCost, all.missingUsage, all.estimated, all.pending
	report.Coverage.Complete = len(ids) > 0 && report.Coverage.MissingChannels == 0 && all.missingCost == 0 && all.missingUsage == 0 && all.estimated == 0 && all.pending == 0
	billingDays := make([]BillingRow, 0, len(report.Days))
	for _, day := range report.Days {
		billingDays = append(billingDays, day.BillingRow)
	}
	report.RoundingDifference, err = billingRoundingDifference(billingDays, report.Total.BillingRow)
	return report, err
}

func upstreamBillingModelLess(left, right UpstreamBillingModelRow) bool {
	if left.ModelName != right.ModelName {
		return left.ModelName < right.ModelName
	}
	if left.UpstreamModelName != right.UpstreamModelName {
		return left.UpstreamModelName < right.UpstreamModelName
	}
	return left.BillingGroup < right.BillingGroup
}

func GetUpstreamBillingUsageDetails(ctx context.Context, groupID int, month, startDate, endDate, day, modelName string, cursor int64) (*UpstreamBillingUsageDetails, error) {
	if cursor < 0 || len(modelName) > 255 {
		return nil, errors.New("invalid upstream usage detail cursor or model")
	}
	start, end, err := upstreamBillingPeriod(month, startDate, endDate)
	if err != nil {
		return nil, err
	}
	if day != "" {
		at, err := time.ParseInLocation("2006-01-02", day, billingLocation)
		if err != nil || at.Unix() < start || at.Unix() >= end {
			return nil, errors.New("detail date is outside the selected billing period")
		}
		start, end = at.Unix(), at.AddDate(0, 0, 1).Unix()
	}
	queryCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	group, err := model.GetUpstreamBillingGroup(queryCtx, groupID)
	if err != nil {
		return nil, err
	}
	currency, err := currentBillingCurrency(false)
	if err != nil {
		return nil, err
	}
	view := &UpstreamBillingUsageDetails{Currency: currency, Records: make([]UpstreamBillingUsageRecordView, 0)}
	ids := make([]int, 0, len(group.Channels))
	for _, channel := range group.Channels {
		if channel.Eligible {
			ids = append(ids, channel.ID)
		}
	}
	if len(ids) == 0 {
		return view, nil
	}
	var sequence int64
	pageComplete := errors.New("upstream detail page complete")
	err = model.VisitUpstreamBillingSourceRecords(queryCtx, ids, start, end, func(record model.UpstreamBillingStatementRecord) error {
		sequence++
		record.Sequence = sequence
		if sequence <= cursor || (modelName != "" && record.ModelName != modelName) {
			return nil
		}
		if len(view.Records) == 100 {
			view.NextCursor = view.Records[len(view.Records)-1].Sequence
			return pageComplete
		}
		charge, refund := record.CostQuota, int64(0)
		if charge < 0 {
			charge, refund = 0, -charge
		}
		amount, err := billingRow("", charge, refund, 1, currency)
		if err != nil {
			return err
		}
		value := amount.Amount
		if record.CostMissing {
			value = ""
		}
		view.Records = append(view.Records, UpstreamBillingUsageRecordView{UpstreamBillingStatementRecord: record, Amount: value})
		return nil
	})
	if errors.Is(err, pageComplete) {
		err = nil
	}
	return view, err
}

func PrepareUpstreamBillingStatement(ctx context.Context, groupID, actorID, profileID int, month string) (*model.UpstreamBillingStatement, error) {
	start, end, err := model.BillingMonthBounds(month)
	if err != nil {
		return nil, err
	}
	if end > common.GetTimestamp()-86400 {
		return nil, ErrUpstreamBillingMonthOpen
	}
	if profileID <= 0 || actorID <= 0 {
		return nil, errors.New("upstream billing archive storage profile and actor are required")
	}
	statementID := common.GetUUID()
	published := false
	defer func() {
		if !published {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := model.DB.WithContext(cleanupCtx).Where("statement_id = ? AND NOT EXISTS (SELECT 1 FROM upstream_billing_statements WHERE id = ?)", statementID, statementID).Delete(&model.UpstreamBillingStatementRecord{}).Error; err != nil {
				common.SysError("failed to remove unpublished upstream staging rows: " + err.Error())
			}
		}
	}()
	report, err := buildUpstreamBillingReport(ctx, groupID, month, "", "", statementID)
	if err != nil {
		return nil, err
	}
	branding, err := captureBillingDocumentBranding(ctx)
	if err != nil {
		return nil, err
	}
	report.Issuer, report.PDFLogoPNG, report.PDFOperatingLogoPNG, report.OperatingName, report.PDFFooter = branding.Issuer, branding.LogoPNG, branding.OperatingLogoPNG, branding.OperatingName, branding.Footer
	body, err := common.Marshal(report)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(body)
	statement := &model.UpstreamBillingStatement{ID: statementID, GroupID: groupID, Month: month, StartAt: start, EndAt: end, GroupVersion: report.Group.Version, Snapshot: model.UpstreamBillingSnapshotData(body), SnapshotSHA256: hex.EncodeToString(digest[:]), SourceSHA256: report.SourceSHA256, RecordCount: report.Total.Count, StorageProfileID: profileID, CreatedBy: actorID}
	if err := model.CreateUpstreamBillingStatement(ctx, statement, nil); err != nil {
		return nil, err
	}
	published = true
	return statement, nil
}
