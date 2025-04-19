package forwarder

import (
	"net"
	"time"

	"github.com/go-ping/ping"
	"github.com/mdxabu/bridge/internal/config"
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

func Start() {
	nextIP, err := config.GetDestIpAddress()
	if err != nil {
		logger.Error("Failed to get destination IP addresses: %v", err)
		return
	}

	logger.ClearPingResults()
	logger.PrintTableHeader()

	ipCount := 0

	for {
		ip, ok := nextIP()
		if !ok {
			break
		}
		pingDestination(ip, nil)
		ipCount++
	}

	if ipCount > 0 {
		logger.Info("Completed pinging %d destinations", ipCount)
	} else {
		logger.Warn("No destinations were pinged")
	}

	logger.DisplayPingTable()
}

func StartWithCallback(callback func(PingData)) {
	nextIP, err := config.GetDestIpAddress()
	if err != nil {
		logger.Error("Failed to get destination IP addresses: %v", err)
		return
	}

	for {
		ip, ok := nextIP()
		if !ok {
			break
		}
		go pingDestination(ip, callback)
		time.Sleep(200 * time.Millisecond)
	}
}

func pingDestination(ip string, callback func(PingData)) {
	sourceIP := getSourceIP()

	pinger, err := ping.NewPinger(ip)
	if err != nil {
		logger.PingTable(sourceIP, ip, 0, 0, 100.0, 0)
		logger.Error("Failed to create pinger for %s: %v", ip, err)
		return
	}

	pinger.Count = 5
	pinger.Timeout = 5 * time.Second
	pinger.SetPrivileged(true)

	pinger.OnFinish = func(stats *ping.Statistics) {
		logger.PingTable(sourceIP, ip, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss, stats.AvgRtt)

		if callback != nil {
			callback(PingData{
				Source:      sourceIP,
				Destination: ip,
				Sent:        stats.PacketsSent,
				Received:    stats.PacketsRecv,
				PacketLoss:  stats.PacketLoss,
				RTT:         stats.AvgRtt.Milliseconds(),
				Timestamp:   time.Now().Unix(),
			})
		}
	}

	if err := pinger.Run(); err != nil {
		logger.PingTable(sourceIP, ip, 2, 0, 100.0, 0)
		logger.Error("Ping failed for %s: %v", ip, err)
	}
}

func getSourceIP() string {
	conf, err := config.ParseConfig()
	if err != nil {
		logger.Error("Failed to parse configuration: %v", err)
		return "unknown"
	}

	nat64, err := conf.GetNAT64IP()
	if err != nil {
		logger.Error("Failed to get NAT64 IP: %v", err)
		return "unknown"
	}

	ip := net.ParseIP(nat64)
	if ip == nil {
		logger.Error("Invalid NAT64 IP: %s", nat64)
		return "unknown"
	}

	return nat64
}
