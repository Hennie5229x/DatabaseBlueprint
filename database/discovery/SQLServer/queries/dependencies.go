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
									'V',   -- View
									'FN',  -- Scalar function
									'IF',  -- Inline table-valued function
									'TF',  -- Multi-statement table-valued function
									'P'    -- Stored procedure
								)
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
					AND dependency.referencing_minor_id = 0
		LEFT JOIN 	ScriptableObjects AS referenced	ON dependency.referenced_id = referenced.object_id
		ORDER BY	referencing.SchemaName,	referencing.ObjectName,	referenced.SchemaName, referenced.ObjectName;
	`).Scan(&dependencies).Error

	if err != nil {
		return nil, err
	}

	return dependencies, nil
}
