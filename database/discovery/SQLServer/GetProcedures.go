package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"

	"gorm.io/gorm"
)

func Procedures(db *gorm.DB, directory string) {
	var procedure []sqlservermodels.Procedures

	procedureSpinner := spinner.New("Procedures", "Discovering procedures")
	procedure = queries.SqlServerProcedures(db)

	if len(procedure) == 0 {
		procedureSpinner.Stop("No procedures found")
		return
	}
	err := ScriptProcedures(*db, directory, procedure, func(index int, total int, procedure sqlservermodels.Procedures) {
		procedureSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, total, procedure.Schema, procedure.Name))
	})
	if err != nil {
		procedureSpinner.Stop("Failed")
		panic(err)
	}
	procedureSpinner.Stop(fmt.Sprintf("%d procedures scripted", len(procedure)))
}
