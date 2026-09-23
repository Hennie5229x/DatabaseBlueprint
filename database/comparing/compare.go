package comparing

import (
	"blueprint/connections"
	"blueprint/database"
	sqlserverComparing "blueprint/database/comparing/SQLServer"
	"blueprint/models"
	"fmt"
	"path/filepath"
)

func Compare(input models.CommandInput) {
	var sourceArg string
	var targetArg string
	if len(input.Arguments) > 1 {
		sourceArg = input.Arguments[0]
		targetArg = input.Arguments[1]
	}

	source_id, source_conn := connections.GetConnection(sourceArg)
	target_id, target_conn := connections.GetConnection(targetArg)

	if source_id == "" || source_conn == nil {
		return
	}
	if target_id == "" || target_conn == nil {
		return
	}

	if source_conn.Type != target_conn.Type {
		fmt.Println("Source and Target database type not the same!")
	}

	if models.DatabaseType(source_conn.Id) == models.DatabaseType(target_conn.Id) {
		fmt.Println("Source and Target cannot be the same connection!")
	}

	db_source, _ := database.Connect(*source_conn)
	db_target, _ := database.Connect(*target_conn)

	switch source_conn.Type {

	case models.SqlServer:
		outputPath := filepath.Join("Migrations", sourceArg+"-to-"+targetArg+".sql")
		sqlserverComparing.Init(db_source, db_target, outputPath)
	default:
		fmt.Printf("Compare is not supported for %s\n", source_conn.Type)
		return
	}
}
