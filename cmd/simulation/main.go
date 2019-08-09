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

func main() {

	reporter := zhttp.NewReporter(os.Getenv("ZIPKIN_ENDPOINT_URL"))
	defer reporter.Close()

	var sb strings.Builder
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

func (s *server) AlivenessCheck(ctx context.Context, req *api.AlivenessCheckRequest) (*api.AlivenessCheckResponse, error) {

	resp := new(api.AlivenessCheckResponse)
	resp.Details = &api.ResponseDetails{Code: api.ResponseCode_OK,
		Message: "simulation service is alive!"}

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

func (s *server) RunSimulation(ctx context.Context, req *api.RunSimulationRequest) (*api.RunSimulationResponse, error) {

	var resp = api.RunSimulationResponse{Details: &api.ResponseDetails{
		Code: api.ResponseCode_OK, Message: fmt.Sprintf("simulation %v successfully started", req.Simulation.Uuid)}}

	if err := validateSimulationRequest(req); err != nil {
		resp.Details.Code = api.ResponseCode_ERROR
		resp.Details.Message = fmt.Sprintf("RunSimulationRequest failed validation: %v", err)
		logger.Error(fmt.Sprintf("RunSimulationRequest failed validation: %v", err))
		// protoc generated code requires error in the return params, return nil here so that clients
		// of this service can process this FOTAAS error differently than other system errors (e.g.
		// if this service is not available). Intercept this error and handle it via response code &
		// message.
		return &resp, nil
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

	return &resp, nil

}

func (s *server) GetSimulationInfo(ctx context.Context, req *api.GetSimulationInfoRequest) (*api.GetSimulationInfoResponse, error) {

	// TODO: validate the request.

	resp := new(api.GetSimulationInfoResponse)
	resp.Details = new(api.ResponseDetails)

	var info *api.SimulationInfo
	var err error

	if info, err = models.RetrieveSimulationInfo(*req); err != nil {
		resp.Details.Code = api.ResponseCode_ERROR
		resp.Details.Message = fmt.Sprintf("failed to retrieve simulation info with error: %v", err)
		logger.Error(fmt.Sprintf("failed to retrieve simulation info with error: %v", err))
		// protoc generated code requires error in the return params, return nil here so that clients
		// of this service can process this FOTAAS error differently than other system errors (e.g.
		// if this service is not available). Intercept this error and handle it via response code &
		// message.
		return resp, nil
	}

	if info == nil {
		resp.Details = &api.ResponseDetails{Code: api.ResponseCode_WARN,
			Message: fmt.Sprintf("no info found for simulation id: %v", req.SimulationUuid)}
		return resp, nil
	}

	resp.Details = &api.ResponseDetails{Code: api.ResponseCode_OK,
		Message: fmt.Sprintf("found info for simulation id: %v", req.SimulationUuid)}
	resp.SimulationInfo = info

	return resp, nil
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

	switch req.Simulation.GranPrix {
	case api.GranPrix_ABU_DHABI, api.GranPrix_AUSTRALIAN, api.GranPrix_AUSTRIAN, api.GranPrix_AZERBAIJAN,
		api.GranPrix_BAHRAIN, api.GranPrix_BELGIAN, api.GranPrix_BRAZILIAN, api.GranPrix_BRITISH,
		api.GranPrix_CANADIAN, api.GranPrix_CHINESE, api.GranPrix_FRENCH, api.GranPrix_GERMAN,
		api.GranPrix_HUNGARIAN, api.GranPrix_ITALIAN, api.GranPrix_JAPANESE, api.GranPrix_MEXICAN,
		api.GranPrix_MONACO, api.GranPrix_RUSSIAN, api.GranPrix_SINGAPORE, api.GranPrix_SPANISH,
		api.GranPrix_UNITED_STATES:
		break
	default:
		sb.WriteString(" error: invalid GranPrix")
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
		sb.WriteString(" error: invalid GranPrix")
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
		// Short-circuit on the first bad member.
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
