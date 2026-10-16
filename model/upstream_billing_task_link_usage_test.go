package model

import (
	"context"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUpstreamBillingHistoricalLinksRestoreOnlyReviewedUsageAndPreserveFrozenRows(t *testing.T) {
	for _, failed := range []bool{false, true} {
		name := "reported successful usage"
		if failed {
			name = "known full supplier refund"
		}
		t.Run(name, func(t *testing.T) {
			fixture := upstreamTaskLinkReviewFixture(t, failed)
			require.NoError(t, DB.AutoMigrate(&BillingEntry{}, &BillingCorrectionRow{}, &CostAccountingSnapshot{}, &CostAccountingAdjustment{}))
			start, end, err := BillingMonthBounds("2026-09")
			require.NoError(t, err)
			for _, log := range []*Log{&fixture.source, &fixture.settlement} {
				basis, cost := "100", int64(80)
				if log.Type == LogTypeRefund {
					basis = "-60"
					cost = -48
					if failed {
						basis = "-100"
						cost = -80
					}
				}
				snapshot, err := BuildCostAccountingSnapshot(log, &CostAccountingInput{CostBasisQuota: basis, CostDiscount: "0.8", CostQuota: cost})
				require.NoError(t, err)
				require.NoError(t, DB.Create(snapshot).Error)
			}
			var before []UpstreamBillingStatementRecord
			require.NoError(t, VisitUpstreamBillingSourceRecords(context.Background(), []int{19}, start, end, func(record UpstreamBillingStatementRecord) error { before = append(before, record); return nil }))
			require.Len(t, before, 2)
			assert.True(t, before[0].UsageMissing)
			assert.Equal(t, "task_missing_id", before[0].UsageSource)
			// Preserve previously collected evidence exactly; associations affect
			// subsequent live visits, never an existing staged source hash.
			for i := range before {
				before[i].StatementID = "before-review"
				before[i].Sequence = int64(i + 1)
			}
			require.NoError(t, DB.Create(&before).Error)
			frozenJSON, err := common.Marshal(before)
			require.NoError(t, err)
			raw, digest := upstreamTaskLinkManifestBytes(t, fixture.manifest)
			links, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "historical-review", 0)
			require.NoError(t, err)
			require.Len(t, links, 1)
			var after []UpstreamBillingStatementRecord
			require.NoError(t, VisitUpstreamBillingSourceRecords(context.Background(), []int{19}, start, end, func(record UpstreamBillingStatementRecord) error { after = append(after, record); return nil }))
			require.Len(t, after, 2)
			assert.Equal(t, fixture.task.TaskID, after[0].TaskID)
			assert.False(t, after[0].UsageMissing)
			assert.Equal(t, int64(1), after[0].Requests)
			assert.Equal(t, int64(80), after[0].CostQuota)
			if failed {
				assert.Equal(t, "task_historical_failed_refunded", after[0].UsageSource)
				assert.Zero(t, after[0].VideoTokens)
				assert.Zero(t, after[0].VideoSeconds)
				assert.Equal(t, int64(-80), after[1].CostQuota)
			} else {
				assert.Equal(t, "task_historical_final_usage", after[0].UsageSource)
				assert.Equal(t, int64(123), after[0].VideoTokens)
				assert.Equal(t, int64(5), after[0].VideoSeconds)
			}
			assert.Zero(t, after[1].Requests)
			assert.Zero(t, after[1].VideoTokens, "settlement does not duplicate original usage")
			frozen, err := GetUpstreamBillingStatementRecords(&UpstreamBillingStatement{ID: "before-review"}, 0, 1000)
			require.NoError(t, err)
			afterJSON, err := common.Marshal(frozen)
			require.NoError(t, err)
			assert.Equal(t, string(frozenJSON), string(afterJSON))
			var original Log
			require.NoError(t, LOG_DB.First(&original, fixture.source.Id).Error)
			assert.Equal(t, fixture.source.Other, original.Other)
		})
	}
}

