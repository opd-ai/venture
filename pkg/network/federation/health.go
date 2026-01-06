// Package federation provides server-to-server federation protocol.
package federation

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// FederationMode represents the operating mode of the federation system
type FederationMode int

const (
	// FederationModeEnabled means federation is fully operational
	FederationModeEnabled FederationMode = iota
	// FederationModeDegraded means federation is experiencing issues but still operational
	FederationModeDegraded
	// FederationModeLocalOnly means federation is disabled, operating in local-only mode
	FederationModeLocalOnly
)

// String returns a human-readable name for the federation mode
func (m FederationMode) String() string {
	switch m {
	case FederationModeEnabled:
		return "Enabled"
	case FederationModeDegraded:
		return "Degraded"
	case FederationModeLocalOnly:
		return "LocalOnly"
	default:
		return "Unknown"
	}
}

// FederationHealth tracks the health and availability of the federation system
type FederationHealth struct {
	mu                 sync.RWMutex
	mode               FederationMode
	availableServers   int
	totalServers       int
	consecutiveFailures int
	logger             *logrus.Entry
}

// NewFederationHealth creates a new federation health tracker
func NewFederationHealth() *FederationHealth {
	return &FederationHealth{
		mode: FederationModeEnabled,
		logger: logrus.WithFields(logrus.Fields{
			"component": "federation_health",
		}),
	}
}

// GetMode returns the current federation mode (thread-safe)
func (fh *FederationHealth) GetMode() FederationMode {
	fh.mu.RLock()
	defer fh.mu.RUnlock()
	return fh.mode
}

// IsEnabled returns true if federation is in enabled mode
func (fh *FederationHealth) IsEnabled() bool {
	return fh.GetMode() == FederationModeEnabled
}

// IsDegraded returns true if federation is in degraded mode
func (fh *FederationHealth) IsDegraded() bool {
	return fh.GetMode() == FederationModeDegraded
}

// IsLocalOnly returns true if federation is in local-only mode
func (fh *FederationHealth) IsLocalOnly() bool {
	return fh.GetMode() == FederationModeLocalOnly
}

// SetLocalOnly switches federation to local-only mode
func (fh *FederationHealth) SetLocalOnly() {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	
	if fh.mode != FederationModeLocalOnly {
		oldMode := fh.mode
		fh.mode = FederationModeLocalOnly
		fh.logger.WithFields(logrus.Fields{
			"old_mode": oldMode.String(),
			"new_mode": fh.mode.String(),
		}).Warn("Federation switched to local-only mode")
	}
}

// SetEnabled switches federation to enabled mode
func (fh *FederationHealth) SetEnabled() {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	
	if fh.mode != FederationModeEnabled {
		oldMode := fh.mode
		fh.mode = FederationModeEnabled
		fh.consecutiveFailures = 0
		fh.logger.WithFields(logrus.Fields{
			"old_mode": oldMode.String(),
			"new_mode": fh.mode.String(),
		}).Info("Federation switched to enabled mode")
	}
}

// SetDegraded switches federation to degraded mode
func (fh *FederationHealth) SetDegraded() {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	
	if fh.mode != FederationModeDegraded {
		oldMode := fh.mode
		fh.mode = FederationModeDegraded
		fh.logger.WithFields(logrus.Fields{
			"old_mode": oldMode.String(),
			"new_mode": fh.mode.String(),
		}).Warn("Federation switched to degraded mode")
	}
}

// UpdateServerCounts updates the available and total server counts
func (fh *FederationHealth) UpdateServerCounts(available, total int) {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	
	fh.availableServers = available
	fh.totalServers = total
	
	// Auto-adjust mode based on server availability
	if total == 0 {
		// No servers configured, stay in current mode
		return
	}
	
	availabilityRatio := float64(available) / float64(total)
	
	if availabilityRatio == 0 {
		// No servers available, switch to local-only
		if fh.mode != FederationModeLocalOnly {
			fh.mode = FederationModeLocalOnly
			fh.logger.Warn("No federation servers available, switching to local-only mode")
		}
	} else if availabilityRatio < 0.5 {
		// Less than 50% available, switch to degraded
		if fh.mode != FederationModeDegraded && fh.mode != FederationModeLocalOnly {
			fh.mode = FederationModeDegraded
			fh.logger.WithFields(logrus.Fields{
				"available": available,
				"total":     total,
				"ratio":     availabilityRatio,
			}).Warn("Federation server availability low, switching to degraded mode")
		}
	} else {
		// Majority available, switch to enabled
		if fh.mode != FederationModeEnabled {
			fh.mode = FederationModeEnabled
			fh.logger.WithFields(logrus.Fields{
				"available": available,
				"total":     total,
				"ratio":     availabilityRatio,
			}).Info("Federation server availability restored, switching to enabled mode")
		}
	}
}

// RecordFailure records a federation operation failure
func (fh *FederationHealth) RecordFailure() {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	
	fh.consecutiveFailures++
	
	// If we have too many consecutive failures, degrade or go local-only
	if fh.consecutiveFailures >= 10 && fh.mode == FederationModeEnabled {
		fh.mode = FederationModeDegraded
		fh.logger.WithFields(logrus.Fields{
			"consecutive_failures": fh.consecutiveFailures,
		}).Warn("Too many consecutive failures, switching to degraded mode")
	} else if fh.consecutiveFailures >= 20 && fh.mode == FederationModeDegraded {
		fh.mode = FederationModeLocalOnly
		fh.logger.WithFields(logrus.Fields{
			"consecutive_failures": fh.consecutiveFailures,
		}).Error("Critical failure count reached, switching to local-only mode")
	}
}

// RecordSuccess records a successful federation operation
func (fh *FederationHealth) RecordSuccess() {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	
	fh.consecutiveFailures = 0
}

// Stats returns current federation health statistics
func (fh *FederationHealth) Stats() map[string]interface{} {
	fh.mu.RLock()
	defer fh.mu.RUnlock()
	
	return map[string]interface{}{
		"mode":                 fh.mode.String(),
		"available_servers":    fh.availableServers,
		"total_servers":        fh.totalServers,
		"consecutive_failures": fh.consecutiveFailures,
	}
}
