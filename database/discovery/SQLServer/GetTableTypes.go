package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"

	"gorm.io/gorm"
)

func TableTypes(db *gorm.DB, directory string) {
	var tableTypes []sqlservermodels.UserDefinedTableType

	tableTypeSpinner := spinner.New("Table Types", "Discovering table types")
	tableTypes = queries.SqlServerUserDefinedTableTypes(db)

	if len(tableTypes) == 0 {
		tableTypeSpinner.Stop("No table types found")
		return
	}
	err := ScriptTableTypes(*db, directory, tableTypes, func(index int, total int, tableType sqlservermodels.UserDefinedTableType) {
		tableTypeSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, total, tableType.SchemaName, tableType.TypeName))
	})
	if err != nil {
		tableTypeSpinner.Stop("Failed")
		panic(err)
	}
	tableTypeSpinner.Stop(fmt.Sprintf("%d table types scripted", len(tableTypes)))
}

func TableTypeColumns(db *gorm.DB, tableType sqlservermodels.UserDefinedTableType) []sqlservermodels.UserDefinedTableTypeColumn {
	return queries.SqlServerUserDefinedTableTypeColumns(db, tableType.SchemaName, tableType.TypeName)
}

func TableTypeKeys(db *gorm.DB, tableType sqlservermodels.UserDefinedTableType) []sqlservermodels.UserDefinedTableTypeKeyColumn {
	return queries.SqlServerUserDefinedTableTypeKeys(db, tableType.SchemaName, tableType.TypeName)
}

func TableTypeChecks(db *gorm.DB, tableType sqlservermodels.UserDefinedTableType) []sqlservermodels.UserDefinedTableTypeCheckConstraint {
	return queries.SqlServerUserDefinedTableTypeChecks(db, tableType.SchemaName, tableType.TypeName)
}

func TableTypeIndexes(db *gorm.DB, tableType sqlservermodels.UserDefinedTableType) []sqlservermodels.UserDefinedTableTypeIndexColumn {
	return queries.SqlServerUserDefinedTableTypeIndexes(db, tableType.SchemaName, tableType.TypeName)
}
