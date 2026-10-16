package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func upstreamBillingModelFixture(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	oldDB, oldLog := DB, LOG_DB
	DB, LOG_DB = db, db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	initCol()
	require.NoError(t, db.AutoMigrate(&Channel{}, &UpstreamBillingGroup{}, &UpstreamBillingGroupChannel{}, &UpstreamBillingGroupEvent{}, &UpstreamBillingStatement{}, &UpstreamBillingStatementEvent{}, &UpstreamBillingStatementRecord{}, &BillingArtifact{}, &UpstreamBillingAttachment{}, &StorageObject{}))
	t.Cleanup(func() { DB, LOG_DB = oldDB, oldLog; require.NoError(t, sqlDB.Close()) })
}

func TestUpstreamBillingGroupsPreserveRoutingAndExclusiveAuditedMembership(t *testing.T) {
	upstreamBillingModelFixture(t)
	tag := "supplier-blue"
	require.NoError(t, DB.Create(&[]Channel{{Id: 19, Name: "blue one", Key: "private", Group: "routing-a", Tag: &tag, CostDiscount: "0.8"}, {Id: 26, Name: "blue two", Key: "private", Group: "routing-b", CostDiscount: "0"}, {Id: 27, Name: "test channel", Key: "private"}}).Error)
	ctx := context.Background()
	group, err := SaveUpstreamBillingGroup(ctx, 0, 0, "Blue supplier", []int{19, 26, 27}, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), group.Version)
	require.Len(t, group.Channels, 3)
	assert.True(t, group.Channels[1].Eligible, "an explicit zero discount is configured")
	assert.False(t, group.Channels[2].Eligible)
	_, err = SaveUpstreamBillingGroup(ctx, 0, 0, "Second supplier", []int{19}, 2)
	assert.ErrorIs(t, err, ErrBillingConflict)
	_, err = SaveUpstreamBillingGroup(ctx, group.ID, 0, "Stale change", []int{26}, 2)
	assert.ErrorIs(t, err, ErrBillingConflict)
	saved, err := SaveUpstreamBillingGroup(ctx, group.ID, 1, "Blue renamed", []int{19, 26}, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), saved.Version)
	var channel Channel
	require.NoError(t, DB.First(&channel, 19).Error)
	assert.Equal(t, "routing-a", channel.Group)
	assert.Equal(t, tag, channel.GetTag())
	search, err := SearchUpstreamBillingChannels(ctx, tag)
	require.NoError(t, err)
	require.Len(t, search, 1)
	assert.Equal(t, saved.ID, search[0].GroupID)
	var events []UpstreamBillingGroupEvent
	require.NoError(t, DB.Order("id asc").Find(&events).Error)
	require.Len(t, events, 2)
	assert.Equal(t, 2, events[1].ActorID)
	assert.Contains(t, events[1].Snapshot, "Blue renamed")
	assert.Contains(t, events[1].Snapshot, `"channel_id":26`)
	require.NoError(t, DB.Delete(&Channel{}, 26).Error)
	missing, err := GetUpstreamBillingGroup(ctx, group.ID)
	require.NoError(t, err)
	assert.True(t, missing.Channels[1].Missing)
	assert.False(t, missing.Channels[1].Eligible)
}

