package state

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/mdxabu/bridge/internal/logger"
)

type Table struct {
	entries map[string]*Entry
	mu      sync.RWMutex
	timeout time.Duration
	log     *logger.Logger
}

func NewTable(timeout int) *Table {
	return &Table{
		entries: make(map[string]*Entry),
		timeout: time.Duration(timeout) * time.Second,
	}
}

func (t *Table) SetLogger(log *logger.Logger) {
	t.log = log
}

type Entry struct {
	IPv6SrcIP   net.IP
	IPv6SrcPort uint16
	IPv4DstIP   net.IP
	IPv4DstPort uint16
	IPv4SrcIP   net.IP
	IPv4SrcPort uint16
	LastActive  time.Time
}

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
		t.log.Debug("Created state entry: %s:%d -> %s:%d (via %s:%d)", ipv6SrcIP, ipv6SrcPort, ipv4DstIP, ipv4DstPort, ipv4SrcIP, ipv4SrcPort)
	}
}

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

func (t *Table) DeleteEntry(ipv6SrcIP net.IP, ipv6SrcPort uint16, ipv4DstIP net.IP, ipv4DstPort uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := generateKey(ipv6SrcIP, ipv6SrcPort, ipv4DstIP, ipv4DstPort)
	delete(t.entries, key)
	if t.log != nil {
		t.log.Debug("Deleted state entry: %s:%d -> %s:%d", ipv6SrcIP, ipv6SrcPort, ipv4DstIP, ipv4DstPort)
	}
}

func (t *Table) CleanupExpiredEntries() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for key, entry := range t.entries {
		if now.Sub(entry.LastActive) > t.timeout {
			delete(t.entries, key)
			if t.log != nil {
				t.log.Debug("Removed expired state entry: %s:%d -> %s:%d (inactive for %v)", entry.IPv6SrcIP, entry.IPv6SrcPort, entry.IPv4DstIP, entry.IPv4DstPort, now.Sub(entry.LastActive))
			}
		}
	}
}

func (t *Table) StartCleanupRoutine() {
	ticker := time.NewTicker(t.timeout / 2)
	defer ticker.Stop()
	for range ticker.C {
		t.CleanupExpiredEntries()
	}
}

func generateKey(ipv6SrcIP net.IP, ipv6SrcPort uint16, ipv4DstIP net.IP, ipv4DstPort uint16) string {
	return fmt.Sprintf("%s:%d-%s:%d", ipv6SrcIP.String(), ipv6SrcPort, ipv4DstIP.String(), ipv4DstPort)
}
