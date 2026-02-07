package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/batch"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// CredentialStore defines the interface for credential storage operations.
type CredentialStore interface {
	List() ([]string, error)
	Get(name string) (*secrets.StoreCredentials, error)
}

// StoreFactory is a function that creates a credential store.
type StoreFactory func() (CredentialStore, error)

// ClientFactory is a function that creates an API client.
type ClientFactory func(handle, accessToken string) api.APIClient

// clientFactory allows overriding the client creation for testing.
var clientFactory ClientFactory = defaultClientFactory

// secretsStoreFactory allows overriding the secrets store creation for testing.
var secretsStoreFactory StoreFactory = defaultSecretsStoreFactory

// formatterWriter is the output writer for formatters (can be overridden in tests).
var formatterWriter io.Writer = os.Stdout

func defaultClientFactory(handle, accessToken string) api.APIClient {
	return api.NewClient(handle, accessToken)
}

func defaultSecretsStoreFactory() (CredentialStore, error) {
	return secrets.NewStore()
}

var ordersCmd = &cobra.Command{
	Use:   "orders",
	Short: "Manage orders",
}

var ordersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		limit, _ := cmd.Flags().GetInt("limit")

		opts := &api.OrdersListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}
		if from != "" {
			since, err := parseTimeFlag(from, "from")
			if err != nil {
				return err
			}
			opts.Since = since
		}
		if to != "" {
			until, err := parseTimeFlag(to, "to")
			if err != nil {
				return err
			}
			opts.Until = until
		}

		// If --limit is set, treat it as the maximum number of orders to return.
		// This is agent-friendly (fetches multiple pages if needed) and also works
		// around APIs that cap/ignore page_size.
		resp := &api.OrdersListResponse{}
		if limit > 0 {
			curPage := opts.Page
			perPage := opts.PageSize
			if perPage <= 0 || perPage > limit {
				perPage = limit
			}

			items := make([]api.OrderSummary, 0, limit)
			totalCount := 0
			hasMore := false
			var pagination api.Pagination

			for len(items) < limit {
				pageOpts := *opts
				pageOpts.Page = curPage
				pageOpts.PageSize = perPage

				pageResp, err := client.ListOrders(cmd.Context(), &pageOpts)
				if err != nil {
					return fmt.Errorf("failed to list orders: %w", err)
				}
				if totalCount == 0 {
					totalCount = pageResp.TotalCount
					pagination = pageResp.Pagination
				}
				items = append(items, pageResp.Items...)
				hasMore = pageResp.HasMore

				if !pageResp.HasMore || len(pageResp.Items) == 0 {
					break
				}
				curPage++
			}

			if len(items) > limit {
				items = items[:limit]
				hasMore = true
			}

			resp.Items = items
			resp.Page = opts.Page
			resp.PageSize = perPage
			resp.TotalCount = totalCount
			resp.HasMore = hasMore
			resp.Pagination = pagination
			resp.Pagination.CurrentPage = opts.Page
			resp.Pagination.PerPage = perPage
		} else {
			r, err := client.ListOrders(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("failed to list orders: %w", err)
			}
			resp = r
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		// Optional expansion/enrichment for list JSON output.
		//
		// Notes:
		// - List endpoints do not include line items, so expanding products/customers
		//   requires per-order detail calls.
		// - We keep this JSON-only to avoid changing the default table output.
		expands, _ := cmd.Flags().GetStringSlice("expand")
		jobs, _ := cmd.Flags().GetInt("jobs")
		if jobs <= 0 {
			jobs = 4
		}

		expandDetails := false
		expandCustomer := false
		expandProducts := false
		for _, e := range expands {
			switch strings.ToLower(strings.TrimSpace(e)) {
			case "":
				continue
			case "details", "detail", "order":
				expandDetails = true
			case "customer":
				expandCustomer = true
			case "products", "product":
				expandProducts = true
			default:
				return fmt.Errorf("invalid --expand value %q (supported: details, customer, products)", e)
			}
		}
		if expandCustomer || expandProducts {
			expandDetails = true
		}

		if outputFormat == "json" && expandDetails && len(resp.Items) > 0 {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			sem := make(chan struct{}, jobs)
			var wg sync.WaitGroup
			var mu sync.Mutex
			var firstErr error

			details := make([]api.Order, len(resp.Items))

			productCache := map[string]*api.Product{}
			var productMu sync.Mutex

			customerCache := map[string]*api.Customer{}
			var customerMu sync.Mutex

			for i := range resp.Items {
				wg.Add(1)
				sem <- struct{}{}
				go func(i int, orderID string) {
					defer wg.Done()
					defer func() { <-sem }()

					if firstErr != nil {
						return
					}

					o, err := client.GetOrder(ctx, orderID)
					if err != nil {
						mu.Lock()
						if firstErr == nil {
							firstErr = fmt.Errorf("failed to expand order details for %s: %w", orderID, err)
							cancel()
						}
						mu.Unlock()
						return
					}

					if expandCustomer && o.Customer == nil && o.CustomerID != "" {
						cid := o.CustomerID
						var c *api.Customer
						customerMu.Lock()
						c = customerCache[cid]
						customerMu.Unlock()
						if c == nil {
							cc, err := client.GetCustomer(ctx, cid)
							if err != nil {
								mu.Lock()
								if firstErr == nil {
									firstErr = fmt.Errorf("failed to expand customer for order %s: %w", orderID, err)
									cancel()
								}
								mu.Unlock()
								return
							}
							c = cc
							customerMu.Lock()
							customerCache[cid] = cc
							customerMu.Unlock()
						}
						o.Customer = c
					}

					if expandProducts && len(o.LineItems) > 0 {
						for li := range o.LineItems {
							pid := strings.TrimSpace(o.LineItems[li].ProductID)
							if pid == "" {
								continue
							}

							var p *api.Product
							productMu.Lock()
							p = productCache[pid]
							productMu.Unlock()
							if p == nil {
								pp, err := client.GetProduct(ctx, pid)
								if err != nil {
									mu.Lock()
									if firstErr == nil {
										firstErr = fmt.Errorf("failed to expand products for order %s: %w", orderID, err)
										cancel()
									}
									mu.Unlock()
									return
								}
								p = pp
								productMu.Lock()
								productCache[pid] = pp
								productMu.Unlock()
							}
							o.LineItems[li].Product = p
						}
					}

					// Copy out to maintain original ordering.
					details[i] = *o
				}(i, resp.Items[i].ID)
			}

			wg.Wait()
			if firstErr != nil {
				return firstErr
			}

			expanded := &api.ListResponse[api.Order]{
				Items:      details,
				Pagination: resp.Pagination,
				Page:       resp.Page,
				PageSize:   resp.PageSize,
				TotalCount: resp.TotalCount,
				HasMore:    resp.HasMore,
			}
			return formatter.JSON(expanded)
		}

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ORDER", "NUMBER", "STATUS", "TOTAL", "CUSTOMER", "CREATED"}
		var rows [][]string
		for _, o := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("order", o.ID),
				o.OrderNumber,
				o.Status,
				o.TotalPrice + " " + o.Currency,
				o.CustomerEmail,
				o.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		if resp.TotalCount > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d orders\n", len(resp.Items), resp.TotalCount)
		} else {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d orders\n", len(resp.Items))
		}
		return nil
	},
}

var ordersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get order details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return handleError(cmd, err, "orders", "")
		}

		orderID := args[0]
		order, err := client.GetOrder(cmd.Context(), orderID)
		if err != nil {
			return handleError(cmd, err, "orders", orderID)
		}

		// Optional expansion/enrichment (opt-in because it can add API calls).
		expands, expandErr := cmd.Flags().GetStringSlice("expand")
		if expandErr == nil && len(expands) > 0 {
			var expandCustomer, expandProducts bool
			for _, e := range expands {
				switch strings.ToLower(strings.TrimSpace(e)) {
				case "":
					continue
				case "customer":
					expandCustomer = true
				case "products", "product":
					expandProducts = true
				default:
					return fmt.Errorf("invalid --expand value %q (supported: customer, products)", e)
				}
			}

			if expandCustomer && order.Customer == nil && order.CustomerID != "" {
				c, err := client.GetCustomer(cmd.Context(), order.CustomerID)
				if err != nil {
					return fmt.Errorf("failed to expand customer: %w", err)
				}
				order.Customer = c
			}

			if expandProducts && len(order.LineItems) > 0 {
				cache := map[string]*api.Product{}
				for i := range order.LineItems {
					pid := strings.TrimSpace(order.LineItems[i].ProductID)
					if pid == "" {
						continue
					}
					if p, ok := cache[pid]; ok {
						order.LineItems[i].Product = p
						continue
					}
					p, err := client.GetProduct(cmd.Context(), pid)
					if err != nil {
						return fmt.Errorf("failed to expand products: %w", err)
					}
					cache[pid] = p
					order.LineItems[i].Product = p
				}
			}
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			// Make jq safer: null line_items is a footgun for `.line_items[]`.
			if order.LineItems == nil {
				order.LineItems = []api.OrderLineItem{}
			}
			return formatter.JSON(order)
		}

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Order ID:       %s\n", order.ID)
		_, _ = fmt.Fprintf(out, "Order Number:   %s\n", order.OrderNumber)
		_, _ = fmt.Fprintf(out, "Status:         %s\n", order.Status)
		_, _ = fmt.Fprintf(out, "Payment:        %s\n", order.PaymentStatus)
		_, _ = fmt.Fprintf(out, "Fulfillment:    %s\n", order.FulfillStatus)
		_, _ = fmt.Fprintf(out, "Total:          %s %s\n", order.TotalPrice, order.Currency)
		_, _ = fmt.Fprintf(out, "Customer:       %s <%s>\n", order.CustomerName, order.CustomerEmail)
		_, _ = fmt.Fprintf(out, "Created:        %s\n", order.CreatedAt.Format(time.RFC3339))
		if len(order.LineItems) > 0 {
			_, _ = fmt.Fprintln(out, "\nLine items:")
			for _, li := range order.LineItems {
				title := li.Title
				if title == "" {
					title = li.Name
				}
				vendor := li.Vendor
				if vendor == "" && li.Product != nil {
					vendor = li.Product.Vendor
				}
				if vendor != "" {
					_, _ = fmt.Fprintf(out, "  %dx %s (%s)\n", li.Quantity, title, vendor)
				} else {
					_, _ = fmt.Fprintf(out, "  %dx %s\n", li.Quantity, title)
				}
			}
		}
		if order.Customer != nil {
			_, _ = fmt.Fprintln(out, "\nExpanded customer:")
			_, _ = fmt.Fprintf(out, "  ID:    %s\n", order.Customer.ID)
			if order.Customer.Email != "" {
				_, _ = fmt.Fprintf(out, "  Email: %s\n", order.Customer.Email)
			}
			if order.Customer.Phone != "" {
				_, _ = fmt.Fprintf(out, "  Phone: %s\n", order.Customer.Phone)
			}
		}
		return nil
	},
}

var ordersCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel an order",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		batchFile, _ := cmd.Flags().GetString("batch")
		if batchFile != "" {
			return cancelOrdersBatch(cmd, batchFile)
		}

		if len(args) == 0 {
			return fmt.Errorf("order ID required (or use --batch)")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would cancel order %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Cancel order %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
				return nil
			}
		}

		if err := client.CancelOrder(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to cancel order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Order %s cancelled.\n", args[0])
		return nil
	},
}

func cancelOrdersBatch(cmd *cobra.Command, filename string) error {
	client, err := getClient(cmd)
	if err != nil {
		return err
	}

	items, err := batch.ReadItems(filename)
	if err != nil {
		return fmt.Errorf("failed to read batch file: %w", err)
	}

	var results []batch.Result
	for i, item := range items {
		var input struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(item, &input); err != nil {
			results = append(results, batch.Result{Index: i, Success: false, Error: "invalid JSON: " + err.Error()})
			continue
		}
		if input.ID == "" {
			results = append(results, batch.Result{Index: i, Success: false, Error: "missing id field"})
			continue
		}

		if err := client.CancelOrder(cmd.Context(), input.ID); err != nil {
			results = append(results, batch.Result{ID: input.ID, Index: i, Success: false, Error: err.Error()})
		} else {
			results = append(results, batch.Result{ID: input.ID, Index: i, Success: true})
		}
	}

	return batch.WriteResults(os.Stdout, results)
}

// --- Create / Update ---

var ordersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an order (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would create order\n")
			return nil
		}

		var req api.OrderCreateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		order, err := client.CreateOrder(cmd.Context(), &req)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(order)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created order %s\n", order.ID)
		return nil
	},
}

var ordersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an order (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would update order %s\n", args[0])
			return nil
		}

		var req api.OrderUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		order, err := client.UpdateOrder(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(order)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated order %s\n", order.ID)
		return nil
	},
}

// --- Search ---

var ordersSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("query")
		status, _ := cmd.Flags().GetString("status")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.OrderSearchOptions{
			Query:    query,
			Status:   status,
			Page:     page,
			PageSize: pageSize,
		}
		if from != "" {
			since, err := parseTimeFlag(from, "from")
			if err != nil {
				return err
			}
			opts.Since = since
		}
		if to != "" {
			until, err := parseTimeFlag(to, "to")
			if err != nil {
				return err
			}
			opts.Until = until
		}

		resp, err := client.SearchOrders(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to search orders: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ORDER", "NUMBER", "STATUS", "TOTAL", "CUSTOMER", "CREATED"}
		var rows [][]string
		for _, o := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("order", o.ID),
				o.OrderNumber,
				o.Status,
				o.TotalPrice + " " + o.Currency,
				o.CustomerEmail,
				o.CreatedAt.Format("2006-01-02 15:04"),
			})
		}
		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d orders\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

// --- Split ---

var ordersSplitCmd = &cobra.Command{
	Use:   "split <id>",
	Short: "Split an order into two orders",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lineItemIDsStr, _ := cmd.Flags().GetString("line-item-ids")
		if lineItemIDsStr == "" {
			return fmt.Errorf("--line-item-ids must not be empty")
		}
		lineItemIDs := strings.Split(lineItemIDsStr, ",")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.SplitOrder(cmd.Context(), args[0], lineItemIDs)
		if err != nil {
			return fmt.Errorf("failed to split order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Split order %s into new order %s\n", resp.OriginalOrder.ID, resp.NewOrder.ID)
		return nil
	},
}

// --- Status update commands ---

var ordersUpdateStatusCmd = &cobra.Command{
	Use:   "update-status <id>",
	Short: "Update order status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would update status for order %s\n", args[0])
			return nil
		}

		status, _ := cmd.Flags().GetString("status")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		order, err := client.UpdateOrderStatus(cmd.Context(), args[0], status)
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(order)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated order %s status to %s\n", order.ID, order.Status)
		return nil
	},
}

var ordersUpdateDeliveryStatusCmd = &cobra.Command{
	Use:   "update-delivery-status <id>",
	Short: "Update order delivery status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would update delivery status for order %s\n", args[0])
			return nil
		}

		status, _ := cmd.Flags().GetString("status")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		order, err := client.UpdateOrderDeliveryStatus(cmd.Context(), args[0], status)
		if err != nil {
			return fmt.Errorf("failed to update order delivery status: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(order)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated order %s delivery status\n", order.ID)
		return nil
	},
}

var ordersUpdatePaymentStatusCmd = &cobra.Command{
	Use:   "update-payment-status <id>",
	Short: "Update order payment status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would update payment status for order %s\n", args[0])
			return nil
		}

		status, _ := cmd.Flags().GetString("status")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		order, err := client.UpdateOrderPaymentStatus(cmd.Context(), args[0], status)
		if err != nil {
			return fmt.Errorf("failed to update order payment status: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(order)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated order %s payment status\n", order.ID)
		return nil
	},
}

// --- Shipment commands ---

var ordersExecuteShipmentCmd = &cobra.Command{
	Use:   "execute-shipment <id>",
	Short: "Execute shipment for an order (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.ExecuteShipment(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to execute shipment: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

var ordersBulkExecuteShipmentCmd = &cobra.Command{
	Use:   "bulk-execute-shipment",
	Short: "Execute shipments for multiple orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		idsStr, _ := cmd.Flags().GetString("order-ids")
		if idsStr == "" {
			return fmt.Errorf("--order-ids must not be empty")
		}
		orderIDs := strings.Split(idsStr, ",")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.BulkExecuteShipment(cmd.Context(), orderIDs)
		if err != nil {
			return fmt.Errorf("failed to bulk execute shipments: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Executed shipments: %d successful, %d failed\n",
			len(resp.Successful), len(resp.Failed))
		return nil
	},
}

// --- Tags subcommands ---

var ordersTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage order tags",
}

var ordersTagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all order tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.ListOrderTags(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list order tags: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

var ordersTagsGetCmd = &cobra.Command{
	Use:   "get <order-id>",
	Short: "Get tags for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetOrderTags(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get order tags: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Tags: %s\n", strings.Join(resp.Tags, ", "))
		return nil
	},
}

var ordersTagsUpdateCmd = &cobra.Command{
	Use:   "update <order-id>",
	Short: "Update tags for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would update tags for order %s\n", args[0])
			return nil
		}

		tagsStr, _ := cmd.Flags().GetString("tags")
		if tagsStr == "" {
			return fmt.Errorf("--tags must not be empty")
		}
		tags := strings.Split(tagsStr, ",")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		order, err := client.UpdateOrderTags(cmd.Context(), args[0], tags)
		if err != nil {
			return fmt.Errorf("failed to update order tags: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(order)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated tags for order %s\n", order.ID)
		return nil
	},
}

// --- Transactions / Action Logs / Labels / Messages ---

var ordersTransactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Get order transactions (raw JSON)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetOrderTransactions(cmd.Context(), nil)
		if err != nil {
			return fmt.Errorf("failed to get order transactions: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

var ordersActionLogsCmd = &cobra.Command{
	Use:   "action-logs <id>",
	Short: "Get action logs for an order (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetOrderActionLogs(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get order action logs: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

var ordersLabelsCmd = &cobra.Command{
	Use:   "labels",
	Short: "Get order delivery labels (raw JSON)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetOrderLabels(cmd.Context(), nil)
		if err != nil {
			return fmt.Errorf("failed to get order labels: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

var ordersCreateMessageCmd = &cobra.Command{
	Use:   "create-message <id>",
	Short: "Create a message on an order (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would create message for order %s\n", args[0])
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.PostOrderMessage(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to create order message: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

// --- Archived orders ---

var ordersArchivedCmd = &cobra.Command{
	Use:   "archived-orders",
	Short: "Manage archived orders",
}

var ordersArchivedListCmd = &cobra.Command{
	Use:   "list",
	Short: "List archived orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")

		opts := &api.ArchivedOrdersListOptions{
			Page:     page,
			PageSize: pageSize,
		}
		if from != "" {
			since, err := parseTimeFlag(from, "from")
			if err != nil {
				return err
			}
			opts.Since = since
		}
		if to != "" {
			until, err := parseTimeFlag(to, "to")
			if err != nil {
				return err
			}
			opts.Until = until
		}

		resp, err := client.ListArchivedOrders(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list archived orders: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ORDER", "NUMBER", "STATUS", "TOTAL", "CUSTOMER", "CREATED"}
		var rows [][]string
		for _, o := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("order", o.ID),
				o.OrderNumber,
				o.Status,
				o.TotalPrice + " " + o.Currency,
				o.CustomerEmail,
				o.CreatedAt.Format("2006-01-02 15:04"),
			})
		}
		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d archived orders\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var ordersArchivedCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an archived orders report (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would create archived orders report\n")
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.CreateArchivedOrdersReport(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create archived orders report: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

// --- Delivery subcommands ---

var ordersDeliveryCmd = &cobra.Command{
	Use:   "delivery",
	Short: "Manage order delivery",
}

var ordersDeliveryGetCmd = &cobra.Command{
	Use:   "get <order-id>",
	Short: "Get delivery information for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		delivery, err := client.GetOrderDelivery(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get order delivery: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(delivery)
		}

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Delivery ID:      %s\n", delivery.ID)
		_, _ = fmt.Fprintf(out, "Order ID:         %s\n", delivery.OrderID)
		_, _ = fmt.Fprintf(out, "Status:           %s\n", delivery.Status)
		_, _ = fmt.Fprintf(out, "Carrier:          %s\n", delivery.Carrier)
		_, _ = fmt.Fprintf(out, "Tracking Number:  %s\n", delivery.TrackingNumber)
		if delivery.TrackingURL != "" {
			_, _ = fmt.Fprintf(out, "Tracking URL:     %s\n", delivery.TrackingURL)
		}
		return nil
	},
}

var ordersDeliveryUpdateCmd = &cobra.Command{
	Use:   "update <order-id>",
	Short: "Update delivery information for an order (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would update delivery for order %s\n", args[0])
			return nil
		}

		var req api.OrderDeliveryUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		delivery, err := client.UpdateOrderDelivery(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update order delivery: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(delivery)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated delivery for order %s\n", delivery.OrderID)
		return nil
	},
}

func getClient(cmd *cobra.Command) (api.APIClient, error) {
	storeName, _ := cmd.Flags().GetString("store")
	if storeName == "" {
		storeName = os.Getenv("SHOPLINE_STORE")
	}

	store, err := secretsStoreFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to open credential store: %w", err)
	}

	if storeName == "" {
		names, err := store.List()
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			return nil, fmt.Errorf("no store profiles configured, run 'shopline auth login'")
		}
		if len(names) == 1 {
			storeName = names[0]
		} else {
			return nil, fmt.Errorf("multiple profiles configured, use --store to select one")
		}
	}

	creds, err := store.Get(storeName)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %s", storeName)
	}

	return clientFactory(creds.Handle, creds.AccessToken), nil
}

