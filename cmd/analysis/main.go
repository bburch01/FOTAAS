package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	pb "github.com/bburch01/FOTAAS/api"
	mdl "github.com/bburch01/FOTAAS/internal/app/analysis/models"
	logging "github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/joho/godotenv"
	"github.com/openzipkin/zipkin-go"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	reporterhttp "github.com/openzipkin/zipkin-go/reporter/http"
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

	if err = mdl.InitDB(); err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize database driver with error: %v", err))
	}

}

func (s *server) HealthCheck(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {

	// Assume good health until a health check test fails.
	var hcr = pb.HealthCheckResponse{ServerStatus: &pb.ServerStatus{Code: pb.StatusCode_OK, Message: "analysis service healthy"}}

	err := mdl.PingDB()
	if err != nil {
		hcr.ServerStatus.Code = pb.StatusCode_ERROR
		hcr.ServerStatus.Message = fmt.Sprintf("failed to ping database with error: %v", err.Error())
		return &hcr, nil
	}

	return &hcr, nil

}

func main() {

	// setup a span reporter that will support trace information in the zipkin ui
	reporter := reporterhttp.NewReporter(os.Getenv("ZIPKIN_ENDPOINT_URL"))
	defer reporter.Close()

	// create the local service endpoint
	zipkinendpoint := []string {os.Getenv("ANALYSIS_SERVICE_HOST"), ":", os.Getenv("ANALYSIS_SERVICE_PORT")}
	zipkinendpointstr := strings.Join(zipkinendpoint,"")
	endpoint, err := zipkin.NewEndpoint("analysis-service", zipkinendpointstr)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create local zipkin endpoint with error: %v", err))
	}

	// initialize the tracer
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create zipkin tracer with error: %v", err))
	}

	zipkinendpoint = nil
	zipkinendpoint = []string {":", os.Getenv("ANALYSIS_SERVICE_PORT")}
	zipkinendpointstr = strings.Join(zipkinendpoint,"")

	listen, err := net.Listen("tcp", zipkinendpointstr)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to listen on tcp port with error: %v", err))
	}

	svr := grpc.NewServer(grpc.StatsHandler(zipkingrpc.NewServerHandler(tracer)))

	pb.RegisterAnalysisServiceServer(svr, &server{})
	if err := svr.Serve(listen); err != nil {
		logger.Fatal(fmt.Sprintf("failed to serve on tcp port with error: %v", err))
	}
}
