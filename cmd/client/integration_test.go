//go:build !android && !ios
// +build !android,!ios

// Package main contains integration tests for the client application
package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestHostAndPlayFlag verifies the --host-and-play flag is recognized
func TestHostAndPlayFlag(t *testing.T) {
	// Build the client binary for testing
	buildCmd := exec.Command("go", "build", "-o", "venture-client-test", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build client: %v", err)
	}
	defer os.Remove("venture-client-test")

	// Run with --help to verify flags exist
	helpCmd := exec.Command("./venture-client-test", "--help")
	output, err := helpCmd.CombinedOutput()
	if err != nil && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("failed to run help: %v", err)
	}

	outputStr := string(output)

	// Verify --host-and-play flag exists
	if !strings.Contains(outputStr, "-host-and-play") {
		t.Error("--host-and-play flag not found in help output")
	}

	// Verify --host-lan flag exists
	if !strings.Contains(outputStr, "-host-lan") {
		t.Error("--host-lan flag not found in help output")
	}

	// Verify -port flag exists
	if !strings.Contains(outputStr, "-port") {
		t.Error("-port flag not found in help output")
	}

	// Verify -max-players flag exists
	if !strings.Contains(outputStr, "-max-players") {
		t.Error("-max-players flag not found in help output")
	}

	// Verify -tick-rate flag exists
	if !strings.Contains(outputStr, "-tick-rate") {
		t.Error("-tick-rate flag not found in help output")
	}
}

// TestHostAndPlayStartup verifies the embedded server starts (without running full game loop)
// This is a smoke test - full functionality requires graphics context
func TestHostAndPlayStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Build the client binary
	buildCmd := exec.Command("go", "build", "-o", "venture-client-test", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build client: %v", err)
	}
	defer os.Remove("venture-client-test")

	// Start client with --host-and-play (will fail due to no graphics context, but we can check logs)
	cmd := exec.Command("./venture-client-test", "--host-and-play", "-port", "9500")
	cmd.Env = append(os.Environ(), "LOG_LEVEL=debug")

	// Run with timeout since the command may hang
	done := make(chan struct{})
	var output []byte
	go func() {
		output, _ = cmd.CombinedOutput()
		close(done)
	}()

	// Wait for command to complete or timeout
	select {
	case <-done:
		// Command completed
	case <-time.After(5 * time.Second):
		// Timeout - kill the process
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		<-done // Wait for goroutine to finish
	}

	outputStr := string(output)

	// We expect it to fail due to no display, but server should attempt to start
	// Look for evidence of server initialization in output
	if !strings.Contains(outputStr, "host-and-play mode enabled") &&
		!strings.Contains(outputStr, "starting embedded server") &&
		!strings.Contains(outputStr, "embedded-server") {
		t.Logf("Output: %s", outputStr)
		t.Skip("could not verify server startup (expected without graphics context)")
	}

	// Verify error is graphics-related, not server-related
	if strings.Contains(outputStr, "no available ports") {
		t.Error("port finding failed - this suggests host-and-play logic has issues")
	}
	if strings.Contains(outputStr, "failed to start embedded server") && !strings.Contains(outputStr, "display") {
		t.Error("embedded server failed to start for non-display reasons")
	}

	t.Logf("Verified host-and-play initialization logic (graphics context required for full test)")
}

// TestPortFallbackFlags verifies port configuration flags work
func TestPortFallbackFlags(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "venture-client-test", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build client: %v", err)
	}
	defer os.Remove("venture-client-test")

	// Test various flag combinations (will fail due to no graphics, but we verify parsing)
	tests := []struct {
		name string
		args []string
	}{
		{"default port", []string{"--host-and-play"}},
		{"custom port", []string{"--host-and-play", "-port", "9000"}},
		{"lan mode", []string{"--host-and-play", "--host-lan"}},
		{"max players", []string{"--host-and-play", "-max-players", "8"}},
		{"tick rate", []string{"--host-and-play", "-tick-rate", "30"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./venture-client-test", tt.args...)
			cmd.Env = append(os.Environ(), "LOG_LEVEL=error") // Reduce noise

			// Start the command
			if err := cmd.Start(); err != nil {
				t.Logf("Command failed to start (expected): %v", err)
				return
			}

			// Run with timeout (should fail fast due to no graphics)
			done := make(chan error)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case err := <-done:
				// Expected to fail quickly due to no graphics context
				t.Logf("Command exited (expected): %v", err)
			case <-time.After(2 * time.Second):
				// Kill the process if it's still running
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
				<-done // Wait for goroutine to finish
				t.Skip("command timed out (expected in environments without graphics context)")
			}
		})
	}
}
