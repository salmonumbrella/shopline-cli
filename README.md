# 🛒 Shopline CLI — E-commerce in your terminal.

Shopline in your terminal. Manage orders, products, inventory, customers, fulfillments, shipping, webhooks, and more.

**Built for humans and AI agents alike.** Structured output, rich error messages, and self-documenting commands make this CLI ideal for automation and LLM-driven workflows.

## Features

- **Authentication** - OAuth flow with secure token storage and automatic refresh
- **Balances** - view store credits and account balances
- **Customers** - manage customers, addresses, groups, and membership
- **Fulfillments** - manage fulfillments, shipments, and return orders
- **Inventory** - manage inventory levels and warehouse locations
- **Multiple stores** - manage multiple Shopline store profiles
- **Orders** - view and manage orders, draft orders, and refunds
- **Products** - manage products, collections, categories, and media
- **Promotions** - manage coupons, discounts, gift cards, and sales
- **Shipping** - configure shipping zones, carriers, and delivery options
- **Webhooks** - configure and manage webhook endpoints

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/shopline-cli
```

### From Source

```bash
go install github.com/salmonumbrella/shopline-cli/cmd/shopline@latest
```

### From Releases

Download the latest binary from the [Releases](https://github.com/salmonumbrella/shopline-cli/releases) page.

## Quick Start

### 1. Authenticate

Choose one of two methods:

**Browser:**
```bash
shopline auth login
```

**Terminal:**
```bash
shopline auth add mystore
# You'll be prompted securely for OAuth credentials
```

### 2. Test Authentication

```bash
shopline auth status
```

## Configuration

### Store Selection

Specify the store using either a flag or environment variable:

```bash
# Via flag
shopline orders list --store mystore

# Via environment
export SHOPLINE_STORE=mystore
shopline orders list
```

### Environment Variables

- `SHOPLINE_STORE` - Default store profile name to use
- `SHOPLINE_OUTPUT` - Output format: `text` (default) or `json`
- `SHOPLINE_COLOR` - Color mode: `auto` (default), `always`, or `never`
- `SHOPLINE_DEBUG` - Enable HTTP request debug logging to stderr
- `SHOPLINE_RETRY_BASE` - Base backoff delay for network error retries (e.g. `200ms`)
- `SHOPLINE_RETRY_MAX` - Max backoff delay for network error retries (e.g. `2s`)
- `SHOPLINE_RETRY_BUDGET` - Total retry budget for network errors (e.g. `5s`, `0` disables)
- `SHOPLINE_RETRY_JITTER` - Jitter factor (0-1) applied to backoff (e.g. `0.2`)
- `NO_COLOR` - Set to any value to disable colors (standard convention)

## Security

### Credential Storage

Credentials are stored securely in your system's keychain:
- **macOS**: Keychain Access
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Windows**: Credential Manager

## Commands

### Authentication

```bash
shopline auth login                     # Authenticate via browser (recommended)
shopline auth add <name>                # Add credentials manually (prompts securely)
shopline auth list                      # List configured store profiles
shopline auth remove <name>             # Remove store profile
shopline auth status [--store <name>]   # Check authentication status
```

### Orders & Fulfillment

```bash
shopline orders list [--status <status>] [--from <date>] [--to <date>]
shopline orders get <orderId>
shopline orders cancel <orderId> [--reason <reason>]
shopline orders close <orderId>
shopline orders reopen <orderId>

shopline draft-orders list
shopline draft-orders get <draftOrderId>
shopline draft-orders create --data '{...}'
shopline draft-orders complete <draftOrderId>

shopline fulfillments list [--order-id <id>]
shopline fulfillments get <fulfillmentId>
shopline fulfillments create --order-id <id> --tracking-number <num> ...

shopline fulfillment-orders list [--order-id <id>]
shopline fulfillment-orders get <fulfillmentOrderId>

shopline refunds list [--order-id <id>]
shopline refunds get <refundId>
shopline refunds create --order-id <id> --amount <n> ...

