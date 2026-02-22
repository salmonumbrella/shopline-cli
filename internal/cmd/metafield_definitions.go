package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var metafieldDefinitionsCmd = &cobra.Command{
	Use:   "metafield-definitions",
	Short: "Manage metafield definitions",
}

var metafieldDefinitionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List metafield definitions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		ownerType, _ := cmd.Flags().GetString("owner-type")
		namespace, _ := cmd.Flags().GetString("namespace")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.MetafieldDefinitionsListOptions{
			Page:      page,
			PageSize:  pageSize,
			OwnerType: ownerType,
			Namespace: namespace,
		}

		resp, err := client.ListMetafieldDefinitions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list metafield definitions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "NAMESPACE", "KEY", "TYPE", "OWNER TYPE"}
		var rows [][]string
		for _, d := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("metafield_definition", d.ID),
				d.Name,
				d.Namespace,
				d.Key,
				d.Type,
				d.OwnerType,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d metafield definitions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var metafieldDefinitionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get metafield definition details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		def, err := client.GetMetafieldDefinition(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get metafield definition: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(def)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Definition ID:  %s\n", def.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:           %s\n", def.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Namespace:      %s\n", def.Namespace)
		_, _ = fmt.Fprintf(outWriter(cmd), "Key:            %s\n", def.Key)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:           %s\n", def.Type)
		_, _ = fmt.Fprintf(outWriter(cmd), "Owner Type:     %s\n", def.OwnerType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:    %s\n", def.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", def.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", def.UpdatedAt.Format(time.RFC3339))

		if len(def.Validations) > 0 {
			_, _ = fmt.Fprintln(outWriter(cmd), "\nValidations:")
			for _, v := range def.Validations {
				_, _ = fmt.Fprintf(outWriter(cmd), "  %s (%s): %s\n", v.Name, v.Type, v.Value)
			}
		}
		return nil
	},
}

var metafieldDefinitionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a metafield definition",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create metafield definition") {
			return nil
		}

		name, _ := cmd.Flags().GetString("name")
		namespace, _ := cmd.Flags().GetString("namespace")
		key, _ := cmd.Flags().GetString("key")
		fieldType, _ := cmd.Flags().GetString("type")
		ownerType, _ := cmd.Flags().GetString("owner-type")
		description, _ := cmd.Flags().GetString("description")

		req := &api.MetafieldDefinitionCreateRequest{
			Name:        name,
			Namespace:   namespace,
			Key:         key,
			Type:        fieldType,
			OwnerType:   ownerType,
			Description: description,
		}

		def, err := client.CreateMetafieldDefinition(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create metafield definition: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(def)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created metafield definition %s\n", def.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:      %s\n", def.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Namespace: %s\n", def.Namespace)
		_, _ = fmt.Fprintf(outWriter(cmd), "Key:       %s\n", def.Key)
		return nil
	},
}

var metafieldDefinitionsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a metafield definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update metafield definition %s", args[0])) {
			return nil
		}

		var req api.MetafieldDefinitionUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		def, err := client.UpdateMetafieldDefinition(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update metafield definition: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(def)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated metafield definition %s\n", def.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:      %s\n", def.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Namespace: %s\n", def.Namespace)
		_, _ = fmt.Fprintf(outWriter(cmd), "Key:       %s\n", def.Key)
		return nil
	},
}

var metafieldDefinitionsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a metafield definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete metafield definition %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteMetafieldDefinition(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete metafield definition: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted metafield definition %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(metafieldDefinitionsCmd)

	metafieldDefinitionsCmd.AddCommand(metafieldDefinitionsListCmd)
	metafieldDefinitionsListCmd.Flags().String("owner-type", "", "Filter by owner type (product, variant, customer, etc.)")
	metafieldDefinitionsListCmd.Flags().String("namespace", "", "Filter by namespace")
	metafieldDefinitionsListCmd.Flags().Int("page", 1, "Page number")
	metafieldDefinitionsListCmd.Flags().Int("page-size", 20, "Results per page")

	metafieldDefinitionsCmd.AddCommand(metafieldDefinitionsGetCmd)

	metafieldDefinitionsCmd.AddCommand(metafieldDefinitionsCreateCmd)
	metafieldDefinitionsCreateCmd.Flags().String("name", "", "Definition name")
	metafieldDefinitionsCreateCmd.Flags().String("namespace", "", "Metafield namespace")
	metafieldDefinitionsCreateCmd.Flags().String("key", "", "Metafield key")
	metafieldDefinitionsCreateCmd.Flags().String("type", "", "Value type (single_line_text_field, number_integer, etc.)")
	metafieldDefinitionsCreateCmd.Flags().String("owner-type", "", "Owner type (product, variant, customer, etc.)")
	metafieldDefinitionsCreateCmd.Flags().String("description", "", "Definition description")
	_ = metafieldDefinitionsCreateCmd.MarkFlagRequired("name")
	_ = metafieldDefinitionsCreateCmd.MarkFlagRequired("namespace")
	_ = metafieldDefinitionsCreateCmd.MarkFlagRequired("key")
	_ = metafieldDefinitionsCreateCmd.MarkFlagRequired("type")
	_ = metafieldDefinitionsCreateCmd.MarkFlagRequired("owner-type")

	metafieldDefinitionsCmd.AddCommand(metafieldDefinitionsUpdateCmd)
	addJSONBodyFlags(metafieldDefinitionsUpdateCmd)
	metafieldDefinitionsUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	metafieldDefinitionsCmd.AddCommand(metafieldDefinitionsDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "metafield-definitions",
		Description: "Manage metafield definitions",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "metafield_definition",
	})
}
