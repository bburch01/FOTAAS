package models

import (
	"fmt"
	"log"
	"testing"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

	"github.com/bburch01/FOTAAS/api"
	"github.com/joho/godotenv"
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

func TestRetrieveSimulatedTelemetryData(t *testing.T) {

	var data *api.TelemetryData
	var err error
	var startTime, endTime time.Time

	req := new(api.GetTelemetryDataRequest)
	req.SearchBy = new(api.GetTelemetryDataRequest_SearchBy)

	req.Simulated = true
	req.SimulationUuid = "a75c9b70-48a7-4c3a-bc80-db545bcdaaf5"
	req.Constructor = api.Constructor_MERCEDES
	req.CarNumber = 44
	req.SearchBy.Constructor = true
	req.SearchBy.CarNumber = true
	req.SearchBy.DateRange = true
	req.SearchBy.HighAlarm = true
	req.SearchBy.LowAlarm = true

	startTime, err = time.Parse(time.RFC3339, "2019-07-14T00:00:00Z")
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}

	endTime, err = time.Parse(time.RFC3339, "2019-07-16T23:59:59Z")
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}

	req.DateRangeBegin, err = ipbts.TimestampProto(startTime)
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}
	req.DateRangeEnd, err = ipbts.TimestampProto(endTime)
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}

	if data, err = RetrieveTelemetryData(*req); err != nil {
		t.Error("failed to retrieve simulated telemetry data with error: ", err)
		t.FailNow()
	}

	logger.Debug(fmt.Sprintf("telemetry data datum count: %v", len(data.TelemetryDatumMap)))

}

func TestRetrieveTelemetryData(t *testing.T) {

	var data *api.TelemetryData
	var err error
	var startTime, endTime time.Time

	req := new(api.GetTelemetryDataRequest)
	req.SearchBy = new(api.GetTelemetryDataRequest_SearchBy)
	req.Constructor = api.Constructor_HAAS
	req.CarNumber = 8
	req.Simulated = true
	req.SimulationUuid = "dca35e1b-b10f-4098-b8e8-34e1d30f35cc"
	req.SearchBy.DateRange = true
	req.SearchBy.Constructor = true
	req.SearchBy.CarNumber = true

	startTime, err = time.Parse(time.RFC3339, "2019-07-08T00:00:00Z")
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}

	endTime, err = time.Parse(time.RFC3339, "2019-07-20T00:00:00Z")
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}

	req.DateRangeBegin, err = ipbts.TimestampProto(startTime)
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}
	req.DateRangeEnd, err = ipbts.TimestampProto(endTime)
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}

	if data, err = RetrieveTelemetryData(*req); err != nil {
		t.Error("failed to retrieve telemetry data with error: ", err)
		t.FailNow()
	}

	if data == nil {
		t.Error("telemetry data was nil, no rows returned...")
		t.FailNow()
	}

	logger.Debug(fmt.Sprintf("telemetry data datum count: %v", len(data.TelemetryDatumMap)))

}
