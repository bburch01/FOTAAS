//go:generate go run generated/assets/assets_generate.go
//go:generate go run generated/templates/templates_generate.go

package main

import (
	"context"
	"encoding/gob"

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
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var logger *zap.Logger

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type msg struct {
	Num int
}

type statusTest struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Result string `json:"result"`
}

type statusTestSequence struct {
	IsComplete  string       `json:"isComplete"`
	StatusTests []statusTest `json:"statusTests"`
}

type User struct {
	Username      string
	Authenticated bool
}

var store *sessions.CookieStore

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

	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

	store = sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	gob.Register(User{})

}

func main() {

	r := mux.NewRouter()

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(assets.Assets)))

	r.HandleFunc("/", aboutHandler).Methods("GET")
	//r.HandleFunc("/", indexHandler).Methods("GET")
	//r.HandleFunc("/login", loginHandler).Methods("POST")
	//r.HandleFunc("/logout", logoutHandler).Methods("GET")
	//r.HandleFunc("/forbidden", forbiddenHandler).Methods("GET")
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

func getUser(s *sessions.Session) User {

	val := s.Values["user"]
	var user = User{}
	user, ok := val.(User)
	if !ok {
		return User{Authenticated: false}
	}
	return user

}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "cookie-name")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.FormValue("code") != "code" {
		if r.FormValue("code") == "" {
			session.AddFlash("Must enter a code")
		}
		session.AddFlash("The code was incorrect")
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/forbidden", http.StatusFound)
		return
	}

	username := r.FormValue("username")

	user := &User{
		Username:      username,
		Authenticated: true,
	}

	session.Values["user"] = user

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/about", http.StatusFound)

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "cookie-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["user"] = User{}
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)

}

func forbiddenHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "cookie-name")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashMessages := session.Flashes()
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, err := templates.Templates.Open("/forbidden.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("forbidden").Parse(string(templateBytes))
	t.Execute(w, flashMessages)

}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "cookie-name")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := getUser(session)

	file, err := templates.Templates.Open("/index.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("index").Parse(string(templateBytes))
	t.Execute(w, user)

}

func echoHandler(w http.ResponseWriter, r *http.Request) {

	/*
		session, err := store.Get(r, "cookie-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUser(session)

		if auth := user.Authenticated; !auth {
			session.AddFlash("You don't have access!")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}
	*/

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

	logger.Debug("made it into echoWebSocketHandler()...")

	/*
		session, err := store.Get(r, "cookie-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUser(session)

		if auth := user.Authenticated; !auth {
			session.AddFlash("You don't have access!")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}
	*/

	var sb strings.Builder
	sb.WriteString("Origin header: ")
	sb.WriteString(r.Header.Get("Origin"))
	sb.WriteString(" not allowed.")
	sb.WriteString(" Host was: ")
	sb.WriteString(r.Host)

	if r.Header.Get("Origin") != "https://"+r.Host {
		logger.Error("http error, origin not allowed, 403.")
		http.Error(w, sb.String(), 403)
		return
	}

	//conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		logger.Error("Could not open websocket connection.")
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	go echo(conn)

}

func echo(conn *websocket.Conn) {

	logger.Debug("made it into echo()...")

	for {

		m := msg{}

		err := conn.ReadJSON(&m)
		if err != nil {
			logger.Error(fmt.Sprintf("Error reading json.: %v", err))
			return
			//fmt.Println("Error reading json.", err)
		}

		logger.Debug(fmt.Sprintf("Message from browser: %#v\n", m))
		//fmt.Printf("Got message: %#v\n", m)

		// echo back the message to the browser ???
		if err = conn.WriteJSON(m); err != nil {
			logger.Error(fmt.Sprintf("Error writing json: %v", err))
			return
			//fmt.Println(err)
		}

		logger.Debug("read from websocket connection completed...")

	}

	/*
		st := make([]statusTest, 7, 7)

		st[0] = statusTest{Name: "Telemetry Service Aliveness", State: "Complete", Result: "PASS"}
		st[1] = statusTest{Name: "Analysis Service Aliveness", State: "Complete", Result: "FAIL"}
		st[2] = statusTest{Name: "Simulation Service Aliveness", State: "Complete", Result: "PASS"}
		st[3] = statusTest{Name: "Start Simulation", State: "Complete", Result: "PASS"}
		st[4] = statusTest{Name: "Poll For Simulation Complete", State: "Complete", Result: "FAIL"}
		st[5] = statusTest{Name: "Retrieve Simulation Data", State: "Complete", Result: "PASS"}
		st[6] = statusTest{Name: "Simulation Data Analysis", State: "In Progress", Result: "UNKNOWN"}

		sts := statusTestSequence{IsComplete: "false", StatusTests: st}

		for {

			m := msg{}

			err := conn.ReadJSON(&m)
			if err != nil {
				//fmt.Println("Error reading json.", err)
				if strings.Contains(err.Error(), "close 1001") {
					logger.Debug("got close on websocket...")
					//fmt.Print("got close on websocket")
					return
				}
			}

			logger.Debug(fmt.Sprintf("Got message: %#v\n", m))
			//fmt.Printf("Got message: %#v\n", m)

			stsJSON, err := json.Marshal(sts)
			if err != nil {
				logger.Error(fmt.Sprintf("json marshal error: %v", err))
				//fmt.Printf("Error: %s", err)
				return
			}

			logger.Error(fmt.Sprintf("stsJSON: %v", string(stsJSON)))
			//fmt.Printf("stsJSON: %v", string(stsJSON))

			if err = conn.WriteJSON(sts); err != nil {
				logger.Error(fmt.Sprintf("json write error: %v", err))
				//fmt.Println(err)
			}

		}
	*/
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

	/*
		session, err := store.Get(r, "cookie-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUser(session)

		if auth := user.Authenticated; !auth {
			session.AddFlash("You don't have access!")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}
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

	/*
		session, err := store.Get(r, "cookie-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUser(session)

		if auth := user.Authenticated; !auth {
			session.AddFlash("You don't have access!")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}
	*/

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

	/*
		session, err := store.Get(r, "cookie-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUser(session)

		if auth := user.Authenticated; !auth {
			session.AddFlash("You don't have access!")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}
	*/

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

	/*
		session, err := store.Get(r, "cookie-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUser(session)

		if auth := user.Authenticated; !auth {
			session.AddFlash("You don't have access!")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}
	*/

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

	/*
		session, err := store.Get(r, "cookie-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUser(session)

		if auth := user.Authenticated; !auth {
			session.AddFlash("You don't have access!")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}
	*/

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

	/*
		session, err := store.Get(r, "cookie-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUser(session)

		if auth := user.Authenticated; !auth {
			session.AddFlash("You don't have access!")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}
	*/

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
