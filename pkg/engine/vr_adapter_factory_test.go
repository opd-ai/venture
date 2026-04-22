package engine

import "testing"

func TestNewRuntimeHeadsetAdapter_DefaultBuildUsesStub(t *testing.T) {
	adapter := NewRuntimeHeadsetAdapter()
	if adapter == nil {
		t.Fatal("expected non-nil headset adapter")
	}
	if _, ok := adapter.(*StubHeadsetAdapter); !ok {
		t.Fatalf("expected stub headset adapter in default build, got %T", adapter)
	}
}

func TestNewRuntimeControllerAdapter_DefaultBuildUsesStub(t *testing.T) {
	adapter := NewRuntimeControllerAdapter()
	if adapter == nil {
		t.Fatal("expected non-nil controller adapter")
	}
	if _, ok := adapter.(*StubControllerAdapter); !ok {
		t.Fatalf("expected stub controller adapter in default build, got %T", adapter)
	}
}
