package data

import (

	//ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	//pbts "github.com/golang/protobuf/ptypes/timestamp"

	"sync"
	"testing"

	//"github.com/bburch01/FOTAAS/internal/app/simulation/data"
	"github.com/bburch01/FOTAAS/internal/app/simulation/models"
	"github.com/bburch01/FOTAAS/internal/app/telemetry"

	"github.com/bburch01/FOTAAS/api"
	"github.com/google/uuid"
	//"github.com/bburch01/FOTAAS/internal/app/telemetry"
	//"github.com/bburch01/FOTAAS/internal/app/simulation"
	//ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
	//spinner "github.com/briandowns/spinner"
)

func TestGenerateSimulatedTelemetryDataNoAlarm(t *testing.T) {

	var sampleRateInMillis int32
	var simID string
	var simMember models.SimulationMember
	var sim models.Simulation

	simMemberDataMap := make(map[string]map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData)

	simMemberMap := make(map[string]models.SimulationMember)
	simID = uuid.New().String()

	simMemberID := uuid.New().String()
	simMember = models.SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: api.Constructor_HAAS,
		CarNumber: 8, ForceAlarm: false, NoAlarms: true,
	}
	simMemberMap[simMemberID] = simMember

	simMemberID = uuid.New().String()
	simMember = models.SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: api.Constructor_MERCEDES,
		CarNumber: 44, ForceAlarm: false, NoAlarms: true,
	}
	simMemberMap[simMemberID] = simMember

	sim = models.Simulation{ID: simID, DurationInMinutes: int32(1), SampleRate: api.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: api.SimulationRateMultiplier_X1, GranPrix: api.GranPrix_UNITED_STATES,
		Track: api.Track_AUSTIN, SimulationMembers: simMemberMap}

	var wg sync.WaitGroup
	errChan := make(chan error, len(sim.SimulationMembers))
	resultsChan := make(chan SimMemberData, len(sim.SimulationMembers))
	wg.Add(len(sim.SimulationMembers))
	for _, v := range sim.SimulationMembers {
		go GenerateSimulatedTelemetryData(sim, v, &wg, resultsChan, errChan)
	}
	wg.Wait()
	close(resultsChan)
	close(errChan)

	if err := <-errChan; err != nil {
		t.Error("failed with error from GenerateSimulatedTelemetryData: ", err)
	}

	// Retrieve and aggregate the simulation data (for all of the sim members) from the results channel
	for v := range resultsChan {
		simMemberDataMap[v.SimMemberID] = v.SimData
	}

	switch sim.SampleRate {
	case api.SampleRate_SR_1_MS:
		sampleRateInMillis = 1
	case api.SampleRate_SR_10_MS:
		sampleRateInMillis = 10
	case api.SampleRate_SR_100_MS:
		sampleRateInMillis = 100
	case api.SampleRate_SR_1000_MS:
		sampleRateInMillis = 1000
	default:
		t.Error("invalid sample rate: ", sim.SampleRate)
	}

	simDurationInMillis := sim.DurationInMinutes * 60000
	expectedDatumCount := simDurationInMillis / sampleRateInMillis

	for _, v := range simMemberDataMap {

		if len(v) != len(api.TelemetryDatumDescription_name) {
			t.Error("invalid datum description count, expected: ", len(api.TelemetryDatumDescription_name), "got: ", len(v))
			t.FailNow()
		}

		for _, v2 := range v {
			if int32(len(v2.Data)) != expectedDatumCount {
				t.Error("invalid datum count, expected: ", expectedDatumCount, "got: ", len(v2.Data))
				t.FailNow()
			}
			if v2.AlarmExists {
				t.Error("invalid alarm exists flag, expected false got: ", v2.AlarmExists)
				t.FailNow()
			}

			for _, v3 := range v2.Data {

				if _, err := uuid.Parse(v3.Uuid); err != nil {
					t.Error("invalid datum uuid: ", v3.Uuid)
					t.FailNow()
				}

				if _, ok := api.TelemetryDatumDescription_value[v3.Description.String()]; !ok {
					t.Error("invalid telemetry datum description: ", v3.Description)
					t.FailNow()
				}

				if _, ok := api.TelemetryDatumUnit_value[v3.Unit.String()]; !ok {
					t.Error("invalid telemetry datum unit: ", v3.Unit)
					t.FailNow()
				}

				dp := telemetryDatumParametersMap[v3.Description]

				if !((dp.RangeLowValue <= v3.Value) && (v3.Value <= dp.RangeHighValue)) {
					t.Error("invalid datum value, expected ", dp.RangeLowValue, " <= value <=", dp.RangeHighValue,
						" for ", v3.Description, " got value: ", v3.Value)
					t.FailNow()
				}

				if v3.HighAlarm {
					t.Error("invalid datum high alarm flag, expected false got: ", v3.HighAlarm)
					t.FailNow()
				}

				if v3.LowAlarm {
					t.Error("invalid datum low alarm flag, expected false got: ", v3.LowAlarm)
					t.FailNow()
				}

				if !v3.Simulated {
					t.Error("invalid simulated datum flag, expected true got: ", v3.Simulated)
					t.FailNow()
				}

				if _, err := uuid.Parse(v3.SimulationUuid); err != nil {
					t.Error("invalid datum uuid: ", v3.SimulationUuid)
					t.FailNow()
				}

				if v3.SimulationTransmitSequenceNumber < 0 || v3.SimulationTransmitSequenceNumber >= expectedDatumCount {
					t.Error("invalid datum simulation transmit sequence number, expected 0 <= value <=", expectedDatumCount-1,
						" got value: ", v3.SimulationTransmitSequenceNumber)
					t.FailNow()
				}
			}
		}
	}
}

