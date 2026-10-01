# Changelog

## 0.3.0
- A route read by id next to its list is named `get` + the list's name: `client.API.ClientGetOrders` (was `client.API.ClientOrders2`), `client.API.FulfillmentGetDeliveries` (was `client.API.FulfillmentDeliveries2`), `client.API.FulfillmentGetShipments` (was `client.API.FulfillmentShipments2`), `client.API.PublicGetStorefrontBlog` (was `client.API.PublicStorefrontBlog2`). Each old name stays as a deprecated alias.
- Query fields the API refuses a request without are now required: `key` on GET /api/v1/fulfillment/licenses/validate.

