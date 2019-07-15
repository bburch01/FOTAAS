package simulation

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/app/simulation/data"
	"github.com/bburch01/FOTAAS/internal/app/simulation/models"
	"github.com/bburch01/FOTAAS/internal/app/telemetry"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

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

	//var startTime time.Time
	//var elapsedTime time.Duration

	sim.State = "INITIALIZING"
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
		go data.GenerateSimulatedTelemetryData(*sim, v, &wg, resultsChan, errChan)
	}
	wg.Wait()
	close(resultsChan)
	close(errChan)

	// Check the errChan, on the first error, set sim.FinalStatusCode & sim.FinalStatusMessage
	// to the error info, attempt to persist the simulation and bail-out.
	if err := <-errChan; err != nil {
		logger.Error(fmt.Sprintf("simulation %v failed to start with error: %v", sim.ID, err))
		sim.State = "FAILED_TO_START"
		if err := sim.UpdateState(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = "simulation failed to start with a server-side error"
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	var simMemberDataMap = make(map[string]map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData)

	// Retrieve and aggregate the simulation data (for all of the sim members) from the results channel
	for v := range resultsChan {
		simMemberDataMap[v.SimMemberID] = v.SimData
		logger.Debug(fmt.Sprintf("gen data result sim member: %v datum count: %v", v.SimMemberID, len(v.SimData)))
	}

	var sampleRateInMillis int32

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
		sim.State = "FAILED_TO_START"
		if err := sim.UpdateState(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = fmt.Sprintf("invalid sample rate")
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	var simRateMultiplier int32

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
		sim.State = "FAILED_TO_START"
		if err := sim.UpdateState(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = fmt.Sprintf("invalid simulation rate multiplier")
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	sleepDuration := time.Duration(sampleRateInMillis/simRateMultiplier) * time.Millisecond
	datumCount := (sim.DurationInMinutes * 60000) / sampleRateInMillis
	percentComplete := float32(0.0)

	percentCompleteIncrement := float32(float32(100.0) / float32(datumCount))
	percentCompleteIncrement = float32(math.Floor(float64(percentCompleteIncrement*100)) / 100)

	var sb strings.Builder

	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_PORT"))
	telemetrySvcEndpoint := sb.String()

	conn, err := grpc.Dial(telemetrySvcEndpoint, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		logger.Error(fmt.Sprintf("simulation %v failed to start with error: %v", sim.ID, err))
		sim.State = "FAILED_TO_START"
		if err := sim.UpdateState(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = "simulation failed to start with a server-side error"
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	// TODO: determine what is the appropriate deadline for transmit requests, possibly scaling
	// based on the datum count.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	defer cancel()

	client := api.NewTelemetryServiceClient(conn)

	sim.StartTimestamp, err = ipbts.TimestampProto(time.Now())
	if err != nil {
		logger.Error(fmt.Sprintf("simulation %v failed to start with error: %v", sim.ID, err))
		sim.State = "FAILED_TO_START"
		if err := sim.UpdateState(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = "simulation failed to start with a server-side error"
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	if err := sim.UpdateStartTimestamp(); err != nil {
		logger.Error(fmt.Sprintf("simulation %v failed to start with error: %v", sim.ID, err))
		sim.State = "FAILED_TO_START"
		if err := sim.UpdateState(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = "simulation failed to start with a server-side error"
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	sim.State = "IN_PROGRESS"
	if err := sim.UpdateState(); err != nil {
		logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = fmt.Sprintf("simulation failed to start with server-side error: %v", err)
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	sim.PercentComplete = percentComplete
	if err := sim.UpdatePercentComplete(); err != nil {
		logger.Error(fmt.Sprintf("simulation %v failed to start with error: %v", sim.ID, err))
		sim.State = "FAILED_TO_START"
		if err := sim.UpdateState(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = "simulation failed to start with a server-side error"
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	var resp *api.TransmitTelemetryResponse
	var req api.TransmitTelemetryRequest
	var transmissionCount int

	// Main simulation loop
	for idx := int32(0); idx < datumCount; idx++ {
		for _, v := range sim.SimulationMembers {

			tdata := api.TelemetryData{}

			simMemberData := simMemberDataMap[v.ID]
			datumMap := make(map[string]*api.TelemetryDatum, len(simMemberData))

			for _, v2 := range simMemberData {
				v2.Data[idx].SimulationTransmitSequenceNumber = idx
				datumMap[v2.Data[idx].Uuid] = &v2.Data[idx]
			}

			tdata.TelemetryDatumMap = datumMap
			req.TelemetryData = &tdata

			resp, err = client.TransmitTelemetry(ctx, &req)
			if err != nil {
				logger.Error(fmt.Sprintf("simulation %v failed with error: %v", sim.ID, err))
				sim.State = "FAILED"
				if err := sim.UpdateState(); err != nil {
					logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
				}
				sim.FinalStatusCode = "ERROR"
				if err := sim.UpdateFinalStatusCode(); err != nil {
					logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
				}
				sim.FinalStatusMessage = "simulation failed to start with a server-side error"
				if err := sim.UpdateFinalStatusMessage(); err != nil {
					logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
				}
				return
			}

			for _, v := range resp.Details {
				if v.Code != api.ResponseCode_OK {
					logger.Error(fmt.Sprintf("simulation %v failed with telemetry service code: %v", sim.ID, v.Code))
					logger.Error(fmt.Sprintf("simulation %v failed with telemetry service message: %v", sim.ID, v.Message))
					sim.State = "FAILED"
					if err := sim.UpdateState(); err != nil {
						logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
					}
					sim.FinalStatusCode = "ERROR"
					if err := sim.UpdateFinalStatusCode(); err != nil {
						logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
					}
					sim.FinalStatusMessage = "simulation failed to start with a server-side error"
					if err := sim.UpdateFinalStatusMessage(); err != nil {
						logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
					}
					return
				}
			}
		}

		transmissionCount++
		time.Sleep(sleepDuration)
		percentComplete += percentCompleteIncrement
		percentComplete = float32(math.Floor(float64(percentComplete*100)) / 100)

		sim.PercentComplete = percentComplete
		if err := sim.UpdatePercentComplete(); err != nil {
			logger.Error(fmt.Sprintf("simulation %v failed with error: %v", sim.ID, err))
			sim.State = "FAILED"
			if err := sim.UpdateState(); err != nil {
				logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
			}
			sim.FinalStatusCode = "ERROR"
			if err := sim.UpdateFinalStatusCode(); err != nil {
				logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
			}
			sim.FinalStatusMessage = "simulation failed to start with a server-side error"
			if err := sim.UpdateFinalStatusMessage(); err != nil {
				logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
			}
			return
		}

	}

	sim.EndTimestamp, err = ipbts.TimestampProto(time.Now())
	if err != nil {
		logger.Error(fmt.Sprintf("simulation %v failed with error: %v", sim.ID, err))
		sim.State = "FAILED"
		if err := sim.UpdateState(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = "simulation failed to start with a server-side error"
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	if err := sim.UpdateEndTimestamp(); err != nil {
		logger.Error(fmt.Sprintf("simulation %v failed with error: %v", sim.ID, err))
		sim.State = "FAILED"
		if err := sim.UpdateState(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = "simulation failed to start with a server-side error"
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	sim.State = "COMPLETED"
	if err := sim.UpdateState(); err != nil {
		logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		return
	}

	sim.PercentComplete = 100.0
	if err := sim.UpdatePercentComplete(); err != nil {
		logger.Error(fmt.Sprintf("simulation %v failed with error: %v", sim.ID, err))
		sim.State = "FAILED"
		if err := sim.UpdateState(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusCode = "ERROR"
		if err := sim.UpdateFinalStatusCode(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		sim.FinalStatusMessage = "simulation failed to start with a server-side error"
		if err := sim.UpdateFinalStatusMessage(); err != nil {
			logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		}
		return
	}

	sim.FinalStatusCode = "OK"
	if err := sim.UpdateFinalStatusCode(); err != nil {
		logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		return
	}

	sim.FinalStatusMessage = "simulation completed normally"
	if err := sim.UpdateFinalStatusMessage(); err != nil {
		logger.Error(fmt.Sprintf("failed to update simulation %v with error: %v", sim.ID, err))
		return
	}
	logger.Debug(fmt.Sprintf("transmissionCount: %v", transmissionCount))

	//elapsedTime = time.Since(startTime)
	//logger.Debug(fmt.Sprintf("simulation execution time: %v", elapsedTime))

	return
}
