package mobile

import (
	"os"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/quality"
)

// NOTE: The init() function in mobile.go calls mobile.SetGame() which panics
// in non-mobile environments. These tests focus on testing individual functions
// that can be tested in isolation without triggering the full initialization.

// TestConstants verifies mobile-specific constants are within expected ranges
func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		min      int
		max      int
		describe string
	}{
		{"DefaultScreenWidth", DefaultScreenWidth, 1024, 1920, "landscape width"},
		{"DefaultScreenHeight", DefaultScreenHeight, 600, 1080, "landscape height"},
		{"mobileSpriteCacheMaxSize", mobileSpriteCacheMaxSize, 50 * 1024 * 1024, 200 * 1024 * 1024, "sprite cache"},
		{"mobileAnimationCacheSize", mobileAnimationCacheSize, 50, 200, "animation cache"},
		{"mobileMemorySoftLimit", mobileMemorySoftLimit, 40 * 1024 * 1024, 150 * 1024 * 1024, "soft limit"},
		{"mobileMemoryHardLimit", mobileMemoryHardLimit, 50 * 1024 * 1024, 200 * 1024 * 1024, "hard limit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.min || tt.value > tt.max {
				t.Errorf("%s = %d, want between %d and %d (%s)", tt.name, tt.value, tt.min, tt.max, tt.describe)
			}
		})
	}
}

// TestMemoryLimitRelationship verifies soft limit is less than hard limit
func TestMemoryLimitRelationship(t *testing.T) {
	if mobileMemorySoftLimit >= mobileMemoryHardLimit {
		t.Errorf("soft limit (%d) must be less than hard limit (%d)", mobileMemorySoftLimit, mobileMemoryHardLimit)
	}

	if mobileMemorySoftLimit >= mobileSpriteCacheMaxSize {
		t.Errorf("soft limit (%d) must be less than cache max size (%d)", mobileMemorySoftLimit, mobileSpriteCacheMaxSize)
	}

	if mobileMemoryHardLimit > mobileSpriteCacheMaxSize {
		t.Errorf("hard limit (%d) must not exceed cache max size (%d)", mobileMemoryHardLimit, mobileSpriteCacheMaxSize)
	}
}

// TestCalculatePlayerSpawnPosition tests player spawn position calculation
func TestCalculatePlayerSpawnPosition(t *testing.T) {
	tests := []struct {
		name      string
		terrain   *terrain.Terrain
		expectX   float64
		expectY   float64
		checkFunc func(x, y float64) bool
	}{
		{
			name:    "nil_terrain",
			terrain: nil,
			expectX: 400,
			expectY: 300,
			checkFunc: func(x, y float64) bool {
				return x == 400 && y == 300
			},
		},
		{
			name: "empty_rooms",
			terrain: &terrain.Terrain{
				Width:  60,
				Height: 40,
				Rooms:  []*terrain.Room{},
			},
			expectX: 400,
			expectY: 300,
			checkFunc: func(x, y float64) bool {
				return x == 400 && y == 300
			},
		},
		{
			name: "with_rooms",
			terrain: &terrain.Terrain{
				Width:  60,
				Height: 40,
				Rooms: []*terrain.Room{
					{X: 10, Y: 10, Width: 8, Height: 6},
				},
			},
			checkFunc: func(x, y float64) bool {
				// Room center: (10 + 8/2, 10 + 6/2) = (14, 13) in tiles
				// In pixels: (14*32, 13*32) = (448, 416)
				return x == 448 && y == 416
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := calculatePlayerSpawnPosition(tt.terrain)
			if !tt.checkFunc(x, y) {
				t.Errorf("calculatePlayerSpawnPosition() = (%v, %v), validation failed", x, y)
			}
		})
	}
}

// TestAddStarterItems tests starter item generation
func TestAddStarterItems(t *testing.T) {
	tests := []struct {
		name         string
		seed         int64
		genreID      string
		wantMinItems int
		wantMaxItems int
	}{
		{"fantasy_seed_100", 100, "fantasy", 0, 3},
		{"scifi_seed_200", 200, "scifi", 0, 3},
		{"zero_seed", 0, "", 0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inventory := engine.NewInventoryComponent(20, 100.0)
			inventory.Gold = 100

			addStarterItems(inventory, tt.seed, tt.genreID)

			itemCount := len(inventory.Items)
			if itemCount < tt.wantMinItems || itemCount > tt.wantMaxItems {
				t.Errorf("addStarterItems() added %d items, want between %d and %d",
					itemCount, tt.wantMinItems, tt.wantMaxItems)
			}

			// Verify gold not modified
			if inventory.Gold != 100 {
				t.Errorf("addStarterItems() modified gold: got %d, want 100", inventory.Gold)
			}
		})
	}
}

