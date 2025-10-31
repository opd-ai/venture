// Package engine provides weather system components for the ECS.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// TestWeatherComponent_Type tests component type identifier.
func TestWeatherComponent_Type(t *testing.T) {
	comp := NewWeatherComponent(particles.DefaultWeatherConfig())
	if got := comp.Type(); got != "weather" {
		t.Errorf("Type() = %v, want %v", got, "weather")
	}
}

// TestNewWeatherComponent tests component initialization.
func TestNewWeatherComponent(t *testing.T) {
	config := particles.DefaultWeatherConfig()
	comp := NewWeatherComponent(config)

	if comp == nil {
		t.Fatal("NewWeatherComponent() returned nil")
	}
	if comp.System != nil {
		t.Error("New component should have nil System (created on activation)")
	}
	if comp.Active {
		t.Error("New component should be inactive")
	}
	if comp.Transitioning {
		t.Error("New component should not be transitioning")
	}
	if comp.TransitionTotal != 5.0 {
		t.Errorf("TransitionTotal = %v, want 5.0", comp.TransitionTotal)
	}
	if comp.Config.Type != config.Type {
		t.Error("Config not stored correctly")
	}
}

// TestWeatherComponent_StartWeather tests weather activation.
func TestWeatherComponent_StartWeather(t *testing.T) {
	config := particles.DefaultWeatherConfig()
	comp := NewWeatherComponent(config)

	err := comp.StartWeather()
	if err != nil {
		t.Fatalf("StartWeather() error = %v", err)
	}

	if !comp.Active {
		t.Error("Weather should be active after StartWeather()")
	}
	if !comp.Transitioning {
		t.Error("Weather should be transitioning after StartWeather()")
	}
	if !comp.FadingIn {
		t.Error("Weather should be fading in after StartWeather()")
	}
	if comp.System == nil {
		t.Error("Weather system should be created by StartWeather()")
	}
	if comp.TransitionTime != 0 {
		t.Error("TransitionTime should be reset to 0")
	}
}

// TestWeatherComponent_StartWeather_AlreadyActive tests idempotency.
func TestWeatherComponent_StartWeather_AlreadyActive(t *testing.T) {
	config := particles.DefaultWeatherConfig()
	comp := NewWeatherComponent(config)

	// Start once
	err := comp.StartWeather()
	if err != nil {
		t.Fatalf("StartWeather() error = %v", err)
	}

	// Complete transition
	comp.Transitioning = false

	// Start again (should be no-op)
	err = comp.StartWeather()
	if err != nil {
		t.Errorf("StartWeather() when already active should not error, got %v", err)
	}
}

// TestWeatherComponent_StopWeather tests weather deactivation.
func TestWeatherComponent_StopWeather(t *testing.T) {
	config := particles.DefaultWeatherConfig()
	comp := NewWeatherComponent(config)

	// Start weather first
	comp.StartWeather()
	comp.Transitioning = false // Complete fade-in

	// Stop weather
	comp.StopWeather()

	if !comp.Transitioning {
		t.Error("Weather should be transitioning after StopWeather()")
	}
	if comp.FadingIn {
		t.Error("Weather should be fading out after StopWeather()")
	}
	if comp.TransitionTime != 0 {
		t.Error("TransitionTime should be reset to 0")
	}
}

