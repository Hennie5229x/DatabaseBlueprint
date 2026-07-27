package sqlserver

import (
	"blueprint/cli/spinner"
	sqlserver_models "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	discoverymodels "blueprint/database/discovery/models"

	"fmt"

	"gorm.io/gorm"
)

func Tables(db *gorm.DB, directory string) {

	var tables []discoverymodels.Tables

	tableSpinner := spinner.New("Tables", "Discovering tables")
	tables = queries.SqlServerTables(*db)

	if len(tables) == 0 {
		tableSpinner.Stop("No tables found")
		return
	}
	err := ScriptTables(*db, directory, tables, func(index int, total int, table discoverymodels.Tables) {
		tableSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, total, table.Schema, table.Name))
	})
	if err != nil {
		tableSpinner.Stop("Failed")
		panic(err)
	}
	tableSpinner.Stop(fmt.Sprintf("%d tables scripted", len(tables)))
}

func ForeignKeys(db *gorm.DB, directory string) {

	var fkeys []sqlserver_models.ForeignKeyColumn

	foreignKeySpinner := spinner.New("Foreign Keys", "Discovering foreign key")
	fkeys = queries.SqlServerForeignKeys(db)

	if len(fkeys) == 0 {
		foreignKeySpinner.Stop("No tables found")
		return
	}
	err := ScriptForeignKeys(*db, directory, fkeys, func(index int, total int, fk sqlserver_models.ForeignKeyColumn) {
		foreignKeySpinner.Update(fmt.Sprintf("[%d/%d] %s", index+1, total, fk.ForeignKeyName))
	})
	if err != nil {
		foreignKeySpinner.Stop("Failed")
		panic(err)
	}
	foreignKeySpinner.Stop(fmt.Sprintf("%d foreign keys scripted", len(fkeys)))
}
