package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var domainsCmd = &cobra.Command{
	Use:     "domains",
	Aliases: []string{"domain"},
	Short:   "Manage domains",
}

var domainsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.DomainsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   api.DomainStatus(status),
		}

		if cmd.Flags().Changed("primary") {
			primary, _ := cmd.Flags().GetBool("primary")
			opts.Primary = &primary
		}

		resp, err := client.ListDomains(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list domains: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "HOST", "PRIMARY", "SSL", "STATUS", "VERIFIED"}
		var rows [][]string
		for _, d := range resp.Items {
			primary := "no"
			if d.Primary {
				primary = "yes"
			}
			ssl := "no"
			if d.SSL {
				ssl = "yes"
			}
			verified := "no"
			if d.Verified {
				verified = "yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("domain", d.ID),
				d.Host,
				primary,
				ssl,
				string(d.Status),
				verified,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d domains\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var domainsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get domain details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		domain, err := client.GetDomain(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get domain: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(domain)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Domain ID:          %s\n", domain.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Host:               %s\n", domain.Host)
		_, _ = fmt.Fprintf(outWriter(cmd), "Primary:            %t\n", domain.Primary)
		_, _ = fmt.Fprintf(outWriter(cmd), "SSL:                %t\n", domain.SSL)
		if domain.SSLStatus != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "SSL Status:         %s\n", domain.SSLStatus)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:             %s\n", domain.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Verified:           %t\n", domain.Verified)
		if domain.VerifiedAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Verified At:        %s\n", domain.VerifiedAt.Format(time.RFC3339))
		}
		if domain.ExpiresAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Expires At:         %s\n", domain.ExpiresAt.Format(time.RFC3339))
		}
		if !domain.Verified && domain.VerificationDNS != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nVerification DNS:   %s\n", domain.VerificationDNS)
			_, _ = fmt.Fprintf(outWriter(cmd), "Verification Token: %s\n", domain.VerificationToken)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "\nCreated:            %s\n", domain.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:            %s\n", domain.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var domainsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		primary, _ := cmd.Flags().GetBool("primary")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create domain: %s", host)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.DomainCreateRequest{
			Host:    host,
			Primary: primary,
		}

		domain, err := client.CreateDomain(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create domain: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(domain)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created domain %s\n", domain.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Host:   %s\n", domain.Host)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status: %s\n", domain.Status)
		if domain.VerificationDNS != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nTo verify ownership, add this DNS record:\n")
			_, _ = fmt.Fprintf(outWriter(cmd), "  %s\n", domain.VerificationDNS)
			_, _ = fmt.Fprintf(outWriter(cmd), "  Token: %s\n", domain.VerificationToken)
		}

		return nil
	},
}

var domainsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &api.DomainUpdateRequest{}

		if cmd.Flags().Changed("primary") {
			v, _ := cmd.Flags().GetBool("primary")
			req.Primary = &v
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update domain %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		domain, err := client.UpdateDomain(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update domain: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(domain)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated domain %s\n", domain.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Host:    %s\n", domain.Host)
		_, _ = fmt.Fprintf(outWriter(cmd), "Primary: %t\n", domain.Primary)

		return nil
	},
}

var domainsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete domain %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete domain %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteDomain(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete domain: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted domain %s\n", args[0])
		return nil
	},
}

var domainsVerifyCmd = &cobra.Command{
	Use:   "verify <id>",
	Short: "Verify domain ownership",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would verify domain %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		domain, err := client.VerifyDomain(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to verify domain: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(domain)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Verification initiated for domain %s\n", domain.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Host:     %s\n", domain.Host)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:   %s\n", domain.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Verified: %t\n", domain.Verified)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(domainsCmd)

	domainsCmd.AddCommand(domainsListCmd)
	domainsListCmd.Flags().Int("page", 1, "Page number")
	domainsListCmd.Flags().Int("page-size", 20, "Results per page")
	domainsListCmd.Flags().String("status", "", "Filter by status (active, pending, verifying, failed, expired)")
	domainsListCmd.Flags().Bool("primary", false, "Filter by primary status")

	domainsCmd.AddCommand(domainsGetCmd)

	domainsCmd.AddCommand(domainsCreateCmd)
	domainsCreateCmd.Flags().String("host", "", "Domain host (required)")
	domainsCreateCmd.Flags().Bool("primary", false, "Set as primary domain")
	_ = domainsCreateCmd.MarkFlagRequired("host")

	domainsCmd.AddCommand(domainsUpdateCmd)
	domainsUpdateCmd.Flags().Bool("primary", false, "Set as primary domain")

	domainsCmd.AddCommand(domainsDeleteCmd)
	domainsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	domainsCmd.AddCommand(domainsVerifyCmd)
}
