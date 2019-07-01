package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	zgrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	zhttp "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/app/simulation"
	"github.com/bburch01/FOTAAS/internal/app/simulation/data"
	"github.com/bburch01/FOTAAS/internal/app/simulation/models"
	"github.com/bburch01/FOTAAS/internal/app/telemetry"
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
	if err = models.InitDB(); err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize database driver with error: %v", err))
	}
}

func (s *server) HealthCheck(ctx context.Context, in *api.HealthCheckRequest) (*api.HealthCheckResponse, error) {
	var hcr = api.HealthCheckResponse{ServerStatus: &api.ServerStatus{
		Code: api.StatusCode_OK, Message: "simulation service healthy"}}

	if err := models.PingDB(); err != nil {
		hcr.ServerStatus.Code = api.StatusCode_ERROR
		hcr.ServerStatus.Message = fmt.Sprintf("failed to ping database with error: %v", err.Error())
		return &hcr, nil
	}

	return &hcr, nil
}

// Need to update this function such that it no longer returns a api.RunSimulationResponse or an error
// All status/results information needs to be persisted to the simulation service db, then clients
// will use this service to get simulation status/results during & after the simulation run
func (s *server) RunSimulation(ctx context.Context, req *api.RunSimulationRequest) (*api.RunSimulationResponse, error) {

	var resp api.RunSimulationResponse
	//var sim = req.Simulation
	//var simMemberMap = req.Simulation.SimulationMemberMap
	//var status api.ServerStatus
	var simData map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
	var simMemberData map[string]map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
	var wg sync.WaitGroup
	var err error

	//REFACTOR OPPORTUNITY
	// This is really ugly but until there is a clever refactor, convert the proto objects
	// contained in req into FOTAAS domain model objects in order to gain db CRUD behaviors
	// This is necessary because the protobuf code cannot be modified. The refactor might
	// be based on wrapping the protobuf objects by FOTAAS domain model objects.
	var sim *models.Simulation = models.NewFromRunSimulationRequest(*req)

	resultsChan := make(chan simulation.SimResult, len(sim.SimulationMembers))
	wg.Add(len(sim.SimulationMembers))

	// For each entry in the simMemberMap (i.e. for each car running in the simulation), validate the simMember
	// and generate telemetry data for it. If validation or data generation fails for any member, don't run
	// the simulation.
	for _, v := range sim.SimulationMembers {
		if err = validate(v); err != nil {
			resp.ServerStatus.Code = api.StatusCode_ERROR
			resp.ServerStatus.Message = fmt.Sprintf("simulation member validation failed with error: %v", err)
			return &resp, nil
		}
		if simData, err = data.GenerateSimulatedTelemetryData(sim, v); err != nil {
			resp.ServerStatus.Code = api.StatusCode_ERROR
			resp.ServerStatus.Message = fmt.Sprintf("simulation member data generation failed with error: %v", err)
			return &resp, nil
		}
		simMemberData[v.ID] = simData
	}

	/*
		for _, v := range sim.SimulationMembers {
			go simulation.StartSimulation(simMemberData[v.ID], *v, &wg, resultsChan)
		}
	*/

	wg.Wait()
	close(resultsChan)

	// Need to convert to persisting simulation results to the simulation service db
	/*
		for res := range resultsChan {
			statusMap[res.UUID] = &res.Status
		}
	*/

	//resp.ServerStatus = statusMap
	//return &resp, nil
	return &resp, nil
}

func (s *server) GetSimulationStatus(ctx context.Context, in *api.GetSimulationStatusRequest) (*api.GetSimulationStatusResponse, error) {
	var gssr = api.GetSimulationStatusResponse{}
	return &gssr, nil
}

func main() {
	var sb strings.Builder

	reporter := zhttp.NewReporter(os.Getenv("ZIPKIN_ENDPOINT_URL"))
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

	svr := grpc.NewServer(grpc.StatsHandler(zgrpc.NewServerHandler(tracer)))

	api.RegisterSimulationServiceServer(svr, &server{})

	if err := svr.Serve(listener); err != nil {
		logger.Fatal(fmt.Sprintf("failed to serve on simulation service port %v with error: %v", simulationSvcPort, err))
	}
}

func validate(simMember models.SimulationMember) error {
	if _, err := uuid.Parse(simMember.ID); err != nil {
		return err
	}
	return nil
}
