package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerDependencies(db *gorm.DB) ([]models.DependencyRow, error) {
	var dependencies []models.DependencyRow

	err := db.Raw(`
		WITH ScriptableObjects AS
		(
			SELECT	o.object_id,
					o.schema_id,
					s.name AS SchemaName,
					o.name AS ObjectName,
					o.type AS ObjectTypeCode,
					o.type_desc AS ObjectType
			FROM 	sys.objects AS o
			JOIN 	sys.schemas AS s ON o.schema_id = s.schema_id
			WHERE 	o.type IN	(
									'U',   -- Table
									'F',   -- Foreign key constraint
									'V',   -- View
									'FN',  -- Scalar function
									'IF',  -- Inline table-valued function
									'TF',  -- Multi-statement table-valued function
									'P',   -- Stored procedure
									'TT',  -- User-defined table type
									'SO',  -- Sequence object
									'SN',  -- Synonym
									'TR'   -- DML trigger
								)
			AND 	o.is_ms_shipped = 0
		)
		SELECT 		DISTINCT
					ReferencingObjectId   = referencing.object_id,
					ReferencingSchema     = referencing.SchemaName,
					ReferencingObject     = referencing.ObjectName,
					ReferencingTypeCode   = referencing.ObjectTypeCode,
					ReferencingType       = referencing.ObjectType,

					ReferencedObjectId    = referenced.object_id,
					ReferencedSchema      = referenced.SchemaName,
					ReferencedObject      = referenced.ObjectName,
					ReferencedTypeCode    = referenced.ObjectTypeCode,
					ReferencedType        = referenced.ObjectType
		FROM 		ScriptableObjects AS referencing
		LEFT JOIN 	sys.sql_expression_dependencies AS dependency ON dependency.referencing_id = referencing.object_id
		LEFT JOIN 	ScriptableObjects AS referenced	ON dependency.referenced_id = referenced.object_id

		UNION

		SELECT 		DISTINCT
					ReferencingObjectId   = parent_table.object_id,
					ReferencingSchema     = parent_table.SchemaName,
					ReferencingObject     = parent_table.ObjectName,
					ReferencingTypeCode   = parent_table.ObjectTypeCode,
					ReferencingType       = parent_table.ObjectType,

					ReferencedObjectId    = referenced.object_id,
					ReferencedSchema      = referenced.SchemaName,
					ReferencedObject      = referenced.ObjectName,
					ReferencedTypeCode    = referenced.ObjectTypeCode,
					ReferencedType        = referenced.ObjectType
		FROM 		sys.objects AS child_object
		JOIN 		ScriptableObjects AS parent_table ON child_object.parent_object_id = parent_table.object_id
					AND parent_table.ObjectTypeCode = 'U'
		JOIN 		sys.sql_expression_dependencies AS dependency ON dependency.referencing_id = child_object.object_id
		JOIN 		ScriptableObjects AS referenced ON dependency.referenced_id = referenced.object_id
		WHERE 		child_object.type IN ('D', 'C')

		UNION

		SELECT 		DISTINCT
					ReferencingObjectId   = foreign_key.object_id,
					ReferencingSchema     = foreign_key.SchemaName,
					ReferencingObject     = foreign_key.ObjectName,
					ReferencingTypeCode   = foreign_key.ObjectTypeCode,
					ReferencingType       = foreign_key.ObjectType,

					ReferencedObjectId    = parent_table.object_id,
					ReferencedSchema      = parent_table.SchemaName,
					ReferencedObject      = parent_table.ObjectName,
					ReferencedTypeCode    = parent_table.ObjectTypeCode,
					ReferencedType        = parent_table.ObjectType
		FROM 		sys.foreign_keys AS fk
		JOIN 		ScriptableObjects AS foreign_key ON fk.object_id = foreign_key.object_id
		JOIN 		ScriptableObjects AS parent_table ON fk.parent_object_id = parent_table.object_id

		UNION

		SELECT 		DISTINCT
					ReferencingObjectId   = foreign_key.object_id,
					ReferencingSchema     = foreign_key.SchemaName,
					ReferencingObject     = foreign_key.ObjectName,
					ReferencingTypeCode   = foreign_key.ObjectTypeCode,
					ReferencingType       = foreign_key.ObjectType,

					ReferencedObjectId    = referenced_table.object_id,
					ReferencedSchema      = referenced_table.SchemaName,
					ReferencedObject      = referenced_table.ObjectName,
					ReferencedTypeCode    = referenced_table.ObjectTypeCode,
					ReferencedType        = referenced_table.ObjectType
		FROM 		sys.foreign_keys AS fk
		JOIN 		ScriptableObjects AS foreign_key ON fk.object_id = foreign_key.object_id
		JOIN 		ScriptableObjects AS referenced_table ON fk.referenced_object_id = referenced_table.object_id
		ORDER BY	ReferencingSchema, ReferencingObject, ReferencedSchema, ReferencedObject;
	`).Scan(&dependencies).Error

	if err != nil {
		return nil, err
	}

	return dependencies, nil
}
