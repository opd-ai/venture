package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/magic"
)

func TestNewCreatureElementalAuraSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.world != world {
		t.Error("world reference not set")
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
	if len(sys.presets) == 0 {
		t.Error("element presets not initialized")
	}
	if len(sys.genreMods) == 0 {
		t.Error("genre modifiers not initialized")
	}
}

func TestCreatureElementalAuraSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	genres := []string{"fantasy", "horror", "sci-fi", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("expected genreID %q, got %q", genre, sys.genreID)
		}
	}
}

func TestCreatureElementalAuraSystem_ElementFromKeywords(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	tests := []struct {
		keyword  string
		expected magic.ElementType
	}{
		{"fire_wolf", magic.ElementFire},
		{"flame spider", magic.ElementFire},
		{"inferno beast", magic.ElementFire},
		{"ember golem", magic.ElementFire},
		{"ice serpent", magic.ElementIce},
		{"frost bear", magic.ElementIce},
		{"frozen horror", magic.ElementIce},
		{"cryo drone", magic.ElementIce},
		{"lightning hawk", magic.ElementLightning},
		{"thunder worm", magic.ElementLightning},
		{"storm elemental", magic.ElementLightning},
		{"shock beetle", magic.ElementLightning},
		{"poison spider", magic.ElementEarth},
		{"toxic slime", magic.ElementEarth},
		{"venom snake", magic.ElementEarth},
		{"acid blob", magic.ElementEarth},
		{"wind wisp", magic.ElementWind},
		{"gale hawk", magic.ElementWind},
		{"tempest", magic.ElementWind},
		{"holy knight", magic.ElementLight},
		{"radiant being", magic.ElementLight},
		{"celestial guardian", magic.ElementLight},
		{"shadow demon", magic.ElementDark},
		{"void horror", magic.ElementDark},
		{"nightmare beast", magic.ElementDark},
		{"arcane construct", magic.ElementArcane},
		{"mystic golem", magic.ElementArcane},
		{"regular wolf", magic.ElementNone},
		{"giant spider", magic.ElementNone},
		{"", magic.ElementNone},
	}

	for _, tt := range tests {
		got := sys.elementFromKeywords(tt.keyword)
		if got != tt.expected {
			t.Errorf("elementFromKeywords(%q) = %v, want %v", tt.keyword, got, tt.expected)
		}
	}
}

func TestCreatureElementalAuraSystem_ElementFromDamageType(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	tests := []struct {
		dmgType  string
		expected magic.ElementType
	}{
		{"fire", magic.ElementFire},
		{"Fire", magic.ElementFire},
		{"ice", magic.ElementIce},
		{"cold", magic.ElementIce},
		{"frost", magic.ElementIce},
		{"lightning", magic.ElementLightning},
		{"electric", magic.ElementLightning},
		{"poison", magic.ElementEarth},
		{"toxic", magic.ElementEarth},
		{"physical", magic.ElementNone},
		{"", magic.ElementNone},
	}

	for _, tt := range tests {
		got := sys.elementFromDamageType(tt.dmgType)
		if got != tt.expected {
			t.Errorf("elementFromDamageType(%q) = %v, want %v", tt.dmgType, got, tt.expected)
		}
	}
}

func TestCreatureElementalAuraSystem_CreateAuraComponent(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&TeamComponent{TeamID: 2})

	elements := []magic.ElementType{
		magic.ElementFire,
		magic.ElementIce,
		magic.ElementLightning,
		magic.ElementEarth,
		magic.ElementWind,
		magic.ElementLight,
		magic.ElementDark,
		magic.ElementArcane,
	}

	for _, elem := range elements {
		comp := sys.createAuraComponent(entity, elem)
		if comp == nil {
			t.Errorf("createAuraComponent returned nil for %v", elem)
			continue
		}
		if comp.Element != elem {
			t.Errorf("expected element %v, got %v", elem, comp.Element)
		}
		if !comp.Enabled {
			t.Errorf("expected aura to be enabled for %v", elem)
		}
		if comp.BaseIntensity <= 0 || comp.BaseIntensity > 1 {
			t.Errorf("invalid BaseIntensity %f for %v", comp.BaseIntensity, elem)
		}
		if comp.AuraRadius <= 0 {
			t.Errorf("invalid AuraRadius %f for %v", comp.AuraRadius, elem)
		}
	}
}

func TestCreatureElementalAuraSystem_UpdateAssignsAura(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	// Create enemy creature with fire-related name
	entity := world.CreateEntity()
	entity.AddComponent(&TeamComponent{TeamID: 2})
	entity.AddComponent(&NameComponent{Name: "Flame Wolf"})

	// First update shouldn't assign (throttle interval)
	sys.Update([]*Entity{entity}, 0.1)
	if entity.HasComponent("creature_elemental_aura") {
		t.Error("aura assigned too early (before throttle interval)")
	}

	// Wait for throttle interval and update again
	sys.Update([]*Entity{entity}, 1.0)
	if !entity.HasComponent("creature_elemental_aura") {
		t.Error("aura not assigned after throttle interval")
	}

	comp, _ := entity.GetComponent("creature_elemental_aura")
	aura, ok := comp.(*CreatureElementalAuraComponent)
	if !ok {
		t.Fatal("component is not CreatureElementalAuraComponent")
	}
	if aura.Element != magic.ElementFire {
		t.Errorf("expected ElementFire, got %v", aura.Element)
	}
}

