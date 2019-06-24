package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/bburch01/FOTAAS/api"
	ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/joho/godotenv"

	uid "github.com/google/uuid"
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

	if resp, err := runSimulation(); err != nil {
		log.Fatalf("an error occurred while trying to run the simulation: %v", err)
	} else {
		for i, v := range resp.ServerStatus {
			log.Printf("uuid: %v status code: %v status msg: %v", i, v.Code, v.Message)
		}
	}

}

func transmit() (*pb.TransmitTelemetryResponse, error) {

	var svcaddr string

	var resp *pb.TransmitTelemetryResponse
	var req pb.TransmitTelemetryRequest
	var td pb.TelemetryData
	var tdm map[string]*pb.TelemetryDatum

	td.GrandPrix = pb.GrandPrix_ABU_DHABI
	td.Track = pb.Track_YAS_MARINA
	td.Constructor = pb.Constructor_MERCEDES
	td.CarNumber = 44

	tdm = make(map[string]*pb.TelemetryDatum)

	uuid := uid.New().String()
	tdm[uuid] = newTelemetryDatum(uuid, pb.TelemetryDatumDescription_ENGINE_COOLANT_TEMP, pb.TelemetryDatumUnit_DEGREE_CELCIUS,
		ts.TimestampNow(), 24.468994, 54.601784, 7.0, 75.65, false, false)

	uuid = uid.New().String()
	tdm[uuid] = newTelemetryDatum(uuid, pb.TelemetryDatumDescription_ENGINE_OIL_PRESSURE, pb.TelemetryDatumUnit_BAR,
		ts.TimestampNow(), 24.468994, 54.601784, 7.0, 143.78, false, false)

	uuid = uid.New().String()
	tdm[uuid] = newTelemetryDatum(uuid, pb.TelemetryDatumDescription_ENGINE_RPM, pb.TelemetryDatumUnit_RPM,
		ts.TimestampNow(), 24.468994, 54.601784, 7.0, 11558.74, false, false)

	uuid = uid.New().String()
	tdm[uuid] = newTelemetryDatum(uuid, pb.TelemetryDatumDescription_G_FORCE, pb.TelemetryDatumUnit_G,
		ts.TimestampNow(), 24.468994, 54.601784, 7.0, 2.45, false, false)

	uuid = uid.New().String()
	tdm[uuid] = newTelemetryDatum(uuid, pb.TelemetryDatumDescription_G_FORCE_DIRECTION, pb.TelemetryDatumUnit_RADIAN,
		ts.TimestampNow(), 24.468994, 54.601784, 7.0, 3.85, false, false)

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

	var client pb.TelemetryServiceClient

	client = pb.NewTelemetryServiceClient(conn)

	resp, err = client.TransmitTelemetry(ctx, &req)
	if err != nil {
		return resp, err
	}

	return resp, nil

}

func runSimulation() (*pb.RunSimulationResponse, error) {

	var svcaddr string

	var resp *pb.RunSimulationResponse
	var req pb.RunSimulationRequest
	var simID string

	simmap := make(map[string]*pb.Simulation)

	simID = uid.New().String()
	sim := pb.Simulation{Uuid: simID, DurationInMinutes: int32(1), SampleRate: pb.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: pb.SimulationRateMultiplier_X2, GrandPrix: pb.GrandPrix_UNITED_STATES,
		Track: pb.Track_AUSTIN, Constructor: pb.Constructor_HAAS, CarNumber: 8, ForceAlarm: false, NoAlarms: true,
	}
	simmap[simID] = &sim

	simID = uid.New().String()
	sim = pb.Simulation{Uuid: simID, DurationInMinutes: int32(1), SampleRate: pb.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: pb.SimulationRateMultiplier_X2, GrandPrix: pb.GrandPrix_UNITED_STATES,
		Track: pb.Track_AUSTIN, Constructor: pb.Constructor_MERCEDES, CarNumber: 44, ForceAlarm: false, NoAlarms: true,
	}
	simmap[simID] = &sim

	req.SimulationMap = simmap

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

	var client pb.SimulationServiceClient

	client = pb.NewSimulationServiceClient(conn)

	resp, err = client.RunSimulation(ctx, &req)
	if err != nil {
		return resp, err
	}

	return resp, nil

}

func newTelemetryDatum(uuid string, desc pb.TelemetryDatumDescription, unit pb.TelemetryDatumUnit,
	tstamp *timestamp.Timestamp, lat float64, lon float64, elev float64, val float64,
	ha bool, la bool) *pb.TelemetryDatum {

	return &pb.TelemetryDatum{Uuid: uuid, Description: desc, Unit: unit, Timestamp: tstamp,
		Latitude: lat, Longitude: lon, Elevation: elev, Value: val, HighAlarm: ha, LowAlarm: la}

}
