# Changelog

## 0.4.0
- Webhook deliveries are retried and logged. `client.API.WebhookSubscriptionsDeliveries(ctx, &WebhookSubscriptionsDeliveriesArgs{…})` lists them (newest first, with every attempt), `WebhookSubscriptionsGetDeliveries(ctx, id)` reads one, `WebhookSubscriptionsDeliveriesRetry(ctx, id)` sends one again (202; 409 when it is queued or its subscription is off), and `WebhookSubscriptionsEventTypes(ctx)` returns the event catalogue.
- `WebhookSubscriptionsUpdate` takes `URL` and `Events` too (it only took `Active`); `Active: true` re-enables a subscription Serront switched off for failing.
- `events` takes prefixes (`serront.order.*`). Newly listed (they were sent but undocumented): `serront.billing.canceled.v1`, `serront.shipping_credit.topped_up.v1`; new: `serront.webhook_subscription.disabled.v1`.

## 0.3.0
- A route read by id next to its list is named `get` + the list's name: `client.API.ClientGetOrders` (was `client.API.ClientOrders2`), `client.API.FulfillmentGetDeliveries` (was `client.API.FulfillmentDeliveries2`), `client.API.FulfillmentGetShipments` (was `client.API.FulfillmentShipments2`), `client.API.PublicGetStorefrontBlog` (was `client.API.PublicStorefrontBlog2`). Each old name stays as a deprecated alias.
- Query fields the API refuses a request without are now required: `key` on GET /api/v1/fulfillment/licenses/validate.

