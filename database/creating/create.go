package creating

import (
	"blueprint/cli/spinner"
	"blueprint/connections"
	"blueprint/database"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	"blueprint/database/scripting"
	"blueprint/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

func Create(input models.CommandInput) {
	argument := ""
	if len(input.Arguments) > 0 {
		argument = input.Arguments[0]
	}

	id, conn := connections.GetConnection(argument)
	if id == "" || conn == nil {
		return
	}

	directory := conn.Name
	applyConnectionOverrides(conn, input.StringFlags)

	connectionDetails := fmt.Sprintf("Server: %s\nPort: %s\nDatabase: %s\nUser: %s\n", conn.Server, conn.Port, conn.Database, conn.User)
	if !scripting.AskYesNo(fmt.Sprintf("%s\nDo you want to create %s?", connectionDetails, argument), false) {
		return
	}

	switch conn.Type {
	case models.SqlServer:
		if err := createSQLServer(*conn, directory); err != nil {
			panic(err)
		}
	default:
		fmt.Printf("Create is not supported for %s\n", conn.Type)
		return
	}
}

func createSQLServer(conn models.Connection, directory string) error {
	masterDB, err := database.SQLServerConnectToDatabase(conn, "master")
	if err != nil {
		return err
	}

	exists, err := database.SQLServerDatabaseExists(masterDB, conn.Database)
	if err != nil {
		return err
	}

	if exists {
		if !scripting.AskYesNo(fmt.Sprintf("Database %s already exists. Overwrite it?", conn.Database), false) {
			return nil
		}

		if err := database.SQLServerDropDatabase(masterDB, conn.Database); err != nil {
			return err
		}
	}

	metadata, err := loadDatabaseMetadata(directory)
	if err != nil {
		return err
	}

	if err := database.SQLServerCreateDatabase(masterDB, conn.Database, metadata.Collation); err != nil {
		return err
	}

	targetDB, err := database.SQLServerConnectToDatabase(conn, conn.Database)
	if err != nil {
		return err
	}

	startTime := time.Now()

	if _, err := executeRunOrder(targetDB, directory); err != nil {
		return err
	}

	if _, err := executeTableData(targetDB, directory); err != nil {
		return err
	}

	if _, err := executeForeignKeys(targetDB, directory); err != nil {
		return err
	}

	fmt.Printf("\nTotal time: %.2fs\n", time.Since(startTime).Seconds())

	return nil
}

func loadDatabaseMetadata(directory string) (sqlservermodels.DatabaseMetadata, error) {
	path := filepath.Join(directory, "Database.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return sqlservermodels.DatabaseMetadata{}, nil
		}
		return sqlservermodels.DatabaseMetadata{}, err
	}

	var metadata sqlservermodels.DatabaseMetadata
	if err := json.Unmarshal(contents, &metadata); err != nil {
		return sqlservermodels.DatabaseMetadata{}, err
	}

	return metadata, nil
}

func executeRunOrder(db *gorm.DB, directory string) (int, error) {
	runOrderPath := filepath.Join(directory, "RunOrder.json")
	contents, err := os.ReadFile(runOrderPath)
	if err != nil {
		return 0, err
	}

	var runOrder sqlservermodels.RunOrderFile
	if err := json.Unmarshal(contents, &runOrder); err != nil {
		return 0, err
	}

	schemaItems := make(sqlservermodels.RunOrderFile, 0, len(runOrder))
	for _, item := range runOrder {
		if !strings.HasPrefix(filepath.ToSlash(item.File), "ForeignKeys/") {
			schemaItems = append(schemaItems, item)
		}
	}

	schemaSpinner := spinner.New("Schema", "Creating schema")
	pauseForSpinner()
	err = db.Transaction(func(tx *gorm.DB) error {
		for index, item := range schemaItems {
			schemaSpinner.Update(fmt.Sprintf("[%d/%d] %s", index+1, len(schemaItems), item.Name))

			filePath := filepath.Join(directory, item.File)
			contents, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			if err := tx.Exec(string(contents)).Error; err != nil {
				return fmt.Errorf("execute %s: %w", item.File, err)
			}
		}
		return nil
	})
	if err != nil {
		schemaSpinner.Stop("Failed")
		return 0, err
	}
	schemaSpinner.Stop(fmt.Sprintf("%d objects created", len(schemaItems)))

	return len(schemaItems), nil
}

