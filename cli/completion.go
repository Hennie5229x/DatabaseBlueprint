package cli

import (
	"blueprint/connections"
	"strings"

	"github.com/spf13/cobra"
)

func connectionNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	config, err := connections.LoadConnections()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	partial := strings.ToLower(toComplete)
	completion := make([]string, 0, len(config.Connections))
	for _, connection := range config.Connections {
		if strings.HasPrefix(strings.ToLower(connection.Name), partial) {
			completion = append(completion, connection.Name)
		}
	}

	return completion, cobra.ShellCompDirectiveNoFileComp
}
