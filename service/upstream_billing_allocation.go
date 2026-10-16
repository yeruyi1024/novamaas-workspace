package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/QuantumNous/new-api/model"
	"github.com/shopspring/decimal"
)

// Amount is the user's net cost; ChannelAmount is the whole channel's net
// cost. Shares use exact quota rather than rounded display amounts.
type UpstreamBillingChannelUserRow struct {
	ChannelID     int     `json:"channel_id"`
	ChannelName   string  `json:"channel_name"`
	UserID        int     `json:"user_id"`
	Username      string  `json:"username"`
	DisplayName   string  `json:"display_name"`
	ChannelAmount string  `json:"channel_amount"`
	SharePercent  *string `json:"share_percent"`
	UpstreamBillingRow
}

type upstreamBillingChannelUserKey struct{ channelID, userID int }

type upstreamBillingAllocation struct {
	users    map[upstreamBillingChannelUserKey]*upstreamBillingAggregate
	channels map[int]*upstreamBillingAggregate
}

func (allocation *upstreamBillingAllocation) Add(record model.UpstreamBillingStatementRecord) error {
	if allocation.users == nil {
		allocation.users = make(map[upstreamBillingChannelUserKey]*upstreamBillingAggregate)
		allocation.channels = make(map[int]*upstreamBillingAggregate)
	}
	key := upstreamBillingChannelUserKey{record.ChannelID, record.UserID}
	if allocation.users[key] == nil {
		allocation.users[key] = &upstreamBillingAggregate{}
	}
	if err := allocation.users[key].Add(record); err != nil {
		return err
	}
	if allocation.channels[record.ChannelID] == nil {
		allocation.channels[record.ChannelID] = &upstreamBillingAggregate{}
	}
	return allocation.channels[record.ChannelID].Add(record)
}

func (allocation *upstreamBillingAllocation) Rows(ctx context.Context, group model.UpstreamBillingGroupView, currency BillingCurrency) ([]UpstreamBillingChannelUserRow, error) {
	ids := make([]int, 0, len(allocation.users))
	for key := range allocation.users {
		ids = append(ids, key.userID)
	}
	identities, err := model.GetBillingUserIdentities(ctx, ids)
	if err != nil {
		return nil, err
	}
	names := make(map[int]string, len(group.Channels))
	for _, channel := range group.Channels {
		names[channel.ID] = channel.Name
	}
	rows := make([]UpstreamBillingChannelUserRow, 0, len(allocation.users))
	for key, total := range allocation.users {
		row, err := total.Row("", currency)
		if err != nil {
			return nil, err
		}
		channel, err := allocation.channels[key.channelID].Row("", currency)
		if err != nil {
			return nil, err
		}
		identity := identities[key.userID]
		if identity.Username == "" {
			identity.Username = fmt.Sprintf("user-%d", key.userID)
		}
		name := names[key.channelID]
		if name == "" {
			name = fmt.Sprintf("channel-%d", key.channelID)
		}
		rows = append(rows, UpstreamBillingChannelUserRow{ChannelID: key.channelID, ChannelName: name, UserID: key.userID, Username: identity.Username, DisplayName: identity.DisplayName, ChannelAmount: channel.Amount, SharePercent: upstreamBillingCostShare(row.ChargeQuota-row.RefundQuota, channel.ChargeQuota-channel.RefundQuota, channel.MissingCostRecords), UpstreamBillingRow: row})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ChannelID != rows[j].ChannelID {
			return rows[i].ChannelID < rows[j].ChannelID
		}
		return rows[i].UserID < rows[j].UserID
	})
	return rows, nil
}

func upstreamBillingCostShare(userNet, channelNet, missingCost int64) *string {
	if channelNet <= 0 || missingCost != 0 {
		return nil
	}
	value := decimal.NewFromInt(userNet).Mul(decimal.NewFromInt(100)).DivRound(decimal.NewFromInt(channelNet), 4).StringFixed(4)
	return &value
}

// Verify monetary allocations independently from the report accumulator. Names
// are frozen snapshot metadata; archive retries never resolve mutable users.
func verifyUpstreamBillingAllocation(snapshot *UpstreamBillingSnapshot, users map[upstreamBillingChannelUserKey]UpstreamBillingRow, channels map[int]UpstreamBillingRow) error {
	if snapshot.SchemaVersion == 1 {
		return nil
	}
	if len(snapshot.ChannelUsers) != len(users) {
		return model.ErrBillingEvidenceIntegrity
	}
	channelNames := make(map[int]string)
	for _, channel := range snapshot.Group.Channels {
		if channel.Eligible {
			name := channel.Name
			if name == "" {
				name = fmt.Sprintf("channel-%d", channel.ID)
			}
			channelNames[channel.ID] = name
		}
	}
	for index, expected := range snapshot.ChannelUsers {
		if index > 0 {
			previous := snapshot.ChannelUsers[index-1]
			if previous.ChannelID > expected.ChannelID || (previous.ChannelID == expected.ChannelID && previous.UserID >= expected.UserID) {
				return model.ErrBillingEvidenceIntegrity
			}
		}
		key := upstreamBillingChannelUserKey{expected.ChannelID, expected.UserID}
		actual, found := users[key]
		channel, channelFound := channels[expected.ChannelID]
		if !found || !channelFound || expected.ChannelName != channelNames[expected.ChannelID] || !actual.matchesArchiveRow(expected.UpstreamBillingRow, snapshot.Currency) {
			return model.ErrBillingEvidenceIntegrity
		}
		amount, err := billingRow("", channel.ChargeQuota, channel.RefundQuota, channel.Count, snapshot.Currency)
		if err != nil || amount.Amount != expected.ChannelAmount {
			return model.ErrBillingEvidenceIntegrity
		}
		share := upstreamBillingCostShare(actual.ChargeQuota-actual.RefundQuota, channel.ChargeQuota-channel.RefundQuota, channel.MissingCostRecords)
		if (share == nil) != (expected.SharePercent == nil) || (share != nil && *share != *expected.SharePercent) {
			return model.ErrBillingEvidenceIntegrity
		}
		delete(users, key)
	}
	if len(users) != 0 {
		return model.ErrBillingEvidenceIntegrity
	}
	return nil
}
