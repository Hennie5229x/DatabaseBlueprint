package discovery

import (
	"blueprint/appinfo"
	"blueprint/cli/spinner"
	"blueprint/connections"
	"blueprint/database"
	discoverymodels "blueprint/database/discovery/models"
	sqlserver "blueprint/database/discovery/schema/SQLServer"
	"blueprint/models"

	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ClearDirectory(path string) error {
	// Remove all files & folders
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	// Recreate the empty directory.
	return os.MkdirAll(path, 0o755)
}

func GetTables(args []string) {
	var Argument string = ""
	if len(args) > 0 {
		Argument = args[0]
	}

	directory := appinfo.ScriptDirectory

	ClearDirectory(directory)

	tableSpinner := spinner.New("Tables", "Discovering tables")

	id, conn := connections.GetConnection(Argument)
	if id == "" || conn == nil {
		return
	}

	db, _ := database.Connect(*conn)

	var tables []discoverymodels.Tables

	// ------------
	// SQL Server
	// ------------
	if conn.Type == models.SqlServer {

		tables = sqlserver.SqlServerTables(*db)

		if len(tables) == 0 {
			tableSpinner.Stop("No tables found")
			return
		}

		var foreignKeyDefinition strings.Builder
		var tableDefinition string = ""

		for index, t := range tables {
			tableSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, len(tables), t.Schema, t.Name))

			tableDefinition = sqlserver.WriteFile_SQLServerTable(*db, t.Schema, t.Name)
			foreignKeyDefinition.WriteString(sqlserver.WriteFile_SQLServerFK(*db, t.Schema, t.Name))

			// Write table files
			tablesPath := filepath.Join(directory, "Tables", fmt.Sprintf("%s.%s", t.Name, "sql"))
			os.MkdirAll(filepath.Dir(tablesPath), 0755)
			err := os.WriteFile(tablesPath, []byte(tableDefinition), 0644)
			if err != nil {
				tableSpinner.Stop("Failed")
				panic(err)
			}

			// Write FK file
			foreignKeyPath := filepath.Join(directory, "ForeignKeys", "ForeignKeys.sql")
			os.MkdirAll(filepath.Dir(foreignKeyPath), 0755)
			err = os.WriteFile(foreignKeyPath, []byte(foreignKeyDefinition.String()), 0644)
			if err != nil {
				tableSpinner.Stop("Failed")
				panic(err)
			}
		}

		tableSpinner.Stop(fmt.Sprintf("%d tables scripted", len(tables)))

	}

}
