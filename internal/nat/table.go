package nat

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// SessionState represents the state of a NAT session
type SessionState struct {
	ID              string
	Protocol        uint8
	IPv6SrcIP       net.IP
	IPv6SrcPort     uint16
	IPv6DstIP       net.IP
	IPv6DstPort     uint16
	IPv4SrcIP       net.IP
	IPv4SrcPort     uint16
	IPv4DstIP       net.IP
	IPv4DstPort     uint16
	CreatedAt       time.Time
	LastActivity    time.Time
	BytesSent       uint64
	BytesReceived   uint64
	PacketsSent     uint64
	PacketsReceived uint64
	State           string // NEW, ESTABLISHED, CLOSING, CLOSED
}

// NATTable manages NAT sessions
type NATTable struct {
	sessions       map[string]*SessionState
	portMappings   map[uint16]*SessionState // Maps allocated ports to sessions
	mu             sync.RWMutex
	nextPort       uint16
	portRangeStart uint16
	portRangeEnd   uint16
	timeoutTCP     time.Duration
	timeoutUDP     time.Duration
}

// NewNATTable creates a new NAT table
func NewNATTable() *NATTable {
	return &NATTable{
		sessions:       make(map[string]*SessionState),
		portMappings:   make(map[uint16]*SessionState),
		nextPort:       10000,
		portRangeStart: 10000,
		portRangeEnd:   65000,
		timeoutTCP:     300 * time.Second,  // 5 minutes for TCP
		timeoutUDP:     60 * time.Second,   // 1 minute for UDP
	}
}

// CreateSession creates a new NAT session
func (nt *NATTable) CreateSession(protocol uint8, ipv6Src net.IP, ipv6SrcPort uint16, ipv6Dst net.IP, ipv6DstPort uint16, ipv4Dst net.IP) (*SessionState, error) {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	// Generate session ID
	sessionID := fmt.Sprintf("%d:%s:%d->%s:%d", protocol, ipv6Src, ipv6SrcPort, ipv6Dst, ipv6DstPort)

	// Check if session already exists
	if session, exists := nt.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
		return session, nil
	}

	// Allocate a new port for this session
	port, err := nt.allocatePort()
	if err != nil {
		return nil, err
	}

	// Extract IPv4 from NAT64 address
	ipv4DstPort := ipv6DstPort

	// Create new session
	session := &SessionState{
		ID:            sessionID,
		Protocol:      protocol,
		IPv6SrcIP:     ipv6Src,
		IPv6SrcPort:   ipv6SrcPort,
		IPv6DstIP:     ipv6Dst,
		IPv6DstPort:   ipv6DstPort,
		IPv4SrcIP:     net.ParseIP("10.64.0.1").To4(), // NAT gateway address
		IPv4SrcPort:   port,
		IPv4DstIP:     ipv4Dst,
		IPv4DstPort:   ipv4DstPort,
		CreatedAt:     time.Now(),
		LastActivity:  time.Now(),
		State:         "NEW",
	}

	// Store session
	nt.sessions[sessionID] = session
	nt.portMappings[port] = session

	return session, nil
}

// LookupSessionIPv6toIPv4 looks up a session for IPv6 to IPv4 translation
func (nt *NATTable) LookupSessionIPv6toIPv4(protocol uint8, srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16) (*SessionState, bool) {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	sessionID := fmt.Sprintf("%d:%s:%d->%s:%d", protocol, srcIP, srcPort, dstIP, dstPort)
	session, exists := nt.sessions[sessionID]
	
	if exists {
		return session, true
	}
	
	return nil, false
}

// LookupSessionIPv4toIPv6 looks up a session for IPv4 to IPv6 translation (reverse)
func (nt *NATTable) LookupSessionIPv4toIPv6(protocol uint8, dstPort uint16) (*SessionState, bool) {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	session, exists := nt.portMappings[dstPort]
	if exists && session.Protocol == protocol {
		return session, true
	}
	
	return nil, false
}

// UpdateSession updates session statistics
func (nt *NATTable) UpdateSession(sessionID string, bytesSent uint64, direction string) {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	session, exists := nt.sessions[sessionID]
	if !exists {
		return
	}

	session.LastActivity = time.Now()
	
	if direction == "outbound" {
		session.BytesSent += bytesSent
		session.PacketsSent++
	} else {
		session.BytesReceived += bytesSent
		session.PacketsReceived++
	}

	// Update state based on activity
	if session.State == "NEW" && session.PacketsReceived > 0 {
		session.State = "ESTABLISHED"
	}
}

// RemoveSession removes a session from the NAT table
func (nt *NATTable) RemoveSession(sessionID string) {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	session, exists := nt.sessions[sessionID]
	if !exists {
		return
	}

	// Free the port
	delete(nt.portMappings, session.IPv4SrcPort)
	delete(nt.sessions, sessionID)
}

// CleanupExpiredSessions removes expired sessions
func (nt *NATTable) CleanupExpiredSessions() int {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	now := time.Now()
	removed := 0

	for sessionID, session := range nt.sessions {
		var timeout time.Duration
		
		// Set timeout based on protocol
		if session.Protocol == 6 { // TCP
			timeout = nt.timeoutTCP
		} else if session.Protocol == 17 { // UDP
			timeout = nt.timeoutUDP
		} else {
			timeout = 30 * time.Second
		}

		// Remove if expired
		if now.Sub(session.LastActivity) > timeout {
			delete(nt.portMappings, session.IPv4SrcPort)
			delete(nt.sessions, sessionID)
			removed++
		}
	}

	return removed
}

// GetAllSessions returns all active sessions
func (nt *NATTable) GetAllSessions() []*SessionState {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	sessions := make([]*SessionState, 0, len(nt.sessions))
	for _, session := range nt.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// GetSessionCount returns the number of active sessions
func (nt *NATTable) GetSessionCount() int {
	nt.mu.RLock()
	defer nt.mu.RUnlock()
	
	return len(nt.sessions)
}

// GetStats returns NAT table statistics
func (nt *NATTable) GetStats() map[string]interface{} {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	tcpCount := 0
	udpCount := 0
	totalBytesSent := uint64(0)
	totalBytesReceived := uint64(0)

	for _, session := range nt.sessions {
		if session.Protocol == 6 {
			tcpCount++
		} else if session.Protocol == 17 {
			udpCount++
		}
		totalBytesSent += session.BytesSent
		totalBytesReceived += session.BytesReceived
	}

	return map[string]interface{}{
		"total_sessions":   len(nt.sessions),
		"tcp_sessions":     tcpCount,
		"udp_sessions":     udpCount,
		"bytes_sent":       totalBytesSent,
		"bytes_received":   totalBytesReceived,
		"allocated_ports":  len(nt.portMappings),
	}
}

// allocatePort allocates a new port for NAT
func (nt *NATTable) allocatePort() (uint16, error) {
	attempts := 0
	maxAttempts := int(nt.portRangeEnd - nt.portRangeStart)

	for attempts < maxAttempts {
		port := nt.nextPort
		nt.nextPort++
		
		if nt.nextPort > nt.portRangeEnd {
			nt.nextPort = nt.portRangeStart
		}

		// Check if port is available
		if _, exists := nt.portMappings[port]; !exists {
			return port, nil
		}

		attempts++
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", nt.portRangeStart, nt.portRangeEnd)
}

// StartCleanupRoutine starts a background goroutine to cleanup expired sessions
func (nt *NATTable) StartCleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			removed := nt.CleanupExpiredSessions()
			if removed > 0 {
				// Log cleanup if needed
				_ = removed
			}
		}
	}()
}
