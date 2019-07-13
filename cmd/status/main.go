package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	zgrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	zhttp "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/app/telemetry/models"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
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

func (s *server) GetSystemStatus(ctx context.Context, req *api.GetSystemStatusRequest) (*api.GetSystemStatusResponse, error) {

	//var resp api.GetSystemStatusResponse
	var statusReport api.SystemStatusReport

	statusReport.TelemetryServiceAliveness = runServiceAlivenessTest("telemetry")
	statusReport.AnalysisServiceAliveness = runServiceAlivenessTest("analysis")
	statusReport.SimulationServiceAliveness = runServiceAlivenessTest("simulation")

	resp := api.GetSystemStatusResponse{Details: &api.ResponseDetails{
		Code: api.ResponseCode_OK, Message: "successfully completed system status checks"},
		SystemStatusReport: &statusReport}

	return &resp, nil
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

func runServiceAlivenessTest(svcname string) api.TestResult {

	var svcEndpoint string
	var resp *api.HealthCheckResponse
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

	//For now, use context.WithDeadline instead of context.WithTimeout
	//ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)

	// TODO: determine what is the appropriate deadline for health check requests
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	switch svcname {
	case "telemetry":
		client := api.NewTelemetryServiceClient(conn)
		resp, err = client.HealthCheck(ctx, &api.HealthCheckRequest{})
		if err != nil {
			logger.Error(fmt.Sprintf("%v service aliveness check failed with error: %v", svcname, err))
			return api.TestResult_FAIL
		}
	case "analysis":
		client := api.NewAnalysisServiceClient(conn)
		resp, err = client.HealthCheck(ctx, &api.HealthCheckRequest{})
		if err != nil {
			logger.Error(fmt.Sprintf("%v service aliveness check failed with error: %v", svcname, err))
			return api.TestResult_FAIL
		}
	case "simulation":
		client := api.NewSimulationServiceClient(conn)
		resp, err = client.HealthCheck(ctx, &api.HealthCheckRequest{})
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
