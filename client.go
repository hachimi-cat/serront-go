// Package serront is the Go SDK for the serront.com services-storefront
// REST API. Sister to @forjio/serront (JS) and forjio-serront (Python).
//
// Auth = Bearer token — an sk_live_… API key from the dashboard (or a
// Huudis-minted access token). Pass Config.Token or set SERRONT_TOKEN.
// The Public resource (buyer-facing storefront + tokenized order
// endpoints) needs no token at all.
//
// Every response rides the Forjio envelope {data, error, meta}; the
// client unwraps it and returns *Error (carrying the envelope's
// error.code) on failure.
package serront

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client is the Serront typed client.
type Client struct {
	token   string
	baseURL string
	httpc   *http.Client

	// Resource namespaces — mirror the JS + Python SDKs.
	Storefront           *StorefrontResource
	Services             *ServicesResource
	Orders               *OrdersResource
	Modules              *ModulesResource
	Ledger               *LedgerResource
	Payouts              *PayoutsResource
	APIKeys              *APIKeysResource
	WebhookSubscriptions *WebhookSubscriptionsResource
	Billing              *BillingResource
	Public               *PublicResource

	// API has every feature route, one method each (generated from the API
	// spec: api_generated.go), Bearer-authenticated like every other call.
	API *GeneratedAPI
}

// Config holds the credentials + endpoint overrides.
type Config struct {
	// Token is the Bearer token (an sk_live_… API key or a Huudis-minted
	// JWT). Defaults to the SERRONT_TOKEN env var. Optional — Public
	// works without it.
	Token string
	// BaseURL overrides the API base. Default: https://serront.com.
	BaseURL string
	// HTTP overrides the http.Client. Default: 30s timeout.
	HTTP *http.Client
}

// New constructs a Serront client.
//
// Example:
//
//	c := serront.New(serront.Config{Token: os.Getenv("SERRONT_TOKEN")})
//	page, err := c.Orders.List(ctx, &serront.OrderListParams{Status: "requested"})
func New(cfg Config) *Client {
	token := cfg.Token
	if token == "" {
		token = os.Getenv("SERRONT_TOKEN")
	}
	base := cfg.BaseURL
	if base == "" {
		base = "https://serront.com"
	}
	httpc := cfg.HTTP
	if httpc == nil {
		httpc = &http.Client{Timeout: 30 * time.Second}
	}
	c := &Client{
		token:   token,
		baseURL: strings.TrimRight(base, "/"),
		httpc:   httpc,
	}
	c.Storefront = &StorefrontResource{c: c}
	c.Services = &ServicesResource{c: c}
	c.Orders = &OrdersResource{c: c}
	c.Modules = &ModulesResource{c: c}
	c.Ledger = &LedgerResource{c: c}
	c.Payouts = &PayoutsResource{c: c}
	c.APIKeys = &APIKeysResource{c: c}
	c.WebhookSubscriptions = &WebhookSubscriptionsResource{c: c}
	c.Billing = &BillingResource{c: c}
	c.Public = &PublicResource{c: c}
	c.API = &GeneratedAPI{c: c}
	return c
}

