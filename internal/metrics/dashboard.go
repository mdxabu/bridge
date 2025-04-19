package metrics

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/mdxabu/bridge/internal/gateway/forwarder"
	"github.com/mdxabu/bridge/internal/logger"
)

type PingData struct {
	Source      string  `json:"source"`
	Destination string  `json:"destination"`
	Sent        int     `json:"sent"`
	Received    int     `json:"received"`
	PacketLoss  float64 `json:"packet_loss"`
	RTT         int64   `json:"rtt_ms"`
	Timestamp   int64   `json:"timestamp"`
}

var (
	pingResults []PingData
	lock        sync.Mutex
)

func recordResult(data forwarder.PingData) {
	lock.Lock()
	defer lock.Unlock()

	pingResults = append(pingResults, PingData{
		Source:      data.Source,
		Destination: data.Destination,
		Sent:        data.Sent,
		Received:    data.Received,
		PacketLoss:  data.PacketLoss,
		RTT:         data.RTT, // already int64
		Timestamp:   time.Now().Unix(),
	})
}

func StartWebDashboard(nat64 bool) {
	go forwarder.StartWithCallback(recordResult)

	http.HandleFunc("/", serveLoginPage)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/dashboard", requireAuth(serveDashboard))
	http.HandleFunc("/api/data", requireAuth(serveData))

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	logger.Info("Dashboard running at http://localhost:8080")
	logger.Info("Username: admin, Password: admin")
	logger.Info("Press Ctrl+C to stop the server")
	http.ListenAndServe(":8080", nil)
}

// Simple in-memory session
var session = struct {
	sync.Mutex
	loggedIn bool
}{}

func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("web/login.html"))
	tmpl.Execute(w, nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "admin" && password == "admin" {
		session.Lock()
		session.loggedIn = true
		session.Unlock()
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("web/dashboard.html"))
	tmpl.Execute(w, nil)
}

func serveData(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pingResults)
}

func requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session.Lock()
		defer session.Unlock()
		if !session.loggedIn {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		handler(w, r)
	}
}
