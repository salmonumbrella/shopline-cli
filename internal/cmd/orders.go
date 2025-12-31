package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/batch"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.OrdersListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
		}

		resp, err := client.ListOrders(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list orders: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NUMBER", "STATUS", "TOTAL", "CUSTOMER", "CREATED"}
		var rows [][]string
		for _, o := range resp.Items {
			rows = append(rows, []string{
				o.ID,
				o.OrderNumber,
				o.Status,
				o.TotalPrice + " " + o.Currency,
				o.CustomerEmail,
				o.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d orders\n", len(resp.Items), resp.TotalCount)
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
			return err
		}

		order, err := client.GetOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(order)
		}

		fmt.Printf("Order ID:       %s\n", order.ID)
		fmt.Printf("Order Number:   %s\n", order.OrderNumber)
		fmt.Printf("Status:         %s\n", order.Status)
		fmt.Printf("Payment:        %s\n", order.PaymentStatus)
		fmt.Printf("Fulfillment:    %s\n", order.FulfillStatus)
		fmt.Printf("Total:          %s %s\n", order.TotalPrice, order.Currency)
		fmt.Printf("Customer:       %s <%s>\n", order.CustomerName, order.CustomerEmail)
		fmt.Printf("Created:        %s\n", order.CreatedAt.Format(time.RFC3339))
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
			fmt.Printf("[DRY-RUN] Would cancel order %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Cancel order %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.CancelOrder(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to cancel order: %w", err)
		}

		fmt.Printf("Order %s cancelled.\n", args[0])
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

	format := outfmt.FormatText
	if outputFormat == "json" {
		format = outfmt.FormatJSON
	}

	f := outfmt.New(formatterWriter, format, colorMode)
	if query != "" {
		f = f.WithQuery(query)
	}
	return f
}

func init() {
	rootCmd.AddCommand(ordersCmd)

	ordersCmd.AddCommand(ordersListCmd)
	ordersListCmd.Flags().String("status", "", "Filter by status")
	ordersListCmd.Flags().Int("page", 1, "Page number")
	ordersListCmd.Flags().Int("page-size", 20, "Results per page")

	ordersCmd.AddCommand(ordersGetCmd)
	ordersCmd.AddCommand(ordersCancelCmd)
	ordersCancelCmd.Flags().String("batch", "", "Batch input file (JSON array or NDJSON)")
}
