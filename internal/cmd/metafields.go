package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var metafieldsCmd = &cobra.Command{
	Use:   "metafields",
	Short: "Manage metafields",
}

var metafieldsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List metafields",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		namespace, _ := cmd.Flags().GetString("namespace")
		key, _ := cmd.Flags().GetString("key")
		ownerType, _ := cmd.Flags().GetString("owner-type")
		ownerID, _ := cmd.Flags().GetString("owner-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.MetafieldsListOptions{
			Page:      page,
			PageSize:  pageSize,
			Namespace: namespace,
			Key:       key,
			OwnerType: ownerType,
			OwnerID:   ownerID,
		}

		resp, err := client.ListMetafields(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list metafields: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAMESPACE", "KEY", "VALUE", "TYPE", "OWNER", "CREATED"}
		var rows [][]string
		for _, m := range resp.Items {
			value := m.Value
			if len(value) > 30 {
				value = value[:27] + "..."
			}
			owner := m.OwnerType
			if m.OwnerID != "" {
				owner = fmt.Sprintf("%s:%s", m.OwnerType, m.OwnerID)
			}
			rows = append(rows, []string{
				outfmt.FormatID("metafield", m.ID),
				m.Namespace,
				m.Key,
				value,
				m.ValueType,
				owner,
				m.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d metafields\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var metafieldsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get metafield details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		metafield, err := client.GetMetafield(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get metafield: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(metafield)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Metafield ID:   %s\n", metafield.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Namespace:      %s\n", metafield.Namespace)
		_, _ = fmt.Fprintf(outWriter(cmd), "Key:            %s\n", metafield.Key)
		_, _ = fmt.Fprintf(outWriter(cmd), "Value:          %s\n", metafield.Value)
		_, _ = fmt.Fprintf(outWriter(cmd), "Value Type:     %s\n", metafield.ValueType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:    %s\n", metafield.Description)
		if metafield.OwnerType != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Owner Type:     %s\n", metafield.OwnerType)
			_, _ = fmt.Fprintf(outWriter(cmd), "Owner ID:       %s\n", metafield.OwnerID)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", metafield.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", metafield.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var metafieldsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a metafield",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create metafield") {
			return nil
		}

		namespace, _ := cmd.Flags().GetString("namespace")
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		valueType, _ := cmd.Flags().GetString("type")
		description, _ := cmd.Flags().GetString("description")
		ownerType, _ := cmd.Flags().GetString("owner-type")
		ownerID, _ := cmd.Flags().GetString("owner-id")

		req := &api.MetafieldCreateRequest{
			Namespace:   namespace,
			Key:         key,
			Value:       value,
			ValueType:   valueType,
			Description: description,
			OwnerType:   ownerType,
			OwnerID:     ownerID,
		}

		metafield, err := client.CreateMetafield(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create metafield: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(metafield)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created metafield %s\n", metafield.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Namespace: %s\n", metafield.Namespace)
		_, _ = fmt.Fprintf(outWriter(cmd), "Key:       %s\n", metafield.Key)
		_, _ = fmt.Fprintf(outWriter(cmd), "Value:     %s\n", metafield.Value)
		return nil
	},
}

var metafieldsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a metafield",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update metafield %s", args[0])) {
			return nil
		}

		var req api.MetafieldUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		metafield, err := client.UpdateMetafield(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update metafield: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(metafield)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated metafield %s\n", metafield.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Namespace: %s\n", metafield.Namespace)
		_, _ = fmt.Fprintf(outWriter(cmd), "Key:       %s\n", metafield.Key)
		_, _ = fmt.Fprintf(outWriter(cmd), "Value:     %s\n", metafield.Value)
		return nil
	},
}

var metafieldsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a metafield",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete metafield %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteMetafield(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete metafield: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted metafield %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(metafieldsCmd)

	metafieldsCmd.AddCommand(metafieldsListCmd)
	metafieldsListCmd.Flags().String("namespace", "", "Filter by namespace")
	metafieldsListCmd.Flags().String("key", "", "Filter by key")
	metafieldsListCmd.Flags().String("owner-type", "", "Filter by owner type (product, order, customer, etc)")
	metafieldsListCmd.Flags().String("owner-id", "", "Filter by owner ID")
	metafieldsListCmd.Flags().Int("page", 1, "Page number")
	metafieldsListCmd.Flags().Int("page-size", 20, "Results per page")

	metafieldsCmd.AddCommand(metafieldsGetCmd)

	metafieldsCmd.AddCommand(metafieldsCreateCmd)
	metafieldsCreateCmd.Flags().String("namespace", "", "Metafield namespace")
	metafieldsCreateCmd.Flags().String("key", "", "Metafield key")
	metafieldsCreateCmd.Flags().String("value", "", "Metafield value")
	metafieldsCreateCmd.Flags().String("type", "string", "Value type (string, integer, json, etc)")
	metafieldsCreateCmd.Flags().String("description", "", "Metafield description")
	metafieldsCreateCmd.Flags().String("owner-type", "", "Owner type (product, order, customer, etc)")
	metafieldsCreateCmd.Flags().String("owner-id", "", "Owner ID")
	_ = metafieldsCreateCmd.MarkFlagRequired("namespace")
	_ = metafieldsCreateCmd.MarkFlagRequired("key")
	_ = metafieldsCreateCmd.MarkFlagRequired("value")

	metafieldsCmd.AddCommand(metafieldsUpdateCmd)
	addJSONBodyFlags(metafieldsUpdateCmd)
	metafieldsUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	metafieldsCmd.AddCommand(metafieldsDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "metafields",
		Description: "Manage metafields",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "metafield",
	})
}
