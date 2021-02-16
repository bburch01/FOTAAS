package status

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var logger *zap.Logger

func init() {

	var logMode logging.LogMode
	var err error

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err = godotenv.Load(); err != nil {
		log.Panicf("failed to load environment variables with error: %v", err)
	}

	if logMode, err = logging.LogModeForString(os.Getenv("LOG_MODE")); err != nil {
		log.Panicf("failed to initialize logging subsystem with error: %v", err)
	}

	if logger, err = logging.NewLogger(logMode, os.Getenv("LOG_DIR"), os.Getenv("LOG_FILE_NAME")); err != nil {
		log.Panicf("failed to initialize logging subsystem with error: %v", err)
	}

}

func CheckServiceAliveness(svcname string) api.TestResult {

	var svcEndpoint string
	var resp *api.AlivenessCheckResponse
	var sb strings.Builder

	switch svcname {
	case "telemetry":
		sb.WriteString(os.Getenv("TELEMETRY_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("TELEMETRY_SERVICE_PORT"))
		svcEndpoint = sb.String()
	case "analysis":
		sb.WriteString(os.Getenv("ANALYSIS_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("ANALYSIS_SERVICE_PORT"))
		svcEndpoint = sb.String()
	case "simulation":
		sb.WriteString(os.Getenv("SIMULATION_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("SIMULATION_SERVICE_PORT"))
		svcEndpoint = sb.String()
	default:
		logger.Error(fmt.Sprintf("service aliveness check failed, invalid service name: %v", svcname))
		return api.TestResult_FAIL
	}

	conn, err := grpc.Dial(svcEndpoint, grpc.WithInsecure())
	if err != nil {
		logger.Error(fmt.Sprintf("%v service aliveness check failed with error: %v", svcname, err))
		return api.TestResult_FAIL
	}
	defer conn.Close()

	// TODO: determine what the appropriate deadline should be for this service call.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	switch svcname {
	case "telemetry":
		client := api.NewTelemetryServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			logger.Error(fmt.Sprintf("%v service aliveness check failed with error: %v", svcname, err))
			return api.TestResult_FAIL
		}
	case "analysis":
		client := api.NewAnalysisServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			logger.Error(fmt.Sprintf("%v service aliveness check failed with error: %v", svcname, err))
			return api.TestResult_FAIL
		}
	case "simulation":
		client := api.NewSimulationServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			logger.Error(fmt.Sprintf("%v service aliveness check failed with error: %v", svcname, err))
			return api.TestResult_FAIL
		}
	default:
		logger.Error(fmt.Sprintf("service aliveness check failed, invalid service name: %v", svcname))
		return api.TestResult_FAIL
	}

	switch resp.Details.Code {
	case api.ResponseCode_OK:
		return api.TestResult_PASS
	case api.ResponseCode_ERROR:
		logger.Error(fmt.Sprintf("%v service aliveness test failed with message: %v", svcname, resp.Details.Message))
		return api.TestResult_FAIL
	default:
		logger.Error(fmt.Sprintf("service aliveness check failed, invalid service status code: %v", resp.Details.Code.String()))
		return api.TestResult_FAIL
	}
}

func StartSimulation(simID string, simDurationInMinutes int32) api.TestResult {

	var req api.RunSimulationRequest
	var forceAlarmFlag, noAlarmFlag bool
	var sb strings.Builder
	var simulationSvcEndpoint string

	forceAlarmFlag = true
	noAlarmFlag = false

	simMemberMap := make(map[string]*api.SimulationMember)

	simMemberID := uuid.New().String()
	simMember1 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_HAAS,
		CarNumber: 8, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
	}
	simMemberMap[simMemberID] = &simMember1

	/*
		simMemberID = uuid.New().String()
		simMember2 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_HAAS,
			CarNumber: 20, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
		}
		simMemberMap[simMemberID] = &simMember2
	*/

	simMemberID = uuid.New().String()
	simMember3 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_MERCEDES,
		CarNumber: 44, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
	}
	simMemberMap[simMemberID] = &simMember3

	/*
		simMemberID = uuid.New().String()
		simMember4 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_MERCEDES,
			CarNumber: 77, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
		}
		simMemberMap[simMemberID] = &simMember4
	*/

	/*
		simMemberID = uuid.New().String()
		simMember5 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_FERRARI,
			CarNumber: 5, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
		}
		simMemberMap[simMemberID] = &simMember5

		simMemberID = uuid.New().String()
		simMember6 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_FERRARI,
			CarNumber: 16, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
		}
		simMemberMap[simMemberID] = &simMember6

		simMemberID = uuid.New().String()
		simMember7 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_RED_BULL_RACING,
			CarNumber: 33, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
		}
		simMemberMap[simMemberID] = &simMember7

		simMemberID = uuid.New().String()
		simMember8 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_RED_BULL_RACING,
			CarNumber: 10, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
		}
		simMemberMap[simMemberID] = &simMember8

		simMemberID = uuid.New().String()
		simMember9 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_MCLAREN,
			CarNumber: 55, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
		}
		simMemberMap[simMemberID] = &simMember9

		simMemberID = uuid.New().String()
		simMember10 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_MCLAREN,
			CarNumber: 4, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
		}
		simMemberMap[simMemberID] = &simMember10

		simMemberID = uuid.New().String()
		simMember11 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_WILLIAMS,
			CarNumber: 88, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
		}
		simMemberMap[simMemberID] = &simMember11

		simMemberID = uuid.New().String()
		simMember12 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_WILLIAMS,
			CarNumber: 63, ForceAlarm: forceAlarmFlag, NoAlarms: noAlarmFlag,
		}
		simMemberMap[simMemberID] = &simMember12
	*/

	sim := api.Simulation{Uuid: simID, DurationInMinutes: simDurationInMinutes, SampleRate: api.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: api.SimulationRateMultiplier_X1, GranPrix: api.GranPrix_UNITED_STATES,
		Track: api.Track_AUSTIN, SimulationMemberMap: simMemberMap}

	req.Simulation = &sim

	sb.WriteString(os.Getenv("SIMULATION_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("SIMULATION_SERVICE_PORT"))
	simulationSvcEndpoint = sb.String()

	conn, err := grpc.Dial(simulationSvcEndpoint, grpc.WithInsecure())
	if err != nil {
		logger.Error(fmt.Sprintf("start simulation test failed with error: %v", err))
		return api.TestResult_FAIL
	}
	defer conn.Close()

	// TODO: determine what the appropriate deadline should be for this service call.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client = api.NewSimulationServiceClient(conn)

	resp, err := client.RunSimulation(ctx, &req)
	if err != nil {
		logger.Error(fmt.Sprintf("start simulation test failed with error: %v", err))
		return api.TestResult_FAIL
	}

	switch resp.Details.Code {
	case api.ResponseCode_OK:
		return api.TestResult_PASS
	case api.ResponseCode_ERROR:
		logger.Error(fmt.Sprintf("start simulation test failed with simulation service message: %v", resp.Details.Message))
		return api.TestResult_FAIL
	default:
		logger.Error(fmt.Sprintf("start simulation test failed, invalid service status code: %v", resp.Details.Code.String()))
		return api.TestResult_FAIL
	}

}

func PollForSimulationComplete(simID string, simDurationInMinutes int32) api.TestResult {

	var simulationSvcEndpoint string
	var sb strings.Builder
	var req api.GetSimulationInfoRequest
	var resp *api.GetSimulationInfoResponse

	req.SimulationUuid = simID

	sb.WriteString(os.Getenv("SIMULATION_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("SIMULATION_SERVICE_PORT"))
	simulationSvcEndpoint = sb.String()

	conn, err := grpc.Dial(simulationSvcEndpoint, grpc.WithInsecure())
	if err != nil {
		logger.Error(fmt.Sprintf("start simulation test failed with error: %v", err))
		return api.TestResult_FAIL
	}
	defer conn.Close()

	// TODO: determine what the appropriate deadline should be for this service call.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)

	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client = api.NewSimulationServiceClient(conn)

	pollCount := int32(0)
	// TODO: this poll count is going to have to change depending on deployment.
	// Currently, the GKE cluster's minimal CPU allotment takes 3x as long to
	// complete the simulation as the iMac does. Probably going to need to
	// make this pollCount configurable via an env var.
	for pollCount < (600) {

		resp, err = client.GetSimulationInfo(ctx, &req)
		if err != nil {
			logger.Error(fmt.Sprintf("start simulation test failed with error: %v", err))
			return api.TestResult_FAIL
		}

		switch resp.Details.Code {
		case api.ResponseCode_OK:
			if resp.SimulationInfo.State == api.SimulationState_COMPLETED {
				return api.TestResult_PASS
			}
		case api.ResponseCode_ERROR:
			logger.Error(fmt.Sprintf("poll for simulation complete test failed with simulation service message: %v", resp.Details.Message))
			return api.TestResult_FAIL
		case api.ResponseCode_INFO, api.ResponseCode_WARN:
			sb.Reset()
			sb.WriteString("poll for simulation complete test failed, ")
			sb.WriteString(fmt.Sprintf("simulation service response code: %v ", resp.Details.Code.String()))
			sb.WriteString(fmt.Sprintf("simulation service response message: %v ", resp.Details.Message))
			logger.Error(sb.String())
			return api.TestResult_FAIL
		default:
			logger.Error(fmt.Sprintf("poll for simulation complete test failed, invalid service status code: %v", resp.Details.Code.String()))
			return api.TestResult_FAIL
		}

		time.Sleep(time.Duration(time.Second))
		pollCount++

	}

	return api.TestResult_FAIL

}

func RetrieveSimulationData(simID string) api.TestResult {

	dataReq := new(api.GetTelemetryDataRequest)
	dataReq.SearchBy = new(api.GetTelemetryDataRequest_SearchBy)
	dataReq.Simulated = true
	dataReq.SimulationUuid = simID
	dataReq.SearchBy.DateRange = true
	dataReq.SearchBy.Constructor = true
	dataReq.SearchBy.CarNumber = true

	var sb strings.Builder
	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_PORT"))
	telemetrySvcEndpoint := sb.String()

	conn, err := grpc.Dial(telemetrySvcEndpoint, grpc.WithInsecure())
	if err != nil {
		logger.Error(fmt.Sprintf("retrieve simulation data test failed with error: %v", err))
		return api.TestResult_FAIL
	}
	defer conn.Close()

	// TODO: determine what the appropriate deadline should be for this service call.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client = api.NewTelemetryServiceClient(conn)

	var startTime, endTime time.Time

	year, month, day := time.Now().Date()

	sb.Reset()
	sb.WriteString(strconv.Itoa(year))
	sb.WriteString("-")
	if int(month) < 10 {
		sb.WriteString("0")
	}
	sb.WriteString(strconv.Itoa(int(month)))
	sb.WriteString("-")
	if day < 10 {
		sb.WriteString("0")
	}
	sb.WriteString(strconv.Itoa(day))

	startTime, err = time.Parse(time.RFC3339, sb.String()+"T00:00:00Z")
	if err != nil {
		logger.Error(fmt.Sprintf("retrieve simulation data test failed with error: %v", err))
		return api.TestResult_FAIL
	}

	endTime, err = time.Parse(time.RFC3339, sb.String()+"T23:59:59Z")
	if err != nil {
		logger.Error(fmt.Sprintf("retrieve simulation data test failed with error: %v", err))
		return api.TestResult_FAIL
	}

	dataReq.DateRangeBegin, err = ipbts.TimestampProto(startTime)
	if err != nil {
		logger.Error(fmt.Sprintf("retrieve simulation data test failed with error: %v", err))
		return api.TestResult_FAIL
	}
	dataReq.DateRangeEnd, err = ipbts.TimestampProto(endTime)
	if err != nil {
		logger.Error(fmt.Sprintf("retrieve simulation data test failed with error: %v", err))
		return api.TestResult_FAIL
	}

	dataReq.Constructor = api.Constructor_HAAS
	dataReq.CarNumber = 8

	resp, err := client.GetTelemetryData(ctx, dataReq)
	if err != nil {
		logger.Error(fmt.Sprintf("retrieve simulation data test failed with error: %v", err))
		return api.TestResult_FAIL
	}

	if resp.Details.Code != api.ResponseCode_OK {
		logger.Error(fmt.Sprintf("retrieve simulation data test failed with telemetry service message: %v", resp.Details.Message))
		return api.TestResult_FAIL
	}

	// 1500 looks like a magic number here. The simulation that is run as part of the system status tests will always
	// produce 1500 telemetry datums per constructor/car number.
	if resp.TelemetryData == nil || len(resp.TelemetryData.TelemetryDatumMap) != 1500 {
		logger.Error(fmt.Sprintf("retrieve simulation data test failed, invalid telemetry datum count for constructor %v car number %v",
			dataReq.Constructor.String(), dataReq.CarNumber))
		return api.TestResult_FAIL
	}

	return api.TestResult_PASS
}

func SimulationDataAnalysis(simID string) api.TestResult {

	var analysisSvcEndpoint string
	var startTime, endTime time.Time
	var sb strings.Builder
	var resp *api.GetAlarmAnalysisResponse
	var err error

	req := new(api.GetAlarmAnalysisRequest)

	year, month, day := time.Now().Date()

	sb.WriteString(strconv.Itoa(year))
	sb.WriteString("-")
	if int(month) < 10 {
		sb.WriteString("0")
	}
	sb.WriteString(strconv.Itoa(int(month)))
	sb.WriteString("-")
	if day < 10 {
		sb.WriteString("0")
	}
	sb.WriteString(strconv.Itoa(day))

	startTime, err = time.Parse(time.RFC3339, sb.String()+"T00:00:00Z")
	if err != nil {
		logger.Error(fmt.Sprintf("retrieve simulation data test failed with error: %v", err))
		return api.TestResult_FAIL
	}

	endTime, err = time.Parse(time.RFC3339, sb.String()+"T23:59:59Z")
	if err != nil {
		logger.Error(fmt.Sprintf("retrieve simulation data test failed with error: %v", err))
		return api.TestResult_FAIL
	}

	req.DateRangeBegin, err = ipbts.TimestampProto(startTime)
	if err != nil {
		logger.Error(fmt.Sprintf("simulation data analysis test failed with error: %v", err))
		return api.TestResult_FAIL
	}

	req.DateRangeEnd, err = ipbts.TimestampProto(endTime)
	if err != nil {
		logger.Error(fmt.Sprintf("simulation data analysis test failed with error: %v", err))
		return api.TestResult_FAIL
	}

	req.Simulated = true

	req.SimulationUuid = simID

	sb.Reset()
	sb.WriteString(os.Getenv("ANALYSIS_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("ANALYSIS_SERVICE_PORT"))
	analysisSvcEndpoint = sb.String()

	conn, err := grpc.Dial(analysisSvcEndpoint, grpc.WithInsecure())
	if err != nil {
		logger.Error(fmt.Sprintf("simulation data analysis test failed with error: %v", err))
		return api.TestResult_FAIL
	}
	defer conn.Close()

	// TODO: determine what the appropriate deadline should be for this service call.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client = api.NewAnalysisServiceClient(conn)

	resp, err = client.GetAlarmAnalysis(ctx, req)
	if err != nil {
		logger.Error(fmt.Sprintf("simulation data analysis test failed with error: %v", err))
		return api.TestResult_FAIL
	}

	if resp.Details.Code != api.ResponseCode_OK {
		logger.Error(fmt.Sprintf("simulation data analysis test failed with analysis service message: %v", resp.Details.Message))
		return api.TestResult_FAIL
	}

	// TODO: add more tests on the contents of the GetAlarmAnalysisResponse. Also,
	// add tests for the GetConstructorAlarmAnalysis service call.
	// Need to be able to set the number of cars include in the simulation in StartSimulation.
	if len(resp.AlarmAnalysisData.AlarmCounts) != 2 {
		logger.Error(fmt.Sprintf("simulation data analysis test failed with invalid alarm count: %v",
			len(resp.AlarmAnalysisData.AlarmCounts)))
		return api.TestResult_FAIL
	}

	return api.TestResult_PASS

}
