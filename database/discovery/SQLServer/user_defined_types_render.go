package sqlserver

import (
	sqlserver_models "blueprint/database/discovery/SQLServer/models"
	"fmt"
	"strings"
)

func GenerateCreateUserDefinedType(userDefinedType sqlserver_models.UserDefinedType) string {
	definition := fmt.Sprintf(
		"CREATE TYPE %s.%s FROM %s",
		quoteSqlServerIdentifier(userDefinedType.SchemaName),
		quoteSqlServerIdentifier(userDefinedType.TypeName),
		userDefinedTypeBaseDefinition(userDefinedType),
	)

	if userDefinedType.IsNullable {
		definition += " NULL"
	} else {
		definition += " NOT NULL"
	}

	return definition + ";"
}

func userDefinedTypeBaseDefinition(userDefinedType sqlserver_models.UserDefinedType) string {
	baseType := userDefinedType.BaseTypeName
	switch strings.ToLower(baseType) {
	case "varchar", "char", "varbinary", "binary":
		if userDefinedType.MaxLength == -1 {
			return strings.ToUpper(baseType) + "(MAX)"
		}
		return fmt.Sprintf("%s(%d)", strings.ToUpper(baseType), userDefinedType.MaxLength)
	case "nvarchar", "nchar":
		if userDefinedType.MaxLength == -1 {
			return strings.ToUpper(baseType) + "(MAX)"
		}
		return fmt.Sprintf("%s(%d)", strings.ToUpper(baseType), userDefinedType.MaxLength/2)
	case "decimal", "numeric":
		return fmt.Sprintf("%s(%d,%d)", strings.ToUpper(baseType), userDefinedType.Precision, userDefinedType.Scale)
	case "datetime2", "datetimeoffset", "time":
		return fmt.Sprintf("%s(%d)", strings.ToUpper(baseType), userDefinedType.Scale)
	case "timestamp":
		return "ROWVERSION"
	default:
		return strings.ToUpper(baseType)
	}
}
