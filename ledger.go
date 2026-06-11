package serront

import (
	"context"
	"net/url"
	"strconv"
)

// LedgerResource is the seller's money views, proxied from Plugipay
// (Bearer auth; Payment module required — 409 PAYMENT_MODULE_DISABLED
// when off).
type LedgerResource struct {
	c *Client
}

// Balance calls GET /api/v1/ledger/balance — collapsed balance +
// per-code rows.
func (r *LedgerResource) Balance(ctx context.Context) (*LedgerBalance, error) {
	var out LedgerBalance
	if err := r.c.do(ctx, "GET", "/api/v1/ledger/balance", nil, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// LedgerEntriesParams filters Ledger.Entries.
type LedgerEntriesParams struct {
	// Limit caps the page size (1-100, default 50).
	Limit int
	// Cursor is the opaque cursor from a previous page.
	Cursor string
	// Code filters by ledger code.
	Code string
	// SourceType / SourceID filter by the originating record.
	SourceType string
	SourceID   string
}

// Entries calls GET /api/v1/ledger/entries — recent ledger entries
// (cursor-paged).
func (r *LedgerResource) Entries(ctx context.Context, params *LedgerEntriesParams) (*LedgerEntryList, error) {
	q := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if params.Code != "" {
			q.Set("code", params.Code)
		}
		if params.SourceType != "" {
			q.Set("sourceType", params.SourceType)
		}
		if params.SourceID != "" {
			q.Set("sourceId", params.SourceID)
		}
	}
	var out LedgerEntryList
	if err := r.c.do(ctx, "GET", "/api/v1/ledger/entries", q, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
