package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerUserDefinedTableTypes(db *gorm.DB) []models.UserDefinedTableType {
	var tableTypes []models.UserDefinedTableType

	err := db.Raw(`
		SELECT      s.name AS SchemaName,
					tt.name AS TypeName,
					tt.is_memory_optimized AS IsMemoryOptimized
		FROM        sys.table_types AS tt
		JOIN        sys.schemas AS s ON s.schema_id = tt.schema_id
		ORDER BY    s.name, tt.name;
	`).Scan(&tableTypes).Error

	if err != nil {
		panic(err)
	}

	return tableTypes
}
