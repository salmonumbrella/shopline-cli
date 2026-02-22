# Shopline CLI

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
go install github.com/salmonumbrella/shopline-cli/cmd/spl@latest
```

### From Releases

Download the latest binary from the [Releases](https://github.com/salmonumbrella/shopline-cli/releases) page.

## Quick Start

### 1. Authenticate

**Browser:**
```bash
spl auth login                            # Opens browser for interactive login
```

**Terminal:**
```bash
spl auth add mystore                      # Add credentials (prompts securely)
```

### 2. Verify Setup

```bash
spl auth status                           # Show current config
```

### 3. List Orders

```bash
spl o ls -S paid                          # List paid orders
spl o ls -S paid -o json                  # JSON output
```

## Configuration

### Store Selection

```bash
spl o ls -s mystore                       # Via flag
export SHOPLINE_STORE=mystore                  # Via environment
spl o ls                                  # Uses default store
```

### Environment Variables

```bash
export SHOPLINE_STORE=mystore                  # Default store profile
export SHOPLINE_OUTPUT=json                    # Output format: text (default) or json
export SHOPLINE_COLOR=auto                     # Color mode: auto, always, or never
export SHOPLINE_CREDENTIALS_DIR=~/.openclaw/credentials/shopline-cli  # Tool-specific credential dir
export CW_CREDENTIALS_DIR=~/.openclaw/credentials                     # OpenClaw shared credential root
export SHOPLINE_DEBUG=1                        # Enable HTTP debug logging to stderr
export SHOPLINE_RETRY_BASE=200ms               # Base backoff delay for retries
export SHOPLINE_RETRY_MAX=2s                   # Max backoff delay for retries
export SHOPLINE_RETRY_BUDGET=5s                # Total retry budget (0 disables)
export SHOPLINE_RETRY_JITTER=0.2               # Jitter factor (0-1)
export NO_COLOR=1                              # Disable colors (standard convention)
```

When present, `~/.openclaw/.env` is auto-loaded at startup. Existing environment variables are not overridden.

## Security

### Credential Storage

Credentials are stored securely in your system's keychain:
- **macOS**: Keychain Access
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Windows**: Credential Manager

OpenClaw integration:
- `~/.openclaw/.env` is loaded automatically on startup (if it exists).
- `SHOPLINE_CREDENTIALS_DIR` sets an explicit credentials directory for this CLI.
- `CW_CREDENTIALS_DIR` sets a shared OpenClaw credentials root; this CLI stores credentials under `CW_CREDENTIALS_DIR/shopline-cli/`.

## Commands

### Authentication

```bash
spl auth login                            # Authenticate via browser (recommended)
spl auth add <name>                       # Add credentials manually
spl auth ls                               # List configured store profiles
spl auth rm <name>                        # Remove store profile
spl auth status -s <name>                 # Check authentication status
```

### Orders & Fulfillment

```bash
spl o ls                                  # List orders
spl o ls -S paid -f 2024-01-01 -t 2024-12-31  # Filter by status and date
spl o g <id>                              # Get order details
spl o cancel <id> --rsn "Customer request"  # Cancel order
spl o close <id>                          # Close order
spl o reopen <id>                         # Reopen order

spl oa g <id>                             # Get order attribution
spl ork ls --oid <id>                     # List order risks
spl ork mk --oid <id> --data '{...}'      # Create order risk

spl do ls                                 # List draft orders
spl do g <id>                             # Get draft order
spl do mk --data '{...}'                  # Create draft order
spl do complete <id>                      # Complete draft order

spl ff ls --oid <id>                      # List fulfillments for order
spl ff g <id>                             # Get fulfillment
spl ff mk --oid <id> --tn <num> ...       # Create fulfillment with tracking

spl fo ls --oid <id>                      # List fulfillment orders
spl fo g <id>                             # Get fulfillment order

spl fs ls                                 # List fulfillment services
spl fs mk -n <name> ...                   # Create fulfillment service

spl rf ls --oid <id>                      # List refunds for order
spl rf g <id>                             # Get refund
spl rf mk --oid <id> --amt <n> ...        # Create refund

spl ro ls                                 # List return orders
spl ro g <id>                             # Get return order

