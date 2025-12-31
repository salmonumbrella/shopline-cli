package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var scriptTagsCmd = &cobra.Command{
	Use:   "script-tags",
	Short: "Manage script tags",
}

var scriptTagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List script tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		src, _ := cmd.Flags().GetString("src")

		opts := &api.ScriptTagsListOptions{
			Page:     page,
			PageSize: pageSize,
			Src:      src,
		}

		resp, err := client.ListScriptTags(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list script tags: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "SRC", "EVENT", "DISPLAY SCOPE", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			rows = append(rows, []string{
				t.ID,
				t.Src,
				t.Event,
				t.DisplayScope,
				t.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d script tags\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var scriptTagsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get script tag details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		tag, err := client.GetScriptTag(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get script tag: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(tag)
		}

		fmt.Printf("ID:            %s\n", tag.ID)
		fmt.Printf("Src:           %s\n", tag.Src)
		fmt.Printf("Event:         %s\n", tag.Event)
		fmt.Printf("Display Scope: %s\n", tag.DisplayScope)
		fmt.Printf("Created:       %s\n", tag.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:       %s\n", tag.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var scriptTagsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a script tag",
	RunE: func(cmd *cobra.Command, args []string) error {
		src, _ := cmd.Flags().GetString("src")
		event, _ := cmd.Flags().GetString("event")
		displayScope, _ := cmd.Flags().GetString("display-scope")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.ScriptTagCreateRequest{
			Src:          src,
			Event:        event,
			DisplayScope: displayScope,
		}

		tag, err := client.CreateScriptTag(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create script tag: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(tag)
		}

		fmt.Printf("Created script tag %s\n", tag.ID)
		fmt.Printf("Src:           %s\n", tag.Src)
		fmt.Printf("Event:         %s\n", tag.Event)
		fmt.Printf("Display Scope: %s\n", tag.DisplayScope)

		return nil
	},
}

var scriptTagsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a script tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[dry-run] Would delete script tag %s\n", args[0])
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete script tag %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteScriptTag(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete script tag: %w", err)
		}

		fmt.Printf("Deleted script tag %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scriptTagsCmd)

	scriptTagsCmd.AddCommand(scriptTagsListCmd)
	scriptTagsListCmd.Flags().Int("page", 1, "Page number")
	scriptTagsListCmd.Flags().Int("page-size", 20, "Results per page")
	scriptTagsListCmd.Flags().String("src", "", "Filter by script source URL")

	scriptTagsCmd.AddCommand(scriptTagsGetCmd)

	scriptTagsCmd.AddCommand(scriptTagsCreateCmd)
	scriptTagsCreateCmd.Flags().String("src", "", "Script source URL")
	scriptTagsCreateCmd.Flags().String("event", "", "Event trigger (e.g., onload)")
	scriptTagsCreateCmd.Flags().String("display-scope", "", "Display scope (e.g., all, online_store)")
	_ = scriptTagsCreateCmd.MarkFlagRequired("src")

	scriptTagsCmd.AddCommand(scriptTagsDeleteCmd)
}