func executeForeignKeys(db *gorm.DB, directory string) (int, error) {
	contents, err := os.ReadFile(filepath.Join(directory, "RunOrder.json"))
	if err != nil {
		return 0, err
	}

	var runOrder sqlservermodels.RunOrderFile
	if err := json.Unmarshal(contents, &runOrder); err != nil {
		return 0, err
	}

	foreignKeys := make(sqlservermodels.RunOrderFile, 0)
	for _, item := range runOrder {
		if strings.HasPrefix(filepath.ToSlash(item.File), "ForeignKeys/") {
			foreignKeys = append(foreignKeys, item)
		}
	}

	foreignKeySpinner := spinner.New("Foreign Keys", "Applying foreign keys")
	pauseForSpinner()
	for index, item := range foreignKeys {
		foreignKeySpinner.Update(fmt.Sprintf("[%d/%d] %s", index+1, len(foreignKeys), item.Name))
		pauseForSpinner()

		fileContents, err := os.ReadFile(filepath.Join(directory, item.File))
		if err != nil {
			foreignKeySpinner.Stop("Failed")
			return 0, err
		}
		if err := db.Exec(string(fileContents)).Error; err != nil {
			foreignKeySpinner.Stop("Failed")
			return 0, fmt.Errorf("execute %s: %w", item.File, err)
		}
	}
	foreignKeySpinner.Stop(fmt.Sprintf("%d foreign keys applied", len(foreignKeys)))

	return len(foreignKeys), nil
}

func executeTableData(db *gorm.DB, directory string) (int, error) {
	tableDataPath := filepath.Join(directory, "TableData")
	dataFiles, err := sqlFilesRecursive(tableDataPath)
	if err != nil {
		return 0, err
	}

	dataSpinner := spinner.New("Data", "Inserting table data")
	pauseForSpinner()
	if len(dataFiles) == 0 {
		dataSpinner.Stop("0 rows inserted")
		return 0, nil
	}

	tableBatches := make(map[string]*tableDataBatch)
	tableNames := make([]string, 0)
	for _, filePath := range dataFiles {
		relativePath, err := filepath.Rel(tableDataPath, filePath)
		if err != nil {
			dataSpinner.Stop("Failed")
			return 0, err
		}

		tableName := filepath.Dir(relativePath)
		batch := tableBatches[tableName]
		if batch == nil {
			batch = &tableDataBatch{Name: filepath.ToSlash(tableName)}
			tableBatches[tableName] = batch
			tableNames = append(tableNames, tableName)
		}

		contents, err := os.ReadFile(filePath)
		if err != nil {
			dataSpinner.Stop("Failed")
			return 0, err
		}
		batch.SQL.Write(contents)
		batch.SQL.WriteString("\n")
		batch.Rows++
	}
	sort.Strings(tableNames)

	rowCount := len(dataFiles)
	err = db.Transaction(func(tx *gorm.DB) error {
		for index, tableName := range tableNames {
			batch := tableBatches[tableName]
			dataSpinner.Update(fmt.Sprintf("[%d/%d] %s", index+1, len(tableNames), batch.Name))
			pauseForSpinner()

			if err := tx.Exec(batch.SQL.String()).Error; err != nil {
				return fmt.Errorf("execute TableData/%s: %w", batch.Name, err)
			}
		}
		return nil
	})
	if err != nil {
		dataSpinner.Stop("Failed")
		return 0, err
	}

	dataSpinner.Stop(fmt.Sprintf("%d rows inserted", rowCount))

	return rowCount, nil
}

type tableDataBatch struct {
	Name string
	SQL  strings.Builder
	Rows int
}

func sqlFilesRecursive(directory string) ([]string, error) {
	if _, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	files := make([]string, 0)
	if err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			return nil
		}

		files = append(files, path)
		return nil
	}); err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

func pauseForSpinner() {
	time.Sleep(50 * time.Millisecond)
}

func applyConnectionOverrides(conn *models.Connection, overrides map[string]string) {
	if overrides["server"] != "" {
		conn.Server = overrides["server"]
	}
	if overrides["port"] != "" {
		conn.Port = overrides["port"]
	}
	if overrides["database"] != "" {
		conn.Database = overrides["database"]
	}
	if overrides["user"] != "" {
		conn.User = overrides["user"]
	}
	if overrides["password"] != "" {
		conn.Password = overrides["password"]
	}
}