func TestUpstreamBillingStatementsRequireArchiveHashesAndRealSessionThenRemainImmutable(t *testing.T) {
	upstreamBillingModelFixture(t)
	require.NoError(t, DB.Create(&Channel{Id: 19, Key: "private", CostDiscount: "0.8"}).Error)
	group, err := SaveUpstreamBillingGroup(context.Background(), 0, 0, "Supplier", []int{19}, 1)
	require.NoError(t, err)
	start, end, err := BillingMonthBounds("2026-08")
	require.NoError(t, err)
	body, err := common.Marshal(struct {
		Group UpstreamBillingGroupView `json:"group"`
	}{*group})
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &UpstreamBillingStatement{ID: common.GetUUID(), GroupID: group.ID, GroupVersion: group.Version, Month: "2026-08", StartAt: start, EndAt: end, Snapshot: UpstreamBillingSnapshotData(body), SnapshotSHA256: hex.EncodeToString(hash[:]), SourceSHA256: hex.EncodeToString(hash[:]), CreatedBy: 1, StorageProfileID: 1}
	require.NoError(t, CreateUpstreamBillingStatement(context.Background(), statement, nil))
	assert.Equal(t, StatementPreparing, statement.Status)
	require.NoError(t, DB.Model(statement).Update("lease_owner", "worker").Error)
	statement.ManifestSHA256, statement.PDFSHA256, statement.ExcelSHA256 = statement.SnapshotSHA256, statement.SnapshotSHA256, statement.SnapshotSHA256
	assert.ErrorIs(t, CompleteUpstreamBillingStatementArchive(statement, "stale-worker", true), ErrBillingConflict)
	require.NoError(t, CompleteUpstreamBillingStatementArchive(statement, "worker", true))
	current, err := GetUpstreamBillingStatement(statement.ID)
	require.NoError(t, err)
	assert.Equal(t, statement.ManifestSHA256, current.ManifestSHA256, "GORM must persist acronym fields")
	_, err = ActUpstreamBillingStatement(statement.ID, "confirm", statement.ManifestSHA256, "", "", 2)
	assert.ErrorIs(t, err, ErrBillingConflict)
	_, err = ActUpstreamBillingStatement(statement.ID, "confirm", "stale-digest", "", "real-session", 2)
	assert.ErrorIs(t, err, ErrBillingConflict)
	for _, kind := range []string{"snapshot", "pdf", "xlsx", "manifest"} {
		require.NoError(t, DB.Create(&BillingArtifact{StatementID: statement.ID, Kind: kind, SHA256: statement.SnapshotSHA256, Size: 10}).Error)
	}
	review, err := GetUpstreamBillingReview(statement.ID)
	require.NoError(t, err)
	confirmed, err := ActUpstreamBillingStatementWithReview(statement.ID, "confirm", statement.ManifestSHA256, review.EvidenceSHA256, "", "real-session", 2)
	require.NoError(t, err)
	assert.Equal(t, StatementConfirmed, confirmed.Status)
	assert.Equal(t, 2, confirmed.ConfirmedBy)
	assert.Equal(t, confirmed.ConfirmedAt, confirmed.ArchivedAt)
	repeated, err := ActUpstreamBillingStatementWithReview(statement.ID, "confirm", statement.ManifestSHA256, review.EvidenceSHA256, "", "real-session", 2)
	require.NoError(t, err)
	assert.Equal(t, confirmed.ConfirmedAt, repeated.ConfirmedAt)
	_, err = ActUpstreamBillingStatement(statement.ID, "void", "", "Please replace", "real-session", 2)
	assert.ErrorIs(t, err, ErrBillingConflict)
	var confirms int64
	require.NoError(t, DB.Model(&UpstreamBillingStatementEvent{}).Where("statement_id = ? AND action = ?", statement.ID, "confirm").Count(&confirms).Error)
	assert.Equal(t, int64(1), confirms)
	duplicate := *statement
	duplicate.ID = common.GetUUID()
	assert.ErrorIs(t, CreateUpstreamBillingStatement(context.Background(), &duplicate, nil), ErrBillingConflict)
	assert.Equal(t, statement.Snapshot, confirmed.Snapshot)
}

