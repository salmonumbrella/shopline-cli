package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Token utilities",
}

var tokenInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get current access token info",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetTokenInfo(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get token info: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
	tokenCmd.AddCommand(tokenInfoCmd)
}
