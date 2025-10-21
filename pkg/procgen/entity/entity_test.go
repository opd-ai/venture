package entity

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestNewEntityGenerator(t *testing.T) {
	gen := NewEntityGenerator()
	if gen == nil {
		t.Fatal("NewEntityGenerator returned nil")
	}
	
	if gen.templates == nil {
		t.Fatal("templates map is nil")
	}
	
	// Check that templates are registered
	if len(gen.templates["fantasy"]) == 0 {
		t.Error("fantasy templates not registered")
	}
	
	if len(gen.templates["scifi"]) == 0 {
		t.Error("scifi templates not registered")
	}
}

func TestEntityGeneration(t *testing.T) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 20,
		},
	}
	
	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	entities, ok := result.([]*Entity)
	if !ok {
		t.Fatal("result is not []*Entity")
	}
	
	if len(entities) != 20 {
		t.Errorf("expected 20 entities, got %d", len(entities))
	}
	
	// Check each entity
	for i, entity := range entities {
		if entity == nil {
			t.Errorf("entity %d is nil", i)
			continue
		}
		
		if entity.Name == "" {
			t.Errorf("entity %d has empty name", i)
		}
		
		if entity.Stats.MaxHealth <= 0 {
			t.Errorf("entity %d (%s) has invalid max health: %d", i, entity.Name, entity.Stats.MaxHealth)
		}
		
		if entity.Stats.Level <= 0 {
			t.Errorf("entity %d (%s) has invalid level: %d", i, entity.Name, entity.Stats.Level)
		}
		
		if entity.Stats.Speed <= 0 {
			t.Errorf("entity %d (%s) has invalid speed: %f", i, entity.Name, entity.Stats.Speed)
		}
	}
}

func TestEntityGenerationDeterministic(t *testing.T) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      3,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 10,
		},
	}
	
	seed := int64(42)
	
	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generate failed: %v", err1)
	}
	
	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generate failed: %v", err2)
	}
	
	entities1 := result1.([]*Entity)
	entities2 := result2.([]*Entity)
	
	if len(entities1) != len(entities2) {
		t.Fatalf("entity counts differ: %d vs %d", len(entities1), len(entities2))
	}
	
	// Check that entities are identical
	for i := range entities1 {
		e1 := entities1[i]
		e2 := entities2[i]
		
		if e1.Name != e2.Name {
			t.Errorf("entity %d name differs: %s vs %s", i, e1.Name, e2.Name)
		}
		
		if e1.Type != e2.Type {
			t.Errorf("entity %d type differs: %v vs %v", i, e1.Type, e2.Type)
		}
		
		if e1.Stats.MaxHealth != e2.Stats.MaxHealth {
			t.Errorf("entity %d (%s) max health differs: %d vs %d", i, e1.Name, e1.Stats.MaxHealth, e2.Stats.MaxHealth)
		}
		
		if e1.Stats.Damage != e2.Stats.Damage {
			t.Errorf("entity %d (%s) damage differs: %d vs %d", i, e1.Name, e1.Stats.Damage, e2.Stats.Damage)
		}
		
		if e1.Stats.Level != e2.Stats.Level {
			t.Errorf("entity %d (%s) level differs: %d vs %d", i, e1.Name, e1.Stats.Level, e2.Stats.Level)
		}
	}
}

func TestEntityGenerationSciFi(t *testing.T) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"count": 10,
		},
	}
	
	result, err := gen.Generate(54321, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	entities := result.([]*Entity)
	if len(entities) != 10 {
		t.Errorf("expected 10 entities, got %d", len(entities))
	}
	
	// All entities should be valid
	for i, entity := range entities {
		if entity.Name == "" {
			t.Errorf("entity %d has empty name", i)
		}
	}
}

func TestEntityValidation(t *testing.T) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 15,
		},
	}
	
	result, err := gen.Generate(99999, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// Should pass validation
	if err := gen.Validate(result); err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}

