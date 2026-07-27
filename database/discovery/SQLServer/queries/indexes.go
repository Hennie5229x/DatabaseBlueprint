package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerIndexes(db *gorm.DB, tableName string) []models.IndexColumn {
	var indexes []models.IndexColumn

	err := db.Raw(`
		SELECT		i.index_id AS IndexID,
					i.name AS IndexName,
					CASE WHEN i.type = 1 THEN 'CLUSTERED' ELSE 'NONCLUSTERED' END AS IndexType,
					i.is_unique AS IsUnique,
					c.name AS ColumnName,
					ic.key_ordinal AS KeyOrdinal,
					ic.is_descending_key AS IsDescending,
					ic.is_included_column AS IsIncluded,
					ic.index_column_id AS IncludeOrder,
					i.has_filter AS HasFilter,
					i.filter_definition AS FilterDefinition,
					CAST(1 AS bit) AS IsUserDefinedName
		FROM 		sys.indexes i
		JOIN 		sys.index_columns ic ON ic.object_id = i.object_id
					AND ic.index_id = i.index_id
		JOIN 		sys.columns c ON c.object_id = ic.object_id
					AND c.column_id = ic.column_id
		WHERE		i.object_id = OBJECT_ID(?, 'U')
		AND 		i.type IN (1, 2)
		AND 		i.is_primary_key = 0
		AND 		i.is_unique_constraint = 0
		AND 		i.is_hypothetical = 0
		ORDER BY 	i.index_id, ic.key_ordinal, ic.index_column_id;
	`, tableName).Scan(&indexes).Error

	if err != nil {
		panic(err)
	}

	return indexes
}
