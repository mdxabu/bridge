package state

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/mdxabu/bridge/internal/logger" // Corrected import path
)

// Table represents the stateful NAT connection tracking table (BIB).
type Table struct {
	entries map[string]*Entry // Key: IPv6Src:Port-IPv4Dest:Port
	mu      sync.RWMutex
	timeout time.Duration
	log     *logger.Logger
}

// NewTable creates a new state table with the given timeout.
func NewTable(timeout int) *Table {
	return &Table{
		entries: make(map[string]*Entry),
		timeout: time.Duration(timeout) * time.Second,
		// Logger needs to be set from the gateway
	}
}

// SetLogger sets the logger for the state table.
func (t *Table) SetLogger(log *logger.Logger) {
	t.log = log
}

// Entry represents an entry in the state table.
type Entry struct {
	IPv6SrcIP   net.IP
	IPv6SrcPort uint16
	IPv4DstIP   net.IP
	IPv4DstPort uint16
	IPv4SrcIP   net.IP // Gateway's external IPv4
	IPv4SrcPort uint16 // Unique port allocated by the gateway
	LastActive  time.Time
}

// CreateEntry creates a new entry in the state table.
func (t *Table) CreateEntry(ipv6SrcIP net.IP, ipv6SrcPort uint16, ipv4DstIP net.IP, ipv4DstPort uint16, ipv4SrcIP net.IP, ipv4SrcPort uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := generateKey(ipv6SrcIP, ipv6SrcPort, ipv4DstIP, ipv4DstPort)
	t.entries[key] = &Entry{
		IPv6SrcIP:   ipv6SrcIP,
		IPv6SrcPort: ipv6SrcPort,
		IPv4DstIP:   ipv4DstIP,
		IPv4DstPort: ipv4DstPort,
		IPv4SrcIP:   ipv4SrcIP,
		IPv4SrcPort: ipv4SrcPort,
		LastActive:  time.Now(),
	}
	if t.log != nil {
		t.log.Debugf("Created state entry: %s:%d -> %s:%d (via %s:%d)", ipv6SrcIP, ipv6SrcPort, ipv4DstIP, ipv4DstPort, ipv4SrcIP, ipv4SrcPort)
	}
}

// LookupIPv6ToIPv4 finds an entry based on the IPv6 source and IPv4 destination.
func (t *Table) LookupIPv6ToIPv4(ipv6SrcIP net.IP, ipv6SrcPort uint16, ipv4DstIP net.IP, ipv4DstPort uint16) *Entry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	key := generateKey(ipv6SrcIP, ipv6SrcPort, ipv4DstIP, ipv4DstPort)
	if entry, ok := t.entries[key]; ok {
		entry.LastActive = time.Now()
		return entry
	}
	return nil
}

// LookupIPv4ToIPv6 finds an entry based on the IPv4 destination and source.
func (t *Table) LookupIPv4ToIPv6(ipv4DstIP net.IP, ipv4DstPort uint16, ipv4SrcIP net.IP, ipv4SrcPort uint16) *Entry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, entry := range t.entries {
		if entry.IPv4SrcIP.Equal(ipv4DstIP) && entry.IPv4SrcPort == ipv4DstPort &&
			entry.IPv4DstIP.Equal(ipv4SrcIP) && entry.IPv4DstPort == ipv4SrcPort {
			entry.LastActive = time.Now()
			return entry
		}
	}
	return nil
}

// DeleteEntry deletes an entry from the state table.
func (t *Table) DeleteEntry(ipv6SrcIP net.IP, ipv6SrcPort uint16, ipv4DstIP net.IP, ipv4DstPort uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := generateKey(ipv6SrcIP, ipv6SrcPort, ipv4DstIP, ipv4DstPort)
	delete(t.entries, key)
	if t.log != nil {
		t.log.Debugf("Deleted state entry: %s:%d -> %s:%d", ipv6SrcIP, ipv6SrcPort, ipv4DstIP, ipv4DstPort)
	}
}

// CleanupExpiredEntries removes entries that have exceeded the timeout.
func (t *Table) CleanupExpiredEntries() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for key, entry := range t.entries {
		if now.Sub(entry.LastActive) > t.timeout {
			delete(t.entries, key)
			if t.log != nil {
				t.log.Debugf("Removed expired state entry: %s:%d -> %s:%d (inactive for %v)", entry.IPv6SrcIP, entry.IPv6SrcPort, entry.IPv4DstIP, entry.IPv4DstPort, now.Sub(entry.LastActive))
			}
		}
	}
}

// StartCleanupRoutine starts a goroutine to periodically clean up expired entries.
func (t *Table) StartCleanupRoutine() {
	ticker := time.NewTicker(t.timeout / 2) // Check for expired entries every half of the timeout period
	defer ticker.Stop()
	for range ticker.C {
		t.CleanupExpiredEntries()
	}
}

// generateKey creates a unique key for the state table entry.
func generateKey(ipv6SrcIP net.IP, ipv6SrcPort uint16, ipv4DstIP net.IP, ipv4DstPort uint16) string {
	return fmt.Sprintf("%s:%d-%s:%d", ipv6SrcIP.String(), ipv6SrcPort, ipv4DstIP.String(), ipv4DstPort)
}
