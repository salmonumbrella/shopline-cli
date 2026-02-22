# Shopline CLI Agent Guide

## Quick Discovery
1. `spl schema` or `spl schema list` (add `--output json` for structured output)
2. `spl schema get <resource>` (add `--output json` for structured output)
3. `spl help --json`
   - Use `spl help --json --deep` for the full command tree in one call

## ID Conventions
- Text tables show IDs as `[resource:$id]` (copy/pasteable).
- Positional args and any `--*-id` flags accept `[resource:$id]`.
- JSON output always contains raw `id` fields.

## Output + Filtering
- Use `--output json` for machine-readable results.
- Use `--json` as shorthand for `--output json`.
- Use `--query` (or `--jq`) for JQ-style filtering on JSON output.
- Use `--fields id,order_number,...` as a shorthand for common projections (builds an internal jq query).
- On `orders`, `products`, `customers`, `--fields minimal|default|debug` expands to a preset field list.
- List commands return a pagination envelope (`{ items, pagination, ... }`). Use `--items-only` to emit just the `items` array.
- Data goes to stdout; errors go to stderr.
- Avoid `2>&1 | jq ...` unless you *want* stderr mixed into your JSON stream.

## Pagination + Sorting
- `--limit` sets page size for list commands (0 uses API defaults).
- `orders list`, `products list`, and `customers list` treat `--limit N` as “return up to N items” and auto-paginate across pages to reach N.
- `--page` and `--page-size` remain available on list commands.
- `--sort-by` and `--desc` work on list commands that support sorting.

## Date Filters
- `--from` / `--to` accept `YYYY-MM-DD` or RFC3339.
- Supported on: `orders list`, `draft-orders list`, `return-orders list`, `abandoned-checkouts list`, `order-attribution list`.

## Desire-Path Aliases
- Plural/singular: `orders` ↔ `order`, `products` ↔ `product`, etc.
- Verb aliases: `list` ↔ `ls`, `get` ↔ `show`, `create` ↔ `new`/`add`, `update` ↔ `edit`, `delete` ↔ `del`/`rm`, `cancel` ↔ `void`.
- Short aliases: `orders` → `ord`, `products` → `prod`, `customers` → `cust`, `inventory` → `inv`, `draft-orders` → `drafts`, `gift-cards` → `giftcard`/`gc`, `discount-codes` → `discounts`, `webhooks` → `hooks`, `shipping` → `ship`, `livestreams` → `live`/`livestream`/`streams`, `message-center` → `mc`/`messages`.

## Examples
```bash
# Discover resources
spl schema
spl help --json

# List orders (with structured IDs)
spl orders list --limit 5

# Copy/paste IDs directly
spl orders get [order:$ord_123]
spl refunds list --order-id [order:$ord_123]

# Sort products
spl products list --sort-by created_at --desc

# Date filters
spl orders list --from 2024-01-01 --to 2024-01-31

# JSON list convenience (items only)
spl orders list --limit 50 -o json --items-only

# Orders with line items (extra API calls; pulls order details for each list item)
spl orders list --limit 20 --desc -o json --items-only --expand details

# Field presets (agent-friendly JSON slices)
spl orders list --limit 20 --desc --json --items-only --fields minimal
spl products list --limit 20 --desc --json --items-only --fields default
spl customers list --limit 20 --desc --json --items-only --fields debug

# Order detail with expanded customer + product info on line items (extra API calls)
spl orders get [order:$ord_123] -o json --expand customer,products
```

## Name-to-ID Resolution (`--by`)

Instead of searching and extracting IDs manually, use `--by` on get commands to resolve a human-readable name/email/number to an ID in a single step:

```bash
# Old way (2 steps)
spl customers search --q "john@example.com" --output json | jq -r '.items[0].id'
spl customers get cust_abc123

# New way (1 step)
spl customers get --by john@example.com
spl orders get --by ORD-12345
spl products get --by "Widget Pro"
spl gifts get --by "Summer Gift"
spl promotions get --by "Flash Sale"
spl addon-products get --by "Bundle Deal"
spl customer-groups get --by "VIP Members"
```

Available on: `customers` (by email), `orders` (by number/query), `products` (by title), `gifts` (by title), `promotions` (by title), `addon-products` (by title), `customer-groups` (by name).

If the lookup matches multiple results, the first match is returned.

## Admin API Commands

Some commands use the Admin API proxy for undocumented Shopline admin endpoints. These require separate auth:

```bash
export SHOPLINE_ADMIN_BASE_URL=<base-url>
export SHOPLINE_ADMIN_TOKEN=<token>
export SHOPLINE_ADMIN_MERCHANT_ID=<merchant-id>
# Token and merchant ID can also be passed per-command: --admin-token / --admin-merchant-id
```

### Orders (Admin)
```bash
spl orders comment <order-id> --text "Note" --private
spl orders admin-refund <order-id> --performer-id usr_123 --amount 500 --remark "Damaged"
spl orders receipt-reissue <order-id>
```

### Products (Admin)
```bash
spl products hide <product-id>
spl products publish <product-id>
spl products unpublish <product-id>
```

### Shipping
```bash
spl shipping status <order-id>
spl shipping tracking <order-id>
spl shipping execute <order-id> --order-number ORD-123 --performer-id usr_123
spl shipping print-label <order-id> --upsert
```
Alias: `ship`

### Livestreams
```bash
spl livestreams list --type live --page 1
spl livestreams get <stream-id>
spl livestreams create --title "Sale" --platform facebook --owner "Host"
spl livestreams update <stream-id> --post-title "New Title"
spl livestreams delete <stream-id>
spl livestreams add-products <stream-id> --body '{"products":[...]}'
spl livestreams remove-products <stream-id> --product-ids id1,id2
spl livestreams start <stream-id> --platform facebook
spl livestreams end <stream-id>
spl livestreams comments <stream-id> --page 1
```
Aliases: `live`, `livestream`, `streams`

### Message Center
```bash
spl message-center list --platform line --state open
spl message-center send <conversation-id> --platform line --content "Hello"
```
Aliases: `mc`, `messages`

## Shorthand Flags (Orders & Promotions)

Orders and promotions accept individual property flags on create/update, so you don't need to build raw JSON payloads:

```bash
# Orders
spl orders create --email user@example.com --note "Test order" --tags "vip,rush"
spl orders update ord_123 --note "Updated note" --tags "priority"

# Promotions
spl promotions create --title "Sale" --discount-type percentage --discount-value 20 --starts-at 2026-03-01
spl promotions update promo_123 --title "New Title" --discount-value 30
```

These flags are merged into the request body alongside any `--data` JSON you provide.

## Dry-Run Support

All write commands (create, update, delete, cancel, activate, deactivate, etc.) support `--dry-run`. It prints the HTTP method, URL, and request body without sending the request:

```bash
spl orders create --dry-run --email test@example.com
spl products delete prod_123 --dry-run
spl promotions activate promo_123 --dry-run
spl customers update cust_456 --dry-run --data '{"first_name":"Jane"}'
```

Use `--dry-run` to verify what the CLI will send before making real changes.
