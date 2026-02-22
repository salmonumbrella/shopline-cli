package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var storefrontTokensCmd = &cobra.Command{
	Use:   "storefront-tokens",
	Short: "Manage storefront access tokens",
}

var storefrontTokensListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storefront access tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.StorefrontTokensListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListStorefrontTokens(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list storefront tokens: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("storefront_token", t.ID),
				t.Title,
				t.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d storefront tokens\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var storefrontTokensGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get storefront token details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		token, err := client.GetStorefrontToken(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get storefront token: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(token)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Token ID: %s\n", token.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:    %s\n", token.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:  %s\n", token.CreatedAt.Format(time.RFC3339))

		return nil
	},
}

var storefrontTokensCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a storefront access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create storefront token '%s'", title)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.StorefrontTokenCreateRequest{
			Title: title,
		}

		token, err := client.CreateStorefrontToken(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create storefront token: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(token)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created storefront token %s\n", token.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title: %s\n", token.Title)
		if token.AccessToken != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nAccess Token: %s\n", token.AccessToken)
			_, _ = fmt.Fprintln(outWriter(cmd), "(Save this token - it will not be shown again)")
		}

		return nil
	},
}

var storefrontTokensDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a storefront access token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete storefront token %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete storefront token %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteStorefrontToken(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete storefront token: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted storefront token %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(storefrontTokensCmd)

	storefrontTokensCmd.AddCommand(storefrontTokensListCmd)
	storefrontTokensListCmd.Flags().Int("page", 1, "Page number")
	storefrontTokensListCmd.Flags().Int("page-size", 20, "Results per page")

	storefrontTokensCmd.AddCommand(storefrontTokensGetCmd)

	storefrontTokensCmd.AddCommand(storefrontTokensCreateCmd)
	storefrontTokensCreateCmd.Flags().String("title", "", "Token title")
	_ = storefrontTokensCreateCmd.MarkFlagRequired("title")

	storefrontTokensCmd.AddCommand(storefrontTokensDeleteCmd)
}
