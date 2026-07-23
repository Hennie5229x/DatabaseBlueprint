package crud

import (
	"blueprint/connections"
	"fmt"
	"slices"

	"github.com/charmbracelet/huh"
)

func Delete(args []string) {

	var Argument string = ""
	if len(args) > 0 {
		Argument = args[0]
	}

	id, conn := connections.GetConnection(Argument)
	if id == "" || conn == nil {
		//fmt.Println("❌ Failed to connect!")
		return
	}

	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Are you sure you want to delete this Connection?").
				Options(
					huh.NewOption("No", "N"),
					huh.NewOption("Yes", "Y"),
				).
				Value(&choice),
		))

	if err := form.Run(); err != nil {
		return
	}

	if choice == "Y" {

		config, err := connections.LoadConnections()
		if err != nil {
			panic(err)
		}

		for i := range config.Connections {
			if id == config.Connections[i].Id {
				config.Connections = slices.Delete(config.Connections, i, i+1)
				break
			}
		}

		// Write to config.json file
		if err := connections.SaveConnections(config); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("✅ Connection deleted.")
	}

}
