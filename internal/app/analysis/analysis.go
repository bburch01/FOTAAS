package analysis

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/bburch01/FOTAAS/api"
	"google.golang.org/grpc"
)

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

		for _, v := range telemetryData.TelemetryDatumMap {

		}

	case api.ResponseCode_ERROR:
	default:
	}

	return data, nil

}
