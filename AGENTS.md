# Shopline CLI Agent Guide

## Quick Discovery
1. `shopline schema` or `shopline schema list` (add `--output json` for structured output)
2. `shopline schema get <resource>` (add `--output json` for structured output)
3. `shopline help --json`
   - Use `shopline help --json --deep` for the full command tree in one call

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
- `orders list` special-case: `--limit N` is treated as “return up to N orders” and will auto-paginate across pages to reach N (useful when the API caps/ignores `page_size`).
- `--page` and `--page-size` remain available on list commands.
- `--sort-by` and `--desc` work on list commands that support sorting.

## Date Filters
- `--from` / `--to` accept `YYYY-MM-DD` or RFC3339.
- Supported on: `orders list`, `draft-orders list`, `return-orders list`, `abandoned-checkouts list`, `order-attribution list`.

## Desire-Path Aliases
- Plural/singular: `orders` ↔ `order`, `products` ↔ `product`, etc.
- Verb aliases: `list` ↔ `ls`, `get` ↔ `show`, `create` ↔ `new`/`add`, `update` ↔ `edit`, `delete` ↔ `del`/`rm`, `cancel` ↔ `void`.
- Short aliases: `orders` → `ord`, `products` → `prod`, `customers` → `cust`, `inventory` → `inv`, `draft-orders` → `drafts`, `gift-cards` → `giftcard`/`gc`, `discount-codes` → `discounts`, `webhooks` → `hooks`.

## Examples
```bash
# Discover resources
shopline schema
shopline help --json

# List orders (with structured IDs)
shopline orders list --limit 5

# Copy/paste IDs directly
shopline orders get [order:$ord_123]
shopline refunds list --order-id [order:$ord_123]

# Sort products
shopline products list --sort-by created_at --desc

# Date filters
shopline orders list --from 2024-01-01 --to 2024-01-31

# JSON list convenience (items only)
shopline orders list --limit 50 -o json --items-only

# Orders with line items (extra API calls; pulls order details for each list item)
shopline orders list --limit 20 --desc -o json --items-only --expand details

# Field presets (agent-friendly JSON slices)
shopline orders list --limit 20 --desc --json --items-only --fields minimal
shopline products list --limit 20 --desc --json --items-only --fields default
shopline customers list --limit 20 --desc --json --items-only --fields debug

# Order detail with expanded customer + product info on line items (extra API calls)
shopline orders get [order:$ord_123] -o json --expand customer,products
```

## Name-to-ID Resolution (`--by`)

Instead of searching and extracting IDs manually, use `--by` on get commands to resolve a human-readable name/email/number to an ID in a single step:

```bash
# Old way (2 steps)
shopline customers search --q "john@example.com" --output json | jq -r '.items[0].id'
shopline customers get cust_abc123

# New way (1 step)
shopline customers get --by john@example.com
shopline orders get --by ORD-12345
shopline products get --by "Widget Pro"
shopline gifts get --by "Summer Gift"
shopline promotions get --by "Flash Sale"
shopline addon-products get --by "Bundle Deal"
shopline customer-groups get --by "VIP Members"
```

Available on: `customers` (by email), `orders` (by number/query), `products` (by title), `gifts` (by title), `promotions` (by title), `addon-products` (by title), `customer-groups` (by name).

If the lookup matches multiple results, the first match is returned.

## Shorthand Flags (Orders & Promotions)

Orders and promotions accept individual property flags on create/update, so you don't need to build raw JSON payloads:

```bash
# Orders
shopline orders create --email user@example.com --note "Test order" --tags "vip,rush"
shopline orders update ord_123 --note "Updated note" --tags "priority"

# Promotions
shopline promotions create --title "Sale" --discount-type percentage --discount-value 20 --starts-at 2026-03-01
shopline promotions update promo_123 --title "New Title" --discount-value 30
```

These flags are merged into the request body alongside any `--data` JSON you provide.

## Dry-Run Support

All write commands (create, update, delete, cancel, activate, deactivate, etc.) support `--dry-run`. It prints the HTTP method, URL, and request body without sending the request:

```bash
shopline orders create --dry-run --email test@example.com
shopline products delete prod_123 --dry-run
shopline promotions activate promo_123 --dry-run
shopline customers update cust_456 --dry-run --data '{"first_name":"Jane"}'
```

Use `--dry-run` to verify what the CLI will send before making real changes.
