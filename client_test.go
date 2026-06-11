package serront

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	c := New(Config{Token: "test"})
	if c == nil {
		t.Fatal("client is nil")
	}
	if c.Storefront == nil || c.Services == nil || c.Orders == nil || c.Modules == nil ||
		c.Ledger == nil || c.Payouts == nil || c.APIKeys == nil ||
		c.WebhookSubscriptions == nil || c.Billing == nil || c.Public == nil {
		t.Fatal("resource namespaces not wired")
	}
}

func TestOrdersListUnwrapsEnvelopeAndSendsFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/orders" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok_123" {
			t.Errorf("auth header = %q", got)
		}
		q := r.URL.Query()
		for key, want := range map[string]string{
			"status": "requested", "paymentStatus": "unpaid",
			"serviceId": "svc_1", "q": "wedding", "limit": "5", "cursor": "cur_abc",
		} {
			if got := q.Get(key); got != want {
				t.Errorf("%s = %q, want %q", key, got, want)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":  map[string]any{"orders": []any{}, "counts": map[string]int{"requested": 2}, "cursor": nil, "hasMore": false},
			"error": nil,
			"meta":  map[string]any{"requestId": "req_o", "timestamp": "now"},
		})
	}))
	defer srv.Close()

	c := New(Config{Token: "tok_123", BaseURL: srv.URL})
	out, err := c.Orders.List(context.Background(), &OrderListParams{
		Status: StatusRequested, PaymentStatus: PaymentUnpaid,
		ServiceID: "svc_1", Q: "wedding", Limit: 5, Cursor: "cur_abc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Counts["requested"] != 2 || out.HasMore {
		t.Errorf("unexpected list: %+v", out)
	}
}

func TestEnvelopeErrorBecomesTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":  nil,
			"error": map[string]any{"code": "INVALID_TRANSITION", "message": "cannot move", "param": "status"},
			"meta":  map[string]any{"requestId": "req_x", "timestamp": "now"},
		})
	}))
	defer srv.Close()

	c := New(Config{Token: "tok", BaseURL: srv.URL})
	status := StatusCompleted
	_, err := c.Orders.Update(context.Background(), "ord_1", OrderUpdateInput{Status: &status})
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Code != "INVALID_TRANSITION" || apiErr.Status != 409 || apiErr.RequestID != "req_x" || apiErr.Param != "status" {
		t.Errorf("unexpected error: %+v", apiErr)
	}
}

func TestPublicNeedsNoToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("unexpected auth header %q", got)
		}
		if r.URL.Path != "/api/v1/public/storefront/studio/order" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":  map[string]any{"number": 1, "accessToken": "at_x", "quotedPriceIdr": 500000, "discountAmountIdr": 0},
			"error": nil,
			"meta":  map[string]any{"requestId": "req_p", "timestamp": "now"},
		})
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, Token: ""})
	out, err := c.Public.CreateOrder(context.Background(), "studio", PublicOrderCreateInput{
		ServiceSlug: "logo-design", BuyerName: "Budi", BuyerEmail: "budi@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.AccessToken != "at_x" || out.QuotedPriceIDR != 500000 {
		t.Errorf("unexpected order: %+v", out)
	}
}

func TestAuthedWithoutTokenFailsFast(t *testing.T) {
	t.Setenv("SERRONT_TOKEN", "")
	c := New(Config{BaseURL: "http://127.0.0.1:0"})
	_, err := c.Services.List(context.Background())
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Code != "AUTH_REQUIRED" {
		t.Fatalf("expected AUTH_REQUIRED, got %v", err)
	}
}

