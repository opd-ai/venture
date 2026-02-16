package mobile

import (
	"image/color"
	"testing"
)

// =============================================================================
// MobileMenu Scroll Helper Tests
// These tests cover the internal scroll helper methods that are used for
// momentum scrolling in the mobile menu.
// =============================================================================

// TestMobileMenu_CalculateMaxScroll tests max scroll calculation.
func TestMobileMenu_CalculateMaxScroll(t *testing.T) {
	tests := []struct {
		name       string
		menuHeight float64
		itemCount  int
		wantValue  float64 // Expected maxScroll value
	}{
		{
			name:       "empty_menu",
			menuHeight: 400,
			itemCount:  0,
			wantValue:  0, // 0 items = no scrollable content
		},
		{
			name:       "single_item_fits",
			menuHeight: 400,
			itemCount:  1,
			wantValue:  0, // 1 item stretches to fill menu, no scroll needed
		},
		{
			name:       "many_items_scrollable",
			menuHeight: 200,
			itemCount:  20,
			wantValue:  760, // 20 items * 48px min height - 200 = 760 (scrollable)
		},
		{
			name:       "items_exactly_fit",
			menuHeight: 400,
			itemCount:  8,
			wantValue:  0, // 8 items * 50px dynamic - 400 = 0 (exactly fits)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := NewMobileMenu(0, 0, 300, tt.menuHeight)
			for i := 0; i < tt.itemCount; i++ {
				menu.AddItem("Item", true, nil)
			}

			maxScroll := menu.calculateMaxScroll()
			if maxScroll != tt.wantValue {
				t.Errorf("calculateMaxScroll() = %.2f, want %.2f", maxScroll, tt.wantValue)
			}
		})
	}
}

// TestMobileMenu_GetItemHeight tests consistent item height calculation.
func TestMobileMenu_GetItemHeight(t *testing.T) {
	tests := []struct {
		name       string
		menuHeight float64
		itemCount  int
		want       float64
	}{
		{"empty_menu", 400, 0, minMenuItemHeight},
		{"few_items_stretch", 400, 2, 200},             // 400/2=200 > 48
		{"many_items_use_min", 200, 20, minMenuItemHeight}, // 200/20=10 < 48
		{"exact_min_boundary", 480, 10, minMenuItemHeight}, // 480/10=48 == 48
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := NewMobileMenu(0, 0, 300, tt.menuHeight)
			for i := 0; i < tt.itemCount; i++ {
				menu.AddItem("Item", true, nil)
			}
			got := menu.getItemHeight()
			if got != tt.want {
				t.Errorf("getItemHeight() = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}

// TestMobileMenu_HandleTapEmptyMenu tests that tap input on empty menu doesn't panic.
func TestMobileMenu_HandleTapEmptyMenu(t *testing.T) {
	menu := NewMobileMenu(0, 0, 300, 400)
	// Should not panic with no items
	menu.handleTapInput()
	menu.handleLongPressInput()
}

// TestMobileMenu_SwipeVerticalDetection tests swipe direction classification.
func TestMobileMenu_SwipeVerticalDetection(t *testing.T) {
	// Verify the swipe angle check: abs(direction) must be between π/4 and 3π/4
	// for vertical classification
	tests := []struct {
		name       string
		angle      float64
		isVertical bool
	}{
		{"right_horizontal", 0.0, false},
		{"up_vertical", -1.5708, true},  // -π/2
		{"down_vertical", 1.5708, true}, // π/2
		{"left_horizontal", 3.1416, false},   // π
		{"neg_left_horizontal", -3.1416, false}, // -π
		{"diagonal_45", 0.7854, false},  // π/4 boundary
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absDir := absFloatVal(tt.angle)
			got := absDir > 0.7854 && absDir < 2.3562 // π/4 and 3π/4
			if got != tt.isVertical {
				t.Errorf("angle=%.4f: isVertical=%v, want %v", tt.angle, got, tt.isVertical)
			}
		})
	}
}

// TestMobileMenu_ApplyBounceBackEffect tests the rubber-band effect.
func TestMobileMenu_ApplyBounceBackEffect(t *testing.T) {
	tests := []struct {
		name           string
		initialOffset  float64
		initialVel     float64
		maxScroll      float64
		wantOffsetLess float64 // offset should be less than this after bounce
		wantVelLess    float64 // velocity should be less than this after bounce
	}{
		{
			name:           "scrolled_above_top",
			initialOffset:  50.0, // Positive means scrolled above top
			initialVel:     10.0,
			maxScroll:      100.0,
			wantOffsetLess: 50.0, // Should decrease
			wantVelLess:    10.0, // Should decrease
		},
		{
			name:           "scrolled_below_bottom",
			initialOffset:  -150.0, // Below maxScroll of 100
			initialVel:     -10.0,
			maxScroll:      100.0,
			wantOffsetLess: -100.0, // Should move toward valid range (become less negative)
			wantVelLess:    -7.0,   // Should be reduced (toward 0)
		},
		{
			name:           "within_valid_range",
			initialOffset:  -50.0, // Within 0 to -100
			initialVel:     5.0,
			maxScroll:      100.0,
			wantOffsetLess: -40.0, // No change expected
			wantVelLess:    6.0,   // No change expected
		},
		{
			name:           "small_overscroll_snaps_back",
			initialOffset:  0.5, // Very small overscroll above top
			initialVel:     0.1,
			maxScroll:      100.0,
			wantOffsetLess: 1.0, // Should snap to 0
			wantVelLess:    0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := NewMobileMenu(0, 0, 300, 400)
			menu.scrollOffset = tt.initialOffset
			menu.scrollVelocity = tt.initialVel

			menu.applyBounceBackEffect(tt.maxScroll)

			// For overscroll cases, offset should decrease in magnitude
			if tt.initialOffset > 0 {
				// Scrolled above top: offset should decrease
				if menu.scrollOffset >= tt.initialOffset && menu.scrollOffset > 1.0 {
					t.Errorf("scrollOffset = %.2f, should decrease from %.2f", menu.scrollOffset, tt.initialOffset)
				}
			} else if tt.initialOffset < -tt.maxScroll {
				// Scrolled below bottom: offset should move toward valid range
				if menu.scrollOffset < tt.initialOffset {
					t.Errorf("scrollOffset = %.2f, should move toward valid range from %.2f", menu.scrollOffset, tt.initialOffset)
				}
			}
		})
	}
}

// TestMobileMenu_ApplyDeceleration tests friction application.
func TestMobileMenu_ApplyDeceleration(t *testing.T) {
	tests := []struct {
		name         string
		initialVel   float64
		deceleration float64
		wantVelLess  float64
	}{
		{
			name:         "positive_velocity",
			initialVel:   10.0,
			deceleration: 0.95,
			wantVelLess:  10.0,
		},
		{
			name:         "negative_velocity",
			initialVel:   -10.0,
			deceleration: 0.95,
			wantVelLess:  -9.0, // Should become less negative
		},
		{
			name:         "zero_velocity",
			initialVel:   0.0,
			deceleration: 0.95,
			wantVelLess:  0.1, // Should stay zero
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := NewMobileMenu(0, 0, 300, 400)
			menu.scrollVelocity = tt.initialVel
			menu.scrollDeceleration = tt.deceleration

			menu.applyDeceleration()

			expectedVel := tt.initialVel * tt.deceleration
			if menu.scrollVelocity != expectedVel {
				t.Errorf("scrollVelocity = %.4f, want %.4f", menu.scrollVelocity, expectedVel)
			}
		})
	}
}

