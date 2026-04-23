package aitypes_test

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine/aitypes"
)

func TestNodeStatus_String(t *testing.T) {
	tests := []struct {
		status aitypes.NodeStatus
		want   string
	}{
		{aitypes.NodeSuccess, "Success"},
		{aitypes.NodeFailure, "Failure"},
		{aitypes.NodeRunning, "Running"},
		{aitypes.NodeStatus(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("NodeStatus(%d).String() = %q, want %q", int(tt.status), got, tt.want)
			}
		})
	}
}

func TestNodeStatus_Values(t *testing.T) {
	// Ensure the iota values are stable (Success=0, Failure=1, Running=2).
	if int(aitypes.NodeSuccess) != 0 {
		t.Errorf("NodeSuccess = %d, want 0", int(aitypes.NodeSuccess))
	}
	if int(aitypes.NodeFailure) != 1 {
		t.Errorf("NodeFailure = %d, want 1", int(aitypes.NodeFailure))
	}
	if int(aitypes.NodeRunning) != 2 {
		t.Errorf("NodeRunning = %d, want 2", int(aitypes.NodeRunning))
	}
}
