package federation

import (
	"testing"
)

func TestNewFederationHealth(t *testing.T) {
	fh := NewFederationHealth()
	
	if fh == nil {
		t.Fatal("NewFederationHealth returned nil")
	}
	
	if fh.GetMode() != FederationModeEnabled {
		t.Errorf("Expected initial mode to be Enabled, got %v", fh.GetMode())
	}
}

func TestFederationModeString(t *testing.T) {
	tests := []struct {
		mode     FederationMode
		expected string
	}{
		{FederationModeEnabled, "Enabled"},
		{FederationModeDegraded, "Degraded"},
		{FederationModeLocalOnly, "LocalOnly"},
		{FederationMode(999), "Unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.mode.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.mode.String())
			}
		})
	}
}

func TestFederationHealthModeChecks(t *testing.T) {
	fh := NewFederationHealth()
	
	// Initially enabled
	if !fh.IsEnabled() {
		t.Error("Expected IsEnabled to be true initially")
	}
	if fh.IsDegraded() {
		t.Error("Expected IsDegraded to be false initially")
	}
	if fh.IsLocalOnly() {
		t.Error("Expected IsLocalOnly to be false initially")
	}
	
	// Switch to degraded
	fh.SetDegraded()
	if fh.IsEnabled() {
		t.Error("Expected IsEnabled to be false after SetDegraded")
	}
	if !fh.IsDegraded() {
		t.Error("Expected IsDegraded to be true after SetDegraded")
	}
	if fh.IsLocalOnly() {
		t.Error("Expected IsLocalOnly to be false after SetDegraded")
	}
	
	// Switch to local-only
	fh.SetLocalOnly()
	if fh.IsEnabled() {
		t.Error("Expected IsEnabled to be false after SetLocalOnly")
	}
	if fh.IsDegraded() {
		t.Error("Expected IsDegraded to be false after SetLocalOnly")
	}
	if !fh.IsLocalOnly() {
		t.Error("Expected IsLocalOnly to be true after SetLocalOnly")
	}
}

func TestFederationHealthSetEnabled(t *testing.T) {
	fh := NewFederationHealth()
	
	fh.SetDegraded()
	fh.SetEnabled()
	
	if fh.GetMode() != FederationModeEnabled {
		t.Errorf("Expected mode to be Enabled, got %v", fh.GetMode())
	}
	
	stats := fh.Stats()
	if stats["consecutive_failures"].(int) != 0 {
		t.Error("Expected consecutive_failures to be reset to 0 when enabled")
	}
}

func TestFederationHealthUpdateServerCountsNoServers(t *testing.T) {
	fh := NewFederationHealth()
	
	fh.UpdateServerCounts(0, 5)
	
	if fh.GetMode() != FederationModeLocalOnly {
		t.Errorf("Expected LocalOnly mode with 0 available servers, got %v", fh.GetMode())
	}
}

func TestFederationHealthUpdateServerCountsLowAvailability(t *testing.T) {
	fh := NewFederationHealth()
	
	fh.UpdateServerCounts(2, 5) // 40% availability
	
	if fh.GetMode() != FederationModeDegraded {
		t.Errorf("Expected Degraded mode with low availability, got %v", fh.GetMode())
	}
}

func TestFederationHealthUpdateServerCountsHighAvailability(t *testing.T) {
	fh := NewFederationHealth()
	
	// First degrade
	fh.SetDegraded()
	
	// Then restore
	fh.UpdateServerCounts(4, 5) // 80% availability
	
	if fh.GetMode() != FederationModeEnabled {
		t.Errorf("Expected Enabled mode with high availability, got %v", fh.GetMode())
	}
}

func TestFederationHealthUpdateServerCountsAllAvailable(t *testing.T) {
	fh := NewFederationHealth()
	
	fh.UpdateServerCounts(5, 5) // 100% availability
	
	if fh.GetMode() != FederationModeEnabled {
		t.Errorf("Expected Enabled mode with all servers available, got %v", fh.GetMode())
	}
}

func TestFederationHealthRecordFailure(t *testing.T) {
	fh := NewFederationHealth()
	
	// Record failures and check mode transitions
	for i := 1; i <= 9; i++ {
		fh.RecordFailure()
		if fh.GetMode() != FederationModeEnabled {
			t.Errorf("Expected Enabled mode after %d failures, got %v", i, fh.GetMode())
		}
	}
	
	// 10th failure should trigger degraded mode
	fh.RecordFailure()
	if fh.GetMode() != FederationModeDegraded {
		t.Errorf("Expected Degraded mode after 10 failures, got %v", fh.GetMode())
	}
	
	// Continue recording failures
	for i := 11; i <= 19; i++ {
		fh.RecordFailure()
		if fh.GetMode() != FederationModeDegraded {
			t.Errorf("Expected Degraded mode after %d failures, got %v", i, fh.GetMode())
		}
	}
	
	// 20th failure should trigger local-only mode
	fh.RecordFailure()
	if fh.GetMode() != FederationModeLocalOnly {
		t.Errorf("Expected LocalOnly mode after 20 failures, got %v", fh.GetMode())
	}
}

func TestFederationHealthRecordSuccess(t *testing.T) {
	fh := NewFederationHealth()
	
	// Record some failures
	fh.RecordFailure()
	fh.RecordFailure()
	fh.RecordFailure()
	
	stats := fh.Stats()
	if stats["consecutive_failures"].(int) != 3 {
		t.Error("Expected 3 consecutive failures")
	}
	
	// Record success
	fh.RecordSuccess()
	
	stats = fh.Stats()
	if stats["consecutive_failures"].(int) != 0 {
		t.Error("Expected consecutive_failures to be reset to 0 after success")
	}
}

func TestFederationHealthStats(t *testing.T) {
	fh := NewFederationHealth()
	
	fh.UpdateServerCounts(2, 5) // 40% availability - triggers degraded mode
	fh.RecordFailure()
	fh.RecordFailure()
	
	stats := fh.Stats()
	
	if stats["mode"].(string) != "Degraded" {
		t.Errorf("Expected mode to be Degraded in stats, got %s", stats["mode"].(string))
	}
	
	if stats["available_servers"].(int) != 2 {
		t.Errorf("Expected 2 available_servers, got %d", stats["available_servers"].(int))
	}
	
	if stats["total_servers"].(int) != 5 {
		t.Errorf("Expected 5 total_servers, got %d", stats["total_servers"].(int))
	}
	
	if stats["consecutive_failures"].(int) != 2 {
		t.Errorf("Expected 2 consecutive_failures, got %d", stats["consecutive_failures"].(int))
	}
}

func TestFederationHealthIdempotentModeChanges(t *testing.T) {
	fh := NewFederationHealth()
	
	// Set mode multiple times - should not cause issues
	fh.SetEnabled()
	fh.SetEnabled()
	if fh.GetMode() != FederationModeEnabled {
		t.Error("Expected mode to remain Enabled")
	}
	
	fh.SetDegraded()
	fh.SetDegraded()
	if fh.GetMode() != FederationModeDegraded {
		t.Error("Expected mode to remain Degraded")
	}
	
	fh.SetLocalOnly()
	fh.SetLocalOnly()
	if fh.GetMode() != FederationModeLocalOnly {
		t.Error("Expected mode to remain LocalOnly")
	}
}
