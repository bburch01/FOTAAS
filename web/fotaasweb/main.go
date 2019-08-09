//go:generate go run generated/assets/assets_generate.go
//go:generate go run generated/templates/templates_generate.go

package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/pkg/logging"
	"github.com/bburch01/FOTAAS/web/fotaasweb/generated/assets"
	"github.com/bburch01/FOTAAS/web/fotaasweb/generated/templates"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	//"github.com/bburch01/FOTAAS/api"
	//"google.golang.org/grpc"
	"github.com/gorilla/mux"
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

func main() {

	r := mux.NewRouter()

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(assets.Assets)))

	r.HandleFunc("/", aboutHandler).Methods("GET")
	r.HandleFunc("/about", aboutHandler).Methods("GET")
	r.HandleFunc("/aliveness", alivenessHandler).Methods("GET")
	r.HandleFunc("/status", statusHandler).Methods("GET")
	r.HandleFunc("/simulation", simulationHandler).Methods("GET")
	r.HandleFunc("/analysis", analysisHandler).Methods("GET")
	r.HandleFunc("/telemetry", telemetryHandler).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	http.ListenAndServe(":8080", r)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {

	type Person struct {
		UserName string
	}

	p := Person{UserName: "Barry"}

	file, err := templates.Templates.Open("/about.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("about").Parse(string(templateBytes))
	t.Execute(w, p)
}

func alivenessHandler(w http.ResponseWriter, r *http.Request) {

	type AlivenessPageData struct {
		RespMap map[string]api.AlivenessCheckResponse
	}

	apd := AlivenessPageData{RespMap: make(map[string]api.AlivenessCheckResponse)}

	resp, err := checkByName("telemetry")
	if err != nil {
		resp = &api.AlivenessCheckResponse{Details: &api.ResponseDetails{
			Code: api.ResponseCode_ERROR, Message: fmt.Sprintf("aliveness check failed with error: %v", err)}}
	}
	apd.RespMap["telemetry"] = *resp

	resp, err = checkByName("analysis")
	if err != nil {
		resp = &api.AlivenessCheckResponse{Details: &api.ResponseDetails{
			Code: api.ResponseCode_ERROR, Message: fmt.Sprintf("aliveness check failed with error: %v", err)}}
	}
	apd.RespMap["analysis"] = *resp

	resp, err = checkByName("simulation")
	if err != nil {
		resp = &api.AlivenessCheckResponse{Details: &api.ResponseDetails{
			Code: api.ResponseCode_ERROR, Message: fmt.Sprintf("aliveness check failed with error: %v", err)}}
	}
	apd.RespMap["simulation"] = *resp

	resp, err = checkByName("status")
	if err != nil {
		resp = &api.AlivenessCheckResponse{Details: &api.ResponseDetails{
			Code: api.ResponseCode_ERROR, Message: fmt.Sprintf("aliveness check failed with error: %v", err)}}
	}
	apd.RespMap["status"] = *resp

	file, err := templates.Templates.Open("/aliveness.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("aliveness").Parse(string(templateBytes))
	t.Execute(w, apd)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	file, err := templates.Templates.Open("/status.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("status").Parse(string(templateBytes))
	t.Execute(w, nil)
}

func simulationHandler(w http.ResponseWriter, r *http.Request) {
	file, err := templates.Templates.Open("/simulation.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("simulation").Parse(string(templateBytes))
	t.Execute(w, nil)
}

func analysisHandler(w http.ResponseWriter, r *http.Request) {
	file, err := templates.Templates.Open("/analysis.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("analysis").Parse(string(templateBytes))
	t.Execute(w, nil)
}

func telemetryHandler(w http.ResponseWriter, r *http.Request) {
	file, err := templates.Templates.Open("/telemetry.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("telemetry").Parse(string(templateBytes))
	t.Execute(w, nil)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	file, err := templates.Templates.Open("/404space.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("404space").Parse(string(templateBytes))
	t.Execute(w, nil)
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
