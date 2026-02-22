package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/auth"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:     "auth",
	Aliases: []string{"config", "store", "profile", "profiles"},
	Short:   "Manage authentication",
	Long:    `Add, remove, and manage store authentication profiles.`,
}

var authAddCmd = &cobra.Command{
	Use:     "login",
	Aliases: []string{"add"},
	Short:   "Add a new store profile via browser",
	Long:    `Opens a browser window for interactive authentication setup. Use --no-browser to print the URL only.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would add store profile via browser") {
			return nil
		}

		restoreNoBrowser := setNoBrowserEnv(authNoBrowser)
		defer restoreNoBrowser()

		server, err := auth.NewServer()
		if err != nil {
			return fmt.Errorf("failed to create auth server: %w", err)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
		defer cancel()

		creds, err := server.Run(ctx)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		store, err := secrets.NewStore()
		if err != nil {
			return fmt.Errorf("failed to open credential store: %w", err)
		}

		if err := store.Save(creds); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "\nSuccessfully added store: %s\n", creds.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Try: shopline --store %s orders list --limit 5\n", creds.Handle)
		return nil
	},
}

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured store profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := secrets.NewStore()
		if err != nil {
			return fmt.Errorf("failed to open credential store: %w", err)
		}

		names, err := store.List()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		if len(names) == 0 {
			_, _ = fmt.Fprintln(outWriter(cmd), "No store profiles configured. Use 'spl auth login' to add one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tHANDLE\tCREATED\tSTATUS")

		for _, name := range names {
			creds, err := store.Get(name)
			if err != nil {
				continue
			}

			status := "OK"
			if creds.IsOld() {
				status = "ROTATE"
			}

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				creds.Name,
				creds.Handle,
				creds.CreatedAt.Format("2006-01-02"),
				status,
			)
		}
		_ = w.Flush()

		return nil
	},
}

var authRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a store profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would remove store profile %s", args[0])) {
			return nil
		}

		store, err := secrets.NewStore()
		if err != nil {
			return fmt.Errorf("failed to open credential store: %w", err)
		}

		if err := store.Delete(name); err != nil {
			return fmt.Errorf("failed to remove profile: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Removed store profile: %s\n", name)
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		storeName, _ := cmd.Flags().GetString("store")
		if storeName == "" {
			storeName = os.Getenv(defaultStoreEnvName)
		}
		storeName = resolveStoreAlias(storeName)

		store, err := secrets.NewStore()
		if err != nil {
			return fmt.Errorf("failed to open credential store: %w", err)
		}

		if storeName == "" {
			names, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}
			if len(names) == 1 {
				storeName = names[0]
			} else if len(names) == 0 {
				_, _ = fmt.Fprintln(outWriter(cmd), "Not authenticated. Use 'spl auth login' to add a profile.")
				return nil
			} else {
				_, _ = fmt.Fprintln(outWriter(cmd), "Multiple profiles configured. Use --store to select one:")
				for _, n := range names {
					_, _ = fmt.Fprintf(outWriter(cmd), "  - %s\n", n)
				}
				return nil
			}
		}

		creds, err := resolveStoreCredentials(store, storeName)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Profile:  %s\n", creds.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:   %s\n", creds.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:  %s\n", creds.CreatedAt.Format("2006-01-02 15:04:05"))

		if creds.IsOld() {
			_, _ = fmt.Fprintln(outWriter(cmd), "\nWarning: Credentials are older than 90 days. Consider rotating them.")
		}

		return nil
	},
}

var authNoBrowser bool

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authAddCmd)
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authRemoveCmd)
	authCmd.AddCommand(authStatusCmd)

	authAddCmd.Flags().BoolVar(&authNoBrowser, "no-browser", false, "Print auth URL without opening browser automatically")
}

func setNoBrowserEnv(enabled bool) func() {
	if !enabled {
		return func() {}
	}

	previous, existed := os.LookupEnv("SHOPLINE_NO_BROWSER")
	_ = os.Setenv("SHOPLINE_NO_BROWSER", "1")

	return func() {
		if existed {
			_ = os.Setenv("SHOPLINE_NO_BROWSER", previous)
			return
		}
		_ = os.Unsetenv("SHOPLINE_NO_BROWSER")
	}
}
