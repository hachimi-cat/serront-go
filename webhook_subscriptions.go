package serront

import (
	"context"
	"net/url"
)

// WebhookSubscriptionsResource manages outbound webhook endpoints
// receiving serront.order.* events (Bearer auth).
type WebhookSubscriptionsResource struct {
	c *Client
}

// List calls GET /api/v1/webhook-subscriptions — secrets never
// included.
func (r *WebhookSubscriptionsResource) List(ctx context.Context) ([]WebhookSubscription, error) {
	var out struct {
		Subscriptions []WebhookSubscription `json:"subscriptions"`
	}
	if err := r.c.do(ctx, "GET", "/api/v1/webhook-subscriptions", nil, nil, false, &out); err != nil {
		return nil, err
	}
	return out.Subscriptions, nil
}

// WebhookSubscriptionCreateInput is the payload for
// WebhookSubscriptions.Create.
type WebhookSubscriptionCreateInput struct {
	URL string `json:"url"`
	// Events defaults to ["*"]; entries are "*" or versioned serront
	// event types (serront.order.created.v1, …).
	Events []string `json:"events,omitempty"`
}

// Create calls POST /api/v1/webhook-subscriptions — the signing Secret
// is returned ONCE here and never again.
func (r *WebhookSubscriptionsResource) Create(ctx context.Context, input WebhookSubscriptionCreateInput) (*WebhookSubscriptionCreated, error) {
	var out WebhookSubscriptionCreated
	if err := r.c.do(ctx, "POST", "/api/v1/webhook-subscriptions", nil, input, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update calls PATCH /api/v1/webhook-subscriptions/:id — pause/resume
// delivery.
func (r *WebhookSubscriptionsResource) Update(ctx context.Context, id string, active bool) (*WebhookSubscription, error) {
	var out WebhookSubscription
	payload := map[string]bool{"active": active}
	path := "/api/v1/webhook-subscriptions/" + url.PathEscape(id)
	if err := r.c.do(ctx, "PATCH", path, nil, payload, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete calls DELETE /api/v1/webhook-subscriptions/:id.
func (r *WebhookSubscriptionsResource) Delete(ctx context.Context, id string) error {
	path := "/api/v1/webhook-subscriptions/" + url.PathEscape(id)
	return r.c.do(ctx, "DELETE", path, nil, nil, false, nil)
}
