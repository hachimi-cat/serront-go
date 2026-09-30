# serront-go

Typed Go client for the [serront.com](https://serront.com) services-storefront REST API.

Current version: **v0.2.0** — adds `Client.API`, every feature route
generated from the API spec. (v0.1.0: storefront, services, orders
(+ proof download), modules, ledger, payouts, API keys, webhook
subscriptions, billing and the public buyer surface.)

```bash
go get github.com/hachimi-cat/serront-go
```

```go
import (
	"context"
	"os"

	serront "github.com/hachimi-cat/serront-go"
)

// Bearer token from Config.Token or the SERRONT_TOKEN env var — use an
// sk_live_… API key from Dashboard → Settings → API keys.
c := serront.New(serront.Config{Token: "sk_live_xxx"})
ctx := context.Background()

// Storefront — the seller's public profile
settings, err := c.Storefront.Get(ctx) // nil until the first Put
_, err = c.Storefront.Put(ctx, serront.StorefrontPutInput{
	Slug:        "studio-renata",
	DisplayName: "Studio Renata",
	Bio:         "Brand & logo design studio.",
	Published:   true,
})

// Services — the catalog (fixed / hourly / package pricing)
services, err := c.Services.List(ctx)
svc, err := c.Services.Create(ctx, serront.ServiceCreateInput{
	Slug:        "logo-design",
	Name:        "Logo design",
	PricingType: serront.PricingPackage,
	Packages: []serront.ServicePackage{
		{Name: "Basic", PriceIDR: 500000},
		{Name: "Full identity", PriceIDR: 2500000},
	},
})
order1 := 1
_, err = c.Services.Update(ctx, svc.ID, serront.ServiceUpdateInput{SortOrder: &order1})

// Orders — the order desk
page, err := c.Orders.List(ctx, &serront.OrderListParams{
	Status:        serront.StatusRequested,
	PaymentStatus: serront.PaymentClaimed,
	Q:             "wedding",
})
order, err := c.Orders.Get(ctx, page.Orders[0].ID) // + full thread
_, err = c.Orders.Reply(ctx, order.ID, serront.OrderReplyInput{Body: "On it!"})
status, quote := serront.StatusConfirmed, 750000
_, err = c.Orders.Update(ctx, order.ID, serront.OrderUpdateInput{Status: &status, QuotedPriceIDR: &quote})
_, err = c.Orders.ConfirmPayment(ctx, order.ID)
proof, err := c.Orders.DownloadProof(ctx, order.ID) // {Data, ContentType}

// Modules — Payment (Plugipay) / Marketing (Ripllo) integrations
mods, err := c.Modules.Get(ctx)
enabled := true
_, err = c.Modules.Set(ctx, serront.ModulesSetInput{Payment: &enabled})

// Ledger + payouts (Plugipay-held; Payment module required)
balance, err := c.Ledger.Balance(ctx)
entries, err := c.Ledger.Entries(ctx, &serront.LedgerEntriesParams{Limit: 20})
payouts, err := c.Payouts.List(ctx, &serront.PayoutListParams{Status: serront.PayoutPending})
payout, err := c.Payouts.Create(ctx, serront.PayoutCreateInput{Amount: 1000000})
_, err = c.Payouts.Cancel(ctx, payout.ID)

// API keys + webhook subscriptions
keys, err := c.APIKeys.List(ctx)
created, err := c.APIKeys.Create(ctx, "ci") // created.Key shown ONCE
hook, err := c.WebhookSubscriptions.Create(ctx, serront.WebhookSubscriptionCreateInput{
	URL:    "https://example.com/serront-hook",
	Events: []string{"serront.order.created.v1"},
}) // hook.Secret shown ONCE

// Billing — current plan + upgrade checkout (tiers: free / starter / growth / business)
info, err := c.Billing.Get(ctx)
out, err := c.Billing.Checkout(ctx, serront.TierStarter) // redirect the browser to out.HostedURL

// Public (buyer) surface — no token required
view, err := c.Public.GetStorefront(ctx, "studio-renata")
placed, err := c.Public.CreateOrder(ctx, "studio-renata", serront.PublicOrderCreateInput{
	ServiceSlug: "logo-design",
	PackageName: "Basic",
	BuyerName:   "Budi",
	BuyerEmail:  "budi@example.com",
	DiscountCode: "WELCOME10", // Marketing module
})
check, err := c.Public.ValidateDiscount(ctx, "studio-renata", "WELCOME10", 500000)
myOrder, err := c.Public.GetOrder(ctx, placed.AccessToken)
_, err = c.Public.ReplyOrder(ctx, placed.AccessToken, "Any update?")
data, _ := os.ReadFile("transfer.png")
_, err = c.Public.UploadProof(ctx, placed.AccessToken, serront.ProofUploadInput{
	Data: data, ContentType: "image/png",
})
_, err = c.Public.ClaimPayment(ctx, placed.AccessToken) // "I have transferred"
pay, err := c.Public.Pay(ctx, placed.AccessToken)       // Payment module → HostedURL
```

## Endpoints covered

| Resource | Methods |
| --- | --- |
| `Storefront` | `Get`, `Put` (slug / display / bio / WhatsApp / bank accounts / branding / publish) |
| `Services` | `List`, `Create` (403 `LIMIT_REACHED` past the tier's service limit), `Get`, `Update`, `Delete` (409 while orders reference it) |
| `Orders` | `List` (Status / PaymentStatus / ServiceID / Q / Limit / Cursor + per-status counts), `Get`, `Update` (status and/or quote), `Reply` (with `IsInternal`), `ConfirmPayment`, `DownloadProof` |
| `Modules` | `Get`, `Set` (payment / marketing toggles; first enable provisions the partner workspace) |
| `Ledger` | `Balance`, `Entries` (cursor-paged; 409 `PAYMENT_MODULE_DISABLED` when off) |
| `Payouts` | `List`, `Create`, `Cancel`, `Balance`, `GetBankAccount`, `UpdateBankAccount`, `MarkInTransit`, `MarkPaid`, `MarkFailed` |
| `APIKeys` | `List`, `Create` (plaintext returned once), `Delete` |
| `WebhookSubscriptions` | `List`, `Create` (secret returned once), `Update` (active), `Delete` |
| `Billing` | `Get` (subscription + effectiveTier + tier table), `Checkout(tier)` → hosted checkout URL |
| `Public` | `GetStorefront`, `CreateOrder`, `ValidateDiscount`, `GetOrder`, `ReplyOrder`, `ClaimPayment`, `Pay`, `UploadProof` (no token) |
| `API` | Every feature route, one method each — generated from the API spec (`api_generated.go`) |

`c.API.<Area><Action>(ctx, pathParams…, *<Area><Action>Args)` covers every
route of the API, Bearer-authenticated like the resources above (the
`/api/v1/public/*` routes without a token), and returns the response's
`data` as `json.RawMessage` (routes that answer with a file or an event
stream — invoice PDFs, proofs, `…/stream` — are for the resources above).
Required fields are plain values, optional ones pointers (`serront.Ptr`),
slices or maps; `Body` passes the whole JSON body (a body field named `body`
is `BodyField`).

```go
data, err := c.API.OrdersUpdate(ctx, orderID, &serront.OrdersUpdateArgs{
	Status: serront.Ptr("confirmed"),
})
```

Errors return `*serront.Error` carrying the API envelope's `error.code`
(`NOT_FOUND`, `VALIDATION_ERROR`, `LIMIT_REACHED`, `UPGRADE_REQUIRED`,
`INVALID_TRANSITION`, …), the HTTP status, and the `meta.requestId`.

## Family

Sister to:
- [`@forjio/serront`](https://www.npmjs.com/package/@forjio/serront) (JS/TS)
- [`forjio-serront`](https://pypi.org/project/forjio-serront/) (Python)
