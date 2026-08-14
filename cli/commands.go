package cli

import (
	"blueprint/connections"
	conn_crud "blueprint/connections/crud"
	"fmt"

	"blueprint/database/scripting"
	"blueprint/models"
)

var Commands = []models.Commands{

	// ---- System ----
	{
		Name:        "version",
		Description: "Shows application version",
		Run:         Version,
		Category:    models.System,
	},
	{
		Name:        "completion",
		Description: "Generate shell completion scripts",
		Run:         Completion,
		Category:    models.System,
		Hide:        true,
	},
	// ---- Connections ----
	{
		Name:        "list",
		Description: "Lists database connections",
		Run:         conn_crud.List,
		Category:    models.Connections,
	},
	{
		Name:        "test",
		Description: "Test the database connection",
		Run:         connections.Test,
		Category:    models.Connections,

		Usage: models.CommandUsage{
			Arguments: []models.UsageItem{
				{
					Name:        "connection-name",
					Description: "Name of a saved connection",
				},
			},
			Flags: []models.UsageItem{
				{
					Name:        "-h, --help",
					Description: "Show usage and examples",
				},
			},
			Examples: []string{
				"blue test Production",
			},
		},
	},
	{
		Name:        "add",
		Description: "Add a database connection",
		Run:         conn_crud.Add,
		Category:    models.Connections,
	},
	{
		Name:        "edit",
		Description: "Edit a database connection",
		Run:         conn_crud.Edit,
		Category:    models.Connections,

		Usage: models.CommandUsage{
			Arguments: []models.UsageItem{
				{
					Name:        "connection-name",
					Description: "Name of a saved connection",
				},
			},
			Flags: []models.UsageItem{
				{
					Name:        "-h, --help",
					Description: "Show usage and examples",
				},
			},
			Examples: []string{
				"blue edit Production",
			},
		},
	},
	{
		Name:        "delete",
		Description: "Delete a database connection",
		Run:         conn_crud.Delete,
		Category:    models.Connections,

		Usage: models.CommandUsage{
			Arguments: []models.UsageItem{
				{
					Name:        "connection-name",
					Description: "Name of a saved connection",
				},
			},
			Flags: []models.UsageItem{
				{
					Name:        "-h, --help",
					Description: "Show usage and examples",
				},
			},
			Examples: []string{
				"blue delete Production",
			},
		},
	},
	// ---- Database ----
	{
		Name:        "script",
		Description: "Script database objects",
		Run:         scripting.Script,
		Category:    models.Database,

		Usage: models.CommandUsage{
			Arguments: []models.UsageItem{
				{
					Name:        "connection-name",
					Description: "Name of a saved connection",
				},
			},
			Flags: []models.UsageItem{
				{
					Name:        "-h, --help",
					Description: "Show usage and examples",
				},
				{
					Name:        "--data-only",
					Description: "Script data only",
				},
			},
			Examples: []string{
				"blue script Production",
				"blue script Production --data-only",
			},
		},
	},
}

/*
	SubCommands: []models.Commands{
		{
			Name:        "test",
			Description: "Test the database connection",
			Run:         connections.TestConnection, //connections.TestConnection,
		},
		{
			Name:        "add",
			Description: "Add a database connection",
			Run:         connections.AddConnections,
		},
		{
			Name:        "edit",
			Description: "Edit a database connection",
			Run:         List,
		},
		{
			Name:        "delete",
			Description: "Delete a database connection",
			Run:         List,
			SubCommands: []models.Commands{
				{
					Name:        "subsub",
					Description: "sub sub sub sub",
					Run:         List,
				},
			},
		},
	}, */

func testFunc(args []string) {
	fmt.Println("TEST", args)
}
