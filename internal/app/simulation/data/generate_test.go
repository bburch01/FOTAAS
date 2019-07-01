package data

import (
	"fmt"
	"testing"
	"time"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/app/telemetry"
	"github.com/briandowns/spinner"
	"github.com/google/uuid"
	//ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
	//spinner "github.com/briandowns/spinner"
)

func TestGenerateSimulatedTelemetryDataForceAlarm(t *testing.T) {

	var sim api.Simulation
	//var tstamp time.Time
	var simData map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
	var err error
	var simDurationInMinutes int32
	var sampleRate api.SampleRate
	var sampleRateInMilliseconds int32
	var expectedSimDataLength int32
	var actualSimDataLength int32

	simDurationInMinutes = 1
	sampleRate = api.SampleRate_SR_1000_MS

	sim = api.Simulation{Uuid: uuid.New().String(), DurationInMinutes: simDurationInMinutes, SampleRate: sampleRate,
		GrandPrix: api.GrandPrix_UNITED_STATES, Track: api.Track_AUSTIN,
	}

	if simData, err = GenerateSimulatedTelemetryData(sim); err != nil {
		t.Error("failed to generate simulation data with error: ", err)
	}

	expectedSimDataLength = int32(len(telemetryDatumParametersMap))
	actualSimDataLength = int32(len(simData))
	if actualSimDataLength != expectedSimDataLength {
		t.Error("failed with incorrect simData length, expected: ", expectedSimDataLength, "got: ", actualSimDataLength)
	}

	// Only 1 of the simData should have alarmExists set to true.
	var simDataAlarmCount int
	for _, v := range simData {
		if v.AlarmExists {
			simDataAlarmCount++
		}
	}
	if simDataAlarmCount != 1 {
		t.Error("incorrect simData alarm exists count, expected 1 got: ", simDataAlarmCount)
	}

	simDurationInMilliseconds := sim.DurationInMinutes * 60000

	switch sampleRate {
	case api.SampleRate_SR_1_MS:
		sampleRateInMilliseconds = 1
	case api.SampleRate_SR_10_MS:
		sampleRateInMilliseconds = 10
	case api.SampleRate_SR_100_MS:
		sampleRateInMilliseconds = 100
	case api.SampleRate_SR_1000_MS:
		sampleRateInMilliseconds = 1000
	default:
		t.Error("invalid sample rate")
	}

	var expectedDatumCount = simDurationInMilliseconds / sampleRateInMilliseconds
	for _, v1 := range simData {
		if v1.AlarmExists {
			datumCount := int32(len(v1.Data))
			if datumCount != expectedDatumCount {
				t.Error("incorrect datum count, expected: ", expectedDatumCount, " got: ", datumCount)
			}
		}
	}

	//TODO: remove, for debug only
	/*
		for _, v1 := range simData {
			if v1.AlarmExists {
				for idx, v2 := range v1.Data {
					logger.Debug(fmt.Sprintf("idx: %v desc: %v high alarm: %v low alarm %v value: %v, alarm index: %v", idx, v2.Description.String(),
						v2.HighAlarm, v2.LowAlarm, v2.Value, v1.AlarmIndex))
				}
			}
		}
	*/

	// Confirm that only 1 datum is the alarm datum and that it has either high alarm or
	// low alarm set but not both.
	var datumWithAlarmCount int
	for _, v1 := range simData {
		if v1.AlarmExists {
			for _, v2 := range v1.Data {
				if v2.HighAlarm && v2.LowAlarm {
					t.Error("datum high alarm and low alarm both set to true")
				}
				if v2.HighAlarm || v2.LowAlarm {
					datumWithAlarmCount++
				}
			}
		}
	}
	if datumWithAlarmCount != 1 {
		t.Error("incorrect datum with alarm count, expected 1 got ", datumWithAlarmCount)
	}

	// Confirm that the alarm datum value is within the valid range.
	for _, v := range simData {
		if v.AlarmExists {
			if !((0 <= v.AlarmIndex) && (v.AlarmIndex <= len(v.Data))) {
				t.Error("simulatedTelemetryData alarm index is invalid")
			}
			alarmDatum := v.Data[v.AlarmIndex]
			dp := telemetryDatumParametersMap[v.DatumDesc]
			switch v.AlarmMode {
			case telemetry.Low:
				if !(alarmDatum.Value <= dp.LowAlarmValue) {
					t.Error("invalid datum value, expected ", alarmDatum.Value, " to be <= ", dp.LowAlarmValue)
				}
			case telemetry.High:
				if !(alarmDatum.Value >= dp.HighAlarmValue) {
					t.Error("invalid datum value, expected ", alarmDatum.Value, " to be >= ", dp.HighAlarmValue)
				}
			default:
				t.Error("simulatedTelemetryData alarm mode invalid")
			}
			break
		}
	}

	// Confirm that all datum values preceeding the alarm value are within the valid range and that all
	// datum values following the alarm datum have been set to 0.0 .
	for _, v1 := range simData {
		if v1.AlarmExists {
			dp := telemetryDatumParametersMap[v1.DatumDesc]
			for i, v2 := range v1.Data {
				if i < v1.AlarmIndex {
					switch v1.AlarmMode {
					case telemetry.Low:
						if !((dp.LowAlarmValue <= v2.Value) && (v2.Value <= dp.RangeHighValue)) {
							t.Error("datum index: ", i, " invalid pre-alarm datum value ", v2.Value,
								" expected to be between ", dp.LowAlarmValue, " and ", dp.RangeHighValue)
						}
					case telemetry.High:
						if !((dp.RangeLowValue <= v2.Value) && (v2.Value <= dp.HighAlarmValue)) {
							t.Error("datum index: ", i, " invalid pre-alarm datum value ", v2.Value,
								" expected to be between ", dp.RangeLowValue, " and ", dp.HighAlarmValue)
						}
					default:
						t.Error("simulatedTelemetryData alarm mode invalid")
					}
				} else if i > v1.AlarmIndex {
					if v2.Value != 0.0 {
						t.Error("invalid post-alarm datum value, expected 0.0 got ", v2.Value)
					}
				}
			}
			break
		}
	}

	/*
		for _, v1 := range simData {
			if v1.AlarmExists {
				logger.Debug(fmt.Sprintf("simData datumDesc: %v alarmExists: %v", v1.DatumDesc, v1.AlarmExists))
				for _, v2 := range v1.Data {
					if tstamp, err = ts.Timestamp(v2.Timestamp); err != nil {
						t.Error("failed to convert google.protobuf.timestamp to time.Time with error: ", err)
					}
					logger.Debug(fmt.Sprintf("datum uuid: %v desc: %v unit: %v timestamp: %v value: %v", v2.Uuid, v2.Description.String(),
						v2.Unit.String(), tstamp, v2.Value))
				}
			}
		}
	*/

}

