package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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
				outfmt.FormatID("market", m.ID),
				m.Name,
				m.Handle,
				fmt.Sprintf("%t", m.Primary),
				fmt.Sprintf("%t", m.Enabled),
				fmt.Sprintf("%d", len(m.Countries)),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d markets\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Market ID:   %s\n", market.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:        %s\n", market.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:      %s\n", market.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Primary:     %t\n", market.Primary)
		_, _ = fmt.Fprintf(outWriter(cmd), "Enabled:     %t\n", market.Enabled)
		_, _ = fmt.Fprintf(outWriter(cmd), "Countries:   %v\n", market.Countries)
		_, _ = fmt.Fprintf(outWriter(cmd), "Currencies:  %v\n", market.Currencies)
		_, _ = fmt.Fprintf(outWriter(cmd), "Languages:   %v\n", market.Languages)
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
		if checkDryRun(cmd, "[DRY-RUN] Would create market") {
			return nil
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Created market %s\n", market.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:   %s\n", market.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle: %s\n", market.Handle)
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete market %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete market %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteMarket(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete market: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted market %s\n", args[0])
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