func TestBillingGetAndCheckout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/billing":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"subscription":  map[string]any{"id": nil, "accountId": "acc_1", "tier": "free", "status": "active"},
					"effectiveTier": "free",
					"earlyAccess":   false,
					"tiers":         []any{map[string]any{"id": "starter", "name": "Starter", "priceIdr": 79000, "serviceLimit": 10}},
				},
				"error": nil,
				"meta":  map[string]any{"requestId": "req_b", "timestamp": "now"},
			})
		case "/api/v1/billing/checkout":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["tier"] != "starter" {
				t.Errorf("tier = %q", body["tier"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":  map[string]any{"checkoutSessionId": "cs_1", "hostedUrl": "https://pay.example/cs_1"},
				"error": nil,
				"meta":  map[string]any{"requestId": "req_c", "timestamp": "now"},
			})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(Config{Token: "tok", BaseURL: srv.URL})
	info, err := c.Billing.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.EffectiveTier != TierFree || len(info.Tiers) != 1 || info.Tiers[0].ServiceLimit != 10 {
		t.Errorf("unexpected billing info: %+v", info)
	}
	out, err := c.Billing.Checkout(context.Background(), TierStarter)
	if err != nil {
		t.Fatal(err)
	}
	if out.HostedURL != "https://pay.example/cs_1" {
		t.Errorf("hostedUrl = %q", out.HostedURL)
	}
}

func TestProofUploadAndDownload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "PUT" && r.URL.Path == "/api/v1/public/orders/at_x/proof":
			if got := r.Header.Get("Content-Type"); got != "image/png" {
				t.Errorf("Content-Type = %q", got)
			}
			body, _ := io.ReadAll(r.Body)
			if len(body) != 3 {
				t.Errorf("body length = %d", len(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":  map[string]any{"hasProof": true},
				"error": nil,
				"meta":  map[string]any{"requestId": "req_u", "timestamp": "now"},
			})
		case r.Method == "GET" && r.URL.Path == "/api/v1/orders/ord_1/proof":
			if got := r.Header.Get("Authorization"); got != "Bearer tok" {
				t.Errorf("auth header = %q", got)
			}
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte{1, 2, 3})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(Config{Token: "tok", BaseURL: srv.URL})
	ok, err := c.Public.UploadProof(context.Background(), "at_x", ProofUploadInput{
		Data: []byte{1, 2, 3}, ContentType: "image/png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("hasProof = false")
	}
	dl, err := c.Orders.DownloadProof(context.Background(), "ord_1")
	if err != nil {
		t.Fatal(err)
	}
	if dl.ContentType != "image/png" || len(dl.Data) != 3 {
		t.Errorf("unexpected download: %+v", dl)
	}
}

func TestStorefrontPutAndModulesSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "PUT" && r.URL.Path == "/api/v1/storefront":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["slug"] != "studio" {
				t.Errorf("slug = %v", body["slug"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":  map[string]any{"accountId": "acc_1", "slug": "studio", "displayName": "Studio", "published": true},
				"error": nil,
				"meta":  map[string]any{"requestId": "req_s", "timestamp": "now"},
			})
		case r.Method == "PATCH" && r.URL.Path == "/api/v1/modules":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"modules":           map[string]any{"payment": true},
					"plugipayAccountId": "acc_p",
					"ripploAccountId":   nil,
					"available":         map[string]any{"payment": true, "marketing": true},
				},
				"error": nil,
				"meta":  map[string]any{"requestId": "req_m", "timestamp": "now"},
			})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(Config{Token: "tok", BaseURL: srv.URL})
	sf, err := c.Storefront.Put(context.Background(), StorefrontPutInput{
		Slug: "studio", DisplayName: "Studio", Published: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sf.Slug != "studio" {
		t.Errorf("slug = %q", sf.Slug)
	}
	enabled := true
	mods, err := c.Modules.Set(context.Background(), ModulesSetInput{Payment: &enabled})
	if err != nil {
		t.Fatal(err)
	}
	if !mods.Modules.Payment || mods.PlugipayAccountID == nil || *mods.PlugipayAccountID != "acc_p" {
		t.Errorf("unexpected modules: %+v", mods)
	}
}