spl shp ls                                # List shipments
spl shp g <id>                            # Get shipment

spl ac ls                                 # List abandoned checkouts
spl ac g <id>                             # Get abandoned checkout
spl ac send-recovery <id>                 # Send recovery email

spl ct prepare                            # Prepare cart
spl ct exchange                           # Exchange cart
```

### Products & Catalog

```bash
spl p ls                                  # List products
spl p ls -S active --vn "ACME"            # Filter by status and vendor
spl p g <id>                              # Get product
spl p mk --ti "Blue T-Shirt" --pr 29.99  # Create product
spl p up <id> --ti "Red T-Shirt"          # Update product
spl p rm <id>                             # Delete product

spl col ls                                # List collections
spl col g <id>                            # Get collection
spl col mk --ti "Summer" ...              # Create collection
spl col up <id> --ti "Winter"             # Update collection
spl col rm <id>                           # Delete collection

spl smc ls                                # List smart collections
spl smc mk --ti "New Arrivals" --rules '{...}'  # Create smart collection

spl cat ls                                # List categories
spl cat g <id>                            # Get category

spl txn ls                                # List taxonomies
spl txn mk -n <name> ...                  # Create taxonomy

spl pl ls                                 # List product listings
spl prv ls --pid <id>                     # List product reviews
spl prc mk --review-id <id> --data '{...}'  # Create review comment
spl psu ls                                # List product subscriptions

spl ap ls                                 # List addon products
spl ap mk --data '{...}'                  # Create addon product

spl tg ls                                 # List tags
spl tg mk -n <name>                       # Create tag

spl lb ls                                 # List labels
spl lb mk -n <name> ...                   # Create label

spl sc ls                                 # List size charts
spl sc mk --data '{...}'                  # Create size chart

spl mds ls --pid <id>                     # List media for product
spl mds upload <id> --file <path>         # Upload media
spl md mk --data '{...}'                  # Create image
```

### Inventory & Warehouses

```bash
spl inv ls --lid <id>                     # List inventory at location
spl inv g <id>                            # Get inventory item
spl inv adjust <id> --delta <n>           # Adjust inventory
spl inv set <id> --qty <n>                # Set inventory quantity

spl il ls --lid <id>                      # List inventory levels
spl il set --inventory-item-id <id> --lid <id> --qty <n>  # Set level

spl wh ls                                 # List warehouses
spl wh g <id>                             # Get warehouse
spl wh mk -n <name> ...                   # Create warehouse

spl loc ls                                # List locations
spl loc g <id>                            # Get location

spl pur ls                                # List purchase orders
spl pur g <id>                            # Get purchase order
```

### Customers & Membership

```bash
spl cu ls                                 # List customers
spl cu ls -e "john@example.com"           # Filter by email
spl cu g <id>                             # Get customer
spl cu find "john"                        # Search customers
spl cu mk -e "john@example.com" --fn "John" --ln "Doe"  # Create customer
spl cu up <id> -e "new@example.com"       # Update customer
spl cu rm <id>                            # Delete customer

spl ca ls <customerId>                    # List customer addresses
spl ca g <customerId> <addressId>         # Get address

spl cg ls                                 # List customer groups
spl cg g <id>                             # Get customer group

spl cbl ls                                # List customer blacklist
spl cbl mk --data '{...}'                 # Add to blacklist

spl css ls                                # List saved searches
spl css mk -n <name> -q <query>           # Create saved search

spl mp ls --cid <id>                      # List member points
spl mp adjust <id> --points <n>           # Adjust points

spl mem ls                                # List membership tiers
spl scr ls --cid <id>                     # List store credits

spl uc ls                                 # List user coupons
spl uc assign --coupon-id <id> --uid <id>  # Assign coupon
spl uc redeem --cd <code>                 # Redeem coupon

spl ucr ls                                # List user credits
spl ucr bulk-update --data '{...}'        # Bulk update credits

spl wl ls --cid <id>                      # List wish lists
spl wli ls --wish-list-id <id>            # List wish list items
spl wli mk --wish-list-id <id> --pid <id>  # Add to wish list

