package serront

import (
	"context"
	"net/url"
	"strconv"
)

// PayoutsResource is the seller's payouts, proxied to Plugipay (Bearer
// auth; Payment module required — 409 PAYMENT_MODULE_DISABLED when
// off).
type PayoutsResource struct {
	c *Client
}

// PayoutListParams filters Payouts.List.
type PayoutListParams struct {
	// Limit caps the page size (1-100, default 50).
	Limit int
	// Cursor is the opaque cursor from a previous page.
	Cursor string
	// Status filters: pending / in_transit / paid / failed / cancelled.
	Status string
}

// List calls GET /api/v1/payouts — history (cursor-paged).
func (r *PayoutsResource) List(ctx context.Context, params *PayoutListParams) (*PayoutList, error) {
	q := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if params.Status != "" {
			q.Set("status", params.Status)
		}
	}
	var out PayoutList
	if err := r.c.do(ctx, "GET", "/api/v1/payouts", q, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PayoutCreateInput is the payload for Payouts.Create.
type PayoutCreateInput struct {
	// Amount is whole rupiah, positive.
	Amount            int     `json:"amount"`
	BankCode          *string `json:"bankCode,omitempty"`
	BankName          string  `json:"bankName,omitempty"`
	BankAccountNumber string  `json:"bankAccountNumber,omitempty"`
	BankAccountHolder string  `json:"bankAccountHolder,omitempty"`
	Note              *string `json:"note,omitempty"`
}

// Create calls POST /api/v1/payouts — request a payout (manual flow →
// pending).
func (r *PayoutsResource) Create(ctx context.Context, input PayoutCreateInput) (*Payout, error) {
	var out Payout
	if err := r.c.do(ctx, "POST", "/api/v1/payouts", nil, input, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Cancel calls POST /api/v1/payouts/:id/cancel — pending → cancelled.
func (r *PayoutsResource) Cancel(ctx context.Context, id string) (*Payout, error) {
	var out Payout
	path := "/api/v1/payouts/" + url.PathEscape(id) + "/cancel"
	if err := r.c.do(ctx, "POST", path, nil, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Balance calls GET /api/v1/payouts/balance — available = ledger −
// in-flight payouts.
func (r *PayoutsResource) Balance(ctx context.Context) (*PayoutBalance, error) {
	var out PayoutBalance
	if err := r.c.do(ctx, "GET", "/api/v1/payouts/balance", nil, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBankAccount calls GET /api/v1/payouts/bank-account — merchant
// default bank details.
func (r *PayoutsResource) GetBankAccount(ctx context.Context) (*PayoutBankAccount, error) {
	var out PayoutBankAccount
	if err := r.c.do(ctx, "GET", "/api/v1/payouts/bank-account", nil, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BankAccountUpdateInput is the payload for Payouts.UpdateBankAccount.
type BankAccountUpdateInput struct {
	BankCode          *string `json:"bankCode,omitempty"`
	BankName          string  `json:"bankName"`
	BankAccountNumber string  `json:"bankAccountNumber"`
	BankAccountHolder string  `json:"bankAccountHolder"`
}

// UpdateBankAccount calls PATCH /api/v1/payouts/bank-account.
func (r *PayoutsResource) UpdateBankAccount(ctx context.Context, input BankAccountUpdateInput) (*PayoutBankAccount, error) {
	var out PayoutBankAccount
	if err := r.c.do(ctx, "PATCH", "/api/v1/payouts/bank-account", nil, input, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MarkInTransit calls POST /api/v1/payouts/:id/mark-in-transit —
// operator transition (manual mode). reference may be nil.
func (r *PayoutsResource) MarkInTransit(ctx context.Context, id string, reference *string) (*Payout, error) {
	var out Payout
	path := "/api/v1/payouts/" + url.PathEscape(id) + "/mark-in-transit"
	payload := map[string]*string{"reference": reference}
	if err := r.c.do(ctx, "POST", path, nil, payload, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MarkPaid calls POST /api/v1/payouts/:id/mark-paid — operator
// transition. reference may be nil.
func (r *PayoutsResource) MarkPaid(ctx context.Context, id string, reference *string) (*Payout, error) {
	var out Payout
	path := "/api/v1/payouts/" + url.PathEscape(id) + "/mark-paid"
	payload := map[string]*string{"reference": reference}
	if err := r.c.do(ctx, "POST", path, nil, payload, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MarkFailed calls POST /api/v1/payouts/:id/mark-failed — operator
// transition.
func (r *PayoutsResource) MarkFailed(ctx context.Context, id string, failureReason string) (*Payout, error) {
	var out Payout
	path := "/api/v1/payouts/" + url.PathEscape(id) + "/mark-failed"
	payload := map[string]string{"failureReason": failureReason}
	if err := r.c.do(ctx, "POST", path, nil, payload, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