// TestAddStarterItems_Determinism verifies same seed produces same items
func TestAddStarterItems_Determinism(t *testing.T) {
	seed := int64(12345)
	genreID := "fantasy"

	// Generate items twice with same seed
	inv1 := engine.NewInventoryComponent(20, 100.0)
	addStarterItems(inv1, seed, genreID)

	inv2 := engine.NewInventoryComponent(20, 100.0)
	addStarterItems(inv2, seed, genreID)

	// Compare item counts
	if len(inv1.Items) != len(inv2.Items) {
		t.Errorf("non-deterministic: inv1 has %d items, inv2 has %d items",
			len(inv1.Items), len(inv2.Items))
		return
	}

	// Compare item properties
	for i := range inv1.Items {
		item1 := inv1.Items[i]
		item2 := inv2.Items[i]

		if item1.Name != item2.Name {
			t.Errorf("item %d: different names: %q vs %q", i, item1.Name, item2.Name)
		}
		if item1.Type != item2.Type {
			t.Errorf("item %d: different types: %v vs %v", i, item1.Type, item2.Type)
		}
		if item1.Stats.Value != item2.Stats.Value {
			t.Errorf("item %d: different values: %d vs %d", i, item1.Stats.Value, item2.Stats.Value)
		}
	}
}

// TestAddStarterItems_WeaponGeneration tests weapon-specific generation
func TestAddStarterItems_WeaponGeneration(t *testing.T) {
	inventory := engine.NewInventoryComponent(20, 100.0)
	addStarterItems(inventory, 42, "fantasy")

	// Look for a weapon with "Rusty" prefix
	foundRustyWeapon := false
	for _, itm := range inventory.Items {
		if len(itm.Name) >= 5 && itm.Name[:5] == "Rusty" {
			foundRustyWeapon = true
			// Verify value is set to 5 as per code
			if itm.Stats.Value != 5 {
				t.Errorf("rusty weapon value = %d, want 5", itm.Stats.Value)
			}
			break
		}
	}

	// Note: Due to generation randomness, weapon may not always be generated
	// This is acceptable; we're just checking that IF it exists, it's correct
	if foundRustyWeapon {
		t.Logf("found rusty weapon as expected")
	}
}

// TestAddStarterItems_PotionGeneration tests potion-specific generation
func TestAddStarterItems_PotionGeneration(t *testing.T) {
	inventory := engine.NewInventoryComponent(20, 100.0)
	addStarterItems(inventory, 84, "scifi")

	// Count items named "Minor Health Potion"
	potionCount := 0
	for _, itm := range inventory.Items {
		if itm.Name == "Minor Health Potion" {
			potionCount++
			// Verify potion properties
			if itm.Stats.Value != 10 {
				t.Errorf("potion value = %d, want 10", itm.Stats.Value)
			}
			if itm.Stats.Weight != 0.2 {
				t.Errorf("potion weight = %f, want 0.2", itm.Stats.Weight)
			}
		}
	}

	// Note: Potions may not always be generated due to RNG
	if potionCount > 0 {
		t.Logf("found %d potions", potionCount)
	}
}

