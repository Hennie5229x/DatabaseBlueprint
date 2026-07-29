package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerProcedures(db *gorm.DB) []models.Procedures {
	var procs []models.Procedures

	err := db.Raw(`
		SELECT      s.name AS [Schema],
					p.name AS [Name],
					sm.definition AS [Definition]
		FROM        sys.procedures AS p
		JOIN        sys.schemas AS s ON p.schema_id = s.schema_id
		JOIN        sys.sql_modules AS sm ON sm.object_id = p.object_id
		ORDER BY    s.name, p.name;
	`).Scan(&procs).Error

	if err != nil {
		panic(err)
	}

	return procs
}
