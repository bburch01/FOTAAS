package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bburch01/FOTAAS/api"
	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	pbts "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/joho/godotenv"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func init() {
	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv error: %v", err)
	}
}

func main() {

	/*
		if resp, err := transmit(); err != nil {
			log.Fatalf("an error occurred while trying to transmit telemetry data: %v", err)
		} else {
			for i, v := range resp.ServerStatus {
				log.Printf("uuid: %v status code: %v status msg: %v", i, v.Code, v.Message)
			}
		}
	*/

	/*
		if resp, err := runSimulation(); err != nil {
			log.Fatalf("an error occurred while trying to run the simulation: %v", err)
		} else {
			for i, v := range resp.ServerStatus {
				log.Printf("uuid: %v status code: %v status msg: %v", i, v.Code, v.Message)
			}
		}
	*/

}

func transmit() (*api.TransmitTelemetryResponse, error) {

	var svcaddr string

	var resp *api.TransmitTelemetryResponse
	var req api.TransmitTelemetryRequest
	var td api.TelemetryData
	var tdm map[string]*api.TelemetryDatum

	td.GranPrix = api.GranPrix_ABU_DHABI
	td.Track = api.Track_YAS_MARINA
	td.Constructor = api.Constructor_MERCEDES
	td.CarNumber = 44

	tdm = make(map[string]*api.TelemetryDatum)

	simID := uuid.New().String()
	tdm[simID] = newTelemetryDatum(simID, api.TelemetryDatumDescription_ENGINE_COOLANT_TEMP, api.TelemetryDatumUnit_DEGREE_CELCIUS,
		ipbts.TimestampNow(), 24.468994, 54.601784, 7.0, 75.65, false, false)

	simID = uuid.New().String()
	tdm[simID] = newTelemetryDatum(simID, api.TelemetryDatumDescription_ENGINE_OIL_PRESSURE, api.TelemetryDatumUnit_BAR,
		ipbts.TimestampNow(), 24.468994, 54.601784, 7.0, 143.78, false, false)

	simID = uuid.New().String()
	tdm[simID] = newTelemetryDatum(simID, api.TelemetryDatumDescription_ENGINE_RPM, api.TelemetryDatumUnit_RPM,
		ipbts.TimestampNow(), 24.468994, 54.601784, 7.0, 11558.74, false, false)

	simID = uuid.New().String()
	tdm[simID] = newTelemetryDatum(simID, api.TelemetryDatumDescription_G_FORCE, api.TelemetryDatumUnit_G,
		ipbts.TimestampNow(), 24.468994, 54.601784, 7.0, 2.45, false, false)

	simID = uuid.New().String()
	tdm[simID] = newTelemetryDatum(simID, api.TelemetryDatumDescription_G_FORCE_DIRECTION, api.TelemetryDatumUnit_RADIAN,
		ipbts.TimestampNow(), 24.468994, 54.601784, 7.0, 3.85, false, false)

	td.TelemetryDatumMap = tdm

	req.TelemetryData = &td

	svcaddr = os.Getenv("TELEMETRY_SERVICE_HOST") + os.Getenv("TELEMETRY_SERVICE_PORT")

	conn, err := grpc.Dial(svcaddr, grpc.WithInsecure())
	if err != nil {
		return resp, err
	}
	defer conn.Close()

	//time.Sleep(time.Duration(sampleRateInMillis) * time.Millisecond)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//currentSimTime = currentSimTime.Add(time.Duration(sampleRateInMillis) * time.Millisecond)

	//ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)

	// TODO: determine what is the appropriate deadline for transmit requests, possibly scaling
	// based on the size of TelemetryData
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client api.TelemetryServiceClient

	client = api.NewTelemetryServiceClient(conn)

	resp, err = client.TransmitTelemetry(ctx, &req)
	if err != nil {
		return resp, err
	}

	return resp, nil

}

func runSimulation() (*api.RunSimulationResponse, error) {

	var svcaddr string

	var resp *api.RunSimulationResponse
	var req api.RunSimulationRequest
	var simID string

	//TODO: the model now is 1 sim to many simMember
	//simmap := make(map[string]*api.Simulation)

	simID = uuid.New().String()
	sim := api.Simulation{Uuid: simID, DurationInMinutes: int32(1), SampleRate: api.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: api.SimulationRateMultiplier_X2, GranPrix: api.GranPrix_UNITED_STATES,
		Track: api.Track_AUSTIN,
	}
	//simmap[simID] = &sim

	/*
		simID = uuid.New().String()
		sim = api.Simulation{Uuid: simID, DurationInMinutes: int32(1), SampleRate: api.SampleRate_SR_1000_MS,
			SimulationRateMultiplier: api.SimulationRateMultiplier_X2, GranPrix: api.GranPrix_UNITED_STATES,
			Track: api.Track_AUSTIN,
		}
		simmap[simID] = &sim
	*/

	req.Simulation = &sim

	svcaddr = os.Getenv("SIMULATION_SERVICE_HOST") + os.Getenv("SIMULATION_SERVICE_PORT")

	conn, err := grpc.Dial(svcaddr, grpc.WithInsecure())
	if err != nil {
		return resp, err
	}
	defer conn.Close()

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)

	// TODO: determine what is the appropriate deadline for transmit requests, possibly scaling
	// based on the size of SimulationMap.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client api.SimulationServiceClient

	client = api.NewSimulationServiceClient(conn)

	resp, err = client.RunSimulation(ctx, &req)
	if err != nil {
		return resp, err
	}

	return resp, nil

}

func newTelemetryDatum(uuid string, desc api.TelemetryDatumDescription, unit api.TelemetryDatumUnit,
	tstamp *pbts.Timestamp, lat float64, lon float64, elev float64, val float64,
	ha bool, la bool) *api.TelemetryDatum {

	return &api.TelemetryDatum{Uuid: uuid, Description: desc, Unit: unit, Timestamp: tstamp,
		Latitude: lat, Longitude: lon, Elevation: elev, Value: val, HighAlarm: ha, LowAlarm: la}

}
