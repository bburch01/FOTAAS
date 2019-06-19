package data

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	pb "github.com/bburch01/FOTAAS/api"
	ts "github.com/bburch01/FOTAAS/fotaas-internal/pkg/protobuf/timestamp"

	//tel "github.com/bburch01/FOTAAS/fotaas-internal/app/telemetry"
	tel "github.com/bburch01/FOTAAS/fotaas-internal/app/telemetry"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"

	//ts "github.com/bburch01/FOTAAS/fotaas-internal/pkg/protobuf/timestamp"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"

	uid "github.com/google/uuid"
	randutil "github.com/jmcvetta/randutil"

	//randutil "github.com/jmcvetta/randutil"
	//uid "github.com/google/uuid"
	//uuid "github.com/satori/go.uuid"

	logging "github.com/bburch01/FOTAAS/fotaas-internal/pkg/logging"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	//logging "github.com/bburch01/FOTAAS/fotaas-internal/pkg/logging"
	//"github.com/joho/godotenv"
	//"go.uber.org/zap"
)

var logger *zap.Logger

func init() {

	var lm logging.LogMode
	var err error

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err = godotenv.Load(); err != nil {
		log.Panicf("failed to load environment variables with error: %v", err)
	}

	if lm, err = logging.LogModeForString(os.Getenv("LOG_MODE")); err != nil {
		log.Panicf("failed to initialize logging subsystem with error: %v", err)
	}

	if logger, err = logging.NewLogger(lm, os.Getenv("LOG_DIR"), os.Getenv("LOG_FILE_NAME")); err != nil {
		log.Panicf("failed to initialize logging subsystem with error: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

}

var alarmEventChoices = []randutil.Choice{
	{Weight: 5, Item: true},
	{Weight: 95, Item: false},
}

type rampDirection int

const (
	up rampDirection = iota
	down
)

func (rd rampDirection) String() string {
	return [...]string{"up", "down"}[rd]
}

var alarmRampDirectionChoices = []randutil.Choice{
	{Weight: 50, Item: up},
	{Weight: 50, Item: down},
}

type raceSegment int

const (
	initial raceSegment = iota
	intermediate
	final
)

/*
type alarmMode int

const (
	high alarmMode = iota
	low
)

func (am alarmMode) String() string {
	return [...]string{"high", "low"}[am]
}
*/

/*
type alarmParams struct {
	desc pb.TelemetryDatumDescription
	mode alarmMode
}
*/

var alarmTypeChoices = []randutil.Choice{
	{Weight: 35, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_TIRE_PRESSURE_FL, Mode: tel.Low}},
	{Weight: 35, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_TIRE_PRESSURE_FR, Mode: tel.Low}},
	{Weight: 35, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_TIRE_PRESSURE_RL, Mode: tel.Low}},
	{Weight: 35, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_TIRE_PRESSURE_RR, Mode: tel.Low}},
	{Weight: 15, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_ENGINE_OIL_PRESSURE, Mode: tel.High}},
	{Weight: 15, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_ENGINE_OIL_PRESSURE, Mode: tel.Low}},
	{Weight: 15, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_ENGINE_COOLANT_TEMP, Mode: tel.High}},
	{Weight: 10, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_ENGINE_OIL_TEMP, Mode: tel.High}},
	{Weight: 10, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_BRAKE_TEMP_FL, Mode: tel.High}},
	{Weight: 10, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_BRAKE_TEMP_FR, Mode: tel.High}},
	{Weight: 10, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_BRAKE_TEMP_RL, Mode: tel.High}},
	{Weight: 10, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_BRAKE_TEMP_RR, Mode: tel.High}},
	{Weight: 5, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_MGUK_OUTPUT, Mode: tel.High}},
	{Weight: 5, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_MGUK_OUTPUT, Mode: tel.Low}},
	{Weight: 5, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_MGUH_OUTPUT, Mode: tel.High}},
	{Weight: 5, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_MGUH_OUTPUT, Mode: tel.Low}},
	{Weight: 2, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_ENERGY_STORAGE_LEVEL, Mode: tel.High}},
	{Weight: 2, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_ENERGY_STORAGE_LEVEL, Mode: tel.Low}},
	{Weight: 2, Item: tel.AlarmParams{Desc: pb.TelemetryDatumDescription_ENERGY_STORAGE_TEMP, Mode: tel.High}},
}

