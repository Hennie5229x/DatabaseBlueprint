package crud

import (
	"blueprint/connections"
	"blueprint/models"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/google/uuid"
)

func NameExist(Name string, config models.ConnectionsFile) bool {

	var Exsists bool = false

	for i := range config.Connections {
		if strings.EqualFold(config.Connections[i].Name, Name) { // Case insensitive match
			Exsists = true
		}
	}

	return Exsists
}

func Add(args []string) {
	options := make([]huh.Option[models.DatabaseType], 0, len(models.DatabaseTypes))

	for _, db := range models.DatabaseTypes {
		options = append(options,
			huh.NewOption(db.Name, db.Value),
		)
	}

	var (
		name     string
		server   string
		port     string
		database string
		username string
		password string
		dbType   models.DatabaseType
	)

	// Input Form
	form := huh.NewForm(
		huh.NewGroup(

			huh.NewInput().
				Title("Connection Name").
				Value(&name),

			huh.NewSelect[models.DatabaseType]().
				Title("Database Type").
				Options(options...).
				Value(&dbType),

			huh.NewInput().
				Title("Server").
				Value(&server),

			huh.NewInput().
				Title("Port").
				Value(&port),

			huh.NewInput().
				Title("Database").
				Value(&database),

			huh.NewInput().
				Title("Username").
				Value(&username),

			huh.NewInput().
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&password),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	//fmt.Println()
	//fmt.Println("Connection Summary")
	//fmt.Println("------------------")
	//fmt.Printf("%-15s %s\n", "Name:", name)
	//fmt.Printf("%-15s %s\n", "Type:", dbType)
	//fmt.Printf("%-15s %s\n", "Server:", server)
	//fmt.Printf("%-15s %s\n", "Database:", database)
	//fmt.Printf("%-15s %s\n", "Username:", username)
	//fmt.Printf("%-15s %s\n", "Password:", strings.Repeat("*", len(password)))
	//fmt.Println()

	config, err := connections.LoadConnections()
	if err != nil {
		panic(err)
	}

	if NameExist(name, *config) {
		fmt.Println("❌ Name already exists!", "'"+name+"'")
		return
	}

	connection := models.Connection{
		Id:       uuid.NewString(),
		Type:     dbType,
		Name:     name,
		Server:   server,
		Port:     port,
		Database: database,
		User:     username,
		Password: password,
	}

	config.Connections = append(config.Connections, connection)

	// Write to config.json file
	if err := connections.SaveConnections(config); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("✅ Connection saved.")
}
