package models

type DatabaseType string

const (
	SqlServer DatabaseType = "sqlserver"
	//PostgreSql DatabaseType = "postgresql"
	//MySql      DatabaseType = "mysql"
	//SQLite     DatabaseType = "sqlite"
)

type DatabaseTypeInfo struct {
	Name  string
	Value DatabaseType
}

var DatabaseTypes = []DatabaseTypeInfo{
	{
		Name:  "SQL Server",
		Value: SqlServer,
	},
	/*
		{
			Name:  "PostgreSQL",
			Value: PostgreSql,
		},
		{
			Name:  "MySQL",
			Value: MySql,
		},
		{
			Name:  "SQLite",
			Value: SQLite,
		},
	*/
}