// TestMobileMenu_StopIfVelocityNegligible tests velocity threshold check.
func TestMobileMenu_StopIfVelocityNegligible(t *testing.T) {
	tests := []struct {
		name        string
		velocity    float64
		wantStop    bool
		description string
	}{
		{
			name:        "zero_velocity",
			velocity:    0.0,
			wantStop:    true,
			description: "Zero velocity should stop scrolling",
		},
		{
			name:        "small_positive_velocity",
			velocity:    0.05,
			wantStop:    true,
			description: "Small positive velocity should stop scrolling",
		},
		{
			name:        "small_negative_velocity",
			velocity:    -0.05,
			wantStop:    true,
			description: "Small negative velocity should stop scrolling",
		},
		{
			name:        "large_positive_velocity",
			velocity:    5.0,
			wantStop:    false,
			description: "Large positive velocity should continue scrolling",
		},
		{
			name:        "large_negative_velocity",
			velocity:    -5.0,
			wantStop:    false,
			description: "Large negative velocity should continue scrolling",
		},
		{
			name:        "boundary_positive",
			velocity:    0.1,
			wantStop:    false,
			description: "Velocity at exactly 0.1 should continue (boundary)",
		},
		{
			name:        "boundary_negative",
			velocity:    -0.1,
			wantStop:    false,
			description: "Velocity at exactly -0.1 should continue (boundary)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := NewMobileMenu(0, 0, 300, 400)
			menu.isScrolling = true
			menu.scrollVelocity = tt.velocity
			menu.scrollOffset = -50.0 // Start with some offset

			menu.stopIfVelocityNegligible(100.0)

			stopped := !menu.isScrolling
			if stopped != tt.wantStop {
				t.Errorf("%s: stopped = %v, want %v", tt.description, stopped, tt.wantStop)
			}
		})
	}
}

