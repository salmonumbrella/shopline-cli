package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var refundsCmd = &cobra.Command{
	Use:   "refunds",
	Short: "Manage order refunds",
}

type refundsAdminClient interface {
	AdminRefundOrder(ctx context.Context, orderID string, req *api.AdminRefundRequest) (json.RawMessage, error)
}

var refundsAdminClientFactory = func(cmd *cobra.Command) (refundsAdminClient, error) {
	return getAdminClient(cmd)
}

var refundsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List refunds",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.RefundsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListRefunds(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list refunds: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER", "STATUS", "AMOUNT", "CURRENCY", "NOTE", "ITEMS", "CREATED"}
		var rows [][]string
		for _, r := range resp.Items {
			note := r.Note
			if len(note) > 20 {
				note = note[:17] + "..."
			}
			rows = append(rows, []string{
				outfmt.FormatID("refund", r.ID),
				r.OrderID,
				r.Status,
				r.Amount,
				r.Currency,
				note,
				fmt.Sprintf("%d", len(r.LineItems)),
				r.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d refunds\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var refundsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get refund details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		refund, err := client.GetRefund(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get refund: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(refund)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Refund ID:      %s\n", refund.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Order ID:       %s\n", refund.OrderID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:         %s\n", refund.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Amount:         %s %s\n", refund.Amount, refund.Currency)
		if refund.Note != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Note:           %s\n", refund.Note)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Restock:        %t\n", refund.Restock)
		if !refund.ProcessedAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Processed At:   %s\n", refund.ProcessedAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", refund.CreatedAt.Format(time.RFC3339))

		if len(refund.LineItems) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nLine Items (%d):\n", len(refund.LineItems))
			for _, item := range refund.LineItems {
				_, _ = fmt.Fprintf(outWriter(cmd), "  - Line Item: %s, Qty: %d, Subtotal: %.2f\n",
					item.LineItemID, item.Quantity, item.Subtotal)
			}
		}
		return nil
	},
}

var refundsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a refund",
	Long:  "Create a refund using --body/--body-file JSON input and/or shorthand flags (--order-id, --amount, --note, --restock). Shorthand flags override fields from JSON input. If Open API refunds are unavailable (404), this command automatically falls back to Admin API refund flow.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would create refund") {
			return nil
		}

		hasBody := cmd.Flags().Changed("body") || cmd.Flags().Changed("body-file")
		hasFlags := cmd.Flags().Changed("order-id") || cmd.Flags().Changed("amount") ||
			cmd.Flags().Changed("note") || cmd.Flags().Changed("restock")
		if !hasBody && !hasFlags {
			return fmt.Errorf("provide refund data via --body/--body-file or individual flags (--order-id, --amount, --note, --restock)")
		}

		var req api.RefundCreateRequest
		if hasBody {
			if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
				return err
			}
		}

		if cmd.Flags().Changed("order-id") {
			req.OrderID, _ = cmd.Flags().GetString("order-id")
		}
		if cmd.Flags().Changed("amount") {
			req.Amount, _ = cmd.Flags().GetFloat64("amount")
		}
		if cmd.Flags().Changed("note") {
			req.Note, _ = cmd.Flags().GetString("note")
		}
		if cmd.Flags().Changed("restock") {
			req.Restock, _ = cmd.Flags().GetBool("restock")
		}

		if strings.TrimSpace(req.OrderID) == "" {
			return fmt.Errorf("order_id is required (set --order-id or include order_id in --body/--body-file)")
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		refund, err := client.CreateRefund(cmd.Context(), &req)
		if err != nil {
			if !isNotFoundAPIError(err) {
				return fmt.Errorf("failed to create refund: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Open API /refunds endpoint returned 404; attempting Admin API fallback.")

			adminResp, meta, fallbackErr := createRefundViaAdminFallback(cmd, client, &req)
			if fallbackErr != nil {
				return fmt.Errorf("failed to create refund: %v; admin fallback failed: %w", err, fallbackErr)
			}

			if outputFormat == "json" {
				return formatter.JSON(map[string]any{
					"id":                       "",
					"order_id":                 req.OrderID,
					"status":                   "submitted",
					"amount":                   req.Amount,
					"currency":                 meta.Currency,
					"note":                     req.Note,
					"restock":                  req.Restock,
					"via":                      "admin_fallback",
					"admin_amount_minor":       meta.AmountMinor,
					"performer_id":             meta.PerformerID,
					"order_payment_updated_at": meta.PaymentUpdatedAt,
					"admin_response":           adminResp,
				})
			}

			_, _ = fmt.Fprintf(
				outWriter(cmd),
				"Created refund via admin fallback for order %s (amount minor units: %d)\n",
				req.OrderID,
				meta.AmountMinor,
			)
			return nil
		}

		if outputFormat == "json" {
			return formatter.JSON(refund)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created refund %s (status: %s)\n", refund.ID, refund.Status)
		return nil
	},
}

var refundsOrderCmd = &cobra.Command{
	Use:   "order <order-id>",
	Short: "List refunds for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		resp, err := client.ListOrderRefunds(cmd.Context(), args[0])
		if err != nil {
			if !isNotFoundAPIError(err) {
				return fmt.Errorf("failed to list order refunds: %w", err)
			}

			logsRaw, logsErr := client.GetOrderActionLogs(cmd.Context(), args[0])
			if logsErr != nil {
				return fmt.Errorf("failed to list order refunds: %w; action-log fallback failed: %w", err, logsErr)
			}
			logs, parseErr := parseOrderActionLogs(logsRaw)
			if parseErr != nil {
				return fmt.Errorf("failed to list order refunds: %w; action-log fallback failed: %w", err, parseErr)
			}
			refundLogs := filterRefundActionLogs(logs)

			if outputFormat == "json" {
				return formatter.JSON(map[string]any{
					"items":                  []api.Refund{},
					"total_count":            0,
					"order_id":               args[0],
					"endpoint_unavailable":   true,
					"reason":                 "refunds_endpoint_404",
					"inferred_refund_events": refundLogs,
				})
			}

			out := outWriter(cmd)
			_, _ = fmt.Fprintln(out, "Refund endpoint unavailable (404) for this store; using action logs as fallback.")
			if len(refundLogs) == 0 {
				_, _ = fmt.Fprintf(out, "No refund-related action logs found for order %s.\n", args[0])
				return nil
			}

			headers := []string{"ACTION", "PERFORMER", "CREATED"}
			rows := make([][]string, 0, len(refundLogs))
			for _, log := range refundLogs {
				rows = append(rows, []string{
					firstNonEmpty(log.Key, log.Name),
					firstNonEmpty(
						log.PerformerName,
						log.PerformerID,
						log.UserID,
						stringFromMap(log.PerformData, "performer_name"),
						stringFromMap(log.PerformData, "performer_id"),
						stringFromMap(log.PerformData, "user_id"),
						stringFromMap(log.Data, "performer_name"),
						stringFromMap(log.Data, "performer_id"),
						stringFromMap(log.Data, "user_id"),
					),
					strings.TrimSpace(log.CreatedAt),
				})
			}
			formatter.Table(headers, rows)
			_, _ = fmt.Fprintf(out, "\nShowing %d refund-related action logs for order %s\n", len(refundLogs), args[0])
			return nil
		}

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "STATUS", "AMOUNT", "CURRENCY", "NOTE", "ITEMS", "CREATED"}
		var rows [][]string
		for _, r := range resp.Items {
			note := r.Note
			if len(note) > 20 {
				note = note[:17] + "..."
			}
			rows = append(rows, []string{
				outfmt.FormatID("refund", r.ID),
				r.Status,
				r.Amount,
				r.Currency,
				note,
				fmt.Sprintf("%d", len(r.LineItems)),
				r.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d refunds for order %s\n", len(resp.Items), args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(refundsCmd)

	refundsCmd.AddCommand(refundsListCmd)
	refundsListCmd.Flags().Int("page", 1, "Page number")
	refundsListCmd.Flags().Int("page-size", 20, "Results per page")

	refundsCmd.AddCommand(refundsGetCmd)
	refundsCmd.AddCommand(refundsCreateCmd)
	addJSONBodyFlags(refundsCreateCmd)
	refundsCreateCmd.Flags().String("order-id", "", "Order ID to refund")
	refundsCreateCmd.Flags().Float64("amount", 0, "Refund amount")
	refundsCreateCmd.Flags().String("note", "", "Optional refund note")
	refundsCreateCmd.Flags().Bool("restock", false, "Whether refunded items should be restocked")
	refundsCreateCmd.Flags().String("performer-id", "", "Admin fallback only: performer ID for refund")
	refundsCreateCmd.Flags().String("payment-updated-at", "", "Admin fallback only: order payment updated timestamp")
	refundsCmd.AddCommand(refundsOrderCmd)

	schema.Register(schema.Resource{
		Name:        "refunds",
		Description: "Manage order refunds",
		Commands:    []string{"list", "get", "create", "order"},
		IDPrefix:    "refund",
	})
}

type refundsAdminFallbackMeta struct {
	PerformerID      string
	PaymentUpdatedAt string
	AmountMinor      int
	Currency         string
}

func createRefundViaAdminFallback(cmd *cobra.Command, client api.APIClient, req *api.RefundCreateRequest) (json.RawMessage, *refundsAdminFallbackMeta, error) {
	adminClient, err := refundsAdminClientFactory(cmd)
	if err != nil {
		return nil, nil, err
	}

	order, err := client.GetOrder(cmd.Context(), req.OrderID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load order %s for admin fallback: %w", req.OrderID, err)
	}

	currency := strings.TrimSpace(order.Currency)
	amountMinor, err := refundAmountToMinorUnits(req.Amount, currency)
	if err != nil {
		return nil, nil, err
	}

	performerID, _ := cmd.Flags().GetString("performer-id")
	paymentUpdatedAt, _ := cmd.Flags().GetString("payment-updated-at")
	performerID = strings.TrimSpace(performerID)
	paymentUpdatedAt = strings.TrimSpace(paymentUpdatedAt)

	if performerID == "" || paymentUpdatedAt == "" {
		logsRaw, logsErr := client.GetOrderActionLogs(cmd.Context(), req.OrderID)
		if logsErr != nil {
			return nil, nil, fmt.Errorf("failed to auto-detect admin refund fields from action logs: %w", logsErr)
		}
		logs, parseErr := parseOrderActionLogs(logsRaw)
		if parseErr != nil {
			return nil, nil, parseErr
		}

		if performerID == "" {
			performerID = detectPerformerID(logs)
		}
		if paymentUpdatedAt == "" {
			paymentUpdatedAt = detectPaymentUpdatedAt(logs)
		}
	}

	if paymentUpdatedAt == "" && !order.UpdatedAt.IsZero() {
		paymentUpdatedAt = order.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}

	if performerID == "" {
		return nil, nil, fmt.Errorf(
			"missing performer ID for admin fallback; pass --performer-id or ensure order action logs include performer_id/user_id",
		)
	}
	if paymentUpdatedAt == "" {
		return nil, nil, fmt.Errorf(
			"missing payment updated timestamp for admin fallback; pass --payment-updated-at or ensure order action logs include updated_payment_status entries",
		)
	}

	adminReq := &api.AdminRefundRequest{
		PerformerID:           performerID,
		Amount:                amountMinor,
		OrderPaymentUpdatedAt: paymentUpdatedAt,
	}
	if note := strings.TrimSpace(req.Note); note != "" {
		adminReq.RefundRemark = note
	}

	result, err := adminClient.AdminRefundOrder(cmd.Context(), req.OrderID, adminReq)
	if err != nil {
		return nil, nil, err
	}

	meta := &refundsAdminFallbackMeta{
		PerformerID:      performerID,
		PaymentUpdatedAt: paymentUpdatedAt,
		AmountMinor:      amountMinor,
		Currency:         currency,
	}
	return result, meta, nil
}

func isNotFoundAPIError(err error) bool {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == 404 || strings.EqualFold(strings.TrimSpace(apiErr.Code), "HTTP_404")
	}
	return false
}

type orderActionLog struct {
	Key           string         `json:"key"`
	Name          string         `json:"name"`
	CreatedAt     string         `json:"created_at"`
	PerformerName string         `json:"performer_name"`
	PerformerID   string         `json:"performer_id"`
	UserID        string         `json:"user_id"`
	Data          map[string]any `json:"data"`
	PerformData   map[string]any `json:"perform_data"`
}

func parseOrderActionLogs(raw json.RawMessage) ([]orderActionLog, error) {
	var wrapped struct {
		Items   []orderActionLog `json:"items"`
		Results []orderActionLog `json:"results"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil {
		switch {
		case len(wrapped.Items) > 0:
			return wrapped.Items, nil
		case len(wrapped.Results) > 0:
			return wrapped.Results, nil
		}
	}

	var logs []orderActionLog
	if err := json.Unmarshal(raw, &logs); err != nil {
		return nil, fmt.Errorf("failed to parse order action logs for admin fallback: %w", err)
	}
	return logs, nil
}

func detectPerformerID(logs []orderActionLog) string {
	var (
		bestID   string
		bestTime time.Time
		hasBest  bool
		fallback string
	)

	for _, log := range logs {
		id := firstNonEmpty(
			log.PerformerID,
			log.UserID,
			stringFromMap(log.PerformData, "performer_id"),
			stringFromMap(log.PerformData, "user_id"),
			stringFromMap(log.Data, "performer_id"),
			stringFromMap(log.Data, "user_id"),
		)
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		fallback = id
		if t, ok := parseActionLogTime(log.CreatedAt); ok {
			if !hasBest || t.After(bestTime) {
				bestID = id
				bestTime = t
				hasBest = true
			}
		}
	}

	if hasBest {
		return bestID
	}
	return fallback
}

func detectPaymentUpdatedAt(logs []orderActionLog) string {
	var (
		bestValue string
		bestTime  time.Time
		hasBest   bool
		fallback  string
	)

	for _, log := range logs {
		if !isPaymentUpdatedLog(log) {
			continue
		}
		created := strings.TrimSpace(log.CreatedAt)
		if created == "" {
			continue
		}
		fallback = created
		if t, ok := parseActionLogTime(created); ok {
			if !hasBest || t.After(bestTime) {
				bestValue = created
				bestTime = t
				hasBest = true
			}
		}
	}

	if hasBest {
		return bestValue
	}
	return fallback
}

func parseActionLogTime(value string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func isPaymentUpdatedLog(log orderActionLog) bool {
	key := strings.ToLower(strings.TrimSpace(firstNonEmpty(log.Key, log.Name)))
	if key == "updated_payment_status" || key == "updated_order_payment_status" || strings.Contains(key, "payment_status") {
		return true
	}
	if _, ok := log.Data["updated_payment_status"]; ok {
		return true
	}
	_, ok := log.PerformData["updated_payment_status"]
	return ok
}

func filterRefundActionLogs(logs []orderActionLog) []orderActionLog {
	filtered := make([]orderActionLog, 0, len(logs))
	for _, log := range logs {
		key := strings.ToLower(strings.TrimSpace(firstNonEmpty(log.Key, log.Name)))
		if strings.Contains(key, "refund") || mapHasKeyContaining(log.Data, "refund") || mapHasKeyContaining(log.PerformData, "refund") {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

func mapHasKeyContaining(values map[string]any, token string) bool {
	if len(values) == 0 {
		return false
	}
	needle := strings.ToLower(strings.TrimSpace(token))
	if needle == "" {
		return false
	}
	for key := range values {
		if strings.Contains(strings.ToLower(strings.TrimSpace(key)), needle) {
			return true
		}
	}
	return false
}

func stringFromMap(m map[string]any, key string) string {
	if len(m) == 0 {
		return ""
	}
	raw, ok := m[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

var zeroDecimalCurrencies = map[string]struct{}{
	"BIF": {},
	"CLP": {},
	"DJF": {},
	"GNF": {},
	"JPY": {},
	"KMF": {},
	"KRW": {},
	"MGA": {},
	"PYG": {},
	"RWF": {},
	"TWD": {},
	"UGX": {},
	"VND": {},
	"VUV": {},
	"XAF": {},
	"XOF": {},
	"XPF": {},
}

var threeDecimalCurrencies = map[string]struct{}{
	"BHD": {},
	"IQD": {},
	"JOD": {},
	"KWD": {},
	"LYD": {},
	"OMR": {},
	"TND": {},
}

func refundAmountToMinorUnits(amount float64, currency string) (int, error) {
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be a positive value")
	}

	cur := strings.ToUpper(strings.TrimSpace(currency))
	decimals := 2
	if _, ok := zeroDecimalCurrencies[cur]; ok {
		decimals = 0
	} else if _, ok := threeDecimalCurrencies[cur]; ok {
		decimals = 3
	}

	factor := math.Pow10(decimals)
	minor := int(math.Round(amount * factor))
	if minor <= 0 {
		return 0, fmt.Errorf("amount is too small after conversion to minor units")
	}
	return minor, nil
}
