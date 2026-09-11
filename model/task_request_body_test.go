package model

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTaskRequestBodyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&TaskRequestBody{}))
	previousDB := DB
	DB = db
	t.Cleanup(func() { DB = previousDB })
	return db
}

func TestSaveTaskRequestSnapshotsStoresBodiesInSeparateRows(t *testing.T) {
	db := setupTaskRequestBodyTestDB(t)
	original := []byte(`{"model":"public-model","prompt":"hello"}`)
	upstream := []byte(`{"model":"upstream-model","content":[{"type":"text","text":"hello"}]}`)

	require.NoError(t, SaveTaskRequestSnapshots("task_snapshot", "request_snapshot", original, upstream))

	snapshots, exists, err := GetArchivedRequestSnapshots("task_snapshot", "request_snapshot")
	require.NoError(t, err)
	require.True(t, exists)
	assert.JSONEq(t, string(original), string(snapshots.Original))
	assert.JSONEq(t, string(upstream), string(snapshots.Upstream))

	var records []TaskRequestBody
	require.NoError(t, db.Order("reference_id ASC").Find(&records).Error)
	require.Len(t, records, 2)
	assert.NotEqual(t, records[0].ReferenceID, records[1].ReferenceID)
	for _, record := range records {
		assert.LessOrEqual(t, record.BodySize, int64(MaxTaskRequestBodyBytes))
		assert.Len(t, record.BodySHA256, 64)
	}
}

func TestSaveTaskRequestSnapshotsRejectsOversizedUpstreamBodyAtomically(t *testing.T) {
	db := setupTaskRequestBodyTestDB(t)
	original := []byte(`{"prompt":"safe"}`)
	upstream := bytes.Repeat([]byte(" "), MaxTaskRequestBodyBytes+1)

	err := SaveTaskRequestSnapshots("task_oversized", "request_oversized", original, upstream)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTaskRequestBodyTooLarge))
	var count int64
	require.NoError(t, db.Model(&TaskRequestBody{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestUpstreamRequestBodyReferenceIsBoundedForMaximumTaskID(t *testing.T) {
	reference, err := upstreamRequestBodyReference(string(bytes.Repeat([]byte("t"), 191)), "")

	require.NoError(t, err)
	assert.LessOrEqual(t, len(reference), 191)
	assert.Len(t, reference, len(upstreamRequestBodyReferencePrefix)+sha256.Size*2)
}
