package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	pb "github.com/bburch01/FOTAAS/api"
	sim "github.com/bburch01/FOTAAS/internal/app/simulation"
	gen "github.com/bburch01/FOTAAS/internal/app/simulation/data"
	mdl "github.com/bburch01/FOTAAS/internal/app/simulation/models"
	tel "github.com/bburch01/FOTAAS/internal/app/telemetry"
	logging "github.com/bburch01/FOTAAS/internal/pkg/logging"
	uid "github.com/google/uuid"
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
	var hcr = pb.HealthCheckResponse{ServerStatus: &pb.ServerStatus{
		Code: pb.StatusCode_OK, Message: "simulation service healthy"}}

	if err := mdl.PingDB(); err != nil {
		hcr.ServerStatus.Code = pb.StatusCode_ERROR
		hcr.ServerStatus.Message = fmt.Sprintf("failed to ping database with error: %v", err.Error())
		return &hcr, nil
	}

	return &hcr, nil
}

// Need to update this function such that it no longer returns a pb.RunSimulationResponse or an error
// All status/results information needs to be persisted to the simulation service db, then clients
// will use this service to get simulation status/results during & after the simulation run
func (s *server) RunSimulation(ctx context.Context, in *pb.RunSimulationRequest) (*pb.RunSimulationResponse, error) {
	var resp pb.RunSimulationResponse
	var simmap = in.SimulationMap
	var status pb.ServerStatus
	var statusMap = make(map[string]*pb.ServerStatus)
	var simData map[pb.TelemetryDatumDescription]tel.SimulatedTelemetryData
	var wg sync.WaitGroup

	resultsChan := make(chan sim.SimResult, len(simmap))
	wg.Add(len(simmap))

	// For each entry in the simmap (i.e. for each car running in the simulation), generate telemetry data
	// and start a simulation worker.
	for _, v := range simmap {
		err := validate(v)
		if err != nil {
			status.Code = pb.StatusCode_ERROR
			status.Message = fmt.Sprintf("simulation validation failed with error: %v", err)
			statusMap[v.Uuid] = &status
			break
		}
		if simData, err = gen.GenerateSimulatedTelemetryData(*v); err != nil {
			status.Code = pb.StatusCode_ERROR
			status.Message = fmt.Sprintf("failed to generate simulation data with error: %v", err)
			statusMap[v.Uuid] = &status
			logger.Error(fmt.Sprintf("failed to generate simulation data with error: %v", err))
			break
		}
		go sim.StartSimulation(simData, *v, &wg, resultsChan)
	}

	wg.Wait()
	close(resultsChan)

	// Need to convert to persisting simulation results to the simulation service db
	for res := range resultsChan {
		statusMap[res.UUID] = &res.Status
	}

	resp.ServerStatus = statusMap
	return &resp, nil
}

func (s *server) GetSimulationStatus(ctx context.Context, in *pb.GetSimulationStatusRequest) (*pb.GetSimulationStatusResponse, error) {
	var gssr = pb.GetSimulationStatusResponse{}
	return &gssr, nil
}

func main() {
	var sb strings.Builder

	reporter := reporterhttp.NewReporter(os.Getenv("ZIPKIN_ENDPOINT_URL"))
	defer reporter.Close()

	sb.WriteString(os.Getenv("SIMULATION_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("SIMULATION_SERVICE_PORT"))
	simulationSvcEndpoint := sb.String()

	zipkinLocalEndpoint, err := zipkin.NewEndpoint("simulation-service", simulationSvcEndpoint)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create zipkin local endpoint with error: %v", err))
	}

	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zipkinLocalEndpoint))
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create zipkin tracer with error: %v", err))
	}

	sb.Reset()
	sb.WriteString(":")
	sb.WriteString(os.Getenv("SIMULATION_SERVICE_PORT"))
	simulationSvcPort := sb.String()

	listener, err := net.Listen("tcp", simulationSvcPort)
	if err != nil {
		logger.Fatal(fmt.Sprintf("tcp failed to listen on simulation service port %v with error: %v", simulationSvcPort, err))
	}

	svr := grpc.NewServer(grpc.StatsHandler(zipkingrpc.NewServerHandler(tracer)))

	pb.RegisterSimulationServiceServer(svr, &server{})

	if err := svr.Serve(listener); err != nil {
		logger.Fatal(fmt.Sprintf("failed to serve on simulation service port %v with error: %v", simulationSvcPort, err))
	}
}

func validate(sim *pb.Simulation) error {
	if _, err := uid.Parse(sim.Uuid); err != nil {
		return err
	}
	return nil
}
