package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	cmd "github.com/idlab-discover/AIBoMGen-cli/cmd/aibomgen-cli"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

// Version is set at build time
var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	if err := fang.Execute(
		context.Background(),
		cmd.GetRootCmd(),
		fang.WithColorSchemeFunc(ui.FangColorScheme),
	); err != nil {
		os.Exit(1)
	}
}