func TestUpstreamBillingStatementFreezesQualityFlagsInDirectAndStagedRecords(t *testing.T) {
	for _, staged := range []bool{false, true} {
		t.Run(map[bool]string{false: "direct records", true: "staged records"}[staged], func(t *testing.T) {
			upstreamBillingModelFixture(t)
			require.NoError(t, DB.Create(&Channel{Id: 19, Key: "private", CostDiscount: "0.8"}).Error)
			group, err := SaveUpstreamBillingGroup(context.Background(), 0, 0, "Supplier", []int{19}, 1)
			require.NoError(t, err)
			start, end, err := BillingMonthBounds("2026-08")
			require.NoError(t, err)
			body, err := common.Marshal(struct {
				Group UpstreamBillingGroupView `json:"group"`
			}{*group})
			require.NoError(t, err)
			hash := sha256.Sum256(body)
			statement := &UpstreamBillingStatement{ID: common.GetUUID(), GroupID: group.ID, GroupVersion: group.Version, Month: "2026-08", StartAt: start, EndAt: end, Snapshot: UpstreamBillingSnapshotData(body), SnapshotSHA256: hex.EncodeToString(hash[:]), CreatedBy: 1, StorageProfileID: 1, RecordCount: 1}
			records := []UpstreamBillingStatementRecord{{StatementID: statement.ID, Sequence: 1, ChannelID: 19, OccurredAt: start, LogType: LogTypeConsume, CostMissing: true, UsageMissing: true, UsageEstimated: true, TaskPending: true}}
			if staged {
				require.NoError(t, DB.Create(&records).Error)
				records = nil
			}
			require.NoError(t, CreateUpstreamBillingStatement(context.Background(), statement, records))
			frozen, err := GetUpstreamBillingStatementRecords(statement, 0, 100)
			require.NoError(t, err)
			require.Len(t, frozen, 1)
			assert.True(t, frozen[0].CostMissing)
			assert.True(t, frozen[0].UsageMissing)
			assert.True(t, frozen[0].UsageEstimated)
			assert.True(t, frozen[0].TaskPending)
		})
	}
}

func TestUpstreamBillingStatementFreezesKnownMissingChannelButRejectsConcurrentDeletion(t *testing.T) {
	for _, missingBeforeRead := range []bool{false, true} {
		t.Run(map[bool]string{false: "deleted after report", true: "missing in frozen report"}[missingBeforeRead], func(t *testing.T) {
			upstreamBillingModelFixture(t)
			require.NoError(t, DB.Create(&Channel{Id: 19, Key: "private", CostDiscount: "0.8"}).Error)
			group, err := SaveUpstreamBillingGroup(context.Background(), 0, 0, "Supplier", []int{19}, 1)
			require.NoError(t, err)
			if missingBeforeRead {
				require.NoError(t, DB.Delete(&Channel{}, 19).Error)
				group, err = GetUpstreamBillingGroup(context.Background(), group.ID)
				require.NoError(t, err)
			}
			body, err := common.Marshal(struct {
				Group UpstreamBillingGroupView `json:"group"`
			}{*group})
			require.NoError(t, err)
			if !missingBeforeRead {
				require.NoError(t, DB.Delete(&Channel{}, 19).Error)
			}
			start, end, err := BillingMonthBounds("2026-08")
			require.NoError(t, err)
			hash := sha256.Sum256(body)
			statement := &UpstreamBillingStatement{ID: common.GetUUID(), GroupID: group.ID, GroupVersion: group.Version, Month: "2026-08", StartAt: start, EndAt: end, Snapshot: UpstreamBillingSnapshotData(body), SnapshotSHA256: hex.EncodeToString(hash[:]), CreatedBy: 1, StorageProfileID: 1}
			err = CreateUpstreamBillingStatement(context.Background(), statement, nil)
			if missingBeforeRead {
				require.NoError(t, err)
				assert.Equal(t, StatementPreparing, statement.Status)
			} else {
				assert.ErrorIs(t, err, ErrBillingConflict)
			}
		})
	}
}

