package serront

// Order statuses (the work side). requested → confirmed|declined,
// confirmed → in_progress → delivered → completed; canceled from any
// pre-completed, non-terminal state (seller action).
const (
	StatusRequested  = "requested"
	StatusConfirmed  = "confirmed"
	StatusDeclined   = "declined"
	StatusInProgress = "in_progress"
	StatusDelivered  = "delivered"
	StatusCompleted  = "completed"
	StatusCanceled   = "canceled"
)

// Payment statuses (the money side): unpaid → payment_claimed (buyer)
// → payment_confirmed (seller; also directly from unpaid).
const (
	PaymentUnpaid    = "unpaid"
	PaymentClaimed   = "payment_claimed"
	PaymentConfirmed = "payment_confirmed"
)

// Pricing types. fixed | hourly carry a base PriceIDR; package carries
// 1-5 packages.
const (
	PricingFixed   = "fixed"
	PricingHourly  = "hourly"
	PricingPackage = "package"
)

// Billing tiers.
const (
	TierFree     = "free"
	TierStarter  = "starter"
	TierGrowth   = "growth"
	TierBusiness = "business"
)

// Payout statuses.
const (
	PayoutPending   = "pending"
	PayoutInTransit = "in_transit"
	PayoutPaid      = "paid"
	PayoutFailed    = "failed"
	PayoutCancelled = "cancelled"
)

// BankAccount is one manual bank-transfer account shown to buyers.
type BankAccount struct {
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountHolder string `json:"accountHolder"`
}

