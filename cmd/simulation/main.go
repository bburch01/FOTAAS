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
	"github.com/bburch01/FOTAAS/internal/app/simulation"
	"github.com/bburch01/FOTAAS/internal/app/simulation/models"
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

	//var resp api.RunSimulationResponse

	var rsr = api.RunSimulationResponse{ServerStatus: &api.ServerStatus{
		Code: api.StatusCode_OK, Message: fmt.Sprintf("simulation %v successfully started", req.Simulation.Uuid)}}

	//var sim = req.Simulation
	//var simMemberMap = req.Simulation.SimulationMemberMap
	//var status api.ServerStatus
	//var simData map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
	//var simMemberData map[string]map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
	//var wg sync.WaitGroup
	var err error

	// Validate the simulation request
	if err = validateSimulationRequest(req); err != nil {
		rsr.ServerStatus.Code = api.StatusCode_ERROR
		rsr.ServerStatus.Message = fmt.Sprintf("RunSimulationRequest failed validation: %v", err)
		return &rsr, nil
	}

	for _, v := range req.Simulation.SimulationMemberMap {
		logger.Debug(fmt.Sprintf("req simulation member: %v", *v))
	}

	//REFACTOR OPPORTUNITY
	// This is really ugly but until a clever refactor, convert the proto objects
	// contained in req into FOTAAS domain model objects in order to gain db CRUD behaviors.
	// This is necessary because the protobuf code cannot be modified. One of the big
	// problems with this is that all the useful enums in the protobuf objects get lost
	// in translation. A possible refactor would be to have the FOTAAS domain model
	// object wrap the protobuf object and redeclare all of the enums.
	var sim *models.Simulation = models.NewFromRunSimulationRequest(*req)

	// Start the simulation asynchronously (i.e. don't wait on a response from the goroutine).
	// Simulation progress/status is persisted to the FOTAAS simulation db.
	go simulation.StartSimulation(sim)

	return &rsr, nil

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

func validateSimulationRequest(req *api.RunSimulationRequest) error {

	var sb strings.Builder
	var invalidRequest = false

	sb.WriteString("invalid RunSimulationRequest: ")

	if _, err := uuid.Parse(req.Simulation.Uuid); err != nil {
		sb.WriteString("error: invalid uuid")
		invalidRequest = true
	}

	if req.Simulation.DurationInMinutes <= 0 {
		sb.WriteString(" error: DurationInMinutes must be > 0")
		invalidRequest = true
	}

	switch req.Simulation.SampleRate {
	case api.SampleRate_SR_1_MS:
		break
	case api.SampleRate_SR_10_MS:
		break
	case api.SampleRate_SR_100_MS:
		break
	case api.SampleRate_SR_1000_MS:
		break
	default:
		sb.WriteString(" error: invalid SampleRate")
		invalidRequest = true
	}

	switch req.Simulation.SimulationRateMultiplier {
	case api.SimulationRateMultiplier_X1:
		break
	case api.SimulationRateMultiplier_X2:
		break
	case api.SimulationRateMultiplier_X4:
		break
	case api.SimulationRateMultiplier_X8:
		break
	case api.SimulationRateMultiplier_X10:
		break
	case api.SimulationRateMultiplier_X20:
		break
	default:
		sb.WriteString(" error: invalid SimulationRateMultiplier")
		invalidRequest = true
	}

	switch req.Simulation.GrandPrix {
	case api.GrandPrix_ABU_DHABI, api.GrandPrix_AUSTRALIAN, api.GrandPrix_AUSTRIAN, api.GrandPrix_AZERBAIJAN,
		api.GrandPrix_BAHRAIN, api.GrandPrix_BELGIAN, api.GrandPrix_BRAZILIAN, api.GrandPrix_BRITISH,
		api.GrandPrix_CANADIAN, api.GrandPrix_CHINESE, api.GrandPrix_FRENCH, api.GrandPrix_GERMAN,
		api.GrandPrix_HUNGARIAN, api.GrandPrix_ITALIAN, api.GrandPrix_JAPANESE, api.GrandPrix_MEXICAN,
		api.GrandPrix_MONACO, api.GrandPrix_RUSSIAN, api.GrandPrix_SINGAPORE, api.GrandPrix_SPANISH,
		api.GrandPrix_UNITED_STATES:
		break
	default:
		sb.WriteString(" error: invalid GrandPrix")
		invalidRequest = true
	}

	switch req.Simulation.Track {
	case api.Track_AUSTIN, api.Track_BAKU, api.Track_CATALUNYA_BARCELONA, api.Track_HOCKENHEIM,
		api.Track_HUNGARORING, api.Track_INTERLAGOS_SAU_PAULO, api.Track_MARINA_BAY,
		api.Track_MELBOURNE, api.Track_MEXICO_CITY, api.Track_MONTE_CARLO, api.Track_MONTREAL,
		api.Track_MONZA, api.Track_PAUL_RICARD_LE_CASTELLET, api.Track_SAKHIR,
		api.Track_SHANGHAI, api.Track_SILVERSTONE, api.Track_SOCHI, api.Track_SPA_FRANCORCHAMPS,
		api.Track_SPIELBERG_RED_BULL_RING, api.Track_SUZUKA, api.Track_YAS_MARINA:
		break
	default:
		sb.WriteString(" error: invalid GrandPrix")
		invalidRequest = true
	}

	if req.Simulation.SimulationMemberMap == nil {
		sb.WriteString(" error: SimulationMemberMap must not be nil")
		invalidRequest = true
	}

	if len(req.Simulation.SimulationMemberMap) == 0 {
		sb.WriteString(" error: SimulationMemberMap must contain at least 1 member")
		invalidRequest = true
	} else {
		// Short-circuit on the first bad member but at least report all the errors
		// found with that member.
		for _, v := range req.Simulation.SimulationMemberMap {

			if _, err := uuid.Parse(v.Uuid); err != nil {
				sb.WriteString(" simulation member ")
				sb.WriteString(v.Uuid)
				sb.WriteString(" error: invalid uuid")
				invalidRequest = true
			}

			if _, err := uuid.Parse(v.SimulationUuid); err != nil {
				sb.WriteString(" simulation member ")
				sb.WriteString(v.Uuid)
				sb.WriteString(" error: invalid simulation uuid")
				invalidRequest = true
			}

			switch v.Constructor {
			case api.Constructor_ALPHA_ROMEO, api.Constructor_FERRARI, api.Constructor_HAAS, api.Constructor_MCLAREN,
				api.Constructor_MERCEDES, api.Constructor_RACING_POINT, api.Constructor_RED_BULL_RACING,
				api.Constructor_SCUDERIA_TORO_ROSO, api.Constructor_WILLIAMS:
				break
			default:
				sb.WriteString(" simulation member ")
				sb.WriteString(v.Uuid)
				sb.WriteString(" error: invalid constructor")
				invalidRequest = true
			}

			if v.CarNumber < 0 {
				sb.WriteString(" simulation member ")
				sb.WriteString(v.Uuid)
				sb.WriteString(" error: CarNumber must be > 0")
				invalidRequest = true
			}

			if v.ForceAlarm && v.NoAlarms {
				sb.WriteString(" simulation member ")
				sb.WriteString(v.Uuid)
				sb.WriteString(" error: ForceAlarms & NoAlarms must not both be true")
				invalidRequest = true
			}

			if invalidRequest {
				break
			}

		}
	}

	if invalidRequest {
		return fmt.Errorf("%v", sb.String())
	}

	return nil
}
