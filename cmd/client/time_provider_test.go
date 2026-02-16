//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"
	"time"
)

func TestRealTimeProvider(t *testing.T) {
	tp := RealTimeProvider{}

	before := time.Now()
	got := tp.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("RealTimeProvider.Now() = %v, want between %v and %v", got, before, after)
	}
}

func TestMockTimeProvider(t *testing.T) {
	fixed := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	tp := MockTimeProvider{FixedTime: fixed}

	got := tp.Now()
	if !got.Equal(fixed) {
		t.Errorf("MockTimeProvider.Now() = %v, want %v", got, fixed)
	}

	// Verify determinism: multiple calls return same time
	got2 := tp.Now()
	if !got2.Equal(fixed) {
		t.Errorf("MockTimeProvider.Now() second call = %v, want %v", got2, fixed)
	}
}

func TestDefaultTimeProvider(t *testing.T) {
	tp := DefaultTimeProvider()
	if tp == nil {
		t.Fatal("DefaultTimeProvider() returned nil")
	}

	if _, ok := tp.(RealTimeProvider); !ok {
		t.Errorf("DefaultTimeProvider() returned %T, want RealTimeProvider", tp)
	}
}

func TestTimeProviderInterface(t *testing.T) {
	// Verify both types satisfy the interface
	var _ TimeProvider = RealTimeProvider{}
	var _ TimeProvider = MockTimeProvider{}
}

func TestMockTimeProviderDeterminism(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
	}{
		{"epoch", time.Unix(0, 0)},
		{"future", time.Date(2030, 6, 15, 0, 0, 0, 0, time.UTC)},
		{"with nanos", time.Date(2026, 1, 1, 0, 0, 0, 123456789, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := MockTimeProvider{FixedTime: tt.time}
			got := tp.Now()
			if got.Unix() != tt.time.Unix() {
				t.Errorf("Unix() = %d, want %d", got.Unix(), tt.time.Unix())
			}
			if got.UnixNano() != tt.time.UnixNano() {
				t.Errorf("UnixNano() = %d, want %d", got.UnixNano(), tt.time.UnixNano())
			}
		})
	}
}