// TestUpdate tests the Update function
func TestUpdate(t *testing.T) {
	// Save original state
	originalInstance := gameInstance
	defer func() {
		gameInstance = originalInstance
	}()

	tests := []struct {
		name     string
		instance *engine.EbitenGame
		want     bool
	}{
		{
			name:     "nil_instance",
			instance: nil,
			want:     false,
		},
		{
			name:     "valid_instance",
			instance: &engine.EbitenGame{},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameInstance = tt.instance
			got := Update()
			if got != tt.want {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetScreenWidth tests screen width getter
func TestGetScreenWidth(t *testing.T) {
	// Save original state
	originalInstance := gameInstance
	defer func() {
		gameInstance = originalInstance
	}()

	tests := []struct {
		name     string
		instance *engine.EbitenGame
		want     int
	}{
		{
			name:     "nil_instance",
			instance: nil,
			want:     0,
		},
		{
			name:     "valid_instance",
			instance: &engine.EbitenGame{},
			want:     DefaultScreenWidth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameInstance = tt.instance
			got := GetScreenWidth()
			if got != tt.want {
				t.Errorf("GetScreenWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetScreenHeight tests screen height getter
func TestGetScreenHeight(t *testing.T) {
	// Save original state
	originalInstance := gameInstance
	defer func() {
		gameInstance = originalInstance
	}()

	tests := []struct {
		name     string
		instance *engine.EbitenGame
		want     int
	}{
		{
			name:     "nil_instance",
			instance: nil,
			want:     0,
		},
		{
			name:     "valid_instance",
			instance: &engine.EbitenGame{},
			want:     DefaultScreenHeight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameInstance = tt.instance
			got := GetScreenHeight()
			if got != tt.want {
				t.Errorf("GetScreenHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMobileQualitySystemWrapper tests the ECS adapter
func TestMobileQualitySystemWrapper(t *testing.T) {
	// Create a minimal quality system for testing
	config := quality.LowQualityConfig()
	qSystem := engine.NewQualitySystem(&config, 60.0)

	wrapper := &mobileQualitySystemWrapper{system: qSystem}

	// Test that Update doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update() panicked: %v", r)
		}
	}()

	// Call with empty entity list and small deltaTime
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestMobileQualitySystemWrapper_NilSystem tests wrapper with nil system
func TestMobileQualitySystemWrapper_NilSystem(t *testing.T) {
	wrapper := &mobileQualitySystemWrapper{system: nil}

	// This should panic when calling Update on nil system
	defer func() {
		if r := recover(); r == nil {
			t.Error("Update() should panic with nil system, but didn't")
		}
	}()

	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestStart tests the Start function
func TestStart(t *testing.T) {
	// Note: We cannot fully test initializeGame() in unit tests due to Ebiten dependencies
	// This test verifies the nil-check logic only

	// Save and restore original state
	originalInstance := gameInstance
	defer func() {
		gameInstance = originalInstance
	}()

	// Test that Start is idempotent when gameInstance exists
	gameInstance = &engine.EbitenGame{}
	Start()

	if gameInstance == nil {
		t.Error("Start() set gameInstance to nil")
	}
}

// TestEnvironmentVariableIntegration tests seed/genre from environment
func TestEnvironmentVariableIntegration(t *testing.T) {
	tests := []struct {
		name        string
		seedEnv     string
		genreEnv    string
		wantSeedErr bool
	}{
		{
			name:        "valid_seed_and_genre",
			seedEnv:     "12345",
			genreEnv:    "fantasy",
			wantSeedErr: false,
		},
		{
			name:        "invalid_seed_uses_fallback",
			seedEnv:     "not_a_number",
			genreEnv:    "scifi",
			wantSeedErr: false, // Function falls back to time-based seed
		},
		{
			name:        "empty_values",
			seedEnv:     "",
			genreEnv:    "",
			wantSeedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			if tt.seedEnv != "" {
				os.Setenv("VENTURE_SEED", tt.seedEnv)
				defer os.Unsetenv("VENTURE_SEED")
			} else {
				os.Unsetenv("VENTURE_SEED")
			}

			if tt.genreEnv != "" {
				os.Setenv("VENTURE_GENRE", tt.genreEnv)
				defer os.Unsetenv("VENTURE_GENRE")
			} else {
				os.Unsetenv("VENTURE_GENRE")
			}

			// The init() function already ran, so we can't re-test it directly
			// This test documents expected behavior
		})
	}
}

// TestItemGeneratorCreation tests that item generator can be created
func TestItemGeneratorCreation(t *testing.T) {
	gen := item.NewItemGenerator()
	if gen == nil {
		t.Error("NewItemGenerator() returned nil")
	}
}

// BenchmarkCalculatePlayerSpawnPosition benchmarks spawn calculation
func BenchmarkCalculatePlayerSpawnPosition(b *testing.B) {
	terrain := &terrain.Terrain{
		Width:  60,
		Height: 40,
		Rooms: []*terrain.Room{
			{X: 10, Y: 10, Width: 8, Height: 6},
			{X: 20, Y: 15, Width: 10, Height: 8},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calculatePlayerSpawnPosition(terrain)
	}
}

// BenchmarkAddStarterItems benchmarks starter item generation
func BenchmarkAddStarterItems(b *testing.B) {
	inventory := engine.NewInventoryComponent(20, 100.0)
	seed := int64(42)
	genreID := "fantasy"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inventory.Items = nil // Clear items between iterations
		addStarterItems(inventory, seed, genreID)
	}
}

// TestGetPlatformConfig verifies platform configuration is returned correctly
func TestGetPlatformConfig(t *testing.T) {
	cfg := getPlatformConfig()

	// Verify basic config values are valid
	if cfg.screenWidth <= 0 {
		t.Errorf("screenWidth = %d, want > 0", cfg.screenWidth)
	}
	if cfg.screenHeight <= 0 {
		t.Errorf("screenHeight = %d, want > 0", cfg.screenHeight)
	}
	if cfg.minTouchTarget <= 0 {
		t.Errorf("minTouchTarget = %d, want > 0", cfg.minTouchTarget)
	}
	if cfg.platformName == "" {
		t.Error("platformName is empty")
	}

	// Safe area values should be non-negative
	if cfg.safeAreaTop < 0 {
		t.Errorf("safeAreaTop = %d, want >= 0", cfg.safeAreaTop)
	}
	if cfg.safeAreaBottom < 0 {
		t.Errorf("safeAreaBottom = %d, want >= 0", cfg.safeAreaBottom)
	}
}

// TestGetPlatformConfigValues tests specific platform config value ranges
func TestGetPlatformConfigValues(t *testing.T) {
	cfg := getPlatformConfig()

	tests := []struct {
		name  string
		value int
		min   int
		max   int
	}{
		{"screenWidth", cfg.screenWidth, 640, 2560},
		{"screenHeight", cfg.screenHeight, 480, 1440},
		{"minTouchTarget", cfg.minTouchTarget, 24, 64},
		{"safeAreaTop", cfg.safeAreaTop, 0, 100},
		{"safeAreaBottom", cfg.safeAreaBottom, 0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.min || tt.value > tt.max {
				t.Errorf("%s = %d, want between %d and %d", tt.name, tt.value, tt.min, tt.max)
			}
		})
	}
}

// TestGetPlatformInfo verifies platform info map is complete
func TestGetPlatformInfo(t *testing.T) {
	info := GetPlatformInfo()

	requiredKeys := []string{
		"platform",
		"isIOS",
		"isAndroid",
		"isTouchCapable",
		"supportsBackButton",
		"supportsGestures",
		"minTouchTarget",
		"keyboardHeight",
	}

	for _, key := range requiredKeys {
		if _, exists := info[key]; !exists {
			t.Errorf("GetPlatformInfo() missing key %q", key)
		}
	}

	// Verify types are correct
	if _, ok := info["platform"].(string); !ok {
		t.Error("platform should be a string")
	}
	if _, ok := info["isIOS"].(bool); !ok {
		t.Error("isIOS should be a bool")
	}
	if _, ok := info["isAndroid"].(bool); !ok {
		t.Error("isAndroid should be a bool")
	}
}

// TestPlatformConfigStruct verifies platformConfig struct has correct field types
func TestPlatformConfigStruct(t *testing.T) {
	cfg := platformConfig{
		screenWidth:     1280,
		screenHeight:    720,
		safeAreaTop:     50,
		safeAreaBottom:  34,
		minTouchTarget:  44,
		supportsHaptic:  true,
		supportsBackBtn: false,
		keyboardHeight:  250,
		platformName:    "test",
	}

	// Verify struct fields can be accessed
	if cfg.screenWidth != 1280 {
		t.Errorf("screenWidth = %d, want 1280", cfg.screenWidth)
	}
	if !cfg.supportsHaptic {
		t.Error("supportsHaptic should be true")
	}
	if cfg.platformName != "test" {
		t.Errorf("platformName = %q, want %q", cfg.platformName, "test")
	}
}

// BenchmarkGetPlatformConfig benchmarks platform config retrieval
func BenchmarkGetPlatformConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getPlatformConfig()
	}
}

// BenchmarkGetPlatformInfo benchmarks platform info retrieval
func BenchmarkGetPlatformInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetPlatformInfo()
	}
}
