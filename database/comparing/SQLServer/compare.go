package sqlserver

import (
	"blueprint/cli/spinner"
	sqlserverModels "blueprint/database/discovery/SQLServer/models"
	sqlserverQueries "blueprint/database/discovery/SQLServer/queries"
	discoveryModels "blueprint/database/discovery/models"
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

var activeComparison *comparisonSummary

func Init(dbSource *gorm.DB, dbTarget *gorm.DB, outputPath string) {
	comparison := newComparisonSummary()
	activeComparison = comparison
	defer func() { activeComparison = nil }()

	compareWithSpinner(comparison, "Schemas", "Schema", func() { compareSchemas(dbSource, dbTarget) })
	compareWithSpinner(comparison, "User Types", "User-defined type", func() { compareUserDefinedTypes(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Table Types", "Table type", func() { compareTableTypes(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Type Columns", "Table type column", func() { compareTableTypeColumns(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Type Keys", "Table type key", func() { compareTableTypeKeys(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Type Checks", "Table type check", func() { compareTableTypeChecks(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Type Indexes", "Table type index", func() { compareTableTypeIndexes(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Sequences", "Sequence", func() { compareSequences(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Synonyms", "Synonym", func() { compareSynonyms(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Tables", "Table", func() { compareTables(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Columns", "Column", func() { compareColumns(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Defaults", "Default constraint", func() { compareDefaultConstraints(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Primary Keys", "Primary key", func() { comparePrimaryKeys(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Unique Constraints", "Unique constraint", func() { compareUniqueConstraints(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Check Constraints", "Check constraint", func() { compareCheckConstraints(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Indexes", "Index", func() { compareIndexes(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Foreign Keys", "Foreign key", func() { compareForeignKeys(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Views", "View", func() { compareViews(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Functions", "Function", func() { compareFunctions(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Procedures", "Procedure", func() { compareProcedures(dbSource, dbTarget) })
	compareWithSpinner(comparison, "Triggers", "Trigger", func() { compareTriggers(dbSource, dbTarget) })

	if err := GenerateMigration(dbSource, dbTarget, outputPath); err != nil {
		fmt.Printf("Failed to generate migration script: %v\n", err)
	}
}

type comparisonSummary struct {
	changes map[string]map[string]struct{}
}

func newComparisonSummary() *comparisonSummary {
	return &comparisonSummary{changes: make(map[string]map[string]struct{})}
}

func (summary *comparisonSummary) record(objectType string, key string) {
	if summary.changes[objectType] == nil {
		summary.changes[objectType] = make(map[string]struct{})
	}
	summary.changes[objectType][key] = struct{}{}
}

func (summary *comparisonSummary) count(objectType string) int {
	return len(summary.changes[objectType])
}

func compareWithSpinner(summary *comparisonSummary, label string, objectType string, compare func()) {
	comparisonSpinner := spinner.New(label, "Comparing")
	compare()
	comparisonSpinner.Stop(fmt.Sprintf("%s : %d changes", label, summary.count(objectType)))
}

func compareSchemas(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"Schema",
		sqlserverQueries.SqlServerSchemas(dbSource),
		sqlserverQueries.SqlServerSchemas(dbTarget),
		func(schema sqlserverModels.Schemas) string { return schema.Name },
		func(_, _ sqlserverModels.Schemas) bool { return true },
	)
}

func compareSequences(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"Sequence",
		sqlserverQueries.SqlServerSequences(dbSource),
		sqlserverQueries.SqlServerSequences(dbTarget),
		func(sequence sqlserverModels.Sequences) string {
			return sequence.SchemaName + "." + sequence.SequenceName
		},
		func(source, target sqlserverModels.Sequences) bool {
			return reflect.DeepEqual(source, target)
		},
	)
}

func compareViews(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"View",
		sqlserverQueries.SqlServerViews(*dbSource),
		sqlserverQueries.SqlServerViews(*dbTarget),
		func(view sqlserverModels.Views) string { return view.Schema + "." + view.View },
		func(source, target sqlserverModels.Views) bool {
			return definitionsEqual(source.Definition, target.Definition)
		},
	)
}

func compareFunctions(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"Function",
		sqlserverQueries.SqlServerFunctions(dbSource),
		sqlserverQueries.SqlServerFunctions(dbTarget),
		func(function sqlserverModels.Functions) string {
			return function.Schema + "." + function.Name
		},
		func(source, target sqlserverModels.Functions) bool {
			return definitionsEqual(source.Definition, target.Definition)
		},
	)
}

func compareProcedures(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"Procedure",
		sqlserverQueries.SqlServerProcedures(dbSource),
		sqlserverQueries.SqlServerProcedures(dbTarget),
		func(procedure sqlserverModels.Procedures) string {
			return procedure.Schema + "." + procedure.Name
		},
		func(source, target sqlserverModels.Procedures) bool {
			return definitionsEqual(source.Definition, target.Definition)
		},
	)
}

func compareTriggers(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"Trigger",
		sqlserverQueries.SqlServerTriggers(dbSource),
		sqlserverQueries.SqlServerTriggers(dbTarget),
		func(trigger sqlserverModels.Triggers) string {
			return trigger.SchemaName + "." + trigger.TriggerName
		},
		func(source, target sqlserverModels.Triggers) bool {
			return definitionsEqual(source.Definition, target.Definition) &&
				source.IsInsteadOf == target.IsInsteadOf &&
				source.IsDisabled == target.IsDisabled &&
				source.IsNotForReplication == target.IsNotForReplication
		},
	)
}

func compareSynonyms(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"Synonym",
		sqlserverQueries.SqlServerSynonyms(dbSource),
		sqlserverQueries.SqlServerSynonyms(dbTarget),
		func(synonym sqlserverModels.Synonyms) string {
			return synonym.SchemaName + "." + synonym.SynonymName
		},
		func(source, target sqlserverModels.Synonyms) bool {
			return source.BaseObjectName == target.BaseObjectName
		},
	)
}

func compareUserDefinedTypes(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"User-defined type",
		sqlserverQueries.SqlServerUserDefinedTypes(dbSource),
		sqlserverQueries.SqlServerUserDefinedTypes(dbTarget),
		func(userType sqlserverModels.UserDefinedType) string {
			return userType.SchemaName + "." + userType.TypeName
		},
		func(source, target sqlserverModels.UserDefinedType) bool {
			return reflect.DeepEqual(source, target)
		},
	)
}

func compareTableTypes(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"Table type",
		sqlserverQueries.SqlServerUserDefinedTableTypes(dbSource),
		sqlserverQueries.SqlServerUserDefinedTableTypes(dbTarget),
		func(tableType sqlserverModels.UserDefinedTableType) string {
			return tableType.SchemaName + "." + tableType.TypeName
		},
		func(source, target sqlserverModels.UserDefinedTableType) bool {
			return reflect.DeepEqual(source, target)
		},
	)
}

func compareTableTypeColumns(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareExistingTableTypes(dbSource, dbTarget, func(schemaName, typeName string) {
		objectKey := schemaName + "." + typeName
		compareObjects(
			"Table type column",
			sqlserverQueries.SqlServerUserDefinedTableTypeColumns(dbSource, schemaName, typeName),
			sqlserverQueries.SqlServerUserDefinedTableTypeColumns(dbTarget, schemaName, typeName),
			func(column sqlserverModels.UserDefinedTableTypeColumn) string {
				return objectKey + "." + column.ColumnName
			},
			func(source, target sqlserverModels.UserDefinedTableTypeColumn) bool {
				return reflect.DeepEqual(source, target)
			},
		)
	})
}

func compareTableTypeKeys(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareExistingTableTypes(dbSource, dbTarget, func(schemaName, typeName string) {
		objectKey := schemaName + "." + typeName
		compareObjects(
			"Table type key",
			sqlserverQueries.SqlServerUserDefinedTableTypeKeys(dbSource, schemaName, typeName),
			sqlserverQueries.SqlServerUserDefinedTableTypeKeys(dbTarget, schemaName, typeName),
			func(key sqlserverModels.UserDefinedTableTypeKeyColumn) string {
				return objectKey + "." + key.ConstraintName + "." + fmt.Sprint(key.KeyOrdinal)
			},
			func(source, target sqlserverModels.UserDefinedTableTypeKeyColumn) bool {
				return source.ConstraintName == target.ConstraintName &&
					source.ConstraintType == target.ConstraintType &&
					source.IndexName == target.IndexName &&
					source.IndexType == target.IndexType &&
					source.ColumnName == target.ColumnName &&
					source.KeyOrdinal == target.KeyOrdinal &&
					source.IsDescending == target.IsDescending &&
					source.BucketCount == target.BucketCount
			},
		)
	})
}

func compareTableTypeChecks(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareExistingTableTypes(dbSource, dbTarget, func(schemaName, typeName string) {
		objectKey := schemaName + "." + typeName
		compareObjects(
			"Table type check",
			sqlserverQueries.SqlServerUserDefinedTableTypeChecks(dbSource, schemaName, typeName),
			sqlserverQueries.SqlServerUserDefinedTableTypeChecks(dbTarget, schemaName, typeName),
			func(check sqlserverModels.UserDefinedTableTypeCheckConstraint) string {
				return objectKey + "." + check.ConstraintName
			},
			func(source, target sqlserverModels.UserDefinedTableTypeCheckConstraint) bool {
				return source.ConstraintName == target.ConstraintName &&
					source.ParentColumnID == target.ParentColumnID &&
					source.Definition == target.Definition
			},
		)
	})
}

func compareTableTypeIndexes(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareExistingTableTypes(dbSource, dbTarget, func(schemaName, typeName string) {
		objectKey := schemaName + "." + typeName
		compareObjects(
			"Table type index",
			sqlserverQueries.SqlServerUserDefinedTableTypeIndexes(dbSource, schemaName, typeName),
			sqlserverQueries.SqlServerUserDefinedTableTypeIndexes(dbTarget, schemaName, typeName),
			func(index sqlserverModels.UserDefinedTableTypeIndexColumn) string {
				return objectKey + "." + index.IndexName + "." + fmt.Sprint(index.KeyOrdinal) + "." + fmt.Sprint(index.IncludeOrder)
			},
			func(source, target sqlserverModels.UserDefinedTableTypeIndexColumn) bool {
				return source.IndexName == target.IndexName &&
					source.IndexType == target.IndexType &&
					source.IsUnique == target.IsUnique &&
					source.ColumnName == target.ColumnName &&
					source.KeyOrdinal == target.KeyOrdinal &&
					source.IsDescending == target.IsDescending &&
					source.IsIncluded == target.IsIncluded &&
					source.IncludeOrder == target.IncludeOrder &&
					source.HasFilter == target.HasFilter &&
					source.FilterDefinition == target.FilterDefinition &&
					source.BucketCount == target.BucketCount
			},
		)
	})
}

func compareTables(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"Table",
		sqlserverQueries.SqlServerTables(*dbSource),
		sqlserverQueries.SqlServerTables(*dbTarget),
		func(table discoveryModels.Tables) string {
			return table.Schema + "." + table.Name
		},
		func(_, _ discoveryModels.Tables) bool { return true },
	)
}

func compareColumns(dbSource *gorm.DB, dbTarget *gorm.DB) {
	sourceTables := sqlserverQueries.SqlServerTables(*dbSource)
	targetTables := sqlserverQueries.SqlServerTables(*dbTarget)

	targetTableMap := make(map[string]struct{}, len(targetTables))
	for _, table := range targetTables {
		targetTableMap[table.Schema+"."+table.Name] = struct{}{}
	}

	for _, table := range sourceTables {
		tableKey := table.Schema + "." + table.Name
		if _, found := targetTableMap[tableKey]; !found {
			continue
		}

		compareObjects(
			"Column",
			sqlserverQueries.SqlServerColumns(dbSource, tableKey),
			sqlserverQueries.SqlServerColumns(dbTarget, tableKey),
			func(column sqlserverModels.Column) string {
				return tableKey + "." + column.ColumnName
			},
			func(source, target sqlserverModels.Column) bool {
				return reflect.DeepEqual(source, target)
			},
		)
	}
}

func compareDefaultConstraints(dbSource *gorm.DB, dbTarget *gorm.DB) {
	sourceTables := sqlserverQueries.SqlServerTables(*dbSource)
	targetTables := sqlserverQueries.SqlServerTables(*dbTarget)

	targetTableMap := make(map[string]struct{}, len(targetTables))
	for _, table := range targetTables {
		targetTableMap[table.Schema+"."+table.Name] = struct{}{}
	}

	for _, table := range sourceTables {
		tableKey := table.Schema + "." + table.Name
		if _, found := targetTableMap[tableKey]; !found {
			continue
		}

		compareObjects(
			"Default constraint",
			sqlserverQueries.SqlServerDefaultConstraints(dbSource, tableKey),
			sqlserverQueries.SqlServerDefaultConstraints(dbTarget, tableKey),
			func(constraint sqlserverModels.DefaultConstraint) string {
				return tableKey + "." + constraint.ConstraintName
			},
			func(source, target sqlserverModels.DefaultConstraint) bool {
				return reflect.DeepEqual(source, target)
			},
		)
	}
}

func comparePrimaryKeys(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareExistingTables(dbSource, dbTarget, func(tableKey string) {
		compareObjects(
			"Primary key",
			sqlserverQueries.SqlServerPrimaryKeys(dbSource, tableKey),
			sqlserverQueries.SqlServerPrimaryKeys(dbTarget, tableKey),
			func(key sqlserverModels.PrimaryKeyColumn) string {
				return constraintKey(tableKey, key.ConstraintName, key.IsSystemNamed, key.ColumnName, key.KeyOrdinal)
			},
			func(source, target sqlserverModels.PrimaryKeyColumn) bool {
				return sameConstraintName(source.ConstraintName, source.IsSystemNamed, target.ConstraintName, target.IsSystemNamed) &&
					source.IsSystemNamed == target.IsSystemNamed &&
					source.IndexType == target.IndexType &&
					source.ColumnName == target.ColumnName &&
					source.KeyOrdinal == target.KeyOrdinal &&
					source.IsDescending == target.IsDescending
			},
		)
	})
}

func compareUniqueConstraints(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareExistingTables(dbSource, dbTarget, func(tableKey string) {
		compareObjects(
			"Unique constraint",
			sqlserverQueries.SqlServerUniqueConstraints(dbSource, tableKey),
			sqlserverQueries.SqlServerUniqueConstraints(dbTarget, tableKey),
			func(constraint sqlserverModels.UniqueConstraintColumn) string {
				return constraintKey(tableKey, constraint.ConstraintName, constraint.IsSystemNamed, constraint.ColumnName, constraint.KeyOrdinal)
			},
			func(source, target sqlserverModels.UniqueConstraintColumn) bool {
				return sameConstraintName(source.ConstraintName, source.IsSystemNamed, target.ConstraintName, target.IsSystemNamed) &&
					source.IsSystemNamed == target.IsSystemNamed &&
					source.IndexType == target.IndexType &&
					source.ColumnName == target.ColumnName &&
					source.KeyOrdinal == target.KeyOrdinal &&
					source.IsDescending == target.IsDescending
			},
		)
	})
}

func compareCheckConstraints(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareExistingTables(dbSource, dbTarget, func(tableKey string) {
		compareObjects(
			"Check constraint",
			sqlserverQueries.SqlServerCheckConstraints(dbSource, tableKey),
			sqlserverQueries.SqlServerCheckConstraints(dbTarget, tableKey),
			func(constraint sqlserverModels.CheckConstraint) string {
				return tableKey + "." + constraint.ConstraintName
			},
			func(source, target sqlserverModels.CheckConstraint) bool {
				return source.ConstraintName == target.ConstraintName &&
					source.Definition == target.Definition
			},
		)
	})
}

func compareIndexes(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareExistingTables(dbSource, dbTarget, func(tableKey string) {
		compareObjects(
			"Index",
			sqlserverQueries.SqlServerIndexes(dbSource, tableKey),
			sqlserverQueries.SqlServerIndexes(dbTarget, tableKey),
			func(index sqlserverModels.IndexColumn) string {
				return tableKey + "." + index.IndexName + "." + fmt.Sprint(index.KeyOrdinal) + "." + fmt.Sprint(index.IncludeOrder)
			},
			func(source, target sqlserverModels.IndexColumn) bool {
				return source.IndexName == target.IndexName &&
					source.IndexType == target.IndexType &&
					source.IsUnique == target.IsUnique &&
					source.ColumnName == target.ColumnName &&
					source.KeyOrdinal == target.KeyOrdinal &&
					source.IsDescending == target.IsDescending &&
					source.IsIncluded == target.IsIncluded &&
					source.IncludeOrder == target.IncludeOrder &&
					source.HasFilter == target.HasFilter &&
					source.FilterDefinition == target.FilterDefinition
			},
		)
	})
}

func compareForeignKeys(dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareObjects(
		"Foreign key",
		sqlserverQueries.SqlServerForeignKeys(dbSource),
		sqlserverQueries.SqlServerForeignKeys(dbTarget),
		func(key sqlserverModels.ForeignKeyColumn) string {
			return key.ParentSchema + "." + key.ParentTable + "." + key.ForeignKeyName + "." + fmt.Sprint(key.KeyOrdinal)
		},
		func(source, target sqlserverModels.ForeignKeyColumn) bool {
			return source.ForeignKeyName == target.ForeignKeyName &&
				source.ParentSchema == target.ParentSchema &&
				source.ParentTable == target.ParentTable &&
				source.ColumnName == target.ColumnName &&
				source.KeyOrdinal == target.KeyOrdinal &&
				source.ReferencedSchema == target.ReferencedSchema &&
				source.ReferencedTable == target.ReferencedTable &&
				source.ReferencedColumn == target.ReferencedColumn &&
				source.DeleteAction == target.DeleteAction &&
				source.UpdateAction == target.UpdateAction
		},
	)
}

func compareExistingTables(dbSource *gorm.DB, dbTarget *gorm.DB, compare func(string)) {
	sourceTables := sqlserverQueries.SqlServerTables(*dbSource)
	targetTables := sqlserverQueries.SqlServerTables(*dbTarget)

	targetTableMap := make(map[string]struct{}, len(targetTables))
	for _, table := range targetTables {
		targetTableMap[table.Schema+"."+table.Name] = struct{}{}
	}

	for _, table := range sourceTables {
		tableKey := table.Schema + "." + table.Name
		if _, found := targetTableMap[tableKey]; found {
			compare(tableKey)
		}
	}
}

func compareExistingTableTypes(dbSource *gorm.DB, dbTarget *gorm.DB, compare func(string, string)) {
	sourceTableTypes := sqlserverQueries.SqlServerUserDefinedTableTypes(dbSource)
	targetTableTypes := sqlserverQueries.SqlServerUserDefinedTableTypes(dbTarget)

	targetTableTypeMap := make(map[string]struct{}, len(targetTableTypes))
	for _, tableType := range targetTableTypes {
		targetTableTypeMap[tableType.SchemaName+"."+tableType.TypeName] = struct{}{}
	}

	for _, tableType := range sourceTableTypes {
		key := tableType.SchemaName + "." + tableType.TypeName
		if _, found := targetTableTypeMap[key]; found {
			compare(tableType.SchemaName, tableType.TypeName)
		}
	}
}

func constraintKey(tableKey, constraintName string, isSystemNamed bool, columnName string, ordinal int) string {
	if isSystemNamed {
		return tableKey + ".<system>." + columnName + "." + fmt.Sprint(ordinal)
	}
	return tableKey + "." + constraintName + "." + fmt.Sprint(ordinal)
}

func sameConstraintName(sourceName string, sourceSystemNamed bool, targetName string, targetSystemNamed bool) bool {
	return sourceName == targetName || (sourceSystemNamed && targetSystemNamed)
}

func compareObjects[T any](
	objectType string,
	sourceObjects []T,
	targetObjects []T,
	key func(T) string,
	equal func(T, T) bool,
) {
	targetMap := make(map[string]T, len(targetObjects))
	for _, object := range targetObjects {
		targetMap[key(object)] = object
	}

	for _, sourceObject := range sourceObjects {
		objectKey := key(sourceObject)
		targetObject, found := targetMap[objectKey]

		if !found {
			if activeComparison != nil {
				activeComparison.record(objectType, objectKey)
			}
		} else if !equal(sourceObject, targetObject) {
			if activeComparison != nil {
				activeComparison.record(objectType, objectKey)
			}
		}
	}

	sourceMap := make(map[string]struct{}, len(sourceObjects))
	for _, sourceObject := range sourceObjects {
		sourceMap[key(sourceObject)] = struct{}{}
	}
	for _, targetObject := range targetObjects {
		objectKey := key(targetObject)
		if _, found := sourceMap[objectKey]; !found && activeComparison != nil {
			activeComparison.record(objectType, objectKey)
		}
	}
}
