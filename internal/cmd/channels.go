package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var channelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Manage sales channels",
}

var channelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sales channels",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		limit, _ := cmd.Flags().GetInt("limit")

		opts := &api.ChannelsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		if cmd.Flags().Changed("active") {
			active, _ := cmd.Flags().GetBool("active")
			opts.Active = &active
		}

		resp := &api.ChannelsListResponse{}
		if limit > 0 {
			curPage := opts.Page
			perPage := opts.PageSize
			if perPage <= 0 || perPage > limit {
				perPage = limit
			}

			items := make([]api.Channel, 0, limit)
			totalCount := 0
			hasMore := false

			for len(items) < limit {
				pageOpts := *opts
				pageOpts.Page = curPage
				pageOpts.PageSize = perPage

				pageResp, err := client.ListChannels(cmd.Context(), &pageOpts)
				if err != nil {
					return fmt.Errorf("failed to list channels: %w", err)
				}
				if totalCount == 0 {
					totalCount = pageResp.TotalCount
				}
				items = append(items, pageResp.Items...)
				hasMore = pageResp.HasMore

				if !pageResp.HasMore || len(pageResp.Items) == 0 {
					break
				}
				curPage++
			}

			if len(items) > limit {
				items = items[:limit]
				hasMore = true
			}

			resp.Items = items
			resp.Page = opts.Page
			resp.PageSize = perPage
			resp.TotalCount = totalCount
			resp.HasMore = hasMore
		} else {
			r, err := client.ListChannels(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("failed to list channels: %w", err)
			}
			resp = r
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "HANDLE", "TYPE", "ACTIVE", "PRODUCTS"}
		var rows [][]string
		for _, c := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("channel", c.ID),
				c.Name,
				c.Handle,
				c.Type,
				fmt.Sprintf("%t", c.Active),
				fmt.Sprintf("%d", c.ProductCount),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d channels\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var channelsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get channel details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		channel, err := client.GetChannel(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get channel: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(channel)
		}

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Channel ID:        %s\n", channel.ID)
		_, _ = fmt.Fprintf(out, "Name:              %s\n", channel.Name)
		_, _ = fmt.Fprintf(out, "Handle:            %s\n", channel.Handle)
		_, _ = fmt.Fprintf(out, "Type:              %s\n", channel.Type)
		_, _ = fmt.Fprintf(out, "Active:            %t\n", channel.Active)
		_, _ = fmt.Fprintf(out, "Products:          %d\n", channel.ProductCount)
		_, _ = fmt.Fprintf(out, "Collections:       %d\n", channel.CollectionCount)
		_, _ = fmt.Fprintf(out, "Remote Fulfillment: %t\n", channel.SupportsRemoteFulfillment)
		_, _ = fmt.Fprintf(out, "Created:           %s\n", channel.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(out, "Updated:           %s\n", channel.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var channelsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a sales channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create channel") {
			return nil
		}

		name, _ := cmd.Flags().GetString("name")
		handle, _ := cmd.Flags().GetString("handle")
		channelType, _ := cmd.Flags().GetString("type")

		req := &api.ChannelCreateRequest{
			Name:   name,
			Handle: handle,
			Type:   channelType,
		}

		channel, err := client.CreateChannel(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create channel: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(channel)
		}

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Created channel %s\n", channel.ID)
		_, _ = fmt.Fprintf(out, "Name:   %s\n", channel.Name)
		_, _ = fmt.Fprintf(out, "Handle: %s\n", channel.Handle)
		_, _ = fmt.Fprintf(out, "Type:   %s\n", channel.Type)
		return nil
	},
}

var channelsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a sales channel",
	Long:  "Update a sales channel using either --body/--body-file (raw JSON) or individual flags (--name, --active).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update channel %s", args[0])) {
			return nil
		}

		hasBody := cmd.Flags().Changed("body") || cmd.Flags().Changed("body-file")
		hasFlags := cmd.Flags().Changed("name") || cmd.Flags().Changed("active")

		if hasBody && hasFlags {
			return fmt.Errorf("use either --body/--body-file or individual flags, not both")
		}
		if !hasBody && !hasFlags {
			return fmt.Errorf("provide channel data via --body/--body-file or individual flags (--name, --active)")
		}

		var req api.ChannelUpdateRequest
		if hasBody {
			if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
				return err
			}
		} else {
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = &name
			}
			if cmd.Flags().Changed("active") {
				active, _ := cmd.Flags().GetBool("active")
				req.Active = &active
			}
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		channel, err := client.UpdateChannel(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update channel: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(channel)
		}

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Updated channel %s\n", channel.ID)
		_, _ = fmt.Fprintf(out, "Name:   %s\n", channel.Name)
		_, _ = fmt.Fprintf(out, "Handle: %s\n", channel.Handle)
		_, _ = fmt.Fprintf(out, "Type:   %s\n", channel.Type)
		return nil
	},
}

var channelsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a sales channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete channel %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete channel %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteChannel(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete channel: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted channel %s\n", args[0])
		return nil
	},
}

var channelsProductsCmd = &cobra.Command{
	Use:   "products <channel-id>",
	Short: "List products in a channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.ListChannelProducts(cmd.Context(), args[0], page, pageSize)
		if err != nil {
			return fmt.Errorf("failed to list channel products: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"PRODUCT ID", "PUBLISHED"}
		var rows [][]string
		for _, p := range resp.Items {
			rows = append(rows, []string{
				p.ProductID,
				fmt.Sprintf("%t", p.Published),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d products in channel\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var channelsPublishCmd = &cobra.Command{
	Use:   "publish <channel-id> <product-id>",
	Short: "Publish a product to a channel",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would publish product to channel") {
			return nil
		}

		req := &api.ChannelPublishProductRequest{
			ProductID: args[1],
		}

		if err := client.PublishProductToChannel(cmd.Context(), args[0], req); err != nil {
			return fmt.Errorf("failed to publish product: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Published product %s to channel %s\n", args[1], args[0])
		return nil
	},
}

var channelsUnpublishCmd = &cobra.Command{
	Use:   "unpublish <channel-id> <product-id>",
	Short: "Unpublish a product from a channel",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would unpublish product from channel") {
			return nil
		}

		if err := client.UnpublishProductFromChannel(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to unpublish product: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Unpublished product %s from channel %s\n", args[1], args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(channelsCmd)

	channelsCmd.AddCommand(channelsListCmd)
	channelsListCmd.Flags().Bool("active", false, "Filter by active status")
	channelsListCmd.Flags().Int("page", 1, "Page number")
	channelsListCmd.Flags().Int("page-size", 20, "Results per page")

	channelsCmd.AddCommand(channelsGetCmd)

	channelsCmd.AddCommand(channelsCreateCmd)
	channelsCreateCmd.Flags().String("name", "", "Channel name")
	channelsCreateCmd.Flags().String("handle", "", "Channel handle (URL slug)")
	channelsCreateCmd.Flags().String("type", "", "Channel type (online_store, point_of_sale, mobile, etc.)")
	_ = channelsCreateCmd.MarkFlagRequired("name")
	_ = channelsCreateCmd.MarkFlagRequired("type")

	channelsCmd.AddCommand(channelsUpdateCmd)
	addJSONBodyFlags(channelsUpdateCmd)
	channelsUpdateCmd.Flags().String("name", "", "Channel name")
	channelsUpdateCmd.Flags().Bool("active", false, "Channel active status")

	channelsCmd.AddCommand(channelsDeleteCmd)

	channelsCmd.AddCommand(channelsProductsCmd)
	channelsProductsCmd.Flags().Int("page", 1, "Page number")
	channelsProductsCmd.Flags().Int("page-size", 20, "Results per page")

	channelsCmd.AddCommand(channelsPublishCmd)
	channelsCmd.AddCommand(channelsUnpublishCmd)
}
