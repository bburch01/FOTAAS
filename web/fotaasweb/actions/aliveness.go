package actions

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/gobuffalo/buffalo"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

/*
type SimMemberData struct {
	SimMemberID string
	SimData     map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
}
*/

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

// AlivenessHandler is a default handler to serve up
// the aliveness page.
func AlivenessHandler(c buffalo.Context) error {

	respMap := make(map[string]api.AlivenessCheckResponse)

	resp, err := checkByName("telemetry")
	if err != nil {
		return c.Render(200, r.HTML("404space.html"))
		//c.Flash().Add("error", fmt.Sprintf("an internal server error has occurred: %v", err))
		//return c.Redirect(500, "/")
	}
	respMap["telemetry"] = *resp

	resp, err = checkByName("analysis")
	if err != nil {
		return c.Render(200, r.HTML("404space.html"))
		//c.Flash().Add("error", fmt.Sprintf("an internal server error has occurred: %v", err))
		//return c.Redirect(500, "/")
	}
	respMap["telemetry"] = *resp

	resp, err = checkByName("simulation")
	if err != nil {
		return c.Render(200, r.HTML("404space.html"))
		//c.Flash().Add("error", fmt.Sprintf("an internal server error has occurred: %v", err))
		//return c.Redirect(500, "/")
	}
	respMap["telemetry"] = *resp

	resp, err = checkByName("status")
	if err != nil {
		return c.Render(200, r.HTML("404space.html"))
		//c.Flash().Add("error", fmt.Sprintf("an internal server error has occurred: %v", err))
		//return c.Redirect(500, "/")
	}
	respMap["telemetry"] = *resp

	return c.Render(200, r.HTML("aliveness.html"))
}

func checkByName(svcname string) (*api.AlivenessCheckResponse, error) {

	var svcEndpoint string
	var resp *api.AlivenessCheckResponse
	var sb strings.Builder

	switch svcname {
	case "telemetry":
		sb.WriteString(os.Getenv("TELEMETRY_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("TELEMETRY_SERVICE_PORT"))
		svcEndpoint = sb.String()
	case "analysis":
		sb.WriteString(os.Getenv("ANALYSIS_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("ANALYSIS_SERVICE_PORT"))
		svcEndpoint = sb.String()
	case "simulation":
		sb.WriteString(os.Getenv("SIMULATION_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("SIMULATION_SERVICE_PORT"))
		svcEndpoint = sb.String()
	case "status":
		sb.WriteString(os.Getenv("STATUS_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("STATUS_SERVICE_PORT"))
		svcEndpoint = sb.String()
	default:
		return resp, errors.New("invalid service name, valid service names are telemetry, analysis, simulation, status")
	}

	conn, err := grpc.Dial(svcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// TODO: determine what the appropriate deadline should be for this service call.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	switch svcname {
	case "telemetry":
		client := api.NewTelemetryServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			return nil, err
		}
	case "analysis":
		client := api.NewAnalysisServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			return nil, err
		}
	case "simulation":
		client := api.NewSimulationServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			return nil, err
		}
	case "status":
		client := api.NewSystemStatusServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid service name, valid service names are: telemetry, analysis, simulation, status")
	}

	return resp, nil
}
