package simulation

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	pb "github.com/bburch01/FOTAAS/api"
	tel "github.com/bburch01/FOTAAS/internal/app/telemetry"
	logging "github.com/bburch01/FOTAAS/internal/pkg/logging"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type SimResult struct {
	UUID   string
	Status pb.ServerStatus
}

var logger *zap.Logger

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

}

// Sequential version:
/*
func StartSimulation(simData map[pb.TelemetryDatumDescription]tel.SimulatedTelemetryData, sim pb.Simulation) error {

	//var tstamp time.Time
	//var err error
	var startTime time.Time
	var elapsedTime time.Duration
	var tdata pb.TelemetryData
	var sampleRateInMillis int32
	var simRateMultiplier int32
	var svcaddr string
	var resp *pb.TransmitTelemetryResponse
	var req pb.TransmitTelemetryRequest

	switch sim.SampleRate {
	case pb.SampleRate_SR_1_MS:
		sampleRateInMillis = 1
	case pb.SampleRate_SR_10_MS:
		sampleRateInMillis = 10
	case pb.SampleRate_SR_100_MS:
		sampleRateInMillis = 100
	case pb.SampleRate_SR_1000_MS:
		sampleRateInMillis = 1000
	default:
		return errors.New("invalid sample rate in millis")
	}

	switch sim.SimulationRateMultiplier {
	case pb.SimulationRateMultiplier_X1:
		simRateMultiplier = 1
	case pb.SimulationRateMultiplier_X2:
		simRateMultiplier = 2
	case pb.SimulationRateMultiplier_X4:
		simRateMultiplier = 4
	case pb.SimulationRateMultiplier_X8:
		simRateMultiplier = 8
	case pb.SimulationRateMultiplier_X10:
		simRateMultiplier = 10
	case pb.SimulationRateMultiplier_X20:
		simRateMultiplier = 20
	default:
		return errors.New("invalid sample rate in millis")
	}

	sampleRateInMillis = sampleRateInMillis / simRateMultiplier

	datumCount := len(simData[pb.TelemetryDatumDescription_BRAKE_TEMP_FL].Data)

	//start debug logger block
		for _, v1 := range simData {

			// For now, log just the first 20 datum
			ldc := 0
			logger.Debug(fmt.Sprintf("datum count: %v", len(v1.Data)))
			for _, v2 := range v1.Data {
				if tstamp, err = ts.Timestamp(v2.Timestamp); err != nil {
					logger.Error(fmt.Sprintf("failed to convert google.protobuf.timestamp to time.Time with error: %v", err))
				}
				logger.Debug(fmt.Sprintf("datum desc: %v unit: %v timestamp: %v value: %v", v2.Description.String(),
					v2.Unit.String(), tstamp, v2.Value))

				ldc++
				if ldc == 20 {
					break
				}
			}
		}
	//end debug logger block

	startTime = time.Now()

	tdata.GrandPrix = sim.GrandPrix
	tdata.Track = sim.Track
	tdata.Constructor = sim.Constructor
	tdata.CarNumber = sim.CarNumber

	var transmissionCount int

	svcaddr = os.Getenv("TELEMETRY_SERVICE_HOST") + os.Getenv("TELEMETRY_SERVICE_PORT")

	conn, err := grpc.Dial(svcaddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)

	defer cancel()

	var client pb.TelemetryServiceClient

	client = pb.NewTelemetryServiceClient(conn)

	for idx := 0; idx < datumCount; idx++ {

		datumMap := make(map[string]*pb.TelemetryDatum, len(simData))

		for _, v := range simData {

			v.Data[idx].SimulationTransmitSequenceNumber = int32(idx)
			datumMap[v.Data[idx].Uuid] = &v.Data[idx]

		}

		tdata.TelemetryDatumMap = datumMap


		//start debug logger block
			logger.Debug("")
			logger.Debug(fmt.Sprintf("TRANSMISSION START GrandPrix: %v Track: %v Constructor: %v CarNumber: %v",
				tdata.GrandPrix, tdata.Track, tdata.Constructor, tdata.CarNumber))
			for _, v := range datumMap {

				logger.Debug(fmt.Sprintf("datum description: %v simulation: %v simulation id: %v sim tx seq num: %v datum value: %v",
					v.Description, v.Simulated, v.SimulationUuid, v.SimulationTransmitSequenceNumber, v.Value))

			}
			logger.Debug("TRANSMISSION STOP")
			logger.Debug("")

			transmissionCount++
		//end debug logger block

		// Next block is the actual telemetry service transmit request

		req.TelemetryData = &tdata

		resp, err = client.TransmitTelemetry(ctx, &req)
		if err != nil {
			return err
		}

		for i, v := range resp.ServerStatus {
			if v.Code != pb.StatusCode_OK {
				return fmt.Errorf("transmit of telemetry datum UUID %v failed with telemetry service error: %v", i, v.Message)
			}
		}

		transmissionCount++

		time.Sleep(time.Duration(sampleRateInMillis) * time.Millisecond)

	}

	logger.Debug(fmt.Sprintf("transmissionCount: %v", transmissionCount))

	elapsedTime = time.Since(startTime)
	logger.Debug(fmt.Sprintf("simulation execution time: %v", elapsedTime))

	return nil
}
*/

