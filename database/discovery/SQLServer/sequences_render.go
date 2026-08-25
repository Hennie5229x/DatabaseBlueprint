package sqlserver

import (
	sqlserver_models "blueprint/database/discovery/SQLServer/models"
	"fmt"
	"strings"
)

func GenerateCreateSequence(sequence sqlserver_models.Sequences) string {
	definition := fmt.Sprintf(
		"CREATE SEQUENCE %s.%s\n    AS %s\n    START WITH %s\n    INCREMENT BY %s\n    MINVALUE %s\n    MAXVALUE %s",
		quoteSqlServerIdentifier(sequence.SchemaName),
		quoteSqlServerIdentifier(sequence.SequenceName),
		sequenceDataType(sequence),
		sequence.StartValue,
		sequence.IncrementBy,
		sequence.MinValue,
		sequence.MaxValue,
	)

	if sequence.IsCycling {
		definition += "\n    CYCLE"
	} else {
		definition += "\n    NO CYCLE"
	}

	if sequence.IsCached {
		definition += "\n    CACHE"
		if sequence.CacheSize != nil {
			definition += fmt.Sprintf(" %d", *sequence.CacheSize)
		}
	} else {
		definition += "\n    NO CACHE"
	}

	return definition + ";"
}

func sequenceDataType(sequence sqlserver_models.Sequences) string {
	if !strings.EqualFold(sequence.DataTypeSchemaName, "sys") {
		return quoteSqlServerIdentifier(sequence.DataTypeSchemaName) + "." + quoteSqlServerIdentifier(sequence.DataType)
	}

	dataType := strings.ToLower(sequence.DataType)
	if dataType == "decimal" || dataType == "numeric" {
		return fmt.Sprintf("%s(%d,%d)", strings.ToUpper(dataType), sequence.Precision, sequence.Scale)
	}

	return strings.ToUpper(dataType)
}
