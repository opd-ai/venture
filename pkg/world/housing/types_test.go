package housing

import (
	"image/color"
	"testing"
	"time"
)

func TestBuildingSizeString(t *testing.T) {
	tests := []struct {
		size     BuildingSize
		expected string
	}{
		{SizeSmall, "Small"},
		{SizeMedium, "Medium"},
		{SizeLarge, "Large"},
		{SizeEstate, "Estate"},
		{BuildingSize(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.size.String(); got != tt.expected {
				t.Errorf("BuildingSize.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildingSizeTiles(t *testing.T) {
	tests := []struct {
		size     BuildingSize
		expected int
	}{
		{SizeSmall, 8},
		{SizeMedium, 16},
		{SizeLarge, 24},
		{SizeEstate, 32},
	}

	for _, tt := range tests {
		t.Run(tt.size.String(), func(t *testing.T) {
			if got := tt.size.Tiles(); got != tt.expected {
				t.Errorf("BuildingSize.Tiles() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildingSizeSquareUnits(t *testing.T) {
	tests := []struct {
		size     BuildingSize
		expected int
	}{
		{SizeSmall, 64},
		{SizeMedium, 256},
		{SizeLarge, 576},
		{SizeEstate, 1024},
	}

	for _, tt := range tests {
		t.Run(tt.size.String(), func(t *testing.T) {
			if got := tt.size.SquareUnits(); got != tt.expected {
				t.Errorf("BuildingSize.SquareUnits() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPermissionLevelString(t *testing.T) {
	tests := []struct {
		perm     PermissionLevel
		expected string
	}{
		{PermissionNone, "None"},
		{PermissionVisit, "Visit"},
		{PermissionFriend, "Friend"},
		{PermissionCoOwner, "CoOwner"},
		{PermissionLevel(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.perm.String(); got != tt.expected {
				t.Errorf("PermissionLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPermissionSet(t *testing.T) {
	ps := NewPermissionSet()

	// Test default permission
	if got := ps.GetPermission("player1"); got != PermissionNone {
		t.Errorf("GetPermission() = %v, want %v", got, PermissionNone)
	}

	// Set permission
	ps.SetPermission("player1", PermissionFriend)
	if got := ps.GetPermission("player1"); got != PermissionFriend {
		t.Errorf("GetPermission() = %v, want %v", got, PermissionFriend)
	}

	// Change default
	ps.DefaultLevel = PermissionVisit
	if got := ps.GetPermission("player2"); got != PermissionVisit {
		t.Errorf("GetPermission() = %v, want %v", got, PermissionVisit)
	}
}

func TestNewPlot(t *testing.T) {
	plot := NewPlot("player1", Vector2{X: 100, Y: 200}, SizeMedium)

	if plot.OwnerID != "player1" {
		t.Errorf("OwnerID = %v, want player1", plot.OwnerID)
	}

	if plot.Position.X != 100 || plot.Position.Y != 200 {
		t.Errorf("Position = (%v, %v), want (100, 200)", plot.Position.X, plot.Position.Y)
	}

	if plot.Size != SizeMedium {
		t.Errorf("Size = %v, want %v", plot.Size, SizeMedium)
	}

	if plot.Permissions == nil {
		t.Error("Permissions should not be nil")
	}

	if plot.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestPlotBounds(t *testing.T) {
	plot := NewPlot("player1", Vector2{X: 100, Y: 200}, SizeMedium)
	min, max := plot.Bounds()

	// Medium is 16 tiles, so half is 8
	expectedMin := Vector2{X: 92, Y: 192}
	expectedMax := Vector2{X: 108, Y: 208}

	if min != expectedMin {
		t.Errorf("min = %v, want %v", min, expectedMin)
	}

	if max != expectedMax {
		t.Errorf("max = %v, want %v", max, expectedMax)
	}
}

func TestPlotContains(t *testing.T) {
	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeSmall)

	tests := []struct {
		name     string
		point    Vector2
		expected bool
	}{
		{"center", Vector2{X: 100, Y: 100}, true},
		{"inside", Vector2{X: 102, Y: 102}, true},
		{"edge", Vector2{X: 104, Y: 104}, true},
		{"outside", Vector2{X: 110, Y: 110}, false},
		{"negative outside", Vector2{X: 90, Y: 90}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := plot.Contains(tt.point); got != tt.expected {
				t.Errorf("Contains(%v) = %v, want %v", tt.point, got, tt.expected)
			}
		})
	}
}

func TestPlotOverlaps(t *testing.T) {
	plot1 := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)

	tests := []struct {
		name     string
		plot2Pos Vector2
		expected bool
	}{
		{"same position", Vector2{X: 100, Y: 100}, true},
		{"adjacent", Vector2{X: 117, Y: 100}, false}, // 17 tiles apart (8+8+1), no overlap
		{"overlap", Vector2{X: 110, Y: 110}, true},
		{"far away", Vector2{X: 200, Y: 200}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plot2 := NewPlot("player2", tt.plot2Pos, SizeMedium)
			if got := plot1.Overlaps(plot2); got != tt.expected {
				t.Errorf("Overlaps() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHousingComponent(t *testing.T) {
	comp := HousingComponent{PlotID: "plot_001"}

	if comp.Type() != "housing" {
		t.Errorf("Type() = %v, want housing", comp.Type())
	}

	if comp.PlotID != "plot_001" {
		t.Errorf("PlotID = %v, want plot_001", comp.PlotID)
	}
}

func TestPlotColor(t *testing.T) {
	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeSmall)

	// Default color should be gray
	expectedColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	if plot.Color != expectedColor {
		t.Errorf("Color = %v, want %v", plot.Color, expectedColor)
	}

	// Should be able to change color
	newColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	plot.Color = newColor
	if plot.Color != newColor {
		t.Errorf("Color = %v, want %v", plot.Color, newColor)
	}
}

func TestSetDefaultTimeProvider(t *testing.T) {
	// Save original and restore after test
	defer ResetDefaultTimeProvider()

	fixedTime := time.Date(2026, 6, 15, 10, 30, 0, 0, time.UTC)
	mockTP := &MockTimeProvider{CurrentTime: fixedTime}
	SetDefaultTimeProvider(mockTP)

	// NewPlot should use the mock time provider
	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeSmall)
	if !plot.CreatedAt.Equal(fixedTime) {
		t.Errorf("CreatedAt = %v, want %v", plot.CreatedAt, fixedTime)
	}
	if !plot.ModifiedAt.Equal(fixedTime) {
		t.Errorf("ModifiedAt = %v, want %v", plot.ModifiedAt, fixedTime)
	}

	// NewBlueprint should also use the mock time provider
	bp := NewBlueprintWithTime("Test Blueprint", "Author", "fantasy", nil, defaultTimeProvider)
	if !bp.CreatedAt.Equal(fixedTime) {
		t.Errorf("Blueprint CreatedAt = %v, want %v", bp.CreatedAt, fixedTime)
	}
	if !bp.ModifiedAt.Equal(fixedTime) {
		t.Errorf("Blueprint ModifiedAt = %v, want %v", bp.ModifiedAt, fixedTime)
	}
}

func TestResetDefaultTimeProvider(t *testing.T) {
	// Set a mock provider
	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	SetDefaultTimeProvider(&MockTimeProvider{CurrentTime: fixedTime})

	// Reset to real time
	ResetDefaultTimeProvider()

	// NewPlot should use real time (not the fixed time)
	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeSmall)
	if plot.CreatedAt.Equal(fixedTime) {
		t.Error("CreatedAt should not equal fixed time after reset")
	}

	// CreatedAt should be close to now (within 1 second)
	if time.Since(plot.CreatedAt) > time.Second {
		t.Error("CreatedAt should be close to current time after reset")
	}
}

func TestTimeProviderDeterminism(t *testing.T) {
	defer ResetDefaultTimeProvider()

	fixedTime := time.Date(2026, 6, 15, 10, 30, 0, 123456789, time.UTC)
	SetDefaultTimeProvider(&MockTimeProvider{CurrentTime: fixedTime})

	// Create two plots - they should have identical timestamps
	plot1 := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	plot2 := NewPlot("player2", Vector2{X: 200, Y: 200}, SizeLarge)

	if !plot1.CreatedAt.Equal(plot2.CreatedAt) {
		t.Errorf("plot1.CreatedAt = %v, plot2.CreatedAt = %v, want equal", plot1.CreatedAt, plot2.CreatedAt)
	}

	// Create two blueprints - they should have identical timestamps
	bp1 := NewBlueprintWithTime("Blueprint1", "Author1", "fantasy", nil, defaultTimeProvider)
	bp2 := NewBlueprintWithTime("Blueprint2", "Author2", "sci-fi", nil, defaultTimeProvider)

	if !bp1.CreatedAt.Equal(bp2.CreatedAt) {
		t.Errorf("bp1.CreatedAt = %v, bp2.CreatedAt = %v, want equal", bp1.CreatedAt, bp2.CreatedAt)
	}
}

func TestNewMockTimeProviderFromSeed(t *testing.T) {
	tests := []struct {
		seed     int64
		expected time.Time
	}{
		{0, time.Unix(0, 0)},
		{1234567890, time.Unix(1234567890, 0)},
		{-1000, time.Unix(-1000, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.expected.String(), func(t *testing.T) {
			mockTP := NewMockTimeProvider(tt.seed)
			if !mockTP.Now().Equal(tt.expected) {
				t.Errorf("NewMockTimeProvider(%d).Now() = %v, want %v", tt.seed, mockTP.Now(), tt.expected)
			}
		})
	}
}