spl sub ls                                # List subscriptions
spl sub g <id>                            # Get subscription
spl sub mk --data '{...}'                 # Create subscription

spl conv ls                               # List conversations
spl conv g <id>                           # Get conversation
spl conv send <id> -m <text>              # Send message
```

### Pricing & Promotions

```bash
spl pr ls                                 # List price rules
spl pr g <id>                             # Get price rule
spl pr mk --ti "10% Off" --value <n> ...  # Create price rule

spl cpn ls --price-rule-id <id>           # List coupons
spl cpn g <id>                            # Get coupon
spl cpn mk --cd <code> --price-rule-id <id> ...  # Create coupon

spl dc ls                                 # List discount codes
spl dc g <id>                             # Get discount code

spl pm ls                                 # List promotions
spl pm g <id>                             # Get promotion

spl sl ls                                 # List sales
spl flp ls                                # List flash prices
spl fpc ls                                # List flash price campaigns
spl fpc mk --data '{...}'                 # Create flash price campaign
spl gi ls                                 # List gifts

spl gc ls                                 # List gift cards
spl gc g <id>                             # Get gift card
spl gc mk --initial-value <n> --cur <c>   # Create gift card

spl sp ls                                 # List selling plans
spl sp mk --data '{...}'                  # Create selling plan
```

### Shipping & Delivery

```bash
spl sz ls                                 # List shipping zones
spl sz g <id>                             # Get shipping zone
spl sz mk -n <name> --countries <codes>   # Create shipping zone

spl cs ls                                 # List carrier services
spl cs g <id>                             # Get carrier service

spl dop ls                                # List delivery options
spl dop g <id>                            # Get delivery option
spl dop time-slots <id>                   # Get delivery time slots

spl ld ls                                 # List local delivery options
spl pu ls                                 # List pickup locations
```

### Payments & Finance

```bash
spl pay ls --oid <id>                     # List payments for order
spl pay g <id>                            # Get payment

spl po ls                                 # List payouts
spl po g <id>                             # Get payout

spl tx ls --oid <id>                      # List transactions for order
spl tx g <id>                             # Get transaction

spl dis ls                                # List disputes
spl dis g <id>                            # Get dispute

spl bal g                                 # Get balance
spl tax ls                                # List taxes
spl ts ls                                 # List tax services
spl cur ls                                # List currencies
```

### Channels & Markets

```bash
spl ch ls                                 # List channels
spl ch g <id>                             # Get channel

spl chp ls --channel-id <id>              # List channel products
spl chp sync <id> --channel-id <id>       # Sync product to channel

spl mk ls                                 # List markets
spl mk g <id>                             # Get market

spl cnt ls                                # List countries
spl dm ls                                 # List domains
```

### Content & Themes

```bash
spl th ls                                 # List themes
spl th g <id>                             # Get theme
spl th publish <id>                       # Publish theme

spl as ls --tid <id>                      # List assets for theme
spl as g <key> --tid <id>                 # Get asset
spl as upload <key> --file <path> --tid <id>  # Upload asset

spl pg ls                                 # List pages
spl pg g <id>                             # Get page
spl pg mk --ti "About Us" -b <html>       # Create page

spl bl ls                                 # List blogs
spl art ls --bid <id>                     # List articles

spl rd ls                                 # List redirects
spl rd mk --path <from> --target <to>     # Create redirect

spl fi ls                                 # List files
spl fi upload --file <path>               # Upload file
```

### Storefront API

```bash
spl sfpr ls                               # List storefront products
spl sfpm ls                               # List storefront promotions
spl sfc ls                                # List storefront carts
spl sft ls                                # List storefront tokens
spl sfo ls                                # List storefront OAuth
spl sfoa ls                               # List storefront OAuth applications
spl sfoa mk --data '{...}'                # Create OAuth application
```

### Store Settings

```bash
spl shop info                             # Get shop info
spl shop settings                         # Get shop settings
spl set ls                                # List all settings
spl cos g                                 # Get checkout settings

spl mf ls --ot <type> --owid <id>         # List metafields
spl mf g <id>                             # Get metafield
spl mf mk --ns <ns> --key <key> --value <v> ...  # Create metafield

