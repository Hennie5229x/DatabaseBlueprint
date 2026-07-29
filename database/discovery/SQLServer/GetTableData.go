package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"
	"strings"
	"time"

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
		primaryKeys := queries.SqlServerPrimaryKeys(db, qualifiedTableName)

		tableData, err := queries.SqlServerTableData(db, table.Schema, table.Name)

		if err != nil {
			tableDataSpinner.Stop("Failed")
			panic(err)
		}

		if err := ScriptTableData(directory, table.Schema, table.Name, tableData, columns, primaryKeys); err != nil {
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

func generateInsertStatement(schemaName string, tableName string, row map[string]interface{}, columns []sqlservermodels.Column) string {
	columnNames := make([]string, 0, len(columns))
	values := make([]string, 0, len(columns))
	hasIdentityInsert := false

	for _, column := range columns {
		if shouldSkipInsertColumn(column) {
			continue
		}

		columnNames = append(columnNames, quoteSqlServerIdentifier(column.ColumnName))
		values = append(values, formatSQLValue(row[column.ColumnName], column))

		if column.IsIdentity {
			hasIdentityInsert = true
		}
	}

	targetTable := fmt.Sprintf(
		"%s.%s",
		quoteSqlServerIdentifier(schemaName),
		quoteSqlServerIdentifier(tableName),
	)

	insertStatement := fmt.Sprintf(
		"INSERT INTO %s.%s (%s) VALUES (%s);",
		quoteSqlServerIdentifier(schemaName),
		quoteSqlServerIdentifier(tableName),
		strings.Join(columnNames, ", "),
		strings.Join(values, ", "),
	)

	if hasIdentityInsert {
		return fmt.Sprintf(
			"SET IDENTITY_INSERT %s ON;\n%s\nSET IDENTITY_INSERT %s OFF;",
			targetTable,
			insertStatement,
			targetTable,
		)
	}

	return insertStatement
}

func formatSQLValue(value interface{}, column sqlservermodels.Column) string {
	if value == nil {
		return "NULL"
	}

	dataType := strings.ToLower(column.DataType)

	if strings.EqualFold(dataType, "uniqueidentifier") {
		if raw, ok := value.([]byte); ok && len(raw) == 16 {
			var uniqueIdentifier mssql.UniqueIdentifier
			if err := uniqueIdentifier.Scan(raw); err == nil {
				return "'" + uniqueIdentifier.String() + "'"
			}
		}

		if typedValue, ok := value.(string); ok {
			return "'" + strings.ReplaceAll(typedValue, "'", "''") + "'"
		}
	}

	switch typedValue := value.(type) {
	case bool:
		if typedValue {
			return "1"
		}
		return "0"

	case time.Time:
		return "'" + formatSQLTemporalValue(typedValue, column) + "'"

	case string:
		return formatSQLStringValue(typedValue, dataType)

	case []byte:
		return formatSQLBytesValue(typedValue, column)
	}

	return fmt.Sprintf("%v", value)
}

func shouldSkipInsertColumn(column sqlservermodels.Column) bool {
	dataType := strings.ToLower(column.DataType)

	return column.ComputedDefinition != "" || dataType == "rowversion" || dataType == "timestamp"
}

func formatSQLBytesValue(value []byte, column sqlservermodels.Column) string {
	dataType := strings.ToLower(column.DataType)
	if dataType == "bit" {
		if len(value) > 0 && value[0] == 1 {
			return "1"
		}
		return "0"
	}

	if isNumericDataType(dataType) {
		return string(value)
	}

	if isBinaryDataType(dataType) {
		return "0x" + fmt.Sprintf("%x", value)
	}

	return formatSQLStringValue(string(value), dataType)
}

func formatSQLStringValue(value string, dataType string) string {
	value = strings.ReplaceAll(value, "'", "''")
	if isUnicodeStringDataType(dataType) {
		return "N'" + value + "'"
	}

	return "'" + value + "'"
}

func formatSQLTemporalValue(value time.Time, column sqlservermodels.Column) string {
	dataType := strings.ToLower(column.DataType)
	scale := column.Scale

	switch dataType {
	case "date":
		return value.Format("2006-01-02")
	case "time":
		return value.Format("15:04:05") + formatSQLFraction(value.Nanosecond(), scale)
	case "datetime2":
		return value.Format("2006-01-02 15:04:05") + formatSQLFraction(value.Nanosecond(), scale)
	case "datetimeoffset":
		return value.Format("2006-01-02 15:04:05") + formatSQLFraction(value.Nanosecond(), scale) + value.Format(" -07:00")
	default:
		return value.Format("2006-01-02 15:04:05.000")
	}
}

func formatSQLFraction(nanosecond int, scale int) string {
	if scale <= 0 {
		return ""
	}

	if scale > 7 {
		scale = 7
	}

	fraction := fmt.Sprintf("%09d", nanosecond)
	return "." + fraction[:scale]
}

func isUnicodeStringDataType(dataType string) bool {
	switch dataType {
	case "nchar", "nvarchar", "ntext", "xml":
		return true
	default:
		return false
	}
}

func isBinaryDataType(dataType string) bool {
	switch dataType {
	case "binary", "varbinary", "image":
		return true
	default:
		return false
	}
}

func isNumericDataType(dataType string) bool {
	switch dataType {
	case "tinyint", "smallint", "int", "bigint", "decimal", "numeric", "float", "real", "money", "smallmoney":
		return true
	default:
		return false
	}
}

func tableDataFileName(row map[string]interface{}, primaryKeys []sqlservermodels.PrimaryKeyColumn, columnsByName map[string]sqlservermodels.Column, fallbackIndex int) string {
	if len(primaryKeys) == 0 {
		return fmt.Sprintf("%d", fallbackIndex)
	}

	parts := make([]string, 0, len(primaryKeys))
	for _, primaryKey := range primaryKeys {
		column, ok := columnsByName[primaryKey.ColumnName]
		if !ok {
			return fmt.Sprintf("%d", fallbackIndex)
		}

		value, ok := row[primaryKey.ColumnName]
		if !ok || value == nil {
			return fmt.Sprintf("%d", fallbackIndex)
		}

		parts = append(parts, sanitizeFileNamePart(formatColumnValue(value, column.DataType)))
	}

	return strings.Join(parts, "-")
}

func columnsByName(columns []sqlservermodels.Column) map[string]sqlservermodels.Column {
	columnsByName := make(map[string]sqlservermodels.Column, len(columns))
	for _, column := range columns {
		columnsByName[column.ColumnName] = column
	}

	return columnsByName
}

func sanitizeFileNamePart(value string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	value = strings.TrimSpace(replacer.Replace(value))
	if value == "" || value == "." || value == ".." {
		return "_"
	}

	return value
}
