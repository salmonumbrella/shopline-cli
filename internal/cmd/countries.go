package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var countriesCmd = &cobra.Command{
	Use:   "countries",
	Short: "Manage countries",
}

var countriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List countries",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.ListCountries(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list countries: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"CODE", "NAME", "TAX", "TAX NAME", "PROVINCES"}
		var rows [][]string
		for _, c := range resp.Items {
			rows = append(rows, []string{
				c.Code,
				c.Name,
				fmt.Sprintf("%.2f%%", c.Tax),
				c.TaxName,
				fmt.Sprintf("%d", len(c.Provinces)),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d countries\n", len(resp.Items))
		return nil
	},
}

var countriesGetCmd = &cobra.Command{
	Use:   "get <code>",
	Short: "Get country details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		country, err := client.GetCountry(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get country: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(country)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Code:       %s\n", country.Code)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:       %s\n", country.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tax:        %.2f%%\n", country.Tax)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tax Name:   %s\n", country.TaxName)
		_, _ = fmt.Fprintf(outWriter(cmd), "Provinces:  %d\n", len(country.Provinces))

		if len(country.Provinces) > 0 {
			_, _ = fmt.Fprintln(outWriter(cmd), "\nProvinces/States:")
			for _, p := range country.Provinces {
				_, _ = fmt.Fprintf(outWriter(cmd), "  %s - %s (Tax: %.2f%% %s)\n", p.Code, p.Name, p.Tax, p.TaxName)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(countriesCmd)

	countriesCmd.AddCommand(countriesListCmd)

	countriesCmd.AddCommand(countriesGetCmd)
}
