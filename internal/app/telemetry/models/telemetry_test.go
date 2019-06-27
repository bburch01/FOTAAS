package models

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/joho/godotenv"
	//mdl "github.com/bburch01/FOTAAS/internal/app/simulation/models"
	// 	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	//tel "github.com/bburch01/FOTAAS/internal/app/telemetry"
	//ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
)

/*
func TestStartSimulation(t *testing.T) {

	var simData map[pb.TelemetryDatumDescription]tel.SimulatedTelemetryData
	var err error
	var tstamp time.Time

	simDurationInMinutes := int32(1)
	sampleRate := pb.SampleRate_SR_1000_MS

	sim := pb.Simulation{Uuid: uid.New().String(), DurationInMinutes: simDurationInMinutes, SampleRate: sampleRate,
		SimulationRateMultiplier: pb.SimulationRateMultiplier_X20, GrandPrix: pb.GrandPrix_UNITED_STATES,
		Track: pb.Track_AUSTIN, Constructor: pb.Constructor_HAAS, CarNumber: 8, ForceAlarm: false, NoAlarms: true,
	}

	if simData, err = d.GenerateSimulatedTelemetryData(sim); err != nil {
		t.Error("failed to generate simulation data with error: ", err)
	}

	for _, v1 := range simData {
		//logger.Debug(fmt.Sprintf("simData datumDesc: %v alarmExists: %v", v1.DatumDesc, v1.AlarmExists))
		for _, v2 := range v1.Data {
			if tstamp, err = ts.Timestamp(v2.Timestamp); err != nil {
				t.Error("failed to convert google.protobuf.timestamp to time.Time with error: ", err)
			}
			logger.Debug(fmt.Sprintf("datum desc: %v unit: %v timestamp: %v value: %v", v2.Description.String(),
				v2.Unit.String(), tstamp, v2.Value))
		}
	}


}
*/

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

func TestTelemetryModels(t *testing.T) {

	var td TelemetryDatum
	var startTime *timestamp.Timestamp
	var err error

	startTime, err = ts.TimestampProto(time.Now())
	if err != nil {
		t.Error("failed to create timestamp with error: ", err)
	}

	/*
		sim = mdl.Simulation{ID: "yyz100", DurationInMinutes: 60, SampleRate: "SR_1000_MS", GrandPrix: "ITALIAN", Track: "MONZA",
			StartTimestamp: startTime, PercentComplete: 50}
	*/

	/*
		sim.ID = "yyz100"
		sim.DurationInMinutes = 60
		sim.SampleRate = "SR_1000_MS"
		sim.GrandPrix = "ITALIAN"
		sim.Track = "MONZA"
		sim.State = "IN_PROGRESS"
		sim.StartTimestamp = startTime
		sim.EndTimestamp = startTime
		sim.PercentComplete = 50
		sim.FinalStatusCode = "OK"
		sim.FinalStatusMessage = "RUNNING"
	*/

	/*
		td.ID = "yyz100"
		td.Simulated = false
		td.Description = "G_FORCE"
		td.Unit = "G"
		td.Timestamp = startTime
		td.Latitude = float64(0.0)
		td.Longitude = float64(0.0)
		td.Elevation = float64(0.0)
		td.Value = float64(0.0)
		td.HiAlarm = false
		td.LoAlarm = false
	*/
	/*
		err = td.Persist()
		if err != nil {
			t.Error("failed to persist telemetry datum with error: ", err)
		}
	*/

	/*
		err = PingDB()
		if err != nil {
			t.Error("failed to ping db with error: ", err)
		}
	*/

	dbDriver := os.Getenv("DB_DRIVER")
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("TELEMETRY_SERVICE_DB_NAME")

	dbConURL := dbUser + ":" + dbPass + "@tcp(" + dbHost + ")" + "/" + dbName
	db, err = sql.Open(dbDriver, dbConURL)
	if err != nil {
		t.Error("failed to connect to db with error: ", err)
	}
	if err = PingDB(); err != nil {
		t.Error("failed to ping db with error: ", err)
	}

	/*
		ID                               string
		Simulated                        bool
		SimulationID                     string
		SimulationTransmitSequenceNumber int32
		GrandPrix                        string
		Track                            string
		Constructor                      string
		CarNumber                        int32
		Timestamp                        *timestamp.Timestamp
		Latitude                         float64
		Longitude                        float64
		Elevation                        float64
		Description                      string
		Unit                             string
		Value                            float64
		HiAlarm                          bool
		LoAlarm                          bool
	*/

	td.ID = "yyz100"
	td.Simulated = false
	td.SimulationID = ""
	td.SimulationTransmitSequenceNumber = int32(0)
	td.GrandPrix = "ITALIAN"
	td.Track = "MONZA"
	td.Constructor = "HAAS"
	td.CarNumber = int32(10)
	td.Timestamp = startTime
	td.Latitude = float64(0.0)
	td.Longitude = float64(0.0)
	td.Elevation = float64(0.0)
	td.Description = "G_FORCE"
	td.Unit = "G"
	td.Value = float64(0.0)
	td.HiAlarm = false
	td.LoAlarm = false

	err = td.Persist()
	if err != nil {
		t.Error("failed to persist telemetry datum with error: ", err)
	}

}
