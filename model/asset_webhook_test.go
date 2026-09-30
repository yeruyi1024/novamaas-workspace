package model

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestAssetWebhookDeliveryClaimsAcrossSharedDatabase(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			prefix := fmt.Sprintf("test_aw_%d_", time.Now().UnixNano())
			var dialector gorm.Dialector
			switch dialect {
			case "mysql":
				dsn := strings.TrimSpace(os.Getenv("TEST_MYSQL_DSN"))
				if dsn == "" {
					t.Skip("TEST_MYSQL_DSN is not configured")
				}
				dialector = mysql.Open(dsn)
			case "postgres":
				dsn := strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
				if dsn == "" {
					t.Skip("TEST_POSTGRES_DSN is not configured")
				}
				dialector = postgres.Open(dsn)
			default:
				dialector = sqlite.Open("file:" + prefix + "?mode=memory&cache=shared")
			}
			db, err := gorm.Open(dialector, &gorm.Config{NamingStrategy: schema.NamingStrategy{TablePrefix: prefix}})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			if dialect == "sqlite" {
				sqlDB.SetMaxOpenConns(1)
			} else {
				sqlDB.SetMaxOpenConns(2)
			}
			t.Cleanup(func() { _ = sqlDB.Close() })
			require.NoError(t, db.AutoMigrate(&AssetWebhookDelivery{}))
			t.Cleanup(func() { assert.NoError(t, db.Migrator().DropTable(&AssetWebhookDelivery{})) })
			originalDB := DB
			DB = db
			t.Cleanup(func() { DB = originalDB })

			now := common.GetTimestamp()
			delivery := AssetWebhookDelivery{WebhookID: "wh_shared", Payload: `{}`, Status: AssetWebhookDeliveryStatusPending}
			require.NoError(t, db.Create(&delivery).Error)
			// Both nodes must read the same candidate before either conditional
			// UPDATE runs. This reproduces contention without timing-based sleeps.
			ready, release := make(chan struct{}, 2), make(chan struct{})
			t.Cleanup(func() {
				select {
				case <-release:
				default:
					close(release)
				}
			})
			require.NoError(t, db.Callback().Update().Before("gorm:begin_transaction").Register("webhook_claim_barrier", func(_ *gorm.DB) {
				ready <- struct{}{}
				<-release
			}))
			t.Cleanup(func() { _ = db.Callback().Update().Remove("webhook_claim_barrier") })
			type result struct {
				claimed []*AssetWebhookDelivery
				err     error
			}
			results := make(chan result, 2)
			for _, owner := range []string{"node-a", "node-b"} {
				go func() {
					claimed, err := ClaimAssetWebhookDeliveries(now, now+60, owner, 1)
					results <- result{claimed, err}
				}()
			}
			for range 2 {
				select {
				case <-ready:
				case <-time.After(10 * time.Second):
					t.Fatal("both nodes did not reach the claim barrier")
				}
			}
			close(release)
			var claims []*AssetWebhookDelivery
			for range 2 {
				select {
				case result := <-results:
					require.NoError(t, result.err)
					claims = append(claims, result.claimed...)
				case <-time.After(10 * time.Second):
					t.Fatal("claim did not finish")
				}
			}
			require.NoError(t, db.Callback().Update().Remove("webhook_claim_barrier"))
			require.Len(t, claims, 1, "a shared delivery must have only one owner")
			oldOwner := claims[0].LockedBy
			owned, err := OwnsAssetWebhookDeliveryLease(context.Background(), delivery.ID, oldOwner, now)
			require.NoError(t, err)
			assert.True(t, owned)

			// Expired workers cannot finish or discard work, even before takeover.
			require.NoError(t, db.Model(&delivery).Update("lease_until", now-1).Error)
			assert.Error(t, FinishAssetWebhookDelivery(context.Background(), delivery.ID, oldOwner, AssetWebhookDeliveryStatusSucceeded, 204, 0, ""))
			assert.Error(t, SupersedeAssetWebhookDelivery(context.Background(), delivery.ID, oldOwner, "stale status"))
			claimed, err := ClaimAssetWebhookDeliveries(now, now+60, "replacement", 1)
			require.NoError(t, err)
			require.Len(t, claimed, 1)
			owned, err = OwnsAssetWebhookDeliveryLease(context.Background(), delivery.ID, oldOwner, now)
			require.NoError(t, err)
			assert.False(t, owned)
			assert.Error(t, FinishAssetWebhookDelivery(context.Background(), delivery.ID, oldOwner, AssetWebhookDeliveryStatusSucceeded, 204, 0, ""))
			assert.Error(t, SupersedeAssetWebhookDelivery(context.Background(), delivery.ID, oldOwner, "stale status"))
			require.NoError(t, db.First(&delivery, delivery.ID).Error)
			assert.Equal(t, "replacement", delivery.LockedBy)
			assert.Equal(t, AssetWebhookDeliveryStatusDelivering, delivery.Status)
			assert.Zero(t, delivery.Attempts)
			require.NoError(t, FinishAssetWebhookDelivery(context.Background(), delivery.ID, "replacement", AssetWebhookDeliveryStatusSucceeded, 204, 0, ""))
			require.NoError(t, db.First(&delivery, delivery.ID).Error)
			assert.Equal(t, AssetWebhookDeliveryStatusSucceeded, delivery.Status)
			assert.Equal(t, 1, delivery.Attempts)
			claimed, err = ClaimAssetWebhookDeliveries(now+120, now+180, "later-node", 1)
			require.NoError(t, err)
			assert.Empty(t, claimed, "successful deliveries must not be sent again")
		})
	}
}
