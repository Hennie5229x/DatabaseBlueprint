package comparing

import (
	"blueprint/connections"
	"blueprint/database"
	sqlserverComparing "blueprint/database/comparing/SQLServer"
	"blueprint/database/scripting"
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

	outputPath := filepath.Join("Migrations", sourceArg+"-to-"+targetArg+".sql")
	connectionDetails := fmt.Sprintf(
		"Source: %s\nServer: %s\nPort: %s\nDatabase: %s\nUser: %s\n\nTarget: %s\nServer: %s\nPort: %s\nDatabase: %s\nUser: %s",
		sourceArg,
		source_conn.Server,
		source_conn.Port,
		source_conn.Database,
		source_conn.User,
		targetArg,
		target_conn.Server,
		target_conn.Port,
		target_conn.Database,
		target_conn.User,
	)
	if !scripting.AskYesNo(fmt.Sprintf("%s\n\nDo you want to generate a migration from %s to %s?\nOutput: %s", connectionDetails, sourceArg, targetArg, outputPath), false) {
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
		sqlserverComparing.Init(db_source, db_target, outputPath)
	default:
		fmt.Printf("Compare is not supported for %s\n", source_conn.Type)
		return
	}
}
