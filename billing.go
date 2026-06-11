package serront

import "context"

// BillingResource is the workspace plan + Plugipay checkout surface
// (Bearer auth).
type BillingResource struct {
	c *Client
}

// TierDef is one row of the tier table (ids: free / starter / growth /
// business).
type TierDef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// PriceIDR is whole rupiah per month. 0 = free.
	PriceIDR int      `json:"priceIdr"`
	Blurb    string   `json:"blurb"`
	Features []string `json:"features"`
	// ServiceLimit is the max services on the storefront — ENFORCED
	// (403 LIMIT_REACHED).
	ServiceLimit int `json:"serviceLimit"`
	// AgentLimit is the agent-seat limit (Huudis workspace members).
	AgentLimit int `json:"agentLimit"`
	// BrandingRemoval: may the seller hide the "Powered by Serront"
	// footer?
	BrandingRemoval bool `json:"brandingRemoval"`
	// ModuleAccess: may the seller enable the Payment/Marketing
	// modules?
	ModuleAccess bool `json:"moduleAccess"`
}

// BillingSubscription is the workspace's current subscription. A
// workspace with no purchase history reports tier "free" with a nil ID.
type BillingSubscription struct {
	ID                        *string `json:"id"`
	AccountID                 string  `json:"accountId"`
	Tier                      string  `json:"tier"`   // free | starter | growth | business
	Status                    string  `json:"status"` // active | past_due | canceled
	PlugipayCheckoutSessionID *string `json:"plugipayCheckoutSessionId"`
	CurrentPeriodEnd          *string `json:"currentPeriodEnd"`
}

// BillingInfo is the GET /api/v1/billing response.
type BillingInfo struct {
	Subscription BillingSubscription `json:"subscription"`
	// EffectiveTier is the tier the freemium gates actually honor
	// (lapsed/canceled rows fall back to free).
	EffectiveTier string `json:"effectiveTier"`
	// EarlyAccess is always false on serront — freemium limits are
	// enforced day one.
	EarlyAccess bool      `json:"earlyAccess"`
	Tiers       []TierDef `json:"tiers"`
}

// Get calls GET /api/v1/billing — current subscription + effectiveTier
// + the tier table.
func (r *BillingResource) Get(ctx context.Context) (*BillingInfo, error) {
	var out BillingInfo
	if err := r.c.do(ctx, "GET", "/api/v1/billing", nil, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckoutResult carries the Plugipay hosted checkout to redirect the
// browser to.
type CheckoutResult struct {
	CheckoutSessionID string `json:"checkoutSessionId"`
	HostedURL         string `json:"hostedUrl"`
}

// Checkout calls POST /api/v1/billing/checkout for a paid tier
// ("starter" | "growth" | "business"); redirect the browser to
// HostedURL.
func (r *BillingResource) Checkout(ctx context.Context, tier string) (*CheckoutResult, error) {
	var out CheckoutResult
	payload := map[string]string{"tier": tier}
	if err := r.c.do(ctx, "POST", "/api/v1/billing/checkout", nil, payload, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
