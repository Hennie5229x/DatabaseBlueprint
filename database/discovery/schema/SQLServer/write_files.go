package sqlserver

import (
	"fmt"

	"gorm.io/gorm"
)

func WriteFile_SQLServerTable(db gorm.DB, schema string, table string) string {

	var tableDefinition string = ""

	// ---------------------
	// Table Sections
	//----------------------
	tableName := fmt.Sprintf("%s.%s", schema, table)
	columns := SqlServerColumns(&db, tableName)
	defaultConstraints := SqlServerDefaultConstraints(&db, tableName)
	primaryKeys := SqlServerPrimaryKeys(&db, tableName)
	uniqueConstraints := SqlServerUniqueConstraints(&db, tableName)
	checkConstraints := SqlServerCheckConstraints(&db, tableName)
	indexes := SqlServerIndexes(&db, tableName)

	createTableSQL := GenerateCreateTable(
		schema,
		table,
		columns,
		defaultConstraints,
		primaryKeys,
		uniqueConstraints,
		checkConstraints,
	)
	createIndexesSQL := GenerateCreateIndexes(schema, table, indexes)

	//fmt.Println(createTableSQL)
	tableDefinition += createTableSQL

	if createIndexesSQL != "" {
		tableDefinition += fmt.Sprintln("")
		//fmt.Println(createIndexesSQL)
		tableDefinition += createIndexesSQL
	}

	return tableDefinition

}

func WriteFile_SQLServerFK(db gorm.DB, schema string, table string) string {

	var foreignKeyDefinition string = ""

	tableName := fmt.Sprintf("%s.%s", schema, table)

	var sqlServerForeignKeyStatements []string
	foreignKeys := SqlServerForeignKeys(&db, tableName)

	sqlServerForeignKeyStatements = append(
		sqlServerForeignKeyStatements,
		GenerateForeignKeys(schema, table, foreignKeys)...,
	)

	for _, foreignKeySQL := range sqlServerForeignKeyStatements {
		foreignKeyDefinition += foreignKeySQL
		foreignKeyDefinition += fmt.Sprintln("")
	}
	return foreignKeyDefinition
}
