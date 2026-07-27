package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerViews(db gorm.DB) []models.Views {
	var view []models.Views

	err := db.Raw(`
		SELECT      s.name AS [Schema],
					v.name AS [View],
					sm.definition AS [Definition]
		FROM        sys.views AS v
		JOIN        sys.schemas AS s ON v.schema_id = s.schema_id
		JOIN        sys.sql_modules AS sm ON sm.object_id = v.object_id
		ORDER BY    s.name, v.name;


	`, 1).Scan(&view).Error

	if err != nil {
		panic(err)
	}

	return view
}
