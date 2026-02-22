package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var pickupCmd = &cobra.Command{
	Use:   "pickup",
	Short: "Manage store pickup locations",
}

var pickupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pickup locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		locationID, _ := cmd.Flags().GetString("location-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		activeFlag, _ := cmd.Flags().GetString("active")

		opts := &api.PickupListOptions{
			Page:       page,
			PageSize:   pageSize,
			LocationID: locationID,
		}

		if activeFlag != "" {
			active := activeFlag == "true"
			opts.Active = &active
		}

		resp, err := client.ListPickupLocations(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list pickup locations: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "ADDRESS", "CITY", "ACTIVE", "PHONE"}
		var rows [][]string
		for _, l := range resp.Items {
			address := l.Address1
			if l.Address2 != "" {
				address += ", " + l.Address2
			}
			rows = append(rows, []string{
				outfmt.FormatID("pickup_location", l.ID),
				l.Name,
				address,
				l.City,
				strconv.FormatBool(l.Active),
				l.Phone,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d pickup locations\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var pickupGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get pickup location details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		location, err := client.GetPickupLocation(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get pickup location: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(location)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Location ID:      %s\n", location.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:             %s\n", location.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active:           %t\n", location.Active)
		_, _ = fmt.Fprintf(outWriter(cmd), "\n--- Address ---\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "Address 1:        %s\n", location.Address1)
		if location.Address2 != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Address 2:        %s\n", location.Address2)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "City:             %s\n", location.City)
		if location.Province != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Province/State:   %s\n", location.Province)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Country:          %s\n", location.Country)
		if location.ZipCode != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "ZIP Code:         %s\n", location.ZipCode)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "\n--- Contact ---\n")
		if location.Phone != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Phone:            %s\n", location.Phone)
		}
		if location.Email != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Email:            %s\n", location.Email)
		}
		if location.Instructions != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\n--- Instructions ---\n")
			_, _ = fmt.Fprintf(outWriter(cmd), "%s\n", location.Instructions)
		}
		if location.LocationID != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nLinked Location:  %s\n", location.LocationID)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "\nCreated:          %s\n", location.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:          %s\n", location.UpdatedAt.Format(time.RFC3339))

		if len(location.Hours) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\n--- Operating Hours ---\n")
			for _, h := range location.Hours {
				if h.Closed {
					_, _ = fmt.Fprintf(outWriter(cmd), "  %s: Closed\n", h.Day)
				} else {
					_, _ = fmt.Fprintf(outWriter(cmd), "  %s: %s - %s\n", h.Day, h.OpenTime, h.CloseTime)
				}
			}
		}
		return nil
	},
}

var pickupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a pickup location",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create pickup location") {
			return nil
		}

		name, _ := cmd.Flags().GetString("name")
		address1, _ := cmd.Flags().GetString("address1")
		address2, _ := cmd.Flags().GetString("address2")
		city, _ := cmd.Flags().GetString("city")
		province, _ := cmd.Flags().GetString("province")
		country, _ := cmd.Flags().GetString("country")
		zipCode, _ := cmd.Flags().GetString("zip-code")
		phone, _ := cmd.Flags().GetString("phone")
		email, _ := cmd.Flags().GetString("email")
		instructions, _ := cmd.Flags().GetString("instructions")
		active, _ := cmd.Flags().GetBool("active")
		locationID, _ := cmd.Flags().GetString("location-id")

		req := &api.PickupCreateRequest{
			Name:         name,
			Address1:     address1,
			Address2:     address2,
			City:         city,
			Province:     province,
			Country:      country,
			ZipCode:      zipCode,
			Phone:        phone,
			Email:        email,
			Instructions: instructions,
			Active:       active,
			LocationID:   locationID,
		}

		location, err := client.CreatePickupLocation(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create pickup location: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(location)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created pickup location %s: %s\n", location.ID, location.Name)
		return nil
	},
}

var pickupUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a pickup location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update pickup location %s", args[0])) {
			return nil
		}

		var req api.PickupUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		location, err := client.UpdatePickupLocation(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update pickup location: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(location)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated pickup location %s: %s\n", location.ID, location.Name)
		return nil
	},
}

var pickupDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a pickup location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete pickup location %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeletePickupLocation(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete pickup location: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Pickup location %s deleted.\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pickupCmd)

	pickupCmd.AddCommand(pickupListCmd)
	pickupListCmd.Flags().String("location-id", "", "Filter by inventory location ID")
	pickupListCmd.Flags().String("active", "", "Filter by active status (true/false)")
	pickupListCmd.Flags().Int("page", 1, "Page number")
	pickupListCmd.Flags().Int("page-size", 20, "Results per page")

	pickupCmd.AddCommand(pickupGetCmd)

	pickupCmd.AddCommand(pickupCreateCmd)
	pickupCreateCmd.Flags().String("name", "", "Location name")
	pickupCreateCmd.Flags().String("address1", "", "Street address")
	pickupCreateCmd.Flags().String("address2", "", "Address line 2 (apt, suite, etc)")
	pickupCreateCmd.Flags().String("city", "", "City")
	pickupCreateCmd.Flags().String("province", "", "Province or state")
	pickupCreateCmd.Flags().String("country", "", "Country")
	pickupCreateCmd.Flags().String("zip-code", "", "ZIP or postal code")
	pickupCreateCmd.Flags().String("phone", "", "Phone number")
	pickupCreateCmd.Flags().String("email", "", "Email address")
	pickupCreateCmd.Flags().String("instructions", "", "Pickup instructions for customers")
	pickupCreateCmd.Flags().Bool("active", true, "Whether the location is active")
	pickupCreateCmd.Flags().String("location-id", "", "Linked inventory location ID")
	_ = pickupCreateCmd.MarkFlagRequired("name")
	_ = pickupCreateCmd.MarkFlagRequired("address1")
	_ = pickupCreateCmd.MarkFlagRequired("city")
	_ = pickupCreateCmd.MarkFlagRequired("country")

	pickupCmd.AddCommand(pickupUpdateCmd)
	addJSONBodyFlags(pickupUpdateCmd)
	pickupUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	pickupCmd.AddCommand(pickupDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "pickup",
		Description: "Manage store pickup locations",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "pickup_location",
	})
}
