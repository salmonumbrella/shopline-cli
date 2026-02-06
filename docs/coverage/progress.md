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
- [ ] Customers
- Customers core: list/get/search/create/update/delete + tags + subscriptions + LINE lookup (done)
- Customer metafields + app_metafields (done); credits/points pending
- [x] Store credits
- Customer store credits (`GET/POST /customers/{id}/store_credits`) + user credits (`GET /user_credits`, `POST /user_credits/bulk_update`) (done)
- [x] Channels
- [ ] Staffs
- [ ] Webhooks

## Storefront API

- [ ] Carts
- [ ] Storefront tokens
- [ ] Storefront OAuth applications
