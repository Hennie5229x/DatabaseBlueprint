package queries

import (
	"blueprint/database/discovery/models"

	"gorm.io/gorm"
)

func SqlServerTables(db gorm.DB) []models.Tables {
	var table []models.Tables

	err := db.Raw(`
		SELECT  	sch.name AS [Schema],
					tbl.name AS [Name]
		FROM    	sys.tables tbl
		JOIN    	sys.schemas sch ON sch.schema_id = tbl.schema_id
		WHERE   	type = 'U'
		ORDER BY	tbl.name ASC

	`, 1).Scan(&table).Error

	if err != nil {
		panic(err)
	}

	return table
}
