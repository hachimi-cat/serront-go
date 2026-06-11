package serront

import (
	"context"
	"net/url"
)

// ServicesResource is the seller's catalog (Bearer auth).
type ServicesResource struct {
	c *Client
}

// List calls GET /api/v1/services — all services, sortOrder asc.
func (r *ServicesResource) List(ctx context.Context) ([]Service, error) {
	var out []Service
	if err := r.c.do(ctx, "GET", "/api/v1/services", nil, nil, false, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ServiceCreateInput is the payload for Services.Create.
type ServiceCreateInput struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// PricingType is "fixed" / "hourly" (require PriceIDR) or
	// "package" (requires 1-5 Packages).
	PricingType string           `json:"pricingType"`
	PriceIDR    *int             `json:"priceIdr,omitempty"`
	Packages    []ServicePackage `json:"packages,omitempty"`
	Active      *bool            `json:"active,omitempty"`
	SortOrder   *int             `json:"sortOrder,omitempty"`
}

// Create calls POST /api/v1/services. 403 LIMIT_REACHED past the
// tier's service limit (freemium gate, enforced day one).
func (r *ServicesResource) Create(ctx context.Context, input ServiceCreateInput) (*Service, error) {
	var out Service
	if err := r.c.do(ctx, "POST", "/api/v1/services", nil, input, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get calls GET /api/v1/services/:id.
func (r *ServicesResource) Get(ctx context.Context, id string) (*Service, error) {
	var out Service
	if err := r.c.do(ctx, "GET", "/api/v1/services/"+url.PathEscape(id), nil, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ServiceUpdateInput is the payload for Services.Update. Nil fields
// are omitted.
type ServiceUpdateInput struct {
	Slug        *string `json:"slug,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	PricingType *string `json:"pricingType,omitempty"`
	PriceIDR    *int    `json:"priceIdr,omitempty"`
	// Packages replaces the package list; a pointer to an empty slice
	// is rejected for package-priced services.
	Packages *[]ServicePackage `json:"packages,omitempty"`
	Active   *bool             `json:"active,omitempty"`
	// SortOrder reorders — the services list sorts sortOrder asc.
	SortOrder *int `json:"sortOrder,omitempty"`
}

// Update calls PATCH /api/v1/services/:id — partial update incl.
// reorder via SortOrder.
func (r *ServicesResource) Update(ctx context.Context, id string, input ServiceUpdateInput) (*Service, error) {
	var out Service
	if err := r.c.do(ctx, "PATCH", "/api/v1/services/"+url.PathEscape(id), nil, input, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete calls DELETE /api/v1/services/:id — 409 CONFLICT while orders
// reference it (deactivate instead).
func (r *ServicesResource) Delete(ctx context.Context, id string) error {
	return r.c.do(ctx, "DELETE", "/api/v1/services/"+url.PathEscape(id), nil, nil, false, nil)
}
