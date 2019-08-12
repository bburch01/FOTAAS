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
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var logger *zap.Logger

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type msg struct {
	Num int
}

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
	r.HandleFunc("/echo", echoHandler).Methods("GET")
	r.HandleFunc("/echo_ws", echoWebSocketHandler).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	http.ListenAndServe(":8080", r)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {

	file, err := templates.Templates.Open("/echo.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("echo").Parse(string(templateBytes))
	t.Execute(w, nil)

}

func echoWebSocketHandler(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	go echo(conn)

}

func echo(conn *websocket.Conn) {
	for {
		m := msg{}

		err := conn.ReadJSON(&m)
		if err != nil {
			//fmt.Println("Error reading json.", err)
			if strings.Contains(err.Error(), "close 1001") {
				fmt.Print("got close on websocket")
				return
			}
		}

		fmt.Printf("Got message: %#v\n", m)

		if err = conn.WriteJSON(m); err != nil {
			fmt.Println(err)
		}
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {

	/*
			type Person struct {
				UserName string
			}

			p := Person{UserName: "Barry"}

			now use p in t.Execute(w, p)

			p is then available in about.html e.g.:

		        <div>
		          <h1>About Page</h1>
		          <h1>Hello {{.UserName}}!</h1>
				</div>

	*/

	file, err := templates.Templates.Open("/about.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("about").Parse(string(templateBytes))
	t.Execute(w, nil)
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