/*
type telemetryDatumParameters struct {
	unit           pb.TelemetryDatumUnit
	RangeLowValue  float64
	RangeHighValue float64
	highAlarmValue float64
	lowAlarmValue  float64
}

type tel.SimulatedTelemetryData struct {
	datumDesc   pb.TelemetryDatumDescription
	data        []pb.TelemetryDatum
	alarmExists bool
	alarmMode   alarmMode
	alarmIndex  int
}
*/

var telemetryDatumParametersMap = map[pb.TelemetryDatumDescription]tel.TelemetryDatumParameters{
	pb.TelemetryDatumDescription_BRAKE_TEMP_FL: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 750.0, RangeHighValue: 1050.0,
		HighAlarmValue: 1300.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_BRAKE_TEMP_FR: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 750.0, RangeHighValue: 1050.0,
		HighAlarmValue: 1300.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_BRAKE_TEMP_RL: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 750.0, RangeHighValue: 1050.0,
		HighAlarmValue: 1300.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_BRAKE_TEMP_RR: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 750.0, RangeHighValue: 1050.0,
		HighAlarmValue: 1300.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_ENERGY_STORAGE_LEVEL: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_MJ, RangeLowValue: 1.3, RangeHighValue: 3.8,
		HighAlarmValue: 4.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_ENERGY_STORAGE_TEMP: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 50.0, RangeHighValue: 55.0,
		HighAlarmValue: 60.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_ENGINE_COOLANT_TEMP: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 110.0, RangeHighValue: 120.0,
		HighAlarmValue: 140.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_ENGINE_OIL_PRESSURE: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_KPA, RangeLowValue: 300.0, RangeHighValue: 400.0,
		HighAlarmValue: 550.0, LowAlarmValue: 40.0,
	},
	pb.TelemetryDatumDescription_ENGINE_OIL_TEMP: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 110.0, RangeHighValue: 120.0,
		HighAlarmValue: 140.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_ENGINE_RPM: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_RPM, RangeLowValue: 2500.0, RangeHighValue: 13500.00,
		HighAlarmValue: 15000.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_FUEL_CONSUMED: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_KG, RangeLowValue: 0, RangeHighValue: 120.0,
		HighAlarmValue: 125.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_FUEL_FLOW: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_KG_PER_HOUR, RangeLowValue: 10.0, RangeHighValue: 80.0,
		HighAlarmValue: 100.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_G_FORCE: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_G, RangeLowValue: 2.0, RangeHighValue: 6.0,
		HighAlarmValue: 8.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_G_FORCE_DIRECTION: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_RADIAN, RangeLowValue: 0, RangeHighValue: 6.280,
		HighAlarmValue: 0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_MGUH_OUTPUT: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_JPS, RangeLowValue: 16.0, RangeHighValue: 19.0,
		HighAlarmValue: 25.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_MGUK_OUTPUT: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_JPS, RangeLowValue: 16.0, RangeHighValue: 19.0,
		HighAlarmValue: 25.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_SPEED: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_KPH, RangeLowValue: 100.0, RangeHighValue: 350.0,
		HighAlarmValue: 400.0, LowAlarmValue: 0,
	},
	pb.TelemetryDatumDescription_TIRE_PRESSURE_FL: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_BAR, RangeLowValue: 1.1, RangeHighValue: 1.3,
		HighAlarmValue: 1.6, LowAlarmValue: 0.8,
	},
	pb.TelemetryDatumDescription_TIRE_PRESSURE_FR: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_BAR, RangeLowValue: 1.1, RangeHighValue: 1.3,
		HighAlarmValue: 1.6, LowAlarmValue: 0.8,
	},
	pb.TelemetryDatumDescription_TIRE_PRESSURE_RL: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_BAR, RangeLowValue: 1.1, RangeHighValue: 1.3,
		HighAlarmValue: 1.6, LowAlarmValue: 0.8,
	},
	pb.TelemetryDatumDescription_TIRE_PRESSURE_RR: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_BAR, RangeLowValue: 1.1, RangeHighValue: 1.3,
		HighAlarmValue: 1.6, LowAlarmValue: 0.8,
	},
	pb.TelemetryDatumDescription_TIRE_TEMP_FL: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 80.0, RangeHighValue: 120.0,
		HighAlarmValue: 130.0, LowAlarmValue: 70.0,
	},
	pb.TelemetryDatumDescription_TIRE_TEMP_FR: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 80.0, RangeHighValue: 120.0,
		HighAlarmValue: 130.0, LowAlarmValue: 70.0,
	},
	pb.TelemetryDatumDescription_TIRE_TEMP_RL: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 80.0, RangeHighValue: 120.0,
		HighAlarmValue: 130.0, LowAlarmValue: 70.0,
	},
	pb.TelemetryDatumDescription_TIRE_TEMP_RR: tel.TelemetryDatumParameters{
		Unit: pb.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 80.0, RangeHighValue: 120.0,
		HighAlarmValue: 130.0, LowAlarmValue: 70.0,
	},
}

