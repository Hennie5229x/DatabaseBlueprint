package sqlserver

import (
	"blueprint/database/discovery/SQLServer/queries"
	"encoding/json"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

func DatabaseMetadata(db *gorm.DB, directory string) {
	metadata := queries.SqlServerDatabaseMetadata(db)
	contents, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		panic(err)
	}

	path := filepath.Join(directory, "Database.json")
	if err := os.WriteFile(path, append(contents, '\n'), 0o644); err != nil {
		panic(err)
	}
}