func TestGenerateSimulatedTelemetryDataNoAlarm(t *testing.T) {

	var sim api.Simulation
	//var tstamp time.Time
	var simData map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
	var err error
	var simDurationInMinutes int32
	var sampleRate api.SampleRate
	var sampleRateInMilliseconds int32
	var expectedSimDataLength int32
	var actualSimDataLength int32
	var startTime time.Time
	var elapsedTime time.Duration

	simDurationInMinutes = 1
	sampleRate = api.SampleRate_SR_1000_MS

	sim = api.Simulation{Uuid: uuid.New().String(), DurationInMinutes: simDurationInMinutes, SampleRate: sampleRate,
		GrandPrix: api.GrandPrix_UNITED_STATES, Track: api.Track_AUSTIN,
	}

	/*
		var sim api.Simulation
		//var tstamp time.Time
		var simData map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
		var err error
		var simDurationInMinutes int32
		var sampleRateInMilliseconds int32
		var expectedSimDataLength int32
		var actualSimDataLength int32
		var startTime time.Time
		var elapsedTime time.Duration

		simDurationInMinutes = 1
		sampleRateInMilliseconds = 1

		sim = api.Simulation{Uuid: uuid.New().String(), DurationInMinutes: simDurationInMinutes,
			SampleRateInMilliseconds: sampleRateInMilliseconds, TransmitRateInMilliseconds: 5, GrandPrix: api.GrandPrix_UNITED_STATES,
			Track: api.Track_AUSTIN, Constructor: api.Constructor_HAAS, CarNumber: 8, ForceAlarm: false, NoAlarms: true,
		}
	*/

	//cores := runtime.NumCPU()
	//logger.Debug(fmt.Sprintf("this machine has %v CPU cores available to this process", cores))

	// There is a chance (about 5%) that generateSimulatedTelemetryData() will create an alarm for one of the
	// datum descriptions when ForceAlarm is set to false (i.e. and un-forced alarm will occur in the data).
	// This test case is specifically for NO alarms in the simulated data so call generateSimulatedTelemetryData()
	// until alarm exists is false.
	/*
		var alarmFound = true
		for alarmFound {
			startTime = time.Now()
			if simData, err = generateSimulatedTelemetryData(sim); err != nil {
				t.Error("failed to generate simulation data with error: ", err)
			}
			elapsedTime = time.Since(startTime)
			logger.Debug(fmt.Sprintf("generateSimulatedTelemetryData() execution time: %v", elapsedTime))
			//t.Logf("generateSimulatedTelemetryData() execution time: %v", elapsedTime)
			alarmFound = false
			for _, v := range simData {
				if v.AlarmExists {
					logger.Debug("alarm found in data, regenerating data set")
					alarmFound = true
					break
				}
			}
		}
	*/

	startTime = time.Now()
	if simData, err = GenerateSimulatedTelemetryData(sim); err != nil {
		t.Error("failed to generate simulation data with error: ", err)
	}
	elapsedTime = time.Since(startTime)
	logger.Debug(fmt.Sprintf("generateSimulatedTelemetryData() execution time: %v", elapsedTime))

	// The number of simulatedTelemetryData in the simData map must equal the length
	// of the telemetryDataParametersmap (one for each datum description).
	expectedSimDataLength = int32(len(telemetryDatumParametersMap))
	actualSimDataLength = int32(len(simData))
	if actualSimDataLength != expectedSimDataLength {
		t.Error("failed with incorrect simData length, expected: ", expectedSimDataLength, "got: ", actualSimDataLength)
	}

	simDurationInMilliseconds := sim.DurationInMinutes * 60000

	switch sampleRate {
	case api.SampleRate_SR_1_MS:
		sampleRateInMilliseconds = 1
	case api.SampleRate_SR_10_MS:
		sampleRateInMilliseconds = 10
	case api.SampleRate_SR_100_MS:
		sampleRateInMilliseconds = 100
	case api.SampleRate_SR_1000_MS:
		sampleRateInMilliseconds = 1000
	default:
		t.Error("invalid sample rate")
	}

	// The datum count in each of the simulatedTelemetryData in the simData map must equal
	// what the sample rate will produce during the duration of the simulation.
	var expectedDatumCount = simDurationInMilliseconds / sampleRateInMilliseconds
	for _, v1 := range simData {
		datumCount := int32(len(v1.Data))
		if datumCount != expectedDatumCount {
			t.Error("incorrect datum count, expected: ", expectedDatumCount, " got: ", datumCount)
		}
	}

	// Confirm that all of the generated datum values are within the valid range.
	for _, v1 := range simData {
		dp := telemetryDatumParametersMap[v1.DatumDesc]
		for i, v2 := range v1.Data {
			if !((dp.RangeLowValue <= v2.Value) && (v2.Value <= dp.RangeHighValue)) {
				t.Error("datum index: ", i, " invalid datum value ", v2.Value,
					" expected to be between ", dp.RangeLowValue, " and ", dp.RangeHighValue)
			}
		}
	}
}

func BenchmarkGenerateSimulatedTelemetryDataNoAlarm(b *testing.B) {

	var sim api.Simulation
	var err error
	var simDurationInMinutes int32
	var sampleRate api.SampleRate

	simDurationInMinutes = 1
	sampleRate = api.SampleRate_SR_1000_MS

	sim = api.Simulation{Uuid: uuid.New().String(), DurationInMinutes: simDurationInMinutes,
		SampleRate: sampleRate, GrandPrix: api.GrandPrix_UNITED_STATES,
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

/*
func BenchmarkGenerateSimulatedTelemetryDataNoAlarmSequential(b *testing.B) {

	var sim api.Simulation
	var err error
	var simDurationInMinutes int32
	var sampleRate api.SampleRate

	simDurationInMinutes = 1
	sampleRate = api.SampleRate_SR_1000_MS

	sim = api.Simulation{Uuid: uuid.New().String(), DurationInMinutes: simDurationInMinutes,
		SampleRate: sampleRate, GrandPrix: api.GrandPrix_UNITED_STATES,
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
