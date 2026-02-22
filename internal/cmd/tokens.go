package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Manage API tokens",
}

var tokensListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.TokensListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListTokens(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list tokens: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "SCOPES", "EXPIRES", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			expires := "Never"
			if t.ExpiresAt != nil {
				expires = t.ExpiresAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				outfmt.FormatID("token", t.ID),
				t.Title,
				strings.Join(t.Scopes, ", "),
				expires,
				t.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d tokens\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var tokensGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get token details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		token, err := client.GetToken(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get token: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(token)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Token ID:  %s\n", token.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:     %s\n", token.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Scopes:    %s\n", strings.Join(token.Scopes, ", "))
		if token.ExpiresAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Expires:   %s\n", token.ExpiresAt.Format(time.RFC3339))
		} else {
			_, _ = fmt.Fprintf(outWriter(cmd), "Expires:   Never\n")
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:   %s\n", token.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:   %s\n", token.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var tokensCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an API token",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		scopesStr, _ := cmd.Flags().GetString("scopes")

		var scopes []string
		if scopesStr != "" {
			scopes = strings.Split(scopesStr, ",")
			for i := range scopes {
				scopes[i] = strings.TrimSpace(scopes[i])
			}
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create token '%s' with scopes: %s", title, strings.Join(scopes, ", "))) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.TokenCreateRequest{
			Title:  title,
			Scopes: scopes,
		}

		token, err := client.CreateToken(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create token: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(token)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created token %s\n", token.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:        %s\n", token.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Scopes:       %s\n", strings.Join(token.Scopes, ", "))
		if token.AccessToken != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nAccess Token: %s\n", token.AccessToken)
			_, _ = fmt.Fprintln(outWriter(cmd), "(Save this token - it will not be shown again)")
		}

		return nil
	},
}

var tokensDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete token %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete token %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteToken(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete token: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted token %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tokensCmd)

	tokensCmd.AddCommand(tokensListCmd)
	tokensListCmd.Flags().Int("page", 1, "Page number")
	tokensListCmd.Flags().Int("page-size", 20, "Results per page")

	tokensCmd.AddCommand(tokensGetCmd)

	tokensCmd.AddCommand(tokensCreateCmd)
	tokensCreateCmd.Flags().String("title", "", "Token title")
	tokensCreateCmd.Flags().String("scopes", "", "Comma-separated list of scopes")
	_ = tokensCreateCmd.MarkFlagRequired("title")

	tokensCmd.AddCommand(tokensDeleteCmd)
}
