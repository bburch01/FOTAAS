package models

import (
	"fmt"
	"log"
	"testing"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

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

	if err = InitDB(); err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize database driver with error: %v", err))
	}

}

func TestSimulationModels(t *testing.T) {

	var sim Simulation
	//var startTime *pbts.Timestamp
	var err error

	/*
		startTime, err = ipbts.TimestampProto(time.Now())
		if err != nil {
			t.Error("failed to create timestamp with error: ", err)
		}
	*/

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

	/*
		sim = Simulation{ID: simID, DurationInMinutes: 60, SampleRate: api.SampleRate_SR_1000_MS, GranPrix: api.GranPrix_ITALIAN, Track: api.Track_MONZA,
			State: "IN_PROGRESS", StartTimestamp: startTime, PercentComplete: 0, SimulationMembers: simMemberMap}
	*/

	sim = Simulation{ID: simID, DurationInMinutes: 60, SampleRate: api.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: api.SimulationRateMultiplier_X1, GranPrix: api.GranPrix_ITALIAN, Track: api.Track_MONZA,
		SimulationMembers: simMemberMap}

	sim.State = "INITIALIZING"
	err = sim.Create()
	if err != nil {
		t.Error("failed to persist simulation with error: ", err)
	}

	sim.StartTimestamp, err = ipbts.TimestampProto(time.Now())
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
	}
	if err := sim.UpdateStartTimestamp(); err != nil {
		t.Error("failed to update simulation start timestamp with error: ", err)
	}

	sim.State = "IN_PROGRESS"
	if err := sim.UpdateState(); err != nil {
		t.Error("failed to update simulation state with error: ", err)
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

func TestRetrieveSimulationInfo(t *testing.T) {

	req := api.GetSimulationInfoRequest{}
	info := api.SimulationInfo{}

	var err error

	req.Uuid = "87bfb12d-63a2-4633-8c5b-ba3d95335c45"

	if info, err = RetrieveSimulationInfo(req); err != nil {
		t.Error("failed to retrieve simulation info with error: ", err)
		t.FailNow()
	}

	fmt.Printf("\nsimulation id       : %v ", info.Uuid)
	fmt.Printf("\nduration in minutes : %v ", info.DurationInMinutes)
	fmt.Printf("\nsample rate         : %v ", info.SampleRate)
	fmt.Printf("\ngran prix           : %v ", info.GranPrix)
	fmt.Printf("\ntrack               : %v ", info.Track)
	fmt.Printf("\nstate               : %v ", info.State)
	fmt.Printf("\nstart timestamp     : %v ", ipbts.TimestampString(info.StartTimestamp))
	fmt.Printf("\nend timestamp       : %v ", ipbts.TimestampString(info.EndTimestamp))
	fmt.Printf("\npercent complete    : %v ", info.PercentComplete)
	fmt.Printf("\nfinal info code   : %v ", info.FinalStatusCode)
	fmt.Printf("\nfinal info message: %v ", info.FinalStatusMessage)
	fmt.Print("\n")

}
