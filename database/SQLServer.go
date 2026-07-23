package database

import (
	"blueprint/models"
	"fmt"
	"net"
	"net/url"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func SQLServerConnect(conn models.Connection) (*gorm.DB, error) {
	// github.com/denisenkom/go-mssqldb

	host := conn.Server
	if conn.Port != "" {
		host = net.JoinHostPort(conn.Server, conn.Port)
	}
	dsn := fmt.Sprintf(
		"sqlserver://%s:%s@%s?database=%s&trustservercertificate=true",
		url.QueryEscape(conn.User),
		url.QueryEscape(conn.Password),
		host,
		url.QueryEscape(conn.Database),
	)

	return gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
}
