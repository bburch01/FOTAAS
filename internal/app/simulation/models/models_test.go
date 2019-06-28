package models

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	pb "github.com/bburch01/FOTAAS/api"
	ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	uid "github.com/google/uuid"

	"github.com/joho/godotenv"
	//uid "github.com/google/uuid"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
	//tel "github.com/bburch01/FOTAAS/internal/app/telemetry"
	//ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
	//mdl "github.com/bburch01/FOTAAS/internal/app/simulation/models"
	//pb "github.com/bburch01/FOTAAS/api"
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
	var startTime *timestamp.Timestamp
	var err error

	startTime, err = ts.TimestampProto(time.Now())
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

	simID := uid.New().String()

	sim = Simulation{ID: simID, DurationInMinutes: 60, SampleRate: "SR_1000_MS", GrandPrix: "ITALIAN", Track: "MONZA",
		State: "IN_PROGRESS", StartTimestamp: startTime, PercentComplete: 0}

	err = sim.Create()
	if err != nil {
		t.Error("failed to persist simulation with error: ", err)
	}

	simMemberMap := make(map[string]SimulationMember)

	simMemberID := uid.New().String()
	simMember := SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: "HAAS", CarNumber: 10, ForceAlarm: false, NoAlarms: false}
	simMemberMap[simMemberID] = simMember

	simMemberID = uid.New().String()
	simMember = SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: "MERCEDES", CarNumber: 44, ForceAlarm: false, NoAlarms: false}
	simMemberMap[simMemberID] = simMember

	simMemberID = uid.New().String()
	simMember = SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: "WILLIAMS", CarNumber: 3, ForceAlarm: false, NoAlarms: false}
	simMemberMap[simMemberID] = simMember

	for _, m := range simMemberMap {
		err = m.Create()
		if err != nil {
			t.Error("failed to persist simulation_member with error: ", err)
		}
	}

	for _, m := range simMemberMap {
		m.AlarmOccurred = true
		m.AlarmDatumDescription = pb.TelemetryDatumDescription_BRAKE_TEMP_FL.String()
		m.AlarmDatumUnit = pb.TelemetryDatumUnit_DEGREE_CELCIUS.String()
		m.AlarmDatumValue = float64(458.055)
		err = m.UpdateAlarmInfo()
		if err != nil {
			t.Error("failed to update simulation_member alarm info with error: ", err)
		}
	}

	/*
		err = simMember.Create()
		if err != nil {
			t.Error("failed to persist simulation_member with error: ", err)
		}
		simMember = SimulationMember{ID: "sm1001", SimulationID: "yyz1200", Constructor: "MERCEDES", CarNumber: 44, ForceAlarm: false, NoAlarms: false}
		err = simMember.Create()
		if err != nil {
			t.Error("failed to persist simulation_member with error: ", err)
		}
		simMember = SimulationMember{ID: "sm1002", SimulationID: "yyz1200", Constructor: "WILLIAMS", CarNumber: 3, ForceAlarm: false, NoAlarms: false}
		err = simMember.Create()
		if err != nil {
			t.Error("failed to persist simulation_member with error: ", err)
		}
	*/

}