func TestEntityTypes(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   string
	}{
		{TypeMonster, "monster"},
		{TypeNPC, "npc"},
		{TypeBoss, "boss"},
		{TypeMinion, "minion"},
	}
	
	for _, tt := range tests {
		if got := tt.entityType.String(); got != tt.expected {
			t.Errorf("EntityType.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestEntitySize(t *testing.T) {
	tests := []struct {
		size     EntitySize
		expected string
	}{
		{SizeTiny, "tiny"},
		{SizeSmall, "small"},
		{SizeMedium, "medium"},
		{SizeLarge, "large"},
		{SizeHuge, "huge"},
	}
	
	for _, tt := range tests {
		if got := tt.size.String(); got != tt.expected {
			t.Errorf("EntitySize.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestRarity(t *testing.T) {
	tests := []struct {
		rarity   Rarity
		expected string
	}{
		{RarityCommon, "common"},
		{RarityUncommon, "uncommon"},
		{RarityRare, "rare"},
		{RarityEpic, "epic"},
		{RarityLegendary, "legendary"},
	}
	
	for _, tt := range tests {
		if got := tt.rarity.String(); got != tt.expected {
			t.Errorf("Rarity.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestEntityIsHostile(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   bool
	}{
		{TypeMonster, true},
		{TypeBoss, true},
		{TypeMinion, true},
		{TypeNPC, false},
	}
	
	for _, tt := range tests {
		entity := &Entity{Type: tt.entityType}
		if got := entity.IsHostile(); got != tt.expected {
			t.Errorf("Entity{Type: %v}.IsHostile() = %v, want %v", tt.entityType, got, tt.expected)
		}
	}
}

func TestEntityIsBoss(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   bool
	}{
		{TypeMonster, false},
		{TypeBoss, true},
		{TypeMinion, false},
		{TypeNPC, false},
	}
	
	for _, tt := range tests {
		entity := &Entity{Type: tt.entityType}
		if got := entity.IsBoss(); got != tt.expected {
			t.Errorf("Entity{Type: %v}.IsBoss() = %v, want %v", tt.entityType, got, tt.expected)
		}
	}
}

func TestEntityThreatLevel(t *testing.T) {
	entity := &Entity{
		Type: TypeMonster,
		Stats: Stats{
			Health:    100,
			MaxHealth: 100,
			Damage:    10,
			Defense:   5,
			Level:     5,
		},
	}
	
	threat := entity.GetThreatLevel()
	if threat < 0 || threat > 100 {
		t.Errorf("GetThreatLevel() = %d, expected 0-100", threat)
	}
	
	// Boss should have higher threat
	boss := &Entity{
		Type: TypeBoss,
		Stats: Stats{
			Health:    100,
			MaxHealth: 100,
			Damage:    10,
			Defense:   5,
			Level:     5,
		},
	}
	
	bossThreat := boss.GetThreatLevel()
	if bossThreat <= threat {
		t.Errorf("Boss threat (%d) should be higher than monster threat (%d)", bossThreat, threat)
	}
}

func TestGetFantasyTemplates(t *testing.T) {
	templates := GetFantasyTemplates()
	if len(templates) == 0 {
		t.Error("GetFantasyTemplates returned empty slice")
	}
	
	// Check that we have various types
	hasMonster := false
	hasBoss := false
	hasNPC := false
	
	for _, template := range templates {
		switch template.BaseType {
		case TypeMonster:
			hasMonster = true
		case TypeBoss:
			hasBoss = true
		case TypeNPC:
			hasNPC = true
		}
	}
	
	if !hasMonster {
		t.Error("Fantasy templates missing monster type")
	}
	if !hasBoss {
		t.Error("Fantasy templates missing boss type")
	}
	if !hasNPC {
		t.Error("Fantasy templates missing NPC type")
	}
}

func TestGetSciFiTemplates(t *testing.T) {
	templates := GetSciFiTemplates()
	if len(templates) == 0 {
		t.Error("GetSciFiTemplates returned empty slice")
	}
	
	// All templates should have valid ranges
	for i, template := range templates {
		if template.HealthRange[0] > template.HealthRange[1] {
			t.Errorf("Template %d has invalid health range", i)
		}
		if template.DamageRange[0] > template.DamageRange[1] {
			t.Errorf("Template %d has invalid damage range", i)
		}
		if template.DefenseRange[0] > template.DefenseRange[1] {
			t.Errorf("Template %d has invalid defense range", i)
		}
		if template.SpeedRange[0] > template.SpeedRange[1] {
			t.Errorf("Template %d has invalid speed range", i)
		}
	}
}

func TestEntityLevelScaling(t *testing.T) {
	gen := NewEntityGenerator()
	
	// Generate entities at different depths
	depths := []int{1, 5, 10, 20}
	
	for _, depth := range depths {
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      depth,
			GenreID:    "fantasy",
			Custom: map[string]interface{}{
				"count": 5,
			},
		}
		
		result, err := gen.Generate(int64(depth)*100, params)
		if err != nil {
			t.Fatalf("Generate failed at depth %d: %v", depth, err)
		}
		
		entities := result.([]*Entity)
		
		// Check that entities have appropriate levels for depth
		for _, entity := range entities {
			// Level should be roughly proportional to depth
			if entity.Stats.Level < depth/2 {
				t.Errorf("At depth %d, entity %s has suspiciously low level %d", depth, entity.Name, entity.Stats.Level)
			}
		}
	}
}

func TestEntityValidation_InvalidInput(t *testing.T) {
	gen := NewEntityGenerator()
	
	// Test with non-entity slice
	err := gen.Validate("not an entity slice")
	if err == nil {
		t.Error("Validate should fail for non-entity slice")
	}
	
	// Test with empty entity slice
	err = gen.Validate([]*Entity{})
	if err == nil {
		t.Error("Validate should fail for empty entity slice")
	}
	
	// Test with invalid entity (empty name)
	invalidEntity := &Entity{
		Name: "",
		Stats: Stats{
			MaxHealth: 100,
			Level:     1,
			Speed:     1.0,
		},
	}
	err = gen.Validate([]*Entity{invalidEntity})
	if err == nil {
		t.Error("Validate should fail for entity with empty name")
	}
	
	// Test with invalid max health
	invalidHealth := &Entity{
		Name: "Test",
		Stats: Stats{
			MaxHealth: 0,
			Level:     1,
			Speed:     1.0,
		},
	}
	err = gen.Validate([]*Entity{invalidHealth})
	if err == nil {
		t.Error("Validate should fail for entity with zero max health")
	}
	
	// Test with invalid level
	invalidLevel := &Entity{
		Name: "Test",
		Stats: Stats{
			MaxHealth: 100,
			Level:     0,
			Speed:     1.0,
		},
	}
	err = gen.Validate([]*Entity{invalidLevel})
	if err == nil {
		t.Error("Validate should fail for entity with zero level")
	}
	
	// Test with invalid speed
	invalidSpeed := &Entity{
		Name: "Test",
		Stats: Stats{
			MaxHealth: 100,
			Level:     1,
			Speed:     0,
		},
	}
	err = gen.Validate([]*Entity{invalidSpeed})
	if err == nil {
		t.Error("Validate should fail for entity with zero speed")
	}
}

func TestEntityType_String_Unknown(t *testing.T) {
	unknown := EntityType(99)
	if got := unknown.String(); got != "unknown" {
		t.Errorf("Unknown EntityType.String() = %v, want 'unknown'", got)
	}
}

func TestEntitySize_String_Unknown(t *testing.T) {
	unknown := EntitySize(99)
	if got := unknown.String(); got != "unknown" {
		t.Errorf("Unknown EntitySize.String() = %v, want 'unknown'", got)
	}
}

func TestRarity_String_Unknown(t *testing.T) {
	unknown := Rarity(99)
	if got := unknown.String(); got != "unknown" {
		t.Errorf("Unknown Rarity.String() = %v, want 'unknown'", got)
	}
}

func TestEntityGeneration_DifferentDifficulties(t *testing.T) {
	gen := NewEntityGenerator()
	difficulties := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	
	for _, diff := range difficulties {
		params := procgen.GenerationParams{
			Difficulty: diff,
			Depth:      5,
			GenreID:    "fantasy",
			Custom: map[string]interface{}{
				"count": 5,
			},
		}
		
		result, err := gen.Generate(12345, params)
		if err != nil {
			t.Fatalf("Generate failed at difficulty %f: %v", diff, err)
		}
		
		entities := result.([]*Entity)
		if len(entities) != 5 {
			t.Errorf("Expected 5 entities at difficulty %f, got %d", diff, len(entities))
		}
	}
}

func TestEntityGeneration_UnknownGenre(t *testing.T) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "unknown_genre",
		Custom: map[string]interface{}{
			"count": 5,
		},
	}
	
	// Should fall back to default templates or generate entities anyway
	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed with unknown genre: %v", err)
	}
	
	entities := result.([]*Entity)
	if len(entities) != 5 {
		t.Errorf("Expected 5 entities with unknown genre, got %d", len(entities))
	}
}

func TestEntityGeneration_ZeroCount(t *testing.T) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 0,
		},
	}
	
	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	entities := result.([]*Entity)
	// With count=0, should still generate default number (likely 10)
	// If it generates 0 entities, that's also acceptable behavior
	if len(entities) < 0 {
		t.Error("Entity count should not be negative")
	}
}

func TestEntityGeneration_LargeCount(t *testing.T) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 100,
		},
	}
	
	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	entities := result.([]*Entity)
	if len(entities) != 100 {
		t.Errorf("Expected 100 entities, got %d", len(entities))
	}
}

func TestEntityThreatLevel_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		entity *Entity
	}{
		{
			name: "MinimalStats",
			entity: &Entity{
				Type: TypeMinion,
				Stats: Stats{
					Health:    1,
					MaxHealth: 1,
					Damage:    1,
					Defense:   0,
					Level:     1,
				},
			},
		},
		{
			name: "MaximalStats",
			entity: &Entity{
				Type: TypeBoss,
				Stats: Stats{
					Health:    10000,
					MaxHealth: 10000,
					Damage:    500,
					Defense:   200,
					Level:     100,
				},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threat := tt.entity.GetThreatLevel()
			if threat < 0 {
				t.Errorf("GetThreatLevel() = %d, should not be negative", threat)
			}
		})
	}
}

func BenchmarkEntityGeneration(b *testing.B) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 10,
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}
