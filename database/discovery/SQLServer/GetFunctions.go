package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"

	"gorm.io/gorm"
)

func Functions(db *gorm.DB, directory string) {
	var function []sqlservermodels.Functions

	functionSpinner := spinner.New("Functions", "Discovering functions")
	function = queries.SqlServerFunctions(db)

	if len(function) == 0 {
		functionSpinner.Stop("No functions found")
		return
	}
	err := ScriptFunctions(*db, directory, function, func(index int, total int, function sqlservermodels.Functions) {
		functionSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, total, function.Schema, function.Name))
	})
	if err != nil {
		functionSpinner.Stop("Failed")
		panic(err)
	}
	functionSpinner.Stop(fmt.Sprintf("%d functions scripted", len(function)))
}
