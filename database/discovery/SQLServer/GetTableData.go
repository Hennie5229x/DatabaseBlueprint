package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"
	"strings"

	mssql "github.com/microsoft/go-mssqldb"
	"gorm.io/gorm"
)

func TableData(db *gorm.DB, directory string) {
	tables := queries.SqlServerTables(*db)
	tableDataSpinner := spinner.New("Table Data", "Discovering table data")

	if len(tables) == 0 {
		tableDataSpinner.Stop("No tables found")
		return
	}

	for index, table := range tables {
		tableDataSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, len(tables), table.Schema, table.Name))

		qualifiedTableName := fmt.Sprintf("%s.%s", table.Schema, table.Name)
		columns := queries.SqlServerColumns(db, qualifiedTableName)

		tableData, err := queries.SqlServerTableData(db, table.Schema, table.Name)

		if err != nil {
			tableDataSpinner.Stop("Failed")
			panic(err)
		}

		if err := ScriptTableData(directory, table.Name, tableData, columns); err != nil {
			tableDataSpinner.Stop("Failed")
			panic(err)
		}
	}

	tableDataSpinner.Stop(fmt.Sprintf("%d tables data scripted", len(tables)))
}

func formatTableData(rows []map[string]interface{}, columns []sqlservermodels.Column) string {
	formattedRows := make([]string, 0, len(rows))
	for _, row := range rows {
		parts := make([]string, 0, len(columns))
		for _, column := range columns {
			value := row[column.ColumnName]
			parts = append(parts, fmt.Sprintf("%s:%s", column.ColumnName, formatColumnValue(value, column.DataType)))
		}

		formattedRows = append(formattedRows, "map["+strings.Join(parts, " ")+"]")
	}

	return "[" + strings.Join(formattedRows, " ") + "]"
}

func formatColumnValue(value interface{}, dataType string) string {
	if value == nil {
		return "<nil>"
	}

	if strings.EqualFold(dataType, "uniqueidentifier") {
		if raw, ok := value.([]byte); ok && len(raw) == 16 {
			var uniqueIdentifier mssql.UniqueIdentifier
			if err := uniqueIdentifier.Scan(raw); err == nil {
				return uniqueIdentifier.String()
			}
		}
	}

	return fmt.Sprintf("%v", value)
}