// StorefrontSettings is the seller's storefront profile.
type StorefrontSettings struct {
	AccountID string `json:"accountId"`
	// Slug is the public URL slug — globally unique across storefronts.
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio"`
	// WhatsappNumber is E.164 — rendered as a wa.me link on buyer pages.
	WhatsappNumber     *string       `json:"whatsappNumber"`
	ManualBankAccounts []BankAccount `json:"manualBankAccounts"`
	ManualInstructions *string       `json:"manualInstructions"`
	// HideBranding hides "Powered by Serront" — paid perk (Starter+).
	HideBranding bool   `json:"hideBranding"`
	Published    bool   `json:"published"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// ServicePackage is one package of a package-priced service.
type ServicePackage struct {
	Name string `json:"name"`
	// PriceIDR is whole rupiah.
	PriceIDR    int    `json:"priceIdr"`
	Description string `json:"description,omitempty"`
}

// Service is one catalog entry on the storefront.
type Service struct {
	ID        string `json:"id"`
	AccountID string `json:"accountId"`
	// Slug is unique per storefront.
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PricingType string `json:"pricingType"` // fixed | hourly | package
	// PriceIDR is the base price (fixed/hourly); nil for package services.
	PriceIDR *int `json:"priceIdr"`
	// Packages carries 1-5 packages (package services); nil otherwise.
	Packages  []ServicePackage `json:"packages"`
	Active    bool             `json:"active"`
	SortOrder int              `json:"sortOrder"`
	CreatedAt string           `json:"createdAt"`
	UpdatedAt string           `json:"updatedAt"`
}

// OrderService is the slim service shape embedded on order responses.
type OrderService struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	PricingType string `json:"pricingType"`
}

// Order is one order on the seller's desk.
type Order struct {
	ID        string `json:"id"`
	AccountID string `json:"accountId"`
	// Number is the per-seller order counter (#1, #2, …).
	Number        int     `json:"number"`
	ServiceID     string  `json:"serviceId"`
	ClientID      *string `json:"clientId"`
	Status        string  `json:"status"`
	PaymentStatus string  `json:"paymentStatus"`
	// PackageName is the chosen package for package-priced services.
	PackageName   *string `json:"packageName"`
	BuyerName     string  `json:"buyerName"`
	BuyerEmail    string  `json:"buyerEmail"`
	BuyerPhone    *string `json:"buyerPhone"`
	PreferredDate *string `json:"preferredDate"`
	Notes         string  `json:"notes"`
	// QuotedPriceIDR is the quote (whole rupiah) — defaulted from the
	// service price, seller-adjustable via Orders.Update.
	QuotedPriceIDR    int     `json:"quotedPriceIdr"`
	DiscountCode      *string `json:"discountCode"`
	DiscountAmountIDR int     `json:"discountAmountIdr"`
	// AccessToken is the buyer's tokenized-link credential (/o/<token>).
	AccessToken string `json:"accessToken"`
	// HasProof: the buyer uploaded a payment proof — fetch it via
	// Orders.DownloadProof.
	HasProof      bool         `json:"hasProof"`
	LastMessageAt string       `json:"lastMessageAt"`
	CreatedAt     string       `json:"createdAt"`
	UpdatedAt     string       `json:"updatedAt"`
	Service       OrderService `json:"service"`
}

// OrderMessage is one message on an order's thread.
type OrderMessage struct {
	ID         string  `json:"id"`
	OrderID    string  `json:"orderId"`
	AuthorType string  `json:"authorType"` // "seller" | "buyer"
	AuthorName *string `json:"authorName"`
	Body       string  `json:"body"`
	// IsInternal marks a seller-only internal note — never shown to the
	// buyer.
	IsInternal bool   `json:"isInternal"`
	CreatedAt  string `json:"createdAt"`
}

// OrderWithMessages is an order including its full thread (internal
// notes included — this is the seller surface).
type OrderWithMessages struct {
	Order
	Messages []OrderMessage `json:"messages"`
}

// OrderList is the list response: orders plus per-status counts for
// the whole workspace and a keyset cursor for the next page.
type OrderList struct {
	Orders []Order        `json:"orders"`
	Counts map[string]int `json:"counts"`
	// Cursor is the opaque cursor for the next page (nil on the last page).
	Cursor  *string `json:"cursor"`
	HasMore bool    `json:"hasMore"`
}

// ProofDownload is the buyer's payment-proof image bytes.
type ProofDownload struct {
	Data        []byte
	ContentType string
}

// ModulesState is the per-module enabled flags.
type ModulesState struct {
	Payment   bool `json:"payment,omitempty"`
	Marketing bool `json:"marketing,omitempty"`
}

// ModulesView is the GET/PATCH /api/v1/modules response.
type ModulesView struct {
	Modules ModulesState `json:"modules"`
	// PlugipayAccountID is the provisioned Plugipay workspace id
	// (Payment module), if any.
	PlugipayAccountID *string `json:"plugipayAccountId"`
	// RipploAccountID is the provisioned Ripllo workspace id
	// (Marketing module), if any.
	RipploAccountID *string `json:"ripploAccountId"`
	// Available: platform creds present — modules can actually provision.
	Available struct {
		Payment   bool `json:"payment"`
		Marketing bool `json:"marketing"`
	} `json:"available"`
}

// LedgerBalanceRow is one per-code balance row from Plugipay.
type LedgerBalanceRow struct {
	Code     string `json:"code"`
	Balance  int    `json:"balance"`
	Currency string `json:"currency"`
}

// LedgerBalance is the GET /api/v1/ledger/balance response.
type LedgerBalance struct {
	// Balance is the sum across the per-code rows.
	Balance  int                `json:"balance"`
	Currency string             `json:"currency"`
	ByCode   []LedgerBalanceRow `json:"byCode"`
}

// LedgerEntry is one ledger entry (Plugipay-held).
type LedgerEntry struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Amount      int     `json:"amount"`
	Currency    string  `json:"currency"`
	SourceType  *string `json:"sourceType"`
	SourceID    *string `json:"sourceId"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"createdAt"`
}

// LedgerEntryList is the cursor-paged entries response.
type LedgerEntryList struct {
	Entries []LedgerEntry `json:"entries"`
	Cursor  *string       `json:"cursor"`
	HasMore bool          `json:"hasMore"`
}

// Payout is one payout (Plugipay-held; manual
// pending → in_transit → paid state machine).
type Payout struct {
	ID                string  `json:"id"`
	Amount            int     `json:"amount"`
	Currency          string  `json:"currency"`
	Status            string  `json:"status"`
	BankCode          *string `json:"bankCode"`
	BankName          *string `json:"bankName"`
	BankAccountNumber *string `json:"bankAccountNumber"`
	BankAccountHolder *string `json:"bankAccountHolder"`
	Note              *string `json:"note"`
	Reference         *string `json:"reference"`
	FailureReason     *string `json:"failureReason"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

// PayoutList is the cursor-paged payout history.
type PayoutList struct {
	Payouts []Payout `json:"payouts"`
	Cursor  *string  `json:"cursor"`
	HasMore bool     `json:"hasMore"`
}

// PayoutBalance is the GET /api/v1/payouts/balance response.
type PayoutBalance struct {
	// Available = ledger balance − in-flight payouts.
	Available int    `json:"available"`
	Pending   int    `json:"pending"`
	Currency  string `json:"currency"`
}

// PayoutBankAccount is the merchant's default payout bank details.
type PayoutBankAccount struct {
	BankCode          *string `json:"bankCode"`
	BankName          *string `json:"bankName"`
	BankAccountNumber *string `json:"bankAccountNumber"`
	BankAccountHolder *string `json:"bankAccountHolder"`
}

// APIKeySummary is the display-safe API-key shape (no hashes).
type APIKeySummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// KeyPrefix is the display-safe prefix ("sk_live_" + 4 hex chars).
	KeyPrefix  string  `json:"keyPrefix"`
	LastUsedAt *string `json:"lastUsedAt"`
	CreatedAt  string  `json:"createdAt"`
}

// APIKeyCreated is the create response — Key (the plaintext) is
// returned ONCE here and never again.
type APIKeyCreated struct {
	APIKeySummary
	Key string `json:"key"`
}