// apigenRequest is the call behind Client.API (api_generated.go):
// Bearer-authenticated like every other request, with the same envelope
// handling; the buyer-facing /api/v1/public/* routes go without a token,
// as Client.Public does.
func (c *Client) apigenRequest(ctx context.Context, method, path string, query url.Values, body map[string]any) (json.RawMessage, error) {
	var send any
	if body != nil {
		send = body
	}
	var out json.RawMessage
	if err := c.do(ctx, strings.ToUpper(method), path, query, send, strings.HasPrefix(path, "/api/v1/public/"), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// envelope mirrors the Forjio data/error/meta API envelope.
type envelope struct {
	Data  json.RawMessage `json:"data"`
	Error *envelopeError  `json:"error"`
	Meta  *envelopeMeta   `json:"meta"`
}

type envelopeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
	DocURL  string `json:"docUrl,omitempty"`
}

type envelopeMeta struct {
	RequestID string `json:"requestId,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// do builds the request, attaches the bearer (unless noAuth), parses
// the envelope, and decodes the data slot into out (pointer; nil to
// ignore the body).
func (c *Client) do(
	ctx context.Context,
	method, path string,
	query url.Values,
	body any,
	noAuth bool,
	out any,
) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return &Error{Status: 0, Code: "SERIALIZE_FAILED", Message: err.Error()}
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return &Error{Status: 0, Code: "REQUEST_BUILD_FAILED", Message: err.Error()}
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if !noAuth {
		if c.token == "" {
			return &Error{
				Status:  0,
				Code:    "AUTH_REQUIRED",
				Message: "no token configured: set Config.Token or SERRONT_TOKEN",
			}
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.httpc.Do(req)
	if err != nil {
		return &Error{Status: 0, Code: "NETWORK_ERROR", Message: err.Error()}
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return &Error{
			Status:  res.StatusCode,
			Code:    "INVALID_RESPONSE",
			Message: fmt.Sprintf("non-JSON response (HTTP %d)", res.StatusCode),
		}
	}

	requestID := ""
	if env.Meta != nil {
		requestID = env.Meta.RequestID
	}
	if res.StatusCode >= 400 || env.Error != nil {
		code := "UNKNOWN"
		message := fmt.Sprintf("HTTP %d", res.StatusCode)
		param := ""
		if env.Error != nil {
			if env.Error.Code != "" {
				code = env.Error.Code
			}
			if env.Error.Message != "" {
				message = env.Error.Message
			}
			param = env.Error.Param
		}
		return &Error{
			Status:    res.StatusCode,
			Code:      code,
			Message:   message,
			RequestID: requestID,
			Param:     param,
		}
	}

	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return &Error{
				Status:    res.StatusCode,
				Code:      "DECODE_FAILED",
				Message:   err.Error(),
				RequestID: requestID,
			}
		}
	}
	return nil
}

// doRaw sends a raw (non-JSON) request body — the proof upload path.
func (c *Client) doRaw(
	ctx context.Context,
	method, path string,
	data []byte,
	contentType string,
	noAuth bool,
	out any,
) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return &Error{Status: 0, Code: "REQUEST_BUILD_FAILED", Message: err.Error()}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)
	if !noAuth {
		if c.token == "" {
			return &Error{
				Status:  0,
				Code:    "AUTH_REQUIRED",
				Message: "no token configured: set Config.Token or SERRONT_TOKEN",
			}
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.httpc.Do(req)
	if err != nil {
		return &Error{Status: 0, Code: "NETWORK_ERROR", Message: err.Error()}
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return &Error{
			Status:  res.StatusCode,
			Code:    "INVALID_RESPONSE",
			Message: fmt.Sprintf("non-JSON response (HTTP %d)", res.StatusCode),
		}
	}
	requestID := ""
	if env.Meta != nil {
		requestID = env.Meta.RequestID
	}
	if res.StatusCode >= 400 || env.Error != nil {
		code := "UNKNOWN"
		message := fmt.Sprintf("HTTP %d", res.StatusCode)
		param := ""
		if env.Error != nil {
			if env.Error.Code != "" {
				code = env.Error.Code
			}
			if env.Error.Message != "" {
				message = env.Error.Message
			}
			param = env.Error.Param
		}
		return &Error{Status: res.StatusCode, Code: code, Message: message, RequestID: requestID, Param: param}
	}
	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return &Error{Status: res.StatusCode, Code: "DECODE_FAILED", Message: err.Error(), RequestID: requestID}
		}
	}
	return nil
}

// download GETs a binary route (the proof download). Error responses
// still ride the JSON envelope.
func (c *Client) download(ctx context.Context, path string) (*ProofDownload, error) {
	if c.token == "" {
		return nil, &Error{
			Status:  0,
			Code:    "AUTH_REQUIRED",
			Message: "no token configured: set Config.Token or SERRONT_TOKEN",
		}
	}
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return nil, &Error{Status: 0, Code: "REQUEST_BUILD_FAILED", Message: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpc.Do(req)
	if err != nil {
		return nil, &Error{Status: 0, Code: "NETWORK_ERROR", Message: err.Error()}
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)

	if res.StatusCode >= 400 {
		var env envelope
		code, message, requestID, param := "UNKNOWN", fmt.Sprintf("HTTP %d", res.StatusCode), "", ""
		if err := json.Unmarshal(raw, &env); err == nil {
			if env.Meta != nil {
				requestID = env.Meta.RequestID
			}
			if env.Error != nil {
				if env.Error.Code != "" {
					code = env.Error.Code
				}
				if env.Error.Message != "" {
					message = env.Error.Message
				}
				param = env.Error.Param
			}
		}
		return nil, &Error{Status: res.StatusCode, Code: code, Message: message, RequestID: requestID, Param: param}
	}

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return &ProofDownload{Data: raw, ContentType: contentType}, nil
}