// TestWeatherComponent_GetOpacity tests opacity calculation.
func TestWeatherComponent_GetOpacity(t *testing.T) {
	tests := []struct {
		name          string
		active        bool
		transitioning bool
		fadingIn      bool
		transitionPct float64 // Percentage of transition completed
		want          float64
	}{
		{
			name:          "inactive_no_transition",
			active:        false,
			transitioning: false,
			fadingIn:      false,
			transitionPct: 0,
			want:          0.0,
		},
		{
			name:          "active_no_transition",
			active:        true,
			transitioning: false,
			fadingIn:      false,
			transitionPct: 0,
			want:          1.0,
		},
		{
			name:          "fade_in_start",
			active:        true,
			transitioning: true,
			fadingIn:      true,
			transitionPct: 0,
			want:          0.0,
		},
		{
			name:          "fade_in_mid",
			active:        true,
			transitioning: true,
			fadingIn:      true,
			transitionPct: 0.5,
			want:          0.5,
		},
		{
			name:          "fade_in_complete",
			active:        true,
			transitioning: true,
			fadingIn:      true,
			transitionPct: 1.0,
			want:          1.0,
		},
		{
			name:          "fade_out_start",
			active:        true,
			transitioning: true,
			fadingIn:      false,
			transitionPct: 0,
			want:          1.0,
		},
		{
			name:          "fade_out_mid",
			active:        true,
			transitioning: true,
			fadingIn:      false,
			transitionPct: 0.5,
			want:          0.5,
		},
		{
			name:          "fade_out_complete",
			active:        true,
			transitioning: true,
			fadingIn:      false,
			transitionPct: 1.0,
			want:          0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := particles.DefaultWeatherConfig()
			comp := NewWeatherComponent(config)

			comp.Active = tt.active
			comp.Transitioning = tt.transitioning
			comp.FadingIn = tt.fadingIn
			comp.TransitionTotal = 5.0
			comp.TransitionTime = tt.transitionPct * comp.TransitionTotal

			got := comp.GetOpacity()
			if got != tt.want {
				t.Errorf("GetOpacity() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWeatherComponent_UpdateTransition tests transition updates.
func TestWeatherComponent_UpdateTransition(t *testing.T) {
	config := particles.DefaultWeatherConfig()
	comp := NewWeatherComponent(config)

	// Start weather
	comp.StartWeather()

	// Update transition partway
	completed := comp.UpdateTransition(2.5) // 2.5 of 5.0 seconds
	if completed {
		t.Error("Transition should not be complete after 2.5 seconds")
	}
	if !comp.Transitioning {
		t.Error("Should still be transitioning")
	}
	if comp.TransitionTime != 2.5 {
		t.Errorf("TransitionTime = %v, want 2.5", comp.TransitionTime)
	}

	// Complete transition
	completed = comp.UpdateTransition(2.5) // Total 5.0 seconds
	if !completed {
		t.Error("Transition should be complete after 5.0 seconds")
	}
	if comp.Transitioning {
		t.Error("Should not be transitioning after completion")
	}
	if comp.TransitionTime != 0 {
		t.Error("TransitionTime should be reset to 0")
	}
	if !comp.Active {
		t.Error("Should be active after fade-in completes")
	}
}

// TestWeatherComponent_UpdateTransition_FadeOut tests fade-out completion.
func TestWeatherComponent_UpdateTransition_FadeOut(t *testing.T) {
	config := particles.DefaultWeatherConfig()
	comp := NewWeatherComponent(config)

	// Start and complete fade-in
	comp.StartWeather()
	comp.UpdateTransition(5.0)

	// Start fade-out
	comp.StopWeather()

	// Complete fade-out
	completed := comp.UpdateTransition(5.0)
	if !completed {
		t.Error("Transition should be complete")
	}
	if comp.Active {
		t.Error("Should be inactive after fade-out completes")
	}
	if comp.Transitioning {
		t.Error("Should not be transitioning after completion")
	}
}

// TestWeatherComponent_IsFullyActive tests active state check.
func TestWeatherComponent_IsFullyActive(t *testing.T) {
	tests := []struct {
		name          string
		active        bool
		transitioning bool
		want          bool
	}{
		{"inactive", false, false, false},
		{"active_transitioning", true, true, false},
		{"active_complete", true, false, true},
		{"inactive_transitioning", false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := particles.DefaultWeatherConfig()
			comp := NewWeatherComponent(config)

			comp.Active = tt.active
			comp.Transitioning = tt.transitioning

			got := comp.IsFullyActive()
			if got != tt.want {
				t.Errorf("IsFullyActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWeatherComponent_IsFullyInactive tests inactive state check.
func TestWeatherComponent_IsFullyInactive(t *testing.T) {
	tests := []struct {
		name          string
		active        bool
		transitioning bool
		want          bool
	}{
		{"inactive_complete", false, false, true},
		{"active_transitioning", true, true, false},
		{"active_complete", true, false, false},
		{"inactive_transitioning", false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := particles.DefaultWeatherConfig()
			comp := NewWeatherComponent(config)

			comp.Active = tt.active
			comp.Transitioning = tt.transitioning

			got := comp.IsFullyInactive()
			if got != tt.want {
				t.Errorf("IsFullyInactive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWeatherComponent_ChangeWeather tests weather type switching.
func TestWeatherComponent_ChangeWeather(t *testing.T) {
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherRain
	comp := NewWeatherComponent(config)

	// Start with rain
	comp.StartWeather()
	comp.UpdateTransition(5.0) // Complete fade-in

	// Change to snow
	newConfig := config
	newConfig.Type = particles.WeatherSnow
	err := comp.ChangeWeather(newConfig)
	if err != nil {
		t.Fatalf("ChangeWeather() error = %v", err)
	}

	// Should be fading out
	if !comp.Transitioning {
		t.Error("Should be transitioning when changing weather")
	}
	if comp.FadingIn {
		t.Error("Should be fading out when changing from active weather")
	}
	if comp.Config.Type != particles.WeatherSnow {
		t.Error("Config should be updated to new weather type")
	}
}

// TestWeatherComponent_ChangeWeather_Invalid tests invalid config handling.
func TestWeatherComponent_ChangeWeather_Invalid(t *testing.T) {
	config := particles.DefaultWeatherConfig()
	comp := NewWeatherComponent(config)

	// Try to change to invalid config (negative width)
	invalidConfig := config
	invalidConfig.Width = -100
	err := comp.ChangeWeather(invalidConfig)
	if err == nil {
		t.Error("ChangeWeather() with invalid config should return error")
	}
}

// TestWeatherComponent_ChangeWeather_WhenInactive tests changing inactive weather.
func TestWeatherComponent_ChangeWeather_WhenInactive(t *testing.T) {
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherRain
	comp := NewWeatherComponent(config)

	// Change to snow while inactive
	newConfig := config
	newConfig.Type = particles.WeatherSnow
	err := comp.ChangeWeather(newConfig)
	if err != nil {
		t.Fatalf("ChangeWeather() error = %v", err)
	}

	// Should start new weather immediately (fade in)
	if !comp.Active {
		t.Error("Should be active after changing inactive weather")
	}
	if !comp.Transitioning {
		t.Error("Should be transitioning (fade in)")
	}
	if !comp.FadingIn {
		t.Error("Should be fading in when starting new weather")
	}
}
