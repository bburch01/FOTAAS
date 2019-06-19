package simulation

import (
	"fmt"
	"testing"
	"time"

	pb "github.com/bburch01/FOTAAS/api"
	d "github.com/bburch01/FOTAAS/fotaas-internal/app/simulation/data"
	tel "github.com/bburch01/FOTAAS/fotaas-internal/app/telemetry"
	ts "github.com/bburch01/FOTAAS/fotaas-internal/pkg/protobuf/timestamp"

	//tel "github.com/bburch01/FOTAAS/fotaas-internal/app/telemetry"

	uid "github.com/google/uuid"
	//ts "github.com/bburch01/FOTAAS/fotaas-internal/pkg/protobuf/timestamp"
)

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

	// Sequential version:
	/*
		if err = StartSimulation(simData, sim); err != nil {
			t.Error("failed with error: ", err)
		}
	*/
}
