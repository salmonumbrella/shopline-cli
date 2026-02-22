package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var storefrontOAuthApplicationsCmd = &cobra.Command{
	Use:     "storefront-oauth-applications",
	Aliases: []string{"storefront-oauth-apps"},
	Short:   "Manage storefront OAuth applications (documented endpoints)",
}

var storefrontOAuthApplicationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storefront OAuth applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.ListStorefrontOAuthApplications(cmd.Context(), &api.StorefrontOAuthApplicationsListOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to list OAuth applications: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "CLIENT ID", "SCOPES", "CREATED"}
		var rows [][]string
		for _, a := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("storefront_oauth_app", a.ID),
				a.Name,
				a.ClientID,
				strings.Join(a.Scopes, ", "),
				a.CreatedAt.Format("2006-01-02"),
			})
		}
		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d OAuth applications\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var storefrontOAuthApplicationsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get OAuth application details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		app, err := client.GetStorefrontOAuthApplication(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get OAuth application: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(app)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "ID:            %s\n", app.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:          %s\n", app.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Client ID:     %s\n", app.ClientID)
		if len(app.RedirectURIs) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Redirect URIs: %s\n", strings.Join(app.RedirectURIs, ", "))
		}
		if len(app.Scopes) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Scopes:        %s\n", strings.Join(app.Scopes, ", "))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:       %s\n", app.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:       %s\n", app.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var storefrontOAuthApplicationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a storefront OAuth application",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		redirectURIsStr, _ := cmd.Flags().GetString("redirect-uris")
		scopesStr, _ := cmd.Flags().GetString("scopes")

		var redirectURIs []string
		if redirectURIsStr != "" {
			redirectURIs = strings.Split(redirectURIsStr, ",")
			for i := range redirectURIs {
				redirectURIs[i] = strings.TrimSpace(redirectURIs[i])
			}
		}
		var scopes []string
		if scopesStr != "" {
			scopes = strings.Split(scopesStr, ",")
			for i := range scopes {
				scopes[i] = strings.TrimSpace(scopes[i])
			}
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create OAuth application %q", name)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		app, err := client.CreateStorefrontOAuthApplication(cmd.Context(), &api.StorefrontOAuthApplicationCreateRequest{
			Name:         name,
			RedirectURIs: redirectURIs,
			Scopes:       scopes,
		})
		if err != nil {
			return fmt.Errorf("failed to create OAuth application: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(app)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created OAuth application %s\n", app.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:      %s\n", app.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Client ID: %s\n", app.ClientID)
		return nil
	},
}

var storefrontOAuthApplicationsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a storefront OAuth application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete OAuth application %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if err := client.DeleteStorefrontOAuthApplication(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete OAuth application: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted OAuth application %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(storefrontOAuthApplicationsCmd)
	storefrontOAuthApplicationsCmd.AddCommand(storefrontOAuthApplicationsListCmd)
	storefrontOAuthApplicationsListCmd.Flags().Int("page", 1, "Page number")
	storefrontOAuthApplicationsListCmd.Flags().Int("page-size", 20, "Results per page")

	storefrontOAuthApplicationsCmd.AddCommand(storefrontOAuthApplicationsGetCmd)

	storefrontOAuthApplicationsCmd.AddCommand(storefrontOAuthApplicationsCreateCmd)
	storefrontOAuthApplicationsCreateCmd.Flags().String("name", "", "Application name")
	storefrontOAuthApplicationsCreateCmd.Flags().String("redirect-uris", "", "Comma-separated redirect URIs")
	storefrontOAuthApplicationsCreateCmd.Flags().String("scopes", "", "Comma-separated scopes")
	storefrontOAuthApplicationsCreateCmd.Flags().Bool("dry-run", false, "Preview without making changes")
	_ = storefrontOAuthApplicationsCreateCmd.MarkFlagRequired("name")

	storefrontOAuthApplicationsCmd.AddCommand(storefrontOAuthApplicationsDeleteCmd)
	storefrontOAuthApplicationsDeleteCmd.Flags().Bool("dry-run", false, "Preview without making changes")
}
