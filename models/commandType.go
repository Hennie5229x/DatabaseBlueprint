package models

type CommandType string

const (
	System      CommandType = "system"
	Connections CommandType = "connections"
	Database    CommandType = "database"
)
