package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerUserDefinedTableTypeIndexes(db *gorm.DB, schemaName string, typeName string) []models.UserDefinedTableTypeIndexColumn {
	var indexes []models.UserDefinedTableTypeIndexColumn

	err := db.Raw(`
		SELECT      s.name AS SchemaName,
					tt.name AS TypeName,
					i.index_id AS IndexID,
					i.name AS IndexName,
					i.type_desc AS IndexType,
					i.is_unique AS IsUnique,
					c.name AS ColumnName,
					ic.key_ordinal AS KeyOrdinal,
					ic.is_descending_key AS IsDescending,
					ic.is_included_column AS IsIncluded,
					ic.index_column_id AS IncludeOrder,
					i.has_filter AS HasFilter,
					i.filter_definition AS FilterDefinition,
					ISNULL(hi.bucket_count, 0) AS BucketCount
		FROM        sys.table_types AS tt
		INNER JOIN  sys.schemas AS s ON s.schema_id = tt.schema_id
		INNER JOIN  sys.indexes AS i ON i.object_id = tt.type_table_object_id
		INNER JOIN  sys.index_columns AS ic ON ic.object_id = i.object_id
					AND ic.index_id = i.index_id
		INNER JOIN  sys.columns AS c ON c.object_id = ic.object_id
					AND c.column_id = ic.column_id
		LEFT JOIN   sys.hash_indexes AS hi ON hi.object_id = i.object_id
					AND hi.index_id = i.index_id
		WHERE       s.name = ?
		AND         tt.name = ?
		AND         i.index_id > 0
		AND         i.is_primary_key = 0
		AND         i.is_unique_constraint = 0
		AND         i.is_hypothetical = 0
		ORDER BY    i.index_id, ic.key_ordinal, ic.index_column_id;
	`, schemaName, typeName).Scan(&indexes).Error

	if err != nil {
		panic(err)
	}

	return indexes
}
