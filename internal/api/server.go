package api
import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/mdxabu/bridge/internal/nat"
)

// Server represents the API server
type Server struct {
	bridge      BridgeInterface
	addr        string
	server      *http.Server
	mu          sync.RWMutex
	startTime   time.Time
	isRunning   bool
}

// BridgeInterface defines the interface for bridge operations
type BridgeInterface interface {
	GetStats() map[string]interface{}
	GetActiveSessions() []*nat.SessionState
}

// NewServer creates a new API server
func NewServer(addr string, bridge BridgeInterface) *Server {
	return &Server{
		bridge:    bridge,
		addr:      addr,
		startTime: time.Now(),
	}
}

// Start starts the API server
func (s *Server) Start() error {
	s.mu.Lock()
	s.isRunning = true
	s.mu.Unlock()

	mux := http.NewServeMux()
	
	// Register endpoints
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/health", s.handleHealth)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: s.enableCORS(mux),
	}

	return s.server.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop() error {
	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()

	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// handleStatus returns the bridge status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"status":     s.getStatusString(),
		"uptime":     time.Since(s.startTime).Seconds(),
		"start_time": s.startTime.Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(status)
}

// handleStats returns bridge statistics
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if s.bridge == nil {
		http.Error(w, "Bridge not initialized", http.StatusServiceUnavailable)
		return
	}

	stats := s.bridge.GetStats()
	stats["uptime"] = time.Since(s.startTime).Seconds()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleSessions returns active NAT sessions
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if s.bridge == nil {
		http.Error(w, "Bridge not initialized", http.StatusServiceUnavailable)
		return
	}

	sessions := s.bridge.GetActiveSessions()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// handleHealth returns health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	health := map[string]interface{}{
		"status":  "healthy",
		"running": s.isRunning,
		"uptime":  time.Since(s.startTime).Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// enableCORS enables CORS for the API
func (s *Server) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getStatusString returns a human-readable status string
func (s *Server) getStatusString() string {
	if s.isRunning {
		return "running"
	}
	return "stopped"
}