func GenerateSimulatedTelemetryData(sim pb.Simulation) (map[pb.TelemetryDatumDescription]tel.SimulatedTelemetryData, error) {

	//var datumCount = (sim.DurationInMinutes * 60000) / (sim.SampleRateInMilliseconds)

	var datumCount int32
	var sampleRateInMillis int32
	var simDurationInMillis int32

	var genAlarmChoice randutil.Choice
	var alarmTypeChoice randutil.Choice
	var err error
	var simStartTime = time.Now()
	//var currentSimTime time.Time
	//var datumTimestamp *timestamp.Timestamp
	var genAlarm bool
	//var stData = make([]tel.SimulatedTelemetryData, len(telemetryDatumParametersMap))
	var simulatedTelemetryDataMap = make(map[pb.TelemetryDatumDescription]tel.SimulatedTelemetryData)

	simDurationInMillis = sim.DurationInMinutes * 60000
	switch sim.SampleRate {
	case pb.SampleRate_SR_1_MS:
		sampleRateInMillis = 1
	case pb.SampleRate_SR_10_MS:
		sampleRateInMillis = 10
	case pb.SampleRate_SR_100_MS:
		sampleRateInMillis = 100
	case pb.SampleRate_SR_1000_MS:
		sampleRateInMillis = 1000
	default:
		return simulatedTelemetryDataMap, errors.New("invalid sample rate in millis")
	}

	datumCount = simDurationInMillis / sampleRateInMillis

	// ForceAlarm false, NoAlarms false: alarm generated based on the probabilities declared in alarmEventChoices
	// ForceAlarm true, NoAlarms false: force the generation of an alarm
	// ForceAlarm false, NoAlarms true: do not generate an alarm
	if !sim.ForceAlarm {
		// Determine if there will be an alarm event during the simulation.
		if genAlarmChoice, err = randutil.WeightedChoice(alarmEventChoices); err != nil {
			return simulatedTelemetryDataMap, err
		}
		genAlarm = genAlarmChoice.Item.(bool)

	} else {
		genAlarm = true
	}

	if !sim.ForceAlarm && !sim.NoAlarms {
		// Determine if there will be an alarm event during the simulation.
		if genAlarmChoice, err = randutil.WeightedChoice(alarmEventChoices); err != nil {
			return simulatedTelemetryDataMap, err
		}
		genAlarm = genAlarmChoice.Item.(bool)

	} else if sim.ForceAlarm && !sim.NoAlarms {
		genAlarm = true
	} else if !sim.ForceAlarm && sim.NoAlarms {
		genAlarm = false
	} else {
		return simulatedTelemetryDataMap, errors.New("simulation ForceAlarm & NoAlarm must not both be set to true")
	}

	// Alarm or not, get an alarmTypeChoice to keep the compiler happy.
	if alarmTypeChoice, err = randutil.WeightedChoice(alarmTypeChoices); err != nil {
		return simulatedTelemetryDataMap, err
	}

	errChan := make(chan error, len(telemetryDatumParametersMap))
	resultsChan := make(chan tel.SimulatedTelemetryData, len(telemetryDatumParametersMap))
	sem := make(chan int, runtime.NumCPU())

	var wg sync.WaitGroup
	wg.Add(len(telemetryDatumParametersMap))

	for datumDesc, datumParams := range telemetryDatumParametersMap {
		go telemetryDataGenerationWorker(sim.Uuid, datumDesc, datumParams, sampleRateInMillis, alarmTypeChoice.Item.(tel.AlarmParams), datumCount,
			simStartTime, genAlarm, sem, &wg, resultsChan, errChan)
	}

	wg.Wait()
	close(errChan)
	close(resultsChan)

	if err := <-errChan; err != nil {
		logger.Error(fmt.Sprintf("goroutine error: %v", err))
		return simulatedTelemetryDataMap, err
	}

	/*
		ap := alarmTypeChoice.Item.(alarmParams)

		for datumDesc, tel.DatumParams := range telemetryDatumParametersMap {
			logger.Debug(fmt.Sprintf("generating telemetry data for datum desc: %v", datumDesc.String()))
			values := randFloatsInRange(tel.DatumParams.RangeLowValue, tel.DatumParams.RangeHighValue, datumCount)
			data := make([]pb.TelemetryDatum, datumCount)
			for i, v := range values {
				data[i].Value = v
			}
			simData := tel.SimulatedTelemetryData{datumDesc: datumDesc, data: data, alarmExists: false,
				alarmMode: ap.Mode, alarmIndex: 0}
			if genAlarm {
				if datumDesc.String() == ap.Desc.String() {
					switch ap.Mode {
					case high:
						logger.Debug(fmt.Sprintf("alarm.Desc: %v alarm.Mode: %v range low: %v range high: %v"+
							"ramp dir: up high alarm level: %v", ap.Desc, ap.Mode.String(), tel.DatumParams.RangeLowValue,
							tel.DatumParams.RangeHighValue, tel.DatumParams.HighAlarmValue))
						if err := rampToAlarm(&simData, tel.DatumParams.RangeLowValue, tel.DatumParams.RangeHighValue, up,
							tel.DatumParams.HighAlarmValue); err != nil {
							return tel.SimulatedTelemetryDataMap, err
						}
					case low:
						logger.Debug(fmt.Sprintf("alarm.Desc: %v alarm.Mode: %v range low: %v range high: %v"+
							"ramp dir: down low alarm level: %v", ap.Desc, ap.Mode.String(), tel.DatumParams.RangeLowValue,
							tel.DatumParams.RangeHighValue, tel.DatumParams.LowAlarmValue))
						if err := rampToAlarm(&simData, tel.DatumParams.RangeLowValue, tel.DatumParams.RangeHighValue, down,
							tel.DatumParams.LowAlarmValue); err != nil {
							return tel.SimulatedTelemetryDataMap, err
						}
					}
				}
			}
			currentSimTime = simStartTime
			for i := range simData.Data {
				if i > 0 {
					currentSimTime = currentSimTime.Add(time.Second)
				}
				if datumTimestamp, err = ts.TimestampProto(currentSimTime); err != nil {
					return tel.SimulatedTelemetryDataMap, err
				}
				simData.Data[i].Uuid = uid.New().String()
				simData.Data[i].Description = datumDesc
				simData.Data[i].Unit = tel.DatumParams.unit
				simData.Data[i].Timestamp = datumTimestamp
				//TODO: Currently, lat, long, & elevation are not modeled in the simulation.
				simData.Data[i].Latitude = 0.0
				simData.Data[i].Longitude = 0.0
				simData.Data[i].Elevation = 0.0
				// The datum Value and (if an alarm occurred) HighAlarm (or LowAlarm) were set in
				// the rampToAlarm() function.
			}
			tel.SimulatedTelemetryDataMap[datumDesc] = simData
		}
	*/

	for std := range resultsChan {
		//logger.Debug(fmt.Sprintf("adding sim data for datum desc: %v to the sim data map", std.datumDesc))
		simulatedTelemetryDataMap[std.DatumDesc] = std
	}

	return simulatedTelemetryDataMap, nil
}

