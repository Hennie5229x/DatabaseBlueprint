package cli

import (
	"blueprint/connections"
	conn_crud "blueprint/connections/crud"

	"blueprint/database/discovery"
	"blueprint/models"
)

var Commands = []models.Commands{
	{
		Name:        "version",
		Description: "Shows application version",
		Run:         Version,
		Category:    models.System,
	},
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
	},
	{
		Name:        "delete",
		Description: "Delete a database connection",
		Run:         conn_crud.Delete,
		Category:    models.Connections,
	},
	{
		Name:        "tables",
		Description: "List all Tables",
		Run:         discovery.GetTables,
		Category:    models.Database,
	},

	/*{
		Name:        "table",
		Description: "",
		Category:    models.Database,
		SubCommands: []models.Commands{
			{
				Name:        "columns",
				Description: "List table columns",
				Run:         discovery.GetColumns,
			},
		},
	},
	*/
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
