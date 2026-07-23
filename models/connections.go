package models

type ConnectionsFile struct {
	Connections []Connection `json:"connections"`
}

type Connection struct {
	Id       string       `json:"id"`
	Type     DatabaseType `json:"type"`
	Name     string       `json:"name"`
	Server   string       `json:"server"`
	Port     string       `json:"port"`
	Database string       `json:"database"`
	User     string       `json:"user"`
	Password string       `json:"password,omitempty"`
}
