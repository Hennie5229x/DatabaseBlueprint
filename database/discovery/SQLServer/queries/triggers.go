package queries

import (
	models "blueprint/database/discovery/SQLServer/models"

	"gorm.io/gorm"
)

func SqlServerTriggers(db *gorm.DB) []models.Triggers {
	var triggers []models.Triggers

	err := db.Raw(`
		SELECT      trg_sch.name AS SchemaName,
		            tbl_sch.name AS TableSchemaName,
		            tbl.name AS TableName,
		            trg.name AS TriggerName,
		            sm.definition AS Definition,
		            trg.is_instead_of_trigger AS IsInsteadOf,
		            trg.is_disabled AS IsDisabled,
		            trg.is_not_for_replication AS IsNotForReplication
		FROM        sys.triggers trg
		JOIN        sys.objects trg_obj ON trg.object_id = trg_obj.object_id
		JOIN        sys.schemas trg_sch ON trg_obj.schema_id = trg_sch.schema_id
		JOIN        sys.tables tbl ON trg.parent_id = tbl.object_id
		JOIN        sys.schemas tbl_sch ON tbl.schema_id = tbl_sch.schema_id
		JOIN        sys.sql_modules sm ON trg.object_id = sm.object_id
		WHERE       trg.parent_class = 1
		AND         trg.is_ms_shipped = 0
		AND         tbl.is_ms_shipped = 0
		ORDER BY    trg_sch.name, trg.name;
	`).Scan(&triggers).Error

	if err != nil {
		panic(err)
	}

	return triggers
}
