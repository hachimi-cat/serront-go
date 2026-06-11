package serront

import (
	"context"
	"net/url"
)

// APIKeysResource manages programmatic access keys (Bearer auth).
type APIKeysResource struct {
	c *Client
}

// List calls GET /api/v1/api-keys — display-safe summaries (hashes
// never leave the database).
func (r *APIKeysResource) List(ctx context.Context) ([]APIKeySummary, error) {
	var out struct {
		APIKeys []APIKeySummary `json:"apiKeys"`
	}
	if err := r.c.do(ctx, "GET", "/api/v1/api-keys", nil, nil, false, &out); err != nil {
		return nil, err
	}
	return out.APIKeys, nil
}

// Create calls POST /api/v1/api-keys — the plaintext Key is returned
// ONCE here and never again.
func (r *APIKeysResource) Create(ctx context.Context, name string) (*APIKeyCreated, error) {
	var out APIKeyCreated
	payload := map[string]string{"name": name}
	if err := r.c.do(ctx, "POST", "/api/v1/api-keys", nil, payload, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete calls DELETE /api/v1/api-keys/:id.
func (r *APIKeysResource) Delete(ctx context.Context, id string) error {
	return r.c.do(ctx, "DELETE", "/api/v1/api-keys/"+url.PathEscape(id), nil, nil, false, nil)
}
