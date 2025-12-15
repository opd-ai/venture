package engine

import (
	"encoding/json"
	"testing"
)

func TestNewStereoscopicComponent(t *testing.T) {
	c := NewStereoscopicComponent()

	if c == nil {
		t.Fatal("NewStereoscopicComponent returned nil")
	}

	if c.Enabled {
		t.Error("Expected Enabled to be false by default")
	}

	if c.IPD != DefaultIPD {
		t.Errorf("Expected IPD %v, got %v", DefaultIPD, c.IPD)
	}

	if c.Convergence != DefaultConvergence {
		t.Errorf("Expected Convergence %v, got %v", DefaultConvergence, c.Convergence)
	}

	if !c.BarrelDistortion {
		t.Error("Expected BarrelDistortion to be true by default")
	}

	if c.CurrentEye != EyeLeft {
		t.Errorf("Expected CurrentEye %v, got %v", EyeLeft, c.CurrentEye)
	}

	// Check eye separation is calculated from IPD
	expectedSeparation := (DefaultIPD / 1000.0) / 2.0
	if c.EyeSeparation != expectedSeparation {
		t.Errorf("Expected EyeSeparation %v, got %v", expectedSeparation, c.EyeSeparation)
	}
}

func TestStereoscopicComponent_Type(t *testing.T) {
	c := NewStereoscopicComponent()
	if c.Type() != "stereoscopic" {
		t.Errorf("Expected type 'stereoscopic', got '%s'", c.Type())
	}
}

func TestStereoscopicComponent_SetIPD(t *testing.T) {
	tests := []struct {
		name        string
		inputIPD    float64
		expectedIPD float64
	}{
		{"normal value", 65.0, 65.0},
		{"minimum clamp", 50.0, MinIPD},
		{"maximum clamp", 80.0, MaxIPD},
		{"at minimum", MinIPD, MinIPD},
		{"at maximum", MaxIPD, MaxIPD},
		{"negative clamp", -10.0, MinIPD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewStereoscopicComponent()
			c.SetIPD(tt.inputIPD)

			if c.GetIPD() != tt.expectedIPD {
				t.Errorf("Expected IPD %v, got %v", tt.expectedIPD, c.GetIPD())
			}

			// Verify eye separation updated
			expectedSeparation := (tt.expectedIPD / 1000.0) / 2.0
			if c.EyeSeparation != expectedSeparation {
				t.Errorf("Expected EyeSeparation %v, got %v", expectedSeparation, c.EyeSeparation)
			}
		})
	}
}

func TestStereoscopicComponent_GetEyeOffset(t *testing.T) {
	c := NewStereoscopicComponent()
	c.SetIPD(64.0) // 64mm IPD

	expectedSeparation := (64.0 / 1000.0) / 2.0 // 0.032 meters

	leftOffset := c.GetEyeOffset(EyeLeft)
	rightOffset := c.GetEyeOffset(EyeRight)

	if leftOffset != -expectedSeparation {
		t.Errorf("Expected left offset %v, got %v", -expectedSeparation, leftOffset)
	}

	if rightOffset != expectedSeparation {
		t.Errorf("Expected right offset %v, got %v", expectedSeparation, rightOffset)
	}

	// Left should be negative, right should be positive
	if leftOffset >= 0 {
		t.Error("Left eye offset should be negative")
	}

	if rightOffset <= 0 {
		t.Error("Right eye offset should be positive")
	}
}

func TestStereoscopicComponent_SetEnabled(t *testing.T) {
	c := NewStereoscopicComponent()

	if c.IsEnabled() {
		t.Error("Expected disabled by default")
	}

	c.SetEnabled(true)
	if !c.IsEnabled() {
		t.Error("Expected enabled after SetEnabled(true)")
	}

	c.SetEnabled(false)
	if c.IsEnabled() {
		t.Error("Expected disabled after SetEnabled(false)")
	}
}

