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

func (s *server) AlivenessCheck(ctx context.Context, req *api.AlivenessCheckRequest) (*api.AlivenessCheckResponse, error) {

	resp := new(api.AlivenessCheckResponse)
	resp.Details = &api.ResponseDetails{Code: api.ResponseCode_OK,
		Message: "telemetry service alive"}

	if err := models.PingDB(); err != nil {
		resp.Details.Code = api.ResponseCode_ERROR
		resp.Details.Message = fmt.Sprintf("failed to ping database with error: %v", err.Error())
		logger.Error(fmt.Sprintf("failed to ping database with error: %v", err))
		// protoc generated code requires error in the return params, return nil here so that clients
		// of this service call process this FOTAAS error differently than other system errors (e.g.
		// if this service is not available). Intercept this error and handle it via response code &
		// message.
		return resp, nil
	}

	return resp, nil
}

func (s *server) TransmitTelemetry(ctx context.Context, req *api.TransmitTelemetryRequest) (*api.TransmitTelemetryResponse, error) {

	var resp api.TransmitTelemetryResponse
	var datum models.TelemetryDatum
	var datumMap = req.TelemetryData.TelemetryDatumMap
	var status api.ResponseDetails
	var statusMap = make(map[string]*api.ResponseDetails)

	for i, v := range datumMap {
		err := validate(v)
		if err != nil {
			status.Code = api.ResponseCode_ERROR
			status.Message = fmt.Sprintf("telemetry datum validation failed with error: %v", err)
		} else {
			datum.ID = v.Uuid
			if v.Simulated {
				datum.Simulated = v.Simulated
				datum.SimulationID = v.SimulationUuid
				datum.SimulationTransmitSequenceNumber = v.SimulationTransmitSequenceNumber
			}
			datum.Description = v.Description.String()

			datum.GranPrix = v.GranPrix.String()
			datum.Track = v.Track.String()
			datum.Constructor = v.Constructor.String()
			datum.CarNumber = v.CarNumber

			datum.Unit = v.Unit.String()
			datum.Timestamp = v.Timestamp
			datum.Latitude = v.Latitude
			datum.Longitude = v.Longitude
			datum.Elevation = v.Elevation
			datum.Value = v.Value
			datum.HiAlarm = v.HighAlarm
			datum.LoAlarm = v.LowAlarm
			err = datum.Create()
			if err != nil {
				status.Code = api.ResponseCode_ERROR
				status.Message = fmt.Sprintf("server side error: %v", err)
			} else {
				status.Code = api.ResponseCode_OK
				status.Message = fmt.Sprintf("telemetry datum successfully processed.")
			}
		}
		statusMap[i] = &status
	}

	resp.Details = statusMap
	return &resp, nil
}

func (s *server) GetTelemetryData(ctx context.Context, req *api.GetTelemetryDataRequest) (*api.GetTelemetryDataResponse, error) {

	// TODO: need to validate the request (all search terms present and valid)

	var data *api.TelemetryData
	var err error

	resp := new(api.GetTelemetryDataResponse)
	resp.Details = new(api.ResponseDetails)

	if data, err = models.RetrieveTelemetryData(*req); err != nil {
		resp.Details.Code = api.ResponseCode_ERROR
		resp.Details.Message = fmt.Sprintf("failed to retrieve simulated telemetry data with error: %v", err)
		logger.Error(fmt.Sprintf("failed to retrieve simulated telemetry data with error: %v", err))
		// protoc generated code requires error in the return params, return nil here so that clients
		// of this service call process this FOTAAS error differently than other system errors (e.g.
		// if this service is not available). Intercept this error and handle it via response code &
		// message.
		return resp, nil
	}

	resp.Details = &api.ResponseDetails{Code: api.ResponseCode_OK,
		Message: fmt.Sprintf("found %v telemetry datum", len(data.TelemetryDatumMap))}

	resp.TelemetryData = data

	return resp, nil

}

/*
func (s *server) GetSimulatedTelemetryData(ctx context.Context, req *api.GetSimulatedTelemetryDataRequest) (*api.GetSimulatedTelemetryDataResponse, error) {

	// TODO: need to validate the request (all search terms present and valid)

	var resp api.GetSimulatedTelemetryDataResponse
	var data *api.TelemetryData
	var err error

	if data, err = models.RetrieveSimulatedTelemetryData(*req); err != nil {
		resp.Details.Code = api.ResponseCode_ERROR
		resp.Details.Message = fmt.Sprintf("failed to retrieve simulated telemetry data with error: %v", err)
		logger.Error(fmt.Sprintf("failed to retrieve simulated telemetry data with error: %v", err))
		// protoc generated code requires error in the return params, return nil here so that clients
		// of this service call process this FOTAAS error differently than other system errors (e.g.
		// if this service is not available). Intercept this error and handle it via response code &
		// message.
		return &resp, nil
	}

	resp.Details = &api.ResponseDetails{Code: api.ResponseCode_OK,
		Message: fmt.Sprintf("found %v simulated telemetry datum", len(data.TelemetryDatumMap))}

	resp.TelemetryData = data

	return &resp, nil
}
*/

