package main

import (
	"fmt"
	"os"

	"github.com/salmonumbrella/shopline-cli/internal/cmd"
	"github.com/salmonumbrella/shopline-cli/internal/env"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := env.LoadOpenClawEnv(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to load ~/.openclaw/.env: %v\n", err)
	}

	if err := cmd.Execute(version, commit, date); err != nil {
		os.Exit(cmd.GetExitCode(err))
	}
}
