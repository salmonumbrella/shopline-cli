package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var storefrontOAuthCmd = &cobra.Command{
	Use:   "storefront-oauth",
	Short: "Manage storefront OAuth clients",
}

var storefrontOAuthListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storefront OAuth clients",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.StorefrontOAuthClientsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListStorefrontOAuthClients(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list OAuth clients: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "CLIENT ID", "SCOPES", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("storefront_oauth_client", c.ID),
				c.Name,
				c.ClientID,
				strings.Join(c.Scopes, ", "),
				c.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d OAuth clients\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var storefrontOAuthGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get OAuth client details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		oauthClient, err := client.GetStorefrontOAuthClient(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get OAuth client: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(oauthClient)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "ID:            %s\n", oauthClient.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:          %s\n", oauthClient.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Client ID:     %s\n", oauthClient.ClientID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Redirect URIs: %s\n", strings.Join(oauthClient.RedirectURIs, ", "))
		_, _ = fmt.Fprintf(outWriter(cmd), "Scopes:        %s\n", strings.Join(oauthClient.Scopes, ", "))
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:       %s\n", oauthClient.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:       %s\n", oauthClient.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var storefrontOAuthCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a storefront OAuth client",
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

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create OAuth client '%s'", name)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.StorefrontOAuthClientCreateRequest{
			Name:         name,
			RedirectURIs: redirectURIs,
			Scopes:       scopes,
		}

		oauthClient, err := client.CreateStorefrontOAuthClient(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create OAuth client: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(oauthClient)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created OAuth client %s\n", oauthClient.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:          %s\n", oauthClient.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Client ID:     %s\n", oauthClient.ClientID)
		if oauthClient.ClientSecret != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nClient Secret: %s\n", oauthClient.ClientSecret)
			_, _ = fmt.Fprintln(outWriter(cmd), "(Save this secret - it will not be shown again)")
		}

		return nil
	},
}

var storefrontOAuthUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an OAuth client",
	Args:  cobra.ExactArgs(1),
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

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update OAuth client %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.StorefrontOAuthClientUpdateRequest{
			Name:         name,
			RedirectURIs: redirectURIs,
			Scopes:       scopes,
		}

		oauthClient, err := client.UpdateStorefrontOAuthClient(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update OAuth client: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(oauthClient)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated OAuth client %s\n", oauthClient.ID)
		return nil
	},
}

var storefrontOAuthDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an OAuth client",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete OAuth client %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete OAuth client %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteStorefrontOAuthClient(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete OAuth client: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted OAuth client %s\n", args[0])
		return nil
	},
}

var storefrontOAuthRotateCmd = &cobra.Command{
	Use:   "rotate-secret <id>",
	Short: "Rotate an OAuth client secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would rotate secret for OAuth client %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, "Rotate OAuth client secret? This will invalidate the existing secret. [y/N] ") {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		oauthClient, err := client.RotateStorefrontOAuthClientSecret(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to rotate OAuth client secret: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(oauthClient)
		}

		_, _ = fmt.Fprintln(outWriter(cmd), "OAuth client secret rotated")
		if oauthClient.ClientSecret != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nNew Client Secret: %s\n", oauthClient.ClientSecret)
			_, _ = fmt.Fprintln(outWriter(cmd), "(Save this secret - it will not be shown again)")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(storefrontOAuthCmd)

	storefrontOAuthCmd.AddCommand(storefrontOAuthListCmd)
	storefrontOAuthListCmd.Flags().Int("page", 1, "Page number")
	storefrontOAuthListCmd.Flags().Int("page-size", 20, "Results per page")

	storefrontOAuthCmd.AddCommand(storefrontOAuthGetCmd)

	storefrontOAuthCmd.AddCommand(storefrontOAuthCreateCmd)
	storefrontOAuthCreateCmd.Flags().String("name", "", "OAuth client name")
	storefrontOAuthCreateCmd.Flags().String("redirect-uris", "", "Comma-separated list of redirect URIs")
	storefrontOAuthCreateCmd.Flags().String("scopes", "", "Comma-separated list of scopes")
	_ = storefrontOAuthCreateCmd.MarkFlagRequired("name")
	_ = storefrontOAuthCreateCmd.MarkFlagRequired("redirect-uris")

	storefrontOAuthCmd.AddCommand(storefrontOAuthUpdateCmd)
	storefrontOAuthUpdateCmd.Flags().String("name", "", "OAuth client name")
	storefrontOAuthUpdateCmd.Flags().String("redirect-uris", "", "Comma-separated list of redirect URIs")
	storefrontOAuthUpdateCmd.Flags().String("scopes", "", "Comma-separated list of scopes")

	storefrontOAuthCmd.AddCommand(storefrontOAuthDeleteCmd)

	storefrontOAuthCmd.AddCommand(storefrontOAuthRotateCmd)
}
