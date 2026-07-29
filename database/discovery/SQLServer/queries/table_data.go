package queries

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func SqlServerTableData(db *gorm.DB, schemaName string, tableName string) ([]map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	schemaName = strings.TrimSpace(schemaName)
	tableName = strings.TrimSpace(tableName)

	if schemaName == "" {
		return nil, fmt.Errorf("schema name cannot be empty")
	}

	if tableName == "" {
		return nil, fmt.Errorf("table name cannot be empty")
	}

	qualifiedTableName := fmt.Sprintf("%s.%s", quoteSqlServerIdentifier(schemaName), quoteSqlServerIdentifier(tableName))

	query := fmt.Sprintf("SELECT * FROM %s", qualifiedTableName)

	var tableData []map[string]interface{}

	result := db.Raw(query).Find(&tableData)
	if result.Error != nil {
		return nil, fmt.Errorf(
			"read table %s: %w",
			qualifiedTableName,
			result.Error,
		)
	}

	return tableData, nil
}

func quoteSqlServerIdentifier(value string) string {
	return "[" + strings.ReplaceAll(value, "]", "]]") + "]"
}
