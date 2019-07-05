package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	pbts "github.com/golang/protobuf/ptypes/timestamp"

	"github.com/bburch01/FOTAAS/api"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
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

}

func TestSimulationModels(t *testing.T) {

	var sim Simulation
	var startTime *pbts.Timestamp
	var err error

	startTime, err = ipbts.TimestampProto(time.Now())
	if err != nil {
		t.Error("failed to create timestamp with error: ", err)
	}

	dbDriver := os.Getenv("DB_DRIVER")
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("SIMULATION_SERVICE_DB_NAME")

	dbConURL := dbUser + ":" + dbPass + "@tcp(" + dbHost + ")" + "/" + dbName
	db, err = sql.Open(dbDriver, dbConURL)
	if err != nil {
		t.Error("failed to connect to db with error: ", err)
	}
	if err = PingDB(); err != nil {
		t.Error("failed to ping db with error: ", err)
	}

	simID := uuid.New().String()

	simMemberMap := make(map[string]SimulationMember)

	simMemberID := uuid.New().String()
	simMember := SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: api.Constructor_HAAS, CarNumber: 10, ForceAlarm: false, NoAlarms: false}
	simMemberMap[simMemberID] = simMember

	simMemberID = uuid.New().String()
	simMember = SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: api.Constructor_MERCEDES, CarNumber: 44, ForceAlarm: false, NoAlarms: false}
	simMemberMap[simMemberID] = simMember

	simMemberID = uuid.New().String()
	simMember = SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: api.Constructor_WILLIAMS, CarNumber: 3, ForceAlarm: false, NoAlarms: false}
	simMemberMap[simMemberID] = simMember

	sim = Simulation{ID: simID, DurationInMinutes: 60, SampleRate: api.SampleRate_SR_1000_MS, GrandPrix: api.GrandPrix_ITALIAN, Track: api.Track_MONZA,
		State: "IN_PROGRESS", StartTimestamp: startTime, PercentComplete: 0, SimulationMembers: simMemberMap}

	err = sim.Create()
	if err != nil {
		t.Error("failed to persist simulation with error: ", err)
	}

	for _, m := range sim.SimulationMembers {
		m.AlarmOccurred = true
		m.AlarmDatumDescription = api.TelemetryDatumDescription_BRAKE_TEMP_FL.String()
		m.AlarmDatumUnit = api.TelemetryDatumUnit_DEGREE_CELCIUS.String()
		m.AlarmDatumValue = float64(458.055)
		err = m.UpdateAlarmInfo()
		if err != nil {
			t.Error("failed to update simulation_member alarm info with error: ", err)
		}
	}

	var simMembers []SimulationMember

	simMembers, err = sim.FindAllMembers()
	if err != nil {
		t.Error("failed to retrieve simulation members with error: ", err)
	}

	for _, m := range simMembers {
		fmt.Printf("\nsimulation member: %v", m)
	}

}