func TestGenerateSimulatedTelemetryDataForceAlarm(t *testing.T) {

	var sampleRateInMillis int32
	var simID string
	var simMember models.SimulationMember
	var sim models.Simulation

	simMemberDataMap := make(map[string]map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData)

	simMemberMap := make(map[string]models.SimulationMember)
	simID = uuid.New().String()

	simMemberID := uuid.New().String()
	simMember = models.SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: api.Constructor_HAAS,
		CarNumber: 8, ForceAlarm: true, NoAlarms: false,
	}
	simMemberMap[simMemberID] = simMember

	simMemberID = uuid.New().String()
	simMember = models.SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: api.Constructor_MERCEDES,
		CarNumber: 44, ForceAlarm: true, NoAlarms: false,
	}
	simMemberMap[simMemberID] = simMember

	sim = models.Simulation{ID: simID, DurationInMinutes: int32(1), SampleRate: api.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: api.SimulationRateMultiplier_X1, GranPrix: api.GranPrix_UNITED_STATES,
		Track: api.Track_AUSTIN, SimulationMembers: simMemberMap}

	var wg sync.WaitGroup
	errChan := make(chan error, len(sim.SimulationMembers))
	resultsChan := make(chan SimMemberData, len(sim.SimulationMembers))
	wg.Add(len(sim.SimulationMembers))
	for _, v := range sim.SimulationMembers {
		go GenerateSimulatedTelemetryData(sim, v, &wg, resultsChan, errChan)
	}
	wg.Wait()
	close(resultsChan)
	close(errChan)

	if err := <-errChan; err != nil {
		t.Error("failed with error from GenerateSimulatedTelemetryData: ", err)
	}

	// Retrieve and aggregate the simulation data (for all of the sim members) from the results channel
	for v := range resultsChan {
		simMemberDataMap[v.SimMemberID] = v.SimData
	}

	switch sim.SampleRate {
	case api.SampleRate_SR_1_MS:
		sampleRateInMillis = 1
	case api.SampleRate_SR_10_MS:
		sampleRateInMillis = 10
	case api.SampleRate_SR_100_MS:
		sampleRateInMillis = 100
	case api.SampleRate_SR_1000_MS:
		sampleRateInMillis = 1000
	default:
		t.Error("invalid sample rate: ", sim.SampleRate)
	}

	simDurationInMillis := sim.DurationInMinutes * 60000
	expectedDatumCount := simDurationInMillis / sampleRateInMillis

	for _, v := range simMemberDataMap {

		if len(v) != len(api.TelemetryDatumDescription_name) {
			t.Error("invalid datum description count, expected: ", len(api.TelemetryDatumDescription_name), "got: ", len(v))
			t.FailNow()
		}

		alarmCount := 0
		for _, v2 := range v {

			if int32(len(v2.Data)) != expectedDatumCount {
				t.Error("invalid datum count, expected: ", expectedDatumCount, "got: ", len(v2.Data))
				t.FailNow()
			}

			if v2.AlarmExists {
				alarmCount++
			}

			datumWithAlarmCount := 0
			for i, v3 := range v2.Data {

				if _, err := uuid.Parse(v3.Uuid); err != nil {
					t.Error("invalid datum uuid: ", v3.Uuid)
					t.FailNow()
				}

				if _, ok := api.TelemetryDatumDescription_value[v3.Description.String()]; !ok {
					t.Error("invalid telemetry datum description: ", v3.Description)
					t.FailNow()
				}

				if _, ok := api.TelemetryDatumUnit_value[v3.Unit.String()]; !ok {
					t.Error("invalid telemetry datum unit: ", v3.Unit)
					t.FailNow()
				}

				if v3.HighAlarm && v3.LowAlarm {
					t.Error("telemetry datum high alarm and low alarm both set to true")
					t.FailNow()
				}

				if v3.HighAlarm || v3.LowAlarm {
					datumWithAlarmCount++
				}

				if !v3.Simulated {
					t.Error("invalid simulated datum flag, expected true got: ", v3.Simulated)
					t.FailNow()
				}

				if _, err := uuid.Parse(v3.SimulationUuid); err != nil {
					t.Error("invalid datum uuid: ", v3.SimulationUuid)
					t.FailNow()
				}

				if v3.SimulationTransmitSequenceNumber < 0 || v3.SimulationTransmitSequenceNumber >= expectedDatumCount {
					t.Error("invalid datum simulation transmit sequence number, expected 0 <= value <=", expectedDatumCount-1,
						" got value: ", v3.SimulationTransmitSequenceNumber)
					t.FailNow()
				}

				// Confirm that all datum values preceeding the alarm value are within the valid range and that all
				// datum values following the alarm datum have been set to 0.0 (as per design). Else f there is no alarm
				// for the datum description, confirm that all datum are in valid range.
				dp := telemetryDatumParametersMap[v3.Description]
				if v2.AlarmExists {
					if i < v2.AlarmIndex {
						switch v2.AlarmMode {
						case telemetry.Low:
							if !((dp.LowAlarmValue <= v3.Value) && (v3.Value <= dp.RangeHighValue)) {
								t.Error("datum index: ", i, " invalid pre-alarm datum value ", v3.Value,
									" expected to be between ", dp.LowAlarmValue, " and ", dp.RangeHighValue)
							}
						case telemetry.High:
							if !((dp.RangeLowValue <= v3.Value) && (v3.Value <= dp.HighAlarmValue)) {
								t.Error("datum index: ", i, " invalid pre-alarm datum value ", v3.Value,
									" expected to be between ", dp.RangeLowValue, " and ", dp.HighAlarmValue)
							}
						default:
							t.Error("simulatedTelemetryData alarm mode invalid")
						}
					} else if i > v2.AlarmIndex {
						if v3.Value != 0.0 {
							t.Error("invalid post-alarm datum value, expected 0.0 got ", v3.Value)
						}
					}
				} else {

					if !((dp.RangeLowValue <= v3.Value) && (v3.Value <= dp.RangeHighValue)) {
						t.Error("invalid datum value, expected ", dp.RangeLowValue, " <= value <=", dp.RangeHighValue,
							" for ", v3.Description, " got value: ", v3.Value)
						t.FailNow()
					}

				}

			}
			if v2.AlarmExists && datumWithAlarmCount != 1 {
				t.Error("incorrect datum with alarm count, expected 1 got ", datumWithAlarmCount)
			}

		}
		if alarmCount == 0 {
			t.Error("no alarm found in telemetry data")
			t.FailNow()
		} else if alarmCount > 1 {
			t.Error("more than 1 alarm found in telemetry data")
			t.FailNow()

		}
	}
}

