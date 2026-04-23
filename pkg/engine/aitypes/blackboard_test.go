package aitypes_test

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine/aitypes"
)

func TestBlackboard_SetGet(t *testing.T) {
	bb := aitypes.NewBlackboard()
	bb.Set("key", "value")
	val, ok := bb.Get("key")
	if !ok {
		t.Fatal("Get: expected ok=true")
	}
	if val != "value" {
		t.Fatalf("Get: got %v, want %q", val, "value")
	}
}

func TestBlackboard_Get_Missing(t *testing.T) {
	bb := aitypes.NewBlackboard()
	_, ok := bb.Get("missing")
	if ok {
		t.Fatal("Get: expected ok=false for missing key")
	}
}

func TestBlackboard_Clear(t *testing.T) {
	bb := aitypes.NewBlackboard()
	bb.Set("a", 1)
	bb.Set("b", 2)
	bb.Clear()
	_, ok := bb.Get("a")
	if ok {
		t.Fatal("Clear: key 'a' still present after Clear")
	}
	_, ok = bb.Get("b")
	if ok {
		t.Fatal("Clear: key 'b' still present after Clear")
	}
}

func TestBlackboard_GetFloat64(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantVal float64
		wantOk  bool
	}{
		{"present float64", float64(3.14), 3.14, true},
		{"wrong type int", 42, 0.0, false},
		{"missing key", nil, 0.0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bb := aitypes.NewBlackboard()
			if tt.value != nil {
				bb.Set("k", tt.value)
			}
			got, ok := bb.GetFloat64("k")
			if ok != tt.wantOk {
				t.Errorf("GetFloat64 ok=%v, want %v", ok, tt.wantOk)
			}
			if ok && got != tt.wantVal {
				t.Errorf("GetFloat64 val=%v, want %v", got, tt.wantVal)
			}
		})
	}
}

func TestBlackboard_GetBool(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantVal bool
		wantOk  bool
	}{
		{"present true", true, true, true},
		{"present false", false, false, true},
		{"wrong type string", "yes", false, false},
		{"missing key", nil, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bb := aitypes.NewBlackboard()
			if tt.value != nil {
				bb.Set("flag", tt.value)
			}
			got, ok := bb.GetBool("flag")
			if ok != tt.wantOk {
				t.Errorf("GetBool ok=%v, want %v", ok, tt.wantOk)
			}
			if ok && got != tt.wantVal {
				t.Errorf("GetBool val=%v, want %v", got, tt.wantVal)
			}
		})
	}
}

func TestBlackboard_SetRNG_GetRNG(t *testing.T) {
	bb := aitypes.NewBlackboard()
	rng := bb.GetRNG()
	if rng == nil {
		t.Fatal("GetRNG returned nil on default blackboard")
	}
}

func TestBlackboardWithSeed_Determinism(t *testing.T) {
	const seed = int64(42)
	bb1 := aitypes.NewBlackboardWithSeed(seed)
	bb2 := aitypes.NewBlackboardWithSeed(seed)

	for i := 0; i < 10; i++ {
		v1 := bb1.GetRNG().Float64()
		v2 := bb2.GetRNG().Float64()
		if v1 != v2 {
			t.Fatalf("RNG non-deterministic at step %d: %v != %v", i, v1, v2)
		}
	}
}

func TestBlackboard_Overwrite(t *testing.T) {
	bb := aitypes.NewBlackboard()
	bb.Set("k", "first")
	bb.Set("k", "second")
	val, ok := bb.Get("k")
	if !ok || val != "second" {
		t.Fatalf("Set overwrite: got %v %v, want 'second' true", val, ok)
	}
}
