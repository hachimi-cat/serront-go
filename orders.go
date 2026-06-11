package serront

import (
	"context"
	"net/url"
	"strconv"
)

// OrdersResource is the seller's order desk (Bearer auth).
type OrdersResource struct {
	c *Client
}

// OrderListParams filters Orders.List.
type OrderListParams struct {
	// Status filters to one status ("requested", "confirmed",
	// "declined", "in_progress", "delivered", "completed", "canceled")
	// or "all". Empty = all.
	Status string
	// PaymentStatus filters to one payment status ("unpaid",
	// "payment_claimed", "payment_confirmed") or "all".
	PaymentStatus string
	// ServiceID filters to orders for one service.
	ServiceID string
	// Q is a free-text search across buyer name/email, notes, message
	// bodies and service name.
	Q string
	// Limit caps the page size (1-100, default 50).
	Limit int
	// Cursor is the opaque cursor from a previous page.
	Cursor string
}

// List calls GET /api/v1/orders — newest-activity-first, with
// per-status counts.
func (r *OrdersResource) List(ctx context.Context, params *OrderListParams) (*OrderList, error) {
	q := url.Values{}
	if params != nil {
		if params.Status != "" {
			q.Set("status", params.Status)
		}
		if params.PaymentStatus != "" {
			q.Set("paymentStatus", params.PaymentStatus)
		}
		if params.ServiceID != "" {
			q.Set("serviceId", params.ServiceID)
		}
		if params.Q != "" {
			q.Set("q", params.Q)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
	}
	var out OrderList
	if err := r.c.do(ctx, "GET", "/api/v1/orders", q, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get calls GET /api/v1/orders/:id — order + full thread (incl.
// internal notes).
func (r *OrdersResource) Get(ctx context.Context, id string) (*OrderWithMessages, error) {
	var out OrderWithMessages
	if err := r.c.do(ctx, "GET", "/api/v1/orders/"+url.PathEscape(id), nil, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// OrderUpdateInput is the payload for Orders.Update. Nil fields are
// omitted; provide Status and/or QuotedPriceIDR.
type OrderUpdateInput struct {
	Status         *string `json:"status,omitempty"`
	QuotedPriceIDR *int    `json:"quotedPriceIdr,omitempty"`
}

// Update calls PATCH /api/v1/orders/:id — status transition and/or the
// quote. Illegal transitions 409 INVALID_TRANSITION.
func (r *OrdersResource) Update(ctx context.Context, id string, input OrderUpdateInput) (*Order, error) {
	var out Order
	if err := r.c.do(ctx, "PATCH", "/api/v1/orders/"+url.PathEscape(id), nil, input, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// OrderReplyInput is the payload for Orders.Reply.
type OrderReplyInput struct {
	Body string `json:"body"`
	// IsInternal posts a seller-only internal note — never shown to
	// the buyer.
	IsInternal bool   `json:"isInternal,omitempty"`
	AuthorName string `json:"authorName,omitempty"`
}

// Reply calls POST /api/v1/orders/:id/messages — seller reply (or
// internal note when IsInternal is true).
func (r *OrdersResource) Reply(ctx context.Context, id string, input OrderReplyInput) (*OrderMessage, error) {
	var out OrderMessage
	path := "/api/v1/orders/" + url.PathEscape(id) + "/messages"
	if err := r.c.do(ctx, "POST", path, nil, input, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ConfirmPayment calls POST /api/v1/orders/:id/confirm-payment —
// unpaid|payment_claimed → payment_confirmed.
func (r *OrdersResource) ConfirmPayment(ctx context.Context, id string) (*Order, error) {
	var out Order
	path := "/api/v1/orders/" + url.PathEscape(id) + "/confirm-payment"
	if err := r.c.do(ctx, "POST", path, nil, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DownloadProof calls GET /api/v1/orders/:id/proof — the buyer's
// payment-proof image (404 NOT_FOUND when none uploaded).
func (r *OrdersResource) DownloadProof(ctx context.Context, id string) (*ProofDownload, error) {
	return r.c.download(ctx, "/api/v1/orders/"+url.PathEscape(id)+"/proof")
}
