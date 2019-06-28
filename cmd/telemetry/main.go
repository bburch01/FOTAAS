package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	pb "github.com/bburch01/FOTAAS/api"
	mdl "github.com/bburch01/FOTAAS/internal/app/telemetry/models"
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
	var hcr = pb.HealthCheckResponse{ServerStatus: &pb.ServerStatus{Code: pb.StatusCode_OK, Message: "telemetry service healthy"}}

	if err := mdl.PingDB(); err != nil {
		hcr.ServerStatus.Code = pb.StatusCode_ERROR
		hcr.ServerStatus.Message = fmt.Sprintf("failed to ping database with error: %v", err.Error())
		return &hcr, nil
	}

	return &hcr, nil

}

func (s *server) TransmitTelemetry(ctx context.Context, in *pb.TransmitTelemetryRequest) (*pb.TransmitTelemetryResponse, error) {

	var resp pb.TransmitTelemetryResponse
	var tdmap = in.TelemetryData.TelemetryDatumMap
	var td mdl.TelemetryDatum
	var status pb.ServerStatus
	var ttsm = make(map[string]*pb.ServerStatus)

	td.GrandPrix = in.TelemetryData.GrandPrix.String()
	td.Track = in.TelemetryData.Track.String()
	td.Constructor = in.TelemetryData.Constructor.String()
	td.CarNumber = in.TelemetryData.CarNumber

	for i, v := range tdmap {
		err := validate(v)
		if err != nil {
			status.Code = pb.StatusCode_ERROR
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
				status.Code = pb.StatusCode_ERROR
				status.Message = fmt.Sprintf("server side error: %v", err)
			} else {
				status.Code = pb.StatusCode_OK
				status.Message = fmt.Sprintf("telemetry datum successfully processed.")
				logger.Info(fmt.Sprintf("successfully processed telemetry datum uuid: %v", td.ID))
			}
		}
		ttsm[i] = &status
	}

	resp.ServerStatus = ttsm
	return &resp, nil
}

func validate(td *pb.TelemetryDatum) error {

	// Check the uuid for valid format
	if _, err := uid.Parse(td.Uuid); err != nil {
		return err
	}

	// Check that the telemetry datum unit is valid for the description
	switch td.Description {
	case pb.TelemetryDatumDescription_BRAKE_TEMP_FL, pb.TelemetryDatumDescription_BRAKE_TEMP_FR, pb.TelemetryDatumDescription_BRAKE_TEMP_RL,
		pb.TelemetryDatumDescription_BRAKE_TEMP_RR, pb.TelemetryDatumDescription_ENGINE_COOLANT_TEMP, pb.TelemetryDatumDescription_ENGINE_OIL_TEMP,
		pb.TelemetryDatumDescription_ENERGY_STORAGE_TEMP, pb.TelemetryDatumDescription_TIRE_TEMP_FL, pb.TelemetryDatumDescription_TIRE_TEMP_FR,
		pb.TelemetryDatumDescription_TIRE_TEMP_RL, pb.TelemetryDatumDescription_TIRE_TEMP_RR:
		if td.Unit != pb.TelemetryDatumUnit_DEGREE_CELCIUS {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_DEGREE_CELCIUS got %v",
				td.Description.String(), td.Unit.String())
		}
	case pb.TelemetryDatumDescription_TIRE_PRESSURE_FL, pb.TelemetryDatumDescription_TIRE_PRESSURE_FR, pb.TelemetryDatumDescription_TIRE_PRESSURE_RL,
		pb.TelemetryDatumDescription_TIRE_PRESSURE_RR:
		if td.Unit != pb.TelemetryDatumUnit_BAR {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_BAR got %v",
				td.Description.String(), td.Unit.String())
		}
	case pb.TelemetryDatumDescription_MGUK_OUTPUT, pb.TelemetryDatumDescription_MGUH_OUTPUT:
		if td.Unit != pb.TelemetryDatumUnit_JPS {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_JPS got %v",
				td.Description.String(), td.Unit.String())
		}
	case pb.TelemetryDatumDescription_SPEED:
		if td.Unit != pb.TelemetryDatumUnit_KPH {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KPH got %v",
				td.Description.String(), td.Unit.String())
		}
	case pb.TelemetryDatumDescription_ENGINE_OIL_PRESSURE:
		if td.Unit != pb.TelemetryDatumUnit_KPA {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KPA got %v",
				td.Description.String(), td.Unit.String())
		}
	case pb.TelemetryDatumDescription_G_FORCE:
		if td.Unit != pb.TelemetryDatumUnit_G {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_G got %v",
				td.Description.String(), td.Unit.String())
		}
	case pb.TelemetryDatumDescription_FUEL_CONSUMED:
		if td.Unit != pb.TelemetryDatumUnit_KG {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KG got %v",
				td.Description.String(), td.Unit.String())
		}
	case pb.TelemetryDatumDescription_FUEL_FLOW:
		if td.Unit != pb.TelemetryDatumUnit_KG_PER_HOUR {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KG_PER_HOUR got %v",
				td.Description.String(), td.Unit.String())
		}
	case pb.TelemetryDatumDescription_ENGINE_RPM:
		if td.Unit != pb.TelemetryDatumUnit_RPM {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_RPM got %v",
				td.Description.String(), td.Unit.String())
		}
	case pb.TelemetryDatumDescription_ENERGY_STORAGE_LEVEL:
		if td.Unit != pb.TelemetryDatumUnit_MJ {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_MJ got %v",
				td.Description.String(), td.Unit.String())
		}
	case pb.TelemetryDatumDescription_G_FORCE_DIRECTION:
		if td.Unit != pb.TelemetryDatumUnit_RADIAN {
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

	reporter := reporterhttp.NewReporter(os.Getenv("ZIPKIN_ENDPOINT_URL"))
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

	svr := grpc.NewServer(grpc.StatsHandler(zipkingrpc.NewServerHandler(tracer)))

	pb.RegisterTelemetryServiceServer(svr, &server{})

	if err := svr.Serve(listener); err != nil {
		logger.Fatal(fmt.Sprintf("failed to serve on telemetry service port %v with error: %v", telemetrySvcPort, err))
	}

	/*
		reporter := reporterhttp.NewReporter(os.Getenv("ZIPKIN_ENDPOINT_URL"))
		defer reporter.Close()

		zipkinendpoint := []string {os.Getenv("TELEMETRY_SERVICE_HOST"), ":", os.Getenv("TELEMETRY_SERVICE_PORT")}
		zipkinendpointstr := strings.Join(zipkinendpoint,"")
		endpoint, err := zipkin.NewEndpoint("telemetry-service", zipkinendpointstr)

		if err != nil {
			logger.Fatal(fmt.Sprintf("failed to create local zipkin endpoint with error: %v", err))
		}

		tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
		if err != nil {
			logger.Fatal(fmt.Sprintf("failed to create zipkin tracer with error: %v", err))
		}

		zipkinendpoint = nil
		zipkinendpoint = []string {":", os.Getenv("TELEMETRY_SERVICE_PORT")}
		zipkinendpointstr = strings.Join(zipkinendpoint,"")

		listen, err := net.Listen("tcp", zipkinendpointstr)

		if err != nil {
			logger.Fatal(fmt.Sprintf("failed to listen on tcp port with error: %v", err))
		}

		svr := grpc.NewServer(grpc.StatsHandler(zipkingrpc.NewServerHandler(tracer)))

		pb.RegisterTelemetryServiceServer(svr, &server{})
		if err := svr.Serve(listen); err != nil {
			logger.Fatal(fmt.Sprintf("failed to serve on tcp port with error: %v", err))
		}
	*/
}
