package cli

import (
	"blueprint/appinfo"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	if err := newRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	var showVersion bool

	rootCmd := &cobra.Command{
		Use:          appinfo.CLIName,
		Short:        appinfo.Name,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				Version(emptyInput())
				return
			}

			PrintBanner()
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Show version")
	rootCmd.AddGroup(
		&cobra.Group{ID: "system", Title: "System Commands"},
		&cobra.Group{ID: "connections", Title: "Connection Commands"},
		&cobra.Group{ID: "database", Title: "Database Commands"},
	)
	rootCmd.AddCommand(
		newVersionCommand(),
		newCompletionCommand(rootCmd),
		newListCommand(),
		newAddCommand(),
		newEditCommand(),
		newDeleteCommand(),
		newTestCommand(),
		newScriptCommand(),
	)

	return rootCmd
}
