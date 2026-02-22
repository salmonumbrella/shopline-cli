package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var multipassCmd = &cobra.Command{
	Use:   "multipass",
	Short: "Manage multipass authentication",
}

var multipassStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get multipass configuration status",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		multipass, err := client.GetMultipass(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get multipass status: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(multipass)
		}

		status := "Disabled"
		if multipass.Enabled {
			status = "Enabled"
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:  %s\n", status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created: %s\n", multipass.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated: %s\n", multipass.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var multipassEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable multipass authentication",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would enable multipass authentication") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		multipass, err := client.EnableMultipass(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to enable multipass: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(multipass)
		}

		_, _ = fmt.Fprintln(outWriter(cmd), "Multipass authentication enabled")
		if multipass.Secret != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nSecret: %s\n", multipass.Secret)
			_, _ = fmt.Fprintln(outWriter(cmd), "(Save this secret - it will not be shown again)")
		}

		return nil
	},
}

var multipassDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable multipass authentication",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would disable multipass authentication") {
			return nil
		}

		if !confirmAction(cmd, "Disable multipass authentication? [y/N] ") {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DisableMultipass(cmd.Context()); err != nil {
			return fmt.Errorf("failed to disable multipass: %w", err)
		}

		_, _ = fmt.Fprintln(outWriter(cmd), "Multipass authentication disabled")
		return nil
	},
}

var multipassRotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate multipass secret",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would rotate multipass secret") {
			return nil
		}

		if !confirmAction(cmd, "Rotate multipass secret? This will invalidate all existing tokens. [y/N] ") {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		multipass, err := client.RotateMultipassSecret(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to rotate multipass secret: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(multipass)
		}

		_, _ = fmt.Fprintln(outWriter(cmd), "Multipass secret rotated")
		if multipass.Secret != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nNew Secret: %s\n", multipass.Secret)
			_, _ = fmt.Fprintln(outWriter(cmd), "(Save this secret - it will not be shown again)")
		}

		return nil
	},
}

var multipassTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Generate a multipass login token",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		returnTo, _ := cmd.Flags().GetString("return-to")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would generate multipass token for %s", email)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.MultipassTokenRequest{
			Email:    email,
			ReturnTo: returnTo,
		}

		token, err := client.GenerateMultipassToken(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to generate multipass token: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(token)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Token:   %s\n", token.Token)
		_, _ = fmt.Fprintf(outWriter(cmd), "URL:     %s\n", token.URL)
		_, _ = fmt.Fprintf(outWriter(cmd), "Expires: %s\n", token.ExpiresAt.Format(time.RFC3339))

		return nil
	},
}

// Documented multipass endpoints

var multipassSecretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage multipass secret (documented endpoints)",
}

var multipassSecretGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get multipass secret",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetMultipassSecret(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get multipass secret: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var multipassSecretCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create multipass secret (may return existing secret)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Optional body (docs don't require it).
		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")

		var req json.RawMessage
		var hasBody bool
		var err error
		if strings.TrimSpace(body) != "" || strings.TrimSpace(bodyFile) != "" {
			req, err = readJSONBodyFlags(cmd)
			if err != nil {
				return err
			}
			hasBody = true
		}

		if checkDryRun(cmd, "[DRY-RUN] Would create multipass secret") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		var anyBody any
		if hasBody {
			anyBody = req
		}

		resp, err := client.CreateMultipassSecret(cmd.Context(), anyBody)
		if err != nil {
			return fmt.Errorf("failed to create multipass secret: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var multipassLinkingsCmd = &cobra.Command{
	Use:   "linkings",
	Short: "Manage multipass linking records (documented endpoints)",
}

var multipassLinkingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active multipass linkings (raw JSON)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerIDs, _ := cmd.Flags().GetStringSlice("customer-id")
		resp, err := client.ListMultipassLinkings(cmd.Context(), customerIDs)
		if err != nil {
			return fmt.Errorf("failed to list multipass linkings: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var multipassCustomersCmd = &cobra.Command{
	Use:   "customers",
	Short: "Manage multipass customer linkings (documented endpoints)",
}

var multipassCustomersLinkCmd = &cobra.Command{
	Use:   "link <customer-id>",
	Short: "Update customer's multipass linking (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would update multipass linking") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.UpdateMultipassCustomerLinking(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update multipass linking: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var multipassCustomersUnlinkCmd = &cobra.Command{
	Use:     "unlink <customer-id>",
	Aliases: []string{"delete", "del", "rm"},
	Short:   "Delete customer's multipass linking (marks inactive)",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !confirmAction(cmd, fmt.Sprintf("Delete multipass linking for customer %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if checkDryRun(cmd, "[DRY-RUN] Would delete multipass linking") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.DeleteMultipassCustomerLinking(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to delete multipass linking: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(multipassCmd)

	multipassCmd.AddCommand(multipassStatusCmd)

	multipassCmd.AddCommand(multipassEnableCmd)

	multipassCmd.AddCommand(multipassDisableCmd)

	multipassCmd.AddCommand(multipassRotateCmd)

	multipassCmd.AddCommand(multipassTokenCmd)
	multipassTokenCmd.Flags().String("email", "", "Customer email address")
	multipassTokenCmd.Flags().String("return-to", "", "URL to redirect after login")
	_ = multipassTokenCmd.MarkFlagRequired("email")

	// Documented endpoints
	multipassCmd.AddCommand(multipassSecretCmd)
	multipassSecretCmd.AddCommand(multipassSecretGetCmd)
	multipassSecretCmd.AddCommand(multipassSecretCreateCmd)
	addJSONBodyFlags(multipassSecretCreateCmd)

	multipassCmd.AddCommand(multipassLinkingsCmd)
	multipassLinkingsCmd.AddCommand(multipassLinkingsListCmd)
	multipassLinkingsListCmd.Flags().StringSlice("customer-id", nil, "Customer id filter (repeatable)")

	multipassCmd.AddCommand(multipassCustomersCmd)
	multipassCustomersCmd.AddCommand(multipassCustomersLinkCmd)
	addJSONBodyFlags(multipassCustomersLinkCmd)
	multipassCustomersCmd.AddCommand(multipassCustomersUnlinkCmd)

	schema.Register(schema.Resource{
		Name:        "multipass",
		Description: "Manage multipass authentication",
		Commands:    []string{"status", "enable", "disable", "rotate", "token", "secret", "linkings", "customers"},
	})
}
