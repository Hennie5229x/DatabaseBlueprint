package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerForeignKeys(db *gorm.DB) []models.ForeignKeyColumn {
	var foreignKeys []models.ForeignKeyColumn

	err := db.Raw(`
		SELECT		fk.name AS ForeignKeyName,
					fk.object_id AS ForeignKeyObjectID,
					psch.name AS ParentSchema,
					pt.name AS ParentTable,
					pc.name AS ColumnName,
					fkc.constraint_column_id AS KeyOrdinal,
					rsch.name AS ReferencedSchema,
					rt.name AS ReferencedTable,
					rc.name AS ReferencedColumn,
					fk.delete_referential_action_desc AS DeleteAction,
					fk.update_referential_action_desc AS UpdateAction
		FROM 		sys.foreign_keys fk
		JOIN 		sys.foreign_key_columns fkc ON fkc.constraint_object_id = fk.object_id
		JOIN 		sys.tables pt ON pt.object_id = fkc.parent_object_id
		JOIN 		sys.schemas psch ON psch.schema_id = pt.schema_id
		JOIN 		sys.columns pc ON pc.object_id = fkc.parent_object_id
					AND pc.column_id = fkc.parent_column_id
		JOIN 		sys.tables rt ON rt.object_id = fkc.referenced_object_id
		JOIN 		sys.schemas rsch ON rsch.schema_id = rt.schema_id
		JOIN 		sys.columns rc ON rc.object_id = fkc.referenced_object_id
					AND rc.column_id = fkc.referenced_column_id
		ORDER BY 	fk.object_id, fkc.constraint_column_id;
	`).Scan(&foreignKeys).Error

	if err != nil {
		panic(err)
	}

	return foreignKeys
}
