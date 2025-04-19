package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/mdxabu/bridge/internal/gateway/forwarder"
	"github.com/mdxabu/bridge/internal/logger"
)

const (
	adminUsername = "admin"
	adminPassword = "admin"
)

type MetricData struct {
	Timestamp   int64   `json:"timestamp"`
	Source      string  `json:"source"`
	Destination string  `json:"destination"`
	Sent        int     `json:"sent"`
	Received    int     `json:"received"`
	Loss        float64 `json:"loss"`
	RTT         float64 `json:"rtt"`
}

var (
	metrics     []MetricData
	metricMutex sync.Mutex
	sessions    = make(map[string]bool)
	sessionLock sync.RWMutex

	ipPoolSize            int
	ipPoolUsed            int
	ipPoolMutex           sync.RWMutex
	poolExhaustedNotified bool
)

func AddMetric(source, destination string, sent, received int, loss, rtt float64) {
	metricMutex.Lock()
	defer metricMutex.Unlock()

	if len(metrics) > 100 {
		metrics = metrics[1:]
	}

	metrics = append(metrics, MetricData{
		Timestamp:   time.Now().Unix(),
		Source:      source,
		Destination: destination,
		Sent:        sent,
		Received:    received,
		Loss:        loss,
		RTT:         rtt,
	})
}

func SetIPPoolSize(size int) {
	ipPoolMutex.Lock()
	defer ipPoolMutex.Unlock()
	ipPoolSize = size
}

func UpdateIPPoolUsage(used int) {
	ipPoolMutex.Lock()
	defer ipPoolMutex.Unlock()

	ipPoolUsed = used

	if ipPoolUsed >= ipPoolSize && !poolExhaustedNotified {
		logger.Warn("IP POOL EXHAUSTED: All available IP addresses have been allocated!")
		fmt.Printf("    Used: %d/%d addresses\n\n", ipPoolUsed, ipPoolSize)
		poolExhaustedNotified = true
	} else if ipPoolUsed < ipPoolSize && poolExhaustedNotified {
		logger.Info("IP addresses available again in the pool")
		poolExhaustedNotified = false
	}
}

func GetIPPoolStatus() (int, int, float64) {
	ipPoolMutex.RLock()
	defer ipPoolMutex.RUnlock()

	if ipPoolSize == 0 {
		return 0, 0, 0
	}

	usagePercent := float64(ipPoolUsed) / float64(ipPoolSize) * 100
	return ipPoolUsed, ipPoolSize, usagePercent
}

type IPPoolStatusData struct {
	Used         int     `json:"used"`
	Total        int     `json:"total"`
	UsagePercent float64 `json:"usage_percent"`
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != adminUsername || pass != adminPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func generateSessionID() string {
	return time.Now().Format("20060102150405") + "-" +
		time.Now().Add(time.Millisecond).Format("20060102150405")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		username := r.Form.Get("username")
		password := r.Form.Get("password")

		if username == adminUsername && password == adminPassword {
			sessionID := generateSessionID()

			sessionLock.Lock()
			sessions[sessionID] = true
			sessionLock.Unlock()

			cookie := http.Cookie{
				Name:     "session",
				Value:    sessionID,
				Path:     "/",
				MaxAge:   3600,
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)

			go collectPingDataOnce()

			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, filepath.Join("web", "login.html"))
}

func isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	sessionLock.RLock()
	defer sessionLock.RUnlock()
	return sessions[cookie.Value]
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, filepath.Join("web", "dashboard.html"))
}

func protectedDataHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dataHandler(w, r)
}

func protectedStartMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	startMetricsHandler(w, r)
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	metricMutex.Lock()
	defer metricMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, "Failed to encode metrics data", http.StatusInternalServerError)
		return
	}
}

func startMetricsHandler(w http.ResponseWriter, r *http.Request) {
	go collectPingDataOnce()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func collectPingDataOnce() {
	forwarder.StartWithCallback(func(data forwarder.PingData) {
		AddMetric(
			data.Source,
			data.Destination,
			data.Sent,
			data.Received,
			data.PacketLoss,
			float64(data.RTT),
		)
	})
}

func collectPingData() {
	go func() {
		for {
			forwarder.StartWithCallback(func(data forwarder.PingData) {
				AddMetric(
					data.Source,
					data.Destination,
					data.Sent,
					data.Received,
					data.PacketLoss,
					float64(data.RTT),
				)
			})

			time.Sleep(30 * time.Second)
		}
	}()
}

func checkAuthHandler(w http.ResponseWriter, r *http.Request) {
	if isAuthenticated(r) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"authenticated"}`))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"unauthenticated"}`))
	}
}

func ipPoolStatusHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	used, total, percent := GetIPPoolStatus()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")

	data := IPPoolStatusData{
		Used:         used,
		Total:        total,
		UsagePercent: percent,
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode IP pool data", http.StatusInternalServerError)
		return
	}
}

func StartWebDashboard(nat64 bool) {
	collectPingData()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/login", http.StatusFound)
		} else {
			http.NotFound(w, r)
		}
	})

	http.HandleFunc("/login", loginHandler)

	http.HandleFunc("/api/data", protectedDataHandler)
	http.HandleFunc("/api/start-metrics", protectedStartMetricsHandler)
	http.HandleFunc("/api/ip-pool-status", ipPoolStatusHandler)

	http.HandleFunc("/api/check-auth", checkAuthHandler)

	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.ServeFile(w, r, filepath.Join("web", "dashboard.html"))
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err == nil {
			sessionLock.Lock()
			delete(sessions, cookie.Value)
			sessionLock.Unlock()
		}

		expiredCookie := http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		}
		http.SetCookie(w, &expiredCookie)

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join("web", "static")))))

	logger.Info("Dashboard running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
