package database

import (
	"blueprint/models"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func SQLServerConnect(conn models.Connection) (*gorm.DB, error) {
	return SQLServerConnectToDatabase(conn, conn.Database)
}

func SQLServerConnectToDatabase(conn models.Connection, databaseName string) (*gorm.DB, error) {
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
		url.QueryEscape(databaseName),
	)

	logger := gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	return gorm.Open(sqlserver.Open(dsn), &gorm.Config{Logger: logger})
}
