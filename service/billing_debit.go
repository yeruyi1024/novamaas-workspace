package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// Only this reviewed two-row file may authorize recovery. The file itself
// stays with finance; the public repository contains no customer identifiers.
const approvedDebitFileSHA256 = "4533bf95b2c001de0192d70954a0d0a3f945850fa4f69d99001a49b650703229"

func ReviewedDebits() ([]model.ReviewedDebit, error) {
	document, err := os.ReadFile("approved-billing-debits.json") // /data in the release container
	if err != nil {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	if fmt.Sprintf("%x", sha256.Sum256(document)) != approvedDebitFileSHA256 {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	var file struct {
		Cases []model.ReviewedDebit `json:"cases"`
	}
	if err := common.Unmarshal(document, &file); err != nil {
		return nil, err
	}
	if len(file.Cases) != 2 {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	return file.Cases, nil
}

type ReviewedDebitRow struct {
	model.DebitReceipt
	UserID     int    `json:"user_id"`
	Username   string `json:"username"`
	RequestID  string `json:"request_id"`
	UpstreamID string `json:"upstream_request_id,omitempty"`
	Quota      int64  `json:"quota"`
	Amount     string `json:"amount"`
	Source     string `json:"source"`
}

func RunReviewedDebit(ctx context.Context, c model.ReviewedDebit, actor int, apply bool) (ReviewedDebitRow, error) {
	cur, err := currentBillingCurrency(true)
	if err != nil || cur.Code != "CNY" || cur.QuotaPerUnit != "500000" || cur.Rate != "7" {
		return ReviewedDebitRow{}, model.ErrBillingEvidenceIntegrity
	}
	amount, err := billingRow("", c.Quota, 0, 1, cur)
	if err != nil {
		return ReviewedDebitRow{}, err
	}
	row := ReviewedDebitRow{UserID: c.UserID, Username: c.Username, RequestID: c.RequestID, Quota: c.Quota, Amount: amount.Charge, Source: "usage_log"}
	if c.Supplier != nil {
		row.Source = "supplier_statement"
		row.UpstreamID = c.UpstreamID
	}
	row.DebitReceipt, err = model.RecoverReviewedDebit(ctx, c, actor, apply)
	if !apply && (errors.Is(err, model.ErrBillingConflict) || errors.Is(err, model.ErrBillingHistoryBlocked) || errors.Is(err, model.ErrBillingInsufficientQuota)) {
		row.Status = "blocked"
		return row, nil
	}
	return row, err
}
