//go:generate go run generated/assets/assets_generate.go
//go:generate go run generated/templates/templates_generate.go

package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bburch01/FOTAAS/web/fotaasweb/generated/assets"
	"github.com/bburch01/FOTAAS/web/fotaasweb/generated/templates"
	"github.com/gorilla/mux"
)

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
	file, err := templates.Templates.Open("/aliveness.html")
	if err != nil {
		log.Panicf("failed to open template with error: %v", err)
	}
	defer file.Close()

	templateBytes, _ := ioutil.ReadAll(file)

	t, _ := template.New("aliveness").Parse(string(templateBytes))
	t.Execute(w, nil)
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
