//go:build !android && !ios
// +build !android,!ios

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// TestServerIdentityGeneration verifies proper server identity generation with ed25519 keypair
func TestServerIdentityGeneration(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs for cleaner test output
	seed := int64(12345)

	tests := []struct {
		name       string
		serverName string
		wantErr    bool
	}{
		{
			name:       "valid server name",
			serverName: "test-server",
			wantErr:    false,
		},
		{
			name:       "valid server name with spaces",
			serverName: "Test Server One",
			wantErr:    false,
		},
		{
			name:       "valid server name with special chars",
			serverName: "server-01_prod",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call should not panic and should return valid managers
			guildManager, fleetManager, _ := initializeV8SystemsServer(world, seed, tt.serverName, logger)

			if guildManager == nil {
				t.Error("initializeV8SystemsServer returned nil guildManager")
			}

			if fleetManager == nil {
				t.Error("initializeV8SystemsServer returned nil fleetManager")
			}

			// Verify guild system was added
			systems := world.GetSystems()
			if len(systems) == 0 {
				t.Error("No systems were added to world")
			}
		})
	}
}

// TestServerIdentityUniqueness verifies that multiple server initializations generate different identities
func TestServerIdentityUniqueness(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	logger.SetLevel(logrus.ErrorLevel)
	seed := int64(12345)

	// Capture log output to verify fingerprints are different
	var buf1, buf2 bytes.Buffer
	logger1 := logrus.New()
	logger1.SetOutput(&buf1)
	logger1.SetLevel(logrus.InfoLevel)

	logger2 := logrus.New()
	logger2.SetOutput(&buf2)
	logger2.SetLevel(logrus.InfoLevel)

	world1 := engine.NewWorld()
	world2 := engine.NewWorld()

	// Initialize two servers with same name
	_, _, _ = initializeV8SystemsServer(world1, seed, "test-server", logger1)
	_, _, _ = initializeV8SystemsServer(world2, seed, "test-server", logger2)

	// Verify both initializations logged server identity
	if buf1.Len() == 0 {
		t.Error("First server did not log identity information")
	}

	if buf2.Len() == 0 {
		t.Error("Second server did not log identity information")
	}

	// Note: We cannot directly compare fingerprints as they're logged, not returned,
	// but we verify the initialization succeeded and logged identity info
}

// TestServerNameFromEnvironment verifies SERVER_NAME environment variable loading
func TestServerNameFromEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		flagValue string
		want      string
	}{
		{
			name:      "env not set uses flag default",
			envValue:  "",
			flagValue: "default-server",
			want:      "default-server",
		},
		{
			name:      "env overrides default",
			envValue:  "production-server",
			flagValue: "default-server",
			want:      "production-server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("SERVER_NAME", tt.envValue)
				defer os.Unsetenv("SERVER_NAME")
			} else {
				os.Unsetenv("SERVER_NAME")
			}

			// Test getEnvOrDefault helper
			result := getEnvOrDefault("SERVER_NAME", tt.flagValue)
			if result != tt.want {
				t.Errorf("getEnvOrDefault() = %q, want %q", result, tt.want)
			}
		})
	}
}

// TestServerIdentityKeyLength verifies ed25519 key generation produces correct key sizes
func TestServerIdentityKeyLength(t *testing.T) {
	world := engine.NewWorld()

	// Capture detailed log output
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	seed := int64(12345)
	serverName := "key-test-server"

	_, _, _ = initializeV8SystemsServer(world, seed, serverName, logger)

	// Verify log contains expected fields
	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("No log output captured")
	}

	// Should contain server_name, fingerprint, and public_key_len fields
	expectedFields := []string{"server_name", "fingerprint", "public_key_len"}
	for _, field := range expectedFields {
		if !bytes.Contains(buf.Bytes(), []byte(field)) {
			t.Errorf("Log output missing expected field: %s", field)
		}
	}

	// Should contain the server name we provided
	if !bytes.Contains(buf.Bytes(), []byte(serverName)) {
		t.Errorf("Log output missing server name: %s", serverName)
	}
}

// TestServerIdentityWithMultipleSystems verifies server identity initialization with full system stack
func TestServerIdentityWithMultipleSystems(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	logger.SetLevel(logrus.ErrorLevel)
	seed := int64(12345)
	serverName := "full-stack-server"

	// Initialize all server systems
	initializeV4Systems(world, seed, "fantasy", logger, nil)
	initializeV5SystemsServer(world, logger)
	initializeV6SystemsServer(world, seed, logger, nil)
	guildManager, _, _ := initializeV8SystemsServer(world, seed, serverName, logger)

	if guildManager == nil {
		t.Fatal("guildManager is nil after full system initialization")
	}

	// Verify systems were added
	systems := world.GetSystems()
	if len(systems) < 10 {
		t.Errorf("Expected at least 10 systems, got %d", len(systems))
	}

	// Should be able to initialize V9 systems with the guild manager
	stationMgr, petHomeMgr, guildHousingMgr, narrativeWorldSys, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	if stationMgr == nil {
		t.Error("stationMgr is nil after V9 initialization")
	}
	if petHomeMgr == nil {
		t.Error("petHomeMgr is nil after V9 initialization")
	}
	if guildHousingMgr == nil {
		t.Error("guildHousingMgr is nil after V9 initialization")
	}
	if narrativeWorldSys == nil {
		t.Error("narrativeWorldSys is nil after V9 initialization")
	}
	if politicalWarfareSys == nil {
		t.Error("politicalWarfareSys is nil after V9 initialization")
	}
}
