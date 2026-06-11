package serront

import "context"

// StorefrontResource is the seller's storefront profile (Bearer auth).
type StorefrontResource struct {
	c *Client
}

// Get calls GET /api/v1/storefront — current settings (nil until the
// first Put).
func (r *StorefrontResource) Get(ctx context.Context) (*StorefrontSettings, error) {
	var out *StorefrontSettings
	if err := r.c.do(ctx, "GET", "/api/v1/storefront", nil, nil, false, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// StorefrontPutInput is the payload for Storefront.Put (full replace).
type StorefrontPutInput struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio,omitempty"`
	// WhatsappNumber is E.164 (e.g. +62812…); nil omits, a pointer to
	// "" is rejected server-side — clear with an explicit null via a
	// nil pointer left out of the JSON (omitempty).
	WhatsappNumber     *string       `json:"whatsappNumber,omitempty"`
	ManualBankAccounts []BankAccount `json:"manualBankAccounts,omitempty"`
	ManualInstructions *string       `json:"manualInstructions,omitempty"`
	// HideBranding is a paid perk — 403 UPGRADE_REQUIRED on Free.
	HideBranding bool `json:"hideBranding,omitempty"`
	Published    bool `json:"published,omitempty"`
}

// Put calls PUT /api/v1/storefront — upsert settings (full replace).
func (r *StorefrontResource) Put(ctx context.Context, input StorefrontPutInput) (*StorefrontSettings, error) {
	var out StorefrontSettings
	if err := r.c.do(ctx, "PUT", "/api/v1/storefront", nil, input, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
