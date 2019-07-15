package data

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	pbts "github.com/golang/protobuf/ptypes/timestamp"

	"github.com/bburch01/FOTAAS/api"

	"github.com/bburch01/FOTAAS/internal/app/simulation/models"
	"github.com/bburch01/FOTAAS/internal/app/telemetry"

	//"github.com/bburch01/FOTAAS/internal/app/simulation"
	//"github.com/bburch01/FOTAAS/internal/app/simulation"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/google/uuid"
	"github.com/jmcvetta/randutil"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	//logging "github.com/bburch01/FOTAAS/internal/pkg/logging"
	//"github.com/joho/godotenv"
	//"go.uber.org/zap"
	//tel "github.com/bburch01/FOTAAS/internal/app/telemetry"
	//ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
	//randutil "github.com/jmcvetta/randutil"
	//uid "github.com/google/uuid"
	//uuid "github.com/satori/go.uuid"
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

type SimMemberData struct {
	SimMemberID string
	SimData     map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
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
	desc api.TelemetryDatumDescription
	mode alarmMode
}
*/

var alarmTypeChoices = []randutil.Choice{
	{Weight: 35, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_TIRE_PRESSURE_FL, Mode: telemetry.Low}},
	{Weight: 35, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_TIRE_PRESSURE_FR, Mode: telemetry.Low}},
	{Weight: 35, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_TIRE_PRESSURE_RL, Mode: telemetry.Low}},
	{Weight: 35, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_TIRE_PRESSURE_RR, Mode: telemetry.Low}},
	{Weight: 15, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_ENGINE_OIL_PRESSURE, Mode: telemetry.High}},
	{Weight: 15, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_ENGINE_OIL_PRESSURE, Mode: telemetry.Low}},
	{Weight: 15, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_ENGINE_COOLANT_TEMP, Mode: telemetry.High}},
	{Weight: 10, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_ENGINE_OIL_TEMP, Mode: telemetry.High}},
	{Weight: 10, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_BRAKE_TEMP_FL, Mode: telemetry.High}},
	{Weight: 10, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_BRAKE_TEMP_FR, Mode: telemetry.High}},
	{Weight: 10, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_BRAKE_TEMP_RL, Mode: telemetry.High}},
	{Weight: 10, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_BRAKE_TEMP_RR, Mode: telemetry.High}},
	{Weight: 5, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_MGUK_OUTPUT, Mode: telemetry.High}},
	{Weight: 5, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_MGUK_OUTPUT, Mode: telemetry.Low}},
	{Weight: 5, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_MGUH_OUTPUT, Mode: telemetry.High}},
	{Weight: 5, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_MGUH_OUTPUT, Mode: telemetry.Low}},
	{Weight: 2, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_ENERGY_STORAGE_LEVEL, Mode: telemetry.High}},
	{Weight: 2, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_ENERGY_STORAGE_LEVEL, Mode: telemetry.Low}},
	{Weight: 2, Item: telemetry.AlarmParams{Desc: api.TelemetryDatumDescription_ENERGY_STORAGE_TEMP, Mode: telemetry.High}},
}

