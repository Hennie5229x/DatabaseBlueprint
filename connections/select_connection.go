package connections

import (
	"blueprint/models"
	"fmt"
	"net"

	"github.com/charmbracelet/huh"
)

func ConnectionsSelectForm(config models.ConnectionsFile) string {

	var (
		id string
	)

	options := make([]huh.Option[string], 0, len(config.Connections))

	var optionString string = ""
	for _, db := range config.Connections {
		host := db.Server
		if db.Port != "" {
			host = net.JoinHostPort(db.Server, db.Port)
		}
		optionString = PadRight(db.Name, 20) + "🖥  " + PadRight(host, 17) + " ➔ 🛢  " + db.Database
		options = append(options,
			huh.NewOption(optionString, db.Id),
		)
	}

	form := huh.NewForm(
		huh.NewGroup(

			huh.NewSelect[string]().
				Title("Test a conenction").
				Options(options...).
				Value(&id),
		),
	)

	if err := form.Run(); err != nil {
		panic(err)
	}

	return id
}

func GetConnection(Name string) (string, *models.Connection) {
	if Name == "" {
		fmt.Println("⚠️  Connection name required")
		return "", nil
	}
	/*
		config, _ := LoadConnections()

		id := ConnectionsSelectForm(*config)

		conn := *FindConnectionById(config, id)

		return id, &conn
	*/
	id, conn := FindConnectionByName(Name)
	if conn == nil || id == "" {
		fmt.Println("❌ Cannot find connection", "'"+Name+"'")
		return "", nil
	}
	return id, conn
}
