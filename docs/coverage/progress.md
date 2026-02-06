# Coverage Progress

This is a human-maintained checklist that we update as we close gaps surfaced by `docs/coverage/report.md`.

How to refresh the report:

```bash
./scripts/download_shopline_openapi_docs.py --urls docs/shopline-openapi/urls_endpoints.txt --force
go run ./cmd/shopline-coverage
```

## Open API

- [ ] Orders
- [ ] Order metafields
- [x] Products (endpoints wired; types still partially raw for metafields/stocks)
- [ ] Customers
- [ ] Store credits
- [ ] Channels
- [ ] Staffs
- [ ] Webhooks

## Storefront API

- [ ] Carts
- [ ] Storefront tokens
- [ ] Storefront OAuth applications
