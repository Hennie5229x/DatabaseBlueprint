package sqlserver

import (
	discoverySQLServer "blueprint/database/discovery/SQLServer"
	sqlserverModels "blueprint/database/discovery/SQLServer/models"
	sqlserverQueries "blueprint/database/discovery/SQLServer/queries"
	discoveryModels "blueprint/database/discovery/models"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func addConstraintMigrations(statements *[]migrationStatement, dbSource *gorm.DB, dbTarget *gorm.DB) {
	compareExistingTablesForMigration(dbSource, dbTarget, func(table discoveryModels.Tables, tableKey string) {
		addPrimaryKeyMigrations(statements, dbSource, dbTarget, table, tableKey)
		addUniqueConstraintMigrations(statements, dbSource, dbTarget, table, tableKey)
		addCheckConstraintMigrations(statements, dbSource, dbTarget, table, tableKey)
		addIndexMigrations(statements, dbSource, dbTarget, table, tableKey)
	})
}

func addPrimaryKeyMigrations(statements *[]migrationStatement, dbSource *gorm.DB, dbTarget *gorm.DB, table discoveryModels.Tables, tableKey string) {
	source := groupPrimaryKeys(sqlserverQueries.SqlServerPrimaryKeys(dbSource, tableKey))
	target := groupPrimaryKeys(sqlserverQueries.SqlServerPrimaryKeys(dbTarget, tableKey))

	for key, sourceRows := range source {
		targetRows, exists := target[key]
		if exists && primaryKeyRowsEqual(sourceRows, targetRows) {
			continue
		}

		name := sourceRows[0].ConstraintName
		sql := createPrimaryKeySQL(table, sourceRows)
		if exists {
			if sourceRows[0].IsSystemNamed || targetRows[0].IsSystemNamed {
				sql = "-- TODO: Replace changed system-named primary key safely."
			} else {
				sql = dropConstraintSQL(table, targetRows[0].ConstraintName) + "\n" + sql
			}
		}

		*statements = append(*statements, migrationStatement{order: 65, key: "Primary key " + tableKey + "." + name, sql: sql})
	}
}

func addUniqueConstraintMigrations(statements *[]migrationStatement, dbSource *gorm.DB, dbTarget *gorm.DB, table discoveryModels.Tables, tableKey string) {
	source := groupUniqueConstraints(sqlserverQueries.SqlServerUniqueConstraints(dbSource, tableKey))
	target := groupUniqueConstraints(sqlserverQueries.SqlServerUniqueConstraints(dbTarget, tableKey))

	for key, sourceRows := range source {
		targetRows, exists := target[key]
		if exists && uniqueConstraintRowsEqual(sourceRows, targetRows) {
			continue
		}

		name := sourceRows[0].ConstraintName
		sql := createUniqueConstraintSQL(table, sourceRows)
		if exists {
			if sourceRows[0].IsSystemNamed || targetRows[0].IsSystemNamed {
				sql = "-- TODO: Replace changed system-named unique constraint safely."
			} else {
				sql = dropConstraintSQL(table, targetRows[0].ConstraintName) + "\n" + sql
			}
		}

		*statements = append(*statements, migrationStatement{order: 66, key: "Unique constraint " + tableKey + "." + name, sql: sql})
	}
}

func addCheckConstraintMigrations(statements *[]migrationStatement, dbSource *gorm.DB, dbTarget *gorm.DB, table discoveryModels.Tables, tableKey string) {
	source := groupCheckConstraints(sqlserverQueries.SqlServerCheckConstraints(dbSource, tableKey))
	target := groupCheckConstraints(sqlserverQueries.SqlServerCheckConstraints(dbTarget, tableKey))

	for key, sourceConstraint := range source {
		targetConstraint, exists := target[key]
		if exists && checkConstraintsEqual(sourceConstraint, targetConstraint) {
			continue
		}

		sql := createCheckConstraintSQL(table, sourceConstraint)
		if exists {
			if sourceConstraint.IsSystemNamed || targetConstraint.IsSystemNamed {
				sql = "-- TODO: Replace changed system-named check constraint safely."
			} else {
				sql = dropConstraintSQL(table, targetConstraint.ConstraintName) + "\n" + sql
			}
		}

		*statements = append(*statements, migrationStatement{order: 67, key: "Check constraint " + tableKey + "." + key, sql: sql})
	}
}

func addIndexMigrations(statements *[]migrationStatement, dbSource *gorm.DB, dbTarget *gorm.DB, table discoveryModels.Tables, tableKey string) {
	source := groupIndexes(sqlserverQueries.SqlServerIndexes(dbSource, tableKey))
	target := groupIndexes(sqlserverQueries.SqlServerIndexes(dbTarget, tableKey))

	for key, sourceRows := range source {
		targetRows, exists := target[key]
		if exists && indexRowsEqual(sourceRows, targetRows) {
			continue
		}

		indexName := sourceRows[0].IndexName
		sql := discoverySQLServer.GenerateCreateIndexes(table.Schema, table.Name, sourceRows)
		if exists {
			if indexName == "" || targetRows[0].IndexName == "" {
				sql = "-- TODO: Replace changed unnamed index safely."
			} else {
				sql = fmt.Sprintf("DROP INDEX %s ON %s.%s;\n%s", quoteIdentifier(targetRows[0].IndexName), quoteIdentifier(table.Schema), quoteIdentifier(table.Name), sql)
			}
		}

		*statements = append(*statements, migrationStatement{order: 68, key: "Index " + tableKey + "." + indexName, sql: sql})
	}
}

func groupPrimaryKeys(rows []sqlserverModels.PrimaryKeyColumn) map[string][]sqlserverModels.PrimaryKeyColumn {
	return groupRows(rows, func(row sqlserverModels.PrimaryKeyColumn) string {
		if row.IsSystemNamed {
			return "<system-primary>"
		}
		return row.ConstraintName
	})
}

func groupUniqueConstraints(rows []sqlserverModels.UniqueConstraintColumn) map[string][]sqlserverModels.UniqueConstraintColumn {
	return groupRows(rows, func(row sqlserverModels.UniqueConstraintColumn) string {
		if row.IsSystemNamed {
			return "<system-unique>"
		}
		return row.ConstraintName
	})
}

func groupCheckConstraints(rows []sqlserverModels.CheckConstraint) map[string]sqlserverModels.CheckConstraint {
	result := make(map[string]sqlserverModels.CheckConstraint, len(rows))
	for _, row := range rows {
		key := row.ConstraintName
		if row.IsSystemNamed {
			key = "<system-check>"
		}
		result[key] = row
	}
	return result
}

func groupIndexes(rows []sqlserverModels.IndexColumn) map[string][]sqlserverModels.IndexColumn {
	return groupRows(rows, func(row sqlserverModels.IndexColumn) string {
		if row.IndexName == "" {
			return "<unnamed-index>"
		}
		return row.IndexName
	})
}

func groupRows[T any](rows []T, key func(T) string) map[string][]T {
	result := make(map[string][]T)
	for _, row := range rows {
		groupKey := key(row)
		result[groupKey] = append(result[groupKey], row)
	}
	return result
}

func primaryKeyRowsEqual(source []sqlserverModels.PrimaryKeyColumn, target []sqlserverModels.PrimaryKeyColumn) bool {
	if len(source) != len(target) {
		return false
	}
	for index := range source {
		if source[index].IndexType != target[index].IndexType ||
			source[index].ColumnName != target[index].ColumnName ||
			source[index].KeyOrdinal != target[index].KeyOrdinal ||
			source[index].IsDescending != target[index].IsDescending {
			return false
		}
	}
	return true
}

func uniqueConstraintRowsEqual(source []sqlserverModels.UniqueConstraintColumn, target []sqlserverModels.UniqueConstraintColumn) bool {
	if len(source) != len(target) {
		return false
	}
	for index := range source {
		if source[index].IndexType != target[index].IndexType ||
			source[index].ColumnName != target[index].ColumnName ||
			source[index].KeyOrdinal != target[index].KeyOrdinal ||
			source[index].IsDescending != target[index].IsDescending {
			return false
		}
	}
	return true
}

func indexRowsEqual(source []sqlserverModels.IndexColumn, target []sqlserverModels.IndexColumn) bool {
	if len(source) != len(target) {
		return false
	}
	for index := range source {
		left := source[index]
		right := target[index]
		if left.IndexType != right.IndexType ||
			left.IsUnique != right.IsUnique ||
			left.ColumnName != right.ColumnName ||
			left.KeyOrdinal != right.KeyOrdinal ||
			left.IsDescending != right.IsDescending ||
			left.IsIncluded != right.IsIncluded ||
			left.IncludeOrder != right.IncludeOrder ||
			left.HasFilter != right.HasFilter ||
			left.FilterDefinition != right.FilterDefinition {
			return false
		}
	}
	return true
}

func checkConstraintsEqual(source sqlserverModels.CheckConstraint, target sqlserverModels.CheckConstraint) bool {
	return source.ConstraintName == target.ConstraintName &&
		source.IsSystemNamed == target.IsSystemNamed &&
		source.Definition == target.Definition
}

func createPrimaryKeySQL(table discoveryModels.Tables, rows []sqlserverModels.PrimaryKeyColumn) string {
	columns := make([]string, 0, len(rows))
	for _, row := range rows {
		columns = append(columns, "    "+quoteIdentifier(row.ColumnName)+indexDirection(row.IsDescending))
	}
	name := ""
	if !rows[0].IsSystemNamed {
		name = " CONSTRAINT " + quoteIdentifier(rows[0].ConstraintName)
	}
	return fmt.Sprintf("ALTER TABLE %s.%s ADD%s PRIMARY KEY %s\n(\n%s\n);", quoteIdentifier(table.Schema), quoteIdentifier(table.Name), name, rows[0].IndexType, strings.Join(columns, ",\n"))
}

func createUniqueConstraintSQL(table discoveryModels.Tables, rows []sqlserverModels.UniqueConstraintColumn) string {
	columns := make([]string, 0, len(rows))
	for _, row := range rows {
		columns = append(columns, "    "+quoteIdentifier(row.ColumnName)+indexDirection(row.IsDescending))
	}
	name := ""
	if !rows[0].IsSystemNamed {
		name = " CONSTRAINT " + quoteIdentifier(rows[0].ConstraintName)
	}
	return fmt.Sprintf("ALTER TABLE %s.%s ADD%s UNIQUE %s\n(\n%s\n);", quoteIdentifier(table.Schema), quoteIdentifier(table.Name), name, rows[0].IndexType, strings.Join(columns, ",\n"))
}

func createCheckConstraintSQL(table discoveryModels.Tables, constraint sqlserverModels.CheckConstraint) string {
	name := ""
	if !constraint.IsSystemNamed {
		name = " CONSTRAINT " + quoteIdentifier(constraint.ConstraintName)
	}
	return fmt.Sprintf("ALTER TABLE %s.%s ADD%s CHECK %s;", quoteIdentifier(table.Schema), quoteIdentifier(table.Name), name, constraint.Definition)
}

func dropConstraintSQL(table discoveryModels.Tables, constraintName string) string {
	return fmt.Sprintf("ALTER TABLE %s.%s DROP CONSTRAINT %s;", quoteIdentifier(table.Schema), quoteIdentifier(table.Name), quoteIdentifier(constraintName))
}

func indexDirection(descending bool) string {
	if descending {
		return " DESC"
	}
	return " ASC"
}
