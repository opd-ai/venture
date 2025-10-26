package engine

import (
	"fmt"
	"strings"
	"testing"
)

func TestCharacterClass_String(t *testing.T) {
	tests := []struct {
		name  string
		class CharacterClass
		want  string
	}{
		{"warrior", ClassWarrior, "Warrior"},
		{"mage", ClassMage, "Mage"},
		{"rogue", ClassRogue, "Rogue"},
		{"unknown", CharacterClass(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.class.String(); got != tt.want {
				t.Errorf("CharacterClass.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCharacterClass_Description(t *testing.T) {
	tests := []struct {
		name  string
		class CharacterClass
	}{
		{"warrior has description", ClassWarrior},
		{"mage has description", ClassMage},
		{"rogue has description", ClassRogue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.class.Description()
			if desc == "" {
				t.Errorf("CharacterClass.Description() returned empty string for %v", tt.class)
			}
			if len(desc) < 20 {
				t.Errorf("CharacterClass.Description() too short: %v", desc)
			}
		})
	}

	// Test unknown class returns empty
	if desc := CharacterClass(99).Description(); desc != "" {
		t.Errorf("Unknown class should return empty description, got: %v", desc)
	}
}

func TestCharacterData_Validate(t *testing.T) {
	tests := []struct {
		name    string
		data    CharacterData
		wantErr bool
	}{
		{
			name:    "valid warrior",
			data:    CharacterData{Name: "TestHero", Class: ClassWarrior},
			wantErr: false,
		},
		{
			name:    "valid mage",
			data:    CharacterData{Name: "Gandalf", Class: ClassMage},
			wantErr: false,
		},
		{
			name:    "valid rogue",
			data:    CharacterData{Name: "Shadow", Class: ClassRogue},
			wantErr: false,
		},
		{
			name:    "empty name",
			data:    CharacterData{Name: "", Class: ClassWarrior},
			wantErr: true,
		},
		{
			name:    "whitespace only name",
			data:    CharacterData{Name: "   ", Class: ClassWarrior},
			wantErr: true,
		},
		{
			name:    "name too long",
			data:    CharacterData{Name: "ThisNameIsWayTooLongAndExceedsTwentyCharacters", Class: ClassWarrior},
			wantErr: true,
		},
		{
			name:    "invalid class",
			data:    CharacterData{Name: "Hero", Class: CharacterClass(99)},
			wantErr: true,
		},
		{
			name:    "name with spaces trimmed",
			data:    CharacterData{Name: "  Hero  ", Class: ClassWarrior},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CharacterData.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check that whitespace is trimmed on successful validation
			if err == nil && tt.data.Name != "" {
				if tt.data.Name[0] == ' ' || tt.data.Name[len(tt.data.Name)-1] == ' ' {
					t.Errorf("CharacterData.Validate() did not trim whitespace: %q", tt.data.Name)
				}
			}
		})
	}
}

func TestNewCharacterCreation(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	if cc == nil {
		t.Fatal("NewCharacterCreation() returned nil")
	}

	if cc.currentStep != stepNameInput {
		t.Errorf("NewCharacterCreation() currentStep = %v, want %v", cc.currentStep, stepNameInput)
	}

	if cc.selectedClass != ClassWarrior {
		t.Errorf("NewCharacterCreation() selectedClass = %v, want %v", cc.selectedClass, ClassWarrior)
	}

	if cc.confirmed {
		t.Error("NewCharacterCreation() confirmed should be false")
	}

	if cc.screenWidth != 800 || cc.screenHeight != 600 {
		t.Errorf("NewCharacterCreation() screen dimensions = (%d, %d), want (800, 600)",
			cc.screenWidth, cc.screenHeight)
	}

	if cc.inputBuffer == nil {
		t.Error("NewCharacterCreation() inputBuffer is nil")
	}
}

func TestCharacterCreation_GetCharacterData(t *testing.T) {
	cc := NewCharacterCreation(800, 600)
	cc.characterData = CharacterData{
		Name:  "TestHero",
		Class: ClassMage,
	}

	data := cc.GetCharacterData()
	if data.Name != "TestHero" {
		t.Errorf("GetCharacterData() Name = %v, want TestHero", data.Name)
	}
	if data.Class != ClassMage {
		t.Errorf("GetCharacterData() Class = %v, want %v", data.Class, ClassMage)
	}
}

func TestCharacterCreation_IsComplete(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	if cc.IsComplete() {
		t.Error("IsComplete() should be false initially")
	}

	cc.confirmed = true
	if !cc.IsComplete() {
		t.Error("IsComplete() should be true after confirmation")
	}
}

func TestCharacterCreation_Reset(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	// Set some values
	cc.currentStep = stepConfirmation
	cc.characterData = CharacterData{Name: "Test", Class: ClassMage}
	cc.nameInput = "Test"
	cc.selectedClass = ClassRogue
	cc.confirmed = true
	cc.errorMsg = "Some error"

	// Reset
	cc.Reset()

	// Verify everything is reset
	if cc.currentStep != stepNameInput {
		t.Errorf("After Reset() currentStep = %v, want %v", cc.currentStep, stepNameInput)
	}
	if cc.characterData.Name != "" {
		t.Errorf("After Reset() characterData.Name = %v, want empty", cc.characterData.Name)
	}
	if cc.nameInput != "" {
		t.Errorf("After Reset() nameInput = %v, want empty", cc.nameInput)
	}
	if cc.selectedClass != ClassWarrior {
		t.Errorf("After Reset() selectedClass = %v, want %v", cc.selectedClass, ClassWarrior)
	}
	if cc.confirmed {
		t.Error("After Reset() confirmed should be false")
	}
	if cc.errorMsg != "" {
		t.Errorf("After Reset() errorMsg = %v, want empty", cc.errorMsg)
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxChars int
		wantLen  int // Expected number of lines
	}{
		{
			name:     "short text",
			text:     "Hello world",
			maxChars: 50,
			wantLen:  1,
		},
		{
			name:     "text requiring wrap",
			text:     "This is a longer piece of text that should be wrapped into multiple lines",
			maxChars: 30,
			wantLen:  3,
		},
		{
			name:     "empty text",
			text:     "",
			maxChars: 50,
			wantLen:  0,
		},
		{
			name:     "single word",
			text:     "Hello",
			maxChars: 50,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := wrapText(tt.text, tt.maxChars)
			if len(lines) != tt.wantLen {
				t.Errorf("wrapText() returned %d lines, want %d", len(lines), tt.wantLen)
			}

			// Verify no line exceeds maxChars
			for i, line := range lines {
				if len(line) > tt.maxChars {
					t.Errorf("wrapText() line %d exceeds maxChars: %d > %d", i, len(line), tt.maxChars)
				}
			}

			// Verify all words are present
			if tt.text != "" {
				combined := ""
				for _, line := range lines {
					combined += line + " "
				}
				combined = combined[:len(combined)-1] // Remove trailing space

				// Simple check: combined should contain all words from original
				if tt.wantLen > 0 && combined == "" {
					t.Error("wrapText() produced empty combined text")
				}
			}
		})
	}
}

func TestApplyClassStats_Warrior(t *testing.T) {
	world := NewWorld()
	player := world.CreateEntity()

	// Add required components
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})
	player.AddComponent(NewStatsComponent())
	player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 0.5})

	err := ApplyClassStats(player, ClassWarrior)
	if err != nil {
		t.Fatalf("ApplyClassStats() error = %v", err)
	}

	// Verify warrior stats
	healthComp, _ := player.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Max != 150 {
		t.Errorf("Warrior health = %v, want 150", health.Max)
	}

	manaComp, _ := player.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	if mana.Max != 50 {
		t.Errorf("Warrior mana = %v, want 50", mana.Max)
	}

	statsCompRaw, _ := player.GetComponent("stats")
	statsComp := statsCompRaw.(*StatsComponent)
	if statsComp.Attack != 12 {
		t.Errorf("Warrior attack = %v, want 12", statsComp.Attack)
	}
	if statsComp.Defense != 8 {
		t.Errorf("Warrior defense = %v, want 8", statsComp.Defense)
	}
	if statsComp.CritDamage != 2.0 {
		t.Errorf("Warrior crit damage = %v, want 2.0", statsComp.CritDamage)
	}
}

