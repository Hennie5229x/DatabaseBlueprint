package sqlserver

import (
	sqlserver_models "blueprint/database/discovery/SQLServer/models"
	"fmt"
	"sort"
	"strings"
)

func GenerateCreateTable(
	schemaName string,
	tableName string,
	columns []sqlserver_models.Column,
	defaultConstraints []sqlserver_models.DefaultConstraint,
	primaryKeys []sqlserver_models.PrimaryKeyColumn,
	uniqueConstraints []sqlserver_models.UniqueConstraintColumn,
	checkConstraints []sqlserver_models.CheckConstraint,
) string {
	if schemaName == "" {
		schemaName = "dbo"
	}

	defaultConstraintByColumnID := make(map[int]sqlserver_models.DefaultConstraint, len(defaultConstraints))
	for _, constraint := range defaultConstraints {
		defaultConstraintByColumnID[constraint.ColumnID] = constraint
	}

	sections := make([]string, 0, 4)

	columnDefinitions := sqlServerColumnDefinitions(columns, defaultConstraintByColumnID)
	if len(columnDefinitions) > 0 {
		sections = append(sections, strings.Join(columnDefinitions, ",\n"))
	}

	keyDefinitions := sqlServerKeyDefinitions(primaryKeys, uniqueConstraints)
	if len(keyDefinitions) > 0 {
		sections = append(sections, strings.Join(keyDefinitions, ",\n\n"))
	}

	ruleDefinitions := sqlServerRuleDefinitions(checkConstraints)
	if len(ruleDefinitions) > 0 {
		sections = append(sections, strings.Join(ruleDefinitions, ",\n\n"))
	}

	return fmt.Sprintf(
		"CREATE TABLE %s.%s\n(\n%s\n);",
		quoteSqlServerIdentifier(schemaName),
		quoteSqlServerIdentifier(tableName),
		strings.Join(sections, ",\n\n"),
	)
}

func GenerateCreateIndexes(schemaName string, tableName string, indexes []sqlserver_models.IndexColumn) string {
	if schemaName == "" {
		schemaName = "dbo"
	}

	indexDefinitions := sqlServerIndexDefinitions(schemaName, tableName, indexes)
	return strings.Join(indexDefinitions, "\n\n")
}

func GenerateForeignKeys(foreignKeys []sqlserver_models.ForeignKeyColumn) []string {

	groups := groupForeignKeys(foreignKeys)
	definitions := make([]string, 0, len(groups))

	for _, group := range groups {
		localColumns := make([]string, 0, len(group.Columns))
		referencedColumns := make([]string, 0, len(group.Columns))

		for _, column := range group.Columns {
			localColumns = append(localColumns, "    "+quoteSqlServerIdentifier(column.ColumnName))
			referencedColumns = append(referencedColumns, "    "+quoteSqlServerIdentifier(column.ReferencedColumn))
		}

		definition := fmt.Sprintf(
			"ALTER TABLE %s.%s\nADD CONSTRAINT %s FOREIGN KEY\n(\n%s\n)\nREFERENCES %s.%s\n(\n%s\n)\nON DELETE %s\nON UPDATE %s;",
			quoteSqlServerIdentifier(group.ParentSchema),
			quoteSqlServerIdentifier(group.ParentTable),
			quoteSqlServerIdentifier(group.ForeignKeyName),
			strings.Join(localColumns, ",\n"),
			quoteSqlServerIdentifier(group.ReferencedSchema),
			quoteSqlServerIdentifier(group.ReferencedTable),
			strings.Join(referencedColumns, ",\n"),
			sqlServerReferentialAction(group.DeleteAction),
			sqlServerReferentialAction(group.UpdateAction),
		)

		definitions = append(definitions, definition)
	}

	return definitions
}

