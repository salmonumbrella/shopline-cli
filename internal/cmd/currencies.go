package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var currenciesCmd = &cobra.Command{
	Use:   "currencies",
	Short: "Manage currencies",
}

var currenciesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List currencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.ListCurrencies(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list currencies: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"CODE", "NAME", "SYMBOL", "PRIMARY", "ENABLED", "RATE"}
		var rows [][]string
		for _, c := range resp.Items {
			rows = append(rows, []string{
				c.Code,
				c.Name,
				c.Symbol,
				fmt.Sprintf("%t", c.Primary),
				fmt.Sprintf("%t", c.Enabled),
				fmt.Sprintf("%.4f", c.ExchangeRate),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d currencies\n", len(resp.Items))
		return nil
	},
}

var currenciesGetCmd = &cobra.Command{
	Use:   "get <code>",
	Short: "Get currency details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		currency, err := client.GetCurrency(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get currency: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(currency)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Code:          %s\n", currency.Code)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:          %s\n", currency.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Symbol:        %s\n", currency.Symbol)
		_, _ = fmt.Fprintf(outWriter(cmd), "Primary:       %t\n", currency.Primary)
		_, _ = fmt.Fprintf(outWriter(cmd), "Enabled:       %t\n", currency.Enabled)
		_, _ = fmt.Fprintf(outWriter(cmd), "Auto-Update:   %t\n", currency.AutoUpdate)
		_, _ = fmt.Fprintf(outWriter(cmd), "Exchange Rate: %.4f\n", currency.ExchangeRate)
		return nil
	},
}

var currenciesUpdateCmd = &cobra.Command{
	Use:   "update <code>",
	Short: "Update a currency",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update currency %s", args[0])) {
			return nil
		}

		req := &api.CurrencyUpdateRequest{}

		if cmd.Flags().Changed("enabled") {
			enabled, _ := cmd.Flags().GetBool("enabled")
			req.Enabled = &enabled
		}
		if cmd.Flags().Changed("exchange-rate") {
			rate, _ := cmd.Flags().GetFloat64("exchange-rate")
			req.ExchangeRate = &rate
		}
		if cmd.Flags().Changed("auto-update") {
			autoUpdate, _ := cmd.Flags().GetBool("auto-update")
			req.AutoUpdate = &autoUpdate
		}

		currency, err := client.UpdateCurrency(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update currency: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(currency)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated currency %s\n", currency.Code)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:          %s\n", currency.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Enabled:       %t\n", currency.Enabled)
		_, _ = fmt.Fprintf(outWriter(cmd), "Exchange Rate: %.4f\n", currency.ExchangeRate)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(currenciesCmd)

	currenciesCmd.AddCommand(currenciesListCmd)

	currenciesCmd.AddCommand(currenciesGetCmd)

	currenciesCmd.AddCommand(currenciesUpdateCmd)
	currenciesUpdateCmd.Flags().Bool("enabled", false, "Enable/disable the currency")
	currenciesUpdateCmd.Flags().Float64("exchange-rate", 0, "Exchange rate relative to primary currency")
	currenciesUpdateCmd.Flags().Bool("auto-update", false, "Automatically update exchange rate")
}