/*
func BenchmarkGenerateSimulatedTelemetryDataNoAlarm(b *testing.B) {

	var sim models.Simulation
	var err error
	var simDurationInMinutes int32
	var sampleRate string

	simDurationInMinutes = 1
	sampleRate = api.SampleRate_SR_1000_MS.String()

	sim = models.Simulation{ID: uuid.New().String(), DurationInMinutes: simDurationInMinutes,
		SampleRate: sampleRate, GranPrix: api.GranPrix_UNITED_STATES.String(), Track: api.Track_AUSTIN.String(),
	}

	s := spinner.New(spinner.CharSets[7], 100*time.Millisecond)
	s.Prefix = "Benchmark GenerateSimulatedTelemetryDataNoAlarm in progress: "
	s.Start()
	for i := 0; i < b.N; i++ {
		if _, err = GenerateSimulatedTelemetryData(sim); err != nil {
			b.Error("failed to generate simulation data with error: ", err)
		}
	}
	s.Stop()

}
*/

/*
func BenchmarkGenerateSimulatedTelemetryDataNoAlarmSequential(b *testing.B) {

	var sim api.Simulation
	var err error
	var simDurationInMinutes int32
	var sampleRate api.SampleRate

	simDurationInMinutes = 1
	sampleRate = api.SampleRate_SR_1000_MS

	sim = api.Simulation{Uuid: uuid.New().String(), DurationInMinutes: simDurationInMinutes,
		SampleRate: sampleRate, GranPrix: api.GranPrix_UNITED_STATES,
		Track: api.Track_AUSTIN,
	}

	s := spinner.New(spinner.CharSets[7], 100*time.Millisecond)
	s.Prefix = "Benchmark GenerateSimulatedTelemetryDataNoAlarmSequential in progress: "
	s.Start()
	for i := 0; i < b.N; i++ {
		if _, err = generateSimulatedTelemetryDataSequential(sim); err != nil {
			b.Error("failed to generate simulation data with error: ", err)
		}
	}
	s.Stop()

}
*/
