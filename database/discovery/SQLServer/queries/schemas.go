package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerSchemas(db *gorm.DB) []models.Schemas {
	var schemas []models.Schemas

	err := db.Raw(`
		SELECT      s.name
		FROM        sys.schemas AS s
		LEFT JOIN   sys.database_principals AS p ON p.principal_id = s.principal_id
		WHERE       s.name NOT IN (
			'dbo',
			'guest',
			'sys',
			'INFORMATION_SCHEMA'
		)
		AND ISNULL(p.is_fixed_role, 0) = 0
		ORDER BY s.name
	`).Scan(&schemas).Error

	if err != nil {
		panic(err)
	}

	return schemas
}
