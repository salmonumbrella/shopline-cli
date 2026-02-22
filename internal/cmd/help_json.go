package cmd

import (
	"encoding/json"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type helpFlag struct {
	Name       string `json:"name"`
	Shorthand  string `json:"shorthand,omitempty"`
	Usage      string `json:"usage"`
	Type       string `json:"type"`
	Default    string `json:"default,omitempty"`
	Required   bool   `json:"required,omitempty"`
	Deprecated bool   `json:"deprecated,omitempty"`
	Persistent bool   `json:"persistent,omitempty"`
}

type helpCommand struct {
	Name        string        `json:"name"`
	Use         string        `json:"use"`
	Short       string        `json:"short,omitempty"`
	Long        string        `json:"long,omitempty"`
	Aliases     []string      `json:"aliases,omitempty"`
	Example     string        `json:"example,omitempty"`
	Deprecated  string        `json:"deprecated,omitempty"`
	Flags       []helpFlag    `json:"flags,omitempty"`
	Subcommands []helpCommand `json:"subcommands,omitempty"`
	// Commands is an alias of subcommands for agent-friendliness (some tooling expects "commands").
	Commands []helpCommand `json:"commands,omitempty"`
}

var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Help about any command",
	RunE: func(cmd *cobra.Command, args []string) error {
		target := rootCmd
		if len(args) > 0 {
			found, _, err := rootCmd.Find(args)
			if err != nil {
				return err
			}
			target = found
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		deep, _ := cmd.Flags().GetBool("deep")
		outputFormat, _ := cmd.Flags().GetString("output")
		if jsonOut || outputFormat == "json" {
			return printHelpJSON(cmd, target, deep)
		}

		return target.Help()
	},
}

func init() {
	helpCmd.Flags().Bool("json", false, "Output help as JSON")
	helpCmd.Flags().Bool("deep", false, "For JSON help, include nested subcommands recursively")
}

func printHelpJSON(cmd *cobra.Command, target *cobra.Command, deep bool) error {
	info := buildHelpCommand(target, deep)
	enc := json.NewEncoder(cmd.OutOrStdout())
	if getRequestedOutputMode(cmd, "json") == "json" {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(info)
}

func buildHelpCommand(cmd *cobra.Command, deep bool) helpCommand {
	info := helpCommand{
		Name:       cmd.Name(),
		Use:        cmd.Use,
		Short:      cmd.Short,
		Long:       cmd.Long,
		Aliases:    cmd.Aliases,
		Example:    cmd.Example,
		Deprecated: cmd.Deprecated,
	}

	info.Flags = collectHelpFlags(cmd)
	for _, sub := range cmd.Commands() {
		if sub.Hidden {
			continue
		}
		if deep {
			info.Subcommands = append(info.Subcommands, buildHelpCommand(sub, deep))
		} else {
			info.Subcommands = append(info.Subcommands, helpCommand{
				Name:    sub.Name(),
				Use:     sub.Use,
				Short:   sub.Short,
				Aliases: append([]string{}, sub.Aliases...),
			})
		}
	}

	sort.Slice(info.Flags, func(i, j int) bool { return info.Flags[i].Name < info.Flags[j].Name })
	sort.Slice(info.Subcommands, func(i, j int) bool { return info.Subcommands[i].Name < info.Subcommands[j].Name })
	info.Commands = append([]helpCommand{}, info.Subcommands...)
	return info
}

func collectHelpFlags(cmd *cobra.Command) []helpFlag {
	flags := map[string]helpFlag{}
	addFlag := func(f *pflag.Flag, persistent bool) {
		if f == nil {
			return
		}
		flags[f.Name] = helpFlag{
			Name:       f.Name,
			Shorthand:  f.Shorthand,
			Usage:      f.Usage,
			Type:       f.Value.Type(),
			Default:    f.DefValue,
			Required:   isFlagRequired(f),
			Deprecated: f.Deprecated != "",
			Persistent: persistent,
		}
	}

	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) { addFlag(f, false) })
	cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) { addFlag(f, true) })

	result := make([]helpFlag, 0, len(flags))
	for _, f := range flags {
		result = append(result, f)
	}
	return result
}

func isFlagRequired(f *pflag.Flag) bool {
	if f == nil || f.Annotations == nil {
		return false
	}
	vals, ok := f.Annotations[cobra.BashCompOneRequiredFlag]
	if !ok || len(vals) == 0 {
		return false
	}
	return vals[0] == "true"
}