spl mfd ls                                # List metafield definitions
spl cf ls                                 # List custom fields
spl st ls                                 # List script tags

spl token info                            # Get current access token info
spl tok ls                                # List API tokens
spl tok mk --data '{...}'                 # Create API token
```

### Staff & Operations

```bash
spl sf ls                                 # List staff
spl sf g <id>                             # Get staff member

spl mr g                                  # Get merchant info

spl mup status                            # Multipass status
spl mup enable                            # Enable multipass
spl mup disable                           # Disable multipass
spl mup token -e <email>                  # Generate multipass token
spl mup rotate                            # Rotate multipass secret

spl ol ls -f <date> -t <date>             # List operation logs
```

### Webhooks

```bash
spl hooks ls                              # List webhooks
spl hooks g <id>                          # Get webhook
spl hooks mk --topic orders/create --addr https://example.com/hook  # Create webhook
spl hooks up <id> --addr <url>            # Update webhook
spl hooks rm <id>                         # Delete webhook
```

### Marketing & Campaigns

```bash
spl afc ls                                # List affiliate campaigns
spl afc g <id>                            # Get campaign
spl afc mk --data '{...}'                 # Create campaign

spl me ls                                 # List marketing events
spl me mk --data '{...}'                  # Create marketing event
```

### B2B / Company

```bash
spl cc ls                                 # List company catalogs
spl cc mk --data '{...}'                  # Create catalog

spl ccr ls                                # List company credits
spl ccr adjust --compid <id> --amt <n>    # Adjust company credits

spl cp ls                                 # List catalog pricing
spl cp mk --data '{...}'                  # Create catalog pricing
```

### Customer Data Platform (CDP)

```bash
spl cdp profiles ls                       # List CDP profiles
spl cdp profiles g <id>                   # Get CDP profile
spl cdp events ls                         # List CDP events
spl cdp events g <id>                     # Get CDP event
spl cdp segments ls                       # List CDP segments
spl cdp segments g <id>                   # Get CDP segment
```

### Point of Sale

```bash
spl ppo ls                                # List POS purchase orders
spl ppo g <id>                            # Get POS purchase order
```

### Bulk Operations

```bash
spl bo ls                                 # List bulk operations
spl bo g <id>                             # Get bulk operation
spl bo mk --gql <graphql>                 # Create bulk operation
```

### Schema Discovery

```bash
spl schema                                # List all available resources
spl schema ls                             # List all resources (alias)
spl schema g orders                       # Get details about a resource
spl schema -o json                        # JSON introspection for agents
spl help --json                           # Full help as JSON
```

## Agent-Friendly Design

This CLI is designed for AI agents and automation tools with features that make programmatic interaction reliable and predictable.

For a concise agent workflow guide, see `AGENTS.md`.

### Copy/Pasteable IDs

List tables include IDs in a structured format that can be pasted directly into other commands:

```
[order:$ord_123]
[product:$prod_456]
```

The CLI accepts these in positional args and any `--*-id` flags.

### Rich Errors with Suggestions

```bash
$ spl o g ord_invalid123
Error: Order not found (404)

Suggestions:
  - Verify the order ID is correct
  - Run 'spl orders list' to see available orders
  - Check that you're using the correct store profile
