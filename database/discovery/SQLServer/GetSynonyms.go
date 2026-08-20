package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"

	"gorm.io/gorm"
)

func Synonyms(db *gorm.DB, directory string) {
	var synonyms []sqlservermodels.Synonyms

	synonymsSpinner := spinner.New("Synonyms", "Discovering synonyms")
	synonyms = queries.SqlServerSynonyms(db)

	if len(synonyms) == 0 {
		synonymsSpinner.Stop("No synonyms found")
		return
	}
	err := ScriptSynonyms(*db, directory, synonyms, func(index int, total int, synonyms sqlservermodels.Synonyms) {
		synonymsSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, total, synonyms.SchemaName, synonyms.SynonymName))
	})
	if err != nil {
		synonymsSpinner.Stop("Failed")
		panic(err)
	}
	synonymsSpinner.Stop(fmt.Sprintf("%d synonyms scripted", len(synonyms)))
}