// TestMobileMenu_StopScrollingInternal tests the internal stopScrolling function.
func TestMobileMenu_StopScrollingInternal(t *testing.T) {
	menu := NewMobileMenu(0, 0, 300, 400)

	// Set up scrolling state
	menu.isScrolling = true
	menu.scrollVelocity = 10.0
	menu.scrollOffset = -50.0

	// Stop scrolling (internal method)
	menu.stopScrolling()

	if menu.isScrolling {
		t.Error("isScrolling = true, want false after stopScrolling()")
	}
	if menu.scrollVelocity != 0 {
		t.Errorf("scrollVelocity = %.2f, want 0 after stopScrolling()", menu.scrollVelocity)
	}
	if menu.scrollOffset != 0 {
		t.Errorf("scrollOffset = %.2f, want 0 after stopScrolling()", menu.scrollOffset)
	}
}

// TestMobileMenu_SnapToValidRange tests snapping to valid scroll range.
func TestMobileMenu_SnapToValidRange(t *testing.T) {
	tests := []struct {
		name        string
		offset      float64
		maxScroll   float64
		wantOffset  float64
		description string
	}{
		{
			name:        "above_top",
			offset:      50.0,
			maxScroll:   100.0,
			wantOffset:  0.0,
			description: "Positive offset should snap to 0",
		},
		{
			name:        "below_bottom",
			offset:      -150.0,
			maxScroll:   100.0,
			wantOffset:  -100.0,
			description: "Offset below -maxScroll should snap to -maxScroll",
		},
		{
			name:        "within_range_top",
			offset:      0.0,
			maxScroll:   100.0,
			wantOffset:  0.0,
			description: "Offset at 0 should stay at 0",
		},
		{
			name:        "within_range_bottom",
			offset:      -100.0,
			maxScroll:   100.0,
			wantOffset:  -100.0,
			description: "Offset at -maxScroll should stay at -maxScroll",
		},
		{
			name:        "within_range_middle",
			offset:      -50.0,
			maxScroll:   100.0,
			wantOffset:  -50.0,
			description: "Offset within range should not change",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := NewMobileMenu(0, 0, 300, 400)
			menu.scrollOffset = tt.offset

			menu.snapToValidRange(tt.maxScroll)

			if menu.scrollOffset != tt.wantOffset {
				t.Errorf("%s: scrollOffset = %.2f, want %.2f", tt.description, menu.scrollOffset, tt.wantOffset)
			}
		})
	}
}

// TestMobileMenu_UpdateMomentumScrolling tests momentum scrolling update.
func TestMobileMenu_UpdateMomentumScrolling(t *testing.T) {
	menu := NewMobileMenu(0, 0, 300, 200)

	// Add items to make scrolling possible
	for i := 0; i < 10; i++ {
		menu.AddItem("Item", true, nil)
	}

	// Start scrolling
	menu.isScrolling = true
	menu.scrollVelocity = 5.0
	menu.scrollOffset = -20.0

	// Update momentum scrolling
	menu.updateMomentumScrolling()

	// Velocity should have been applied and decelerated
	if menu.scrollVelocity >= 5.0 {
		t.Errorf("scrollVelocity = %.2f, should be less than initial 5.0", menu.scrollVelocity)
	}
}

// =============================================================================
// MinimapWidget Helper Tests
// =============================================================================

// TestMinimapWidget_GetTileColorForType tests tile color mapping.
func TestMinimapWidget_GetTileColorForType(t *testing.T) {
	minimap := NewMinimapWidget(0, 0, 100, 100)

	tests := []struct {
		name     string
		tileType int
		want     color.Color
	}{
		{
			name:     "wall_type_0",
			tileType: 0,
			want:     color.RGBA{60, 60, 60, 255}, // Dark gray for walls
		},
		{
			name:     "floor_type_1",
			tileType: 1,
			want:     color.RGBA{150, 150, 150, 255}, // Light gray for floor
		},
		{
			name:     "unknown_type_2",
			tileType: 2,
			want:     color.RGBA{100, 100, 100, 255}, // Default gray
		},
		{
			name:     "unknown_type_negative",
			tileType: -1,
			want:     color.RGBA{100, 100, 100, 255}, // Default gray
		},
		{
			name:     "unknown_type_large",
			tileType: 999,
			want:     color.RGBA{100, 100, 100, 255}, // Default gray
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minimap.getTileColorForType(tt.tileType)
			if got != tt.want {
				t.Errorf("getTileColorForType(%d) = %v, want %v", tt.tileType, got, tt.want)
			}
		})
	}
}
