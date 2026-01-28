package cmd

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage webhooks",
}

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		topic, _ := cmd.Flags().GetString("topic")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.WebhooksListOptions{
			Page:     page,
			PageSize: pageSize,
			Topic:    topic,
		}

		resp, err := client.ListWebhooks(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list webhooks: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TOPIC", "ADDRESS", "FORMAT", "API VERSION", "CREATED"}
		var rows [][]string
		for _, w := range resp.Items {
			rows = append(rows, []string{
				w.ID,
				w.Topic,
				w.Address,
				string(w.Format),
				w.APIVersion,
				w.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d webhooks\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var webhooksGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get webhook details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		webhook, err := client.GetWebhook(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get webhook: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(webhook)
		}

		fmt.Printf("Webhook ID:   %s\n", webhook.ID)
		fmt.Printf("Topic:        %s\n", webhook.Topic)
		fmt.Printf("Address:      %s\n", webhook.Address)
		fmt.Printf("Format:       %s\n", webhook.Format)
		fmt.Printf("API Version:  %s\n", webhook.APIVersion)
		fmt.Printf("Created:      %s\n", webhook.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:      %s\n", webhook.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook",
	Long:  "Create a webhook subscription. Note: The API will reject the request with an error if a webhook with the same topic and address already exists.",
	RunE: func(cmd *cobra.Command, args []string) error {
		topic, _ := cmd.Flags().GetString("topic")
		address, _ := cmd.Flags().GetString("address")
		format, _ := cmd.Flags().GetString("format")
		apiVersion, _ := cmd.Flags().GetString("api-version")

		// Validate webhook address is a valid HTTPS URL
		if err := validateWebhookAddress(address); err != nil {
			return err
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create webhook for topic %s at %s\n", topic, address)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.WebhookCreateRequest{
			Address:    address,
			Topic:      topic,
			APIVersion: apiVersion,
		}
		if format != "" {
			req.Format = api.WebhookFormat(format)
		}

		webhook, err := client.CreateWebhook(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create webhook: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(webhook)
		}

		fmt.Printf("Created webhook %s\n", webhook.ID)
		fmt.Printf("Topic:    %s\n", webhook.Topic)
		fmt.Printf("Address:  %s\n", webhook.Address)
		fmt.Printf("Format:   %s\n", webhook.Format)

		return nil
	},
}

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete webhook %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteWebhook(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete webhook: %w", err)
		}

		fmt.Printf("Deleted webhook %s\n", args[0])
		return nil
	},
}

// validateWebhookAddress validates that the address is a valid HTTPS URL.
func validateWebhookAddress(address string) error {
	parsedURL, err := url.Parse(address)
	if err != nil || !strings.EqualFold(parsedURL.Scheme, "https") {
		return fmt.Errorf("webhook address must be a valid HTTPS URL")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(webhooksCmd)

	webhooksCmd.AddCommand(webhooksListCmd)
	webhooksListCmd.Flags().String("topic", "", "Filter by topic (e.g., orders/create, products/update)")
	webhooksListCmd.Flags().Int("page", 1, "Page number")
	webhooksListCmd.Flags().Int("page-size", 20, "Results per page")

	webhooksCmd.AddCommand(webhooksGetCmd)

	webhooksCmd.AddCommand(webhooksCreateCmd)
	webhooksCreateCmd.Flags().String("topic", "", "Webhook topic (e.g., orders/create)")
	webhooksCreateCmd.Flags().String("address", "", "Webhook URL")
	webhooksCreateCmd.Flags().String("format", "json", "Payload format (json/xml)")
	webhooksCreateCmd.Flags().String("api-version", "", "API version for webhook payloads")
	_ = webhooksCreateCmd.MarkFlagRequired("topic")
	_ = webhooksCreateCmd.MarkFlagRequired("address")

	webhooksCmd.AddCommand(webhooksDeleteCmd)
}
