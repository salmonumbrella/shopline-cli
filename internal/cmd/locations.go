package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var locationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Manage store locations",
}

var locationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.LocationsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListLocations(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list locations: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "CITY", "COUNTRY", "ACTIVE", "DEFAULT", "CREATED"}
		var rows [][]string
		for _, l := range resp.Items {
			active := "no"
			if l.Active {
				active = "yes"
			}
			isDefault := "no"
			if l.IsDefault {
				isDefault = "yes"
			}
			rows = append(rows, []string{
				l.ID,
				l.Name,
				l.City,
				l.Country,
				active,
				isDefault,
				l.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d locations\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var locationsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get location details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		location, err := client.GetLocation(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get location: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(location)
		}

		fmt.Printf("Location ID:    %s\n", location.ID)
		fmt.Printf("Name:           %s\n", location.Name)
		fmt.Printf("Address:        %s\n", location.Address1)
		if location.Address2 != "" {
			fmt.Printf("                %s\n", location.Address2)
		}
		fmt.Printf("City:           %s\n", location.City)
		fmt.Printf("Province:       %s\n", location.Province)
		fmt.Printf("Country:        %s (%s)\n", location.Country, location.CountryCode)
		fmt.Printf("ZIP:            %s\n", location.Zip)
		fmt.Printf("Phone:          %s\n", location.Phone)
		fmt.Printf("Active:         %t\n", location.Active)
		fmt.Printf("Default:        %t\n", location.IsDefault)
		fmt.Printf("Created:        %s\n", location.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", location.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var locationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a location",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		address1, _ := cmd.Flags().GetString("address")
		city, _ := cmd.Flags().GetString("city")
		country, _ := cmd.Flags().GetString("country")
		phone, _ := cmd.Flags().GetString("phone")

		req := &api.LocationCreateRequest{
			Name:     name,
			Address1: address1,
			City:     city,
			Country:  country,
			Phone:    phone,
		}

		location, err := client.CreateLocation(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create location: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(location)
		}

		fmt.Printf("Created location %s\n", location.ID)
		fmt.Printf("Name:    %s\n", location.Name)
		fmt.Printf("Address: %s, %s, %s\n", location.Address1, location.City, location.Country)
		return nil
	},
}

var locationsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete location %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteLocation(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete location: %w", err)
		}

		fmt.Printf("Deleted location %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(locationsCmd)

	locationsCmd.AddCommand(locationsListCmd)
	locationsListCmd.Flags().Int("page", 1, "Page number")
	locationsListCmd.Flags().Int("page-size", 20, "Results per page")

	locationsCmd.AddCommand(locationsGetCmd)

	locationsCmd.AddCommand(locationsCreateCmd)
	locationsCreateCmd.Flags().String("name", "", "Location name")
	locationsCreateCmd.Flags().String("address", "", "Street address")
	locationsCreateCmd.Flags().String("city", "", "City")
	locationsCreateCmd.Flags().String("country", "", "Country")
	locationsCreateCmd.Flags().String("phone", "", "Phone number")
	_ = locationsCreateCmd.MarkFlagRequired("name")
	_ = locationsCreateCmd.MarkFlagRequired("address")
	_ = locationsCreateCmd.MarkFlagRequired("city")
	_ = locationsCreateCmd.MarkFlagRequired("country")

	locationsCmd.AddCommand(locationsDeleteCmd)
}
