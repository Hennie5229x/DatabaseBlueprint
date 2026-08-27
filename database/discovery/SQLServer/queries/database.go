package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerDatabaseMetadata(db *gorm.DB) models.DatabaseMetadata {
	var metadata models.DatabaseMetadata

	err := db.Raw(`
		SELECT CONVERT(sysname, DATABASEPROPERTYEX(DB_NAME(), 'Collation')) AS Collation;
	`).Scan(&metadata).Error
	if err != nil {
		panic(err)
	}

	return metadata
}
