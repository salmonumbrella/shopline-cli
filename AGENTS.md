# Shopline CLI Agent Guide

## Quick Discovery
1. `shopline schema` or `shopline schema list`
2. `shopline schema get <resource>`
3. `shopline help --json`

## ID Conventions
- Text tables show IDs as `[resource:$id]` (copy/pasteable).
- Positional args and any `--*-id` flags accept `[resource:$id]`.
- JSON output always contains raw `id` fields.

## Output + Filtering
- Use `--output json` for machine-readable results.
- Use `--query` for JQ-style filtering on JSON output.
- Data goes to stdout; errors go to stderr.

## Pagination + Sorting
- `--limit` sets page size for list commands (0 uses API defaults).
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
```