func generateSimulatedTelemetryDataSequential(sim pb.Simulation) (map[pb.TelemetryDatumDescription]tel.SimulatedTelemetryData, error) {

	//var datumCount = (sim.DurationInMinutes * 60000) / (sim.SampleRateInMilliseconds)

	var datumCount int32
	var sampleRateInMillis int32
	var simDurationInMillis int32

	var genAlarmChoice randutil.Choice
	var alarmTypeChoice randutil.Choice
	var err error
	var simStartTime = time.Now()
	var currentSimTime time.Time
	var datumTimestamp *timestamp.Timestamp
	var genAlarm bool
	var simulatedTelemetryDataMap = make(map[pb.TelemetryDatumDescription]tel.SimulatedTelemetryData)

	simDurationInMillis = sim.DurationInMinutes * 60000
	switch sim.SampleRate {
	case pb.SampleRate_SR_1_MS:
		sampleRateInMillis = 1
	case pb.SampleRate_SR_10_MS:
		sampleRateInMillis = 10
	case pb.SampleRate_SR_100_MS:
		sampleRateInMillis = 100
	case pb.SampleRate_SR_1000_MS:
		sampleRateInMillis = 1000
	default:
		return simulatedTelemetryDataMap, errors.New("invalid sample rate in millis")
	}

	datumCount = simDurationInMillis / sampleRateInMillis

	// ForceAlarm false, NoAlarms false: alarm generated based on the probabilities declared in alarmEventChoices
	// ForceAlarm true, NoAlarms false: force the generation of an alarm
	// ForceAlarm false, NoAlarms true: do not generate an alarm
	if !sim.ForceAlarm {
		// Determine if there will be an alarm event during the simulation.
		if genAlarmChoice, err = randutil.WeightedChoice(alarmEventChoices); err != nil {
			return simulatedTelemetryDataMap, err
		}
		genAlarm = genAlarmChoice.Item.(bool)

	} else {
		genAlarm = true
	}

	if !sim.ForceAlarm && !sim.NoAlarms {
		// Determine if there will be an alarm event during the simulation.
		if genAlarmChoice, err = randutil.WeightedChoice(alarmEventChoices); err != nil {
			return simulatedTelemetryDataMap, err
		}
		genAlarm = genAlarmChoice.Item.(bool)

	} else if sim.ForceAlarm && !sim.NoAlarms {
		genAlarm = true
	} else if !sim.ForceAlarm && sim.NoAlarms {
		genAlarm = false
	} else {
		return simulatedTelemetryDataMap, errors.New("simulation ForceAlarm & NoAlarm must not both be set to true")
	}

	// Alarm or not, get an alarmTypeChoice to keep the compiler happy.
	if alarmTypeChoice, err = randutil.WeightedChoice(alarmTypeChoices); err != nil {
		return simulatedTelemetryDataMap, err
	}

	ap := alarmTypeChoice.Item.(tel.AlarmParams)

	for datumDesc, datumParams := range telemetryDatumParametersMap {
		//logger.Debug(fmt.Sprintf("generating telemetry data for datum desc: %v", datumDesc.String()))
		values := randFloatsInRange(datumParams.RangeLowValue, datumParams.RangeHighValue, datumCount)
		data := make([]pb.TelemetryDatum, datumCount)
		for i, v := range values {
			data[i].Value = v
		}
		simData := tel.SimulatedTelemetryData{DatumDesc: datumDesc, Data: data, AlarmExists: false,
			AlarmMode: ap.Mode, AlarmIndex: 0}
		if genAlarm {
			if datumDesc.String() == ap.Desc.String() {
				switch ap.Mode {
				case tel.High:
					logger.Debug(fmt.Sprintf("alarm.Desc: %v alarm.Mode: %v range low: %v range high: %v"+
						"ramp dir: up high alarm level: %v", ap.Desc, ap.Mode.String(), datumParams.RangeLowValue,
						datumParams.RangeHighValue, datumParams.HighAlarmValue))
					if err := rampToAlarm(&simData, datumParams.RangeLowValue, datumParams.RangeHighValue, up,
						datumParams.HighAlarmValue); err != nil {
						return simulatedTelemetryDataMap, err
					}
				case tel.Low:
					logger.Debug(fmt.Sprintf("alarm.Desc: %v alarm.Mode: %v range low: %v range high: %v"+
						"ramp dir: down low alarm level: %v", ap.Desc, ap.Mode.String(), datumParams.RangeLowValue,
						datumParams.RangeHighValue, datumParams.LowAlarmValue))
					if err := rampToAlarm(&simData, datumParams.RangeLowValue, datumParams.RangeHighValue, down,
						datumParams.LowAlarmValue); err != nil {
						return simulatedTelemetryDataMap, err
					}
				}
			}
		}
		currentSimTime = simStartTime
		for i := range simData.Data {
			if i > 0 {
				currentSimTime = currentSimTime.Add(time.Second)
			}
			if datumTimestamp, err = ts.TimestampProto(currentSimTime); err != nil {
				return simulatedTelemetryDataMap, err
			}
			simData.Data[i].Uuid = uid.New().String()
			simData.Data[i].Description = datumDesc
			simData.Data[i].Unit = datumParams.Unit
			simData.Data[i].Timestamp = datumTimestamp
			//TODO: Currently, lat, long, & elevation are not modeled in the simulation.
			simData.Data[i].Latitude = 0.0
			simData.Data[i].Longitude = 0.0
			simData.Data[i].Elevation = 0.0
			// The datum Value and (if an alarm occurred) HighAlarm (or LowAlarm) were set in
			// the rampToAlarm() function.
		}
		simulatedTelemetryDataMap[datumDesc] = simData
	}

	return simulatedTelemetryDataMap, nil
}

