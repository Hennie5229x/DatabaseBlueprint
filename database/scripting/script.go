package scripting

import (
	"blueprint/appinfo"
	"blueprint/connections"
	"blueprint/database"
	discoverysqlserver "blueprint/database/discovery/SQLServer"
	"blueprint/models"
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

const directory string = appinfo.ScriptDirectory

func Script(args []string) {

	var Argument string = ""
	if len(args) > 0 {
		Argument = args[0]
	}

	var yn bool = AskYesNo(fmt.Sprintf("Do you want to script %s?", Argument), true)

	if !yn {
		return
	}

	startTime := time.Now()

	clearScriptDirectory(directory)

	// Get Connection
	id, conn := connections.GetConnection(Argument)
	if id == "" || conn == nil {
		return
	}

	db, _ := database.Connect(*conn)

	fmt.Printf("----- %s -----\n", Argument)

	switch conn.Type {
	case models.SqlServer:
		Script_SQLServer(db, conn.Database)
	case models.MySql:
		Script_MySQL()
	case models.PostgreSql:
		Script_PostgreSql()
	case models.SQLite:
		Script_SQLite()
	}

	fmt.Printf("\nTotal time: %.2fs\n", time.Since(startTime).Seconds())
}

func Script_SQLServer(db *gorm.DB, databaseName string) {
	discoverysqlserver.UserDefinedTypes(db, directory)
	discoverysqlserver.Tables(db, directory)
	discoverysqlserver.TableTypes(db, directory)
	discoverysqlserver.ForeignKeys(db, directory)
	discoverysqlserver.Views(db, directory)
	discoverysqlserver.Functions(db, directory)
	discoverysqlserver.Procedures(db, directory)
	discoverysqlserver.GenerateRunOrder(db, directory, databaseName)
	discoverysqlserver.TableData(db, directory)
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

// -----------------------
// Yes / No  CLI handler
// -----------------------
func AskYesNo(prompt string, def bool) bool {
	choices := "Y/n"
	if !def {
		choices = "y/N"
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [%s]: ", prompt, choices)
		input, err := reader.ReadString('\n')
		if err != nil {
			return def
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input == "" {
			return def
		}
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Invalid input. Please type 'y' or 'n'.")
	}
}