func sqlServerColumnDefinitions(columns []sqlserver_models.Column, defaultConstraintByColumnID map[int]sqlserver_models.DefaultConstraint) []string {
	definitions := make([]string, 0, len(columns))
	columnNameWidth := 0
	for _, column := range columns {
		if column.ComputedDefinition != "" {
			continue
		}

		width := len(quoteSqlServerIdentifier(column.ColumnName))
		if width > columnNameWidth {
			columnNameWidth = width
		}
	}

	for _, column := range columns {
		if column.ComputedDefinition != "" {
			definition := fmt.Sprintf(
				"    %s AS %s",
				quoteSqlServerIdentifier(column.ColumnName),
				column.ComputedDefinition,
			)

			if column.IsPersisted {
				definition += " PERSISTED"
			}

			definitions = append(definitions, definition)
			continue
		}

		columnIdentifier := quoteSqlServerIdentifier(column.ColumnName)
		definition := fmt.Sprintf(
			"    %-*s %s",
			columnNameWidth,
			columnIdentifier,
			sqlServerColumnDataType(column),
		)

		if column.IsIdentity {
			definition += fmt.Sprintf(
				" IDENTITY(%s,%s)",
				sqlServerIdentityValue(column.IdentitySeed, "1"),
				sqlServerIdentityValue(column.IdentityIncrement, "1"),
			)
		}

		if column.IsNullable {
			definition += " NULL"
		} else {
			definition += " NOT NULL"
		}

		if constraint, ok := defaultConstraintByColumnID[column.ColumnID]; ok {
			if constraint.ConstraintName != "" && !constraint.IsSystemNamed {
				definition += " CONSTRAINT " + quoteSqlServerIdentifier(constraint.ConstraintName)
			}
			definition += " DEFAULT " + constraint.ConstraintValue
		}

		definitions = append(definitions, definition)
	}

	return definitions
}

func sqlServerKeyDefinitions(primaryKeys []sqlserver_models.PrimaryKeyColumn, uniqueConstraints []sqlserver_models.UniqueConstraintColumn) []string {
	definitions := make([]string, 0)

	for _, primaryKey := range groupPrimaryKeys(primaryKeys) {
		constraintName := sqlServerConstraintName(primaryKey.ConstraintName, primaryKey.IsSystemNamed)
		definitions = append(definitions, sqlServerKeyConstraintDefinition("PRIMARY KEY", constraintName, primaryKey.IndexType, primaryKey.Columns))
	}

	for _, uniqueConstraint := range groupUniqueConstraints(uniqueConstraints) {
		constraintName := sqlServerConstraintName(uniqueConstraint.ConstraintName, uniqueConstraint.IsSystemNamed)
		definitions = append(definitions, sqlServerKeyConstraintDefinition("UNIQUE", constraintName, uniqueConstraint.IndexType, uniqueConstraint.Columns))
	}

	return definitions
}

func sqlServerConstraintName(name string, isSystemNamed bool) string {
	if isSystemNamed {
		return ""
	}

	return name
}

func sqlServerRuleDefinitions(checkConstraints []sqlserver_models.CheckConstraint) []string {
	sortedConstraints := append([]sqlserver_models.CheckConstraint(nil), checkConstraints...)
	sort.Slice(sortedConstraints, func(i, j int) bool {
		return sortedConstraints[i].Definition < sortedConstraints[j].Definition
	})

	definitions := make([]string, 0, len(sortedConstraints))
	for _, checkConstraint := range sortedConstraints {
		definitions = append(definitions, "    CHECK "+checkConstraint.Definition)
	}

	return definitions
}

func sqlServerIndexDefinitions(schemaName string, tableName string, indexes []sqlserver_models.IndexColumn) []string {
	groups := groupIndexes(indexes)
	definitions := make([]string, 0, len(groups))

	for _, group := range groups {
		keyColumns := make([]string, 0, len(group.KeyColumns))
		for _, column := range group.KeyColumns {
			keyColumns = append(keyColumns, "    "+sqlServerIndexedColumn(column.ColumnName, column.IsDescending))
		}

		indexName := group.IndexName
		if !group.IsUserDefinedName {
			indexName = sqlServerDeterministicIndexName(tableName, group.KeyColumns, group.IncludeColumns, group.IsUnique)
		}

		definition := fmt.Sprintf(
			"CREATE %s%s INDEX %s\nON %s.%s\n(\n%s\n)",
			sqlServerUniqueKeyword(group.IsUnique),
			group.IndexType,
			quoteSqlServerIdentifier(indexName),
			quoteSqlServerIdentifier(schemaName),
			quoteSqlServerIdentifier(tableName),
			strings.Join(keyColumns, ",\n"),
		)

		if len(group.IncludeColumns) > 0 {
			includeColumns := make([]string, 0, len(group.IncludeColumns))
			for _, column := range group.IncludeColumns {
				includeColumns = append(includeColumns, "    "+quoteSqlServerIdentifier(column.ColumnName))
			}

			definition += "\nINCLUDE\n(\n" + strings.Join(includeColumns, ",\n") + "\n)"
		}

		if group.HasFilter && group.FilterDefinition != "" {
			definition += "\nWHERE " + group.FilterDefinition
		}

		definitions = append(definitions, definition+";")
	}

	return definitions
}