// Concurrent version:
func StartSimulation(simData map[pb.TelemetryDatumDescription]tel.SimulatedTelemetryData, sim pb.Simulation,
	wg *sync.WaitGroup, resultsChan chan SimResult) {

	//var tstamp time.Time
	//var err error
	var startTime time.Time
	var elapsedTime time.Duration
	var tdata pb.TelemetryData
	var sampleRateInMillis int32
	var simRateMultiplier int32
	var svcaddr string
	var resp *pb.TransmitTelemetryResponse
	var req pb.TransmitTelemetryRequest

	defer wg.Done()

	switch sim.SampleRate {
	case pb.SampleRate_SR_1_MS:
		sampleRateInMillis = 1
	case pb.SampleRate_SR_10_MS:
		sampleRateInMillis = 10
	case pb.SampleRate_SR_100_MS:
		sampleRateInMillis = 100
	case pb.SampleRate_SR_1000_MS:
		sampleRateInMillis = 1000
	default:
		resultsChan <- SimResult{UUID: sim.Uuid, Status: pb.ServerStatus{Code: pb.StatusCode_ERROR, Message: "invalid sample rate in millis"}}
		return
	}

	switch sim.SimulationRateMultiplier {
	case pb.SimulationRateMultiplier_X1:
		simRateMultiplier = 1
	case pb.SimulationRateMultiplier_X2:
		simRateMultiplier = 2
	case pb.SimulationRateMultiplier_X4:
		simRateMultiplier = 4
	case pb.SimulationRateMultiplier_X8:
		simRateMultiplier = 8
	case pb.SimulationRateMultiplier_X10:
		simRateMultiplier = 10
	case pb.SimulationRateMultiplier_X20:
		simRateMultiplier = 20
	default:
		resultsChan <- SimResult{UUID: sim.Uuid, Status: pb.ServerStatus{Code: pb.StatusCode_ERROR, Message: "invalid simulation rate multiplier"}}
		return
	}

	sampleRateInMillis = sampleRateInMillis / simRateMultiplier

	datumCount := len(simData[pb.TelemetryDatumDescription_BRAKE_TEMP_FL].Data)

	/*
		for _, v1 := range simData {

			// For now, log just the first 20 datum
			ldc := 0
			logger.Debug(fmt.Sprintf("datum count: %v", len(v1.Data)))
			for _, v2 := range v1.Data {
				if tstamp, err = ts.Timestamp(v2.Timestamp); err != nil {
					logger.Error(fmt.Sprintf("failed to convert google.protobuf.timestamp to time.Time with error: %v", err))
				}
				logger.Debug(fmt.Sprintf("datum desc: %v unit: %v timestamp: %v value: %v", v2.Description.String(),
					v2.Unit.String(), tstamp, v2.Value))

				ldc++
				if ldc == 20 {
					break
				}
			}
		}
	*/

	startTime = time.Now()

	tdata.GrandPrix = sim.GrandPrix
	tdata.Track = sim.Track
	tdata.Constructor = sim.Constructor
	tdata.CarNumber = sim.CarNumber

	var transmissionCount int

	svcaddr = os.Getenv("TELEMETRY_SERVICE_HOST") + os.Getenv("TELEMETRY_SERVICE_PORT")

	conn, err := grpc.Dial(svcaddr, grpc.WithInsecure())
	if err != nil {
		msg := fmt.Sprintf("simulation service error: %v", err)
		resultsChan <- SimResult{UUID: sim.Uuid, Status: pb.ServerStatus{Code: pb.StatusCode_ERROR, Message: msg}}
		return
	}
	defer conn.Close()

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)

	// TODO: determine what is the appropriate deadline for transmit requests, possibly scaling
	// based on the datum count.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client pb.TelemetryServiceClient

	client = pb.NewTelemetryServiceClient(conn)

	for idx := 0; idx < datumCount; idx++ {

		datumMap := make(map[string]*pb.TelemetryDatum, len(simData))

		for _, v := range simData {

			v.Data[idx].SimulationTransmitSequenceNumber = int32(idx)
			datumMap[v.Data[idx].Uuid] = &v.Data[idx]

		}

		tdata.TelemetryDatumMap = datumMap

		/*
			logger.Debug("")
			logger.Debug(fmt.Sprintf("TRANSMISSION START GrandPrix: %v Track: %v Constructor: %v CarNumber: %v",
				tdata.GrandPrix, tdata.Track, tdata.Constructor, tdata.CarNumber))
			for _, v := range datumMap {

				logger.Debug(fmt.Sprintf("datum description: %v simulation: %v simulation id: %v sim tx seq num: %v datum value: %v",
					v.Description, v.Simulated, v.SimulationUuid, v.SimulationTransmitSequenceNumber, v.Value))

			}
			logger.Debug("TRANSMISSION STOP")
			logger.Debug("")

			transmissionCount++
		*/

		// Next block is the actual telemetry service transmit request

		req.TelemetryData = &tdata

		resp, err = client.TransmitTelemetry(ctx, &req)
		if err != nil {
			msg := fmt.Sprintf("simulation service error: %v", err)
			resultsChan <- SimResult{UUID: sim.Uuid, Status: pb.ServerStatus{Code: pb.StatusCode_ERROR, Message: msg}}
			return
		}

		for i, v := range resp.ServerStatus {
			if v.Code != pb.StatusCode_OK {
				msg := fmt.Sprintf("transmit of telemetry datum UUID %v failed with telemetry service error: %v", i, v.Message)
				resultsChan <- SimResult{UUID: sim.Uuid, Status: pb.ServerStatus{Code: pb.StatusCode_ERROR, Message: msg}}
				return
			}
		}

		transmissionCount++

		time.Sleep(time.Duration(sampleRateInMillis) * time.Millisecond)

	}

	logger.Debug(fmt.Sprintf("transmissionCount: %v", transmissionCount))

	elapsedTime = time.Since(startTime)
	logger.Debug(fmt.Sprintf("simulation execution time: %v", elapsedTime))

	return
}
