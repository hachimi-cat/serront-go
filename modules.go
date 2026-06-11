package serront

import "context"

// ModulesResource is the Payment / Marketing integrations registry
// (Bearer auth).
type ModulesResource struct {
	c *Client
}

// Get calls GET /api/v1/modules — current module state + provisioned
// partner workspace ids.
func (r *ModulesResource) Get(ctx context.Context) (*ModulesView, error) {
	var out ModulesView
	if err := r.c.do(ctx, "GET", "/api/v1/modules", nil, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ModulesSetInput is the payload for Modules.Set. Nil fields are
// omitted; provide Payment and/or Marketing.
type ModulesSetInput struct {
	Payment   *bool `json:"payment,omitempty"`
	Marketing *bool `json:"marketing,omitempty"`
}

// Set calls PATCH /api/v1/modules — toggle modules; first enable
// provisions the partner workspace (Plugipay / Ripllo) idempotently.
// 403 UPGRADE_REQUIRED on the Free tier.
func (r *ModulesResource) Set(ctx context.Context, input ModulesSetInput) (*ModulesView, error) {
	var out ModulesView
	if err := r.c.do(ctx, "PATCH", "/api/v1/modules", nil, input, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
