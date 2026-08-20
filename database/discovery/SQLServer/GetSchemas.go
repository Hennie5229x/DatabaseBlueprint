package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"

	"gorm.io/gorm"
)

func Schemas(db *gorm.DB, directory string) {
	var schemas []sqlservermodels.Schemas

	schemasSpinner := spinner.New("Schemas", "Discovering schemas")
	schemas = queries.SqlServerSchemas(db)

	if len(schemas) == 0 {
		schemasSpinner.Stop("No schemas found")
		return
	}
	err := ScriptSchemas(*db, directory, schemas, func(index int, total int, schemas sqlservermodels.Schemas) {
		schemasSpinner.Update(fmt.Sprintf("[%d/%d] %s", index+1, total, schemas.Name))
	})
	if err != nil {
		schemasSpinner.Stop("Failed")
		panic(err)
	}
	schemasSpinner.Stop(fmt.Sprintf("%d schemas scripted", len(schemas)))
}
