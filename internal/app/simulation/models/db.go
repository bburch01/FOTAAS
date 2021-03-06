package models

import (
	"database/sql"
	"log"
	"os"
	"time"

	logging "github.com/bburch01/FOTAAS/internal/pkg/logging"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var db *sql.DB
var logger *zap.Logger

func init() {

	var lm logging.LogMode
	var err error

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err = godotenv.Load(); err != nil {
		log.Panicf("failed to load environment variables with error: %v", err)
	}

	if lm, err = logging.LogModeForString(os.Getenv("LOG_MODE")); err != nil {
		log.Panicf("failed to initialize logging subsystem with error: %v", err)
	}

	if logger, err = logging.NewLogger(lm, os.Getenv("LOG_DIR"), os.Getenv("LOG_FILE_NAME")); err != nil {
		log.Panicf("failed to initialize logging subsystem with error: %v", err)
	}

}

func InitDB() error {
	dbDriver := os.Getenv("DB_DRIVER")
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("SIMULATION_SERVICE_DB_NAME")
	var err error

	dbConURL := dbUser + ":" + dbPass + "@tcp(" + dbHost + ")" + "/" + dbName + "?parseTime=true"
	db, err = sql.Open(dbDriver, dbConURL)
	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(time.Duration(86400))
	db.SetMaxIdleConns(8)

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