func TestUpstreamBillingLargeFrozenSnapshotRoundTripsAndUsesPortableDatabaseTextStorage(t *testing.T) {
	upstreamBillingModelFixture(t)
	body, err := common.Marshal(map[string]any{"model_days": strings.Repeat("frozen-model-summary;", 5000)})
	require.NoError(t, err)
	require.Greater(t, len(body), 64*1024)
	hash := sha256.Sum256(body)
	statement := UpstreamBillingStatement{ID: "large-snapshot", Snapshot: UpstreamBillingSnapshotData(body), SnapshotSHA256: hex.EncodeToString(hash[:])}
	require.NoError(t, DB.Create(&statement).Error)
	stored, err := GetUpstreamBillingStatement(statement.ID)
	require.NoError(t, err)
	assert.Equal(t, statement.Snapshot, stored.Snapshot)
	assert.Equal(t, statement.SnapshotSHA256, stored.SnapshotSHA256)
	encoded, err := common.Marshal(stored)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"snapshot":"{`, "the API remains a JSON string")
	for _, test := range []struct {
		name, want string
		dialector  gorm.Dialector
	}{
		{name: "sqlite", want: "TEXT", dialector: sqlite.Open(":memory:")},
		{name: "mysql", want: "LONGTEXT", dialector: mysql.New(mysql.Config{DSN: "user:pass@tcp(127.0.0.1:1)/fixture", SkipInitializeWithVersion: true})},
		{name: "postgres", want: "TEXT", dialector: postgres.New(postgres.Config{DSN: "host=127.0.0.1 port=1 user=fixture dbname=fixture sslmode=disable"})},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, err := gorm.Open(test.dialector, &gorm.Config{DryRun: true, DisableAutomaticPing: true})
			require.NoError(t, err)
			statement := &gorm.Statement{DB: db}
			require.NoError(t, statement.Parse(&UpstreamBillingStatement{}))
			field := statement.Schema.LookUpField("Snapshot")
			require.NotNil(t, field)
			assert.Equal(t, test.want, db.Migrator().FullDataTypeOf(field).SQL, "schema generation must not impose a MySQL 64 KiB limit or a nonportable longtext type")
			sqlDB, err := db.DB()
			require.NoError(t, err)
			require.NoError(t, sqlDB.Close())
		})
	}
}

