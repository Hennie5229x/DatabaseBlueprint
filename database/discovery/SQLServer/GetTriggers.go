package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"

	"gorm.io/gorm"
)

func Triggers(db *gorm.DB, directory string) {
	var triggers []sqlservermodels.Triggers

	triggerSpinner := spinner.New("Triggers", "Discovering triggers")
	triggers = queries.SqlServerTriggers(db)

	if len(triggers) == 0 {
		triggerSpinner.Stop("No triggers found")
		return
	}

	err := ScriptTriggers(*db, directory, triggers, func(index int, total int, trigger sqlservermodels.Triggers) {
		triggerSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, total, trigger.SchemaName, trigger.TriggerName))
	})
	if err != nil {
		triggerSpinner.Stop("Failed")
		panic(err)
	}

	triggerSpinner.Stop(fmt.Sprintf("%d triggers scripted", len(triggers)))
}
