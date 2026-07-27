package sqlserver

import (
	sqlserver_models "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	discoverymodels "blueprint/database/discovery/models"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

func BuildTableScript(db gorm.DB, schema string, table string) string {
	tableName := fmt.Sprintf("%s.%s", schema, table)
	columns := queries.SqlServerColumns(&db, tableName)
	defaultConstraints := queries.SqlServerDefaultConstraints(&db, tableName)
	primaryKeys := queries.SqlServerPrimaryKeys(&db, tableName)
	uniqueConstraints := queries.SqlServerUniqueConstraints(&db, tableName)
	checkConstraints := queries.SqlServerCheckConstraints(&db, tableName)
	indexes := queries.SqlServerIndexes(&db, tableName)

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

	tableDefinition := createTableSQL
	if createIndexesSQL != "" {
		tableDefinition += fmt.Sprintln("")
		tableDefinition += createIndexesSQL
	}

	return tableDefinition
}

func BuildForeignKeyScript(foreignKeys []sqlserver_models.ForeignKeyColumn) string {
	foreignKeyStatements := GenerateForeignKeys(foreignKeys)
	if len(foreignKeyStatements) == 0 {
		return ""
	}

	return foreignKeyStatements[0] + fmt.Sprintln("")
}

func ScriptTables(db gorm.DB, directory string, tables []discoverymodels.Tables, onProgress func(index int, total int, table discoverymodels.Tables)) error {

	for index, table := range tables {
		if onProgress != nil {
			onProgress(index, len(tables), table)
		}

		tableDefinition := BuildTableScript(db, table.Schema, table.Name)

		// Tables
		tablesPath := filepath.Join(directory, "Tables", fmt.Sprintf("%s.%s", table.Name, "sql"))
		if err := os.MkdirAll(filepath.Dir(tablesPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(tablesPath, []byte(tableDefinition), 0o644); err != nil {
			return err
		}

	}

	return nil
}

func ScriptForeignKeys(db gorm.DB, directory string, foreignKeys []sqlserver_models.ForeignKeyColumn, onProgress func(index int, total int, fk sqlserver_models.ForeignKeyColumn)) error {
	groupsByID := make(map[int][]sqlserver_models.ForeignKeyColumn)
	orderedIDs := make([]int, 0)
	firstByID := make(map[int]sqlserver_models.ForeignKeyColumn)

	for _, foreignKey := range foreignKeys {
		if _, ok := groupsByID[foreignKey.ForeignKeyObjectID]; !ok {
			orderedIDs = append(orderedIDs, foreignKey.ForeignKeyObjectID)
			firstByID[foreignKey.ForeignKeyObjectID] = foreignKey
		}
		groupsByID[foreignKey.ForeignKeyObjectID] = append(groupsByID[foreignKey.ForeignKeyObjectID], foreignKey)
	}

	for index, foreignKeyObjectID := range orderedIDs {
		fkey := firstByID[foreignKeyObjectID]
		if onProgress != nil {
			onProgress(index, len(orderedIDs), fkey)
		}

		foreignKeyDefinition := BuildForeignKeyScript(groupsByID[foreignKeyObjectID])

		// Foreign Keys
		foreignKeyPath := filepath.Join(directory, "ForeignKeys", fmt.Sprintf("%s.%s", fkey.ForeignKeyName, "sql"))

		if err := os.MkdirAll(filepath.Dir(foreignKeyPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(foreignKeyPath, []byte(foreignKeyDefinition), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func ScriptViews(db gorm.DB, directory string, views []sqlserver_models.Views, onProgress func(index int, total int, view sqlserver_models.Views)) error {
	var viewDefinition string = ""

	for index, view := range views {
		if onProgress != nil {
			onProgress(index, len(views), view)
		}

		viewDefinition = view.Definition

		// Views
		viewsPath := filepath.Join(directory, "Views", fmt.Sprintf("%s.%s", view.View, "sql"))
		if err := os.MkdirAll(filepath.Dir(viewsPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(viewsPath, []byte(viewDefinition), 0o644); err != nil {
			return err
		}

	}

	return nil
}
