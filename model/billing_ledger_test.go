package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedBillingCustomer(t *testing.T) int {
	t.Helper()
	truncateTables(t)
	user := User{Id: 901, Username: "billing_customer", Quota: 10000, Status: common.UserStatusEnabled}
	require.NoError(t, DB.Create(&user).Error)
	now := int64(0)
	_, err := SaveBillingAccount(user.Id, 1, 0, "测试企业", "91310000TEST", &now)
	require.NoError(t, err)
	return user.Id
}

func TestBillingSettlementIsAtomicIdempotentAndNotAReservation(t *testing.T) {
	id := seedBillingCustomer(t)
	op := &BillingOperation{ID: "internal-operation", UserID: id, Reserved: 3000, RequestID: "customer-request", ModelName: "model"}
	require.NoError(t, BeginBillingOperation(op))
	var count int64
	require.NoError(t, DB.Model(&BillingEntry{}).Count(&count).Error)
	assert.Zero(t, count, "reservation must not be shown as consumption")
	require.NoError(t, FinishBillingOperation(op.ID, id, 3000, false))
	require.NoError(t, FinishBillingOperation(op.ID, id, 3000, false))
	quota, err := GetUserQuota(id, true)
	require.NoError(t, err)
	assert.Equal(t, 7000, quota)
	var entries []BillingEntry
	require.NoError(t, DB.Find(&entries).Error)
	require.Len(t, entries, 1)
	assert.Equal(t, int64(3000), entries[0].Quota)
	hours, err := GetBillingHours(id, 0, common.GetTimestamp()+3600)
	require.NoError(t, err)
	require.Len(t, hours, 1)
	assert.Equal(t, int64(3000), hours[0].Charge)
	assert.Equal(t, int64(1), hours[0].Count)
	assert.Equal(t, entries[0].Sequence, hours[0].FirstSequence)
	assert.Equal(t, entries[0].Sequence, hours[0].LastSequence)
	assert.ErrorIs(t, FinishBillingOperation(op.ID, id, 0, true), ErrBillingConflict, "settled consumption cannot be refunded as a reservation")
}