// WebhookSubscription is one outbound webhook endpoint receiving
// serront.order.* events.
type WebhookSubscription struct {
	ID  string `json:"id"`
	URL string `json:"url"`
	// Events is "*" or versioned serront event types
	// (serront.order.created.v1, …).
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

// WebhookSubscriptionCreated is the create response — Secret (the
// signing secret) is returned ONCE here and never again.
type WebhookSubscriptionCreated struct {
	WebhookSubscription
	Secret string `json:"secret"`
}

// ─── Public (buyer-facing) types ─────────────────────────────────────

// PublicStorefrontProfile is the published seller profile.
type PublicStorefrontProfile struct {
	Slug           string  `json:"slug"`
	DisplayName    string  `json:"displayName"`
	Bio            string  `json:"bio"`
	WhatsappNumber *string `json:"whatsappNumber"`
	HideBranding   bool    `json:"hideBranding"`
	// OnlinePayment: Payment module on — buyers can pay orders online.
	OnlinePayment bool `json:"onlinePayment"`
	// DiscountCodes: Marketing module on — the order form accepts
	// discount codes.
	DiscountCodes bool `json:"discountCodes"`
}

// PublicService is one active service as shown to buyers.
type PublicService struct {
	ID          string           `json:"id"`
	Slug        string           `json:"slug"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	PricingType string           `json:"pricingType"`
	PriceIDR    *int             `json:"priceIdr"`
	Packages    []ServicePackage `json:"packages"`
}

// PublicStorefrontView is the GET /api/v1/public/storefront/:slug
// response.
type PublicStorefrontView struct {
	Storefront PublicStorefrontProfile `json:"storefront"`
	Services   []PublicService         `json:"services"`
}

// PublicOrderCreated carries the buyer's only credential: the access
// token for the tokenized order page.
type PublicOrderCreated struct {
	Number            int    `json:"number"`
	AccessToken       string `json:"accessToken"`
	QuotedPriceIDR    int    `json:"quotedPriceIdr"`
	DiscountAmountIDR int    `json:"discountAmountIdr"`
}

// DiscountValidation is the validate-discount dry-run result.
type DiscountValidation struct {
	Valid             bool    `json:"valid"`
	Reason            *string `json:"reason"`
	DiscountAmountIDR int     `json:"discountAmountIdr"`
	// QuotedPriceIDR is the net price after the discount (the gross
	// price when invalid).
	QuotedPriceIDR int `json:"quotedPriceIdr"`
	Code           *struct {
		Code        string  `json:"code"`
		Type        string  `json:"type"`
		Description *string `json:"description"`
	} `json:"code"`
}

// PublicPaymentInstructions are the manual-transfer instructions shown
// while payment is not yet confirmed.
type PublicPaymentInstructions struct {
	BankAccounts   []BankAccount `json:"bankAccounts"`
	Instructions   *string       `json:"instructions"`
	WhatsappNumber *string       `json:"whatsappNumber"`
	WhatsappLink   *string       `json:"whatsappLink"`
}

// PublicOrderMessage is one public message in the buyer view.
type PublicOrderMessage struct {
	ID         string  `json:"id"`
	AuthorType string  `json:"authorType"`
	AuthorName *string `json:"authorName"`
	Body       string  `json:"body"`
	CreatedAt  string  `json:"createdAt"`
}

// PublicOrderView is the tokenized buyer-facing order view — public
// messages only, internal notes are never exposed.
type PublicOrderView struct {
	Number            int     `json:"number"`
	Status            string  `json:"status"`
	PaymentStatus     string  `json:"paymentStatus"`
	QuotedPriceIDR    int     `json:"quotedPriceIdr"`
	DiscountCode      *string `json:"discountCode"`
	DiscountAmountIDR int     `json:"discountAmountIdr"`
	// OnlinePayment: Payment module on — Public.Pay mints a hosted
	// checkout.
	OnlinePayment bool    `json:"onlinePayment"`
	PackageName   *string `json:"packageName"`
	BuyerName     string  `json:"buyerName"`
	PreferredDate *string `json:"preferredDate"`
	Notes         string  `json:"notes"`
	CreatedAt     string  `json:"createdAt"`
	Service       struct {
		Slug        string `json:"slug"`
		Name        string `json:"name"`
		PricingType string `json:"pricingType"`
	} `json:"service"`
	Storefront *struct {
		Slug         string `json:"slug"`
		DisplayName  string `json:"displayName"`
		HideBranding bool   `json:"hideBranding"`
	} `json:"storefront"`
	// Payment carries manual-transfer instructions — nil once payment
	// is confirmed.
	Payment  *PublicPaymentInstructions `json:"payment"`
	HasProof bool                       `json:"hasProof"`
	Messages []PublicOrderMessage       `json:"messages"`
}
