package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func upstreamBillingAllocationFixture(t *testing.T) *model.UpstreamBillingGroupView {
	t.Helper()
	group := upstreamBillingServiceFixture(t)
	oldUnit := common.QuotaPerUnit
	oldCurrency := operation_setting.GetGeneralSetting().QuotaDisplayType
	common.QuotaPerUnit = 500000
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeUSD
	t.Cleanup(func() {
		common.QuotaPerUnit = oldUnit
		operation_setting.GetGeneralSetting().QuotaDisplayType = oldCurrency
	})
	require.NoError(t, model.DB.Create(&[]model.User{{Id: 1, Username: "alice", DisplayName: "Alice Corp", Email: "private@example.com", AffCode: "alice-aff"}, {Id: 2, Username: "bob", DisplayName: "Bob Corp", AffCode: "bob-aff"}, {Id: 3, Username: "deleted-customer", AffCode: "deleted-aff"}}).Error)
	require.NoError(t, model.DB.Delete(&model.User{}, 3).Error)
	require.NoError(t, model.DB.Model(&model.Channel{}).Where("id = ?", 26).Update("cost_discount", "0.5").Error)
	require.NoError(t, model.DB.Create(&model.Channel{Id: 28, Name: "Free channel", Key: "private", CostDiscount: "0"}).Error)
	group, err := model.SaveUpstreamBillingGroup(context.Background(), group.ID, group.Version, group.Name, []int{19, 26, 27, 28}, 1)
	require.NoError(t, err)
	start, _, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	for _, event := range []struct {
		user, channel, kind int
		cost                int64
		discount, request   string
	}{
		{user: 1, channel: 19, kind: model.LogTypeConsume, cost: 100, discount: "1", request: "a1"},
		{user: 2, channel: 19, kind: model.LogTypeConsume, cost: 20, discount: "0.2", request: "b1"},
		{user: 2, channel: 19, kind: model.LogTypeRefund, cost: -80, discount: "0.8", request: "b-refund"},
		{user: 1, channel: 26, kind: model.LogTypeConsume, cost: 30, discount: "0.3", request: "a2"},
		{user: 0, channel: 26, kind: model.LogTypeConsume, cost: 10, discount: "0.1", request: "unknown"},
		{user: 3, channel: 26, kind: model.LogTypeConsume, cost: 5, discount: "0.05", request: "deleted"},
		{user: 2, channel: 28, kind: model.LogTypeConsume, cost: 0, discount: "0", request: "free"},
	} {
		basis, prompt, metadata := "100", 10, `{"admin_info":{"usage_billing_path":"upstream"}}`
		if event.kind == model.LogTypeRefund {
			basis, prompt, metadata = "-100", 0, ""
		}
		upstreamBillingCostLog(t, model.Log{UserId: event.user, Type: event.kind, CreatedAt: start + 1, ChannelId: event.channel, ModelName: "text", Quota: 100, PromptTokens: prompt, RequestId: event.request, Other: metadata}, basis, event.discount, event.cost)
	}
	require.NoError(t, model.LOG_DB.Create(&model.Log{UserId: 1, Type: model.LogTypeConsume, CreatedAt: start + 1, ChannelId: 27, ModelName: "excluded", Quota: 999, PromptTokens: 999, RequestId: "excluded"}).Error)
	return group
}

func TestUpstreamBillingChannelUserAllocationUsesExactCostsAndRetainsUnknownAndDeletedUsers(t *testing.T) {
	group := upstreamBillingAllocationFixture(t)
	report, err := GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "", "")
	require.NoError(t, err)
	assert.Equal(t, 2, report.SchemaVersion)
	assert.True(t, report.Coverage.Complete)
	require.Len(t, report.ChannelUsers, 6)
	for i, want := range []struct {
		channel, user                          int
		charge, refund, requests               int64
		amount, channelAmount, username, share string
	}{
		{19, 1, 100, 0, 1, "0.000200", "0.000080", "alice", "250.0000"},
		{19, 2, 20, 80, 1, "-0.000120", "0.000080", "bob", "-150.0000"},
		{26, 0, 10, 0, 1, "0.000020", "0.000090", "user-0", "22.2222"},
		{26, 1, 30, 0, 1, "0.000060", "0.000090", "alice", "66.6667"},
		{26, 3, 5, 0, 1, "0.000010", "0.000090", "user-3", "11.1111"},
		{28, 2, 0, 0, 1, "0.000000", "0.000000", "bob", ""},
	} {
		row := report.ChannelUsers[i]
		assert.Equal(t, want.channel, row.ChannelID)
		assert.Equal(t, want.user, row.UserID)
		assert.Equal(t, want.username, row.Username)
		assert.Equal(t, want.charge, row.ChargeQuota)
		assert.Equal(t, want.refund, row.RefundQuota)
		assert.Equal(t, want.requests, row.Requests)
		assert.Equal(t, want.amount, row.Amount)
		assert.Equal(t, want.channelAmount, row.ChannelAmount)
		if want.share == "" {
			assert.Nil(t, row.SharePercent)
		} else {
			require.NotNil(t, row.SharePercent)
			assert.Equal(t, want.share, *row.SharePercent)
		}
	}
	assert.Equal(t, "Alice Corp", report.ChannelUsers[0].DisplayName)
	assert.Equal(t, int64(7), report.Total.Count)
	body, err := common.Marshal(report.ChannelUsers)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "private@example.com")
	assert.NotContains(t, string(body), "deleted-customer")
}

