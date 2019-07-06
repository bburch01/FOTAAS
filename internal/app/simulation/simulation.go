package simulation

import (
	//"context"
	//"fmt"

	"fmt"
	"log"
	"os"
	"sync"
	"time"

	//ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/app/simulation/data"
	"github.com/bburch01/FOTAAS/internal/app/simulation/models"
	"github.com/bburch01/FOTAAS/internal/app/telemetry"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
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

func StartSimulation(sim *models.Simulation) {

	var startTime time.Time
	//var elapsedTime time.Duration
	//var tdata api.TelemetryData
	var sampleRateInMillis int32
	//should be not needed now: var simDurationInMillis int32
	var simRateMultiplier int32
	//var resp *api.TransmitTelemetryResponse
	//var req api.TransmitTelemetryRequest
	//var sb strings.Builder

	var simMemberDataMap map[string]map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
	//var simData []data.SimMemberData

	if err := sim.Create(); err != nil {
		// Since no simulation status can be persisted, the only this to do is log
		// an error an bail-out.
		logger.Error(fmt.Sprintf("simulation %v failed to start with error: %v", sim.ID, err))
		return
	}

	// Generate simulated telemetry data for all simulation members in advance.
	var wg sync.WaitGroup
	errChan := make(chan error, len(sim.SimulationMembers))
	resultsChan := make(chan data.SimMemberData, len(sim.SimulationMembers))
	wg.Add(len(sim.SimulationMembers))
	for _, v := range sim.SimulationMembers {
		go data.GenerateSimulatedTelemetryData(sim, &v, &wg, resultsChan, errChan)
	}
	wg.Wait()
	close(resultsChan)
	close(errChan)

	// Check the errChan, on the first error, set sim.FinalStatusCode & sim.FinalStatusMessage
	// to the error info, attempt to persist it and bail-out.
	if err := <-errChan; err != nil {
		logger.Error(fmt.Sprintf("simulation %v failed to start with error: ", sim.ID, err))
		sim.State = "FAILED_TO_START"
		sim.FinalStatusCode = "ERROR"
		sim.FinalStatusMessage = fmt.Sprintf("simulation failed to start with error: %v", err)
		if err := sim.Update(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	/*
		type SimMemberData struct {
			SimMemberID string
			SimData     map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
		}
	*/

	// Retrieve and aggregate the simulation data (for all of the sim members) from the results channel

	for v := range resultsChan {
		simMemberDataMap[v.SimMemberID] = v.SimData
	}

	// Main simulation loop. All of the simulation members are part of the same simulation (i.e.
	// multiple simulations are not being run) so concurrency does not make sense here.

	switch sim.SampleRate {
	case api.SampleRate_SR_1_MS:
		sampleRateInMillis = 1
	case api.SampleRate_SR_10_MS:
		sampleRateInMillis = 10
	case api.SampleRate_SR_100_MS:
		sampleRateInMillis = 100
	case api.SampleRate_SR_1000_MS:
		sampleRateInMillis = 1000
	default:
		// This should never happen. Validation occurs in both the protobuf api
		// and in main.go RunSimulation()
		logger.Error(fmt.Sprintf("invalid sample rate for simulation: %v", sim.ID))
		return
	}

	switch sim.SimulationRateMultiplier {
	case api.SimulationRateMultiplier_X1:
		simRateMultiplier = 1
	case api.SimulationRateMultiplier_X2:
		simRateMultiplier = 2
	case api.SimulationRateMultiplier_X4:
		simRateMultiplier = 4
	case api.SimulationRateMultiplier_X8:
		simRateMultiplier = 8
	case api.SimulationRateMultiplier_X10:
		simRateMultiplier = 10
	case api.SimulationRateMultiplier_X20:
		simRateMultiplier = 20
	default:
		// This should never happen. Validation occurs in both the protobuf api
		// and in main.go RunSimulation()
		logger.Error(fmt.Sprintf("invalid simulation rate multiplier for simulation: %v", sim.ID))
		return
	}

	sleepDuration := time.Duration(sampleRateInMillis/simRateMultiplier) * time.Millisecond

	datumCount := (sim.DurationInMinutes * 60000) / sampleRateInMillis

	startTime = time.Now()

	logger.Debug(fmt.Sprintf("sleepDuration: %v datumCount: %v startTime: %v", sleepDuration, datumCount, startTime))

	/*
		tdata.GrandPrix = api.

		//tdata.GrandPrix = sim.GrandPrix
		tdata.Track = sim.Track

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
	*/

	return
}
