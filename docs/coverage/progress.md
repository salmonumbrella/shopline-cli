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
- [ ] Webhooks

## Storefront API

- [x] Carts
- [ ] Storefront tokens
- [ ] Storefront OAuth applications
