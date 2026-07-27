package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerDefaultConstraints(db *gorm.DB, tableName string) []models.DefaultConstraint {
	var constraints []models.DefaultConstraint

	err := db.Raw(`
		SELECT		dc.parent_column_id AS ColumnID,
					c.name AS ColumnName,
					dc.name AS ConstraintName,
					dc.definition AS ConstraintValue
		FROM 		sys.default_constraints dc
		JOIN 		sys.columns c ON c.object_id = dc.parent_object_id
					AND c.column_id = dc.parent_column_id
		WHERE 		dc.parent_object_id = OBJECT_ID(?, 'U')
		ORDER BY 	dc.parent_column_id;
	`, tableName).Scan(&constraints).Error

	if err != nil {
		panic(err)
	}

	return constraints
}
