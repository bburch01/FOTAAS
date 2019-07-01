package simulation

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/app/simulation/models"
	"github.com/bburch01/FOTAAS/internal/app/telemetry"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type SimResult struct {
	UUID   string
	Status api.ServerStatus
}

var logger *zap.Logger

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

}

func StartSimulation(simData map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData, sim models.Simulation,
	wg *sync.WaitGroup, resultsChan chan SimResult) {

	var startTime time.Time
	var elapsedTime time.Duration
	var tdata api.TelemetryData
	var sampleRateInMillis int32
	var simRateMultiplier int32
	var resp *api.TransmitTelemetryResponse
	var req api.TransmitTelemetryRequest
	var sb strings.Builder

	defer wg.Done()

	switch sim.SampleRate {
	case api.SampleRate_SR_1_MS.String():
		sampleRateInMillis = 1
	case api.SampleRate_SR_10_MS.String():
		sampleRateInMillis = 10
	case api.SampleRate_SR_100_MS.String():
		sampleRateInMillis = 100
	case api.SampleRate_SR_1000_MS.String():
		sampleRateInMillis = 1000
	default:
		resultsChan <- SimResult{UUID: sim.Uuid, Status: api.ServerStatus{Code: api.StatusCode_ERROR, Message: "invalid sample rate in millis"}}
		return
	}

	switch sim.SimulationRateMultiplier {
	case api.SimulationRateMultiplier_X1.String():
		simRateMultiplier = 1
	case api.SimulationRateMultiplier_X2.String():
		simRateMultiplier = 2
	case api.SimulationRateMultiplier_X4.String():
		simRateMultiplier = 4
	case api.SimulationRateMultiplier_X8.String():
		simRateMultiplier = 8
	case api.SimulationRateMultiplier_X10.String():
		simRateMultiplier = 10
	case api.SimulationRateMultiplier_X20.String():
		simRateMultiplier = 20
	default:
		resultsChan <- SimResult{UUID: sim.Uuid, Status: api.ServerStatus{Code: api.StatusCode_ERROR, Message: "invalid simulation rate multiplier"}}
		return
	}

	sampleRateInMillis = sampleRateInMillis / simRateMultiplier

	datumCount := len(simData[api.TelemetryDatumDescription_BRAKE_TEMP_FL].Data)

	startTime = time.Now()

	tdata.GrandPrix = sim.GrandPrix
	tdata.Track = sim.Track
	//tdata.Constructor = sim.Constructor
	//tdata.CarNumber = sim.CarNumber

	var transmissionCount int

	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_PORT"))
	telemetrySvcEndpoint := sb.String()

	conn, err := grpc.Dial(telemetrySvcEndpoint, grpc.WithInsecure())
	if err != nil {
		msg := fmt.Sprintf("simulation service error: %v", err)
		resultsChan <- SimResult{UUID: sim.Uuid, Status: api.ServerStatus{Code: api.StatusCode_ERROR, Message: msg}}
		return
	}
	defer conn.Close()

	// TODO: determine what is the appropriate deadline for transmit requests, possibly scaling
	// based on the datum count.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client api.TelemetryServiceClient

	client = api.NewTelemetryServiceClient(conn)

	for idx := 0; idx < datumCount; idx++ {

		datumMap := make(map[string]*api.TelemetryDatum, len(simData))

		for _, v := range simData {

			v.Data[idx].SimulationTransmitSequenceNumber = int32(idx)
			datumMap[v.Data[idx].Uuid] = &v.Data[idx]

		}

		tdata.TelemetryDatumMap = datumMap

		req.TelemetryData = &tdata

		resp, err = client.TransmitTelemetry(ctx, &req)
		if err != nil {
			msg := fmt.Sprintf("simulation service error: %v", err)
			resultsChan <- SimResult{UUID: sim.Uuid, Status: api.ServerStatus{Code: api.StatusCode_ERROR, Message: msg}}
			return
		}

		for i, v := range resp.ServerStatus {
			if v.Code != api.StatusCode_OK {
				msg := fmt.Sprintf("transmit of telemetry datum UUID %v failed with telemetry service error: %v", i, v.Message)
				resultsChan <- SimResult{UUID: sim.Uuid, Status: api.ServerStatus{Code: api.StatusCode_ERROR, Message: msg}}
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
