package housing_crafting

import (
	"testing"
)

// TestNewHousingCraftingSystem tests system creation
func TestNewHousingCraftingSystem(t *testing.T) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)

	if system == nil {
		t.Fatal("newHousingCraftingSystem() returned nil")
	}
	if system.manager != manager {
		t.Error("system.manager not set correctly")
	}
}

// TestHousingCraftingSystemGetCraftingBonus tests crafting bonus retrieval
func TestHousingCraftingSystemGetCraftingBonus(t *testing.T) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)

	tests := []struct {
		name      string
		component *HousingCraftingComponent
		want      float64
	}{
		{
			name:      "nil component",
			component: nil,
			want:      1.0,
		},
		{
			name: "zero multiplier",
			component: &HousingCraftingComponent{
				BonusMultiplier: 0,
			},
			want: 1.0,
		},
		{
			name: "negative multiplier",
			component: &HousingCraftingComponent{
				BonusMultiplier: -0.5,
			},
			want: 1.0,
		},
		{
			name: "valid multiplier",
			component: &HousingCraftingComponent{
				BonusMultiplier: 1.5,
			},
			want: 1.5,
		},
		{
			name: "master quality multiplier",
			component: &HousingCraftingComponent{
				BonusMultiplier: 2.0,
			},
			want: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.GetCraftingBonus(tt.component)
			if got != tt.want {
				t.Errorf("GetCraftingBonus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHousingCraftingSystemGetSkillBonus tests skill bonus retrieval
func TestHousingCraftingSystemGetSkillBonus(t *testing.T) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)

	tests := []struct {
		name      string
		component *HousingCraftingComponent
		skillName string
		want      int
	}{
		{
			name:      "nil component",
			component: nil,
			skillName: "smithing",
			want:      0,
		},
		{
			name: "nil skill bonus map",
			component: &HousingCraftingComponent{
				SkillBonus: nil,
			},
			skillName: "smithing",
			want:      0,
		},
		{
			name: "skill not found",
			component: &HousingCraftingComponent{
				SkillBonus: map[string]int{
					"alchemy": 50,
				},
			},
			skillName: "smithing",
			want:      0,
		},
		{
			name: "skill found",
			component: &HousingCraftingComponent{
				SkillBonus: map[string]int{
					"smithing": 50,
					"alchemy":  25,
				},
			},
			skillName: "smithing",
			want:      50,
		},
		{
			name: "empty skill name",
			component: &HousingCraftingComponent{
				SkillBonus: map[string]int{
					"smithing": 50,
				},
			},
			skillName: "",
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.GetSkillBonus(tt.component, tt.skillName)
			if got != tt.want {
				t.Errorf("GetSkillBonus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHousingCraftingSystemHasRecipe tests recipe checking
func TestHousingCraftingSystemHasRecipe(t *testing.T) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)

	tests := []struct {
		name      string
		component *HousingCraftingComponent
		recipeID  string
		want      bool
	}{
		{
			name:      "nil component",
			component: nil,
			recipeID:  "sword_recipe",
			want:      false,
		},
		{
			name: "empty recipes",
			component: &HousingCraftingComponent{
				ActiveRecipes: []string{},
			},
			recipeID: "sword_recipe",
			want:     false,
		},
		{
			name: "recipe not found",
			component: &HousingCraftingComponent{
				ActiveRecipes: []string{"axe_recipe", "shield_recipe"},
			},
			recipeID: "sword_recipe",
			want:     false,
		},
		{
			name: "recipe found first",
			component: &HousingCraftingComponent{
				ActiveRecipes: []string{"sword_recipe", "axe_recipe", "shield_recipe"},
			},
			recipeID: "sword_recipe",
			want:     true,
		},
		{
			name: "recipe found middle",
			component: &HousingCraftingComponent{
				ActiveRecipes: []string{"axe_recipe", "sword_recipe", "shield_recipe"},
			},
			recipeID: "sword_recipe",
			want:     true,
		},
		{
			name: "recipe found last",
			component: &HousingCraftingComponent{
				ActiveRecipes: []string{"axe_recipe", "shield_recipe", "sword_recipe"},
			},
			recipeID: "sword_recipe",
			want:     true,
		},
		{
			name: "empty recipe ID",
			component: &HousingCraftingComponent{
				ActiveRecipes: []string{"sword_recipe"},
			},
			recipeID: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.HasRecipe(tt.component, tt.recipeID)
			if got != tt.want {
				t.Errorf("HasRecipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHousingCraftingSystemSyncFromStation tests syncing from station manager
func TestHousingCraftingSystemSyncFromStation(t *testing.T) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)

	// Register a station
	station := &CraftingStation{
		ID:      "station1",
		Type:    StationTypeForge,
		Quality: QualityMaster,
		OwnerID: "player1",
		HouseID: "house1",
		SkillBonus: map[string]int{
			"smithing": 75,
		},
		ActiveRecipes: []string{"master_sword", "legendary_axe"},
	}
	err := manager.RegisterStation(station)
	if err != nil {
		t.Fatalf("RegisterStation() error = %v", err)
	}

	t.Run("nil component", func(t *testing.T) {
		err := system.SyncFromStation(nil)
		if err != nil {
			t.Errorf("SyncFromStation(nil) error = %v", err)
		}
	})

	t.Run("empty station ID", func(t *testing.T) {
		component := &HousingCraftingComponent{
			StationID: "",
		}
		err := system.SyncFromStation(component)
		if err != nil {
			t.Errorf("SyncFromStation() with empty StationID error = %v", err)
		}
	})

	t.Run("non-existent station", func(t *testing.T) {
		component := &HousingCraftingComponent{
			StationID: "nonexistent",
		}
		err := system.SyncFromStation(component)
		if err == nil {
			t.Error("SyncFromStation() should error on non-existent station")
		}
	})

	t.Run("successful sync", func(t *testing.T) {
		component := &HousingCraftingComponent{
			StationID: "station1",
		}
		err := system.SyncFromStation(component)
		if err != nil {
			t.Errorf("SyncFromStation() error = %v", err)
		}

		// Verify synced values
		if component.StationType != StationTypeForge {
			t.Errorf("StationType = %v, want %v", component.StationType, StationTypeForge)
		}
		if component.BonusMultiplier != 2.0 { // QualityMaster multiplier
			t.Errorf("BonusMultiplier = %v, want 2.0", component.BonusMultiplier)
		}
		if component.SkillBonus["smithing"] != 75 {
			t.Errorf("SkillBonus[smithing] = %v, want 75", component.SkillBonus["smithing"])
		}
		if len(component.ActiveRecipes) != 2 {
			t.Errorf("ActiveRecipes count = %v, want 2", len(component.ActiveRecipes))
		}
	})
}

// TestHousingCraftingSystemSyncFromStationDeepCopy verifies SyncFromStation makes deep copies
func TestHousingCraftingSystemSyncFromStationDeepCopy(t *testing.T) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)

	station := &CraftingStation{
		ID:      "station1",
		Type:    StationTypeForge,
		Quality: QualityMaster,
		OwnerID: "player1",
		HouseID: "house1",
		SkillBonus: map[string]int{
			"smithing": 75,
		},
		ActiveRecipes: []string{"master_sword", "legendary_axe"},
	}
	manager.RegisterStation(station)

	component := &HousingCraftingComponent{StationID: "station1"}
	if err := system.SyncFromStation(component); err != nil {
		t.Fatalf("SyncFromStation() error = %v", err)
	}

	// Modify the station's data after sync
	station.SkillBonus["smithing"] = 999
	station.ActiveRecipes[0] = "changed_recipe"

	// Component should retain original values (deep copy)
	if component.SkillBonus["smithing"] != 75 {
		t.Errorf("SyncFromStation did not deep copy SkillBonus: got %d, want 75", component.SkillBonus["smithing"])
	}
	if component.ActiveRecipes[0] != "master_sword" {
		t.Errorf("SyncFromStation did not deep copy ActiveRecipes: got %s, want master_sword", component.ActiveRecipes[0])
	}
}

// TestHousingCraftingComponentType tests the component Type method
func TestHousingCraftingComponentType(t *testing.T) {
	component := &HousingCraftingComponent{}
	if got := component.Type(); got != "housing_crafting" {
		t.Errorf("Type() = %v, want housing_crafting", got)
	}
}

// TestHousingCraftingSystemConsistency validates ECS pattern compliance
func TestHousingCraftingSystemConsistency(t *testing.T) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)

	// Register a station
	station := &CraftingStation{
		ID:      "test_station",
		Type:    StationTypeAlchemy,
		Quality: QualityAdvanced,
		OwnerID: "player1",
		HouseID: "house1",
		SkillBonus: map[string]int{
			"alchemy": 50,
		},
		ActiveRecipes: []string{"potion1", "potion2"},
	}
	manager.RegisterStation(station)

	// Create component and sync
	component := &HousingCraftingComponent{
		StationID: "test_station",
	}
	err := system.SyncFromStation(component)
	if err != nil {
		t.Fatalf("SyncFromStation() error = %v", err)
	}

	// Verify all system methods work correctly with synced component
	if bonus := system.GetCraftingBonus(component); bonus != 1.5 { // QualityAdvanced multiplier
		t.Errorf("GetCraftingBonus() = %v, want 1.5", bonus)
	}
	if skillBonus := system.GetSkillBonus(component, "alchemy"); skillBonus != 50 {
		t.Errorf("GetSkillBonus(alchemy) = %v, want 50", skillBonus)
	}
	if !system.HasRecipe(component, "potion1") {
		t.Error("HasRecipe(potion1) = false, want true")
	}
	if system.HasRecipe(component, "nonexistent") {
		t.Error("HasRecipe(nonexistent) = true, want false")
	}
}

// Benchmark tests

func BenchmarkHousingCraftingSystemGetCraftingBonus(b *testing.B) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)
	component := &HousingCraftingComponent{
		BonusMultiplier: 1.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.GetCraftingBonus(component)
	}
}

func BenchmarkHousingCraftingSystemGetSkillBonus(b *testing.B) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)
	component := &HousingCraftingComponent{
		SkillBonus: map[string]int{
			"smithing": 50,
			"alchemy":  25,
			"cooking":  10,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.GetSkillBonus(component, "smithing")
	}
}

func BenchmarkHousingCraftingSystemHasRecipe(b *testing.B) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)
	component := &HousingCraftingComponent{
		ActiveRecipes: []string{"recipe1", "recipe2", "recipe3", "recipe4", "recipe5"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.HasRecipe(component, "recipe3")
	}
}

func BenchmarkHousingCraftingSystemSyncFromStation(b *testing.B) {
	manager := NewStationManager()
	system := newHousingCraftingSystem(manager)

	station := &CraftingStation{
		ID:      "station1",
		Type:    StationTypeForge,
		Quality: QualityMaster,
		OwnerID: "player1",
		HouseID: "house1",
		SkillBonus: map[string]int{
			"smithing": 75,
		},
		ActiveRecipes: []string{"master_sword", "legendary_axe"},
	}
	manager.RegisterStation(station)

	component := &HousingCraftingComponent{
		StationID: "station1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.SyncFromStation(component)
	}
}