func TestStereoscopicComponent_SetConvergence(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"normal value", 15.0, 15.0},
		{"minimum clamp", 0.01, 0.1},
		{"zero clamp", 0.0, 0.1},
		{"negative clamp", -5.0, 0.1},
		{"at minimum", 0.1, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewStereoscopicComponent()
			c.SetConvergence(tt.input)

			if c.GetConvergence() != tt.expected {
				t.Errorf("Expected Convergence %v, got %v", tt.expected, c.GetConvergence())
			}
		})
	}
}

func TestStereoscopicComponent_SetDistortion(t *testing.T) {
	c := NewStereoscopicComponent()

	c.SetDistortion(0.3, 0.4)

	k1, k2 := c.GetDistortion()
	if k1 != 0.3 {
		t.Errorf("Expected K1 0.3, got %v", k1)
	}
	if k2 != 0.4 {
		t.Errorf("Expected K2 0.4, got %v", k2)
	}
}

func TestStereoscopicComponent_SetBarrelDistortion(t *testing.T) {
	c := NewStereoscopicComponent()

	if !c.IsBarrelDistortionEnabled() {
		t.Error("Expected barrel distortion enabled by default")
	}

	c.SetBarrelDistortion(false)
	if c.IsBarrelDistortionEnabled() {
		t.Error("Expected barrel distortion disabled")
	}

	c.SetBarrelDistortion(true)
	if !c.IsBarrelDistortionEnabled() {
		t.Error("Expected barrel distortion enabled")
	}
}

func TestStereoscopicComponent_SetRenderSize(t *testing.T) {
	tests := []struct {
		name             string
		inputW, inputH   int
		expectW, expectH int
	}{
		{"normal size", 1920, 1080, 1920, 1080},
		{"zero width clamp", 0, 1080, 1, 1080},
		{"zero height clamp", 1920, 0, 1920, 1},
		{"negative clamp", -100, -200, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewStereoscopicComponent()
			c.SetRenderSize(tt.inputW, tt.inputH)

			w, h := c.GetRenderSize()
			if w != tt.expectW {
				t.Errorf("Expected width %v, got %v", tt.expectW, w)
			}
			if h != tt.expectH {
				t.Errorf("Expected height %v, got %v", tt.expectH, h)
			}
		})
	}
}

func TestStereoscopicComponent_SetCurrentEye(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"left eye", EyeLeft, EyeLeft},
		{"right eye", EyeRight, EyeRight},
		{"invalid defaults to left", "invalid", EyeLeft},
		{"empty defaults to left", "", EyeLeft},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewStereoscopicComponent()
			c.SetCurrentEye(tt.input)

			if c.GetCurrentEye() != tt.expected {
				t.Errorf("Expected eye %v, got %v", tt.expected, c.GetCurrentEye())
			}
		})
	}
}

func TestStereoscopicComponent_ApplyBarrelDistortion(t *testing.T) {
	c := NewStereoscopicComponent()

	// Test with distortion disabled
	c.SetBarrelDistortion(false)
	x, y := c.ApplyBarrelDistortion(0.5, 0.5)
	if x != 0.5 || y != 0.5 {
		t.Error("With distortion disabled, coordinates should be unchanged")
	}

	// Test with distortion enabled
	c.SetBarrelDistortion(true)
	c.SetDistortion(0.22, 0.24)

	// At center (0, 0), distortion should be identity
	x, y = c.ApplyBarrelDistortion(0.0, 0.0)
	if x != 0.0 || y != 0.0 {
		t.Error("At center, distortion should return (0, 0)")
	}

	// At edge, distortion should magnify
	x, y = c.ApplyBarrelDistortion(1.0, 0.0)
	if x <= 1.0 {
		t.Errorf("At edge, X should be magnified, got %v", x)
	}

	// Test symmetry
	x1, y1 := c.ApplyBarrelDistortion(0.5, 0.3)
	x2, y2 := c.ApplyBarrelDistortion(-0.5, -0.3)
	if x1 != -x2 || y1 != -y2 {
		t.Error("Distortion should be symmetric around center")
	}
}

