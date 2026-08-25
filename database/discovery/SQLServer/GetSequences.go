package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"fmt"

	"gorm.io/gorm"
)

func Sequences(db *gorm.DB, directory string) {
	var sequences []sqlservermodels.Sequences

	sequencesSpinner := spinner.New("Sequences", "Discovering sequences")
	sequences = queries.SqlServerSequences(db)

	if len(sequences) == 0 {
		sequencesSpinner.Stop("No sequences found")
		return
	}

	err := ScriptSequences(*db, directory, sequences, func(index int, total int, sequence sqlservermodels.Sequences) {
		sequencesSpinner.Update(fmt.Sprintf("[%d/%d] %s.%s", index+1, total, sequence.SchemaName, sequence.SequenceName))
	})
	if err != nil {
		sequencesSpinner.Stop("Failed")
		panic(err)
	}

	sequencesSpinner.Stop(fmt.Sprintf("%d sequences scripted", len(sequences)))
}
