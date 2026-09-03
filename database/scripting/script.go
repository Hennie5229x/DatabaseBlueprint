package scripting

import (
	"blueprint/appinfo"
	"blueprint/cli/spinner"
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

var directory string = appinfo.ScriptDirectory

func Script(input models.CommandInput) {

	var Argument string = ""
	if len(input.Arguments) > 0 {
		Argument = input.Arguments[0]
	}
	// Flags
	dataOnly := input.Flags["data-only"]
	schemaOnly := input.Flags["schema-only"]

	serverOveride := input.StringFlags["server"]
	portOveride := input.StringFlags["port"]
	databaseOveride := input.StringFlags["database"]
	userOveride := input.StringFlags["user"]
	passwordOveride := input.StringFlags["password"]

	if dataOnly && schemaOnly {
		fmt.Println("Cannot use --data-only and --schema-only together")
		return
	}

	// Get Connection
	id, conn := connections.GetConnection(Argument)
	if id == "" || conn == nil {
		return
	}
	directory = conn.Name

	// Apply Overide values
	if serverOveride != "" {
		conn.Server = serverOveride
	}
	if portOveride != "" {
		conn.Port = portOveride
	}
	if databaseOveride != "" {
		conn.Database = databaseOveride
	}
	if userOveride != "" {
		conn.User = userOveride
	}
	if passwordOveride != "" {
		conn.Password = passwordOveride
	}

	var connectionDetails string = fmt.Sprintf("Server: %s\nPort: %s\nDatabase: %s\nUser: %s\n",
		conn.Server, conn.Port, conn.Database, conn.User)

	var yn bool = AskYesNo(fmt.Sprintf("%s\nDo you want to script %s?", connectionDetails, Argument), false)

	if !yn {
		return
	}

	startTime := time.Now()
	cleanupSpinner := spinner.New("Cleanup", "Deleting previous scripts")
	if err := clearScriptDirectory(directory); err != nil {
		cleanupSpinner.Stop("Failed")
		panic(err)
	}
	cleanupSpinner.Stop("Directory cleared")

	db, _ := database.Connect(*conn)

	fmt.Printf("----- %s -----\n", Argument)

	switch conn.Type {
	case models.SqlServer:
		Script_SQLServer(db, conn.Database, dataOnly, schemaOnly)
		/*
			case models.MySql:
				Script_MySQL()
			case models.PostgreSql:
				Script_PostgreSql()
			case models.SQLite:
				Script_SQLite()
		*/
	}

	fmt.Printf("\nTotal time: %.2fs\n", time.Since(startTime).Seconds())
}

func Script_SQLServer(db *gorm.DB, databaseName string, dataOnly bool, schemaOnly bool) {

	if !dataOnly {
		discoverysqlserver.DatabaseMetadata(db, directory)
		discoverysqlserver.Schemas(db, directory)
		discoverysqlserver.UserDefinedTypes(db, directory)
		discoverysqlserver.TableTypes(db, directory)
		discoverysqlserver.Sequences(db, directory)
		discoverysqlserver.Synonyms(db, directory)
		discoverysqlserver.Tables(db, directory)
		discoverysqlserver.ForeignKeys(db, directory)
		discoverysqlserver.Views(db, directory)
		discoverysqlserver.Functions(db, directory)
		discoverysqlserver.Procedures(db, directory)
		discoverysqlserver.Triggers(db, directory)
		discoverysqlserver.GenerateRunOrder(db, directory, databaseName)
	}
	if !schemaOnly {
		discoverysqlserver.TableData(db, directory)
	}
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