func TestUpstreamBillingChannelShareIsUnavailableForUnknownOrNonpositiveChannelCost(t *testing.T) {
	group := upstreamBillingAllocationFixture(t)
	start, _, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	require.NoError(t, model.LOG_DB.Create(&model.Log{UserId: 7, Type: model.LogTypeConsume, CreatedAt: start + 2, ChannelId: 19, ModelName: "unknown-cost", Quota: 100, PromptTokens: 10, RequestId: "missing-cost", Other: `{"admin_info":{"usage_billing_path":"upstream"}}`}).Error)
	report, err := GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "", "")
	require.NoError(t, err)
	assert.False(t, report.Coverage.Complete)
	for _, row := range report.ChannelUsers {
		if row.ChannelID == 19 {
			assert.Nil(t, row.SharePercent)
		}
		if row.ChannelID == 26 {
			assert.NotNil(t, row.SharePercent)
		}
	}
	upstreamBillingCostLog(t, model.Log{UserId: 3, Type: model.LogTypeRefund, CreatedAt: start + 3, ChannelId: 26, ModelName: "text", Quota: 100, RequestId: "net-negative"}, "-100", "0.8", -80)
	negative, err := GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "", "")
	require.NoError(t, err)
	for _, row := range negative.ChannelUsers {
		if row.ChannelID == 26 {
			assert.Nil(t, row.SharePercent)
			assert.Equal(t, "-0.000070", row.ChannelAmount)
		}
	}
}

func TestUpstreamBillingChannelSharesUseQuotaEvenWhenDisplayedAmountsRoundToZero(t *testing.T) {
	group := upstreamBillingAllocationFixture(t)
	settings := operation_setting.GetGeneralSetting()
	oldRate, oldSymbol := settings.CustomCurrencyExchangeRate, settings.CustomCurrencySymbol
	settings.QuotaDisplayType = operation_setting.QuotaDisplayTypeCustom
	settings.CustomCurrencyExchangeRate, settings.CustomCurrencySymbol = 0.000001, "¤"
	t.Cleanup(func() { settings.CustomCurrencyExchangeRate, settings.CustomCurrencySymbol = oldRate, oldSymbol })
	report, err := GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "", "")
	require.NoError(t, err)
	require.NotEmpty(t, report.ChannelUsers)
	assert.Equal(t, "0.000000", report.ChannelUsers[0].Amount)
	assert.Equal(t, "0.000000", report.ChannelUsers[0].ChannelAmount)
	require.NotNil(t, report.ChannelUsers[0].SharePercent)
	assert.Equal(t, "250.0000", *report.ChannelUsers[0].SharePercent)
}

func TestUpstreamBillingChannelUserAllocationFreezesIdentityAndRefusesTamperedArchiveDistribution(t *testing.T) {
	group := upstreamBillingAllocationFixture(t)
	statement, err := PrepareUpstreamBillingStatement(context.Background(), group.ID, 1, 1, "2026-08")
	require.NoError(t, err)
	var snapshot UpstreamBillingSnapshot
	require.NoError(t, common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot))
	require.Len(t, snapshot.ChannelUsers, 6)
	require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", 1).Updates(map[string]any{"username": "renamed", "display_name": "New company"}).Error)
	require.NoError(t, model.DB.Model(&model.Channel{}).Where("id = ?", 19).Updates(map[string]any{"name": "Renamed channel", "cost_discount": "0.2"}).Error)
	store := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildUpstreamBillingArchive(context.Background(), statement, store), "archive retries use frozen identity and costs")
	var archived UpstreamBillingSnapshot
	require.NoError(t, common.Unmarshal(store.files["snapshot-0"], &archived))
	assert.Equal(t, "alice", archived.ChannelUsers[0].Username)
	assert.Equal(t, "Alice Corp", archived.ChannelUsers[0].DisplayName)
	assert.Equal(t, "primary", archived.ChannelUsers[0].ChannelName)
	assert.Equal(t, int64(100), archived.ChannelUsers[0].ChargeQuota)
	for _, tamper := range []struct {
		name   string
		mutate func(*UpstreamBillingSnapshot)
	}{
		{"share", func(s *UpstreamBillingSnapshot) { value := "100.0000"; s.ChannelUsers[0].SharePercent = &value }},
		{"channel total", func(s *UpstreamBillingSnapshot) { s.ChannelUsers[0].ChannelAmount = "999.000000" }},
		{"user usage", func(s *UpstreamBillingSnapshot) { s.ChannelUsers[0].InputTokens++ }},
		{"user cost", func(s *UpstreamBillingSnapshot) { s.ChannelUsers[0].ChargeQuota++ }},
		{"missing user", func(s *UpstreamBillingSnapshot) { s.ChannelUsers = s.ChannelUsers[1:] }},
	} {
		t.Run(tamper.name, func(t *testing.T) {
			var changed UpstreamBillingSnapshot
			require.NoError(t, common.UnmarshalJsonStr(string(statement.Snapshot), &changed))
			tamper.mutate(&changed)
			body, err := common.Marshal(changed)
			require.NoError(t, err)
			hash := sha256.Sum256(body)
			copy := *statement
			copy.Snapshot, copy.SnapshotSHA256 = model.UpstreamBillingSnapshotData(body), hex.EncodeToString(hash[:])
			assert.ErrorIs(t, BuildUpstreamBillingArchive(context.Background(), &copy, &memoryUpstreamBillingArchive{files: map[string][]byte{}}), model.ErrBillingEvidenceIntegrity)
		})
	}
}