func telemetryDataGenerationWorker(simUUID string, tdd pb.TelemetryDatumDescription, tdp tel.TelemetryDatumParameters, sampleRateInMillis int32, ap tel.AlarmParams, datumCount int32,
	simStartTime time.Time, genAlarm bool, sem chan int, wg *sync.WaitGroup,
	resultsChan chan tel.SimulatedTelemetryData, errChan chan error) {

	var currentSimTime time.Time
	var datumTimestamp *timestamp.Timestamp
	var err error

	defer wg.Done()
	sem <- 1

	//logger.Debug(fmt.Sprintf("generating telemetry data for datum desc: %v", tdd.String()))
	values := randFloatsInRange(tdp.RangeLowValue, tdp.RangeHighValue, datumCount)
	data := make([]pb.TelemetryDatum, datumCount)
	for i, v := range values {
		data[i].Value = v
	}

	simData := tel.SimulatedTelemetryData{DatumDesc: tdd, Data: data, AlarmExists: false,
		AlarmMode: ap.Mode, AlarmIndex: 0}

	if genAlarm {
		if tdd.String() == ap.Desc.String() {
			switch ap.Mode {
			case tel.High:
				logger.Debug(fmt.Sprintf("alarm.Desc: %v alarm.Mode: %v range low: %v range high: %v"+
					"ramp dir: up high alarm level: %v", ap.Desc, ap.Mode.String(), tdp.RangeLowValue,
					tdp.RangeHighValue, tdp.HighAlarmValue))
				if err := rampToAlarm(&simData, tdp.RangeLowValue, tdp.RangeHighValue, up,
					tdp.HighAlarmValue); err != nil {
					errChan <- err
					<-sem
					return
				}
			case tel.Low:
				logger.Debug(fmt.Sprintf("alarm.Desc: %v alarm.Mode: %v range low: %v range high: %v"+
					"ramp dir: down low alarm level: %v", ap.Desc, ap.Mode.String(), tdp.RangeLowValue,
					tdp.RangeHighValue, tdp.LowAlarmValue))
				if err := rampToAlarm(&simData, tdp.RangeLowValue, tdp.RangeHighValue, down,
					tdp.LowAlarmValue); err != nil {
					errChan <- err
					<-sem
					return
				}
			}
		}
	}

	currentSimTime = simStartTime
	for i := range simData.Data {
		if i > 0 {
			//currentSimTime = currentSimTime.Add(time.Second)
			currentSimTime = currentSimTime.Add(time.Duration(sampleRateInMillis) * time.Millisecond)
		}
		if datumTimestamp, err = ts.TimestampProto(currentSimTime); err != nil {
			errChan <- err
			<-sem
			return
		}
		simData.Data[i].Uuid = uid.New().String()
		simData.Data[i].Simulated = true
		simData.Data[i].SimulationUuid = simUUID
		simData.Data[i].Description = tdd
		simData.Data[i].Unit = tdp.Unit
		simData.Data[i].Timestamp = datumTimestamp
		//TODO: Currently, lat, long, & elevation are not modeled in the simulation.
		simData.Data[i].Latitude = 0.0
		simData.Data[i].Longitude = 0.0
		simData.Data[i].Elevation = 0.0
		// The datum Value and (if an alarm occurred) HighAlarm (or LowAlarm) were set in
		// the rampToAlarm() function.
	}

	//logger.Debug(fmt.Sprintf("appending sim data for datum desc: %v to results channel", simData.datumDesc))
	resultsChan <- simData
	<-sem

}

