package models

import (
	"testing"
	"time"

	//mdl "github.com/bburch01/FOTAAS/internal/app/simulation/models"
	ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

func TestSimulationModels(t *testing.T) {

	var sim Simulation
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

	err = sim.Persist()
	if err != nil {
		t.Error("failed to persist simulation with error: ", err)
	}

}