func sqlServerKeyConstraintDefinition(keyword string, constraintName string, indexType string, columns []orderedColumn) string {
	keyColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		keyColumns = append(keyColumns, "        "+sqlServerIndexedColumn(column.ColumnName, column.IsDescending))
	}

	constraint := keyword
	if constraintName != "" {
		constraint = "CONSTRAINT " + quoteSqlServerIdentifier(constraintName) + " " + constraint
	}

	return fmt.Sprintf(
		"    %s %s\n    (\n%s\n    )",
		constraint,
		indexType,
		strings.Join(keyColumns, ",\n"),
	)
}

func sqlServerIndexedColumn(columnName string, isDescending bool) string {
	if isDescending {
		return quoteSqlServerIdentifier(columnName) + " DESC"
	}

	return quoteSqlServerIdentifier(columnName)
}

func sqlServerUniqueKeyword(isUnique bool) string {
	if isUnique {
		return "UNIQUE "
	}

	return ""
}

func sqlServerColumnDataType(column sqlserver_models.Column) string {
	dataType := strings.ToLower(column.DataType)
	formattedType := dataType

	switch dataType {
	case "varchar", "char", "varbinary", "binary":
		if column.MaxLength == -1 {
			formattedType = fmt.Sprintf("%s(MAX)", dataType)
			break
		}

		formattedType = fmt.Sprintf("%s(%d)", dataType, column.MaxLength)

	case "nvarchar", "nchar":
		if column.MaxLength == -1 {
			formattedType = fmt.Sprintf("%s(MAX)", dataType)
			break
		}

		formattedType = fmt.Sprintf("%s(%d)", dataType, column.MaxLength/2)

	case "decimal", "numeric":
		formattedType = fmt.Sprintf("%s(%d,%d)", dataType, column.Precision, column.Scale)

	case "datetime2", "datetimeoffset", "time":
		formattedType = fmt.Sprintf("%s(%d)", dataType, column.Scale)

	case "timestamp":
		formattedType = "rowversion"
	}

	return strings.ToUpper(formattedType)
}

