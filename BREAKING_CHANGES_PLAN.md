# Shopline API breaking changes plan (Jan 2026)

## Scope
This plan covers the breaking changes listed on the Shopline Open API “coming soon” page for January 22 and January 26, 2026, and any small CLI-facing improvements that follow from them.

## Plan
- [x] Review the breaking-change list and map each item to existing CLI/API client surfaces.
- [x] Customer API: add `subscriptions` support and surface `credit_balance` from Get Customer/Get Customers.
- [x] Customer CLI: show subscription status + credit balance in `customers get` text output.
- [x] Webhooks CLI: document the new duplicate topic+address restriction in `webhooks create` help text.
- [x] Tests: update API and CLI tests to cover new customer fields.
- [x] Sanity check: run focused tests for updated packages.