func TestUpstreamBillingSourceWatermarksExcludeConcurrentBackfillAndCostChanges(t *testing.T) {
	upstreamBillingModelFixture(t)
	require.NoError(t, DB.AutoMigrate(&Log{}, &Task{}, &BillingEntry{}, &BillingCorrectionRow{}, &CostAccountingSnapshot{}, &CostAccountingAdjustment{}))
	start, end, err := BillingMonthBounds("2026-08")
	require.NoError(t, err)
	log := Log{UserId: 1, Type: LogTypeConsume, CreatedAt: start, ChannelId: 19, ModelName: "text", RequestId: "initial", Quota: 100, PromptTokens: 10}
	require.NoError(t, LOG_DB.Create(&log).Error)
	initial, err := BuildCostAccountingSnapshot(&log, &CostAccountingInput{CostBasisQuota: "100", CostDiscount: "0.8", CostQuota: 80})
	require.NoError(t, err)
	require.NoError(t, DB.Create(initial).Error)
	orphan, err := BuildCostAccountingSnapshot(&Log{Id: 999, UserId: 1, Type: LogTypeConsume, CreatedAt: start + 1, ChannelId: 19, ModelName: "orphan", RequestId: "orphan", Quota: 200}, &CostAccountingInput{CostBasisQuota: "200", CostDiscount: "0.8", CostQuota: 160})
	require.NoError(t, err)
	require.NoError(t, DB.Create(orphan).Error)
	var records []UpstreamBillingStatementRecord
	err = VisitUpstreamBillingSourceRecords(context.Background(), []int{19}, start, end, func(record UpstreamBillingStatementRecord) error {
		records = append(records, record)
		if len(records) != 1 {
			return nil
		}
		// This callback supplies a deterministic boundary between source pages,
		// replacing timing/sleep based race tests.
		require.NoError(t, DB.Create(&CostAccountingAdjustment{SnapshotID: orphan.ID, DeltaCostQuota: -20, NewCostQuota: 140}).Error)
		late := Log{UserId: 1, Type: LogTypeConsume, CreatedAt: start + 2, ChannelId: 19, ModelName: "late-backfill", RequestId: "late", Quota: 100, PromptTokens: 10}
		require.NoError(t, LOG_DB.Create(&late).Error)
		lateSnapshot, err := BuildCostAccountingSnapshot(&late, &CostAccountingInput{CostBasisQuota: "100", CostDiscount: "0.8", CostQuota: 80})
		require.NoError(t, err)
		require.NoError(t, DB.Create(lateSnapshot).Error)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "text", records[0].ModelName)
	assert.Equal(t, "orphan", records[1].ModelName)
	assert.Equal(t, int64(160), records[1].CostQuota)
	assert.True(t, records[1].UsageMissing)
	assert.Zero(t, records[1].LatestAdjustmentID)
}

func TestUpstreamBillingCleanupRemovesOnlyOldUnpublishedRowsAndRechecksPublication(t *testing.T) {
	upstreamBillingModelFixture(t)
	now, cutoff := common.GetTimestamp(), common.GetTimestamp()-86400
	records := []UpstreamBillingStatementRecord{
		{StatementID: "old-abandoned", Sequence: 1, CreatedAt: cutoff - 1},
		{StatementID: "recent-collecting", Sequence: 1, CreatedAt: now - 1},
		{StatementID: "legacy-no-time", Sequence: 1},
		{StatementID: "published-during-delete", Sequence: 1, CreatedAt: cutoff - 1},
	}
	for i, status := range []string{StatementConfirmed, StatementFailed, StatementVoid} {
		id := "published-" + status
		require.NoError(t, DB.Create(&UpstreamBillingStatement{ID: id, GroupID: i + 1, Month: "2026-08", Revision: 1, Status: status}).Error)
		records = append(records, UpstreamBillingStatementRecord{StatementID: id, Sequence: 1, CreatedAt: cutoff - 1})
	}
	require.NoError(t, DB.Create(&records).Error)
	// Force publication after the selector has captured IDs but immediately
	// before DELETE executes. No sleeps or real concurrent timing are needed.
	published := false
	require.NoError(t, DB.Callback().Delete().Before("gorm:delete").Register("upstream_cleanup_publication", func(tx *gorm.DB) {
		if published || tx.Statement.Schema == nil || tx.Statement.Schema.Table != "upstream_billing_statement_records" {
			return
		}
		published = true
		err := tx.Session(&gorm.Session{NewDB: true}).Create(&UpstreamBillingStatement{ID: "published-during-delete", GroupID: 10, Month: "2026-08", Revision: 1, Status: StatementPreparing}).Error
		tx.AddError(err)
	}))
	t.Cleanup(func() { require.NoError(t, DB.Callback().Delete().Remove("upstream_cleanup_publication")) })
	removed, err := PurgeUnpublishedUpstreamBillingRecords(context.Background(), cutoff, 1000)
	require.NoError(t, err)
	assert.Equal(t, int64(1), removed)
	assert.True(t, published)
	var remaining []UpstreamBillingStatementRecord
	require.NoError(t, DB.Order("id asc").Find(&remaining).Error)
	identities := make([]string, 0, len(remaining))
	for _, record := range remaining {
		identities = append(identities, record.StatementID)
	}
	assert.NotContains(t, identities, "old-abandoned")
	assert.Contains(t, identities, "recent-collecting")
	assert.Contains(t, identities, "legacy-no-time")
	assert.Contains(t, identities, "published-during-delete")
	for _, status := range []string{StatementConfirmed, StatementFailed, StatementVoid} {
		assert.Contains(t, identities, "published-"+status)
	}
	_, err = PurgeUnpublishedUpstreamBillingRecords(context.Background(), now, 1000)
	assert.Error(t, err, "recent collection must not be deleted by a too-new cutoff")
}