func sqlServerIdentityValue(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func sqlServerReferentialAction(action string) string {
	return strings.ReplaceAll(action, "_", " ")
}

func sqlServerDeterministicIndexName(tableName string, keyColumns []orderedColumn, includeColumns []orderedColumn, isUnique bool) string {
	prefix := "IX"
	if isUnique {
		prefix = "UX"
	}

	parts := []string{prefix, tableName}
	for _, column := range keyColumns {
		parts = append(parts, sanitizeSqlServerNamePart(column.ColumnName))
	}

	if len(includeColumns) > 0 {
		parts = append(parts, "Include")
		for _, column := range includeColumns {
			parts = append(parts, sanitizeSqlServerNamePart(column.ColumnName))
		}
	}

	return strings.Join(parts, "_")
}

func sanitizeSqlServerNamePart(value string) string {
	replacer := strings.NewReplacer(" ", "_", "]", "", "[", "", ".", "_", ",", "_")
	return replacer.Replace(value)
}

func quoteSqlServerIdentifier(identifier string) string {
	escapedIdentifier := strings.ReplaceAll(identifier, "]", "]]")

	return "[" + escapedIdentifier + "]"
}

type orderedColumn struct {
	ColumnName   string
	Ordinal      int
	IsDescending bool
}

type keyConstraintGroup struct {
	ConstraintName string
	IsSystemNamed  bool
	IndexType      string
	Columns        []orderedColumn
}

type foreignKeyGroup struct {
	ForeignKeyName   string
	ParentSchema     string
	ParentTable      string
	Columns          []sqlserver_models.ForeignKeyColumn
	ReferencedSchema string
	ReferencedTable  string
	DeleteAction     string
	UpdateAction     string
}

type indexGroup struct {
	IndexName         string
	IndexType         string
	IsUnique          bool
	HasFilter         bool
	FilterDefinition  string
	IsUserDefinedName bool
	KeyColumns        []orderedColumn
	IncludeColumns    []orderedColumn
}

func groupPrimaryKeys(primaryKeys []sqlserver_models.PrimaryKeyColumn) []keyConstraintGroup {
	if len(primaryKeys) == 0 {
		return nil
	}

	groupsByID := make(map[int]*keyConstraintGroup)
	constraintIDs := make([]int, 0)
	for _, primaryKey := range primaryKeys {
		group := groupsByID[primaryKey.ConstraintObjectID]
		if group == nil {
			group = &keyConstraintGroup{
				ConstraintName: primaryKey.ConstraintName,
				IsSystemNamed:  primaryKey.IsSystemNamed,
				IndexType:      primaryKey.IndexType,
			}
			groupsByID[primaryKey.ConstraintObjectID] = group
			constraintIDs = append(constraintIDs, primaryKey.ConstraintObjectID)
		}

		group.Columns = append(group.Columns, orderedColumn{ColumnName: primaryKey.ColumnName, Ordinal: primaryKey.KeyOrdinal, IsDescending: primaryKey.IsDescending})
	}

	sort.Ints(constraintIDs)
	groups := make([]keyConstraintGroup, 0, len(constraintIDs))
	for _, constraintID := range constraintIDs {
		group := groupsByID[constraintID]
		sort.Slice(group.Columns, func(i, j int) bool { return group.Columns[i].Ordinal < group.Columns[j].Ordinal })
		groups = append(groups, *group)
	}

	return groups
}

func groupUniqueConstraints(uniqueConstraints []sqlserver_models.UniqueConstraintColumn) []keyConstraintGroup {
	if len(uniqueConstraints) == 0 {
		return nil
	}

	groupsByID := make(map[int]*keyConstraintGroup)
	constraintIDs := make([]int, 0)
	for _, uniqueConstraint := range uniqueConstraints {
		group := groupsByID[uniqueConstraint.ConstraintObjectID]
		if group == nil {
			group = &keyConstraintGroup{
				ConstraintName: uniqueConstraint.ConstraintName,
				IsSystemNamed:  uniqueConstraint.IsSystemNamed,
				IndexType:      uniqueConstraint.IndexType,
			}
			groupsByID[uniqueConstraint.ConstraintObjectID] = group
			constraintIDs = append(constraintIDs, uniqueConstraint.ConstraintObjectID)
		}

		group.Columns = append(group.Columns, orderedColumn{ColumnName: uniqueConstraint.ColumnName, Ordinal: uniqueConstraint.KeyOrdinal, IsDescending: uniqueConstraint.IsDescending})
	}

	sort.Ints(constraintIDs)
	groups := make([]keyConstraintGroup, 0, len(constraintIDs))
	for _, constraintID := range constraintIDs {
		group := groupsByID[constraintID]
		sort.Slice(group.Columns, func(i, j int) bool { return group.Columns[i].Ordinal < group.Columns[j].Ordinal })
		groups = append(groups, *group)
	}

	sort.Slice(groups, func(i, j int) bool {
		return sqlServerConstraintSortKey(groups[i]) < sqlServerConstraintSortKey(groups[j])
	})

	return groups
}

func sqlServerConstraintSortKey(group keyConstraintGroup) string {
	parts := []string{group.IndexType}
	for _, column := range group.Columns {
		parts = append(parts, column.ColumnName)
		if column.IsDescending {
			parts = append(parts, "DESC")
		}
	}

	return strings.Join(parts, "|")
}

func groupForeignKeys(foreignKeys []sqlserver_models.ForeignKeyColumn) []foreignKeyGroup {
	if len(foreignKeys) == 0 {
		return nil
	}

	groupsByID := make(map[int]*foreignKeyGroup)
	constraintIDs := make([]int, 0)
	for _, foreignKey := range foreignKeys {
		group := groupsByID[foreignKey.ForeignKeyObjectID]
		if group == nil {
			group = &foreignKeyGroup{
				ForeignKeyName:   foreignKey.ForeignKeyName,
				ParentSchema:     foreignKey.ParentSchema,
				ParentTable:      foreignKey.ParentTable,
				ReferencedSchema: foreignKey.ReferencedSchema,
				ReferencedTable:  foreignKey.ReferencedTable,
				DeleteAction:     foreignKey.DeleteAction,
				UpdateAction:     foreignKey.UpdateAction,
			}
			groupsByID[foreignKey.ForeignKeyObjectID] = group
			constraintIDs = append(constraintIDs, foreignKey.ForeignKeyObjectID)
		}

		group.Columns = append(group.Columns, foreignKey)
	}

	groups := make([]foreignKeyGroup, 0, len(constraintIDs))
	for _, constraintID := range constraintIDs {
		group := groupsByID[constraintID]
		sort.Slice(group.Columns, func(i, j int) bool { return group.Columns[i].KeyOrdinal < group.Columns[j].KeyOrdinal })
		groups = append(groups, *group)
	}

	sort.Slice(groups, func(i, j int) bool {
		return sqlServerForeignKeySortKey(groups[i]) < sqlServerForeignKeySortKey(groups[j])
	})

	return groups
}

func sqlServerForeignKeySortKey(group foreignKeyGroup) string {
	parts := []string{group.ParentSchema, group.ParentTable, group.ForeignKeyName, group.ReferencedSchema, group.ReferencedTable}
	for _, column := range group.Columns {
		parts = append(parts, column.ColumnName, column.ReferencedColumn)
	}
	parts = append(parts, group.DeleteAction, group.UpdateAction)

	return strings.Join(parts, "|")
}

func groupIndexes(indexes []sqlserver_models.IndexColumn) []indexGroup {
	if len(indexes) == 0 {
		return nil
	}

	groupsByID := make(map[int]*indexGroup)
	indexIDs := make([]int, 0)
	for _, index := range indexes {
		group := groupsByID[index.IndexID]
		if group == nil {
			group = &indexGroup{
				IndexName:         index.IndexName,
				IndexType:         index.IndexType,
				IsUnique:          index.IsUnique,
				HasFilter:         index.HasFilter,
				FilterDefinition:  index.FilterDefinition,
				IsUserDefinedName: index.IsUserDefinedName,
			}
			groupsByID[index.IndexID] = group
			indexIDs = append(indexIDs, index.IndexID)
		}

		ordered := orderedColumn{ColumnName: index.ColumnName, Ordinal: index.KeyOrdinal, IsDescending: index.IsDescending}
		if index.IsIncluded {
			ordered.Ordinal = index.IncludeOrder
			group.IncludeColumns = append(group.IncludeColumns, ordered)
			continue
		}

		group.KeyColumns = append(group.KeyColumns, ordered)
	}

	groups := make([]indexGroup, 0, len(indexIDs))
	for _, indexID := range indexIDs {
		group := groupsByID[indexID]
		sort.Slice(group.KeyColumns, func(i, j int) bool { return group.KeyColumns[i].Ordinal < group.KeyColumns[j].Ordinal })
		sort.Slice(group.IncludeColumns, func(i, j int) bool { return group.IncludeColumns[i].Ordinal < group.IncludeColumns[j].Ordinal })
		groups = append(groups, *group)
	}

	sort.Slice(groups, func(i, j int) bool {
		return sqlServerIndexSortKey(groups[i]) < sqlServerIndexSortKey(groups[j])
	})

	return groups
}

func sqlServerIndexSortKey(group indexGroup) string {
	parts := []string{sqlServerUniqueKeyword(group.IsUnique), group.IndexType}
	for _, column := range group.KeyColumns {
		parts = append(parts, column.ColumnName)
		if column.IsDescending {
			parts = append(parts, "DESC")
		}
	}
	parts = append(parts, group.FilterDefinition)
	for _, column := range group.IncludeColumns {
		parts = append(parts, "INCLUDE", column.ColumnName)
	}

	return strings.Join(parts, "|")
}

func sqlServerSynonymsDefinition(synonyms sqlserver_models.Synonyms) string {
	return fmt.Sprintf("CREATE SYNONYM %s.%s\nFOR %s\n", synonyms.SchemaName, synonyms.SynonymName, synonyms.BaseObjectName)
}

func sqlServerSchemaDefinition(schema sqlserver_models.Schemas) string {
	return fmt.Sprintf("CREATE SCHEMA [%s]", schema.Name)
}