```

### Structured Output

- Data goes to stdout, errors and progress to stderr
- Pagination info included in response metadata
- Consistent field naming across all resources

```bash
spl o ls -o json | jq '.items[0]'         # Pipe-friendly: only data goes to stdout
spl o g invalid 2>&1 | jq -r '.error.suggestions[]'  # Parse errors programmatically
```

## Output Formats

### Text

Human-readable tables:

```bash
$ spl o ls
ORDER                      STATUS      TOTAL       CUSTOMER          CREATED
[order:$ord_abc123]         PAID        $125.00     john@example.com  2024-01-15
[order:$ord_def456]         PENDING     $89.50      jane@example.com  2024-01-14
```

### JSON

Machine-readable output:

```bash
$ spl o ls -o json
{
  "items": [
    {"id": "ord_abc123", "status": "PAID", "total_price": "125.00", "currency": "USD"},
    {"id": "ord_def456", "status": "PENDING", "total_price": "89.50", "currency": "USD"}
  ],
  "pagination": {
    "current_page": 1,
    "per_page": 20,
    "total_count": 2,
    "total_pages": 1
  }
}
```

Data goes to stdout, errors and progress to stderr for clean piping.

## Examples

### Create a product with inventory

```bash
spl p mk --ti "Blue T-Shirt" --pr 29.99 --vn "ACME Apparel" -S active
spl il set --inventory-item-id <id> --lid <id> --qty 100
```

### Fulfill an order with tracking

```bash
spl o ls -S unfulfilled                   # Find unfulfilled orders
spl ff mk --oid <id> --tn "1Z999AA10123456784" --tc "UPS" --notify-customer
```

### View customer order history

```bash
spl o ls --cid <id> -f 2024-01-01 -t 2024-12-31 -o json | jq '.items[] | select(.total > 100)'
```

### Switch between stores

```bash
spl o ls -s prod                          # Check production store
spl o ls -s staging                       # Check staging store
export SHOPLINE_STORE=prod                     # Set default
spl o ls                                  # Uses default
```

### Automation

```bash
spl p rm prod_xxx -y                      # Delete without confirmation
spl o ls -l 5 --sb created_at -D -o json  # 5 most recent orders
spl p ls -l 0 -o json                     # All products (no limit)
spl do ls -o json | jq -r '.items[] | select(.created_at < "2024-01-01") | .id' | xargs -I{} spl do rm {} -y  # Batch delete old drafts
```

### Dry-Run Mode

Preview mutations before executing:

```bash
spl o cancel order_xxx --dr --rsn "Customer request"
# [DRY-RUN] Would cancel order
# Order: order_xxx
# Reason: Customer request
# No changes made (dry-run mode)
```

### JQ Filtering

```bash
spl o ls -o json -q '.items[] | select(.status=="PENDING")'  # Pending orders
spl p ls -o json -q '[.items[].id]'       # Extract product IDs
spl cu ls -o json -q '.items[] | select(.email | endswith("@company.com"))'  # Filter by domain
```

## Global Flags

All commands support these flags:

- `-s <name>` / `--store <name>` - Store profile (overrides `SHOPLINE_STORE`)
- `-o <format>` / `--output <format>` - Output format: `text` or `json` (default: text)
- `-j` / `--json` - Shorthand for `-o json`
- `-q <expr>` / `--query <expr>` - JQ filter expression for JSON output
- `-F <fields>` / `--fields <fields>` - Select fields in JSON output
- `-y` / `--yes` - Skip confirmation prompts
- `-l <n>` / `--limit <n>` - Limit results (`0` uses API defaults); `orders list`, `products list`, and `customers list` auto-fetch up to `N`
- `--sort-by <field>` / `--sb <field>` - Sort results by field
- `-D` / `--desc` - Sort descending (requires `--sort-by`)
- `--dry-run` / `--dr` - Preview changes without executing
- `--color <mode>` - Color mode: `auto`, `always`, or `never`
- `--items-only` / `--io` - Output only items array (for JSON)
- `--help` / `-h` - Show help for any command
- `--version` / `-v` - Show version information

## Command Aliases

Every resource command has short aliases for fast typing:

| Command | Aliases |
|---------|---------|
| `orders` | `ord`, `o` |
| `products` | `prod`, `p` |
| `customers` | `cust`, `cu` |
| `refunds` | `ref`, `rf` |
| `collections` | `col` |
| `draft-orders` | `drafts`, `do` |
| `fulfillments` | `ful`, `ff` |
| `payments` | `pay` |
| `transactions` | `tx` |
| `shipments` | `shp` |
| `inventory` | `inv` |
| `inventory-levels` | `il` |
| `warehouses` | `wh` |
| `locations` | `loc` |
| `categories` | `cat` |
| `channels` | `ch` |
| `coupons` | `cpn` |
| `discount-codes` | `discounts`, `dc` |
| `promotions` | `promo`, `pm` |
| `gift-cards` | `giftcard`, `gc` |
| `webhooks` | `hooks` |
| `shipping-zones` | `sz` |
| `themes` | `th` |
| `pages` | `pg` |
| `tags` | `tg` |
| `labels` | `lb` |
| `subscriptions` | `sub` |

Singular forms also work: `spl order ls` = `spl orders ls`.

### Verb Aliases

Subcommands have shorter forms:

| Verb | Aliases |
|------|---------|
| `list` | `ls`, `l` |
| `get` | `show`, `g` |
| `create` | `new`, `add`, `mk` |
| `update` | `edit`, `up` |
| `delete` | `del`, `rm` |
| `cancel` | `void` |
| `search` | `find`, `q` |
| `count` | `cnt` |

### Flag Aliases

Commonly used flags have short aliases to reduce typing:

#### ID Flags

| Flag | Alias |
|------|-------|
| `--order-id` | `--oid` |
| `--product-id` | `--pid` |
| `--customer-id` | `--cid` |
| `--variant-id` | `--vid` |
| `--location-id` | `--lid` |
| `--theme-id` | `--tid` |
| `--user-id` | `--uid` |
| `--blog-id` | `--bid` |
| `--company-id` | `--compid` |

#### Content & Filtering

| Flag | Alias |
|------|-------|
| `--status` | `-S` |
| `--name` | `-n` |
| `--email` | `-e` |
| `--title` | `--ti` |
| `--body` | `-b` |
| `--message` | `-m` |
| `--reason` | `--rsn` |
| `--vendor` | `--vn` |
| `--code` | `--cd` |

#### Amounts & Quantities

| Flag | Alias |
|------|-------|
| `--amount` | `--amt` |
| `--price` | `--pr` |
| `--quantity` | `--qty` |
| `--currency` | `--cur` |

#### Time

| Flag | Alias |
|------|-------|
| `--from` | `-f` |
| `--to` | `-t` |

#### Tracking

| Flag | Alias |
|------|-------|
| `--tracking-number` | `--tn` |
| `--tracking-company` | `--tc` |
| `--tracking-url` | `--tu` |

#### Pagination

| Flag | Alias |
|------|-------|
| `--page` | `--pg` |
| `--page-size` | `--ps` |

### Examples

```bash
# These are equivalent:
spl orders list --status paid
spl o ls -S paid

