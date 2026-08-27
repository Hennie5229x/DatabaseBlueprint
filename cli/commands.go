package cli

import (
	"blueprint/appinfo"
	"blueprint/connections"
	connCrud "blueprint/connections/crud"
	"blueprint/database/creating"
	"blueprint/database/scripting"
	"blueprint/models"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Version(input models.CommandInput) {
	fmt.Printf("%s %s\n", appinfo.Name, appinfo.Version)
}

func VersionCommand() *cobra.Command {
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

func CompletionCommand(rootCmd *cobra.Command) *cobra.Command {
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

func ListCommand() *cobra.Command {
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

func AddCommand() *cobra.Command {
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

func EditCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "edit <connection-name>",
		Short:             "Edit a database connection",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: connectionNameCompletion,
		GroupID:           "connections",
		Run: func(cmd *cobra.Command, args []string) {
			connCrud.Edit(inputFromArgs(args))
		},
	}
}

func DeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "delete <connection-name>",
		Short:             "Delete a database connection",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: connectionNameCompletion,
		GroupID:           "connections",
		Run: func(cmd *cobra.Command, args []string) {
			connCrud.Delete(inputFromArgs(args))
		},
	}
}

func TestCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "test <connection-name>",
		Short:             "Test the database connection",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: connectionNameCompletion,
		GroupID:           "connections",
		Run: func(cmd *cobra.Command, args []string) {
			connections.Test(inputFromArgs(args))
		},
	}
}

func ScriptCommand() *cobra.Command {
	var dataOnly bool
	var schemaOnly bool

	var server string
	var port string
	var database string
	var user string
	var password string

	cmd := &cobra.Command{
		Use:               "script <connection-name>",
		Short:             "Script database objects",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: connectionNameCompletion,
		GroupID:           "database",
		Run: func(cmd *cobra.Command, args []string) {
			scripting.Script(models.CommandInput{
				RawArgs:   args,
				Arguments: args,
				StringFlags: map[string]string{
					"server":   server,
					"port":     port,
					"database": database,
					"user":     user,
					"password": password,
				},
				Flags: map[string]bool{
					"data-only":   dataOnly,
					"schema-only": schemaOnly,
				},
			})
		},
	}

	cmd.Flags().BoolVar(&dataOnly, "data-only", false, "Script data only")
	cmd.Flags().BoolVar(&schemaOnly, "schema-only", false, "Script schema only")

	cmd.Flags().StringVarP(&server, "server", "s", "", "Server override value")
	cmd.Flags().StringVarP(&port, "port", "", "", "Port override value")
	cmd.Flags().StringVarP(&database, "database", "d", "", "Database override value")
	cmd.Flags().StringVarP(&user, "user", "u", "", "User override value")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password override value")

	cmd.MarkFlagsMutuallyExclusive("data-only", "schema-only")

	return cmd
}

func CreateCommand() *cobra.Command {
	var server string
	var port string
	var database string
	var user string
	var password string

	cmd := &cobra.Command{
		Use:               "create <connection-name>",
		Short:             "Create database from exported scripts",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: connectionNameCompletion,
		GroupID:           "database",
		Run: func(cmd *cobra.Command, args []string) {
			creating.Create(models.CommandInput{
				RawArgs:   args,
				Arguments: args,
				StringFlags: map[string]string{
					"server":   server,
					"port":     port,
					"database": database,
					"user":     user,
					"password": password,
				},
				Flags: map[string]bool{},
			})
		},
	}

	cmd.Flags().StringVarP(&server, "server", "s", "", "Server override value")
	cmd.Flags().StringVarP(&port, "port", "", "", "Port override value")
	cmd.Flags().StringVarP(&database, "database", "d", "", "Database override value")
	cmd.Flags().StringVarP(&user, "user", "u", "", "User override value")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password override value")

	return cmd
}

func emptyInput() models.CommandInput {
	return models.CommandInput{Flags: map[string]bool{}, StringFlags: map[string]string{}}
}

func inputFromArgs(args []string) models.CommandInput {
	return models.CommandInput{
		RawArgs:     args,
		Arguments:   args,
		Flags:       map[string]bool{},
		StringFlags: map[string]string{},
	}
}
