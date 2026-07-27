package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerPrimaryKeys(db *gorm.DB, tableName string) []models.PrimaryKeyColumn {
	var primaryKeys []models.PrimaryKeyColumn

	err := db.Raw(`
		SELECT		kc.object_id AS ConstraintObjectID,
					CASE WHEN i.type = 1 THEN 'CLUSTERED' ELSE 'NONCLUSTERED' END AS IndexType,
					c.name AS ColumnName,
					ic.key_ordinal AS KeyOrdinal,
					ic.is_descending_key AS IsDescending
		FROM 		sys.key_constraints kc
		JOIN 		sys.indexes i ON i.object_id = kc.parent_object_id
					AND i.index_id = kc.unique_index_id
		JOIN 		sys.index_columns ic ON ic.object_id = i.object_id
					AND ic.index_id = i.index_id
		JOIN 		sys.columns c ON c.object_id = ic.object_id
					AND c.column_id = ic.column_id
		WHERE 		kc.parent_object_id = OBJECT_ID(?, 'U')
		AND 		kc.type = 'PK'
		ORDER BY 	ic.key_ordinal;
	`, tableName).Scan(&primaryKeys).Error

	if err != nil {
		panic(err)
	}

	return primaryKeys
}
