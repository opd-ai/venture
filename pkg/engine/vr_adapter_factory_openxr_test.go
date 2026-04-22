//go:build vr && !js

package engine

import "testing"

func TestNewRuntimeHeadsetAdapter_VRBuild(t *testing.T) {
	adapter := NewRuntimeHeadsetAdapter()
	if adapter == nil {
		t.Fatal("expected non-nil headset adapter")
	}
}

func TestNewRuntimeControllerAdapter_VRBuild(t *testing.T) {
	adapter := NewRuntimeControllerAdapter()
	if adapter == nil {
		t.Fatal("expected non-nil controller adapter")
	}
}