func TestUpstreamBillingHistoricalLinkReadsRejectChangedSourceAndEvidence(t *testing.T) {
	for _, change := range []string{"source metadata", "evidence body", "task properties", "task final usage"} {
		t.Run(change, func(t *testing.T) {
			fixture := upstreamTaskLinkReviewFixture(t, false)
			raw, digest := upstreamTaskLinkManifestBytes(t, fixture.manifest)
			_, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "historical-review", 0)
			require.NoError(t, err)
			switch change {
			case "source metadata":
				fixture.source.Content = "changed content"
				require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.source.Id).Update("content", fixture.source.Content).Error)
			case "evidence body":
				require.NoError(t, DB.Exec("UPDATE upstream_billing_task_links SET evidence = ? WHERE source_log_id = ?", "{}", fixture.source.Id).Error)
			case "task properties":
				require.NoError(t, DB.Model(&Task{}).Where("id = ?", fixture.task.ID).Update("properties", Properties{OriginModelName: "changed"}).Error)
			case "task final usage":
				require.NoError(t, DB.Model(&Task{}).Where("id = ?", fixture.task.ID).Update("data", []byte(`{"id":"review-provider-task","usage":{"total_tokens":999}}`)).Error)
			}
			_, _, err = upstreamBillingResolveTaskLogs(context.Background(), []Log{fixture.source})
			assert.ErrorIs(t, err, ErrBillingEvidenceIntegrity)
		})
	}
}

func TestUpstreamBillingHistoricalUsageCannotChangeBetweenResolverAndUsageRead(t *testing.T) {
	fixture := upstreamTaskLinkReviewFixture(t, false)
	require.NoError(t, DB.AutoMigrate(&BillingEntry{}, &BillingCorrectionRow{}, &CostAccountingSnapshot{}, &CostAccountingAdjustment{}))
	raw, digest := upstreamTaskLinkManifestBytes(t, fixture.manifest)
	_, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "historical-review", 0)
	require.NoError(t, err)
	changed := false
	require.NoError(t, DB.Callback().Query().After("gorm:query").Register("historical_usage_changed_after_validation", func(tx *gorm.DB) {
		if changed || tx.Statement.Table != "tasks" || !strings.Contains(strings.Join(tx.Statement.Selects, ","), "properties") {
			return
		}
		changed = true
		require.NoError(t, DB.Model(&Task{}).Where("id = ?", fixture.task.ID).Update("data", []byte(`{"id":"review-provider-task","usage":{"total_tokens":999}}`)).Error)
	}))
	defer func() { require.NoError(t, DB.Callback().Query().Remove("historical_usage_changed_after_validation")) }()
	start, end, err := BillingMonthBounds("2026-09")
	require.NoError(t, err)
	err = VisitUpstreamBillingSourceRecords(context.Background(), []int{19}, start, end, func(UpstreamBillingStatementRecord) error { return nil })
	assert.True(t, changed)
	assert.ErrorIs(t, err, ErrBillingEvidenceIntegrity)
}

func TestUpstreamBillingHistoricalLinksNeverUseUnstableClickHouseLogIDs(t *testing.T) {
	fixture := upstreamTaskLinkReviewFixture(t, false)
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeClickHouse)
	defer common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	raw, digest := upstreamTaskLinkManifestBytes(t, fixture.manifest)
	_, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "unsupported-clickhouse", 0)
	assert.ErrorIs(t, err, ErrBillingEvidenceIntegrity)
	resolved, links, err := upstreamBillingResolveTaskLogs(context.Background(), []Log{fixture.source})
	require.NoError(t, err)
	assert.Empty(t, links)
	assert.Equal(t, fixture.source.Other, resolved[0].Other)
}

func TestUpstreamBillingHistoricalLinkAllowsBookkeepingAfterPersistedFinishTime(t *testing.T) {
	for _, failed := range []bool{false, true} {
		name := "settlement"
		if failed {
			name = "failure refund"
		}
		t.Run(name, func(t *testing.T) {
			fixture := upstreamTaskLinkReviewFixture(t, failed)
			posted := fixture.task.FinishTime + 1
			require.NoError(t, LOG_DB.Model(&Log{}).Where("id = ?", fixture.settlement.Id).Update("created_at", posted).Error)
			fixture.candidate["adjustment_chain"].([]any)[0].(map[string]any)["occurred_at"] = posted
			raw, digest := upstreamTaskLinkManifestBytes(t, fixture.manifest)
			links, err := ApplyUpstreamBillingTaskLinks(context.Background(), raw, digest, "delayed-bookkeeping", 0)
			require.NoError(t, err)
			require.Len(t, links, 1)
		})
	}
}
