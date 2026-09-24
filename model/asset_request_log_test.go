package model

import (
	"context"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAssetRequestLogsFilterAndCursorDoNotSkipSameMillisecond(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	previousDB, previousLogDB := DB, LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	DB, LOG_DB = db, db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		DB, LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
	})
	require.NoError(t, MigrateAssetRequestLogs())
	require.NoError(t, InsertAssetRequestLogs(context.Background(), []AssetRequestLog{
		{CreatedAt: 2000, ChannelID: 7, ReplicaID: 91, Source: "sync", Result: "failure", RequestID: "req-1", Operation: "GetAsset"},
		{CreatedAt: 2000, ChannelID: 7, ReplicaID: 91, Source: "sync", Result: "success", RequestID: "req-2", Operation: "CreateAsset"},
		{CreatedAt: 1900, ChannelID: 7, Source: "test", Result: "success", RequestID: "req-3", Operation: "ListAssets"},
		{CreatedAt: 2000, ChannelID: 8, Source: "test", Result: "success", RequestID: "req-4", Operation: "ListAssets"},
	}))
	filter := AssetRequestLogFilter{ChannelID: 7, StartMS: 1000, EndMS: 3000, Limit: 1}
	first, cursor, err := ListAssetRequestLogs(context.Background(), filter)
	require.NoError(t, err)
	require.Len(t, first, 1)
	assert.Equal(t, "req-2", first[0].RequestID)
	require.NotEmpty(t, cursor)
	filter.CursorMS, filter.CursorID, err = ParseAssetRequestLogCursor(cursor)
	require.NoError(t, err)
	second, cursor, err := ListAssetRequestLogs(context.Background(), filter)
	require.NoError(t, err)
	require.Len(t, second, 1)
	assert.Equal(t, "req-1", second[0].RequestID)
	require.NotEmpty(t, cursor)
	filter.CursorMS, filter.CursorID, err = ParseAssetRequestLogCursor(cursor)
	require.NoError(t, err)
	third, cursor, err := ListAssetRequestLogs(context.Background(), filter)
	require.NoError(t, err)
	require.Len(t, third, 1)
	assert.Equal(t, "req-3", third[0].RequestID)
	assert.Empty(t, cursor)

	filtered, _, err := ListAssetRequestLogs(context.Background(), AssetRequestLogFilter{
		ChannelID: 7, ReplicaID: 91, Source: "sync", Result: "failure", StartMS: 1000, EndMS: 3000, Limit: 10,
	})
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	assert.Equal(t, "req-1", filtered[0].RequestID)
}

func TestAssetRequestLogRetentionDeletesOnlyExpiredRowsInSmallBatches(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	previousDB, previousLogDB := DB, LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	DB, LOG_DB = db, db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		DB, LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
	})
	require.NoError(t, MigrateAssetRequestLogs())
	require.NoError(t, InsertAssetRequestLogs(context.Background(), []AssetRequestLog{
		{CreatedAt: 1000, RequestID: "old-1"},
		{CreatedAt: 1001, RequestID: "old-2"},
		{CreatedAt: 2000, RequestID: "current"},
	}))
	count, err := PurgeAssetRequestLogsBefore(context.Background(), 1500, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	logs, _, err := ListAssetRequestLogs(context.Background(), AssetRequestLogFilter{StartMS: 1, EndMS: 3000, Limit: 10})
	require.NoError(t, err)
	assert.Len(t, logs, 2)
	count, err = PurgeAssetRequestLogsBefore(context.Background(), 1500, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	logs, _, err = ListAssetRequestLogs(context.Background(), AssetRequestLogFilter{StartMS: 1, EndMS: 3000, Limit: 10})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, "current", logs[0].RequestID)
}

func TestAssetRequestLogDetailsExpireBeforeMetadata(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	previousDB, previousLogDB := DB, LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	DB, LOG_DB = db, db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		DB, LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
	})
	require.NoError(t, MigrateAssetRequestLogs())
	require.NoError(t, InsertAssetRequestLogEntries(context.Background(), []AssetRequestLogEntry{
		{Log: AssetRequestLog{CreatedAt: 1000, RequestID: "old"}, Detail: AssetRequestLogDetail{RequestURL: "https://example.com/old"}},
		{Log: AssetRequestLog{CreatedAt: 2000, RequestID: "new"}, Detail: AssetRequestLogDetail{RequestURL: "https://example.com/new"}},
	}))
	logs, _, err := ListAssetRequestLogs(context.Background(), AssetRequestLogFilter{StartMS: 1, EndMS: 3000, Limit: 10})
	require.NoError(t, err)
	require.Len(t, logs, 2)
	oldID, newID := logs[1].ID, logs[0].ID
	count, err := PurgeAssetRequestLogDetailsBefore(context.Background(), 1500, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	oldEntry, err := GetAssetRequestLogWithDetail(context.Background(), oldID)
	require.NoError(t, err)
	assert.Nil(t, oldEntry.Detail)
	newEntry, err := GetAssetRequestLogWithDetail(context.Background(), newID)
	require.NoError(t, err)
	require.NotNil(t, newEntry.Detail)
	assert.Equal(t, "https://example.com/new", newEntry.Detail.RequestURL)
	count, err = PurgeAssetRequestLogsBefore(context.Background(), 1500, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
