package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var marketsCmd = &cobra.Command{
	Use:   "markets",
	Short: "Manage markets (regions)",
}

var marketsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List markets",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.MarketsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListMarkets(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list markets: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "HANDLE", "PRIMARY", "ENABLED", "COUNTRIES"}
		var rows [][]string
		for _, m := range resp.Items {
			rows = append(rows, []string{
				m.ID,
				m.Name,
				m.Handle,
				fmt.Sprintf("%t", m.Primary),
				fmt.Sprintf("%t", m.Enabled),
				fmt.Sprintf("%d", len(m.Countries)),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d markets\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var marketsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get market details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		market, err := client.GetMarket(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get market: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(market)
		}

		fmt.Printf("Market ID:   %s\n", market.ID)
		fmt.Printf("Name:        %s\n", market.Name)
		fmt.Printf("Handle:      %s\n", market.Handle)
		fmt.Printf("Primary:     %t\n", market.Primary)
		fmt.Printf("Enabled:     %t\n", market.Enabled)
		fmt.Printf("Countries:   %v\n", market.Countries)
		fmt.Printf("Currencies:  %v\n", market.Currencies)
		fmt.Printf("Languages:   %v\n", market.Languages)
		return nil
	},
}

var marketsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a market",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		handle, _ := cmd.Flags().GetString("handle")
		enabled, _ := cmd.Flags().GetBool("enabled")

		req := &api.MarketCreateRequest{
			Name:    name,
			Handle:  handle,
			Enabled: enabled,
		}

		market, err := client.CreateMarket(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create market: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(market)
		}

		fmt.Printf("Created market %s\n", market.ID)
		fmt.Printf("Name:   %s\n", market.Name)
		fmt.Printf("Handle: %s\n", market.Handle)
		return nil
	},
}

var marketsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a market",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete market %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteMarket(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete market: %w", err)
		}

		fmt.Printf("Deleted market %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(marketsCmd)

	marketsCmd.AddCommand(marketsListCmd)
	marketsListCmd.Flags().Int("page", 1, "Page number")
	marketsListCmd.Flags().Int("page-size", 20, "Results per page")

	marketsCmd.AddCommand(marketsGetCmd)

	marketsCmd.AddCommand(marketsCreateCmd)
	marketsCreateCmd.Flags().String("name", "", "Market name")
	marketsCreateCmd.Flags().String("handle", "", "Market handle (URL slug)")
	marketsCreateCmd.Flags().Bool("enabled", true, "Enable the market")
	_ = marketsCreateCmd.MarkFlagRequired("name")

	marketsCmd.AddCommand(marketsDeleteCmd)
}
