package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	zgrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	zhttp "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/app/status"
	"github.com/bburch01/FOTAAS/internal/app/telemetry/models"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/openzipkin/zipkin-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var logger *zap.Logger

type server struct{}

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

	if err = models.InitDB(); err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize database driver with error: %v", err))
	}

}

func main() {

	var sb strings.Builder

	reporter := zhttp.NewReporter(os.Getenv("ZIPKIN_ENDPOINT_URL"))
	defer reporter.Close()

	sb.WriteString(os.Getenv("STATUS_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("STATUS_SERVICE_PORT"))
	statusSvcEndpoint := sb.String()

	zipkinLocalEndpoint, err := zipkin.NewEndpoint("status-service", statusSvcEndpoint)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create zipkin local endpoint with error: %v", err))
	}

	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zipkinLocalEndpoint))
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create zipkin tracer with error: %v", err))
	}

	sb.Reset()
	sb.WriteString(":")
	sb.WriteString(os.Getenv("STATUS_SERVICE_PORT"))
	statusSvcPort := sb.String()

	listener, err := net.Listen("tcp", statusSvcPort)
	if err != nil {
		logger.Fatal(fmt.Sprintf("tcp failed to listen on status service port %v with error: %v", statusSvcPort, err))
	}

	svr := grpc.NewServer(grpc.StatsHandler(zgrpc.NewServerHandler(tracer)))

	api.RegisterSystemStatusServiceServer(svr, &server{})

	if err := svr.Serve(listener); err != nil {
		logger.Fatal(fmt.Sprintf("failed to serve on status service port %v with error: %v", statusSvcPort, err))
	}

}

func (s *server) AlivenessCheck(ctx context.Context, req *api.AlivenessCheckRequest) (*api.AlivenessCheckResponse, error) {

	resp := new(api.AlivenessCheckResponse)
	resp.Details = &api.ResponseDetails{Code: api.ResponseCode_OK,
		Message: "status service alive"}

	if err := models.PingDB(); err != nil {
		resp.Details.Code = api.ResponseCode_ERROR
		resp.Details.Message = fmt.Sprintf("failed to ping database with error: %v", err.Error())
		logger.Error(fmt.Sprintf("failed to ping database with error: %v", err))
		// protoc generated code requires error in the return params, return nil here so that clients
		// of this service can process this FOTAAS error differently than other system errors (e.g.
		// if this service is not available). Intercept this error and handle it via response code &
		// message.
		return resp, nil
	}

	return resp, nil
}

func (s *server) GetSystemStatus(ctx context.Context, req *api.GetSystemStatusRequest) (*api.GetSystemStatusResponse, error) {

	var statusReport api.SystemStatusReport

	statusReport.TelemetryServiceAliveness = status.CheckServiceAliveness("telemetry")
	statusReport.AnalysisServiceAliveness = status.CheckServiceAliveness("analysis")
	statusReport.SimulationServiceAliveness = status.CheckServiceAliveness("simulation")
	statusReport.StartSimulation = api.TestResult_INCOMPLETE
	statusReport.PollForSimulationComplete = api.TestResult_INCOMPLETE
	statusReport.RetrieveSimulationData = api.TestResult_INCOMPLETE
	statusReport.SimulationDataAnalysis = api.TestResult_INCOMPLETE

	if statusReport.TelemetryServiceAliveness == api.TestResult_FAIL ||
		statusReport.AnalysisServiceAliveness == api.TestResult_FAIL ||
		statusReport.SimulationServiceAliveness == api.TestResult_FAIL {

		resp := api.GetSystemStatusResponse{Details: &api.ResponseDetails{
			Code: api.ResponseCode_INFO, Message: "one or more prerequisite system status tests have failed, test sequence aborted"},
			SystemStatusReport: &statusReport}

		return &resp, nil

	}

	simDurationInMinutes := int32(1)
	simID := uuid.New().String()

	statusReport.StartSimulation = status.StartSimulation(simID, simDurationInMinutes)

	if statusReport.StartSimulation == api.TestResult_FAIL {
		resp := api.GetSystemStatusResponse{Details: &api.ResponseDetails{
			Code: api.ResponseCode_INFO, Message: "one or more prerequisite system status tests have failed, test sequence aborted"},
			SystemStatusReport: &statusReport}

		return &resp, nil
	}

	statusReport.PollForSimulationComplete = status.PollForSimulationComplete(simID, simDurationInMinutes)

	if statusReport.PollForSimulationComplete == api.TestResult_FAIL {
		resp := api.GetSystemStatusResponse{Details: &api.ResponseDetails{
			Code: api.ResponseCode_INFO, Message: "one or more prerequisite system status tests have failed, test sequence aborted"},
			SystemStatusReport: &statusReport}

		return &resp, nil
	}

	statusReport.RetrieveSimulationData = status.RetrieveSimulationData(simID)

	if statusReport.RetrieveSimulationData == api.TestResult_FAIL {
		resp := api.GetSystemStatusResponse{Details: &api.ResponseDetails{
			Code: api.ResponseCode_INFO, Message: "one or more prerequisite system status tests have failed, test sequence aborted"},
			SystemStatusReport: &statusReport}

		return &resp, nil
	}

	statusReport.SimulationDataAnalysis = status.SimulationDataAnalysis(simID)

	if statusReport.SimulationDataAnalysis == api.TestResult_FAIL {
		resp := api.GetSystemStatusResponse{Details: &api.ResponseDetails{
			Code: api.ResponseCode_WARN, Message: "test sequence complete, one or more system status tests failed"},
			SystemStatusReport: &statusReport}

		return &resp, nil
	}

	resp := api.GetSystemStatusResponse{Details: &api.ResponseDetails{
		Code: api.ResponseCode_OK, Message: "system status test sequence complete, all system status tests passed"},
		SystemStatusReport: &statusReport}

	return &resp, nil
}
