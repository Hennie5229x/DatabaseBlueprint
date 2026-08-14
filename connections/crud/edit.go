package crud

import (
	"blueprint/connections"
	"blueprint/models"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/google/uuid"
)

func Edit(input models.CommandInput) {

	var Argument string = ""
	if len(input.Arguments) > 0 {
		Argument = input.Arguments[0]
	}

	id, conn := connections.GetConnection(Argument)

	if id == "" || conn == nil {
		//fmt.Println("❌ Failed to connect!")
		return
	}

	options := make([]huh.Option[models.DatabaseType], 0, len(models.DatabaseTypes))

	for _, db := range models.DatabaseTypes {
		options = append(options,
			huh.NewOption(db.Name, db.Value),
		)
	}

	var (
		name     string              = conn.Name
		server   string              = conn.Server
		port     string              = conn.Port
		database string              = conn.Database
		username string              = conn.User
		password string              = conn.Password
		dbType   models.DatabaseType = conn.Type
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

	config, err := connections.LoadConnections()
	if err != nil {
		panic(err)
	}

	updated := models.Connection{
		Id:       uuid.NewString(),
		Type:     dbType,
		Name:     name,
		Server:   server,
		Port:     port,
		Database: database,
		User:     username,
		Password: password,
	}

	if err := connections.UpdateConnection(config, id, updated); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("✅ Connection updated!")
}
