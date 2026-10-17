package service

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	hosttypes "github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamBillingSettlesWalletTokenAndLedgerOnce(t *testing.T) {
	for _, tc := range []struct {
		name      string
		reason    relaycommon.StreamEndReason
		snapshot  bool
		zeroUsage bool
		tiered    bool
	}{
		{"completed", relaycommon.StreamEndReasonDone, false, false, false},
		{"upstream interrupted", relaycommon.StreamEndReasonScannerErr, false, false, false},
		{"client cancelled", relaycommon.StreamEndReasonClientGone, false, false, false},
		{"converted display usage", relaycommon.StreamEndReasonScannerErr, true, false, false},
		{"tiered converted interrupted usage", relaycommon.StreamEndReasonScannerErr, true, false, true},
		{"no billable usage", relaycommon.StreamEndReasonClientGone, false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			truncate(t)
			const userID, tokenID, channelID = 730, 731, 732
			const initialQuota, reservation = 10_000, 2_000
			seedUser(t, userID, initialQuota)
			seedToken(t, tokenID, userID, "stream-billing-fixture", initialQuota)
			seedChannel(t, channelID)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
			requestID := fmt.Sprintf("stream-billing-%s", tc.name)
			c.Set(common.RequestIdKey, requestID)
			info := &relaycommon.RelayInfo{
				UserId: userID, TokenId: tokenID, TokenKey: "stream-billing-fixture", RequestId: requestID,
				OriginModelName: "stream-billing-fixture", RelayFormat: types.RelayFormatOpenAI,
				ChannelMeta:     &relaycommon.ChannelMeta{ChannelId: channelID},
				ForcePreConsume: true, IsStream: true, StartTime: time.Now(),
				UserSetting: dto.UserSetting{BillingPreference: "wallet_only"},
				PriceData:   hosttypes.PriceData{ModelRatio: 1, CompletionRatio: 2, CacheRatio: 0.1, GroupRatioInfo: hosttypes.GroupRatioInfo{GroupRatio: 1}},
			}
			require.Nil(t, PreConsumeBilling(c, reservation, info))
			quota, err := model.GetUserQuota(userID, false)
			require.NoError(t, err)
			assert.Equal(t, initialQuota-reservation, quota)
			original := &dto.Usage{PromptTokens: 5000, CompletionTokens: 60, TotalTokens: 5060, PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 4800}}
			usage := original
			if tc.snapshot {
				usage = &dto.Usage{PromptTokens: 200, CompletionTokens: 60, TotalTokens: 260, BillingUsage: dto.NewOpenAIChatBillingUsage(original)}
			}
			wantQuota := 800 // 200 uncached + 4800*0.1 cached + 60*2 output.
			if tc.tiered {
				const expr = `len <= 4000 ? tier("short",p+c+cr*0.01) : tier("long",p*3+c*8+cr*0.3)`
				info.TieredBillingSnapshot = &billingexpr.BillingSnapshot{
					BillingMode: "tiered_expr", ExprString: expr, ExprHash: billingexpr.ExprHashString(expr),
					GroupRatio: 1, QuotaPerUnit: 500_000, EstimatedQuotaAfterGroup: reservation,
				}
				wantQuota = 1260 // Full input length selects long: (200*3+60*8+4800*0.3)/2.
			}
			if tc.zeroUsage {
				usage = &dto.Usage{}
				wantQuota = 0
			}
			info.StreamStatus = relaycommon.NewStreamStatus()
			var endErr error
			if tc.reason == relaycommon.StreamEndReasonScannerErr {
				endErr = io.ErrUnexpectedEOF
			}
			info.StreamStatus.SetEndReason(tc.reason, endErr)
			PostTextConsumeQuota(c, info, usage, nil)
			// The lifecycle may attempt cleanup after an error status. Neither a
			// repeated settlement nor a late refund may alter committed money.
			require.NoError(t, info.Billing.Settle(wantQuota))
			info.Billing.Refund(c)
			assert.False(t, info.Billing.NeedsRefund())

			var user model.User
			var token model.Token
			var channel model.Channel
			require.NoError(t, model.DB.First(&user, userID).Error)
			require.NoError(t, model.DB.First(&token, tokenID).Error)
			require.NoError(t, model.DB.First(&channel, channelID).Error)
			assert.Equal(t, initialQuota-wantQuota, user.Quota)
			assert.Equal(t, wantQuota, user.UsedQuota)
			assert.Equal(t, initialQuota-wantQuota, token.RemainQuota)
			assert.Equal(t, wantQuota, token.UsedQuota)
			assert.Equal(t, int64(wantQuota), channel.UsedQuota)
			var operation model.BillingOperation
			require.NoError(t, model.DB.Where("request_id = ?", requestID).First(&operation).Error)
			assert.Equal(t, "settled", operation.State)
			assert.Equal(t, int64(reservation), operation.Reserved)
			assert.Equal(t, int64(wantQuota), operation.Actual)
			var entries []model.BillingEntry
			require.NoError(t, model.DB.Where("request_id = ?", requestID).Find(&entries).Error)
			require.Len(t, entries, 1)
			assert.Equal(t, int64(wantQuota), entries[0].Quota)
			assert.Equal(t, -int64(wantQuota), entries[0].WalletDelta)
			assert.Equal(t, int64(initialQuota-wantQuota), entries[0].BalanceAfter)
			var logs []model.Log
			require.NoError(t, model.LOG_DB.Where("user_id = ? AND type = ?", userID, model.LogTypeConsume).Find(&logs).Error)
			require.Len(t, logs, 1)
			assert.Equal(t, wantQuota, logs[0].Quota)
		})
	}
}