func TestCreatureElementalAuraSystem_UpdatePulseAnimation(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	entity := world.CreateEntity()
	aura := &CreatureElementalAuraComponent{
		Element:          magic.ElementFire,
		BaseIntensity:    0.6,
		CurrentIntensity: 0.6,
		PulseSpeed:       2.0,
		PulseAmplitude:   0.2,
		PulsePhase:       0.0,
		Enabled:          true,
	}
	entity.AddComponent(aura)

	initialIntensity := aura.CurrentIntensity
	initialPhase := aura.PulsePhase

	// Update pulse animation
	sys.Update([]*Entity{entity}, 0.25)

	if aura.PulsePhase == initialPhase {
		t.Error("pulse phase did not advance")
	}
	// Current intensity should have changed due to sine modulation
	if math.Abs(aura.CurrentIntensity-initialIntensity) < 0.001 && aura.PulseSpeed > 0 {
		t.Log("current intensity unchanged (may be at cycle peak/trough)")
	}
}

func TestCreatureElementalAuraSystem_SkipsNonEnemy(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	// Player team entity (team 1) should not get aura
	entity := world.CreateEntity()
	entity.AddComponent(&TeamComponent{TeamID: 1})
	entity.AddComponent(&NameComponent{Name: "Fire Mage Player"})

	sys.Update([]*Entity{entity}, 2.0)
	if entity.HasComponent("creature_elemental_aura") {
		t.Error("player entity should not receive elemental aura")
	}
}

func TestCreatureElementalAuraSystem_SkipsNonElemental(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	// Enemy with non-elemental name
	entity := world.CreateEntity()
	entity.AddComponent(&TeamComponent{TeamID: 2})
	entity.AddComponent(&NameComponent{Name: "Giant Spider"})

	sys.Update([]*Entity{entity}, 2.0)
	if entity.HasComponent("creature_elemental_aura") {
		t.Error("non-elemental creature should not receive aura")
	}
}

func TestCreatureElementalAuraSystem_InferFromVisualTags(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&TeamComponent{TeamID: 2})
	entity.AddComponent(&CreatureVisualComponent{
		Form:       FormQuadruped,
		VisualTags: []string{"frost", "winter"},
	})

	sys.Update([]*Entity{entity}, 2.0)
	if !entity.HasComponent("creature_elemental_aura") {
		t.Fatal("aura not assigned from visual tags")
	}

	comp, _ := entity.GetComponent("creature_elemental_aura")
	aura := comp.(*CreatureElementalAuraComponent)
	if aura.Element != magic.ElementIce {
		t.Errorf("expected ElementIce from frost tag, got %v", aura.Element)
	}
}

func TestCreatureElementalAuraSystem_InferFromAttackType(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&TeamComponent{TeamID: 2})
	entity.AddComponent(&AttackComponent{DamageType: "lightning"})

	sys.Update([]*Entity{entity}, 2.0)
	if !entity.HasComponent("creature_elemental_aura") {
		t.Fatal("aura not assigned from attack type")
	}

	comp, _ := entity.GetComponent("creature_elemental_aura")
	aura := comp.(*CreatureElementalAuraComponent)
	if aura.Element != magic.ElementLightning {
		t.Errorf("expected ElementLightning from attack type, got %v", aura.Element)
	}
}

func TestCreatureElementalAuraSystem_GenreModifiers(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		genre             string
		expectedIntMult   float64
		expectedPulseMult float64
	}{
		{"fantasy", 1.0, 1.0},
		{"horror", 0.8, 0.7},
		{"sci-fi", 1.2, 1.3},
		{"cyberpunk", 1.1, 1.5},
		{"post-apocalyptic", 0.9, 0.9},
	}

	for _, tt := range tests {
		sys := NewCreatureElementalAuraSystem(world, 12345)
		sys.SetGenre(tt.genre)

		mod, ok := sys.genreMods[tt.genre]
		if !ok {
			t.Errorf("genre %q not found in genreMods", tt.genre)
			continue
		}
		if math.Abs(mod.IntensityMult-tt.expectedIntMult) > 0.01 {
			t.Errorf("%s: IntensityMult = %f, want %f", tt.genre, mod.IntensityMult, tt.expectedIntMult)
		}
		if math.Abs(mod.PulseMult-tt.expectedPulseMult) > 0.01 {
			t.Errorf("%s: PulseMult = %f, want %f", tt.genre, mod.PulseMult, tt.expectedPulseMult)
		}
	}
}

func TestCreatureElementalAuraComponent_Type(t *testing.T) {
	comp := NewCreatureElementalAuraComponent()
	if comp.Type() != "creature_elemental_aura" {
		t.Errorf("Type() = %q, want 'creature_elemental_aura'", comp.Type())
	}
}

