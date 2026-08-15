package cli

import (
	"blueprint/appinfo"
	"blueprint/connections"
	connCrud "blueprint/connections/crud"
	"blueprint/database/scripting"
	"blueprint/models"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Version(input models.CommandInput) {
	fmt.Printf("%s %s\n", appinfo.Name, appinfo.Version)
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Shows application version",
		Args:    cobra.NoArgs,
		GroupID: "system",
		Run: func(cmd *cobra.Command, args []string) {
			Version(emptyInput())
		},
	}
}

func newCompletionCommand(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:     "completion [bash|zsh|fish|powershell]",
		Short:   "Generate shell completion scripts",
		Args:    cobra.ExactArgs(1),
		GroupID: "system",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletion(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "Lists database connections",
		Args:    cobra.NoArgs,
		GroupID: "connections",
		Run: func(cmd *cobra.Command, args []string) {
			connCrud.List(emptyInput())
		},
	}
}

func newAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "add",
		Short:   "Add a database connection",
		Args:    cobra.NoArgs,
		GroupID: "connections",
		Run: func(cmd *cobra.Command, args []string) {
			connCrud.Add(emptyInput())
		},
	}
}

func newEditCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "edit <connection-name>",
		Short:   "Edit a database connection",
		Args:    cobra.ExactArgs(1),
		GroupID: "connections",
		Run: func(cmd *cobra.Command, args []string) {
			connCrud.Edit(inputFromArgs(args))
		},
	}
}

func newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <connection-name>",
		Short:   "Delete a database connection",
		Args:    cobra.ExactArgs(1),
		GroupID: "connections",
		Run: func(cmd *cobra.Command, args []string) {
			connCrud.Delete(inputFromArgs(args))
		},
	}
}

func newTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "test <connection-name>",
		Short:   "Test the database connection",
		Args:    cobra.ExactArgs(1),
		GroupID: "connections",
		Run: func(cmd *cobra.Command, args []string) {
			connections.Test(inputFromArgs(args))
		},
	}
}

func newScriptCommand() *cobra.Command {
	var dataOnly bool
	var schemaOnly bool

	cmd := &cobra.Command{
		Use:     "script <connection-name>",
		Short:   "Script database objects",
		Args:    cobra.ExactArgs(1),
		GroupID: "database",
		Run: func(cmd *cobra.Command, args []string) {
			scripting.Script(models.CommandInput{
				RawArgs:   args,
				Arguments: args,
				Flags: map[string]bool{
					"data-only":   dataOnly,
					"schema-only": schemaOnly,
				},
			})
		},
	}

	cmd.Flags().BoolVar(&dataOnly, "data-only", false, "Script data only")
	cmd.Flags().BoolVar(&schemaOnly, "schema-only", false, "Script schema only")
	cmd.MarkFlagsMutuallyExclusive("data-only", "schema-only")

	return cmd
}

func emptyInput() models.CommandInput {
	return models.CommandInput{Flags: map[string]bool{}}
}

func inputFromArgs(args []string) models.CommandInput {
	return models.CommandInput{
		RawArgs:   args,
		Arguments: args,
		Flags:     map[string]bool{},
	}
}