func validate(datum *api.TelemetryDatum) error {

	// Check the uuid for valid format
	if _, err := uuid.Parse(datum.Uuid); err != nil {
		return err
	}

	// Check that the telemetry datum unit is valid for the description
	switch datum.Description {
	case api.TelemetryDatumDescription_BRAKE_TEMP_FL, api.TelemetryDatumDescription_BRAKE_TEMP_FR, api.TelemetryDatumDescription_BRAKE_TEMP_RL,
		api.TelemetryDatumDescription_BRAKE_TEMP_RR, api.TelemetryDatumDescription_ENGINE_COOLANT_TEMP, api.TelemetryDatumDescription_ENGINE_OIL_TEMP,
		api.TelemetryDatumDescription_ENERGY_STORAGE_TEMP, api.TelemetryDatumDescription_TIRE_TEMP_FL, api.TelemetryDatumDescription_TIRE_TEMP_FR,
		api.TelemetryDatumDescription_TIRE_TEMP_RL, api.TelemetryDatumDescription_TIRE_TEMP_RR:
		if datum.Unit != api.TelemetryDatumUnit_DEGREE_CELCIUS {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_DEGREE_CELCIUS got %v",
				datum.Description.String(), datum.Unit.String())
		}
	case api.TelemetryDatumDescription_TIRE_PRESSURE_FL, api.TelemetryDatumDescription_TIRE_PRESSURE_FR, api.TelemetryDatumDescription_TIRE_PRESSURE_RL,
		api.TelemetryDatumDescription_TIRE_PRESSURE_RR:
		if datum.Unit != api.TelemetryDatumUnit_BAR {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_BAR got %v",
				datum.Description.String(), datum.Unit.String())
		}
	case api.TelemetryDatumDescription_MGUK_OUTPUT, api.TelemetryDatumDescription_MGUH_OUTPUT:
		if datum.Unit != api.TelemetryDatumUnit_JPS {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_JPS got %v",
				datum.Description.String(), datum.Unit.String())
		}
	case api.TelemetryDatumDescription_SPEED:
		if datum.Unit != api.TelemetryDatumUnit_KPH {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KPH got %v",
				datum.Description.String(), datum.Unit.String())
		}
	case api.TelemetryDatumDescription_ENGINE_OIL_PRESSURE:
		if datum.Unit != api.TelemetryDatumUnit_KPA {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KPA got %v",
				datum.Description.String(), datum.Unit.String())
		}
	case api.TelemetryDatumDescription_G_FORCE:
		if datum.Unit != api.TelemetryDatumUnit_G {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_G got %v",
				datum.Description.String(), datum.Unit.String())
		}
	case api.TelemetryDatumDescription_FUEL_CONSUMED:
		if datum.Unit != api.TelemetryDatumUnit_KG {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KG got %v",
				datum.Description.String(), datum.Unit.String())
		}
	case api.TelemetryDatumDescription_FUEL_FLOW:
		if datum.Unit != api.TelemetryDatumUnit_KG_PER_HOUR {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_KG_PER_HOUR got %v",
				datum.Description.String(), datum.Unit.String())
		}
	case api.TelemetryDatumDescription_ENGINE_RPM:
		if datum.Unit != api.TelemetryDatumUnit_RPM {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_RPM got %v",
				datum.Description.String(), datum.Unit.String())
		}
	case api.TelemetryDatumDescription_ENERGY_STORAGE_LEVEL:
		if datum.Unit != api.TelemetryDatumUnit_MJ {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_MJ got %v",
				datum.Description.String(), datum.Unit.String())
		}
	case api.TelemetryDatumDescription_G_FORCE_DIRECTION:
		if datum.Unit != api.TelemetryDatumUnit_RADIAN {
			//return errors.New("invalid telemetry datum unit")
			return fmt.Errorf("invalid telemetry datum unit for %v, expected TelemetryDatumUnit_RADIAN got %v",
				datum.Description.String(), datum.Unit.String())
		}
	default:
		return fmt.Errorf("invalid telemetry datum description %v", datum.Description)
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
