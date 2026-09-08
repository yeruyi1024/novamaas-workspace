package model

import (
	"context"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type storageQuerySQLRecorder struct {
	statement string
}

func (recorder *storageQuerySQLRecorder) LogMode(logger.LogLevel) logger.Interface { return recorder }
func (recorder *storageQuerySQLRecorder) Info(context.Context, string, ...any)     {}
func (recorder *storageQuerySQLRecorder) Warn(context.Context, string, ...any)     {}
func (recorder *storageQuerySQLRecorder) Error(context.Context, string, ...any)    {}

func (recorder *storageQuerySQLRecorder) Trace(_ context.Context, _ time.Time, sql func() (string, int64), _ error) {
	recorder.statement, _ = sql()
}

func TestGetStoragePolicyByKeyQuotesReservedColumn(t *testing.T) {
	originalDB := DB
	originalMainType := common.MainDatabaseType()
	originalLogType := common.LogDatabaseType()
	t.Cleanup(func() {
		DB = originalDB
		common.SetDatabaseTypes(originalMainType, originalLogType)
		initCol()
	})

	tests := []struct {
		name          string
		databaseType  common.DatabaseType
		dialector     gorm.Dialector
		whereFragment string
	}{
		{
			name:          "sqlite",
			databaseType:  common.DatabaseTypeSQLite,
			dialector:     sqlite.Open(":memory:"),
			whereFragment: "WHERE `key` = \"relay_media_temp\"",
		},
		{
			name:         "mysql",
			databaseType: common.DatabaseTypeMySQL,
			dialector: mysql.New(mysql.Config{
				DSN:                       "gorm:gorm@tcp(127.0.0.1:3306)/gorm?parseTime=true",
				SkipInitializeWithVersion: true,
			}),
			whereFragment: "WHERE `key` = 'relay_media_temp'",
		},
		{
			name:         "postgres",
			databaseType: common.DatabaseTypePostgreSQL,
			dialector: postgres.New(postgres.Config{
				DSN:                  "host=127.0.0.1 user=gorm dbname=gorm port=5432 sslmode=disable",
				PreferSimpleProtocol: true,
			}),
			whereFragment: `WHERE "key" = 'relay_media_temp'`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := &storageQuerySQLRecorder{}
			db, err := gorm.Open(test.dialector, &gorm.Config{
				DryRun:               true,
				DisableAutomaticPing: true,
				Logger:               recorder,
			})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { _ = sqlDB.Close() })

			DB = db
			common.SetDatabaseTypes(test.databaseType, originalLogType)
			initCol()

			_, err = GetStoragePolicyByKey(StoragePolicyRelayMediaTemp)
			require.NoError(t, err)
			assert.Contains(t, recorder.statement, test.whereFragment)
		})
	}
}
