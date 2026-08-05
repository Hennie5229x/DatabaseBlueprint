package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerUserDefinedTableTypeColumns(db *gorm.DB, schemaName string, typeName string) []models.UserDefinedTableTypeColumn {
	var columns []models.UserDefinedTableTypeColumn

	err := db.Raw(`
		SELECT      s.name AS SchemaName,
					tt.name AS TypeName,
					c.column_id AS ColumnID,
					c.name AS ColumnName,
					CASE
						WHEN column_type.is_user_defined = 1
						THEN QUOTENAME(type_schema.name) + '.' + QUOTENAME(column_type.name)
						ELSE column_type.name
					END AS DataTypeName,
					c.max_length AS MaxLength,
					c.precision AS Precision,
					c.scale AS Scale,
					c.collation_name AS CollationName,
					c.is_nullable AS IsNullable,
					c.is_identity AS IsIdentity,
					c.is_computed AS IsComputed,
					CONVERT(varchar(100), identity_column.seed_value) AS IdentitySeed,
					CONVERT(varchar(100), identity_column.increment_value) AS IdentityIncrement,
					computed_column.definition AS ComputedDefinition,
					ISNULL(computed_column.is_persisted, 0) AS IsPersisted,
					default_constraint.name AS DefaultConstraintName,
					default_constraint.definition AS DefaultDefinition
		FROM        sys.table_types AS tt
		INNER JOIN  sys.schemas AS s ON s.schema_id = tt.schema_id
		INNER JOIN  sys.columns AS c ON c.object_id = tt.type_table_object_id
		INNER JOIN  sys.types AS column_type ON column_type.user_type_id = c.user_type_id
		INNER JOIN  sys.schemas AS type_schema ON type_schema.schema_id = column_type.schema_id
		LEFT JOIN   sys.identity_columns AS identity_column ON identity_column.object_id = c.object_id
					AND identity_column.column_id = c.column_id
		LEFT JOIN   sys.computed_columns AS computed_column ON computed_column.object_id = c.object_id
					AND computed_column.column_id = c.column_id
		LEFT JOIN   sys.default_constraints AS default_constraint ON default_constraint.parent_object_id = c.object_id
					AND default_constraint.parent_column_id = c.column_id
		WHERE       s.name = ?
		AND         tt.name = ?
		ORDER BY    s.name, tt.name, c.column_id;
	`, schemaName, typeName).Scan(&columns).Error

	if err != nil {
		panic(err)
	}

	return columns
}
