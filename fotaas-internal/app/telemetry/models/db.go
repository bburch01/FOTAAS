package models

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("godotenv error: %v", err)
	}
}

func InitDB() error {
	dbDriver := os.Getenv("DB_DRIVER")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("TELEMETRY_SERVICE_DB_NAME")
	var err error
	db, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp("+dbHost+":"+dbPort+")"+"/"+dbName)
	if err != nil {
		return err
	}
	if err = PingDB(); err != nil {
		return err
	}
	return nil
}

func PingDB() error {
	if err := db.Ping(); err != nil {
		return err
	}
	return nil
}
