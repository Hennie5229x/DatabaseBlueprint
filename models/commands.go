package models

import (
	"blueprint/appinfo"
	"fmt"
)

type Commands struct {
	Name        string
	Description string
	Usage       CommandUsage
	Run         func(args []string)
	Category    CommandType
	Hide        bool

	SubCommands []Commands
}

// Print Help for Commands
func (cmd Commands) PrintCommandHelp() {

	fmt.Printf("Command: %s\n", cmd.Name)
	fmt.Printf("Description: %s\n\n", cmd.Description)

	fmt.Println("Usage:")
	fmt.Printf("  %s %s", appinfo.CLIName, cmd.Name)

	for _, argument := range cmd.Usage.Arguments {
		fmt.Printf(" <%s>", argument.Name)
	}

	if len(cmd.Usage.Flags) > 0 {
		fmt.Print(" [flags]")
	}

	fmt.Println()
	fmt.Println()

	if len(cmd.Usage.Arguments) > 0 {
		fmt.Println("Arguments:")

		for _, argument := range cmd.Usage.Arguments {
			fmt.Printf("  %-20s %s\n", argument.Name, argument.Description)
		}

		fmt.Println()
	}

	if len(cmd.Usage.Flags) > 0 {
		fmt.Println("Flags:")

		for _, flag := range cmd.Usage.Flags {
			fmt.Printf("  %-20s %s\n", flag.Name, flag.Description)
		}

		fmt.Println()
	}

	if len(cmd.Usage.Examples) > 0 {
		fmt.Println("Examples:")

		for _, example := range cmd.Usage.Examples {
			fmt.Printf("  %s\n", example)
		}
	}
}