var telemetryDatumParametersMap = map[api.TelemetryDatumDescription]telemetry.TelemetryDatumParameters{
	api.TelemetryDatumDescription_BRAKE_TEMP_FL: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 750.0, RangeHighValue: 1050.0,
		HighAlarmValue: 1300.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_BRAKE_TEMP_FR: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 750.0, RangeHighValue: 1050.0,
		HighAlarmValue: 1300.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_BRAKE_TEMP_RL: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 750.0, RangeHighValue: 1050.0,
		HighAlarmValue: 1300.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_BRAKE_TEMP_RR: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 750.0, RangeHighValue: 1050.0,
		HighAlarmValue: 1300.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_ENERGY_STORAGE_LEVEL: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_MJ, RangeLowValue: 1.3, RangeHighValue: 3.8,
		HighAlarmValue: 4.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_ENERGY_STORAGE_TEMP: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 50.0, RangeHighValue: 55.0,
		HighAlarmValue: 60.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_ENGINE_COOLANT_TEMP: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 110.0, RangeHighValue: 120.0,
		HighAlarmValue: 140.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_ENGINE_OIL_PRESSURE: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_KPA, RangeLowValue: 300.0, RangeHighValue: 400.0,
		HighAlarmValue: 550.0, LowAlarmValue: 40.0,
	},
	api.TelemetryDatumDescription_ENGINE_OIL_TEMP: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 110.0, RangeHighValue: 120.0,
		HighAlarmValue: 140.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_ENGINE_RPM: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_RPM, RangeLowValue: 2500.0, RangeHighValue: 13500.00,
		HighAlarmValue: 15000.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_FUEL_CONSUMED: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_KG, RangeLowValue: 0, RangeHighValue: 120.0,
		HighAlarmValue: 125.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_FUEL_FLOW: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_KG_PER_HOUR, RangeLowValue: 10.0, RangeHighValue: 80.0,
		HighAlarmValue: 100.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_G_FORCE: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_G, RangeLowValue: 2.0, RangeHighValue: 6.0,
		HighAlarmValue: 8.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_G_FORCE_DIRECTION: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_RADIAN, RangeLowValue: 0, RangeHighValue: 6.280,
		HighAlarmValue: 0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_MGUH_OUTPUT: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_JPS, RangeLowValue: 16.0, RangeHighValue: 19.0,
		HighAlarmValue: 25.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_MGUK_OUTPUT: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_JPS, RangeLowValue: 16.0, RangeHighValue: 19.0,
		HighAlarmValue: 25.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_SPEED: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_KPH, RangeLowValue: 100.0, RangeHighValue: 350.0,
		HighAlarmValue: 400.0, LowAlarmValue: 0,
	},
	api.TelemetryDatumDescription_TIRE_PRESSURE_FL: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_BAR, RangeLowValue: 1.1, RangeHighValue: 1.3,
		HighAlarmValue: 1.6, LowAlarmValue: 0.8,
	},
	api.TelemetryDatumDescription_TIRE_PRESSURE_FR: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_BAR, RangeLowValue: 1.1, RangeHighValue: 1.3,
		HighAlarmValue: 1.6, LowAlarmValue: 0.8,
	},
	api.TelemetryDatumDescription_TIRE_PRESSURE_RL: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_BAR, RangeLowValue: 1.1, RangeHighValue: 1.3,
		HighAlarmValue: 1.6, LowAlarmValue: 0.8,
	},
	api.TelemetryDatumDescription_TIRE_PRESSURE_RR: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_BAR, RangeLowValue: 1.1, RangeHighValue: 1.3,
		HighAlarmValue: 1.6, LowAlarmValue: 0.8,
	},
	api.TelemetryDatumDescription_TIRE_TEMP_FL: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 80.0, RangeHighValue: 120.0,
		HighAlarmValue: 130.0, LowAlarmValue: 70.0,
	},
	api.TelemetryDatumDescription_TIRE_TEMP_FR: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 80.0, RangeHighValue: 120.0,
		HighAlarmValue: 130.0, LowAlarmValue: 70.0,
	},
	api.TelemetryDatumDescription_TIRE_TEMP_RL: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 80.0, RangeHighValue: 120.0,
		HighAlarmValue: 130.0, LowAlarmValue: 70.0,
	},
	api.TelemetryDatumDescription_TIRE_TEMP_RR: telemetry.TelemetryDatumParameters{
		Unit: api.TelemetryDatumUnit_DEGREE_CELCIUS, RangeLowValue: 80.0, RangeHighValue: 120.0,
		HighAlarmValue: 130.0, LowAlarmValue: 70.0,
	},
}

