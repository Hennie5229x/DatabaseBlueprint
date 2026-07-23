package discovery

import (
	sqlserver "blueprint/database/discovery/schema/SQLServer"
	"blueprint/models"
	"fmt"

	"gorm.io/gorm"
)

func GetColumns(table string, conn models.Connection, db *gorm.DB) {
	/*
		var Argument string = ""
		if len(args) > 0 {
			Argument = args[0]
		}

		tableName := Argument
	*/

	//id, conn := connections.GetConnection(Argument)
	/*
		if id == "" {
			panic("Connection not found")
		}

		db, _ := database.Connect(*conn)
	*/

	var columns []sqlserver.Column

	if conn.Type == models.SqlServer {
		columns = sqlserver.SqlServerColumns(db, table)
	}

	for _, c := range columns {
		fmt.Printf("%d  %s\n", c.ColumnID, c.ColumnName)
	}
}
