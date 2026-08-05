package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerUserDefinedTableTypeChecks(db *gorm.DB, schemaName string, typeName string) []models.UserDefinedTableTypeCheckConstraint {
	var checks []models.UserDefinedTableTypeCheckConstraint

	err := db.Raw(`
		SELECT      s.name AS SchemaName,
					tt.name AS TypeName,
					cc.object_id AS ConstraintObjectID,
					cc.name AS ConstraintName,
					cc.parent_column_id AS ParentColumnID,
					cc.definition AS Definition
		FROM        sys.table_types AS tt
		INNER JOIN  sys.schemas AS s ON s.schema_id = tt.schema_id
		INNER JOIN  sys.check_constraints AS cc ON cc.parent_object_id = tt.type_table_object_id
		WHERE       s.name = ?
		AND         tt.name = ?
		ORDER BY    cc.object_id;
	`, schemaName, typeName).Scan(&checks).Error

	if err != nil {
		panic(err)
	}

	return checks
}
