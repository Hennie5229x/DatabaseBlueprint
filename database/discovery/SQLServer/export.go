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

func BuildSynonymsScript(synonyms sqlserver_models.Synonyms) string {
	synonymsDefinition := sqlServerSynonymsDefinition(synonyms)
	if len(synonymsDefinition) == 0 {
		return ""
	}

	return synonymsDefinition
}

func BuildSchemaScript(schema sqlserver_models.Schemas) string {
	schemasDefinition := sqlServerSchemaDefinition(schema)
	if len(schemasDefinition) == 0 {
		return ""
	}

	return schemasDefinition
}

func BuildTableTypeScript(db gorm.DB, tableType sqlserver_models.UserDefinedTableType) string {
	columns := TableTypeColumns(&db, tableType)
	keys := TableTypeKeys(&db, tableType)
	checks := TableTypeChecks(&db, tableType)
	indexes := TableTypeIndexes(&db, tableType)

	return GenerateCreateTableType(tableType, columns, keys, checks, indexes)
}

func BuildUserDefinedTypeScript(userDefinedType sqlserver_models.UserDefinedType) string {
	return GenerateCreateUserDefinedType(userDefinedType)
}

func BuildSequenceScript(sequence sqlserver_models.Sequences) string {
	return GenerateCreateSequence(sequence)
}

func BuildTriggerScript(trigger sqlserver_models.Triggers) string {
	return GenerateCreateTrigger(trigger)
}

