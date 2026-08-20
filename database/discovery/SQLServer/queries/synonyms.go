package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerSynonyms(db *gorm.DB) []models.Synonyms {
	var synonyms []models.Synonyms

	err := db.Raw(`
		SELECT      s.name AS SchemaName,
					sy.name AS SynonymName,
					sy.base_object_name AS BaseObjectName
		FROM        sys.synonyms sy
		JOIN        sys.schemas s ON sy.schema_id = s.schema_id
		ORDER BY    s.name, sy.name
	`).Scan(&synonyms).Error

	if err != nil {
		panic(err)
	}

	return synonyms
}
