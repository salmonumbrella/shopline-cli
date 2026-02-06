package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var channelsPricesCmd = &cobra.Command{
	Use:   "prices",
	Short: "Manage product channel prices",
}

var channelsPricesGetCmd = &cobra.Command{
	Use:   "get <channel-id>",
	Short: "Get product channel prices for a channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetChannelPrices(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get channel prices: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var channelsPricesCreateCmd = &cobra.Command{
	Use:   "create <channel-id> <product-id>",
	Short: "Create a product channel price",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := client.CreateChannelProductPrice(cmd.Context(), args[0], args[1], body)
		if err != nil {
			return fmt.Errorf("failed to create channel product price: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var channelsPricesUpdateCmd = &cobra.Command{
	Use:   "update <channel-id> <product-id> <price-id>",
	Short: "Update a product channel price",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := client.UpdateChannelProductPrice(cmd.Context(), args[0], args[1], args[2], body)
		if err != nil {
			return fmt.Errorf("failed to update channel product price: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	channelsCmd.AddCommand(channelsPricesCmd)

	channelsPricesCmd.AddCommand(channelsPricesGetCmd)
	channelsPricesCmd.AddCommand(channelsPricesCreateCmd)
	addJSONBodyFlags(channelsPricesCreateCmd)
	channelsPricesCmd.AddCommand(channelsPricesUpdateCmd)
	addJSONBodyFlags(channelsPricesUpdateCmd)
}
