package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

// enrichError wraps an API error with suggestions based on resource context.
func enrichError(err error, resource, resourceID string) error {
	return api.EnrichError(err, resource, resourceID)
}

// handleError formats and prints a rich error, returning the enriched error.
func handleError(cmd *cobra.Command, err error, resource, resourceID string) error {
	enriched := enrichError(err, resource, resourceID)
	formatted := api.FormatRichError(enriched)
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), formatted)
	return enriched
}