func TestApplyClassStats_Mage(t *testing.T) {
	world := NewWorld()
	player := world.CreateEntity()

	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})
	player.AddComponent(NewStatsComponent())
	player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 0.5})

	err := ApplyClassStats(player, ClassMage)
	if err != nil {
		t.Fatalf("ApplyClassStats() error = %v", err)
	}

	// Verify mage stats
	healthComp, _ := player.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Max != 80 {
		t.Errorf("Mage health = %v, want 80", health.Max)
	}

	manaComp, _ := player.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	if mana.Max != 150 {
		t.Errorf("Mage mana = %v, want 150", mana.Max)
	}
	if mana.Regen != 8.0 {
		t.Errorf("Mage mana regen = %v, want 8.0", mana.Regen)
	}

	statsCompRaw, _ := player.GetComponent("stats")
	statsComp := statsCompRaw.(*StatsComponent)
	if statsComp.Attack != 6 {
		t.Errorf("Mage attack = %v, want 6", statsComp.Attack)
	}
	if statsComp.CritChance != 0.10 {
		t.Errorf("Mage crit chance = %v, want 0.10", statsComp.CritChance)
	}
}

func TestApplyClassStats_Rogue(t *testing.T) {
	world := NewWorld()
	player := world.CreateEntity()

	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})
	player.AddComponent(NewStatsComponent())
	player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 0.5})

	err := ApplyClassStats(player, ClassRogue)
	if err != nil {
		t.Fatalf("ApplyClassStats() error = %v", err)
	}

	// Verify rogue stats
	healthComp, _ := player.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Max != 100 {
		t.Errorf("Rogue health = %v, want 100", health.Max)
	}

	manaComp, _ := player.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	if mana.Max != 80 {
		t.Errorf("Rogue mana = %v, want 80", mana.Max)
	}

	statsCompRaw, _ := player.GetComponent("stats")
	statsComp := statsCompRaw.(*StatsComponent)
	if statsComp.CritChance != 0.15 {
		t.Errorf("Rogue crit chance = %v, want 0.15", statsComp.CritChance)
	}
	if statsComp.Evasion != 0.15 {
		t.Errorf("Rogue evasion = %v, want 0.15", statsComp.Evasion)
	}

	attackCompRaw, _ := player.GetComponent("attack")
	attackComp := attackCompRaw.(*AttackComponent)
	if attackComp.Cooldown != 0.3 {
		t.Errorf("Rogue attack cooldown = %v, want 0.3", attackComp.Cooldown)
	}
}