func ScriptTables(db gorm.DB, directory string, tables []discoverymodels.Tables, onProgress func(index int, total int, table discoverymodels.Tables)) error {

	for index, table := range tables {
		if onProgress != nil {
			onProgress(index, len(tables), table)
		}

		tableDefinition := BuildTableScript(db, table.Schema, table.Name)

		// Tables
		tablesPath := filepath.Join(directory, "Tables", fmt.Sprintf("%s.%s.%s", table.Schema, table.Name, "sql"))
		if err := os.MkdirAll(filepath.Dir(tablesPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(tablesPath, []byte(tableDefinition), 0o644); err != nil {
			return err
		}

	}

	return nil
}

func ScriptTableTypes(db gorm.DB, directory string, tableTypes []sqlserver_models.UserDefinedTableType, onProgress func(index int, total int, tableType sqlserver_models.UserDefinedTableType)) error {
	for index, tableType := range tableTypes {
		if onProgress != nil {
			onProgress(index, len(tableTypes), tableType)
		}

		tableTypeDefinition := BuildTableTypeScript(db, tableType)
		tableTypePath := filepath.Join(directory, "TableTypes", fmt.Sprintf("%s.%s.%s", tableType.SchemaName, tableType.TypeName, "sql"))
		if err := os.MkdirAll(filepath.Dir(tableTypePath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(tableTypePath, []byte(tableTypeDefinition), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func ScriptUserDefinedTypes(directory string, userDefinedTypes []sqlserver_models.UserDefinedType, onProgress func(index int, total int, userDefinedType sqlserver_models.UserDefinedType)) error {
	for index, userDefinedType := range userDefinedTypes {
		if onProgress != nil {
			onProgress(index, len(userDefinedTypes), userDefinedType)
		}

		userDefinedTypeDefinition := BuildUserDefinedTypeScript(userDefinedType)
		userDefinedTypePath := filepath.Join(directory, "DataTypes", fmt.Sprintf("%s.%s.%s", userDefinedType.SchemaName, userDefinedType.TypeName, "sql"))
		if err := os.MkdirAll(filepath.Dir(userDefinedTypePath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(userDefinedTypePath, []byte(userDefinedTypeDefinition), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func ScriptSequences(db gorm.DB, directory string, sequences []sqlserver_models.Sequences, onProgress func(index int, total int, sequence sqlserver_models.Sequences)) error {
	for index, sequence := range sequences {
		if onProgress != nil {
			onProgress(index, len(sequences), sequence)
		}

		sequenceDefinition := BuildSequenceScript(sequence)
		sequencePath := filepath.Join(directory, "Sequences", fmt.Sprintf("%s.%s.%s", sequence.SchemaName, sequence.SequenceName, "sql"))
		if err := os.MkdirAll(filepath.Dir(sequencePath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(sequencePath, []byte(sequenceDefinition), 0o644); err != nil {
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
		foreignKeyPath := filepath.Join(directory, "ForeignKeys", fmt.Sprintf("%s.%s.%s", fkey.ParentSchema, fkey.ForeignKeyName, "sql"))

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
		viewsPath := filepath.Join(directory, "Views", fmt.Sprintf("%s.%s.%s", view.Schema, view.View, "sql"))
		if err := os.MkdirAll(filepath.Dir(viewsPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(viewsPath, []byte(viewDefinition), 0o644); err != nil {
			return err
		}

	}

	return nil
}

func ScriptFunctions(db gorm.DB, directory string, functions []sqlserver_models.Functions, onProgress func(index int, total int, function sqlserver_models.Functions)) error {
	var functionDefinition string = ""

	for index, fn := range functions {
		if onProgress != nil {
			onProgress(index, len(functions), fn)
		}

		functionDefinition = fn.Definition

		// Functions
		functionsPath := filepath.Join(directory, "Functions", fmt.Sprintf("%s.%s.%s", fn.Schema, fn.Name, "sql"))
		if err := os.MkdirAll(filepath.Dir(functionsPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(functionsPath, []byte(functionDefinition), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func ScriptProcedures(db gorm.DB, directory string, procedures []sqlserver_models.Procedures, onProgress func(index int, total int, proc sqlserver_models.Procedures)) error {
	var procDefinition string = ""

	for index, p := range procedures {
		if onProgress != nil {
			onProgress(index, len(procedures), p)
		}

		procDefinition = p.Definition

		// Procedures
		procPath := filepath.Join(directory, "Procedures", fmt.Sprintf("%s.%s.%s", p.Schema, p.Name, "sql"))
		if err := os.MkdirAll(filepath.Dir(procPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(procPath, []byte(procDefinition), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func ScriptTriggers(db gorm.DB, directory string, triggers []sqlserver_models.Triggers, onProgress func(index int, total int, trigger sqlserver_models.Triggers)) error {
	var triggerDefinition string = ""

	for index, trigger := range triggers {
		if onProgress != nil {
			onProgress(index, len(triggers), trigger)
		}

		triggerDefinition = BuildTriggerScript(trigger)

		triggerPath := filepath.Join(directory, "Triggers", fmt.Sprintf("%s.%s.%s", trigger.SchemaName, trigger.TriggerName, "sql"))
		if err := os.MkdirAll(filepath.Dir(triggerPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(triggerPath, []byte(triggerDefinition), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func ScriptSynonyms(db gorm.DB, directory string, synonyms []sqlserver_models.Synonyms, onProgress func(index int, total int, syn sqlserver_models.Synonyms)) error {
	var synonymDefinition string = ""

	for index, sy := range synonyms {
		if onProgress != nil {
			onProgress(index, len(synonyms), sy)
		}

		synonymDefinition = BuildSynonymsScript(sy)

		// Functions
		functionsPath := filepath.Join(directory, "Synonyms", fmt.Sprintf("%s.%s.%s", sy.SchemaName, sy.SynonymName, "sql"))
		if err := os.MkdirAll(filepath.Dir(functionsPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(functionsPath, []byte(synonymDefinition), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func ScriptSchemas(db gorm.DB, directory string, schemas []sqlserver_models.Schemas, onProgress func(index int, total int, sch sqlserver_models.Schemas)) error {
	var schemasDefinition string = ""

	for index, s := range schemas {
		if onProgress != nil {
			onProgress(index, len(schemas), s)
		}

		schemasDefinition = BuildSchemaScript(s)

		// Schemas
		functionsPath := filepath.Join(directory, "Schemas", fmt.Sprintf("%s.%s", s.Name, "sql"))
		if err := os.MkdirAll(filepath.Dir(functionsPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(functionsPath, []byte(schemasDefinition), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func ScriptTableData(directory string, schemaName string, tableName string, rows []map[string]interface{}, columns []sqlserver_models.Column, primaryKeys []sqlserver_models.PrimaryKeyColumn) error {
	tableDataPath := filepath.Join(directory, "TableData", tableName)
	if err := os.MkdirAll(tableDataPath, 0o755); err != nil {
		return err
	}

	columnsByName := columnsByName(columns)
	for index, row := range rows {
		rowDefinition := generateInsertStatement(schemaName, tableName, row, columns)
		fileName := tableDataFileName(row, primaryKeys, columnsByName, index+1)
		rowPath := filepath.Join(tableDataPath, fmt.Sprintf("%s.sql", fileName))
		if err := os.WriteFile(rowPath, []byte(rowDefinition), 0o644); err != nil {
			return err
		}
	}

	return nil
}
