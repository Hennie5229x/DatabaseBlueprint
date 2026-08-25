package sqlserver

import (
	sqlserver_models "blueprint/database/discovery/SQLServer/models"
	"fmt"
	"strings"
)

func GenerateCreateTrigger(trigger sqlserver_models.Triggers) string {
	definition := strings.TrimSpace(trigger.Definition)
	if trigger.IsDisabled {
		definition += fmt.Sprintf(
			"\n\nDISABLE TRIGGER %s.%s ON %s.%s;",
			quoteSqlServerIdentifier(trigger.SchemaName),
			quoteSqlServerIdentifier(trigger.TriggerName),
			quoteSqlServerIdentifier(trigger.TableSchemaName),
			quoteSqlServerIdentifier(trigger.TableName),
		)
	}

	return definition
}