shopline return-orders list
shopline return-orders get <returnOrderId>

shopline shipments list
shopline shipments get <shipmentId>
```

### Products & Catalog

```bash
shopline products list [--status <status>] [--vendor <vendor>]
shopline products get <productId>
shopline products create --title <title> --price <price> ...
shopline products update <productId> [--title <title>] [--price <price>]
shopline products delete <productId>

shopline collections list
shopline collections get <collectionId>
shopline collections create --title <title> ...
shopline collections update <collectionId> [--title <title>]
shopline collections delete <collectionId>

shopline categories list
shopline categories get <categoryId>

shopline product-listings list
shopline product-reviews list [--product-id <id>]
shopline product-subscriptions list

shopline medias list [--product-id <id>]
shopline medias upload <productId> --file <path>
```

### Inventory & Warehouses

```bash
shopline inventory list [--location-id <id>]
shopline inventory get <inventoryItemId>
shopline inventory adjust <inventoryItemId> --delta <n>
shopline inventory set <inventoryItemId> --quantity <n>

shopline inventory-levels list [--location-id <id>]
shopline inventory-levels set --inventory-item-id <id> --location-id <id> --quantity <n>

shopline warehouses list
shopline warehouses get <warehouseId>
shopline warehouses create --name <name> ...

shopline locations list
shopline locations get <locationId>

shopline purchase-orders list
shopline purchase-orders get <purchaseOrderId>
```

### Customers & Membership

```bash
shopline customers list [--email <email>]
shopline customers get <customerId>
shopline customers create --email <email> --first-name <name> ...
shopline customers update <customerId> [--email <email>]
shopline customers delete <customerId>

shopline customer-addresses list <customerId>
shopline customer-addresses get <customerId> <addressId>

shopline customer-groups list
shopline customer-groups get <groupId>

shopline member-points list [--customer-id <id>]
shopline member-points adjust <customerId> --points <n>

shopline membership list
shopline store-credits list [--customer-id <id>]
shopline wish-lists list [--customer-id <id>]
```

### Pricing & Promotions

```bash
shopline price-rules list
shopline price-rules get <priceRuleId>
shopline price-rules create --title <title> --value <n> ...

shopline coupons list [--price-rule-id <id>]
shopline coupons get <couponId>
shopline coupons create --code <code> --price-rule-id <id> ...

shopline discount-codes list
shopline discount-codes get <discountCodeId>

shopline promotions list
shopline promotions get <promotionId>

shopline sales list
shopline flash-price list
shopline gifts list

shopline gift-cards list
shopline gift-cards get <giftCardId>
shopline gift-cards create --initial-value <n> --currency <c> ...
```

### Shipping & Delivery

```bash
shopline shipping-zones list
shopline shipping-zones get <zoneId>
shopline shipping-zones create --name <name> --countries <codes> ...

shopline carrier-services list
shopline carrier-services get <carrierId>

shopline local-delivery list
shopline pickup list
```

### Payments & Finance

```bash
shopline payments list [--order-id <id>]
shopline payments get <paymentId>

shopline payouts list
shopline payouts get <payoutId>

shopline transactions list [--order-id <id>]
shopline transactions get <transactionId>

shopline disputes list
shopline disputes get <disputeId>

shopline balance get
shopline taxes list
shopline currencies list
```

### Channels & Markets

```bash
shopline channels list
shopline channels get <channelId>

shopline channel-products list [--channel-id <id>]
shopline channel-products sync <productId> --channel-id <id>

shopline markets list
shopline markets get <marketId>

shopline countries list
shopline domains list
```

### Content & Themes

```bash
shopline themes list
shopline themes get <themeId>
shopline themes publish <themeId>

shopline assets list [--theme-id <id>]
shopline assets get <assetKey> [--theme-id <id>]
shopline assets upload <assetKey> --file <path> [--theme-id <id>]

shopline pages list
shopline pages get <pageId>
shopline pages create --title <title> --body <html> ...

shopline blogs list
shopline articles list [--blog-id <id>]