func TestStereoscopicComponent_ApplyBarrelDistortion_Deterministic(t *testing.T) {
	c := NewStereoscopicComponent()
	c.SetBarrelDistortion(true)
	c.SetDistortion(0.22, 0.24)

	// Run multiple times with same input
	for i := 0; i < 100; i++ {
		x1, y1 := c.ApplyBarrelDistortion(0.7, 0.4)
		x2, y2 := c.ApplyBarrelDistortion(0.7, 0.4)

		if x1 != x2 || y1 != y2 {
			t.Errorf("Distortion should be deterministic: (%v,%v) != (%v,%v)", x1, y1, x2, y2)
		}
	}
}

func TestStereoscopicComponent_Serialize(t *testing.T) {
	c := NewStereoscopicComponent()
	c.SetEnabled(true)
	c.SetIPD(65.0)
	c.SetConvergence(12.0)
	c.SetDistortion(0.25, 0.30)
	c.SetCurrentEye(EyeRight)

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Verify it's valid JSON
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Serialized data is not valid JSON: %v", err)
	}

	// Verify key fields
	if m["enabled"] != true {
		t.Error("Expected enabled to be true in serialized data")
	}
	if m["ipd"].(float64) != 65.0 {
		t.Errorf("Expected ipd 65.0, got %v", m["ipd"])
	}
}

func TestStereoscopicComponent_Deserialize(t *testing.T) {
	original := NewStereoscopicComponent()
	original.SetEnabled(true)
	original.SetIPD(68.0)
	original.SetConvergence(15.0)
	original.SetDistortion(0.28, 0.32)
	original.SetBarrelDistortion(false)
	original.SetRenderSize(1280, 720)
	original.SetCurrentEye(EyeRight)

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	restored := NewStereoscopicComponent()
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify all fields restored
	if restored.IsEnabled() != original.IsEnabled() {
		t.Errorf("Enabled mismatch: %v != %v", restored.IsEnabled(), original.IsEnabled())
	}
	if restored.GetIPD() != original.GetIPD() {
		t.Errorf("IPD mismatch: %v != %v", restored.GetIPD(), original.GetIPD())
	}
	if restored.GetConvergence() != original.GetConvergence() {
		t.Errorf("Convergence mismatch: %v != %v", restored.GetConvergence(), original.GetConvergence())
	}

	k1o, k2o := original.GetDistortion()
	k1r, k2r := restored.GetDistortion()
	if k1o != k1r || k2o != k2r {
		t.Errorf("Distortion mismatch: (%v,%v) != (%v,%v)", k1o, k2o, k1r, k2r)
	}

	if restored.IsBarrelDistortionEnabled() != original.IsBarrelDistortionEnabled() {
		t.Error("BarrelDistortion mismatch")
	}

	wo, ho := original.GetRenderSize()
	wr, hr := restored.GetRenderSize()
	if wo != wr || ho != hr {
		t.Errorf("RenderSize mismatch: (%v,%v) != (%v,%v)", wo, ho, wr, hr)
	}

	if restored.GetCurrentEye() != original.GetCurrentEye() {
		t.Errorf("CurrentEye mismatch: %v != %v", restored.GetCurrentEye(), original.GetCurrentEye())
	}
}

func TestStereoscopicComponent_ThreadSafety(t *testing.T) {
	c := NewStereoscopicComponent()

	// Run concurrent operations
	done := make(chan bool, 4)

	go func() {
		for i := 0; i < 1000; i++ {
			c.SetIPD(float64(55 + i%20))
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = c.GetIPD()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.SetEnabled(i%2 == 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = c.GetEyeOffset(EyeLeft)
			_ = c.GetEyeOffset(EyeRight)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}
}

func BenchmarkStereoscopicComponent_ApplyBarrelDistortion(b *testing.B) {
	c := NewStereoscopicComponent()
	c.SetBarrelDistortion(true)
	c.SetDistortion(0.22, 0.24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ApplyBarrelDistortion(0.7, 0.4)
	}
}

func BenchmarkStereoscopicComponent_GetEyeOffset(b *testing.B) {
	c := NewStereoscopicComponent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GetEyeOffset(EyeLeft)
		c.GetEyeOffset(EyeRight)
	}
}
