package serront

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Client.API (api_generated.go) goes through apigenRequest: the same bearer
// and envelope handling as every hand-written call.

type apigenSeen struct {
	method, uri, auth, contentType string
	body                           []byte
}

func apigenServer(t *testing.T, status int, env map[string]any) (*httptest.Server, *[]apigenSeen) {
	t.Helper()
	var seen []apigenSeen
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		seen = append(seen, apigenSeen{r.Method, r.URL.RequestURI(), r.Header.Get("Authorization"), r.Header.Get("Content-Type"), body})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(env)
	}))
	t.Cleanup(srv.Close)
	return srv, &seen
}

func apigenOK(data any) map[string]any {
	return map[string]any{"data": data, "error": nil, "meta": map[string]any{"requestId": "req_ok"}}
}

func TestAPI_ListSendsQueryAndToken(t *testing.T) {
	srv, seen := apigenServer(t, 200, apigenOK([]any{map[string]any{"id": "ord_1"}}))
	c := New(Config{Token: "sk_live_test", BaseURL: srv.URL})
	data, err := c.API.OrdersList(context.Background(), &OrdersListArgs{Status: Ptr("requested"), Limit: Ptr(10)})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `[{"id":"ord_1"}]` {
		t.Fatalf("data = %s", data)
	}
	r := (*seen)[0]
	if r.method != "GET" || r.uri != "/api/v1/orders?limit=10&status=requested" || len(r.body) != 0 || r.auth != "Bearer sk_live_test" {
		t.Fatalf("request = %+v", r)
	}
}

func TestAPI_CreateSendsBody(t *testing.T) {
	srv, seen := apigenServer(t, 201, apigenOK(map[string]any{"id": "svc_1"}))
	c := New(Config{Token: "sk_live_test", BaseURL: srv.URL})
	_, err := c.API.ServicesCreate(context.Background(), &ServicesCreateArgs{
		Slug: "logo-design", Name: "Logo design", PricingType: "fixed", PriceIdr: Ptr(1500000), Tags: []string{"design"},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := (*seen)[0]
	if r.method != "POST" || r.uri != "/api/v1/services" || r.auth != "Bearer sk_live_test" || r.contentType != "application/json" {
		t.Fatalf("request = %+v", r)
	}
	if string(r.body) != `{"name":"Logo design","priceIdr":1500000,"pricingType":"fixed","slug":"logo-design","tags":["design"]}` {
		t.Fatalf("body = %s", r.body)
	}
}

func TestAPI_UpdateWithPathParam(t *testing.T) {
	srv, seen := apigenServer(t, 200, apigenOK(map[string]any{"id": "ord_1"}))
	c := New(Config{Token: "sk_live_test", BaseURL: srv.URL})
	if _, err := c.API.OrdersUpdate(context.Background(), "ord 1", &OrdersUpdateArgs{Status: Ptr("confirmed")}); err != nil {
		t.Fatal(err)
	}
	r := (*seen)[0]
	if r.method != "PATCH" || r.uri != "/api/v1/orders/ord%201" || string(r.body) != `{"status":"confirmed"}` {
		t.Fatalf("request = %+v", r)
	}
}

func TestAPI_PublicRouteSendsNoToken(t *testing.T) {
	srv, seen := apigenServer(t, 201, apigenOK(map[string]any{"id": "msg_1"}))
	c := New(Config{Token: "sk_live_test", BaseURL: srv.URL})
	if _, err := c.API.PublicOrdersMessages(context.Background(), "tok_1", &PublicOrdersMessagesArgs{BodyField: "hello"}); err != nil {
		t.Fatal(err)
	}
	r := (*seen)[0]
	if r.uri != "/api/v1/public/orders/tok_1/messages" || r.auth != "" || string(r.body) != `{"body":"hello"}` {
		t.Fatalf("request = %+v", r)
	}
}

func TestAPI_RequiredFieldMissing(t *testing.T) {
	srv, seen := apigenServer(t, 200, apigenOK(nil))
	c := New(Config{Token: "sk_live_test", BaseURL: srv.URL})
	_, err := c.API.ServicesCreate(context.Background(), &ServicesCreateArgs{Slug: "x", Name: "X"})
	if err == nil || !strings.Contains(err.Error(), "PricingType") {
		t.Fatalf("err = %v, want a missing PricingType", err)
	}
	if len(*seen) != 0 {
		t.Fatalf("sent %d requests", len(*seen))
	}
}

func TestAPI_ErrorEnvelope(t *testing.T) {
	srv, _ := apigenServer(t, 404, map[string]any{"data": nil, "error": map[string]any{"code": "NOT_FOUND", "message": "nope"}, "meta": map[string]any{"requestId": "req_9"}})
	_, err := New(Config{Token: "sk_live_test", BaseURL: srv.URL}).API.OrdersGet(context.Background(), "ord_x")
	e, ok := err.(*Error)
	if !ok || e.Status != 404 || e.Code != "NOT_FOUND" || e.RequestID != "req_9" {
		t.Fatalf("err = %#v", err)
	}
}