spl products get abc123
spl p g abc123

spl customers create --email "john@example.com" --first-name "John"
spl cu mk -e "john@example.com" --fn "John"

spl refunds create --order-id abc --amount 50
spl rf mk --oid abc --amt 50

spl webhooks delete abc123
spl hooks rm abc123
```

## Shell Completions

### Bash

```bash
spl completion bash > $(brew --prefix)/etc/bash_completion.d/spl  # macOS
spl completion bash > /etc/bash_completion.d/spl                  # Linux
source <(spl completion bash)                                          # Current session
```

### Zsh

```zsh
spl completion zsh > "${fpath[1]}/_spl"
echo 'source <(spl completion zsh)' >> ~/.zshrc
```

### Fish

```fish
spl completion fish > ~/.config/fish/completions/spl.fish
```

### PowerShell

```powershell
spl completion powershell | Out-String | Invoke-Expression
spl completion powershell >> $PROFILE                                  # Persist
```

## Development

### Prerequisites

Install [lefthook](https://github.com/evilmartians/lefthook) for git hooks:

```bash
brew install lefthook
```

### Setup

```bash
make setup       # Install dev tools (golangci-lint, gofumpt, goimports)
lefthook install # Install git hooks
```

Lefthook runs automatically on:
- **pre-commit**: `golangci-lint` (lint) + `gofumpt` (format check) in parallel
- **pre-push**: `go test -race ./...` (full test suite with race detector)

### Build & Test

```bash
make build    # Build binary
make test     # Run tests with race detector
make lint     # Run linter
make fmt      # Format code
make ci       # Run all checks (fmt-check + lint + test)
```

### Smoke Test

Run a non-destructive Shopline smoke suite (read checks + `--dry-run` write checks):

```bash
scripts/smoke_shopline.sh --store <store-profile>
```

Optional live mutation probe (disabled by default):

```bash
scripts/smoke_shopline.sh --store <store-profile> --allow-mutations
```

## License

MIT

## Links

- [Shopline API Documentation](https://open.shopline.com/documents)
