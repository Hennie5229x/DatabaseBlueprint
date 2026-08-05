package sqlserver

import (
	sqlserver_models "blueprint/database/discovery/SQLServer/models"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func GenerateCreateTableType(tableType sqlserver_models.UserDefinedTableType, columns []sqlserver_models.UserDefinedTableTypeColumn, keys []sqlserver_models.UserDefinedTableTypeKeyColumn, checks []sqlserver_models.UserDefinedTableTypeCheckConstraint, indexes []sqlserver_models.UserDefinedTableTypeIndexColumn) string {
	sections := make([]string, 0, 4)

	if definitions := tableTypeColumnDefinitions(columns); len(definitions) > 0 {
		sections = append(sections, strings.Join(definitions, ",\n"))
	}
	if definitions := tableTypeKeyDefinitions(keys); len(definitions) > 0 {
		sections = append(sections, strings.Join(definitions, ",\n\n"))
	}
	if definitions := tableTypeCheckDefinitions(checks); len(definitions) > 0 {
		sections = append(sections, strings.Join(definitions, ",\n\n"))
	}
	if definitions := tableTypeIndexDefinitions(indexes); len(definitions) > 0 {
		sections = append(sections, strings.Join(definitions, ",\n\n"))
	}

	definition := fmt.Sprintf(
		"CREATE TYPE %s.%s AS TABLE\n(\n%s\n)",
		quoteSqlServerIdentifier(tableType.SchemaName),
		quoteSqlServerIdentifier(tableType.TypeName),
		strings.Join(sections, ",\n\n"),
	)

	if tableType.IsMemoryOptimized {
		definition += "\nWITH\n(\n    MEMORY_OPTIMIZED = ON\n)"
	}

	return definition + ";"
}

func tableTypeColumnDefinitions(columns []sqlserver_models.UserDefinedTableTypeColumn) []string {
	sortedColumns := append([]sqlserver_models.UserDefinedTableTypeColumn(nil), columns...)
	sort.Slice(sortedColumns, func(i, j int) bool { return sortedColumns[i].ColumnID < sortedColumns[j].ColumnID })

	definitions := make([]string, 0, len(sortedColumns))
	for _, column := range sortedColumns {
		definition := "    " + quoteSqlServerIdentifier(column.ColumnName)
		if column.IsComputed && column.ComputedDefinition != "" {
			definition += " AS " + column.ComputedDefinition
			if column.IsPersisted {
				definition += " PERSISTED"
			}
			definitions = append(definitions, definition)
			continue
		}

		definition += " " + tableTypeColumnDataType(column)
		if column.CollationName != "" {
			definition += " COLLATE " + quoteSqlServerIdentifier(column.CollationName)
		}
		if column.IsIdentity {
			definition += fmt.Sprintf(" IDENTITY(%s,%s)", sqlServerIdentityValue(column.IdentitySeed, "1"), sqlServerIdentityValue(column.IdentityIncrement, "1"))
		}
		if column.IsNullable {
			definition += " NULL"
		} else {
			definition += " NOT NULL"
		}
		if column.DefaultDefinition != "" {
			definition += " DEFAULT " + column.DefaultDefinition
		}
		definitions = append(definitions, definition)
	}

	return definitions
}

func tableTypeColumnDataType(column sqlserver_models.UserDefinedTableTypeColumn) string {
	dataType := column.DataTypeName
	switch strings.ToLower(dataType) {
	case "varchar", "char", "varbinary", "binary":
		if column.MaxLength == -1 {
			return strings.ToUpper(dataType) + "(MAX)"
		}
		return fmt.Sprintf("%s(%d)", strings.ToUpper(dataType), column.MaxLength)
	case "nvarchar", "nchar":
		if column.MaxLength == -1 {
			return strings.ToUpper(dataType) + "(MAX)"
		}
		return fmt.Sprintf("%s(%d)", strings.ToUpper(dataType), column.MaxLength/2)
	case "decimal", "numeric":
		return fmt.Sprintf("%s(%d,%d)", strings.ToUpper(dataType), column.Precision, column.Scale)
	case "datetime2", "datetimeoffset", "time":
		return fmt.Sprintf("%s(%d)", strings.ToUpper(dataType), column.Scale)
	case "timestamp":
		return "ROWVERSION"
	default:
		return dataType
	}
}

type tableTypeKeyGroup struct {
	ConstraintName string
	ConstraintType string
	IndexType      string
	BucketCount    int
	Columns        []orderedColumn
}

func tableTypeKeyDefinitions(keys []sqlserver_models.UserDefinedTableTypeKeyColumn) []string {
	groupsByID := make(map[int]*tableTypeKeyGroup)
	constraintIDs := make([]int, 0)
	for _, key := range keys {
		group := groupsByID[key.ConstraintObjectID]
		if group == nil {
			group = &tableTypeKeyGroup{
				ConstraintName: key.ConstraintName,
				ConstraintType: key.ConstraintType,
				IndexType:      key.IndexType,
				BucketCount:    key.BucketCount,
			}
			groupsByID[key.ConstraintObjectID] = group
			constraintIDs = append(constraintIDs, key.ConstraintObjectID)
		}
		group.Columns = append(group.Columns, orderedColumn{ColumnName: key.ColumnName, Ordinal: key.KeyOrdinal, IsDescending: key.IsDescending})
	}

	sort.Ints(constraintIDs)
	definitions := make([]string, 0, len(constraintIDs))
	for _, constraintID := range constraintIDs {
		group := groupsByID[constraintID]
		sort.Slice(group.Columns, func(i, j int) bool { return group.Columns[i].Ordinal < group.Columns[j].Ordinal })
		columns := make([]string, 0, len(group.Columns))
		for _, column := range group.Columns {
			columns = append(columns, "        "+sqlServerIndexedColumn(column.ColumnName, column.IsDescending))
		}

		definition := fmt.Sprintf("    CONSTRAINT %s %s %s\n    (\n%s\n    )", quoteSqlServerIdentifier(group.ConstraintName), group.ConstraintType, group.IndexType, strings.Join(columns, ",\n"))
		if group.BucketCount > 0 {
			definition += "\n    WITH (BUCKET_COUNT = " + strconv.Itoa(group.BucketCount) + ")"
		}
		definitions = append(definitions, definition)
	}

	return definitions
}

func tableTypeCheckDefinitions(checks []sqlserver_models.UserDefinedTableTypeCheckConstraint) []string {
	sortedChecks := append([]sqlserver_models.UserDefinedTableTypeCheckConstraint(nil), checks...)
	sort.Slice(sortedChecks, func(i, j int) bool { return sortedChecks[i].ConstraintObjectID < sortedChecks[j].ConstraintObjectID })

	definitions := make([]string, 0, len(sortedChecks))
	for _, check := range sortedChecks {
		definitions = append(definitions, fmt.Sprintf("    CONSTRAINT %s CHECK %s", quoteSqlServerIdentifier(check.ConstraintName), check.Definition))
	}

	return definitions
}

func tableTypeIndexDefinitions(indexes []sqlserver_models.UserDefinedTableTypeIndexColumn) []string {
	groupsByID := make(map[int][]sqlserver_models.UserDefinedTableTypeIndexColumn)
	indexIDs := make([]int, 0)
	for _, index := range indexes {
		if _, ok := groupsByID[index.IndexID]; !ok {
			indexIDs = append(indexIDs, index.IndexID)
		}
		groupsByID[index.IndexID] = append(groupsByID[index.IndexID], index)
	}
	sort.Ints(indexIDs)

	definitions := make([]string, 0, len(indexIDs))
	for _, indexID := range indexIDs {
		indexColumns := groupsByID[indexID]
		sort.Slice(indexColumns, func(i, j int) bool {
			if indexColumns[i].IsIncluded != indexColumns[j].IsIncluded {
				return !indexColumns[i].IsIncluded
			}
			return indexColumns[i].KeyOrdinal < indexColumns[j].KeyOrdinal
		})

		first := indexColumns[0]
		keyColumns := make([]string, 0, len(indexColumns))
		includeColumns := make([]string, 0, len(indexColumns))
		for _, index := range indexColumns {
			if index.IsIncluded {
				includeColumns = append(includeColumns, "        "+quoteSqlServerIdentifier(index.ColumnName))
				continue
			}
			keyColumns = append(keyColumns, "        "+sqlServerIndexedColumn(index.ColumnName, index.IsDescending))
		}

		definition := fmt.Sprintf("    INDEX %s %s\n    (\n%s\n    )", quoteSqlServerIdentifier(first.IndexName), first.IndexType, strings.Join(keyColumns, ",\n"))
		if len(includeColumns) > 0 {
			definition += "\n    INCLUDE\n    (\n" + strings.Join(includeColumns, ",\n") + "\n    )"
		}
		if first.BucketCount > 0 {
			definition += "\n    WITH (BUCKET_COUNT = " + strconv.Itoa(first.BucketCount) + ")"
		}
		definitions = append(definitions, definition)
	}

	return definitions
}
