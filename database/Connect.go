package database

import (
	"blueprint/models"
	"fmt"

	"gorm.io/gorm"
)

func Connect(conn models.Connection) (*gorm.DB, error) {
	switch conn.Type {

	case models.SqlServer:
		return SQLServerConnect(conn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", conn.Type)
	}
}