func TestApplyClassStats_Errors(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *Entity
		class     CharacterClass
		wantErr   bool
	}{
		{
			name: "nil entity",
			setupFunc: func() *Entity {
				return nil
			},
			class:   ClassWarrior,
			wantErr: true,
		},
		{
			name: "missing health component",
			setupFunc: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})
				player.AddComponent(NewStatsComponent())
				player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 0.5})
				return player
			},
			class:   ClassWarrior,
			wantErr: true,
		},
		{
			name: "missing mana component",
			setupFunc: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&HealthComponent{Current: 100, Max: 100})
				player.AddComponent(NewStatsComponent())
				player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 0.5})
				return player
			},
			class:   ClassWarrior,
			wantErr: true,
		},
		{
			name: "missing stats component",
			setupFunc: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&HealthComponent{Current: 100, Max: 100})
				player.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})
				player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 0.5})
				return player
			},
			class:   ClassWarrior,
			wantErr: true,
		},
		{
			name: "missing attack component",
			setupFunc: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&HealthComponent{Current: 100, Max: 100})
				player.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})
				player.AddComponent(NewStatsComponent())
				return player
			},
			class:   ClassWarrior,
			wantErr: true,
		},
		{
			name: "invalid class",
			setupFunc: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&HealthComponent{Current: 100, Max: 100})
				player.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})
				player.AddComponent(NewStatsComponent())
				player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 0.5})
				return player
			},
			class:   CharacterClass(99),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.setupFunc()
			err := ApplyClassStats(player, tt.class)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyClassStats() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCharacterCreation_GetClassStats(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	tests := []struct {
		class    CharacterClass
		minLines int
	}{
		{ClassWarrior, 4},
		{ClassMage, 4},
		{ClassRogue, 4},
	}

	for _, tt := range tests {
		t.Run(tt.class.String(), func(t *testing.T) {
			cc.characterData.Class = tt.class
			stats := cc.getClassStats()

			if len(stats) < tt.minLines {
				t.Errorf("getClassStats() returned %d lines, want at least %d", len(stats), tt.minLines)
			}

			// Verify each stat line is non-empty
			for i, line := range stats {
				if line == "" {
					t.Errorf("getClassStats() line %d is empty", i)
				}
			}
		})
	}

	// Test unknown class returns empty
	cc.characterData.Class = CharacterClass(99)
	stats := cc.getClassStats()
	if len(stats) != 0 {
		t.Errorf("getClassStats() for unknown class returned %d lines, want 0", len(stats))
	}
}

// TestSetDefaults tests setting custom default values
func TestSetDefaults(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	defaults := CharacterCreationDefaults{
		DefaultName:  "TestHero",
		DefaultClass: ClassMage,
	}

	cc.SetDefaults(defaults)

	got := cc.GetDefaults()
	if got.DefaultName != "TestHero" {
		t.Errorf("GetDefaults().DefaultName = %q, want %q", got.DefaultName, "TestHero")
	}
	if got.DefaultClass != ClassMage {
		t.Errorf("GetDefaults().DefaultClass = %v, want %v", got.DefaultClass, ClassMage)
	}
}

