package serront

import (
	"context"
	"net/url"
)

// PublicResource is the buyer-facing, UNauthenticated surface — the
// published storefront + tokenized order endpoints. No Bearer token
// needed.
type PublicResource struct {
	c *Client
}

// GetStorefront calls GET /api/v1/public/storefront/:slug — published
// profile + active services (404 for unpublished slugs).
func (r *PublicResource) GetStorefront(ctx context.Context, slug string) (*PublicStorefrontView, error) {
	var out PublicStorefrontView
	path := "/api/v1/public/storefront/" + url.PathEscape(slug)
	if err := r.c.do(ctx, "GET", path, nil, nil, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PublicOrderCreateInput is the payload for Public.CreateOrder.
type PublicOrderCreateInput struct {
	ServiceSlug string `json:"serviceSlug"`
	// PackageName is required iff the service's pricing type is
	// "package".
	PackageName string `json:"packageName,omitempty"`
	BuyerName   string `json:"buyerName"`
	BuyerEmail  string `json:"buyerEmail"`
	// BuyerPhone is E.164 (e.g. +62812…).
	BuyerPhone string `json:"buyerPhone,omitempty"`
	// PreferredDate is an ISO date.
	PreferredDate string `json:"preferredDate,omitempty"`
	Notes         string `json:"notes,omitempty"`
	// DiscountCode needs the Marketing module; invalid codes reject the
	// order (DISCOUNT_INVALID).
	DiscountCode string `json:"discountCode,omitempty"`
}

// CreateOrder calls POST /api/v1/public/storefront/:slug/order —
// request a service. The returned AccessToken is the buyer's only
// credential.
func (r *PublicResource) CreateOrder(ctx context.Context, slug string, input PublicOrderCreateInput) (*PublicOrderCreated, error) {
	var out PublicOrderCreated
	path := "/api/v1/public/storefront/" + url.PathEscape(slug) + "/order"
	if err := r.c.do(ctx, "POST", path, nil, input, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ValidateDiscount calls POST
// /api/v1/public/storefront/:slug/validate-discount — Marketing
// module: dry-run a discount code against a gross price (whole
// rupiah). Read-only — never mutates anything.
func (r *PublicResource) ValidateDiscount(ctx context.Context, slug, code string, priceIDR int) (*DiscountValidation, error) {
	var out DiscountValidation
	path := "/api/v1/public/storefront/" + url.PathEscape(slug) + "/validate-discount"
	payload := map[string]any{"code": code, "priceIdr": priceIDR}
	if err := r.c.do(ctx, "POST", path, nil, payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOrder calls GET /api/v1/public/orders/:accessToken — the
// tokenized order view (status + thread + payment instructions;
// internal notes are never exposed).
func (r *PublicResource) GetOrder(ctx context.Context, accessToken string) (*PublicOrderView, error) {
	var out PublicOrderView
	path := "/api/v1/public/orders/" + url.PathEscape(accessToken)
	if err := r.c.do(ctx, "GET", path, nil, nil, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReplyOrder calls POST /api/v1/public/orders/:accessToken/messages —
// a buyer reply on their own order.
func (r *PublicResource) ReplyOrder(ctx context.Context, accessToken string, body string) (*PublicOrderMessage, error) {
	var out PublicOrderMessage
	path := "/api/v1/public/orders/" + url.PathEscape(accessToken) + "/messages"
	payload := map[string]string{"body": body}
	if err := r.c.do(ctx, "POST", path, nil, payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ClaimPayment calls POST
// /api/v1/public/orders/:accessToken/claim-payment — buyer "I have
// transferred" (unpaid → payment_claimed; a re-claim is a 200 no-op).
func (r *PublicResource) ClaimPayment(ctx context.Context, accessToken string) (string, error) {
	var out struct {
		PaymentStatus string `json:"paymentStatus"`
	}
	path := "/api/v1/public/orders/" + url.PathEscape(accessToken) + "/claim-payment"
	if err := r.c.do(ctx, "POST", path, nil, nil, true, &out); err != nil {
		return "", err
	}
	return out.PaymentStatus, nil
}

// Pay calls POST /api/v1/public/orders/:accessToken/pay — Payment
// module: mint a Plugipay hosted checkout for the quote; redirect the
// browser to HostedURL. 409 PAYMENT_MODULE_DISABLED when off.
func (r *PublicResource) Pay(ctx context.Context, accessToken string) (*CheckoutResult, error) {
	var out CheckoutResult
	path := "/api/v1/public/orders/" + url.PathEscape(accessToken) + "/pay"
	if err := r.c.do(ctx, "POST", path, nil, nil, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ProofUploadInput is the payload for Public.UploadProof.
type ProofUploadInput struct {
	// Data is the raw image bytes (≤5MB).
	Data []byte
	// ContentType must be image/png, image/jpeg, image/webp or
	// image/gif.
	ContentType string
}

// UploadProof calls PUT /api/v1/public/orders/:accessToken/proof —
// upload the payment-proof image.
func (r *PublicResource) UploadProof(ctx context.Context, accessToken string, input ProofUploadInput) (bool, error) {
	var out struct {
		HasProof bool `json:"hasProof"`
	}
	path := "/api/v1/public/orders/" + url.PathEscape(accessToken) + "/proof"
	if err := r.c.doRaw(ctx, "PUT", path, input.Data, input.ContentType, true, &out); err != nil {
		return false, err
	}
	return out.HasProof, nil
}
