package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerSchemas(db *gorm.DB) []models.Schemas {
	var schemas []models.Schemas

	err := db.Raw(`
		SELECT  name
		FROM    sys.schemas
		WHERE   name NOT IN (
			'dbo',
			'guest',
			'sys',
			'INFORMATION_SCHEMA',
			'db_owner',
			'db_accessadmin',
			'db_securityadmin',
			'db_ddladmin',
			'db_backupoperator',
			'db_datareader',
			'db_datawriter',
			'db_denydatareader',
			'db_denydatawriter'
		)
		ORDER BY name
	`).Scan(&schemas).Error

	if err != nil {
		panic(err)
	}

	return schemas
}
