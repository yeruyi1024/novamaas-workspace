package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestUpstreamExcelFreezesChannelUserCostsAndRefundShares(t *testing.T) {
	statement, snapshot, records := frozenUpstreamBillingFixture(t)
	snapshot.SchemaVersion = 2
	records[1].UserID = 84
	for i, record := range records {
		row := UpstreamBillingRow{}
		require.NoError(t, row.addArchiveRecord(record))
		amounts, err := billingRow("", row.ChargeQuota, row.RefundQuota, row.Count, snapshot.Currency)
		require.NoError(t, err)
		row.BillingRow = amounts
		channelAmount := "10.500000"
		var share *string
		switch i {
		case 0:
			value := "133.3333"
			share = &value
		case 1:
			value := "-33.3333"
			share = &value
		case 2:
			value := "100.0000"
			share, channelAmount = &value, "12.600000"
		case 3:
			channelAmount = "0.000000"
		}
		snapshot.ChannelUsers = append(snapshot.ChannelUsers, UpstreamBillingChannelUserRow{ChannelID: record.ChannelID, ChannelName: "=channel-name", UserID: record.UserID, Username: "=2+2", DisplayName: "Finance customer", ChannelAmount: channelAmount, SharePercent: share, UpstreamBillingRow: row})
	}
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement.Snapshot, statement.SnapshotSHA256 = model.UpstreamBillingSnapshotData(body), hex.EncodeToString(hash[:])
	workbook, err := RenderUpstreamBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	book, err := excelize.OpenReader(bytes.NewReader(workbook))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, book.Close()) })
	assert.Contains(t, book.GetSheetList(), "渠道用户成本")
	for _, expected := range []struct{ cell, value string }{
		{"A2", "19"}, {"C2", "42"}, {"C3", "84"}, {"D2", "=2+2"},
		{"H2", "14"}, {"H3", "-3.5"}, {"I2", "10.5"},
		{"J2", "133.3333%"}, {"J3", "-33.3333%"}, {"J4", "100.0000%"}, {"J5", "—"},
	} {
		value, err := book.GetCellValue("渠道用户成本", expected.cell, excelize.Options{RawCellValue: true})
		require.NoError(t, err)
		assert.Equal(t, expected.value, value, expected.cell)
	}
	formula, err := book.GetCellFormula("渠道用户成本", "D2")
	require.NoError(t, err)
	assert.Empty(t, formula, "an upstream or customer identity must never become an Excel formula")
}
