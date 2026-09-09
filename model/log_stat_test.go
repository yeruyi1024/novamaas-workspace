package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSumUsedQuotaPreservesQuotaWhenLoadingRateStats(t *testing.T) {
	truncateTables(t)
	now := time.Now().Unix()
	require.NoError(t, LOG_DB.Create(&Log{
		UserId:           2,
		CreatedAt:        now,
		Type:             LogTypeConsume,
		Username:         "regular-user",
		Quota:            12345,
		PromptTokens:     100,
		CompletionTokens: 20,
	}).Error)

	stat, err := SumUsedQuota(
		LogTypeConsume,
		now-60,
		now+60,
		"",
		"regular-user",
		"",
		0,
		"",
	)

	require.NoError(t, err)
	assert.Equal(t, 12345, stat.Quota)
	assert.Equal(t, 1, stat.Rpm)
	assert.Equal(t, 120, stat.Tpm)
}
