package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/batch"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

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
		email, _ := cmd.Flags().GetString("email")
		customerID, _ := cmd.Flags().GetString("customer-id")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		limit, _ := cmd.Flags().GetInt("limit")

		if strings.TrimSpace(email) != "" && strings.TrimSpace(customerID) != "" {
			return fmt.Errorf("--email and --customer-id cannot be used together (use one)")
		}

		var query string
		// Shopline search API accepts raw email as a query string.
		// Special characters (e.g. + tags) are handled server-side.
		if strings.TrimSpace(email) != "" {
			query = strings.TrimSpace(email)
		} else if strings.TrimSpace(customerID) != "" {
			query = strings.TrimSpace(customerID)
		}

		var since *time.Time
		var until *time.Time
		if from != "" {
			parsedSince, err := parseTimeFlag(from, "from")
			if err != nil {
				return err
			}
			since = parsedSince
		}
		if to != "" {
			parsedUntil, err := parseTimeFlag(to, "to")
			if err != nil {
				return err
			}
			until = parsedUntil
		}

		opts := &api.OrdersListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			Since:    since,
			Until:    until,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		var (
			resp     *api.ListResponse[api.OrderSummary]
			fetchErr error
		)
		if query != "" {
			searchOpts := &api.OrderSearchOptions{
				Query:    query,
				Status:   status,
				Since:    since,
				Until:    until,
				Page:     page,
				PageSize: pageSize,
			}
			resp, fetchErr = fetchList(
				cmd.Context(), limit, searchOpts.Page, searchOpts.PageSize,
				func() (*api.ListResponse[api.OrderSummary], error) {
					return client.SearchOrders(cmd.Context(), searchOpts)
				},
				func(page, size int) (*api.ListResponse[api.OrderSummary], error) {
					pageOpts := *searchOpts
					pageOpts.Page = page
					pageOpts.PageSize = size
					return client.SearchOrders(cmd.Context(), &pageOpts)
				},
				"failed to search orders",
			)
		} else {
			resp, fetchErr = fetchList(
				cmd.Context(), limit, opts.Page, opts.PageSize,
				func() (*api.ListResponse[api.OrderSummary], error) {
					return client.ListOrders(cmd.Context(), opts)
				},
				func(page, size int) (*api.ListResponse[api.OrderSummary], error) {
					pageOpts := *opts
					pageOpts.Page = page
					pageOpts.PageSize = size
					return client.ListOrders(cmd.Context(), &pageOpts)
				},
				"failed to list orders",
			)
		}
		if fetchErr != nil {
			return fetchErr
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
			var errMu sync.Mutex
			var firstErr error
			setFirstErr := func(err error) {
				errMu.Lock()
				defer errMu.Unlock()
				if firstErr != nil {
					return
				}
				firstErr = err
				cancel()
			}
			getFirstErr := func() error {
				errMu.Lock()
				defer errMu.Unlock()
				return firstErr
			}

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

					if getFirstErr() != nil {
						return
					}

					o, err := client.GetOrder(ctx, orderID)
					if err != nil {
						setFirstErr(fmt.Errorf("failed to expand order details for %s: %w", orderID, err))
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
								setFirstErr(fmt.Errorf("failed to expand customer for order %s: %w", orderID, err))
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
									setFirstErr(fmt.Errorf("failed to expand products for order %s: %w", orderID, err))
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
			if err := getFirstErr(); err != nil {
				return err
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

		light, _ := cmd.Flags().GetBool("light")
		if outputFormat == "json" {
			if light {
				lightItems := toLightSlice(resp.Items, toLightOrderSummary)
				return formatter.JSON(api.ListResponse[lightOrderSummary]{
					Items:      lightItems,
					Pagination: resp.Pagination,
					Page:       resp.Page,
					PageSize:   resp.PageSize,
					TotalCount: resp.TotalCount,
					HasMore:    resp.HasMore,
				})
			}
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
	Use:   "get [id]",
	Short: "Get order details",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return handleError(cmd, err, "orders", "")
		}

		orderID, err := resolveOrArg(cmd, args, func(query string) (string, error) {
			resp, err := client.SearchOrders(cmd.Context(), &api.OrderSearchOptions{
				Query:    query,
				PageSize: 5,
			})
			if err != nil {
				return "", fmt.Errorf("search failed: %w", err)
			}
			if len(resp.Items) == 0 {
				return "", fmt.Errorf("no order found matching %q", query)
			}

			trimmedQuery := strings.TrimSpace(query)
			exactMatches := make([]string, 0, len(resp.Items))
			for _, item := range resp.Items {
				number := strings.TrimSpace(item.OrderNumber)
				id := strings.TrimSpace(item.ID)
				if strings.EqualFold(number, trimmedQuery) || strings.EqualFold(id, trimmedQuery) {
					exactMatches = append(exactMatches, item.ID)
				}
			}
			if len(exactMatches) == 1 {
				return exactMatches[0], nil
			}
			if len(exactMatches) > 1 {
				return "", fmt.Errorf("multiple exact orders found for %q; use an ID: %s", query, strings.Join(exactMatches, ", "))
			}

			if len(resp.Items) > 1 {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %d orders matched, using first\n", len(resp.Items))
			}
			return resp.Items[0].ID, nil
		})
		if err != nil {
			return err
		}

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

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				return formatter.JSON(toLightOrder(order))
			}
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

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would cancel order %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, fmt.Sprintf("Cancel order %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
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

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	var results []batch.Result
	for i, item := range items {
		if err := ctx.Err(); err != nil {
			results = append(results, batch.Result{Index: i, Success: false, Error: "cancelled: " + err.Error()})
			break
		}

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

		if err := client.CancelOrder(ctx, input.ID); err != nil {
			results = append(results, batch.Result{ID: input.ID, Index: i, Success: false, Error: err.Error()})
		} else {
			results = append(results, batch.Result{ID: input.ID, Index: i, Success: true})
		}
	}

	return batch.WriteResults(outWriter(cmd), results)
}

// --- Create / Update ---

var ordersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an order",
	Long:  "Create an order using either --body/--body-file (raw JSON) or individual flags (--email, --note, --tags).",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would create order") {
			return nil
		}

		hasBody := cmd.Flags().Changed("body") || cmd.Flags().Changed("body-file")
		hasFlags := cmd.Flags().Changed("email") || cmd.Flags().Changed("note") || cmd.Flags().Changed("tags")

		if hasBody && hasFlags {
			return fmt.Errorf("use either --body/--body-file or individual flags, not both")
		}
		if !hasBody && !hasFlags {
			return fmt.Errorf("provide order data via --body/--body-file or individual flags (--email, --note, --tags)")
		}

		var req api.OrderCreateRequest
		if hasBody {
			if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
				return err
			}
		} else {
			email, _ := cmd.Flags().GetString("email")
			note, _ := cmd.Flags().GetString("note")
			tagsStr, _ := cmd.Flags().GetString("tags")

			req.CustomerEmail = email
			req.Note = note
			if tagsStr != "" {
				req.Tags = splitTags(tagsStr)
			}
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
	Short: "Update an order",
	Long:  "Update an order using either --body/--body-file (raw JSON) or individual flags (--note, --tags).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update order %s", args[0])) {
			return nil
		}

		hasBody := cmd.Flags().Changed("body") || cmd.Flags().Changed("body-file")
		hasFlags := cmd.Flags().Changed("note") || cmd.Flags().Changed("tags")

		if hasBody && hasFlags {
			return fmt.Errorf("use either --body/--body-file or individual flags, not both")
		}
		if !hasBody && !hasFlags {
			return fmt.Errorf("provide order data via --body/--body-file or individual flags (--note, --tags)")
		}

		var req api.OrderUpdateRequest
		if hasBody {
			if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
				return err
			}
		} else {
			if cmd.Flags().Changed("note") {
				note, _ := cmd.Flags().GetString("note")
				req.Note = &note
			}
			if cmd.Flags().Changed("tags") {
				tagsStr, _ := cmd.Flags().GetString("tags")
				req.Tags = splitTags(tagsStr)
			}
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

// splitTags splits a comma-separated tags string into a trimmed slice.
func splitTags(s string) []string {
	parts := strings.Split(s, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
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

		query, _ := cmd.Flags().GetString("q")
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update status for order %s", args[0])) {
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update delivery status for order %s", args[0])) {
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update payment status for order %s", args[0])) {
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update tags for order %s", args[0])) {
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create message for order %s", args[0])) {
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
		if checkDryRun(cmd, "[DRY-RUN] Would create archived orders report") {
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update delivery for order %s", args[0])) {
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

// --- Admin API commands ---

var ordersCommentCmd = &cobra.Command{
	Use:   "comment <order-id>",
	Short: "Add a comment to an order (via Admin API)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text, _ := cmd.Flags().GetString("text")
		private, _ := cmd.Flags().GetBool("private")

		if text == "" {
			return fmt.Errorf("--text is required")
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would comment on order %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		req := &api.AdminCommentRequest{
			Comment:   text,
			IsPrivate: private,
		}

		result, err := client.CommentOrder(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to comment on order: %w", err)
		}

		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var ordersCommentsCmd = &cobra.Command{
	Use:     "comments <order-id>",
	Aliases: []string{"cmts"},
	Short:   "List comments on an order (via Admin API)",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		comments, err := client.ListOrderComments(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list order comments: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(comments)
		}

		out := outWriter(cmd)
		if len(comments) == 0 {
			_, _ = fmt.Fprintln(out, "No comments found.")
			return nil
		}

		headers := []string{"ID", "AUTHOR", "PRIVATE", "CREATED", "COMMENT"}
		var rows [][]string
		for _, c := range comments {
			private := "no"
			if c.IsPrivate {
				private = "yes"
			}
			// Truncate long comments for table display
			text := c.Comment
			if utf8.RuneCountInString(text) > 60 {
				runes := []rune(text)
				text = string(runes[:57]) + "..."
			}
			rows = append(rows, []string{
				c.ID,
				c.Author,
				private,
				c.CreatedAt,
				text,
			})
		}
		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(out, "\n%d comment(s)\n", len(comments))
		return nil
	},
}

var ordersAdminRefundCmd = &cobra.Command{
	Use:   "admin-refund <order-id>",
	Short: "Issue an admin refund for an order (via Admin API)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		performerID, _ := cmd.Flags().GetString("performer-id")
		amount, _ := cmd.Flags().GetInt("amount")
		paymentUpdatedAt, _ := cmd.Flags().GetString("payment-updated-at")
		remark, _ := cmd.Flags().GetString("remark")

		if amount <= 0 {
			return fmt.Errorf("--amount must be a positive value (in cents)")
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would refund order %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		req := &api.AdminRefundRequest{
			PerformerID:           performerID,
			Amount:                amount,
			OrderPaymentUpdatedAt: paymentUpdatedAt,
			RefundRemark:          remark,
		}

		result, err := client.AdminRefundOrder(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to refund order: %w", err)
		}

		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var ordersReceiptReissueCmd = &cobra.Command{
	Use:   "receipt-reissue <order-id>",
	Short: "Reissue a receipt for an order (via Admin API)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would reissue receipt for order %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		result, err := client.ReissueReceipt(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to reissue receipt: %w", err)
		}

		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

func init() {
	rootCmd.AddCommand(ordersCmd)

	ordersCmd.AddCommand(ordersListCmd)
	ordersListCmd.Flags().String("status", "", "Filter by status")
	ordersListCmd.Flags().String("email", "", "Filter by customer email (uses order search)")
	ordersListCmd.Flags().String("customer-id", "", "Filter by customer ID (uses order search)")
	ordersListCmd.Flags().String("from", "", "Filter by created date from (YYYY-MM-DD or RFC3339)")
	ordersListCmd.Flags().String("to", "", "Filter by created date to (YYYY-MM-DD or RFC3339)")
	ordersListCmd.Flags().Int("page", 1, "Page number")
	ordersListCmd.Flags().Int("page-size", 20, "Results per page")
	ordersListCmd.Flags().StringSlice("expand", nil, "Expand related resources: details, customer, products (adds API calls)")
	ordersListCmd.Flags().Int("jobs", 4, "Max concurrent API calls for --expand details")
	ordersListCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(ordersListCmd.Flags(), "light", "li")

	ordersCmd.AddCommand(ordersGetCmd)
	ordersGetCmd.Flags().String("by", "", "Find order by order number or query instead of ID")
	ordersGetCmd.Flags().StringSlice("expand", nil, "Expand related resources: customer, products (adds API calls)")
	ordersGetCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(ordersGetCmd.Flags(), "light", "li")
	ordersCmd.AddCommand(ordersCancelCmd)
	ordersCancelCmd.Flags().String("batch", "", "Batch input file (JSON array or NDJSON)")

	// Create / Update
	ordersCmd.AddCommand(ordersCreateCmd)
	addJSONBodyFlags(ordersCreateCmd)
	ordersCreateCmd.Flags().String("email", "", "Customer email")
	ordersCreateCmd.Flags().String("note", "", "Order note")
	ordersCreateCmd.Flags().String("tags", "", "Comma-separated tags")

	ordersCmd.AddCommand(ordersUpdateCmd)
	addJSONBodyFlags(ordersUpdateCmd)
	ordersUpdateCmd.Flags().String("note", "", "Order note")
	ordersUpdateCmd.Flags().String("tags", "", "Comma-separated tags")

	// Search
	ordersCmd.AddCommand(ordersSearchCmd)
	ordersSearchCmd.Flags().String("q", "", "Search query")
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

	// Admin API commands
	ordersCmd.AddCommand(ordersCommentCmd)
	ordersCommentCmd.Flags().String("text", "", "Comment text (required)")
	ordersCommentCmd.Flags().Bool("private", false, "Mark comment as private/internal")
	_ = ordersCommentCmd.MarkFlagRequired("text")
	ordersCommentCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	ordersCmd.AddCommand(ordersCommentsCmd)

	ordersCmd.AddCommand(ordersAdminRefundCmd)
	ordersAdminRefundCmd.Flags().String("performer-id", "", "ID of person performing refund (required)")
	ordersAdminRefundCmd.Flags().Int("amount", 0, "Refund amount in cents (required)")
	ordersAdminRefundCmd.Flags().String("payment-updated-at", "", "Order payment updated timestamp (required)")
	ordersAdminRefundCmd.Flags().String("remark", "", "Refund remark/reason")
	_ = ordersAdminRefundCmd.MarkFlagRequired("performer-id")
	_ = ordersAdminRefundCmd.MarkFlagRequired("amount")
	_ = ordersAdminRefundCmd.MarkFlagRequired("payment-updated-at")
	ordersAdminRefundCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	ordersCmd.AddCommand(ordersReceiptReissueCmd)
	ordersReceiptReissueCmd.Flags().Bool("dry-run", false, "Preview without making changes")

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
			"comment", "comments", "admin-refund", "receipt-reissue",
			"metafields", "app-metafields", "item-metafields", "item-app-metafields",
		},
		IDPrefix: "order",
	})
}
