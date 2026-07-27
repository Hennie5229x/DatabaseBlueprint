package scripting

import (
	"blueprint/appinfo"
	"blueprint/connections"
	"blueprint/database"
	discoverysqlserver "blueprint/database/discovery/SQLServer"
	"blueprint/models"
	"fmt"
	"os"

	"gorm.io/gorm"
)

const directory string = appinfo.ScriptDirectory

func Script(args []string) {
	var Argument string = ""
	if len(args) > 0 {
		Argument = args[0]
	}

	clearScriptDirectory(directory)

	// Get Connection
	id, conn := connections.GetConnection(Argument)
	if id == "" || conn == nil {
		return
	}

	db, _ := database.Connect(*conn)

	fmt.Println(Argument)

	switch conn.Type {
	case models.SqlServer:
		Script_SQLServer(db)
	case models.MySql:
		Script_MySQL()
	case models.PostgreSql:
		Script_PostgreSql()
	case models.SQLite:
		Script_SQLite()
	}
}

func Script_SQLServer(db *gorm.DB) {
	discoverysqlserver.Tables(db, directory)
	discoverysqlserver.ForeignKeys(db, directory)
	discoverysqlserver.Views(db, directory)
}
func Script_MySQL() {

}
func Script_PostgreSql() {

}
func Script_SQLite() {

}

func clearScriptDirectory(path string) error {
	// Remove all files & folders
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	// Recreate the empty directory.
	return os.MkdirAll(path, 0o755)
}
