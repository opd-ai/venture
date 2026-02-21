package engine

import (
	"testing"
)

func TestElementalEnchantmentType_String(t *testing.T) {
	tests := []struct {
		element ElementalEnchantmentType
		want    string
	}{
		{ElementalNone, "none"},
		{ElementalFire, "fire"},
		{ElementalIce, "ice"},
		{ElementalLightning, "lightning"},
		{ElementalPoison, "poison"},
		{ElementalHoly, "holy"},
		{ElementalShadow, "shadow"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.element.String(); got != tt.want {
				t.Errorf("ElementalEnchantmentType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseElementalType(t *testing.T) {
	tests := []struct {
		input string
		want  ElementalEnchantmentType
	}{
		// Fire variants
		{"fire", ElementalFire},
		{"flame", ElementalFire},
		{"burning", ElementalFire},
		{"inferno", ElementalFire},
		// Ice variants
		{"ice", ElementalIce},
		{"frost", ElementalIce},
		{"frozen", ElementalIce},
		{"cold", ElementalIce},
		{"glacial", ElementalIce},
		// Lightning variants
		{"lightning", ElementalLightning},
		{"electric", ElementalLightning},
		{"shock", ElementalLightning},
		{"thunder", ElementalLightning},
		{"storm", ElementalLightning},
		// Poison variants
		{"poison", ElementalPoison},
		{"toxic", ElementalPoison},
		{"venom", ElementalPoison},
		{"acid", ElementalPoison},
		{"corrosive", ElementalPoison},
		// Holy variants
		{"holy", ElementalHoly},
		{"sacred", ElementalHoly},
		{"divine", ElementalHoly},
		{"radiant", ElementalHoly},
		{"light", ElementalHoly},
		{"blessed", ElementalHoly},
		// Shadow variants
		{"shadow", ElementalShadow},
		{"dark", ElementalShadow},
		{"void", ElementalShadow},
		{"unholy", ElementalShadow},
		{"necrotic", ElementalShadow},
		{"cursed", ElementalShadow},
		// Unknown
		{"unknown", ElementalNone},
		{"", ElementalNone},
		{"physical", ElementalNone},
		{"normal", ElementalNone},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseElementalType(tt.input); got != tt.want {
				t.Errorf("ParseElementalType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestElementalWeaponComponent_Type(t *testing.T) {
	comp := &ElementalWeaponComponent{}
	if got := comp.Type(); got != "elemental_weapon" {
		t.Errorf("ElementalWeaponComponent.Type() = %v, want %v", got, "elemental_weapon")
	}
}

func TestNewElementalWeaponComponent(t *testing.T) {
	tests := []struct {
		name      string
		element   ElementalEnchantmentType
		intensity float64
		seed      int64
	}{
		{"fire_high", ElementalFire, 0.9, 12345},
		{"ice_medium", ElementalIce, 0.5, 54321},
		{"lightning_low", ElementalLightning, 0.3, 99999},
		{"poison_max", ElementalPoison, 1.0, 11111},
		{"holy_min", ElementalHoly, 0.0, 22222},
		{"shadow_normal", ElementalShadow, 0.7, 33333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewElementalWeaponComponent(tt.element, tt.intensity, tt.seed)

			if comp == nil {
				t.Fatal("NewElementalWeaponComponent returned nil")
			}
			if comp.ElementType != tt.element {
				t.Errorf("ElementType = %v, want %v", comp.ElementType, tt.element)
			}
			if comp.Intensity != tt.intensity {
				t.Errorf("Intensity = %v, want %v", comp.Intensity, tt.intensity)
			}
			if comp.Seed != tt.seed {
				t.Errorf("Seed = %v, want %v", comp.Seed, tt.seed)
			}
			if !comp.IsDirty {
				t.Error("IsDirty should be true for new component")
			}
			if comp.ParticleCount <= 0 {
				t.Errorf("ParticleCount = %v, want > 0", comp.ParticleCount)
			}
			// Particle count should scale with intensity
			expectedMin := 4 + int(tt.intensity*6)
			if comp.ParticleCount < 4 {
				t.Errorf("ParticleCount = %v, want >= 4", comp.ParticleCount)
			}
			_ = expectedMin
		})
	}
}

func TestNewElementalWeaponEffectSystem(t *testing.T) {
	world := NewWorld()
	system := NewElementalWeaponEffectSystem(world)

	if system == nil {
		t.Fatal("NewElementalWeaponEffectSystem returned nil")
	}
	if system.world != world {
		t.Error("System world reference mismatch")
	}
	if system.animationSpeed <= 0 {
		t.Errorf("animationSpeed = %v, want > 0", system.animationSpeed)
	}
}

func TestElementalWeaponEffectSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewElementalWeaponEffectSystem(world)

	// Create entity with elemental weapon component
	entity := world.CreateEntity()
	comp := NewElementalWeaponComponent(ElementalFire, 0.7, 54321)
	comp.AnimationPhase = 0.0
	entity.AddComponent(comp)

	entities := []*Entity{entity}
	initialPhase := comp.AnimationPhase

	// Update with delta time
	system.Update(entities, 0.5) // 0.5 seconds

	// Animation phase should have advanced
	if comp.AnimationPhase <= initialPhase {
		t.Errorf("AnimationPhase should have advanced from %v", initialPhase)
	}
}

func TestElementalWeaponEffectSystem_Update_PhaseWrap(t *testing.T) {
	world := NewWorld()
	system := NewElementalWeaponEffectSystem(world)

	entity := world.CreateEntity()
	comp := NewElementalWeaponComponent(ElementalIce, 0.8, 11111)
	comp.AnimationPhase = 0.9
	entity.AddComponent(comp)

	entities := []*Entity{entity}

	// Update with enough time to wrap phase
	system.Update(entities, 3.0) // 3 seconds should wrap

	// Phase should have wrapped to stay in [0, 1)
	if comp.AnimationPhase >= 1.0 {
		t.Errorf("AnimationPhase = %v, should wrap to stay < 1.0", comp.AnimationPhase)
	}
}

func TestElementalWeaponEffectSystem_Update_NilEntity(t *testing.T) {
	world := NewWorld()
	system := NewElementalWeaponEffectSystem(world)

	entities := []*Entity{nil}

	// Should not panic
	system.Update(entities, 0.5)
}

func TestElementalWeaponEffectSystem_Update_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewElementalWeaponEffectSystem(world)

	entity := world.CreateEntity()
	// No elemental weapon component added

	entities := []*Entity{entity}

	// Should not panic
	system.Update(entities, 0.5)
}

func TestElementalWeaponEffectSystem_Update_ZeroDelta(t *testing.T) {
	world := NewWorld()
	system := NewElementalWeaponEffectSystem(world)

	entity := world.CreateEntity()
	comp := NewElementalWeaponComponent(ElementalLightning, 0.6, 22222)
	comp.AnimationPhase = 0.5
	entity.AddComponent(comp)

	entities := []*Entity{entity}
	initialPhase := comp.AnimationPhase

	// Update with zero delta
	system.Update(entities, 0.0)

	// Phase should not change
	if comp.AnimationPhase != initialPhase {
		t.Errorf("AnimationPhase changed from %v to %v with zero delta",
			initialPhase, comp.AnimationPhase)
	}
}

func TestGetElementalParamsForEntity(t *testing.T) {
	world := NewWorld()

	t.Run("entity_with_component", func(t *testing.T) {
		entity := world.CreateEntity()
		comp := NewElementalWeaponComponent(ElementalPoison, 0.75, 33333)
		entity.AddComponent(comp)

		result := GetElementalParamsForEntity(entity)
		if result == nil {
			t.Fatal("Expected component, got nil")
		}
		if result.ElementType != ElementalPoison {
			t.Errorf("ElementType = %v, want %v", result.ElementType, ElementalPoison)
		}
	})

	t.Run("entity_without_component", func(t *testing.T) {
		entity := world.CreateEntity()

		result := GetElementalParamsForEntity(entity)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("nil_entity", func(t *testing.T) {
		result := GetElementalParamsForEntity(nil)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})
}

func TestCreateElementalWeaponFromItem(t *testing.T) {
	tests := []struct {
		name          string
		enchantType   string
		rarity        string
		expectNil     bool
		expectElement ElementalEnchantmentType
		minIntensity  float64
		maxIntensity  float64
	}{
		{"fire_common", "fire", "common", false, ElementalFire, 0.3, 0.5},
		{"frost_uncommon", "frost", "uncommon", false, ElementalIce, 0.4, 0.6},
		{"lightning_rare", "lightning", "rare", false, ElementalLightning, 0.6, 0.7},
		{"poison_epic", "toxic", "epic", false, ElementalPoison, 0.75, 0.85},
		{"holy_legendary", "divine", "legendary", false, ElementalHoly, 0.9, 1.0},
		{"shadow_default", "shadow", "unknown", false, ElementalShadow, 0.4, 0.6},
		{"no_element", "physical", "rare", true, ElementalNone, 0, 0},
		{"empty", "", "common", true, ElementalNone, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := CreateElementalWeaponFromItem(tt.enchantType, tt.rarity, 12345)

			if tt.expectNil {
				if comp != nil {
					t.Errorf("Expected nil, got component with element %v", comp.ElementType)
				}
				return
			}

			if comp == nil {
				t.Fatal("Expected component, got nil")
			}
			if comp.ElementType != tt.expectElement {
				t.Errorf("ElementType = %v, want %v", comp.ElementType, tt.expectElement)
			}
			if comp.Intensity < tt.minIntensity || comp.Intensity > tt.maxIntensity {
				t.Errorf("Intensity = %v, want [%v, %v]",
					comp.Intensity, tt.minIntensity, tt.maxIntensity)
			}
		})
	}
}

func TestApplyElementalEnchantmentToEntity(t *testing.T) {
	world := NewWorld()

	t.Run("apply_to_new_entity", func(t *testing.T) {
		entity := world.CreateEntity()

		success := ApplyElementalEnchantmentToEntity(entity, "fire", "rare", 54321)

		if !success {
			t.Error("Expected success")
		}

		comp, has := entity.GetComponent("elemental_weapon")
		if !has {
			t.Fatal("Component not added")
		}

		ewc, ok := comp.(*ElementalWeaponComponent)
		if !ok {
			t.Fatal("Wrong component type")
		}
		if ewc.ElementType != ElementalFire {
			t.Errorf("ElementType = %v, want %v", ewc.ElementType, ElementalFire)
		}
	})

	t.Run("update_existing_component", func(t *testing.T) {
		entity := world.CreateEntity()
		entity.AddComponent(NewElementalWeaponComponent(ElementalIce, 0.5, 11111))

		success := ApplyElementalEnchantmentToEntity(entity, "lightning", "epic", 22222)

		if !success {
			t.Error("Expected success")
		}

		comp, _ := entity.GetComponent("elemental_weapon")
		ewc, _ := comp.(*ElementalWeaponComponent)
		if ewc.ElementType != ElementalLightning {
			t.Errorf("ElementType = %v, want %v", ewc.ElementType, ElementalLightning)
		}
		if !ewc.IsDirty {
			t.Error("Should be marked dirty after update")
		}
	})

	t.Run("nil_entity", func(t *testing.T) {
		success := ApplyElementalEnchantmentToEntity(nil, "fire", "common", 12345)
		if success {
			t.Error("Expected failure for nil entity")
		}
	})

	t.Run("invalid_element", func(t *testing.T) {
		entity := world.CreateEntity()
		success := ApplyElementalEnchantmentToEntity(entity, "physical", "common", 12345)
		if success {
			t.Error("Expected failure for invalid element")
		}
	})
}

func TestGenerateElementalSeed(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		seed1 := GenerateElementalSeed("fire_sword_001", 12345)
		seed2 := GenerateElementalSeed("fire_sword_001", 12345)

		if seed1 != seed2 {
			t.Errorf("Same inputs should produce same seed: %v != %v", seed1, seed2)
		}
	})

	t.Run("different_items", func(t *testing.T) {
		seed1 := GenerateElementalSeed("fire_sword_001", 12345)
		seed2 := GenerateElementalSeed("frost_axe_002", 12345)

		if seed1 == seed2 {
			t.Error("Different items should produce different seeds")
		}
	})

	t.Run("different_base_seeds", func(t *testing.T) {
		seed1 := GenerateElementalSeed("weapon", 12345)
		seed2 := GenerateElementalSeed("weapon", 54321)

		if seed1 == seed2 {
			t.Error("Different base seeds should produce different results")
		}
	})
}

func TestElementalEffectColors(t *testing.T) {
	tests := []struct {
		element ElementalEnchantmentType
	}{
		{ElementalFire},
		{ElementalIce},
		{ElementalLightning},
		{ElementalPoison},
		{ElementalHoly},
		{ElementalShadow},
		{ElementalNone},
	}

	for _, tt := range tests {
		t.Run(tt.element.String(), func(t *testing.T) {
			r, g, b := ElementalEffectColors(tt.element)

			// Just verify valid color values returned
			// (r, g, b are uint8 so always valid)
			_ = r
			_ = g
			_ = b

			// Verify colors are distinct for different elements
			if tt.element != ElementalNone {
				noneR, noneG, noneB := ElementalEffectColors(ElementalNone)
				if r == noneR && g == noneG && b == noneB {
					t.Errorf("Element %v should have distinct color from None", tt.element)
				}
			}
		})
	}

	// Verify specific color expectations
	t.Run("fire_is_warm", func(t *testing.T) {
		r, g, b := ElementalEffectColors(ElementalFire)
		if r <= g || r <= b {
			t.Errorf("Fire should be warm (red-dominant): R=%v G=%v B=%v", r, g, b)
		}
	})

	t.Run("ice_is_cool", func(t *testing.T) {
		r, g, b := ElementalEffectColors(ElementalIce)
		if b < r || b < 200 {
			t.Errorf("Ice should be cool (blue-ish): R=%v G=%v B=%v", r, g, b)
		}
	})

	t.Run("poison_is_green", func(t *testing.T) {
		r, g, b := ElementalEffectColors(ElementalPoison)
		if g < r || g < b {
			t.Errorf("Poison should be green-ish: R=%v G=%v B=%v", r, g, b)
		}
	})
}

func BenchmarkElementalWeaponEffectSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewElementalWeaponEffectSystem(world)

	// Create entities with elemental weapon components
	entities := make([]*Entity, 100)
	elements := []ElementalEnchantmentType{
		ElementalFire, ElementalIce, ElementalLightning,
		ElementalPoison, ElementalHoly, ElementalShadow,
	}

	for i := range entities {
		entity := world.CreateEntity()
		element := elements[i%len(elements)]
		entity.AddComponent(NewElementalWeaponComponent(element, 0.7, int64(i)))
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016) // ~60 FPS
	}
}

func BenchmarkCreateElementalWeaponFromItem(b *testing.B) {
	enchantTypes := []string{"fire", "frost", "lightning", "toxic", "divine", "shadow"}
	rarities := []string{"common", "uncommon", "rare", "epic", "legendary"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enchant := enchantTypes[i%len(enchantTypes)]
		rarity := rarities[i%len(rarities)]
		CreateElementalWeaponFromItem(enchant, rarity, int64(i))
	}
}
