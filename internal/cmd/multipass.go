package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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
		fmt.Printf("Status:  %s\n", status)
		fmt.Printf("Created: %s\n", multipass.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated: %s\n", multipass.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var multipassEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable multipass authentication",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Println("[DRY-RUN] Would enable multipass authentication")
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

		fmt.Println("Multipass authentication enabled")
		if multipass.Secret != "" {
			fmt.Printf("\nSecret: %s\n", multipass.Secret)
			fmt.Println("(Save this secret - it will not be shown again)")
		}

		return nil
	},
}

var multipassDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable multipass authentication",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Println("[DRY-RUN] Would disable multipass authentication")
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Print("Disable multipass authentication? [y/N] ")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DisableMultipass(cmd.Context()); err != nil {
			return fmt.Errorf("failed to disable multipass: %w", err)
		}

		fmt.Println("Multipass authentication disabled")
		return nil
	},
}

var multipassRotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate multipass secret",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Println("[DRY-RUN] Would rotate multipass secret")
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Print("Rotate multipass secret? This will invalidate all existing tokens. [y/N] ")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
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

		fmt.Println("Multipass secret rotated")
		if multipass.Secret != "" {
			fmt.Printf("\nNew Secret: %s\n", multipass.Secret)
			fmt.Println("(Save this secret - it will not be shown again)")
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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would generate multipass token for %s\n", email)
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

		fmt.Printf("Token:   %s\n", token.Token)
		fmt.Printf("URL:     %s\n", token.URL)
		fmt.Printf("Expires: %s\n", token.ExpiresAt.Format(time.RFC3339))

		return nil
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
}
