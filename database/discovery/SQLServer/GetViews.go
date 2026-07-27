package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"

	"gorm.io/gorm"
)

func Views(db *gorm.DB, directory string) {
	var views []sqlservermodels.Views

	viewSpinner := spinner.New("Views", "Discovering views")
	views = queries.SqlServerViews(*db)

	if len(views) == 0 {
		viewSpinner.Stop("No views found")
		return
	}
	err := ScriptViews(*db, directory, views, func(index int, total int, view sqlservermodels.Views) {
		viewSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, total, view.Schema, view.View))
	})
	if err != nil {
		viewSpinner.Stop("Failed")
		panic(err)
	}
	viewSpinner.Stop(fmt.Sprintf("%d views scripted", len(views)))
}