shopline redirects list
shopline redirects create --path <from> --target <to>

shopline files list
shopline files upload --file <path>
```

### Storefront API

```bash
shopline storefront-products list
shopline storefront-promotions list
shopline storefront-carts list
shopline storefront-tokens list
shopline storefront-oauth list
```

### Store Settings

```bash
shopline shop get                               # Get shop info
shopline settings list                          # List all settings
shopline checkout-settings get                  # Get checkout settings

shopline metafields list [--owner-type <type>] [--owner-id <id>]
shopline metafields get <metafieldId>
shopline metafields create --namespace <ns> --key <key> --value <v> ...

shopline metafield-definitions list
shopline custom-fields list
shopline script-tags list
```

### Staff & Operations

```bash
shopline staffs list
shopline staffs get <staffId>

shopline merchants get

shopline operation-logs list [--from <date>] [--to <date>]
```

### Webhooks

```bash
shopline webhooks list
shopline webhooks get <webhookId>
shopline webhooks create --topic orders/create --address https://example.com/hook
shopline webhooks update <webhookId> [--address <url>]
shopline webhooks delete <webhookId>
```

### Bulk Operations

```bash
shopline bulk-operations list
shopline bulk-operations get <operationId>
shopline bulk-operations create --query <graphql>
```

## Agent-Friendly Design

This CLI is designed for AI agents and automation tools with features that make programmatic interaction reliable and predictable.

### Schema Discovery

Agents can discover available API resources and their operations without guessing:

```bash
# List all available resources
shopline schema list

# Get details about a specific resource
shopline schema get orders
shopline schema get products
shopline schema get customers
```

### Rich Errors with Suggestions

Errors include actionable suggestions for recovery:

```bash
$ shopline orders get ord_invalid123
Error: Order not found (404)

Suggestions:
  • Verify the order ID is correct
  • Run 'shopline orders list' to see available orders
  • Check that you're using the correct store profile
```

### Structured Output

JSON output follows consistent conventions that agents can parse reliably:

- Data goes to stdout, errors and progress to stderr
- Pagination info included in response metadata
- Consistent field naming across all resources

```bash
# Pipe-friendly: only data goes to stdout
shopline orders list --output json | jq '.items[0]'

# Parse errors programmatically
shopline orders get invalid 2>&1 | jq -r '.error.suggestions[]'
```

### Automation Flags

Flags designed for non-interactive use:

- `--yes` — Skip confirmation prompts
- `--dry-run` — Preview changes without executing
- `--output json` — Machine-readable output
- `--query` — Built-in JQ filtering
- `--limit 0` — Fetch all results (no pagination limit)

## Output Formats

### Text

Human-readable tables with colors and formatting:

```bash
$ shopline orders list
ORDER_ID                   STATUS      TOTAL       CUSTOMER          CREATED
order_abc123...            PAID        $125.00     john@example.com  2024-01-15
order_def456...            PENDING     $89.50      jane@example.com  2024-01-14

$ shopline products list
PRODUCT_ID                 TITLE              PRICE     INVENTORY   STATUS
prod_xyz789...             Blue T-Shirt       $29.99    150         ACTIVE
prod_uvw012...             Running Shoes      $89.00    42          ACTIVE
```

### JSON

Machine-readable output:

```bash
$ shopline orders list --output json
{
  "orders": [
    {"id": "order_abc123", "status": "PAID", "total": 125.00},
    {"id": "order_def456", "status": "PENDING", "total": 89.50}
  ]
}
```

Data goes to stdout, errors and progress to stderr for clean piping.

## Examples

### Create a product with inventory

```bash
# Create the product
shopline products create \
  --title "Blue T-Shirt" \
  --price 29.99 \
  --vendor "ACME Apparel" \
  --status active

# Set inventory at a location
shopline inventory-levels set \
  --inventory-item-id <inventoryItemId> \
  --location-id <locationId> \
  --quantity 100
```

### Fulfill an order with tracking

```bash
# List orders to find ID
shopline orders list --status unfulfilled

