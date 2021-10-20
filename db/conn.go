package db

import (
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

// DefaultDNS creates default DSN string
func DefaultDSN(host, port, user, password, dbname string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FTokyo", user, password, host, port, dbname)
}

// Connect opens connection to DB
func Connect(dsn string) error {
	// Establish connection
	conn, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// Check connection
	err = conn.Ping()
	if err != nil {
		return err
	}

	// Bind connection
	db = conn
	return nil
}

// Disconnect closes connection
func Disconnect() {
	if db != nil {
		db.Close()
	}
}

// GetConnection returns DB connection
func GetConnection() (*sqlx.DB, error) {
	if db != nil {
		return db, nil
	}
	return nil, errors.New("Connection is not established")
}
