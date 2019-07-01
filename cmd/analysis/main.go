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
	"github.com/bburch01/FOTAAS/internal/app/analysis/models"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/joho/godotenv"
	"github.com/openzipkin/zipkin-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var logger *zap.Logger

type server struct{}

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

	if err = models.InitDB(); err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize database driver with error: %v", err))
	}

}

func (s *server) HealthCheck(ctx context.Context, in *api.HealthCheckRequest) (*api.HealthCheckResponse, error) {

	// Assume good health until a health check test fails.
	var hcr = api.HealthCheckResponse{ServerStatus: &api.ServerStatus{Code: api.StatusCode_OK, Message: "analysis service healthy"}}

	err := models.PingDB()
	if err != nil {
		hcr.ServerStatus.Code = api.StatusCode_ERROR
		hcr.ServerStatus.Message = fmt.Sprintf("failed to ping database with error: %v", err.Error())
		return &hcr, nil
	}

	return &hcr, nil

}

func main() {

	var sb strings.Builder

	reporter := zhttp.NewReporter(os.Getenv("ZIPKIN_ENDPOINT_URL"))
	defer reporter.Close()

	sb.WriteString(os.Getenv("ANALYSIS_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("ANALYSIS_SERVICE_PORT"))
	analysisSvcEndpoint := sb.String()

	zipkinLocalEndpoint, err := zipkin.NewEndpoint("analysis-service", analysisSvcEndpoint)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create zipkin local endpoint with error: %v", err))
	}

	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zipkinLocalEndpoint))
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create zipkin tracer with error: %v", err))
	}

	sb.Reset()
	sb.WriteString(":")
	sb.WriteString(os.Getenv("ANALYSIS_SERVICE_PORT"))
	analysisSvcPort := sb.String()

	listener, err := net.Listen("tcp", analysisSvcPort)
	if err != nil {
		logger.Fatal(fmt.Sprintf("tcp failed to listen on analysis service port %v with error: %v", analysisSvcPort, err))
	}

	svr := grpc.NewServer(grpc.StatsHandler(zgrpc.NewServerHandler(tracer)))

	api.RegisterAnalysisServiceServer(svr, &server{})

	if err := svr.Serve(listener); err != nil {
		logger.Fatal(fmt.Sprintf("failed to serve on analysis service port %v with error: %v", analysisSvcPort, err))
	}
}
