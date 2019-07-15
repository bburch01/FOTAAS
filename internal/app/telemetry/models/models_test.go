package models

import (
	"fmt"
	"log"
	"testing"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

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

func TestRetrieveSimulatedTelemetryData(t *testing.T) {

	//var req api.GetSimulatedTelemetryDataRequest
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

	/*
		for _, v := range data.TelemetryDatumMap {
			logger.Debug(fmt.Sprintf("telemetry data datum: %v", v))
		}
	*/

}

func TestRetrieveTelemetryData(t *testing.T) {

	// The following are valid telemetry db queries:
	// select * from telemetry_datum where constructor='HAAS' and simulation_transmit_sequence_number=1 and timestamp between '2019-07-09' and '2019-07-11';
	// select * from telemetry_datum where constructor='HAAS' and simulation_transmit_sequence_number=1 and timestamp between '2019-7-10 00:00:00' and '2019-7-10 23:59:59';

	var req api.GetTelemetryDataRequest
	var data *api.TelemetryData
	var err error
	var startTime, endTime time.Time

	req.Simulated = false
	req.Constructor = api.Constructor_MERCEDES
	req.CarNumber = 44
	req.SearchBy.DateRange = true

	startTime, err = time.Parse(time.RFC3339, "2019-07-08T00:00:00Z")
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}

	endTime, err = time.Parse(time.RFC3339, "2019-07-12T00:00:00Z")
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

	if data, err = RetrieveTelemetryData(req); err != nil {
		t.Error("failed to retrieve telemetry data with error: ", err)
		t.FailNow()
	}

	logger.Debug(fmt.Sprintf("telemetry data datum count: %v", len(data.TelemetryDatumMap)))

	/*
		for _, v := range data.TelemetryDatumMap {
			logger.Debug(fmt.Sprintf("telemetry data datum: %v", v))
		}
	*/

}
