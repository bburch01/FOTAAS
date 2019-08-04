package main

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.PathPrefix("/assets/images/").Handler(http.StripPrefix("/assets/images/", http.FileServer(http.Dir("./assets/images/"))))
	r.PathPrefix("/assets/css/").Handler(http.StripPrefix("/assets/css/", http.FileServer(http.Dir("./assets/css/"))))

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
	t := template.Must(template.ParseFiles("./templates/about.html"))
	t.Execute(w, nil)
}

func alivenessHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./templates/aliveness.html"))
	t.Execute(w, nil)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./templates/status.html"))
	t.Execute(w, nil)
}

func simulationHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./templates/simulation.html"))
	t.Execute(w, nil)
}

func analysisHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./templates/analysis.html"))
	t.Execute(w, nil)
}

func telemetryHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./templates/telemetry.html"))
	t.Execute(w, nil)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./templates/404space.html")
	//t := template.Must(template.ParseFiles("./templates/404space.html"))
	//t.Execute(w, nil)
}
