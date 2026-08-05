package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"

	"gorm.io/gorm"
)

func UserDefinedTypes(db *gorm.DB, directory string) {
	userDefinedTypes := queries.SqlServerUserDefinedTypes(db)
	userDefinedTypeSpinner := spinner.New("Data Types", "Discovering user-defined types")

	if len(userDefinedTypes) == 0 {
		userDefinedTypeSpinner.Stop("No data types found")
		return
	}

	err := ScriptUserDefinedTypes(directory, userDefinedTypes, func(index int, total int, userDefinedType sqlservermodels.UserDefinedType) {
		userDefinedTypeSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, total, userDefinedType.SchemaName, userDefinedType.TypeName))
	})
	if err != nil {
		userDefinedTypeSpinner.Stop("Failed")
		panic(err)
	}

	userDefinedTypeSpinner.Stop(fmt.Sprintf("%d data types scripted", len(userDefinedTypes)))
}
