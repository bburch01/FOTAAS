package analysis

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type constructorCar struct {
	constructor api.Constructor
	carNumber   int32
}

type alarmCounts struct {
	highAlarmCount int32
	lowAlarmCount  int32
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

func ExtractAlarmAnalysisData(req *api.GetAlarmAnalysisRequest) (*api.AlarmAnalysisData, error) {

	data := new(api.AlarmAnalysisData)

	dataReq := new(api.GetTelemetryDataRequest)
	dataReq.SearchBy = new(api.GetTelemetryDataRequest_SearchBy)
	dataReq.Simulated = req.Simulated
	dataReq.DateRangeBegin = req.DateRangeBegin
	dataReq.DateRangeEnd = req.DateRangeEnd
	dataReq.SearchBy.DateRange = true
	dataReq.SearchBy.HighAlarm = true
	dataReq.SearchBy.LowAlarm = true

	var sb strings.Builder
	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("TELEMETRY_SERVICE_PORT"))
	telemetrySvcEndpoint := sb.String()

	conn, err := grpc.Dial(telemetrySvcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// TODO: determine what is the appropriate deadline for transmit requests, possibly scaling
	// based on the size of SimulationMap.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client = api.NewTelemetryServiceClient(conn)

	resp, err := client.GetTelemetryData(ctx, dataReq)
	if err != nil {
		return nil, err
	}

	switch resp.Details.Code {
	case api.ResponseCode_OK:

		telemetryData := resp.TelemetryData
		if len(telemetryData.TelemetryDatumMap) == 0 {
			// no errors & no telemetry data, caller needs to check for nil
			return nil, nil
		}

		ccac := make(map[constructorCar]alarmCounts)

		for _, v := range telemetryData.TelemetryDatumMap {
			ac := alarmCounts{}
			cc := constructorCar{constructor: v.Constructor, carNumber: v.CarNumber}
			if _, ok := ccac[cc]; !ok {
				if v.HighAlarm {
					ac.highAlarmCount++
				}
				if v.LowAlarm {
					ac.lowAlarmCount++
				}
				ccac[cc] = ac
			} else {
				ac = ccac[cc]
				if v.HighAlarm {
					ac.highAlarmCount++
				}
				if v.LowAlarm {
					ac.lowAlarmCount++
				}
				ccac[cc] = ac
			}
		}

		data.Simulated = req.Simulated
		data.DateRangeBegin = req.DateRangeBegin
		data.DateRangeEnd = req.DateRangeEnd

		for k, v := range ccac {
			ac := api.AlarmAnalysisData_AlarmCountsByConstructorAndCar{}
			ac.Constructor = k.constructor
			ac.CarNumber = k.carNumber
			ac.HighAlarmCount = v.highAlarmCount
			ac.LowAlarmCount = v.lowAlarmCount
			data.AlarmCounts = append(data.AlarmCounts, &ac)
		}

		return data, nil

	case api.ResponseCode_ERROR:
		return nil, fmt.Errorf("failed to retrieve telemetry data, response message from telemetry service was: %v", resp.Details.Message)
	default:
		return nil, fmt.Errorf("failed to retrieve telemetry data, invalid reponse code: %v", resp.Details.Code.String())
	}

}
