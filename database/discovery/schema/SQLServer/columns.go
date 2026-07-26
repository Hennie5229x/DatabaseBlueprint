package sqlserver

import (
	"gorm.io/gorm"
)

func SqlServerColumns(db *gorm.DB, tableName string) []Column {
	var columns []Column

	err := db.Raw(`
		SELECT		c.column_id AS ColumnID,
					c.name AS ColumnName,
					ty.name AS DataType,
					c.max_length AS MaxLength,
					c.precision AS Precision,
					c.scale AS Scale,
					c.is_nullable AS IsNullable,
					c.is_identity AS IsIdentity,
					CONVERT(varchar(100), ic.seed_value) AS IdentitySeed,
					CONVERT(varchar(100), ic.increment_value) AS IdentityIncrement,
					cc.definition AS ComputedDefinition,
					ISNULL(cc.is_persisted, 0) AS IsPersisted
		FROM 		sys.columns c
		JOIN 		sys.types ty ON c.user_type_id = ty.user_type_id
		LEFT JOIN 	sys.identity_columns ic	ON ic.object_id = c.object_id
					AND ic.column_id = c.column_id
		LEFT JOIN 	sys.computed_columns cc	ON cc.object_id = c.object_id
					AND cc.column_id = c.column_id
		WHERE 		c.object_id = OBJECT_ID(?, 'U')
		ORDER BY 	c.column_id;
	`, tableName).Scan(&columns).Error

	if err != nil {
		panic(err)
	}

	return columns
}
