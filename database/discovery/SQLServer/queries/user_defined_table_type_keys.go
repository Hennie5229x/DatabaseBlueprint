package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerUserDefinedTableTypeKeys(db *gorm.DB, schemaName string, typeName string) []models.UserDefinedTableTypeKeyColumn {
	var keys []models.UserDefinedTableTypeKeyColumn

	err := db.Raw(`
		SELECT      s.name AS SchemaName,
					tt.name AS TypeName,
					kc.object_id AS ConstraintObjectID,
					kc.name AS ConstraintName,
					CASE WHEN kc.type = 'PK' THEN 'PRIMARY KEY' ELSE 'UNIQUE' END AS ConstraintType,
					i.name AS IndexName,
					i.type_desc AS IndexType,
					c.name AS ColumnName,
					ic.key_ordinal AS KeyOrdinal,
					ic.is_descending_key AS IsDescending,
					ISNULL(hi.bucket_count, 0) AS BucketCount
		FROM        sys.table_types AS tt
		INNER JOIN  sys.schemas AS s ON s.schema_id = tt.schema_id
		INNER JOIN  sys.key_constraints AS kc ON kc.parent_object_id = tt.type_table_object_id
		INNER JOIN  sys.indexes AS i ON i.object_id = kc.parent_object_id
					AND i.index_id = kc.unique_index_id
		INNER JOIN  sys.index_columns AS ic ON ic.object_id = i.object_id
					AND ic.index_id = i.index_id
		INNER JOIN  sys.columns AS c ON c.object_id = ic.object_id
					AND c.column_id = ic.column_id
		LEFT JOIN   sys.hash_indexes AS hi ON hi.object_id = i.object_id
					AND hi.index_id = i.index_id
		WHERE       s.name = ?
		AND         tt.name = ?
		ORDER BY    kc.object_id, ic.key_ordinal;
	`, schemaName, typeName).Scan(&keys).Error

	if err != nil {
		panic(err)
	}

	return keys
}
