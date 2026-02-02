package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var disputesCmd = &cobra.Command{
	Use:   "disputes",
	Short: "Manage payment disputes",
}

var disputesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List disputes",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")
		reason, _ := cmd.Flags().GetString("reason")

		opts := &api.DisputesListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			Reason:   reason,
		}

		resp, err := client.ListDisputes(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list disputes: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER", "AMOUNT", "CURRENCY", "STATUS", "REASON", "CREATED"}
		var rows [][]string
		for _, d := range resp.Items {
			rows = append(rows, []string{
				d.ID,
				d.OrderID,
				d.Amount,
				d.Currency,
				d.Status,
				d.Reason,
				d.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d disputes\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var disputesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get dispute details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		dispute, err := client.GetDispute(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get dispute: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(dispute)
		}

		fmt.Printf("Dispute ID:   %s\n", dispute.ID)
		fmt.Printf("Order ID:     %s\n", dispute.OrderID)
		fmt.Printf("Payment ID:   %s\n", dispute.PaymentID)
		fmt.Printf("Amount:       %s %s\n", dispute.Amount, dispute.Currency)
		fmt.Printf("Status:       %s\n", dispute.Status)
		fmt.Printf("Reason:       %s\n", dispute.Reason)
		if dispute.NetworkReasonCode != "" {
			fmt.Printf("Network Code: %s\n", dispute.NetworkReasonCode)
		}
		if dispute.EvidenceDueBy != nil {
			fmt.Printf("Evidence Due: %s\n", dispute.EvidenceDueBy.Format(time.RFC3339))
		}
		if dispute.Evidence != nil {
			fmt.Printf("Evidence:\n")
			if dispute.Evidence.CustomerName != "" {
				fmt.Printf("  Customer:   %s\n", dispute.Evidence.CustomerName)
			}
			if dispute.Evidence.CustomerEmail != "" {
				fmt.Printf("  Email:      %s\n", dispute.Evidence.CustomerEmail)
			}
			if dispute.Evidence.ShippingCarrier != "" {
				fmt.Printf("  Carrier:    %s\n", dispute.Evidence.ShippingCarrier)
			}
			if dispute.Evidence.ShippingTrackingNumber != "" {
				fmt.Printf("  Tracking:   %s\n", dispute.Evidence.ShippingTrackingNumber)
			}
		}
		if dispute.ResolvedAt != nil {
			fmt.Printf("Resolved:     %s\n", dispute.ResolvedAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:      %s\n", dispute.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:      %s\n", dispute.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var disputesSubmitCmd = &cobra.Command{
	Use:   "submit <id>",
	Short: "Submit dispute evidence for review",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Submit dispute %s for review? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		dispute, err := client.SubmitDispute(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to submit dispute: %w", err)
		}

		fmt.Printf("Submitted dispute %s (status: %s)\n", dispute.ID, dispute.Status)
		return nil
	},
}

var disputesAcceptCmd = &cobra.Command{
	Use:   "accept <id>",
	Short: "Accept a dispute (concede to customer)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Accept dispute %s? This will concede to the customer. [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		dispute, err := client.AcceptDispute(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to accept dispute: %w", err)
		}

		fmt.Printf("Accepted dispute %s (status: %s)\n", dispute.ID, dispute.Status)
		return nil
	},
}

var disputesEvidenceCmd = &cobra.Command{
	Use:   "evidence <id>",
	Short: "Update dispute evidence",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerName, _ := cmd.Flags().GetString("customer-name")
		customerEmail, _ := cmd.Flags().GetString("customer-email")
		productDesc, _ := cmd.Flags().GetString("product-description")
		shippingCarrier, _ := cmd.Flags().GetString("shipping-carrier")
		trackingNumber, _ := cmd.Flags().GetString("tracking-number")
		shippingDate, _ := cmd.Flags().GetString("shipping-date")

		req := &api.DisputeUpdateEvidenceRequest{
			CustomerName:           customerName,
			CustomerEmail:          customerEmail,
			ProductDescription:     productDesc,
			ShippingCarrier:        shippingCarrier,
			ShippingTrackingNumber: trackingNumber,
			ShippingDate:           shippingDate,
		}

		dispute, err := client.UpdateDisputeEvidence(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update dispute evidence: %w", err)
		}

		fmt.Printf("Updated evidence for dispute %s\n", dispute.ID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(disputesCmd)

	disputesCmd.AddCommand(disputesListCmd)
	disputesListCmd.Flags().Int("page", 1, "Page number")
	disputesListCmd.Flags().Int("page-size", 20, "Results per page")
	disputesListCmd.Flags().String("status", "", "Filter by status (needs_response, under_review, won, lost)")
	disputesListCmd.Flags().String("reason", "", "Filter by reason (fraudulent, product_not_received, etc.)")

	disputesCmd.AddCommand(disputesGetCmd)

	disputesCmd.AddCommand(disputesSubmitCmd)

	disputesCmd.AddCommand(disputesAcceptCmd)

	disputesCmd.AddCommand(disputesEvidenceCmd)
	disputesEvidenceCmd.Flags().String("customer-name", "", "Customer name")
	disputesEvidenceCmd.Flags().String("customer-email", "", "Customer email")
	disputesEvidenceCmd.Flags().String("product-description", "", "Product description")
	disputesEvidenceCmd.Flags().String("shipping-carrier", "", "Shipping carrier")
	disputesEvidenceCmd.Flags().String("tracking-number", "", "Shipping tracking number")
	disputesEvidenceCmd.Flags().String("shipping-date", "", "Shipping date")

	schema.Register(schema.Resource{
		Name:        "disputes",
		Description: "Manage payment disputes",
		Commands:    []string{"list", "get", "submit", "accept", "evidence"},
		IDPrefix:    "dispute",
	})
}
