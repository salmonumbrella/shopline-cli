package cmd

import (
	"fmt"
	"strings"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var customFieldsCmd = &cobra.Command{
	Use:     "custom-fields",
	Aliases: []string{"custom-field", "cf"},
	Short:   "Manage custom field definitions",
}

var customFieldsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom fields",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		ownerType, _ := cmd.Flags().GetString("owner-type")
		fieldType, _ := cmd.Flags().GetString("type")

		opts := &api.CustomFieldsListOptions{
			Page:      page,
			PageSize:  pageSize,
			OwnerType: api.CustomFieldOwnerType(ownerType),
			Type:      api.CustomFieldType(fieldType),
		}

		resp, err := client.ListCustomFields(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list custom fields: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "TYPE", "HINT"}
		var rows [][]string
		for _, f := range resp.Items {
			// Get name from translations (prefer zh-hant, fallback to first available)
			name := ""
			for _, v := range f.NameTranslations {
				name = v
				break
			}
			if n, ok := f.NameTranslations["zh-hant"]; ok {
				name = n
			}

			// Get hint from translations
			hint := ""
			for _, v := range f.HintTranslations {
				hint = v
				break
			}
			if h, ok := f.HintTranslations["zh-hant"]; ok {
				hint = h
			}

			rows = append(rows, []string{
				outfmt.FormatID("custom_field", f.ID),
				name,
				string(f.Type),
				hint,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d custom fields\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var customFieldsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get custom field details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		field, err := client.GetCustomField(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get custom field: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(field)
		}

		// Get name from translations
		name := ""
		for _, v := range field.NameTranslations {
			name = v
			break
		}
		if n, ok := field.NameTranslations["zh-hant"]; ok {
			name = n
		}

		// Get hint from translations
		hint := ""
		for _, v := range field.HintTranslations {
			hint = v
			break
		}
		if h, ok := field.HintTranslations["zh-hant"]; ok {
			hint = h
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Custom Field ID: %s\n", field.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:            %s\n", name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:            %s\n", field.Type)
		if hint != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Hint:            %s\n", hint)
		}

		return nil
	},
}

var customFieldsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom field",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		key, _ := cmd.Flags().GetString("key")
		description, _ := cmd.Flags().GetString("description")
		fieldType, _ := cmd.Flags().GetString("type")
		ownerType, _ := cmd.Flags().GetString("owner-type")
		required, _ := cmd.Flags().GetBool("required")
		searchable, _ := cmd.Flags().GetBool("searchable")
		visible, _ := cmd.Flags().GetBool("visible")
		defaultValue, _ := cmd.Flags().GetString("default-value")
		optionsStr, _ := cmd.Flags().GetString("options")
		validation, _ := cmd.Flags().GetString("validation")
		position, _ := cmd.Flags().GetInt("position")

		var options []string
		if optionsStr != "" {
			options = strings.Split(optionsStr, ",")
			for i := range options {
				options[i] = strings.TrimSpace(options[i])
			}
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create custom field: %s (%s) for %s", name, fieldType, ownerType)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.CustomFieldCreateRequest{
			Name:         name,
			Key:          key,
			Description:  description,
			Type:         api.CustomFieldType(fieldType),
			OwnerType:    api.CustomFieldOwnerType(ownerType),
			Required:     required,
			Searchable:   searchable,
			Visible:      visible,
			DefaultValue: defaultValue,
			Options:      options,
			Validation:   validation,
			Position:     position,
		}

		field, err := client.CreateCustomField(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create custom field: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(field)
		}

		// Get display name from translations
		displayName := ""
		for _, v := range field.NameTranslations {
			displayName = v
			break
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created custom field %s\n", field.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:  %s\n", displayName)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:  %s\n", field.Type)

		return nil
	},
}

var customFieldsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a custom field",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &api.CustomFieldUpdateRequest{}

		if cmd.Flags().Changed("name") {
			req.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("description") {
			req.Description, _ = cmd.Flags().GetString("description")
		}
		if cmd.Flags().Changed("required") {
			v, _ := cmd.Flags().GetBool("required")
			req.Required = &v
		}
		if cmd.Flags().Changed("searchable") {
			v, _ := cmd.Flags().GetBool("searchable")
			req.Searchable = &v
		}
		if cmd.Flags().Changed("visible") {
			v, _ := cmd.Flags().GetBool("visible")
			req.Visible = &v
		}
		if cmd.Flags().Changed("default-value") {
			req.DefaultValue, _ = cmd.Flags().GetString("default-value")
		}
		if cmd.Flags().Changed("options") {
			optionsStr, _ := cmd.Flags().GetString("options")
			options := strings.Split(optionsStr, ",")
			for i := range options {
				options[i] = strings.TrimSpace(options[i])
			}
			req.Options = options
		}
		if cmd.Flags().Changed("validation") {
			req.Validation, _ = cmd.Flags().GetString("validation")
		}
		if cmd.Flags().Changed("position") {
			req.Position, _ = cmd.Flags().GetInt("position")
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update custom field %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		field, err := client.UpdateCustomField(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update custom field: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(field)
		}

		// Get name from translations
		name := ""
		for _, v := range field.NameTranslations {
			name = v
			break
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated custom field %s\n", field.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:    %s\n", name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:    %s\n", field.Type)

		return nil
	},
}

var customFieldsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a custom field",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete custom field %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete custom field %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteCustomField(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete custom field: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted custom field %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(customFieldsCmd)

	customFieldsCmd.AddCommand(customFieldsListCmd)
	customFieldsListCmd.Flags().Int("page", 1, "Page number")
	customFieldsListCmd.Flags().Int("page-size", 20, "Results per page")
	customFieldsListCmd.Flags().String("owner-type", "", "Filter by owner type (product, variant, customer, order, shop)")
	customFieldsListCmd.Flags().String("type", "", "Filter by field type (text, number, date, boolean, select, etc.)")

	customFieldsCmd.AddCommand(customFieldsGetCmd)

	customFieldsCmd.AddCommand(customFieldsCreateCmd)
	customFieldsCreateCmd.Flags().String("name", "", "Field name (required)")
	customFieldsCreateCmd.Flags().String("key", "", "Field key (required)")
	customFieldsCreateCmd.Flags().String("description", "", "Field description")
	customFieldsCreateCmd.Flags().String("type", "text", "Field type (text, number, date, boolean, select, multi_select, file, url, email, json)")
	customFieldsCreateCmd.Flags().String("owner-type", "product", "Owner type (product, variant, customer, order, shop)")
	customFieldsCreateCmd.Flags().Bool("required", false, "Make field required")
	customFieldsCreateCmd.Flags().Bool("searchable", false, "Make field searchable")
	customFieldsCreateCmd.Flags().Bool("visible", true, "Make field visible")
	customFieldsCreateCmd.Flags().String("default-value", "", "Default value")
	customFieldsCreateCmd.Flags().String("options", "", "Comma-separated options for select types")
	customFieldsCreateCmd.Flags().String("validation", "", "Validation regex")
	customFieldsCreateCmd.Flags().Int("position", 0, "Display position")
	_ = customFieldsCreateCmd.MarkFlagRequired("name")
	_ = customFieldsCreateCmd.MarkFlagRequired("key")

	customFieldsCmd.AddCommand(customFieldsUpdateCmd)
	customFieldsUpdateCmd.Flags().String("name", "", "Field name")
	customFieldsUpdateCmd.Flags().String("description", "", "Field description")
	customFieldsUpdateCmd.Flags().Bool("required", false, "Make field required")
	customFieldsUpdateCmd.Flags().Bool("searchable", false, "Make field searchable")
	customFieldsUpdateCmd.Flags().Bool("visible", false, "Make field visible")
	customFieldsUpdateCmd.Flags().String("default-value", "", "Default value")
	customFieldsUpdateCmd.Flags().String("options", "", "Comma-separated options for select types")
	customFieldsUpdateCmd.Flags().String("validation", "", "Validation regex")
	customFieldsUpdateCmd.Flags().Int("position", 0, "Display position")

	customFieldsCmd.AddCommand(customFieldsDeleteCmd)
	customFieldsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}
