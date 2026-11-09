package model

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Logs sharing one second span the page boundary, so paging must break ties by
// the stored id, which the user view replaces with display numbers.
func TestExportUserLogsPagesEveryLogOnceNewestFirst(t *testing.T) {
	truncateTables(t)
	now := time.Now().Unix()
	total := logExportPageSize*2 + 50
	logs := make([]Log, 0, total)
	for i := 0; i < total; i++ {
		group := "standard"
		if i%2 == 1 {
			group = "preferred"
		}
		logs = append(logs, Log{UserId: 5, Type: LogTypeConsume, CreatedAt: now, Group: group, RequestId: fmt.Sprintf("req-%d", i)})
	}
	require.NoError(t, LOG_DB.CreateInBatches(&logs, 500).Error)
	require.NoError(t, LOG_DB.Create(&Log{UserId: 5, Type: LogTypeConsume, CreatedAt: now + 5, Group: "standard", RequestId: "newest"}).Error)
	require.NoError(t, LOG_DB.Create(&Log{UserId: 6, Type: LogTypeConsume, CreatedAt: now, Group: "standard", RequestId: "other-user"}).Error)

	for _, tc := range []struct {
		name  string
		group string
		want  int
	}{
		{"all groups", "", total + 1},
		{"one group", "standard", total/2 + 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var exported []*Log
			err := ExportUserLogs(context.Background(), 5, LogTypeUnknown, now-60, now+60, "", "", tc.group, "", "", func(log *Log) error {
				exported = append(exported, log)
				return nil
			})

			require.NoError(t, err)
			require.Len(t, exported, tc.want)
			assert.Equal(t, "newest", exported[0].RequestId)
			seen := make(map[string]bool, len(exported))
			for i, log := range exported {
				assert.False(t, seen[log.RequestId], "log %s exported twice", log.RequestId)
				seen[log.RequestId] = true
				if i > 0 {
					assert.GreaterOrEqual(t, exported[i-1].CreatedAt, log.CreatedAt, "logs must stay newest first")
				}
			}
		})
	}
}
