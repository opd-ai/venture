package guild

import (
	"testing"
	"time"
)

func TestRealTimeProvider(t *testing.T) {
	rtp := RealTimeProvider{}
	before := time.Now()
	got := rtp.Now()
	after := time.Now()

	// Verify the time is within expected bounds
	if got.Before(before) {
		t.Errorf("RealTimeProvider.Now() returned %v, before %v", got, before)
	}
	if got.After(after) {
		t.Errorf("RealTimeProvider.Now() returned %v, after %v", got, after)
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

func TestMockTimeProvider_Now(t *testing.T) {
	fixedTime := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	mockTP := MockTimeProvider{CurrentTime: fixedTime}

	got := mockTP.Now()
	if !got.Equal(fixedTime) {
		t.Errorf("MockTimeProvider.Now() = %v, want %v", got, fixedTime)
	}
}

func TestMockTimeProvider_SetTime(t *testing.T) {
	mockTP := &MockTimeProvider{}

	newTime := time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC)
	mockTP.SetTime(newTime)

	if !mockTP.CurrentTime.Equal(newTime) {
		t.Errorf("After SetTime, CurrentTime = %v, want %v", mockTP.CurrentTime, newTime)
	}
	if !mockTP.Now().Equal(newTime) {
		t.Errorf("After SetTime, Now() = %v, want %v", mockTP.Now(), newTime)
	}
}

func TestMockTimeProvider_Advance(t *testing.T) {
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mockTP := &MockTimeProvider{CurrentTime: startTime}

	tests := []struct {
		name     string
		duration time.Duration
		want     time.Time
	}{
		{"advance 1 hour", time.Hour, startTime.Add(time.Hour)},
		{"advance 30 minutes", 30 * time.Minute, startTime.Add(time.Hour + 30*time.Minute)},
		{"advance 1 day", 24 * time.Hour, startTime.Add(time.Hour + 30*time.Minute + 24*time.Hour)},
		{"advance negative", -time.Hour, startTime.Add(30*time.Minute + 24*time.Hour)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTP.Advance(tt.duration)
			if !mockTP.Now().Equal(tt.want) {
				t.Errorf("After Advance(%v), Now() = %v, want %v", tt.duration, mockTP.Now(), tt.want)
			}
		})
	}
}

func TestMockTimeProvider_ZeroValue(t *testing.T) {
	mockTP := MockTimeProvider{}

	// Zero value should return time.Time zero value
	got := mockTP.Now()
	if !got.IsZero() {
		t.Errorf("Zero MockTimeProvider.Now() = %v, want zero time", got)
	}
}

// TestTimeProviderInterface verifies both implementations satisfy the interface
func TestTimeProviderInterface(t *testing.T) {
	var _ TimeProvider = RealTimeProvider{}
	var _ TimeProvider = &MockTimeProvider{}
	var _ TimeProvider = MockTimeProvider{}
}

func BenchmarkRealTimeProvider(b *testing.B) {
	rtp := RealTimeProvider{}
	for i := 0; i < b.N; i++ {
		_ = rtp.Now()
	}
}

func BenchmarkMockTimeProvider(b *testing.B) {
	mockTP := MockTimeProvider{CurrentTime: time.Now()}
	for i := 0; i < b.N; i++ {
		_ = mockTP.Now()
	}
}

func BenchmarkMockTimeProvider_Advance(b *testing.B) {
	mockTP := &MockTimeProvider{CurrentTime: time.Now()}
	for i := 0; i < b.N; i++ {
		mockTP.Advance(time.Second)
	}
}
