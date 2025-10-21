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