func getFormatter(cmd *cobra.Command) *outfmt.Formatter {
	outputFormat, _ := cmd.Flags().GetString("output")
	colorMode, _ := cmd.Flags().GetString("color")
	query, _ := cmd.Flags().GetString("query")
	itemsOnly, _ := cmd.Flags().GetBool("items-only")

	format := outfmt.FormatText
	if outputFormat == "json" {
		format = outfmt.FormatJSON
	}

	w := formatterWriter
	// In real CLI execution, cobra's OutOrStdout may be configured (tests, piping, etc).
	// Keep formatterWriter override for unit tests, but otherwise prefer cmd.OutOrStdout.
	if formatterWriter == os.Stdout && cmd != nil && cmd.OutOrStdout() != nil {
		w = cmd.OutOrStdout()
	}

	f := outfmt.New(w, format, colorMode)
	if prefix := idPrefixForCommand(cmd); prefix != "" {
		f = f.WithIDPrefix(prefix)
	}
	if query != "" {
		f = f.WithQuery(query)
	}
	if itemsOnly {
		f = f.WithItemsOnly(true)
	}
	return f
}

func init() {
	rootCmd.AddCommand(ordersCmd)

	ordersCmd.AddCommand(ordersListCmd)
	ordersListCmd.Flags().String("status", "", "Filter by status")
	ordersListCmd.Flags().String("from", "", "Filter by created date from (YYYY-MM-DD or RFC3339)")
	ordersListCmd.Flags().String("to", "", "Filter by created date to (YYYY-MM-DD or RFC3339)")
	ordersListCmd.Flags().Int("page", 1, "Page number")
	ordersListCmd.Flags().Int("page-size", 20, "Results per page")
	ordersListCmd.Flags().StringSlice("expand", nil, "Expand related resources: details, customer, products (adds API calls)")
	ordersListCmd.Flags().Int("jobs", 4, "Max concurrent API calls for --expand details")

	ordersCmd.AddCommand(ordersGetCmd)
	ordersGetCmd.Flags().StringSlice("expand", nil, "Expand related resources: customer, products (adds API calls)")
	ordersCmd.AddCommand(ordersCancelCmd)
	ordersCancelCmd.Flags().String("batch", "", "Batch input file (JSON array or NDJSON)")

	// Create / Update
	ordersCmd.AddCommand(ordersCreateCmd)
	addJSONBodyFlags(ordersCreateCmd)

	ordersCmd.AddCommand(ordersUpdateCmd)
	addJSONBodyFlags(ordersUpdateCmd)

	// Search
	ordersCmd.AddCommand(ordersSearchCmd)
	ordersSearchCmd.Flags().String("query", "", "Search query")
	ordersSearchCmd.Flags().String("status", "", "Filter by status")
	ordersSearchCmd.Flags().String("from", "", "Filter by created date from (YYYY-MM-DD or RFC3339)")
	ordersSearchCmd.Flags().String("to", "", "Filter by created date to (YYYY-MM-DD or RFC3339)")
	ordersSearchCmd.Flags().Int("page", 1, "Page number")
	ordersSearchCmd.Flags().Int("page-size", 20, "Results per page")

	// Split
	ordersCmd.AddCommand(ordersSplitCmd)
	ordersSplitCmd.Flags().String("line-item-ids", "", "Comma-separated line item IDs to split (required)")
	_ = ordersSplitCmd.MarkFlagRequired("line-item-ids")

	// Status updates
	ordersCmd.AddCommand(ordersUpdateStatusCmd)
	ordersUpdateStatusCmd.Flags().String("status", "", "New status (required)")
	_ = ordersUpdateStatusCmd.MarkFlagRequired("status")

	ordersCmd.AddCommand(ordersUpdateDeliveryStatusCmd)
	ordersUpdateDeliveryStatusCmd.Flags().String("status", "", "New delivery status (required)")
	_ = ordersUpdateDeliveryStatusCmd.MarkFlagRequired("status")

	ordersCmd.AddCommand(ordersUpdatePaymentStatusCmd)
	ordersUpdatePaymentStatusCmd.Flags().String("status", "", "New payment status (required)")
	_ = ordersUpdatePaymentStatusCmd.MarkFlagRequired("status")

	// Shipment
	ordersCmd.AddCommand(ordersExecuteShipmentCmd)
	addJSONBodyFlags(ordersExecuteShipmentCmd)

	ordersCmd.AddCommand(ordersBulkExecuteShipmentCmd)
	ordersBulkExecuteShipmentCmd.Flags().String("order-ids", "", "Comma-separated order IDs (required)")
	_ = ordersBulkExecuteShipmentCmd.MarkFlagRequired("order-ids")

	// Tags subcommands
	ordersCmd.AddCommand(ordersTagsCmd)
	ordersTagsCmd.AddCommand(ordersTagsListCmd)
	ordersTagsCmd.AddCommand(ordersTagsGetCmd)
	ordersTagsCmd.AddCommand(ordersTagsUpdateCmd)
	ordersTagsUpdateCmd.Flags().String("tags", "", "Comma-separated tags (required)")
	_ = ordersTagsUpdateCmd.MarkFlagRequired("tags")

	// Transactions / Action Logs / Labels / Messages
	ordersCmd.AddCommand(ordersTransactionsCmd)
	ordersCmd.AddCommand(ordersActionLogsCmd)
	ordersCmd.AddCommand(ordersLabelsCmd)
	ordersCmd.AddCommand(ordersCreateMessageCmd)
	addJSONBodyFlags(ordersCreateMessageCmd)

	// Archived orders
	ordersCmd.AddCommand(ordersArchivedCmd)
	ordersArchivedCmd.AddCommand(ordersArchivedListCmd)
	ordersArchivedListCmd.Flags().Int("page", 1, "Page number")
	ordersArchivedListCmd.Flags().Int("page-size", 20, "Results per page")
	ordersArchivedListCmd.Flags().String("from", "", "Filter by created date from (YYYY-MM-DD or RFC3339)")
	ordersArchivedListCmd.Flags().String("to", "", "Filter by created date to (YYYY-MM-DD or RFC3339)")

	ordersArchivedCmd.AddCommand(ordersArchivedCreateCmd)
	addJSONBodyFlags(ordersArchivedCreateCmd)

	// Delivery subcommands
	ordersCmd.AddCommand(ordersDeliveryCmd)
	ordersDeliveryCmd.AddCommand(ordersDeliveryGetCmd)
	ordersDeliveryCmd.AddCommand(ordersDeliveryUpdateCmd)
	addJSONBodyFlags(ordersDeliveryUpdateCmd)

	schema.Register(schema.Resource{
		Name:        "orders",
		Description: "Manage customer orders",
		Commands: []string{
			"list", "get", "create", "update", "cancel", "search", "split",
			"update-status", "update-delivery-status", "update-payment-status",
			"execute-shipment", "bulk-execute-shipment",
			"tags list", "tags get", "tags update",
			"transactions", "action-logs", "labels", "create-message",
			"archived-orders list", "archived-orders create",
			"delivery get", "delivery update",
			"metafields", "app-metafields", "item-metafields", "item-app-metafields",
		},
		IDPrefix: "order",
	})
}
