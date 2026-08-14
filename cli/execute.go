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
		input, ok := parseCommandInput(cmd, args)
		if !ok {
			return
		}

		if input.Flags["help"] {
			cmd.PrintCommandHelp()
			return
		}

		cmd.Run(input)
		return
	}

	fmt.Printf("'%s' requires a subcommand\n", cmd.Name)
}

func parseCommandInput(cmd models.Commands, args []string) (models.CommandInput, bool) {
	input := models.CommandInput{
		RawArgs:   args,
		Arguments: []string{},
		Flags:     map[string]bool{},
	}

	flagLookup := map[string]string{}
	for _, flag := range cmd.Usage.Flags {
		for _, name := range flag.Names {
			flagLookup[name] = flag.Key
		}
	}

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flagKey, exists := flagLookup[arg]
			if !exists {
				fmt.Printf("Unknown flag '%s' for command '%s'\n", arg, cmd.Name)
				return input, false
			}

			input.Flags[flagKey] = true
			continue
		}

		input.Arguments = append(input.Arguments, arg)
	}

	return input, true
}

// Help flag for normal commands
func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

// --------------------------
// Global Help command
// --------------------------
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
			if cats == string(cmd.Category) && !cmd.Hide {
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

// --------------------------
// Auto Complete Commands & Flags
// --------------------------
func autoCompleteCommands(args []string) {
	validArguments := []string{}
	validFlags := []string{}

	current := ""
	path := []string{}

	if len(args) > 0 {
		current = args[len(args)-1]
		path = args[:len(args)-1]
	}

	// Flag Detection on Command
	current_flag := ""
	if strings.HasPrefix(current, "-") {
		current_flag = current
	}

	//fmt.Println("CUR:", current)
	//fmt.Println("PATH:", path)
	//fmt.Println(last_command)
	//fmt.Println("FLAG:", current_flag)

	commandsAtDepth := resolveCommandPath(Commands, path)
	if len(path) == 0 {
		commandsAtDepth = append([]models.Commands{{Name: "help"}}, commandsAtDepth...)
	}

	flagCommand := resolveCommand(Commands, path)

	var argLen int = len(current)
	var cmdTrimmed string = ""

	if current_flag != "" && flagCommand != nil {
		for _, f := range flagCommand.Usage.Flags {
			for _, flagName := range f.Names {
				if len(current_flag) <= len(flagName) {
					flagTrimmed := flagName[0:len(current_flag)] // Prefix test
					if flagTrimmed == current_flag {
						validFlags = append(validFlags, flagName)
					}
				}
			}
		}
	}

	for _, c := range commandsAtDepth {
		// Build valid Command list
		if argLen <= len(c.Name) {
			cmdTrimmed = string(c.Name[0:argLen]) // Prefix test
			if cmdTrimmed == current {
				validArguments = append(validArguments, c.Name)
			}
		}
	}

	if len(validFlags) > 0 {
		for _, validFlag := range validFlags {
			fmt.Println(validFlag)
		}
	} else if len(validArguments) > 0 {
		for _, validArg := range validArguments {
			fmt.Println(validArg)
		}
	}

	//fmt.Printf("Match(es) found for '%s' => '%s'\n", path, validArguments)

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

func resolveCommand(commands []models.Commands, path []string) *models.Commands {
	if len(path) == 0 {
		return nil
	}

	for _, cmd := range commands {
		if cmd.Name == path[0] {
			if len(path) == 1 {
				return &cmd
			}

			return resolveCommand(cmd.SubCommands, path[1:])
		}
	}

	return nil
}
