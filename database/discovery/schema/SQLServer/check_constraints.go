package sqlserver

import (
	"gorm.io/gorm"
)

func SqlServerCheckConstraints(db *gorm.DB, tableName string) []CheckConstraint {
	var checkConstraints []CheckConstraint

	err := db.Raw(`
		SELECT		cc.object_id AS ConstraintObjectID,
					cc.definition AS Definition
		FROM 		sys.check_constraints cc
		WHERE 		cc.parent_object_id = OBJECT_ID(?, 'U')
		ORDER BY 	cc.object_id;
	`, tableName).Scan(&checkConstraints).Error

	if err != nil {
		panic(err)
	}

	return checkConstraints
}
