package db

// conn.go provides helper functions for connection to DB
import (
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"  // initialize mysql driver
	"github.com/jmoiron/sqlx"
)

var _db *sqlx.DB

// DefaultDSN creates default DSN string
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
	_db = conn
	return nil
}

// Disconnect closes connection
func Disconnect() {
	if _db != nil {
		_db.Close()
	}
}

// GetConnection returns DB connection
func GetConnection() (*sqlx.DB, error) {
	if _db != nil {
		return _db, nil
	}
	return nil, errors.New("Connection is not established")
}
