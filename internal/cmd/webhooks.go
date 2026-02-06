package cmd

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
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
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d webhooks\n", len(resp.Items), resp.TotalCount)
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

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Webhook ID:   %s\n", webhook.ID)
		_, _ = fmt.Fprintf(out, "Topic:        %s\n", webhook.Topic)
		_, _ = fmt.Fprintf(out, "Address:      %s\n", webhook.Address)
		_, _ = fmt.Fprintf(out, "Format:       %s\n", webhook.Format)
		_, _ = fmt.Fprintf(out, "API Version:  %s\n", webhook.APIVersion)
		_, _ = fmt.Fprintf(out, "Created:      %s\n", webhook.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(out, "Updated:      %s\n", webhook.UpdatedAt.Format(time.RFC3339))

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
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would create webhook for topic %s at %s\n", topic, address)
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

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Created webhook %s\n", webhook.ID)
		_, _ = fmt.Fprintf(out, "Topic:    %s\n", webhook.Topic)
		_, _ = fmt.Fprintf(out, "Address:  %s\n", webhook.Address)
		_, _ = fmt.Fprintf(out, "Format:   %s\n", webhook.Format)

		return nil
	},
}

var webhooksUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		topic, _ := cmd.Flags().GetString("topic")
		address, _ := cmd.Flags().GetString("address")
		format, _ := cmd.Flags().GetString("format")
		apiVersion, _ := cmd.Flags().GetString("api-version")

		if !cmd.Flags().Changed("topic") &&
			!cmd.Flags().Changed("address") &&
			!cmd.Flags().Changed("format") &&
			!cmd.Flags().Changed("api-version") {
			return fmt.Errorf("at least one field must be provided to update (topic/address/format/api-version)")
		}

		if cmd.Flags().Changed("address") {
			if err := validateWebhookAddress(address); err != nil {
				return err
			}
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would update webhook %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.WebhookUpdateRequest{}
		if cmd.Flags().Changed("topic") {
			req.Topic = topic
		}
		if cmd.Flags().Changed("address") {
			req.Address = address
		}
		if cmd.Flags().Changed("format") && strings.TrimSpace(format) != "" {
			req.Format = api.WebhookFormat(format)
		}
		if cmd.Flags().Changed("api-version") {
			req.APIVersion = apiVersion
		}

		webhook, err := client.UpdateWebhook(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update webhook: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(webhook)
		}

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Updated webhook %s\n", webhook.ID)
		_, _ = fmt.Fprintf(out, "Topic:    %s\n", webhook.Topic)
		_, _ = fmt.Fprintf(out, "Address:  %s\n", webhook.Address)
		_, _ = fmt.Fprintf(out, "Format:   %s\n", webhook.Format)

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
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would delete webhook %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteWebhook(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete webhook: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted webhook %s\n", args[0])
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

	webhooksCmd.AddCommand(webhooksUpdateCmd)
	webhooksUpdateCmd.Flags().String("topic", "", "Webhook topic (e.g., orders/create)")
	webhooksUpdateCmd.Flags().String("address", "", "Webhook URL")
	webhooksUpdateCmd.Flags().String("format", "", "Payload format (json/xml)")
	webhooksUpdateCmd.Flags().String("api-version", "", "API version for webhook payloads")
	webhooksUpdateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	webhooksCmd.AddCommand(webhooksDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "webhooks",
		Description: "Manage webhook subscriptions",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "webhook",
	})
}
