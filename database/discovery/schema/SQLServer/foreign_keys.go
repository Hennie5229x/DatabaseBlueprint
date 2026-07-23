package sqlserver

import (
	"gorm.io/gorm"
)

func SqlServerForeignKeys(db *gorm.DB, tableName string) []ForeignKeyColumn {
	var foreignKeys []ForeignKeyColumn

	err := db.Raw(`
		SELECT
			fk.object_id AS ForeignKeyObjectID,
			pc.name AS ColumnName,
			fkc.constraint_column_id AS KeyOrdinal,
			rsch.name AS ReferencedSchema,
			rt.name AS ReferencedTable,
			rc.name AS ReferencedColumn,
			fk.delete_referential_action_desc AS DeleteAction,
			fk.update_referential_action_desc AS UpdateAction
		FROM sys.foreign_keys fk
		JOIN sys.foreign_key_columns fkc
			ON fkc.constraint_object_id = fk.object_id
		JOIN sys.columns pc
			ON pc.object_id = fkc.parent_object_id
			AND pc.column_id = fkc.parent_column_id
		JOIN sys.tables rt
			ON rt.object_id = fkc.referenced_object_id
		JOIN sys.schemas rsch
			ON rsch.schema_id = rt.schema_id
		JOIN sys.columns rc
			ON rc.object_id = fkc.referenced_object_id
			AND rc.column_id = fkc.referenced_column_id
		WHERE fk.parent_object_id = OBJECT_ID(?, 'U')
		ORDER BY fk.object_id, fkc.constraint_column_id;
	`, tableName).Scan(&foreignKeys).Error

	if err != nil {
		panic(err)
	}

	return foreignKeys
}
