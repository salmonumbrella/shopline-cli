package main

import (
	"os"

	"github.com/salmonumbrella/shopline-cli/internal/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := cmd.Execute(version, commit, date); err != nil {
		os.Exit(cmd.GetExitCode(err))
	}
}
