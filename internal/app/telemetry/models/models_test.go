package models

import (
	"fmt"
	"log"
	"testing"

	"github.com/bburch01/FOTAAS/api"
	"github.com/joho/godotenv"
	//"github.com/google/uuid"
	//"github.com/joho/godotenv"
	//uid "github.com/google/uuid"
	//tel "github.com/bburch01/FOTAAS/internal/app/telemetry"
	//ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
	//mdl "github.com/bburch01/FOTAAS/internal/app/simulation/models"
)

func init() {

	var err error

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err = godotenv.Load(); err != nil {
		log.Panicf("failed to load environment variables with error: %v", err)
	}

	if err = InitDB(); err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize database driver with error: %v", err))
	}

}

func TestTelemetryModels(t *testing.T) {

	var req api.GetSimulatedTelemetryDataRequest
	var data *api.TelemetryData
	var err error

	req.SimulationUuid = "dc0d88fa-4e7b-4e3b-b10a-55194944e505"
	req.Constructor = api.Constructor_MERCEDES
	req.CarNumber = 44
	req.DatumDescription = api.TelemetryDatumDescription_BRAKE_TEMP_FL

	if data, err = RetrieveSimulatedTelemetryData(req); err != nil {
		t.Error("failed to retrieve simulated telemetry data with error: ", err)
		t.FailNow()
	}

	logger.Debug(fmt.Sprintf("telemetry data gran prix: %v", data.GranPrix))
	logger.Debug(fmt.Sprintf("telemetry data track: %v", data.Track))
	logger.Debug(fmt.Sprintf("telemetry data constructor: %v", data.Constructor))
	logger.Debug(fmt.Sprintf("telemetry data car number: %v", data.CarNumber))
	logger.Debug(fmt.Sprintf("telemetry data datum count: %v", len(data.TelemetryDatumMap)))

	for _, v := range data.TelemetryDatumMap {
		logger.Debug(fmt.Sprintf("telemetry data datum: %v", v))
	}

}
