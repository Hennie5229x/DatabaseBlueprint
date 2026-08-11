package cli

import (
	"blueprint/models"
	"fmt"
	"os"
	"slices"
	"strings"
)

// --------------------------
// Execute CMD Commands
// --------------------------
func Execute() {
	if len(os.Args) < 2 {
		PrintBanner()
		return
	}

	switch os.Args[1] {
	case "help", "h":
		Help()
		return

	default:
		for _, cmd := range Commands {
			if cmd.Name == os.Args[1] {
				executeCommand(cmd, os.Args[2:])
				return
			}
		}

		fmt.Println("Unknown command\nTry 'help' for list of commands")
	}
}

func executeCommand(cmd models.Commands, args []string) {
	if len(args) > 0 {
		// Handle help flags for normal commands
		if isHelpArg(args[0]) {
			cmd.PrintCommandHelp()
			return
		}

		// Handle normal Commands
		for _, subcmd := range cmd.SubCommands {
			if subcmd.Name == args[0] {
				executeCommand(subcmd, args[1:])
				return
			}
		}
	}

	if cmd.Run != nil {
		cmd.Run(args)
		return
	}

	fmt.Printf("'%s' requires a subcommand\n", cmd.Name)
}

// Help flag for normal commands
func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

func Help() {
	cat_slices := []string{}
	for _, cats := range Commands {
		cat_slices = append(cat_slices, string(cats.Category))
	}

	// Remove duplicates Categories
	distinct_cmds := slices.Compact(cat_slices)
	fmt.Println("All Commands")
	for _, cats := range distinct_cmds {
		fmt.Println()
		fmt.Println("----------", strings.ToUpper(cats), "----------")

		for _, cmd := range Commands {
			if cats == string(cmd.Category) {
				if cats == string(models.System) { //Append Help manually
					fmt.Printf("%-20s %s\n", "help", "(List of all commands)")
				}
				fmt.Printf("%-20s %s\n", cmd.Name, "("+cmd.Description+")")

				printSubCommands(cmd, "")
			}
		}
	}
	fmt.Println()
}

const (
	branch = "├── "
	last   = "└── "
	pipe   = "│   "
	space  = "    "
)

func printSubCommands(parentCmd models.Commands, currentIndent string) {
	count := len(parentCmd.SubCommands)

	for i, subcmd := range parentCmd.SubCommands {
		isLast := i == count-1

		prefix := branch
		nextIndent := currentIndent + pipe

		if isLast {
			prefix = last
			nextIndent = currentIndent + space
		}

		name := currentIndent + prefix + subcmd.Name
		fmt.Printf("%-20s (%s)\n", name, subcmd.Description)

		printSubCommands(subcmd, nextIndent)
	}
}
