package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage product tags",
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("q")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.TagsListOptions{
			Page:     page,
			PageSize: pageSize,
			Query:    query,
		}

		resp, err := client.ListTags(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list tags: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "HANDLE", "PRODUCT COUNT", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("tag", t.ID),
				t.Name,
				t.Handle,
				fmt.Sprintf("%d", t.ProductCount),
				t.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d tags\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var tagsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get tag details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		tag, err := client.GetTag(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get tag: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(tag)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Tag ID:         %s\n", tag.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:           %s\n", tag.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:         %s\n", tag.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product Count:  %d\n", tag.ProductCount)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", tag.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", tag.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var tagsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a tag",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create tag '%s'", name)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.TagCreateRequest{
			Name: name,
		}

		tag, err := client.CreateTag(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(tag)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created tag %s\n", tag.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:    %s\n", tag.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:  %s\n", tag.Handle)

		return nil
	},
}

var tagsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete tag %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteTag(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete tag: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted tag %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)

	tagsCmd.AddCommand(tagsListCmd)
	tagsListCmd.Flags().String("q", "", "Search tags by name")
	tagsListCmd.Flags().Int("page", 1, "Page number")
	tagsListCmd.Flags().Int("page-size", 20, "Results per page")

	tagsCmd.AddCommand(tagsGetCmd)

	tagsCmd.AddCommand(tagsCreateCmd)
	tagsCreateCmd.Flags().String("name", "", "Tag name")
	_ = tagsCreateCmd.MarkFlagRequired("name")

	tagsCmd.AddCommand(tagsDeleteCmd)
}