# Create fulfillment with tracking
shopline fulfillments create \
  --order-id <orderId> \
  --tracking-number "1Z999AA10123456784" \
  --tracking-company "UPS" \
  --notify-customer
```

### View customer order history

```bash
shopline orders list \
  --customer-id <customerId> \
  --from 2024-01-01 \
  --to 2024-12-31 \
  --output json | jq '.orders[] | select(.total > 100)'
```

### Switch between stores

```bash
# Check production store
shopline orders list --store prod

# Check staging store
shopline orders list --store staging

# Or set default
export SHOPLINE_STORE=prod
shopline orders list
```

### Automation

Use `--yes` to skip confirmations, `--limit` to control result size, and `--sort-by` for ordering:

```bash
# Delete a product without confirmation prompt
shopline products delete prod_xxx --yes

# Get the 5 most recent orders
shopline orders list --limit 5 --sort-by created_at --desc --output json

# Fetch all products (no pagination limit)
shopline products list --limit 0 --output json

# Pipeline: cancel all draft orders older than 30 days
shopline draft-orders list --output json \
  | jq -r '.items[] | select(.created_at < "2024-01-01") | .id' \
  | xargs -I{} shopline draft-orders delete {} --yes
```

### Dry-Run Mode

Preview mutations before executing:

```bash
shopline orders cancel order_xxx --dry-run --reason "Customer request"

# Output:
# [DRY-RUN] Would cancel order
# ─────────────────────────────────────
# Order: order_xxx
# Customer: john@example.com
# Total: $125.00
# Reason: Customer request
# ─────────────────────────────────────
# No changes made (dry-run mode)
```

### JQ Filtering

Filter JSON output with JQ expressions:

```bash
# Get only pending orders
shopline orders list --output json --query '.orders[] | select(.status=="PENDING")'

# Extract product IDs
shopline products list --output json --query '[.products[].id]'

# Filter customers by email domain
shopline customers list --output json --query '.customers[] | select(.email | endswith("@company.com"))'
```

## Global Flags

All commands support these flags:

- `--store <name>`, `-s` - Store profile to use (overrides SHOPLINE_STORE)
- `--output <format>`, `-o` - Output format: `text` or `json` (default: text)
- `--color <mode>` - Color mode: `auto`, `always`, or `never` (default: auto)
- `--query <expr>` - JQ filter expression for JSON output
- `--yes`, `-y` - Skip confirmation prompts (useful for scripts and automation)
- `--limit <n>` - Limit number of results returned (0 = no limit, fetches all)
- `--sort-by <field>` - Sort results by field name (e.g., `created_at`, `total`)
- `--desc` - Sort descending (requires `--sort-by`)
- `--dry-run` - Preview changes without executing them
- `--help`, `-h` - Show help for any command
- `--version`, `-v` - Show version information

## Shell Completions

Generate shell completions for your preferred shell:

### Bash

```bash
# macOS (Homebrew):
shopline completion bash > $(brew --prefix)/etc/bash_completion.d/shopline

# Linux:
shopline completion bash > /etc/bash_completion.d/shopline

# Or source directly in current session:
source <(shopline completion bash)
```

### Zsh

```zsh
# Save to fpath:
shopline completion zsh > "${fpath[1]}/_shopline"

# Or add to .zshrc for auto-loading:
echo 'autoload -U compinit; compinit' >> ~/.zshrc
echo 'source <(shopline completion zsh)' >> ~/.zshrc
```

### Fish

```fish
shopline completion fish > ~/.config/fish/completions/shopline.fish
```

### PowerShell

```powershell
# Load for current session:
shopline completion powershell | Out-String | Invoke-Expression

# Or add to profile for persistence:
shopline completion powershell >> $PROFILE
```

## Development

After cloning, install git hooks:

```bash
make setup
```

This installs pre-commit and pre-push hooks for linting and testing.

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run linter
make fmt      # Format code
make ci       # Run all checks
```

## License

MIT

## Links

- [Shopline API Documentation](https://open.shopline.com/documents)
