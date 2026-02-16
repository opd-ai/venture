package webrtc

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

func TestDefaultTimeProvider(t *testing.T) {
	tp := DefaultTimeProvider()
	if tp == nil {
		t.Fatal("DefaultTimeProvider() returned nil")
	}
	got := tp.Now()
	if got.IsZero() {
		t.Error("DefaultTimeProvider().Now() returned zero time")
	}
}

func TestMockTimeProvider(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock := &MockTimeProvider{CurrentTime: fixedTime}

	if got := mock.Now(); !got.Equal(fixedTime) {
		t.Errorf("MockTimeProvider.Now() = %v, want %v", got, fixedTime)
	}

	newTime := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	mock.SetTime(newTime)
	if got := mock.Now(); !got.Equal(newTime) {
		t.Errorf("After SetTime, Now() = %v, want %v", got, newTime)
	}

	mock.Advance(2 * time.Hour)
	expected := newTime.Add(2 * time.Hour)
	if got := mock.Now(); !got.Equal(expected) {
		t.Errorf("After Advance, Now() = %v, want %v", got, expected)
	}
}

func TestTimeProviderIntegration(t *testing.T) {
	// Verify TimeProvider is properly wired into constructors
	t.Run("Peer", func(t *testing.T) {
		p, err := NewPeer("test", nil)
		if err != nil {
			t.Fatal(err)
		}
		if p.timeProvider == nil {
			t.Error("Peer.timeProvider is nil")
		}
	})

	t.Run("STUNClient", func(t *testing.T) {
		s := NewSTUNClient(nil)
		if s.timeProvider == nil {
			t.Error("STUNClient.timeProvider is nil")
		}
	})

	t.Run("NATTraversal", func(t *testing.T) {
		n := NewNATTraversal(nil, nil)
		if n.timeProvider == nil {
			t.Error("NATTraversal.timeProvider is nil")
		}
	})

	t.Run("RelayNode", func(t *testing.T) {
		r := NewRelayNode("id", "turn:host:3478", "u", "p", "us", 10)
		if r.timeProvider == nil {
			t.Error("RelayNode.timeProvider is nil")
		}
	})

	t.Run("RelayManager", func(t *testing.T) {
		rm := NewRelayManager(StrategyRoundRobin)
		defer rm.Close()
		if rm.timeProvider == nil {
			t.Error("RelayManager.timeProvider is nil")
		}
	})

	t.Run("RelayConnection", func(t *testing.T) {
		node := NewRelayNode("id", "turn:host:3478", "u", "p", "us", 10)
		rc := NewRelayConnection(node, "local", "relay", 5*time.Minute)
		if rc.timeProvider == nil {
			t.Error("RelayConnection.timeProvider is nil")
		}
	})

	t.Run("SignalingClient", func(t *testing.T) {
		sc := NewSignalingClient("ws://test", "peer1")
		if sc.timeProvider == nil {
			t.Error("SignalingClient.timeProvider is nil")
		}
	})
}
