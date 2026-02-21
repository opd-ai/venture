package engine

import (
	"testing"
)

func TestBuildArchetypeString(t *testing.T) {
	tests := []struct {
		name      string
		archetype BuildArchetype
		want      string
	}{
		{"custom", BuildArchetypeCustom, "Custom"},
		{"tank", BuildArchetypeTank, "Tank"},
		{"dps", BuildArchetypeDPS, "DPS"},
		{"support", BuildArchetypeSupport, "Support"},
		{"hybrid", BuildArchetypeHybrid, "Hybrid"},
		{"battlemage", BuildArchetypeBattlemage, "Battlemage"},
		{"assassin", BuildArchetypeAssassin, "Assassin"},
		{"paladin", BuildArchetypePaladin, "Paladin"},
		{"unknown", BuildArchetype(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.archetype.String(); got != tt.want {
				t.Errorf("BuildArchetype.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildArchetypeDescription(t *testing.T) {
	tests := []struct {
		name      string
		archetype BuildArchetype
		wantLen   int // Minimum description length
	}{
		{"custom", BuildArchetypeCustom, 10},
		{"tank", BuildArchetypeTank, 10},
		{"dps", BuildArchetypeDPS, 10},
		{"support", BuildArchetypeSupport, 10},
		{"hybrid", BuildArchetypeHybrid, 10},
		{"battlemage", BuildArchetypeBattlemage, 10},
		{"assassin", BuildArchetypeAssassin, 10},
		{"paladin", BuildArchetypePaladin, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.archetype.Description()
			if len(got) < tt.wantLen {
				t.Errorf("BuildArchetype.Description() length = %v, want >= %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestBuildTemplateTotalPoints(t *testing.T) {
	template := &BuildTemplate{
		ID:   "test",
		Name: "Test Build",
		Attributes: map[int]int{
			int(AttrStrength):     10,
			int(AttrAgility):      5,
			int(AttrIntelligence): 3,
		},
		Talents: map[string]int{
			"talent_a": 3,
			"talent_b": 2,
		},
		Skills: map[string]int{
			"skill_a": 5,
			"skill_b": 3,
			"skill_c": 2,
		},
		RequiredLevel: 10,
	}

	tests := []struct {
		name     string
		method   func() int
		expected int
	}{
		{"attribute points", template.TotalAttributePoints, 18},
		{"talent points", template.TotalTalentPoints, 5},
		{"skill points", template.TotalSkillPoints, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method(); got != tt.expected {
				t.Errorf("Total %s = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestBuildTemplateValidate(t *testing.T) {
	tests := []struct {
		name      string
		template  *BuildTemplate
		wantError bool
	}{
		{
			name: "valid template",
			template: &BuildTemplate{
				ID:            "test_1",
				Name:          "Test Build",
				RequiredLevel: 5,
			},
			wantError: false,
		},
		{
			name: "missing ID",
			template: &BuildTemplate{
				Name:          "Test Build",
				RequiredLevel: 5,
			},
			wantError: true,
		},
		{
			name: "missing name",
			template: &BuildTemplate{
				ID:            "test_1",
				RequiredLevel: 5,
			},
			wantError: true,
		},
		{
			name: "invalid level",
			template: &BuildTemplate{
				ID:            "test_1",
				Name:          "Test Build",
				RequiredLevel: 0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestBuildTemplateClone(t *testing.T) {
	original := &BuildTemplate{
		ID:          "original",
		Name:        "Original Build",
		Description: "Test description",
		Archetype:   BuildArchetypeDPS,
		Class:       ClassWarrior,
		Attributes: map[int]int{
			int(AttrStrength): 10,
			int(AttrAgility):  5,
		},
		Talents: map[string]int{
			"talent_a": 3,
		},
		Skills: map[string]int{
			"skill_a": 2,
		},
		RequiredLevel: 10,
		IsPreset:      true,
	}

	clone := original.Clone()

	// Verify clone is equal
	if clone.ID != original.ID {
		t.Errorf("Clone ID = %v, want %v", clone.ID, original.ID)
	}
	if clone.Name != original.Name {
		t.Errorf("Clone Name = %v, want %v", clone.Name, original.Name)
	}

	// Modify clone and verify original unchanged
	clone.Attributes[int(AttrStrength)] = 20
	if original.Attributes[int(AttrStrength)] == 20 {
		t.Error("Modifying clone affected original")
	}

	clone.Talents["talent_a"] = 5
	if original.Talents["talent_a"] == 5 {
		t.Error("Modifying clone talents affected original")
	}
}

func TestBuildTemplateComponentType(t *testing.T) {
	comp := NewBuildTemplateComponent()
	if got := comp.Type(); got != "build_template" {
		t.Errorf("Type() = %v, want %v", got, "build_template")
	}
}

func TestNewBuildTemplateComponent(t *testing.T) {
	comp := NewBuildTemplateComponent()

	if comp.MaxTemplates != 10 {
		t.Errorf("MaxTemplates = %v, want %v", comp.MaxTemplates, 10)
	}
	if comp.PendingApply != -1 {
		t.Errorf("PendingApply = %v, want %v", comp.PendingApply, -1)
	}
	if comp.ApplyCooldown != 30.0 {
		t.Errorf("ApplyCooldown = %v, want %v", comp.ApplyCooldown, 30.0)
	}
	if len(comp.Templates) != 0 {
		t.Errorf("Templates length = %v, want %v", len(comp.Templates), 0)
	}
}

func TestBuildTemplateComponentAddRemove(t *testing.T) {
	comp := NewBuildTemplateComponent()

	template := &BuildTemplate{
		ID:            "test_1",
		Name:          "Test Build",
		RequiredLevel: 1,
	}

	// Add template
	index := comp.AddTemplate(template)
	if index != 0 {
		t.Errorf("AddTemplate() = %v, want %v", index, 0)
	}
	if comp.GetTemplateCount() != 1 {
		t.Errorf("GetTemplateCount() = %v, want %v", comp.GetTemplateCount(), 1)
	}

	// Get template
	got := comp.GetTemplate(0)
	if got == nil || got.ID != "test_1" {
		t.Error("GetTemplate(0) returned wrong template")
	}

	// Get by ID
	byID := comp.GetTemplateByID("test_1")
	if byID == nil || byID.Name != "Test Build" {
		t.Error("GetTemplateByID returned wrong template")
	}

	// Invalid index
	if comp.GetTemplate(-1) != nil {
		t.Error("GetTemplate(-1) should return nil")
	}
	if comp.GetTemplate(100) != nil {
		t.Error("GetTemplate(100) should return nil")
	}

	// Remove template
	if !comp.RemoveTemplate(0) {
		t.Error("RemoveTemplate(0) should succeed")
	}
	if comp.GetTemplateCount() != 0 {
		t.Errorf("GetTemplateCount() after remove = %v, want %v", comp.GetTemplateCount(), 0)
	}
}

func TestBuildTemplateComponentMaxCapacity(t *testing.T) {
	comp := NewBuildTemplateComponent()
	comp.MaxTemplates = 3

	// Add templates up to max
	for i := 0; i < 3; i++ {
		template := &BuildTemplate{
			ID:            "test_" + string(rune('a'+i)),
			Name:          "Test Build",
			RequiredLevel: 1,
		}
		index := comp.AddTemplate(template)
		if index != i {
			t.Errorf("AddTemplate() = %v, want %v", index, i)
		}
	}

	// Try to add beyond max
	overflowTemplate := &BuildTemplate{
		ID:            "overflow",
		Name:          "Overflow",
		RequiredLevel: 1,
	}
	index := comp.AddTemplate(overflowTemplate)
	if index != -1 {
		t.Errorf("AddTemplate beyond max = %v, want -1", index)
	}

	if comp.GetAvailableSlots() != 0 {
		t.Errorf("GetAvailableSlots() = %v, want 0", comp.GetAvailableSlots())
	}
}

func TestBuildTemplateComponentPresetProtection(t *testing.T) {
	comp := NewBuildTemplateComponent()

	preset := &BuildTemplate{
		ID:            "preset_tank",
		Name:          "Tank Preset",
		RequiredLevel: 1,
		IsPreset:      true,
	}

	comp.AddTemplate(preset)

	// Cannot remove preset
	if comp.RemoveTemplate(0) {
		t.Error("RemoveTemplate should fail for presets")
	}

	// Cannot update preset
	newTemplate := &BuildTemplate{
		ID:            "modified",
		Name:          "Modified",
		RequiredLevel: 1,
	}
	if comp.UpdateTemplate(0, newTemplate) {
		t.Error("UpdateTemplate should fail for presets")
	}
}

func TestBuildTemplateComponentCooldown(t *testing.T) {
	comp := NewBuildTemplateComponent()
	comp.ApplyCooldown = 10.0
	comp.LastApplyTime = 0.0

	// Can apply at time 10+
	if !comp.CanApplyTemplate(10.0) {
		t.Error("CanApplyTemplate(10.0) should be true")
	}

	// Cannot apply at time 5
	if comp.CanApplyTemplate(5.0) {
		t.Error("CanApplyTemplate(5.0) should be false")
	}

	// Check remaining cooldown
	remaining := comp.GetApplyCooldownRemaining(5.0)
	if remaining != 5.0 {
		t.Errorf("GetApplyCooldownRemaining(5.0) = %v, want 5.0", remaining)
	}

	// Cooldown at 15 should be 0
	remaining = comp.GetApplyCooldownRemaining(15.0)
	if remaining != 0.0 {
		t.Errorf("GetApplyCooldownRemaining(15.0) = %v, want 0.0", remaining)
	}
}

func TestBuildTemplateComponentRequestApply(t *testing.T) {
	comp := NewBuildTemplateComponent()
	comp.ApplyCooldown = 10.0
	comp.LastApplyTime = 0.0

	template := &BuildTemplate{
		ID:            "test_1",
		Name:          "Test Build",
		RequiredLevel: 1,
	}
	comp.AddTemplate(template)

	// Request apply at valid time
	if !comp.RequestApply(0, 15.0) {
		t.Error("RequestApply should succeed when cooldown is ready")
	}
	if comp.PendingApply != 0 {
		t.Errorf("PendingApply = %v, want 0", comp.PendingApply)
	}

	// Reset and try during cooldown
	comp.PendingApply = -1
	comp.LastApplyTime = 10.0
	if comp.RequestApply(0, 15.0) {
		t.Error("RequestApply should fail during cooldown")
	}

	// Invalid index
	if comp.RequestApply(-1, 100.0) {
		t.Error("RequestApply(-1) should fail")
	}
	if comp.RequestApply(100, 100.0) {
		t.Error("RequestApply(100) should fail")
	}
}

func TestBuildTemplateComponentMarkComplete(t *testing.T) {
	comp := NewBuildTemplateComponent()
	comp.PendingApply = 0

	comp.MarkApplyComplete("test_1", 50.0)

	if comp.PendingApply != -1 {
		t.Errorf("PendingApply after complete = %v, want -1", comp.PendingApply)
	}
	if comp.ActiveTemplateID != "test_1" {
		t.Errorf("ActiveTemplateID = %v, want test_1", comp.ActiveTemplateID)
	}
	if comp.LastApplyTime != 50.0 {
		t.Errorf("LastApplyTime = %v, want 50.0", comp.LastApplyTime)
	}
}

func TestBuildTemplateComponentSerialize(t *testing.T) {
	comp := NewBuildTemplateComponent()
	comp.AddTemplate(&BuildTemplate{
		ID:            "test_1",
		Name:          "Test Build",
		RequiredLevel: 5,
		Attributes:    map[int]int{int(AttrStrength): 10},
		Talents:       map[string]int{"talent_a": 2},
		Skills:        map[string]int{"skill_a": 3},
	})
	comp.ActiveTemplateID = "test_1"

	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	newComp := &BuildTemplateComponent{}
	if err := newComp.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if newComp.ActiveTemplateID != "test_1" {
		t.Errorf("Deserialized ActiveTemplateID = %v, want test_1", newComp.ActiveTemplateID)
	}
	if newComp.GetTemplateCount() != 1 {
		t.Errorf("Deserialized template count = %v, want 1", newComp.GetTemplateCount())
	}
}

func TestBuildTemplateComponentGetNames(t *testing.T) {
	comp := NewBuildTemplateComponent()
	comp.AddTemplate(&BuildTemplate{ID: "1", Name: "First", RequiredLevel: 1})
	comp.AddTemplate(&BuildTemplate{ID: "2", Name: "Second", RequiredLevel: 1})

	names := comp.GetTemplateNames()
	if len(names) != 2 {
		t.Fatalf("GetTemplateNames() length = %v, want 2", len(names))
	}
	if names[0] != "First" || names[1] != "Second" {
		t.Errorf("GetTemplateNames() = %v, want [First Second]", names)
	}
}
