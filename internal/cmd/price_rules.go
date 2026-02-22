package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var priceRulesCmd = &cobra.Command{
	Use:   "price-rules",
	Short: "Manage price rules",
}

var priceRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List price rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.PriceRulesListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListPriceRules(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list price rules: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "VALUE TYPE", "VALUE", "TARGET", "CUSTOMER", "USAGE LIMIT", "ONCE/CUSTOMER"}
		var rows [][]string
		for _, pr := range resp.Items {
			usageLimit := "-"
			if pr.UsageLimit > 0 {
				usageLimit = fmt.Sprintf("%d", pr.UsageLimit)
			}
			oncePerCustomer := "No"
			if pr.OncePerCustomer {
				oncePerCustomer = "Yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("price_rule", pr.ID),
				pr.Title,
				pr.ValueType,
				pr.Value,
				pr.TargetType,
				pr.CustomerSelection,
				usageLimit,
				oncePerCustomer,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d price rules\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var priceRulesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get price rule details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		priceRule, err := client.GetPriceRule(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get price rule: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(priceRule)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Price Rule ID:      %s\n", priceRule.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:              %s\n", priceRule.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Value Type:         %s\n", priceRule.ValueType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Value:              %s\n", priceRule.Value)
		_, _ = fmt.Fprintf(outWriter(cmd), "Target Type:        %s\n", priceRule.TargetType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Target Selection:   %s\n", priceRule.TargetSelection)
		_, _ = fmt.Fprintf(outWriter(cmd), "Allocation Method:  %s\n", priceRule.AllocationMethod)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer Selection: %s\n", priceRule.CustomerSelection)
		_, _ = fmt.Fprintf(outWriter(cmd), "Once Per Customer:  %v\n", priceRule.OncePerCustomer)
		if priceRule.UsageLimit > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Usage Limit:        %d\n", priceRule.UsageLimit)
		}
		if !priceRule.StartsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Starts At:          %s\n", priceRule.StartsAt.Format(time.RFC3339))
		}
		if !priceRule.EndsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Ends At:            %s\n", priceRule.EndsAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:            %s\n", priceRule.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var priceRulesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a price rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create price rule") {
			return nil
		}

		title, _ := cmd.Flags().GetString("title")
		valueType, _ := cmd.Flags().GetString("value-type")
		value, _ := cmd.Flags().GetString("value")

		req := &api.PriceRuleCreateRequest{
			Title:     title,
			ValueType: valueType,
			Value:     value,
		}

		priceRule, err := client.CreatePriceRule(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create price rule: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created price rule %s (%s)\n", priceRule.ID, priceRule.Title)
		return nil
	},
}

var priceRulesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a price rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update price rule %s", args[0])) {
			return nil
		}

		var req api.PriceRuleUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		priceRule, err := client.UpdatePriceRule(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update price rule: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(priceRule)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated price rule %s (%s)\n", priceRule.ID, priceRule.Title)
		return nil
	},
}

var priceRulesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a price rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete price rule %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeletePriceRule(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete price rule: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted price rule %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(priceRulesCmd)

	priceRulesCmd.AddCommand(priceRulesListCmd)
	priceRulesListCmd.Flags().Int("page", 1, "Page number")
	priceRulesListCmd.Flags().Int("page-size", 20, "Results per page")

	priceRulesCmd.AddCommand(priceRulesGetCmd)

	priceRulesCmd.AddCommand(priceRulesCreateCmd)
	priceRulesCreateCmd.Flags().String("title", "", "Price rule title")
	priceRulesCreateCmd.Flags().String("value-type", "", "Value type (percentage, fixed_amount)")
	priceRulesCreateCmd.Flags().String("value", "", "Value (e.g., -20 for 20% off)")
	_ = priceRulesCreateCmd.MarkFlagRequired("title")
	_ = priceRulesCreateCmd.MarkFlagRequired("value-type")
	_ = priceRulesCreateCmd.MarkFlagRequired("value")

	priceRulesCmd.AddCommand(priceRulesUpdateCmd)
	addJSONBodyFlags(priceRulesUpdateCmd)
	priceRulesUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	priceRulesCmd.AddCommand(priceRulesDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "price-rules",
		Description: "Manage price rules",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "price_rule",
	})
}
