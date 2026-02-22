package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var sellingPlansCmd = &cobra.Command{
	Use:   "selling-plans",
	Short: "Manage selling plan configurations",
}

var sellingPlansListCmd = &cobra.Command{
	Use:   "list",
	Short: "List selling plans",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.SellingPlansListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
		}

		resp, err := client.ListSellingPlans(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list selling plans: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "FREQUENCY", "DISCOUNT", "TRIAL DAYS", "STATUS", "CREATED"}
		var rows [][]string
		for _, p := range resp.Items {
			frequency := p.Frequency
			if p.FrequencyInterval > 1 {
				frequency = fmt.Sprintf("every %d %s", p.FrequencyInterval, p.Frequency)
			}
			discount := "-"
			if p.DiscountType != "" && p.DiscountValue != "" {
				if p.DiscountType == "percentage" {
					discount = p.DiscountValue + "%"
				} else {
					discount = p.DiscountValue
				}
			}
			trialDays := "-"
			if p.TrialDays > 0 {
				trialDays = fmt.Sprintf("%d days", p.TrialDays)
			}
			rows = append(rows, []string{
				outfmt.FormatID("selling_plan", p.ID),
				p.Name,
				frequency,
				discount,
				trialDays,
				p.Status,
				p.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d selling plans\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var sellingPlansGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get selling plan details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		plan, err := client.GetSellingPlan(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get selling plan: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(plan)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Selling Plan ID:     %s\n", plan.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:                %s\n", plan.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:         %s\n", plan.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Billing Policy:      %s\n", plan.BillingPolicy)
		_, _ = fmt.Fprintf(outWriter(cmd), "Delivery Policy:     %s\n", plan.DeliveryPolicy)
		_, _ = fmt.Fprintf(outWriter(cmd), "Frequency:           %s\n", plan.Frequency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Frequency Interval:  %d\n", plan.FrequencyInterval)
		_, _ = fmt.Fprintf(outWriter(cmd), "Trial Days:          %d\n", plan.TrialDays)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Type:       %s\n", plan.DiscountType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Value:      %s\n", plan.DiscountValue)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:              %s\n", plan.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Position:            %d\n", plan.Position)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:             %s\n", plan.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:             %s\n", plan.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var sellingPlansCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a selling plan",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		billingPolicy, _ := cmd.Flags().GetString("billing-policy")
		deliveryPolicy, _ := cmd.Flags().GetString("delivery-policy")
		frequency, _ := cmd.Flags().GetString("frequency")
		frequencyInterval, _ := cmd.Flags().GetInt("frequency-interval")
		trialDays, _ := cmd.Flags().GetInt("trial-days")
		discountType, _ := cmd.Flags().GetString("discount-type")
		discountValue, _ := cmd.Flags().GetString("discount-value")
		position, _ := cmd.Flags().GetInt("position")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create selling plan '%s' with %s frequency", name, frequency)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.SellingPlanCreateRequest{
			Name:              name,
			Description:       description,
			BillingPolicy:     billingPolicy,
			DeliveryPolicy:    deliveryPolicy,
			Frequency:         frequency,
			FrequencyInterval: frequencyInterval,
			TrialDays:         trialDays,
			DiscountType:      discountType,
			DiscountValue:     discountValue,
			Position:          position,
		}

		plan, err := client.CreateSellingPlan(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create selling plan: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(plan)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created selling plan %s\n", plan.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:       %s\n", plan.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Frequency:  %s\n", plan.Frequency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:     %s\n", plan.Status)

		return nil
	},
}

var sellingPlansDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a selling plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete selling plan %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteSellingPlan(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete selling plan: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted selling plan %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sellingPlansCmd)

	sellingPlansCmd.AddCommand(sellingPlansListCmd)
	sellingPlansListCmd.Flags().String("status", "", "Filter by status (active, inactive)")
	sellingPlansListCmd.Flags().Int("page", 1, "Page number")
	sellingPlansListCmd.Flags().Int("page-size", 20, "Results per page")

	sellingPlansCmd.AddCommand(sellingPlansGetCmd)

	sellingPlansCmd.AddCommand(sellingPlansCreateCmd)
	sellingPlansCreateCmd.Flags().String("name", "", "Selling plan name")
	sellingPlansCreateCmd.Flags().String("description", "", "Description")
	sellingPlansCreateCmd.Flags().String("billing-policy", "", "Billing policy (recurring, one-time)")
	sellingPlansCreateCmd.Flags().String("delivery-policy", "", "Delivery policy (recurring, one-time)")
	sellingPlansCreateCmd.Flags().String("frequency", "", "Frequency (daily, weekly, monthly, quarterly, yearly)")
	sellingPlansCreateCmd.Flags().Int("frequency-interval", 1, "Frequency interval")
	sellingPlansCreateCmd.Flags().Int("trial-days", 0, "Trial period in days")
	sellingPlansCreateCmd.Flags().String("discount-type", "", "Discount type (percentage, fixed)")
	sellingPlansCreateCmd.Flags().String("discount-value", "", "Discount value")
	sellingPlansCreateCmd.Flags().Int("position", 0, "Display position")
	_ = sellingPlansCreateCmd.MarkFlagRequired("name")
	_ = sellingPlansCreateCmd.MarkFlagRequired("frequency")

	sellingPlansCmd.AddCommand(sellingPlansDeleteCmd)
}