func TestBillingFailedRequestRefundsOnceAndNeverBecomesConsumption(t *testing.T) {
	id := seedBillingCustomer(t)
	op := &BillingOperation{ID: "failed-operation", UserID: id, Reserved: 2000}
	require.NoError(t, BeginBillingOperation(op))
	require.NoError(t, ResizeBillingReservation(op.ID, id, 4000))
	require.NoError(t, FinishBillingOperation(op.ID, id, 0, true))
	require.NoError(t, FinishBillingOperation(op.ID, id, 0, true))
	quota, err := GetUserQuota(id, true)
	require.NoError(t, err)
	assert.Equal(t, 10000, quota)
	var count int64
	require.NoError(t, DB.Model(&BillingEntry{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestBillingFundingIsNotConsumptionAndOverrideIsBounded(t *testing.T) {
	id := seedBillingCustomer(t)
	applied, err := PostBillingAdjustment(&BillingEntry{EventKey: "admin-credit", UserID: id, Kind: "funding", WalletDelta: 5000, ActorID: 1}, nil)
	require.NoError(t, err)
	assert.True(t, applied)
	applied, err = PostBillingAdjustment(&BillingEntry{EventKey: "admin-credit", UserID: id, Kind: "funding", WalletDelta: 5000, ActorID: 1}, nil)
	require.NoError(t, err)
	assert.False(t, applied)
	hours, err := GetBillingHours(id, 0, common.GetTimestamp()+3600)
	require.NoError(t, err)
	assert.Empty(t, hours)
	var entry BillingEntry
	require.NoError(t, DB.First(&entry, "event_key = ?", "admin-credit").Error)
	assert.Equal(t, int64(15000), entry.BalanceAfter)
	assert.Zero(t, entry.Quota)
	_, err = PostBillingAdjustment(&BillingEntry{EventKey: "overflow", UserID: id, Kind: "funding", WalletDelta: int64(common.MaxWalletQuota)}, nil)
	require.Error(t, err)
	quota, err := GetUserQuota(id, true)
	require.NoError(t, err)
	assert.Equal(t, 15000, quota)
}

func TestBillingAggregateFailureRollsBackBalanceAndSettlement(t *testing.T) {
	id := seedBillingCustomer(t)
	require.NoError(t, DB.Create(&BillingHour{UserID: id, Hour: common.GetTimestamp() / 3600 * 3600, Charge: int64(common.MaxWalletQuota)}).Error)
	op := &BillingOperation{ID: "overflow-operation", UserID: id, Reserved: 1000}
	require.NoError(t, BeginBillingOperation(op))
	require.Error(t, FinishBillingOperation(op.ID, id, 1200, false))
	quota, err := GetUserQuota(id, true)
	require.NoError(t, err)
	assert.Equal(t, 9000, quota)
	require.NoError(t, DB.First(op, "id = ?", op.ID).Error)
	assert.Equal(t, "reserved", op.State)
	var count int64
	require.NoError(t, DB.Model(&BillingEntry{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestBillingTaskRefundRejectsStalePoller(t *testing.T) {
	id := seedBillingCustomer(t)
	task := &Task{UserId: id, TaskID: "durable-task", Quota: 2000}
	require.NoError(t, DB.Create(task).Error)
	stale := *task
	applied, err := AdjustBillingTaskQuota(task, 0, "model")
	require.NoError(t, err)
	assert.True(t, applied)
	applied, err = AdjustBillingTaskQuota(&stale, 0, "model")
	require.NoError(t, err)
	assert.False(t, applied)
	quota, err := GetUserQuota(id, true)
	require.NoError(t, err)
	assert.Equal(t, 12000, quota)
	var entry BillingEntry
	require.NoError(t, DB.First(&entry).Error)
	assert.Equal(t, int64(-2000), entry.Quota)
}

func TestBillingConfirmationRequiresOwnerSessionAndFrozenDigest(t *testing.T) {
	id := seedBillingCustomer(t)
	digest := strings.Repeat("a", 64)
	statement := &BillingStatement{ID: "statement-1", UserID: id, Month: "2026-01", Revision: 1, Status: StatementDraft, ManifestSHA256: digest, PDFSHA256: strings.Repeat("b", 64)}
	hash := sha256.Sum256([]byte("{}"))
	statement.Snapshot, statement.SnapshotSHA256 = "{}", hex.EncodeToString(hash[:])
	require.NoError(t, DB.Create(statement).Error)
	_, err := ChangeBillingStatement(statement.ID, "issue", "", "", "", 1, true)
	require.NoError(t, err)
	for _, test := range []struct {
		actor         int
		session, hash string
		admin         bool
	}{{1, "admin", digest, true}, {id, "", digest, false}, {id, "session", "wrong", false}} {
		_, err := ChangeBillingStatement(statement.ID, "confirm", test.hash, "", test.session, test.actor, test.admin)
		require.Error(t, err)
	}
	confirmed, err := ChangeBillingStatement(statement.ID, "confirm", digest, "", "owner-session", id, false)
	require.NoError(t, err)
	assert.Positive(t, confirmed.ConfirmedAt)
	_, err = ChangeBillingStatement(statement.ID, "confirm", digest, "", "owner-session", id, false)
	require.NoError(t, err)
	var count int64
	require.NoError(t, DB.Model(&BillingStatementEvent{}).Where("action = ?", "confirm").Count(&count).Error)
	assert.Equal(t, int64(1), count)
	_, err = ChangeBillingStatement(statement.ID, "void", "", "replacement needed", "", 1, true)
	assert.ErrorIs(t, err, ErrBillingConflict)
}

func TestBillingStartAndCorporateProfileVersionAreAudited(t *testing.T) {
	id := seedBillingCustomer(t)
	account, err := GetBillingAccount(id)
	require.NoError(t, err)
	assert.Positive(t, account.AccountingStartAt)
	_, err = SaveBillingAccount(id, id, 0, "stale", "new", nil)
	assert.ErrorIs(t, err, ErrBillingConflict)
	updated, err := SaveBillingAccount(id, id, account.ProfileVersion, "更新企业", "NEW-TAX", nil)
	require.NoError(t, err)
	assert.Equal(t, account.AccountingStartAt, updated.AccountingStartAt)
	var events []BillingAccountEvent
	require.NoError(t, DB.Where("user_id = ?", id).Find(&events).Error)
	require.Len(t, events, 2)
	assert.Contains(t, events[0].Snapshot, "测试企业")
	assert.Contains(t, events[1].Snapshot, "更新企业")
	past := int64(1)
	_, err = SaveBillingAccount(id, 1, updated.ProfileVersion, "更新企业", "NEW-TAX", &past)
	require.Error(t, err)
}

func TestUsageEvidenceCannotBeDeletedOrDisabled(t *testing.T) {
	truncateTables(t)
	originalOptions := common.OptionMap
	originalEnabled := common.LogConsumeEnabled
	common.OptionMap = make(map[string]string)
	t.Cleanup(func() { common.OptionMap = originalOptions; common.LogConsumeEnabled = originalEnabled })
	require.NoError(t, DB.Create(&Log{Type: LogTypeConsume, CreatedAt: 1}).Error)
	deleted, err := DeleteOldLogBatch(context.Background(), common.GetTimestamp(), 100)
	assert.Zero(t, deleted)
	assert.ErrorIs(t, err, ErrUsageLogsRetained)
	assert.ErrorIs(t, UpdateOption("LogConsumeEnabled", "false"), ErrUsageLogsRetained)
	require.NoError(t, updateOptionMap("LogConsumeEnabled", "false"))
	assert.True(t, common.LogConsumeEnabled)
	var count int64
	require.NoError(t, DB.Model(&Log{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestBillingEvidenceCannotBeUpdatedOrDeleted(t *testing.T) {
	id := seedBillingCustomer(t)
	_, err := PostBillingAdjustment(&BillingEntry{EventKey: "protected", UserID: id, Kind: "usage", Quota: 100, WalletDelta: -100}, nil)
	require.NoError(t, err)
	var entry BillingEntry
	require.NoError(t, DB.First(&entry).Error)
	assert.ErrorIs(t, DB.Model(&entry).Update("quota", 500).Error, ErrBillingEvidenceImmutable)
	assert.ErrorIs(t, DB.Delete(&entry).Error, ErrBillingEvidenceImmutable)
	var preserved BillingEntry
	require.NoError(t, DB.First(&preserved, entry.ID).Error)
	assert.Equal(t, int64(100), preserved.Quota)
}

func TestBillingStatementBlocksOpenOperationsAndRetainsVoidedVersions(t *testing.T) {
	id := seedBillingCustomer(t)
	start, end, err := BillingMonthBounds("2020-02")
	require.NoError(t, err)
	require.NoError(t, DB.Model(&BillingAccount{}).Where("user_id = ?", id).Update("accounting_start_at", start).Error)
	op := &BillingOperation{ID: "unresolved-old-request", UserID: id, State: "reserved", CreatedAt: start + 60}
	require.NoError(t, DB.Create(op).Error)
	statement := &BillingStatement{ID: "version-one", UserID: id, Month: "2020-02", StorageProfileID: 1, CreatedBy: 1}
	build := func(account *BillingAccount, hours []BillingHour) (string, string, error) {
		return "{}", strings.Repeat("a", 64), nil
	}
	require.ErrorContains(t, CreateBillingStatement(statement, build), "unfinished")
	require.NoError(t, FinishBillingOperation(op.ID, id, 0, true))
	require.NoError(t, CreateBillingStatement(statement, build))
	assert.Equal(t, start, statement.StartAt)
	assert.Equal(t, end, statement.EndAt)
	duplicate := &BillingStatement{ID: "duplicate", UserID: id, Month: "2020-02"}
	require.ErrorContains(t, CreateBillingStatement(duplicate, build), "active statement")
	_, err = ChangeBillingStatement(statement.ID, "void", "", "replaced draft", "", 1, true)
	require.NoError(t, err)
	replacement := &BillingStatement{ID: "version-two", UserID: id, Month: "2020-02"}
	require.NoError(t, CreateBillingStatement(replacement, build))
	assert.Equal(t, 2, replacement.Revision)
	original, err := GetBillingStatement(statement.ID)
	require.NoError(t, err)
	assert.Equal(t, StatementVoid, original.Status)
	assert.Equal(t, "{}", original.Snapshot)
}

func TestBillingStatementPaginationKeepsSameSecondRecords(t *testing.T) {
	id := seedBillingCustomer(t)
	for _, record := range []BillingStatement{
		{ID: "record-b", UserID: id, Month: "2020-01", Revision: 1, CreatedAt: 100, IssuedAt: 101, Status: StatementIssued},
		{ID: "record-a", UserID: id, Month: "2020-02", Revision: 1, CreatedAt: 100, IssuedAt: 101, Status: StatementConfirmed},
		{ID: "private-draft", UserID: id, Month: "2020-03", Revision: 1, CreatedAt: 99, Status: StatementDraft},
	} {
		require.NoError(t, DB.Create(&record).Error)
	}
	next, err := ListBillingStatements(id, 100, "record-b", "all", true)
	require.NoError(t, err)
	require.Len(t, next, 1)
	assert.Equal(t, "record-a", next[0].ID)
	confirmed, err := ListBillingStatements(id, 0, "", StatementConfirmed, true)
	require.NoError(t, err)
	require.Len(t, confirmed, 1)
	assert.Equal(t, "record-a", confirmed[0].ID)
}

func TestBillingStatementExportsOnlyFrozenMonthlySequenceRange(t *testing.T) {
	id := seedBillingCustomer(t)
	start, end, err := BillingMonthBounds("2020-02")
	require.NoError(t, err)
	require.NoError(t, DB.Model(&BillingAccount{}).Where("user_id = ?", id).Updates(map[string]interface{}{"accounting_start_at": start - 86400, "sequence": 1000}).Error)
	for _, entry := range []BillingEntry{
		{EventKey: "old", UserID: id, Sequence: 5, PostedAt: start - 1, Kind: "usage"},
		{EventKey: "first", UserID: id, Sequence: 11, PostedAt: start, Kind: "usage", Quota: 20},
		{EventKey: "last", UserID: id, Sequence: 12, PostedAt: start + 1, Kind: "usage", Quota: 30},
		{EventKey: "newer", UserID: id, Sequence: 1000, PostedAt: end, Kind: "usage"},
	} {
		require.NoError(t, DB.Create(&entry).Error)
	}
	require.NoError(t, DB.Create(&BillingHour{UserID: id, Hour: start, Charge: 50, Count: 2, FirstSequence: 11, LastSequence: 12}).Error)
	statement := &BillingStatement{ID: "bounded-month", UserID: id, Month: "2020-02"}
	require.NoError(t, CreateBillingStatement(statement, func(*BillingAccount, []BillingHour) (string, string, error) { return "{}", "", nil }))
	assert.Equal(t, int64(11), statement.FromSequence)
	assert.Equal(t, int64(12), statement.ToSequence)
	page, err := GetBillingStatementEntries(statement, 0, 1)
	require.NoError(t, err)
	require.Len(t, page, 1)
	assert.Equal(t, "first", page[0].EventKey)
	page, err = GetBillingStatementEntries(statement, page[0].Sequence, 1000)
	require.NoError(t, err)
	require.Len(t, page, 1)
	assert.Equal(t, "last", page[0].EventKey)
}

func TestBillingArchiveCompletionPersistsDigestsAndRequiresCurrentLease(t *testing.T) {
	id := seedBillingCustomer(t)
	statement := &BillingStatement{ID: "archive-completion", UserID: id, Month: "2020-02", Status: StatementPreparing, LeaseOwner: "current-worker", LeaseUntil: 100}
	require.NoError(t, DB.Create(statement).Error)
	statement.ManifestSHA256, statement.PDFSHA256 = "manifest-digest", "pdf-digest"
	require.NoError(t, CompleteBillingStatementArchive(statement, "current-worker", true))
	stored, err := GetBillingStatement(statement.ID)
	require.NoError(t, err)
	assert.Equal(t, StatementDraft, stored.Status)
	assert.Equal(t, statement.ManifestSHA256, stored.ManifestSHA256)
	assert.Equal(t, statement.PDFSHA256, stored.PDFSHA256)
	assert.Empty(t, stored.LeaseOwner)
	assert.Zero(t, stored.LeaseUntil)
	assert.ErrorIs(t, CompleteBillingStatementArchive(statement, "stale-worker", false), ErrBillingConflict)
	stored, err = GetBillingStatement(statement.ID)
	require.NoError(t, err)
	assert.Equal(t, StatementDraft, stored.Status, "stale failures cannot overwrite the published draft")
}

func TestBillingStatementCannotIssueCorruptedSnapshot(t *testing.T) {
	id := seedBillingCustomer(t)
	statement := &BillingStatement{ID: "corrupt-snapshot", UserID: id, Month: "2020-01", Revision: 1, Status: StatementDraft, Snapshot: "{}", SnapshotSHA256: "does-not-match", ManifestSHA256: "manifest", PDFSHA256: "pdf"}
	require.NoError(t, DB.Create(statement).Error)
	_, err := ChangeBillingStatement(statement.ID, "issue", "", "", "", 1, true)
	assert.ErrorIs(t, err, ErrBillingEvidenceIntegrity)
	preserved, err := GetBillingStatement(statement.ID)
	require.NoError(t, err)
	assert.Equal(t, StatementDraft, preserved.Status)
	assert.Zero(t, preserved.IssuedAt)
}
