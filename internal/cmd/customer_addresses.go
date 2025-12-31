package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var customerAddressesCmd = &cobra.Command{
	Use:   "customer-addresses",
	Short: "Manage customer addresses",
}

var customerAddressesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List customer addresses",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CustomerAddressesListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListCustomerAddresses(cmd.Context(), customerID, opts)
		if err != nil {
			return fmt.Errorf("failed to list customer addresses: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "ADDRESS", "CITY", "COUNTRY", "DEFAULT", "CREATED"}
		var rows [][]string
		for _, a := range resp.Items {
			name := a.FirstName + " " + a.LastName
			isDefault := "no"
			if a.Default {
				isDefault = "yes"
			}
			rows = append(rows, []string{
				a.ID,
				name,
				a.Address1,
				a.City,
				a.Country,
				isDefault,
				a.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d addresses\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var customerAddressesGetCmd = &cobra.Command{
	Use:   "get <address-id>",
	Short: "Get customer address details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")

		address, err := client.GetCustomerAddress(cmd.Context(), customerID, args[0])
		if err != nil {
			return fmt.Errorf("failed to get customer address: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(address)
		}

		fmt.Printf("Address ID:   %s\n", address.ID)
		fmt.Printf("Customer ID:  %s\n", address.CustomerID)
		fmt.Printf("Name:         %s %s\n", address.FirstName, address.LastName)
		if address.Company != "" {
			fmt.Printf("Company:      %s\n", address.Company)
		}
		fmt.Printf("Address:      %s\n", address.Address1)
		if address.Address2 != "" {
			fmt.Printf("              %s\n", address.Address2)
		}
		fmt.Printf("City:         %s\n", address.City)
		if address.Province != "" {
			fmt.Printf("Province:     %s (%s)\n", address.Province, address.ProvinceCode)
		}
		fmt.Printf("Country:      %s (%s)\n", address.Country, address.CountryCode)
		fmt.Printf("ZIP:          %s\n", address.Zip)
		if address.Phone != "" {
			fmt.Printf("Phone:        %s\n", address.Phone)
		}
		fmt.Printf("Default:      %t\n", address.Default)
		fmt.Printf("Created:      %s\n", address.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:      %s\n", address.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var customerAddressesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a customer address",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		address1, _ := cmd.Flags().GetString("address")
		city, _ := cmd.Flags().GetString("city")
		country, _ := cmd.Flags().GetString("country")
		phone, _ := cmd.Flags().GetString("phone")
		isDefault, _ := cmd.Flags().GetBool("default")

		req := &api.CustomerAddressCreateRequest{
			FirstName: firstName,
			LastName:  lastName,
			Address1:  address1,
			City:      city,
			Country:   country,
			Phone:     phone,
			Default:   isDefault,
		}

		address, err := client.CreateCustomerAddress(cmd.Context(), customerID, req)
		if err != nil {
			return fmt.Errorf("failed to create customer address: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(address)
		}

		fmt.Printf("Created address %s\n", address.ID)
		fmt.Printf("Address: %s, %s, %s\n", address.Address1, address.City, address.Country)
		return nil
	},
}

var customerAddressesSetDefaultCmd = &cobra.Command{
	Use:   "set-default <address-id>",
	Short: "Set an address as the default",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")

		address, err := client.SetDefaultCustomerAddress(cmd.Context(), customerID, args[0])
		if err != nil {
			return fmt.Errorf("failed to set default address: %w", err)
		}

		fmt.Printf("Set address %s as default\n", address.ID)
		return nil
	},
}

var customerAddressesDeleteCmd = &cobra.Command{
	Use:   "delete <address-id>",
	Short: "Delete a customer address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete address %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteCustomerAddress(cmd.Context(), customerID, args[0]); err != nil {
			return fmt.Errorf("failed to delete customer address: %w", err)
		}

		fmt.Printf("Deleted address %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(customerAddressesCmd)

	customerAddressesCmd.PersistentFlags().String("customer-id", "", "Customer ID")
	_ = customerAddressesCmd.MarkPersistentFlagRequired("customer-id")

	customerAddressesCmd.AddCommand(customerAddressesListCmd)
	customerAddressesListCmd.Flags().Int("page", 1, "Page number")
	customerAddressesListCmd.Flags().Int("page-size", 20, "Results per page")

	customerAddressesCmd.AddCommand(customerAddressesGetCmd)

	customerAddressesCmd.AddCommand(customerAddressesCreateCmd)
	customerAddressesCreateCmd.Flags().String("first-name", "", "First name")
	customerAddressesCreateCmd.Flags().String("last-name", "", "Last name")
	customerAddressesCreateCmd.Flags().String("address", "", "Street address")
	customerAddressesCreateCmd.Flags().String("city", "", "City")
	customerAddressesCreateCmd.Flags().String("country", "", "Country")
	customerAddressesCreateCmd.Flags().String("phone", "", "Phone number")
	customerAddressesCreateCmd.Flags().Bool("default", false, "Set as default address")
	_ = customerAddressesCreateCmd.MarkFlagRequired("address")
	_ = customerAddressesCreateCmd.MarkFlagRequired("city")
	_ = customerAddressesCreateCmd.MarkFlagRequired("country")

	customerAddressesCmd.AddCommand(customerAddressesSetDefaultCmd)
	customerAddressesCmd.AddCommand(customerAddressesDeleteCmd)
}
