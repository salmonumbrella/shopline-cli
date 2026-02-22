package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mediaCmd = &cobra.Command{
	Use:   "media",
	Short: "Manage media uploads (documented endpoints)",
}

var mediaCreateImageCmd = &cobra.Command{
	Use:     "create-image",
	Aliases: []string{"create", "upload"},
	Short:   "Create image (documented endpoint; raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would create media image") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.CreateMediaImage(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create media image: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(mediaCmd)

	mediaCmd.AddCommand(mediaCreateImageCmd)
	addJSONBodyFlags(mediaCreateImageCmd)
}
