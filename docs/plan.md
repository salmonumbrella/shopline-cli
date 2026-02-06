# Shopline CLI: 100% Coverage + Agent-Friendly Plan

This plan is the living checklist for bringing `shopline-cli` to:

- 100% API coverage against Shopline Open API + Storefront API reference docs
- Agent-friendly ergonomics (discoverability, stable JSON, schema introspection, jq filtering, predictable IDs)

## Source Of Truth

- Shopline Open API reference (mirrored locally via Firecrawl)
  - URL lists: `docs/shopline-openapi/urls_endpoints.txt`, `docs/shopline-openapi/urls_non_endpoints.txt`
  - Local mirror fetch: `scripts/download_shopline_openapi_docs.py`
    - Important: default now uses `onlyMainContent=false` so endpoint URL + method are present for coverage indexing.

## Current Status (Snapshot)

- Docs mirror:
  - Endpoint pages mirrored: 308/308 (local, not committed)
  - Non-endpoint pages: best-effort (some pages are too large / time out)
- CLI agent-friendliness already includes:
  - `--output json`, `--query` (jq), structured IDs in text tables
  - `--items-only` for JSON list commands (unwrap `items`)
  - `orders get --expand customer,products`
  - Order items surfaced correctly via `subtotal_items` => `line_items`

## Guiding Principles (Borrowed From `chatwoot-cli`)

- Fast discovery:
  - `schema list/show` for resource schemas
  - `help --json` for machine-readable CLI help
- Stable machine output:
  - JSON output is consistent across list/get
  - Avoid `null` collections where iteration is common (prefer `[]`)
  - Provide flags to unwrap envelopes (`--items-only`) and to filter output (`--query`)
- “Agent workflow” first:
  - Good defaults + obvious escape hatches
  - Predictable IDs that can be copy/pasted back into commands
  - Expansion flags for common joins (customer, products, line items, metafields)

## Workstreams

### 1) Coverage Inventory (Docs -> Endpoint List)

Goal: generate a canonical list of documented endpoints and compare it to implemented API client + cobra commands.

Tasks:

1. Build an endpoint indexer that reads the mirrored docs pages and outputs a normalized endpoint catalog.
   - Implemented: `go run ./cmd/shopline-coverage`
   - Outputs:
     - `docs/coverage/openapi_endpoints.json`
     - `docs/coverage/code_endpoints.json`
     - `docs/coverage/report.md`
2. Iterate on doc mirroring until parsing is stable for ~100% of endpoint pages.
   - Current state: coverage parsing expects "full" scrapes (firecrawl `onlyMainContent=false`).
3. Extend coverage checker to also account for CLI command coverage (later).

Tests:

- Unit tests for parser normalization (method/path extraction, path param detection)
- Golden tests for report generation (stable ordering)

### 2) Fill API Gaps (internal/api)

Goal: implement missing endpoints in `internal/api` with strong types and tests.

Approach:

- Add endpoints group-by-group (e.g. “Orders”, “Metafields”, “Storefront Carts”, etc.)
- For each group:
  - Add/extend types in `internal/api/*.go`
  - Add methods on `Client`
  - Update `APIClient` interface + `MockClient` (regen if needed)
  - Add `*_test.go` coverage (httptest server verifying path + query + request body)

Tests:

- Request URL/query assertions
- Request body round-trip assertions
- Response shape decoding assertions

### 3) Fill CLI Gaps (internal/cmd)

Goal: every API method that makes sense to expose has a CLI command.

Approach:

- Create resource commands with:
  - `list/get/create/update/delete` (as applicable)
  - consistent flags: `--page`, `--page-size`, `--limit`, `--sort-by`, `--desc`, and resource-specific filters
  - JSON output envelope behavior consistent with other resources

Tests:

- Cobra RunE tests with mock API client (pattern already in repo)
- JSON output assertions

### 4) Agent-Friendly Enhancements

Goal: “I (and ChatGPT) love using this CLI”.

Candidate improvements (in priority order):

1. `help --json` parity
   - Machine-readable help output for every command/flag
2. `schema` improvements
   - Ensure every resource registers schema + ID prefix
   - Add “field presets” like `chatwoot-cli --fields minimal/default/debug` (optional)
3. Output presets
   - For high-traffic resources (orders, products, customers): add `--fields` shortcut for common slices
4. Expansion model
   - Standardize `--expand` across resources
   - Add expansions for metafields and nested relationships
5. Production-safe batching
   - JSONL input mode for bulk operations where it’s safe

Tests:

- Flags exist + precedence (`--jq` alias optional)
- Schema JSON stability tests

## Phase Plan (Milestones)

### Phase 0: Tooling + Docs Mirror

- [x] Extract doc URLs from reference sidebar HTML
- [x] Mirror endpoint pages via Firecrawl to plaintext markdown
- [x] Add endpoint indexer + coverage report generator (`cmd/shopline-coverage`)

Run it locally:

```bash
# Refresh docs mirror (full pages; required for endpoint URL + method parsing)
./scripts/download_shopline_openapi_docs.py --urls docs/shopline-openapi/urls_endpoints.txt --force

# Generate coverage report
go run ./cmd/shopline-coverage
```

### Phase 1: Orders + Order Items + Metafields (Complete the Core Commerce Loop)

- [x] Ensure all orders endpoints from docs exist in `internal/api/orders.go`
- [x] Add missing order item metafields endpoints
- [x] Add CLI commands for those endpoints

### Phase 2: Products + Inventory + Pricing

- [ ] Fill product endpoints (variations, stocks, tags, etc.)
- [ ] Ensure product JSON types decode translation fields consistently
- [ ] CLI coverage + tests

### Phase 3: Customers + Credits + Membership/Points

- [ ] Fill customer endpoints
- [ ] Ensure store credits + user credits endpoints match docs
- [ ] CLI coverage + tests

### Phase 4: Storefront API (Carts, Tokens, OAuth)

- [ ] Fill storefront carts endpoints
- [ ] Add cart item metafields endpoints
- [ ] CLI coverage + tests

### Phase 5: Everything Else + Polish

- [ ] Channels, staff permissions, conversations, etc.
- [ ] Agent-friendly improvements (help/schema/fields/expand)
- [ ] Final coverage report is 100%

## Tracking

We’ll maintain these artifacts:

- `docs/coverage/openapi_endpoints.json` (from docs mirror)
- `docs/coverage/code_endpoints.json` (from code scan)
- `docs/coverage/report.md` (what’s missing)
- `docs/coverage/progress.md` (checklist by resource)
