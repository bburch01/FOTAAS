package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	logging "github.com/bburch01/FOTAAS/fotaas-internal/pkg/logging"
)

var db *sql.DB
var logger *zap.Logger

/*
func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("godotenv error: %v", err)
	}
}
*/

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
	//dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("ANALYSIS_SERVICE_DB_NAME")
	var err error

	logger.Debug(fmt.Sprintf("DB_DRIVER: %v DB_HOST: %v DB_USER: %v DB_PASSWORD: %v ANALYSIS_SERVICE_DB_NAME: %v",
		dbDriver, dbHost, dbUser, dbPass, dbName))

	dbConURL := dbUser + ":" + dbPass + "@tcp(" + dbHost + ")" + "/" + dbName

	logger.Debug(fmt.Sprintf("Database Connection URL: %v", dbConURL))

	//MySQLDB, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp("+dbHost+")"+"/"+dbName)
	db, err = sql.Open(dbDriver, dbConURL)

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
