# Coverage Progress

This is a human-maintained checklist that we update as we close gaps surfaced by `docs/coverage/report.md`.

How to refresh the report:

```bash
./scripts/download_shopline_openapi_docs.py --urls docs/shopline-openapi/urls_endpoints.txt --force
go run ./cmd/shopline-coverage
```

## Open API

- [x] Orders
- [x] Order metafields
- [x] Products (endpoints wired; types still partially raw for metafields/stocks)
- [x] Customers
- Customers core: list/get/search/create/update/delete + tags + subscriptions + LINE lookup (done)
- Customer metafields + app_metafields (done)
- Customer coupon promotions (`GET /customers/{id}/coupon_promotions`) (done)
- [x] Customer groups
- Customer groups core: list/get/create/delete (done)
- Customer group children: `GET /customer_groups/{parent_id}/customer_group_children` + `GET /customer_groups/{parent_id}/customer_group_children/{id}/customer_ids` (done)
- [x] Delivery options
- Delivery options core: list/get (done)
- Delivery config: `GET /delivery_options/delivery_config` (done)
- Delivery time slots (documented): `GET /delivery_options/{delivery_option_id}/delivery_time_slots` (done)
- Delivery stores info update: `PUT /delivery_options/{id}/stores_info` (done)
- [x] Flash price campaigns
- `GET/POST/GET(id)/PUT(id)/DELETE(id) /flash_price_campaigns` (done)
- [x] Store credits
- Customer store credits (`GET/POST /customers/{id}/store_credits`) + user credits (`GET /user_credits`, `POST /user_credits/bulk_update`) (done)
- Member points + membership info (`/customers/membership_info`, `/customers/{id}/member_points`, `/member_point_rules`, `/member_points/bulk_update`, `/customers/{id}/membership_tier/action_logs`) (done)
- [x] Channels
- [x] Staffs
- Staff permissions (`GET /staffs/{id}/permissions`) (done)
- [x] Carts
- Cart items (`POST/PATCH/DELETE /carts/{cart_id}/items`) + prepare/exchange + item metafields/app_metafields (done)
- [x] Settings
- /settings/* documented endpoints (checkout/domains/layouts/theme/users/etc.) (done)
- [x] Merchant metafields
- `/merchants/current/metafields` + `/merchants/current/app_metafields` (done)
- [x] Merchants (documented endpoints)
- `GET /merchants/{merchant_id}` + `POST /merchants/generate_express_link` (done)
- [x] Webhooks
- Includes `PUT /webhooks/{id}` (done)
- [x] Token info
- `GET /token/info` (done)
- [x] Storefront OAuth applications
- `GET/POST/GET(id)/DELETE /storefront/oauth_applications` (done)
- [x] Multipass (documented endpoints)
- `GET/POST /multipass/secret` + `GET /multipass/linkings` + `POST/DELETE /multipass/customers/{customer_id}/linkings` (done)
- [x] Wish list items
- `GET/POST/DELETE /wish_list_items` (done)
- [x] User coupons (docs endpoints)
- `GET /user_coupons/list` + `POST /user_coupons/{coupon_code}/claim` + `POST /user_coupons/{coupon_code}/redeem` (done)
- [x] Promotions (docs endpoints)
- `GET /promotions/coupon-center` (done)
- [x] Media (docs endpoints)
- `POST /media` (done)
- [x] Sales (docs endpoints)
- `POST /sales/{saleId}/delete_products` (done)
- [x] Affiliate campaigns (docs endpoints)
- `GET /affiliate_campaigns/{id}/orders` + `GET /affiliate_campaigns/{id}/summary` + `GET /affiliate_campaigns/{id}/get_products_sales_ranking` + `POST /affiliate_campaigns/{id}/export_report` (done)
- [x] Conversations (docs endpoints)
- `POST /conversations/message` (done)
- [x] Gifts (docs endpoints)
- `PUT /gifts/{id}` + `PUT /gifts/{id}/update_quantity` + `PUT /gifts/update_quantity` + `GET/PUT /gifts/{id}/stocks` (done)
- [x] POS purchase orders (docs endpoints)
- `GET/POST /pos/purchase_orders` + `GET/PUT /pos/purchase_orders/{id}` + `PUT /pos/purchase_orders/bulk_delete` + `POST /pos/purchase_orders/{id}/child` (done)

## Storefront API

- [x] Carts
- [x] Storefront tokens
- [x] Storefront OAuth applications