func GenerateSimulatedTelemetryData(sim models.Simulation, simMember models.SimulationMember, wg *sync.WaitGroup,
	resultsChan chan SimMemberData, errChan chan error) {

	var datumCount int32
	var sampleRateInMillis int32
	var simDurationInMillis int32
	var genAlarmChoice randutil.Choice
	var alarmTypeChoice randutil.Choice
	var err error
	var simStartTime = time.Now()
	var sb strings.Builder

	var genAlarm bool
	var simulatedTelemetryDataMap = make(map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData)

	defer wg.Done()

	simDurationInMillis = sim.DurationInMinutes * 60000

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
		// This should never happen. Validation occurs in both the protobuf api
		// and in main.go RunSimulation()
		sb.WriteString(fmt.Sprintf("invalid sample rate for simulation member: %v", simMember.ID))
		logger.Error(sb.String())
		errChan <- fmt.Errorf(sb.String())
	}

	datumCount = simDurationInMillis / sampleRateInMillis

	// ForceAlarm false, NoAlarms false: alarm generated based on the probabilities declared in alarmEventChoices
	// ForceAlarm true, NoAlarms false: force the generation of an alarm
	// ForceAlarm false, NoAlarms true: do not generate an alarm
	if !simMember.ForceAlarm {
		if genAlarmChoice, err = randutil.WeightedChoice(alarmEventChoices); err != nil {
			errChan <- err
		}
		genAlarm = genAlarmChoice.Item.(bool)
	} else {
		genAlarm = true
	}

	if !simMember.ForceAlarm && !simMember.NoAlarms {
		if genAlarmChoice, err = randutil.WeightedChoice(alarmEventChoices); err != nil {
			errChan <- err
		}
		genAlarm = genAlarmChoice.Item.(bool)

	} else if simMember.ForceAlarm && !simMember.NoAlarms {
		genAlarm = true
	} else if !simMember.ForceAlarm && simMember.NoAlarms {
		genAlarm = false

	} else {
		// This should never happen. Validation occurs in main.go RunSimulation()
		sb.WriteString(fmt.Sprintf("invalid ForceAlarm & NoAlarms combination for simulation member: %v", simMember.ID))
		logger.Error(sb.String())
		errChan <- fmt.Errorf(sb.String())
	}

	// Alarm or not, get an alarmTypeChoice to keep the compiler happy.
	if alarmTypeChoice, err = randutil.WeightedChoice(alarmTypeChoices); err != nil {
		errChan <- err
	}

	workerErrChan := make(chan error, len(telemetryDatumParametersMap))
	workerResultsChan := make(chan telemetry.SimulatedTelemetryData, len(telemetryDatumParametersMap))
	sem := make(chan int, runtime.NumCPU())
	var workerWg sync.WaitGroup
	workerWg.Add(len(telemetryDatumParametersMap))

	logger.Debug(fmt.Sprintf("starting data generation workers for simulation member: %v", simMember.ID))

	for datumDesc, datumParams := range telemetryDatumParametersMap {
		go telemetryDataGenerationWorker(sim, simMember, datumDesc, datumParams, sampleRateInMillis,
			alarmTypeChoice.Item.(telemetry.AlarmParams), datumCount,
			simStartTime, genAlarm, sem, &workerWg, workerResultsChan, workerErrChan)
	}

	workerWg.Wait()
	close(workerErrChan)
	close(workerResultsChan)

	if err := <-workerErrChan; err != nil {
		errChan <- err
	}

	for std := range workerResultsChan {
		simulatedTelemetryDataMap[std.DatumDesc] = std
	}

	smd := SimMemberData{SimMemberID: simMember.ID, SimData: simulatedTelemetryDataMap}
	resultsChan <- smd

	return
}

func telemetryDataGenerationWorker(sim models.Simulation, simMember models.SimulationMember, tdd api.TelemetryDatumDescription, tdp telemetry.TelemetryDatumParameters, sampleRateInMillis int32, ap telemetry.AlarmParams, datumCount int32,
	simStartTime time.Time, genAlarm bool, sem chan int, wg *sync.WaitGroup,
	resultsChan chan telemetry.SimulatedTelemetryData, errChan chan error) {

	var currentSimTime time.Time
	var datumTimestamp *pbts.Timestamp
	var err error

	defer wg.Done()
	sem <- 1

	values := randFloatsInRange(tdp.RangeLowValue, tdp.RangeHighValue, datumCount)
	data := make([]api.TelemetryDatum, datumCount)
	for i, v := range values {
		data[i].Value = v
	}

	simData := telemetry.SimulatedTelemetryData{DatumDesc: tdd, Data: data, AlarmExists: false,
		AlarmMode: ap.Mode, AlarmIndex: 0}

	if genAlarm {
		if tdd.String() == ap.Desc.String() {
			switch ap.Mode {
			case telemetry.High:
				logger.Debug(fmt.Sprintf("alarm.Desc: %v alarm.Mode: %v range low: %v range high: %v"+
					"ramp dir: up high alarm level: %v", ap.Desc, ap.Mode.String(), tdp.RangeLowValue,
					tdp.RangeHighValue, tdp.HighAlarmValue))
				if err := rampToAlarm(&simData, tdp.RangeLowValue, tdp.RangeHighValue, up,
					tdp.HighAlarmValue); err != nil {
					errChan <- err
					<-sem
					return
				}
			case telemetry.Low:
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
		if datumTimestamp, err = ipbts.TimestampProto(currentSimTime); err != nil {
			errChan <- err
			<-sem
			return
		}
		simData.Data[i].Uuid = uuid.New().String()
		simData.Data[i].Simulated = true
		simData.Data[i].SimulationUuid = sim.ID
		simData.Data[i].GranPrix = sim.GranPrix
		simData.Data[i].Track = sim.Track
		simData.Data[i].Constructor = simMember.Constructor
		simData.Data[i].CarNumber = simMember.CarNumber
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

	resultsChan <- simData
	<-sem

	return

}

func randFloatsInRange(min, max float64, n int32) []float64 {
	res := make([]float64, n)
	for i := range res {
		// Round floats down to 2 decimal places.
		res[i] = math.Floor((min+rand.Float64()*(max-min))*100) / 100
	}
	return res
}

func rampToAlarm(simData *telemetry.SimulatedTelemetryData, minVal float64, maxVal float64, rd rampDirection, alarmLevel float64) error {
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