func randFloatsInRange(min, max float64, n int32) []float64 {
	res := make([]float64, n)
	for i := range res {
		// Round floats down to 2 decimal places.
		res[i] = math.Floor((min+rand.Float64()*(max-min))*100) / 100
	}
	return res
}

func rampToAlarm(simData *tel.SimulatedTelemetryData, minVal float64, maxVal float64, rd rampDirection, alarmLevel float64) error {
	var segmentSize = len(simData.Data) / 4
	var alarmReached = false
	var rampFactor float64
	var alarmIndexSet = false
	if rd == down {
		rampFactor = (((minVal + maxVal) / 2) - alarmLevel) / float64(10)
	} else {
		rampFactor = (alarmLevel - ((minVal + maxVal) / 2)) / float64(10)
	}
	var segmentStartChoices = []randutil.Choice{
		{Weight: 1, Item: 0},
		{Weight: 1, Item: 1},
		{Weight: 1, Item: 2},
	}
	segmentStartChoice, err := randutil.WeightedChoice(segmentStartChoices)
	if err != nil {
		return err
	}
	var segmentStart int = segmentSize * segmentStartChoice.Item.(int)
	var segmentStartOffset = segmentStart + (segmentSize / 2)
	for i := (segmentStartOffset); i < len(simData.Data); i++ {
		if rd == down {
			if (simData.Data)[i-1].Value <= alarmLevel && !alarmReached {
				alarmReached = true
			}
		} else {
			if (simData.Data)[i-1].Value >= alarmLevel && !alarmReached {
				alarmReached = true
			}
		}
		if alarmReached {
			// Pad the remaining datum with value = 0.0 since the simulation is effectively over
			// due to an alarm.
			(simData.Data)[i].Value = 0.0
			if !alarmIndexSet {
				if rd == down {
					(simData.Data)[i-1].LowAlarm = true
				} else {
					(simData.Data)[i-1].HighAlarm = true
				}
				simData.AlarmExists = true
				simData.AlarmIndex = i - 1
				alarmIndexSet = true
				logger.Debug(fmt.Sprintf("alarm level %v reached...", alarmLevel))
			}
		} else {
			if rd == down {
				(simData.Data)[i].Value = math.Floor(((simData.Data)[i-1].Value-rampFactor)*100) / 100
			} else {
				(simData.Data)[i].Value = math.Floor(((simData.Data)[i-1].Value+rampFactor)*100) / 100
			}
		}
	}
	if !alarmReached {
		return fmt.Errorf(fmt.Sprintf("failed to ramp %v to alarm level %v", rd.String(), alarmLevel))
	}
	return nil
}
