package database

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func SQLServerDatabaseExists(db *gorm.DB, databaseName string) (bool, error) {
	var count int64
	err := db.Raw("SELECT COUNT(1) FROM sys.databases WHERE name = ?", databaseName).Scan(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func SQLServerDropDatabase(db *gorm.DB, databaseName string) error {
	databaseIdentifier := quoteSQLServerIdentifier(databaseName)
	return db.Exec(fmt.Sprintf("ALTER DATABASE %s SET SINGLE_USER WITH ROLLBACK IMMEDIATE; DROP DATABASE %s;", databaseIdentifier, databaseIdentifier)).Error
}

func SQLServerCreateDatabase(db *gorm.DB, databaseName string) error {
	return db.Exec("CREATE DATABASE " + quoteSQLServerIdentifier(databaseName)).Error
}

func quoteSQLServerIdentifier(identifier string) string {
	escapedIdentifier := strings.ReplaceAll(identifier, "]", "]]")
	return "[" + escapedIdentifier + "]"
}
