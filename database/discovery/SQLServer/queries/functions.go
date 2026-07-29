package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerFunctions(db *gorm.DB) []models.Functions {
	var functions []models.Functions

	err := db.Raw(`
		SELECT      s.name AS [Schema],
					o.name AS [Name],
					sm.definition AS [Definition]
		FROM        sys.objects AS o
		JOIN        sys.schemas AS s ON o.schema_id = s.schema_id
		JOIN        sys.sql_modules AS sm ON sm.object_id = o.object_id
		WHERE       o.type IN ('FN', 'IF', 'TF')
		ORDER BY    s.name, o.name;
	`).Scan(&functions).Error

	if err != nil {
		panic(err)
	}

	return functions
}
