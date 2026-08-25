package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerSequences(db *gorm.DB) []models.Sequences {
	var sequences []models.Sequences

	err := db.Raw(`
		SELECT      sch.name AS SchemaName,
		            seq.name AS SequenceName,
		            typ_sch.name AS DataTypeSchemaName,
		            typ.name AS DataType,
		            seq.precision AS Precision,
		            seq.scale AS Scale,
		            CONVERT(nvarchar(128), seq.start_value) AS StartValue,
		            CONVERT(nvarchar(128), seq.increment) AS IncrementBy,
		            CONVERT(nvarchar(128), seq.minimum_value) AS MinValue,
		            CONVERT(nvarchar(128), seq.maximum_value) AS MaxValue,
		            seq.is_cycling AS IsCycling,
		            seq.is_cached AS IsCached,
		            seq.cache_size AS CacheSize
		FROM        sys.sequences seq
		JOIN        sys.schemas sch ON seq.schema_id = sch.schema_id
		JOIN        sys.types typ ON seq.user_type_id = typ.user_type_id
		JOIN        sys.schemas typ_sch ON typ.schema_id = typ_sch.schema_id
		ORDER BY    sch.name, seq.name;
	`).Scan(&sequences).Error

	if err != nil {
		panic(err)
	}

	return sequences
}
