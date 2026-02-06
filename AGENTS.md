# Shopline CLI Agent Guide

## Quick Discovery
1. `shopline schema` or `shopline schema list` (add `--output json` for structured output)
2. `shopline schema get <resource>` (add `--output json` for structured output)
3. `shopline help --json`

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
