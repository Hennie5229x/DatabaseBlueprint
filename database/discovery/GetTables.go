package discovery

import (
	"blueprint/connections"
	"blueprint/database"
	discoverymodels "blueprint/database/discovery/models"
	sqlserver "blueprint/database/discovery/schema/SQLServer"
	"blueprint/models"
	"fmt"
)

func GetTables(args []string) {

	var Argument string = ""
	if len(args) > 0 {
		Argument = args[0]
	}

	id, conn := connections.GetConnection(Argument)
	if id == "" || conn == nil {
		return
	}

	db, _ := database.Connect(*conn)

	var table []discoverymodels.Tables

	if conn.Type == models.SqlServer {
		table = sqlserver.SqlServerTables(*db)
	}

	var sqlServerForeignKeyStatements []string

	for _, t := range table {
		fmt.Printf("%s.%s\n", t.Schema, t.Name)
		fmt.Println("--------------------------------")
		fmt.Println()

		//GetColumns(t.Name, *conn, db)

		//----------------------
		qualifiedTableName := fmt.Sprintf("%s.%s", t.Schema, t.Name)
		columns := sqlserver.SqlServerColumns(db, qualifiedTableName)
		defaultConstraints := sqlserver.SqlServerDefaultConstraints(db, qualifiedTableName)
		primaryKeys := sqlserver.SqlServerPrimaryKeys(db, qualifiedTableName)
		uniqueConstraints := sqlserver.SqlServerUniqueConstraints(db, qualifiedTableName)
		foreignKeys := sqlserver.SqlServerForeignKeys(db, qualifiedTableName)
		checkConstraints := sqlserver.SqlServerCheckConstraints(db, qualifiedTableName)
		indexes := sqlserver.SqlServerIndexes(db, qualifiedTableName)

		createTableSQL := sqlserver.GenerateCreateTable(
			t.Schema,
			t.Name,
			columns,
			defaultConstraints,
			primaryKeys,
			uniqueConstraints,
			checkConstraints,
		)
		createIndexesSQL := sqlserver.GenerateCreateIndexes(t.Schema, t.Name, indexes)
		sqlServerForeignKeyStatements = append(
			sqlServerForeignKeyStatements,
			sqlserver.GenerateForeignKeys(t.Schema, t.Name, foreignKeys)...,
		)

		fmt.Println(createTableSQL)
		if createIndexesSQL != "" {
			fmt.Println()
			fmt.Println(createIndexesSQL)
		}

		fmt.Println()
	}

	for _, foreignKeySQL := range sqlServerForeignKeyStatements {
		fmt.Println(foreignKeySQL)
		fmt.Println()
	}
}
