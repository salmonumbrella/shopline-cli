# Shopline CLI

A command-line interface for the Shopline e-commerce platform API.

## Installation

### From Source

```bash
go install github.com/salmonumbrella/shopline-cli/cmd/shopline@latest
```

### From Releases

Download the latest binary from the [Releases](https://github.com/salmonumbrella/shopline-cli/releases) page.

## Quick Start

### Authentication

Add a store profile to authenticate with the Shopline API:

```bash
shopline auth add mystore
```

This opens a browser for OAuth authentication. Once complete, your credentials are securely stored.

View your authentication status:

```bash
shopline auth status
```

### Using Multiple Stores

If you have multiple store profiles, specify which one to use:

```bash
shopline orders list --store mystore
# or
export SHOPLINE_STORE=mystore
shopline orders list
```

## Commands

### Orders & Fulfillment

| Command | Description |
|---------|-------------|
| `orders` | Manage orders |
| `draft-orders` | Manage draft orders |
| `fulfillments` | Manage fulfillments |
| `fulfillment-orders` | Manage fulfillment orders |
| `fulfillment-services` | Manage fulfillment services |
| `refunds` | Manage order refunds |
| `return-orders` | Manage return orders |
| `shipments` | Manage shipments |
| `order-risks` | Manage order risk assessments |
| `order-attribution` | Manage order attribution tracking |
| `abandoned-checkouts` | Manage abandoned checkouts |

### Products & Catalog

| Command | Description |
|---------|-------------|
| `products` | Manage products |
| `product-listings` | Manage product listings (products published to sales channels) |
| `product-reviews` | Manage product reviews |
| `product-subscriptions` | Manage product subscriptions |
| `collections` | Manage product collections |
| `smart-collections` | Manage smart collections (auto-populated based on rules) |
| `categories` | Manage product categories |
| `taxonomies` | Manage product taxonomies/categories |
| `tags` | Manage product tags |
| `labels` | Manage product labels |
| `medias` | Manage product media files |
| `size-charts` | Manage size charts |
| `addon-products` | Manage add-on product bundles |

### Inventory & Warehouses

| Command | Description |
|---------|-------------|
| `inventory` | Manage inventory levels |
| `inventory-levels` | Manage inventory levels |
| `warehouses` | Manage warehouses |
| `locations` | Manage store locations |
| `purchase-orders` | Manage purchase orders |

### Customers & Membership

| Command | Description |
|---------|-------------|
| `customers` | Manage customers |
| `customer-addresses` | Manage customer addresses |
| `customer-groups` | Manage customer groups |
| `customer-blacklist` | Manage customer blacklist |
| `customer-saved-searches` | Manage customer saved searches |
| `member-points` | Manage customer member points |
| `membership` | Manage membership tiers |
| `store-credits` | Manage customer store credits |
| `wish-lists` | Manage customer wish lists |
| `subscriptions` | Manage customer subscriptions |
| `conversations` | Manage customer conversations/chat |
| `cdp` | Access Customer Data Platform analytics |

### Pricing & Promotions

| Command | Description |
|---------|-------------|
| `price-rules` | Manage price rules |
| `coupons` | Manage coupons |
| `discount-codes` | Manage discount codes |
| `user-coupons` | Manage user-assigned coupons |
| `promotions` | Manage promotions |
| `sales` | Manage sale campaigns |
| `flash-price` | Manage flash sale pricing |
| `gifts` | Manage gift promotions |
| `gift-cards` | Manage gift cards |
| `selling-plans` | Manage selling plan configurations |

### B2B & Catalogs

| Command | Description |
|---------|-------------|
| `catalog-pricing` | Manage B2B catalog pricing |
| `company-catalogs` | Manage B2B company catalogs |
| `company-credits` | Manage B2B company credits |

### Shipping & Delivery

| Command | Description |
|---------|-------------|
| `shipping-zones` | Manage shipping zones |
| `carrier-services` | Manage carrier services |
| `local-delivery` | Manage local delivery options |
| `pickup` | Manage store pickup locations |

### Payments & Finance

| Command | Description |
|---------|-------------|
| `payments` | Manage payments |
| `payouts` | Manage payment payouts |
| `transactions` | Manage payment transactions |
| `disputes` | Manage payment disputes |
| `balance` | Manage account balance |
| `taxes` | Manage tax settings |
| `tax-services` | Manage tax service providers |
| `currencies` | Manage currencies |

### Channels & Markets

| Command | Description |
|---------|-------------|
| `channels` | Manage sales channels |
| `channel-products` | Manage multi-channel product listings |
| `markets` | Manage markets (regions) |
| `countries` | Manage countries |
| `domains` | Manage domains |

### Marketing & Affiliates

| Command | Description |
|---------|-------------|
| `affiliate-campaigns` | Manage affiliate marketing campaigns |
| `marketing-events` | Manage marketing event tracking |

### Content & Themes

| Command | Description |
|---------|-------------|
| `themes` | Manage themes |
| `assets` | Manage theme assets |
| `pages` | Manage pages |
| `blogs` | Manage blogs |
| `articles` | Manage blog articles |
| `redirects` | Manage URL redirects |
| `files` | Manage files |

### Storefront API

| Command | Description |
|---------|-------------|
| `storefront-products` | View storefront product information |
| `storefront-promotions` | View storefront promotion information |
| `storefront-carts` | Manage storefront shopping carts |
| `storefront-tokens` | Manage storefront access tokens |
| `storefront-oauth` | Manage storefront OAuth clients |

### Store Settings & Configuration

| Command | Description |
|---------|-------------|
| `shop` | Manage shop settings |
| `settings` | Manage store settings |
| `checkout-settings` | Manage checkout settings |
| `metafields` | Manage metafields |
| `metafield-definitions` | Manage metafield definitions |
| `custom-fields` | Manage custom field definitions |
| `script-tags` | Manage script tags |

### Staff & Operations

| Command | Description |
|---------|-------------|
| `staffs` | Manage staff accounts |
| `merchants` | View merchant information |
| `operation-logs` | View operation audit logs |

### Integrations & API

| Command | Description |
|---------|-------------|
| `webhooks` | Manage webhooks |
| `tokens` | Manage API tokens |
| `multipass` | Manage multipass authentication |
| `bulk-operations` | Manage bulk operations |

### Authentication

| Command | Description |
|---------|-------------|
| `auth` | Manage authentication (add, list, remove, status) |

### Utilities

| Command | Description |
|---------|-------------|
| `completion` | Generate the autocompletion script for the specified shell |
| `help` | Help about any command |

## Global Flags

```
-s, --store string     Store profile name (or set SHOPLINE_STORE)
-o, --output string    Output format: text, json (default "text")
    --query string     JQ filter for JSON output
    --color string     Color mode: auto, always, never (default "auto")
    --limit int        Limit number of results
    --sort-by string   Field to sort by
    --desc             Sort in descending order
    --dry-run          Preview changes without executing them
-y, --yes              Skip confirmation prompts
-v, --version          version for shopline
-h, --help             help for shopline
```

## Examples

```bash
# List recent orders
shopline orders list

# Get order details as JSON
shopline orders get ORDER_ID -o json

# List products filtered by vendor
shopline products list --vendor "ACME"

# Adjust inventory
shopline inventory adjust INVENTORY_ID --delta -5

# Create a webhook
shopline webhooks create --topic orders/create --address https://example.com/webhook

# List customers with JSON output and JQ filter
shopline customers list -o json --query '.[] | {id, email}'

# Dry-run an order cancellation
shopline orders cancel ORDER_ID --dry-run

# Generate shell completions
shopline completion bash > ~/.bashrc.d/shopline
shopline completion zsh > ~/.zsh/completions/_shopline
```

## Development

### Prerequisites

- Go 1.21+
- golangci-lint
- gofumpt
- goimports

Install development tools:

```bash
make setup
```

### Build

```bash
make build
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```

### Format

```bash
make fmt
```

### CI

Run all checks (format, lint, test):

```bash
make ci
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run `make ci` to ensure all checks pass
5. Commit your changes (`git commit -m 'feat: add my feature'`)
6. Push to the branch (`git push origin feature/my-feature`)
7. Open a Pull Request

Please follow [Conventional Commits](https://www.conventionalcommits.org/) for commit messages.

## License

MIT License - see [LICENSE](LICENSE) for details.
