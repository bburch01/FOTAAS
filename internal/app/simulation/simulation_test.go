package simulation

import (
	"fmt"
	"testing"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/app/simulation/data"
	"github.com/bburch01/FOTAAS/internal/app/telemetry"
	"github.com/google/uuid"
	// 	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	//tel "github.com/bburch01/FOTAAS/internal/app/telemetry"
	//ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
)

func TestStartSimulation(t *testing.T) {

	var simData map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
	var err error
	var tstamp time.Time

	simDurationInMinutes := int32(1)
	sampleRate := api.SampleRate_SR_1000_MS

	sim := api.Simulation{Uuid: uuid.New().String(), DurationInMinutes: simDurationInMinutes, SampleRate: sampleRate,
		SimulationRateMultiplier: api.SimulationRateMultiplier_X20, GrandPrix: api.GrandPrix_UNITED_STATES,
		Track: api.Track_AUSTIN,
	}

	if simData, err = data.GenerateSimulatedTelemetryData(sim); err != nil {
		t.Error("failed to generate simulation data with error: ", err)
	}

	for _, v1 := range simData {
		//logger.Debug(fmt.Sprintf("simData datumDesc: %v alarmExists: %v", v1.DatumDesc, v1.AlarmExists))
		for _, v2 := range v1.Data {
			if tstamp, err = ipbts.Timestamp(v2.Timestamp); err != nil {
				t.Error("failed to convert google.protobuf.timestamp to time.Time with error: ", err)
			}
			logger.Debug(fmt.Sprintf("datum desc: %v unit: %v timestamp: %v value: %v", v2.Description.String(),
				v2.Unit.String(), tstamp, v2.Value))
		}
	}
}