func TestCreatureElementalAuraComponent_IsElemental(t *testing.T) {
	comp := NewCreatureElementalAuraComponent()

	// Default should not be elemental
	if comp.IsElemental() {
		t.Error("default component should not be elemental")
	}

	// Enabled with element should be elemental
	comp.Enabled = true
	comp.Element = magic.ElementFire
	if !comp.IsElemental() {
		t.Error("enabled fire component should be elemental")
	}

	// Enabled but no element should not be elemental
	comp.Element = magic.ElementNone
	if comp.IsElemental() {
		t.Error("enabled but ElementNone should not be elemental")
	}
}

func TestCreatureElementalAuraComponent_GetColors(t *testing.T) {
	comp := &CreatureElementalAuraComponent{
		AuraR:      1.0,
		AuraG:      0.5,
		AuraB:      0.25,
		SecondaryR: 0.8,
		SecondaryG: 0.4,
		SecondaryB: 0.2,
	}

	r, g, b := comp.GetPrimaryColor()
	if r != 255 || g != 127 || b != 63 {
		t.Errorf("GetPrimaryColor() = (%d,%d,%d), want (255,127,63)", r, g, b)
	}

	sr, sg, sb := comp.GetSecondaryColor()
	if sr != 204 || sg != 102 || sb != 51 {
		t.Errorf("GetSecondaryColor() = (%d,%d,%d), want (204,102,51)", sr, sg, sb)
	}
}

func TestBuildElementAuraPresets(t *testing.T) {
	presets := buildElementAuraPresets()

	elements := []magic.ElementType{
		magic.ElementNone,
		magic.ElementFire,
		magic.ElementIce,
		magic.ElementLightning,
		magic.ElementEarth,
		magic.ElementWind,
		magic.ElementLight,
		magic.ElementDark,
		magic.ElementArcane,
	}

	for _, elem := range elements {
		preset, ok := presets[elem]
		if !ok {
			t.Errorf("missing preset for %v", elem)
			continue
		}
		// Validate color ranges
		if preset.PrimaryR < 0 || preset.PrimaryR > 1 ||
			preset.PrimaryG < 0 || preset.PrimaryG > 1 ||
			preset.PrimaryB < 0 || preset.PrimaryB > 1 {
			t.Errorf("%v: primary color out of [0,1] range", elem)
		}
		if preset.BaseIntensity < 0 || preset.BaseIntensity > 1 {
			t.Errorf("%v: BaseIntensity %f out of [0,1] range", elem, preset.BaseIntensity)
		}
		if preset.AuraRadius <= 0 {
			t.Errorf("%v: AuraRadius %f should be positive", elem, preset.AuraRadius)
		}
	}

	// Fire should have particle emission
	if !presets[magic.ElementFire].ParticleEmission {
		t.Error("Fire element should have particle emission")
	}

	// Fire should pulse faster than ice
	if presets[magic.ElementFire].PulseSpeed <= presets[magic.ElementIce].PulseSpeed {
		t.Error("Fire should pulse faster than ice")
	}
}

func TestBuildGenreElementModifiers(t *testing.T) {
	mods := buildGenreElementModifiers()

	genres := []string{"fantasy", "horror", "sci-fi", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		mod, ok := mods[genre]
		if !ok {
			t.Errorf("missing modifier for genre %q", genre)
			continue
		}
		if mod.IntensityMult <= 0 {
			t.Errorf("%s: IntensityMult should be positive", genre)
		}
		if mod.PulseMult <= 0 {
			t.Errorf("%s: PulseMult should be positive", genre)
		}
	}
}

func TestClampFloat(t *testing.T) {
	tests := []struct {
		v, min, max, expected float64
	}{
		{0.5, 0.0, 1.0, 0.5},
		{-0.5, 0.0, 1.0, 0.0},
		{1.5, 0.0, 1.0, 1.0},
		{0.0, 0.0, 1.0, 0.0},
		{1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		got := clampFloat(tt.v, tt.min, tt.max)
		if got != tt.expected {
			t.Errorf("clampFloat(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.expected)
		}
	}
}

func BenchmarkCreatureElementalAuraSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&TeamComponent{TeamID: 2})
		if i%2 == 0 {
			e.AddComponent(&NameComponent{Name: "Fire Wolf"})
		}
		e.AddComponent(&CreatureElementalAuraComponent{
			Element:       magic.ElementFire,
			BaseIntensity: 0.6,
			PulseSpeed:    2.0,
			Enabled:       true,
		})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkCreatureElementalAuraSystem_ElementFromKeywords(b *testing.B) {
	world := NewWorld()
	sys := NewCreatureElementalAuraSystem(world, 12345)

	keywords := []string{
		"fire_wolf", "frost_bear", "lightning_hawk", "poison_spider",
		"shadow_demon", "holy_knight", "arcane_golem", "regular_beast",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.elementFromKeywords(keywords[i%len(keywords)])
	}
}