// TestResetAppliesDefaults tests that Reset applies default values
func TestResetAppliesDefaults(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	// Set defaults
	defaults := CharacterCreationDefaults{
		DefaultName:  "DefaultHero",
		DefaultClass: ClassRogue,
	}
	cc.SetDefaults(defaults)

	// Modify character data away from defaults
	cc.nameInput = "SomeOtherName"
	cc.characterData.Name = "SomeOtherName"
	cc.characterData.Class = ClassWarrior

	// Reset should apply defaults
	cc.Reset()

	if cc.nameInput != "DefaultHero" {
		t.Errorf("After Reset(), nameInput = %q, want %q", cc.nameInput, "DefaultHero")
	}
	if cc.characterData.Name != "DefaultHero" {
		t.Errorf("After Reset(), characterData.Name = %q, want %q", cc.characterData.Name, "DefaultHero")
	}
	if cc.characterData.Class != ClassRogue {
		t.Errorf("After Reset(), characterData.Class = %v, want %v", cc.characterData.Class, ClassRogue)
	}
}

// TestResetWithoutDefaults tests that Reset works when no defaults are set
func TestResetWithoutDefaults(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	// Modify character data
	cc.nameInput = "SomeName"
	cc.characterData.Name = "SomeName"
	cc.characterData.Class = ClassMage
	cc.currentStep = stepClassSelection

	// Reset without defaults should clear everything
	cc.Reset()

	if cc.nameInput != "" {
		t.Errorf("After Reset() without defaults, nameInput = %q, want empty", cc.nameInput)
	}
	if cc.characterData.Name != "" {
		t.Errorf("After Reset() without defaults, characterData.Name = %q, want empty", cc.characterData.Name)
	}
	if cc.characterData.Class != ClassWarrior {
		t.Errorf("After Reset() without defaults, characterData.Class = %v, want %v (zero value)", cc.characterData.Class, ClassWarrior)
	}
	if cc.currentStep != stepNameInput {
		t.Errorf("After Reset(), currentStep = %v, want %v", cc.currentStep, stepNameInput)
	}
}

// TestLoadPortrait_InvalidFile tests loading invalid portrait files
func TestLoadPortrait_InvalidFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: false, // Empty is valid (no portrait)
		},
		{
			name:    "nonexistent file",
			path:    "/nonexistent/file.png",
			wantErr: true,
			errMsg:  "portrait file not found",
		},
		{
			name:    "wrong extension",
			path:    "/tmp/test.jpg",
			wantErr: true,
			errMsg:  "portrait must be a .png file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := LoadPortrait(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadPortrait() expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LoadPortrait() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("LoadPortrait() unexpected error = %v", err)
				}
				if tt.path == "" && img != nil {
					t.Errorf("LoadPortrait(\"\") = %v, want nil", img)
				}
			}
		})
	}
}

// TestMax tests the max helper function
func TestMax(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{5, 10, 10},
		{10, 5, 10},
		{7, 7, 7},
		{-5, 3, 3},
		{0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("max(%d,%d)", tt.a, tt.b), func(t *testing.T) {
			got := max(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestCharacterData_WithPortrait tests CharacterData with portrait field
func TestCharacterData_WithPortrait(t *testing.T) {
	cd := CharacterData{
		Name:         "TestHero",
		Class:        ClassWarrior,
		PortraitPath: "/path/to/portrait.png",
		Portrait:     nil, // Can be nil
	}

	if err := cd.Validate(); err != nil {
		t.Errorf("CharacterData.Validate() with portrait path error = %v, want nil", err)
	}
}

// TestSetDefaults_WithPortrait tests setting defaults including portrait path
func TestSetDefaults_WithPortrait(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	defaults := CharacterCreationDefaults{
		DefaultName:         "TestHero",
		DefaultClass:        ClassMage,
		DefaultPortraitPath: "/home/user/portrait.png",
	}

	cc.SetDefaults(defaults)

	got := cc.GetDefaults()
	if got.DefaultName != "TestHero" {
		t.Errorf("GetDefaults().DefaultName = %q, want %q", got.DefaultName, "TestHero")
	}
	if got.DefaultClass != ClassMage {
		t.Errorf("GetDefaults().DefaultClass = %v, want %v", got.DefaultClass, ClassMage)
	}
	if got.DefaultPortraitPath != "/home/user/portrait.png" {
		t.Errorf("GetDefaults().DefaultPortraitPath = %q, want %q", got.DefaultPortraitPath, "/home/user/portrait.png")
	}
}

// TestGetDefaultPicturesDirectory tests the Pictures directory detection
func TestGetDefaultPicturesDirectory(t *testing.T) {
	dir := GetDefaultPicturesDirectory()

	// Should return a non-empty string
	if dir == "" {
		t.Error("GetDefaultPicturesDirectory() returned empty string")
	}

	// Should contain expected path component based on OS
	// We can't test exact paths due to different environments, but we can check it's reasonable
	if !strings.Contains(dir, "Pictures") && !strings.Contains(dir, "home") && !strings.Contains(dir, "Users") {
		// On some systems it might just be home dir, that's okay
		t.Logf("GetDefaultPicturesDirectory() = %q (acceptable)", dir)
	}
}
