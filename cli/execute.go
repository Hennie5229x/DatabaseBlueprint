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
	case "__complete":
		autoCompleteCommands(os.Args[2:])
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

// Auto Complete Commands
func autoCompleteCommands(args []string) {
	validArguments := []string{}

	current := ""
	path := []string{}

	if len(args) > 0 {
		current = args[len(args)-1]
		path = args[:len(args)-1]
	}

	commandsAtDepth := resolveCommandPath(Commands, path)

	var argLen int = len(current)
	var cmdTrimmed string = ""

	for _, c := range commandsAtDepth {
		if argLen <= len(c.Name) {
			cmdTrimmed = string(c.Name[0:argLen]) // Prefix test
			if cmdTrimmed == current {
				validArguments = append(validArguments, c.Name)
			}
		}
	}

	if len(validArguments) > 0 {
		for _, validArg := range validArguments {
			fmt.Println(validArg)
		}
	}

	// fmt.Printf("Match(es) found for '%s' => '%s'\n", path, validArguments)

}

func resolveCommandPath(commands []models.Commands, path []string) []models.Commands {
	if len(path) == 0 {
		return commands
	}

	for _, cmd := range commands {
		if cmd.Name == path[0] {
			return resolveCommandPath(cmd.SubCommands, path[1:])
		}
	}

	return nil
}
