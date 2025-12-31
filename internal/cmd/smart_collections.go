package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var smartCollectionsCmd = &cobra.Command{
	Use:   "smart-collections",
	Short: "Manage smart collections (auto-populated based on rules)",
}

var smartCollectionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List smart collections",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.SmartCollectionsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListSmartCollections(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list smart collections: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "HANDLE", "RULES", "DISJUNCTIVE", "PUBLISHED", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			disjunctive := "all match"
			if c.Disjunctive {
				disjunctive = "any match"
			}
			published := "no"
			if c.Published {
				published = "yes"
			}
			rows = append(rows, []string{
				c.ID,
				c.Title,
				c.Handle,
				fmt.Sprintf("%d", len(c.Rules)),
				disjunctive,
				published,
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d smart collections\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var smartCollectionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get smart collection details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		collection, err := client.GetSmartCollection(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get smart collection: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(collection)
		}

		fmt.Printf("Smart Collection ID: %s\n", collection.ID)
		fmt.Printf("Title:               %s\n", collection.Title)
		fmt.Printf("Handle:              %s\n", collection.Handle)
		fmt.Printf("Body HTML:           %s\n", collection.BodyHTML)
		fmt.Printf("Sort Order:          %s\n", collection.SortOrder)
		fmt.Printf("Disjunctive:         %v\n", collection.Disjunctive)
		fmt.Printf("Published:           %v\n", collection.Published)
		fmt.Printf("Published At:        %s\n", collection.PublishedAt.Format(time.RFC3339))
		fmt.Printf("Created:             %s\n", collection.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:             %s\n", collection.UpdatedAt.Format(time.RFC3339))
		fmt.Printf("\nRules (%d):\n", len(collection.Rules))
		for i, rule := range collection.Rules {
			fmt.Printf("  %d. %s %s %q\n", i+1, rule.Column, rule.Relation, rule.Condition)
		}
		return nil
	},
}

var smartCollectionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a smart collection",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		title, _ := cmd.Flags().GetString("title")
		handle, _ := cmd.Flags().GetString("handle")
		bodyHTML, _ := cmd.Flags().GetString("body-html")
		sortOrder, _ := cmd.Flags().GetString("sort-order")
		disjunctive, _ := cmd.Flags().GetBool("disjunctive")
		published, _ := cmd.Flags().GetBool("published")
		rulesJSON, _ := cmd.Flags().GetString("rules")

		var rules []api.Rule
		if rulesJSON != "" {
			if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
				return fmt.Errorf("failed to parse rules JSON: %w", err)
			}
		}

		req := &api.SmartCollectionCreateRequest{
			Title:       title,
			Handle:      handle,
			BodyHTML:    bodyHTML,
			SortOrder:   sortOrder,
			Disjunctive: disjunctive,
			Rules:       rules,
			Published:   published,
		}

		collection, err := client.CreateSmartCollection(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create smart collection: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(collection)
		}

		fmt.Printf("Created smart collection %s\n", collection.ID)
		fmt.Printf("Title:  %s\n", collection.Title)
		fmt.Printf("Handle: %s\n", collection.Handle)
		fmt.Printf("Rules:  %d\n", len(collection.Rules))
		return nil
	},
}

var smartCollectionsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a smart collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete smart collection %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteSmartCollection(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete smart collection: %w", err)
		}

		fmt.Printf("Deleted smart collection %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(smartCollectionsCmd)

	smartCollectionsCmd.AddCommand(smartCollectionsListCmd)
	smartCollectionsListCmd.Flags().Int("page", 1, "Page number")
	smartCollectionsListCmd.Flags().Int("page-size", 20, "Results per page")

	smartCollectionsCmd.AddCommand(smartCollectionsGetCmd)

	smartCollectionsCmd.AddCommand(smartCollectionsCreateCmd)
	smartCollectionsCreateCmd.Flags().String("title", "", "Smart collection title")
	smartCollectionsCreateCmd.Flags().String("handle", "", "Smart collection handle (URL slug)")
	smartCollectionsCreateCmd.Flags().String("body-html", "", "Smart collection description HTML")
	smartCollectionsCreateCmd.Flags().String("sort-order", "", "Product sort order (alpha-asc, alpha-desc, best-selling, created, created-desc, manual, price-asc, price-desc)")
	smartCollectionsCreateCmd.Flags().Bool("disjunctive", false, "Match any rule (true) or all rules (false)")
	smartCollectionsCreateCmd.Flags().Bool("published", true, "Publish the collection")
	smartCollectionsCreateCmd.Flags().String("rules", "", "Rules as JSON array, e.g. '[{\"column\":\"tag\",\"relation\":\"equals\",\"condition\":\"sale\"}]'")
	_ = smartCollectionsCreateCmd.MarkFlagRequired("title")
	_ = smartCollectionsCreateCmd.MarkFlagRequired("rules")

	smartCollectionsCmd.AddCommand(smartCollectionsDeleteCmd)
}
