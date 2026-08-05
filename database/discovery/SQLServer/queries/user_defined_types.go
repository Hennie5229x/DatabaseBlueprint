package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerUserDefinedTypes(db *gorm.DB) []models.UserDefinedType {
	var types []models.UserDefinedType

	err := db.Raw(`
		SELECT      s.name AS SchemaName,
					t.name AS TypeName,
					bt.name AS BaseTypeName,
					t.max_length AS MaxLength,
					t.precision AS Precision,
					t.scale AS Scale,
					t.is_nullable AS IsNullable
		FROM        sys.types AS t
		INNER JOIN  sys.schemas AS s ON s.schema_id = t.schema_id
		INNER JOIN  sys.types AS bt ON bt.user_type_id = t.system_type_id
					AND bt.is_user_defined = 0
		WHERE       t.is_user_defined = 1
		AND         t.is_table_type = 0
		AND         t.is_assembly_type = 0
		ORDER BY    s.name, t.name;
	`).Scan(&types).Error

	if err != nil {
		panic(err)
	}

	return types
}
