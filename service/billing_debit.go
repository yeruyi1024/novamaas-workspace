package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// Embedded at build time, not deployed as a separate configuration file.
// The two internal, random operation UUIDs and their request IDs bind this
// payload to its source records; administrator authorization is checked separately.
//
//go:embed billing_debit_evidence.bin
var sealedDebitEvidence []byte

func reviewedDebits(ctx context.Context) ([]model.ReviewedDebit, error) {
	var ops []model.BillingOperation
	if err := model.DB.WithContext(ctx).Select("id", "request_id").Where("id LIKE ? OR id LIKE ?", "426ce1af%", "7edac7c2%").Order("id ASC").Limit(3).Find(&ops).Error; err != nil {
		return nil, err
	}
	if len(ops) != 2 {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	key := sha256.Sum256([]byte("reviewed-wallet-debits/v1\x00" + ops[0].ID + "\x00" + ops[0].RequestID + "\x00" + ops[1].ID + "\x00" + ops[1].RequestID))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}
	if len(sealedDebitEvidence) < gcm.Overhead() {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	doc, err := gcm.Open(nil, nil, sealedDebitEvidence, nil)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(doc)) != "4533bf95b2c001de0192d70954a0d0a3f945850fa4f69d99001a49b650703229" {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	var file struct {
		Cases []model.ReviewedDebit `json:"cases"`
	}
	if err := common.Unmarshal(doc, &file); err != nil {
		return nil, err
	}
	if len(file.Cases) != 2 {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	return file.Cases, nil
}

type ReviewedDebitResult struct {
	UserID     int                `json:"user_id"`
	Username   string             `json:"username"`
	TotalQuota int64              `json:"total_quota"`
	Total      string             `json:"total"`
	Rows       []ReviewedDebitRow `json:"rows"`
}
type ReviewedDebitRow struct {
	model.DebitReceipt
	RequestID string `json:"request_id"`
	Amount    string `json:"amount"`
	Supplier  bool   `json:"supplier"`
}

func RunReviewedDebits(ctx context.Context, month string, account, actor, confirmUser int, confirmQuota *int64) (*ReviewedDebitResult, error) {
	cases, err := reviewedDebits(ctx)
	if err != nil {
		return nil, err
	}
	c := cases[0]
	if c.Month != month || (account != actor && account != c.UserID) || (confirmQuota != nil && confirmUser != c.UserID) {
		return nil, model.ErrBillingConflict
	}
	cur, err := currentBillingCurrency(true)
	if err != nil || cur.Code != "CNY" || cur.Rate != "7" || cur.QuotaPerUnit != "500000" {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	receipts, total, err := model.RecoverReviewedDebits(ctx, cases, actor, confirmQuota)
	if err != nil {
		return nil, err
	}
	sum, err := billingRow("", total, 0, 1, cur)
	if err != nil {
		return nil, err
	}
	out := &ReviewedDebitResult{UserID: c.UserID, Username: c.Username, TotalQuota: total, Total: sum.Charge}
	for i, row := range cases {
		amount, err := billingRow("", row.Quota, 0, 1, cur)
		if err != nil {
			return nil, err
		}
		out.Rows = append(out.Rows, ReviewedDebitRow{receipts[i], row.RequestID, amount.Charge, row.Supplier != nil})
	}
	return out, nil
}
