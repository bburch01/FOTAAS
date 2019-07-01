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
	var hcr = api.HealthCheckResponse{ServerStatus: &api.ServerStatus{Code: api.StatusCode_OK, Message: "telemetry service healthy"}}

	if err := models.PingDB(); err != nil {
		hcr.ServerStatus.Code = api.StatusCode_ERROR
		hcr.ServerStatus.Message = fmt.Sprintf("failed to ping database with error: %v", err.Error())
		return &hcr, nil
	}

	return &hcr, nil

}

func (s *server) TransmitTelemetry(ctx context.Context, in *api.TransmitTelemetryRequest) (*api.TransmitTelemetryResponse, error) {

	var resp api.TransmitTelemetryResponse
	var tdmap = in.TelemetryData.TelemetryDatumMap
	var td models.TelemetryDatum
	var status api.ServerStatus
	var ttsm = make(map[string]*api.ServerStatus)

	td.GrandPrix = in.TelemetryData.GrandPrix.String()
	td.Track = in.TelemetryData.Track.String()
	td.Constructor = in.TelemetryData.Constructor.String()
	td.CarNumber = in.TelemetryData.CarNumber

	for i, v := range tdmap {
		err := validate(v)
		if err != nil {
			status.Code = api.StatusCode_ERROR
			status.Message = fmt.Sprintf("telemetry datum validation failed with error: %v", err)
		} else {
			td.ID = v.Uuid
			if v.Simulated {
				td.Simulated = v.Simulated
				td.SimulationID = v.SimulationUuid
				td.SimulationTransmitSequenceNumber = v.SimulationTransmitSequenceNumber
			}
			td.Description = v.Description.String()
			td.Unit = v.Unit.String()
			td.Timestamp = v.Timestamp
			td.Latitude = v.Latitude
			td.Longitude = v.Longitude
			td.Elevation = v.Elevation
			td.Value = v.Value
			td.HiAlarm = v.HighAlarm
			td.LoAlarm = v.LowAlarm
			err = td.Create()
			if err != nil {
				status.Code = api.StatusCode_ERROR
				status.Message = fmt.Sprintf("server side error: %v", err)
			} else {
				status.Code = api.StatusCode_OK
				status.Message = fmt.Sprintf("telemetry datum successfully processed.")
				logger.Info(fmt.Sprintf("successfully processed telemetry datum uuid: %v", td.ID))
			}
		}
		ttsm[i] = &status
	}

	resp.ServerStatus = ttsm
	return &resp, nil
}

func validate(td *api.TelemetryDatum) error {

	// Check the uuid for valid format
	if _, err := uuid.Parse(td.Uuid); err != nil {
		return err
	}

	// Check that the telemetry datum unit is valid for the description
	switch td.Description {
	case api.TelemetryDatumDescription_BRAKE_TEMP_FL, api.TelemetryDatumDescription_BRAKE_TEMP_FR, api.TelemetryDatumDescription_BRAKE_TEMP_RL,
		api.TelemetryDatumDescription_BRAKE_TEMP_RR, api.TelemetryDatumDescription_ENGINE_COOLANT_TEMP, api.TelemetryDatumDescription_ENGINE_OIL_TEMP,
		api.TelemetryDatumDescription_ENERGY_STORAGE_TEMP, api.TelemetryDatumDescription_TIRE_TEMP_FL, api.TelemetryDatumDescription_TIRE_TEMP_FR,
		api.TelemetryDatumDescription_TIRE_TEMP_RL, api.TelemetryDatumDescription_TIRE_TEMP_RR:
		if td.Unit != api.TelemetryDatumUnit_DEGREE_CELCIUS {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_DEGREE_CELCIUS got %v",
				td.Description.String(), td.Unit.String())
		}
	case api.TelemetryDatumDescription_TIRE_PRESSURE_FL, api.TelemetryDatumDescription_TIRE_PRESSURE_FR, api.TelemetryDatumDescription_TIRE_PRESSURE_RL,
		api.TelemetryDatumDescription_TIRE_PRESSURE_RR:
		if td.Unit != api.TelemetryDatumUnit_BAR {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_BAR got %v",
				td.Description.String(), td.Unit.String())
		}
	case api.TelemetryDatumDescription_MGUK_OUTPUT, api.TelemetryDatumDescription_MGUH_OUTPUT:
		if td.Unit != api.TelemetryDatumUnit_JPS {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_JPS got %v",
				td.Description.String(), td.Unit.String())
		}
	case api.TelemetryDatumDescription_SPEED:
		if td.Unit != api.TelemetryDatumUnit_KPH {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KPH got %v",
				td.Description.String(), td.Unit.String())
		}
	case api.TelemetryDatumDescription_ENGINE_OIL_PRESSURE:
		if td.Unit != api.TelemetryDatumUnit_KPA {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KPA got %v",
				td.Description.String(), td.Unit.String())
		}
	case api.TelemetryDatumDescription_G_FORCE:
		if td.Unit != api.TelemetryDatumUnit_G {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_G got %v",
				td.Description.String(), td.Unit.String())
		}
	case api.TelemetryDatumDescription_FUEL_CONSUMED:
		if td.Unit != api.TelemetryDatumUnit_KG {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KG got %v",
				td.Description.String(), td.Unit.String())
		}
	case api.TelemetryDatumDescription_FUEL_FLOW:
		if td.Unit != api.TelemetryDatumUnit_KG_PER_HOUR {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KG_PER_HOUR got %v",
				td.Description.String(), td.Unit.String())
		}
	case api.TelemetryDatumDescription_ENGINE_RPM:
		if td.Unit != api.TelemetryDatumUnit_RPM {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_RPM got %v",
				td.Description.String(), td.Unit.String())
		}
	case api.TelemetryDatumDescription_ENERGY_STORAGE_LEVEL:
		if td.Unit != api.TelemetryDatumUnit_MJ {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_MJ got %v",
				td.Description.String(), td.Unit.String())
		}
	case api.TelemetryDatumDescription_G_FORCE_DIRECTION:
		if td.Unit != api.TelemetryDatumUnit_RADIAN {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_RADIAN got %v",
				td.Description.String(), td.Unit.String())
		}
	default:
		return fmt.Errorf("invalid telemetry datum description %v", td.Description)
	}

	return nil
}

func main() {

	var sb strings.Builder

	reporter := zhttp.NewReporter(os.Getenv("ZIPKIN_ENDPOINT_URL"))
	defer reporter.Close()

	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_PORT"))
	telemetrySvcEndpoint := sb.String()

	zipkinLocalEndpoint, err := zipkin.NewEndpoint("telemetry-service", telemetrySvcEndpoint)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create zipkin local endpoint with error: %v", err))
	}

	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zipkinLocalEndpoint))
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create zipkin tracer with error: %v", err))
	}

	sb.Reset()
	sb.WriteString(":")
	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_PORT"))
	telemetrySvcPort := sb.String()

	listener, err := net.Listen("tcp", telemetrySvcPort)
	if err != nil {
		logger.Fatal(fmt.Sprintf("tcp failed to listen on telemetry service port %v with error: %v", telemetrySvcPort, err))
	}

	svr := grpc.NewServer(grpc.StatsHandler(zgrpc.NewServerHandler(tracer)))

	api.RegisterTelemetryServiceServer(svr, &server{})

	if err := svr.Serve(listener); err != nil {
		logger.Fatal(fmt.Sprintf("failed to serve on telemetry service port %v with error: %v", telemetrySvcPort, err))
	}

}
