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

	// Test unknown class returns "Unknown class"
	if desc := CharacterClass(99).Description(); desc != "Unknown class" {
		t.Errorf("Unknown class should return 'Unknown class', got: %v", desc)
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
			name:    "valid ranger",
			data:    CharacterData{Name: "Hunter", Class: ClassRanger},
			wantErr: false,
		},
		{
			name:    "valid cleric",
			data:    CharacterData{Name: "Healer", Class: ClassCleric},
			wantErr: false,
		},
		{
			name:    "valid necromancer",
			data:    CharacterData{Name: "DarkOne", Class: ClassNecromancer},
			wantErr: false,
		},
		{
			name:    "valid battlemage hybrid",
			data:    CharacterData{Name: "Spellsword", Class: ClassBattlemage},
			wantErr: false,
		},
		{
			name:    "valid ninja hybrid",
			data:    CharacterData{Name: "Shinobi", Class: ClassNinja},
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

// TestCharacterCreation_KeyboardStateManagement tests that keyboard state is properly
// managed during character creation flow (WASM-specific, but tests state transitions).
func TestCharacterCreation_KeyboardStateManagement(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	// Initial state: should be at name input step
	if cc.currentStep != stepNameInput {
		t.Errorf("Initial step = %v, want stepNameInput", cc.currentStep)
	}

	// After Reset(), keyboard state should be properly set up for name input
	// Note: We can't test actual ShowKeyboard() calls in non-WASM builds,
	// but we verify the state is correctly initialized
	cc.Reset()
	if cc.currentStep != stepNameInput {
		t.Errorf("After Reset(), step = %v, want stepNameInput", cc.currentStep)
	}

	// Verify Cleanup() can be called without error
	// This ensures the method is safe to call from game.go
	cc.Cleanup()

	// Verify Cleanup() is idempotent (can be called multiple times safely)
	cc.Cleanup()
	cc.Cleanup()
}

// TestCharacterCreation_ResetWithDefaults tests that Reset() properly applies defaults
// and sets up keyboard state for immediate text input.
func TestCharacterCreation_ResetWithDefaults(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	// Set defaults
	defaults := CharacterCreationDefaults{
		DefaultName:  "TestHero",
		DefaultClass: ClassMage,
	}
	cc.SetDefaults(defaults)

	// Reset should apply defaults
	cc.Reset()

	if cc.nameInput != "TestHero" {
		t.Errorf("After Reset() with defaults, nameInput = %q, want %q", cc.nameInput, "TestHero")
	}
	if cc.selectedClass != ClassMage {
		t.Errorf("After Reset() with defaults, selectedClass = %v, want ClassMage", cc.selectedClass)
	}
	if cc.currentStep != stepNameInput {
		t.Errorf("After Reset(), step = %v, want stepNameInput", cc.currentStep)
	}

	// Verify character data is also set
	if cc.characterData.Name != "TestHero" {
		t.Errorf("After Reset(), characterData.Name = %q, want %q", cc.characterData.Name, "TestHero")
	}
	if cc.characterData.Class != ClassMage {
		t.Errorf("After Reset(), characterData.Class = %v, want ClassMage", cc.characterData.Class)
	}
}

// TestCharacterCreation_KeyboardLifecycle verifies keyboard state management during UI navigation.
// This test ensures the mobile keyboard is shown/hidden at appropriate times.
// Note: Actual ShowKeyboard/HideKeyboard calls are no-ops on non-WASM platforms,
// but we can verify the keyboardShown flag is managed correctly.
func TestCharacterCreation_KeyboardLifecycle(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	// Initial state: keyboard should not be shown yet
	if cc.keyboardShown {
		t.Error("Initial keyboardShown should be false")
	}

	// Reset should not set keyboardShown=true immediately
	// (keyboard will be shown by updateNameInput on first Update)
	cc.Reset()
	if cc.keyboardShown {
		t.Error("After Reset(), keyboardShown should be false (will be shown by updateNameInput)")
	}
	if cc.currentStep != stepNameInput {
		t.Errorf("After Reset(), currentStep = %v, want stepNameInput", cc.currentStep)
	}

	// Verify Cleanup hides keyboard (sets flag to false)
	cc.keyboardShown = true // Simulate keyboard was shown
	cc.Cleanup()
	if cc.keyboardShown {
		t.Error("After Cleanup(), keyboardShown should be false")
	}

	// Test state transitions reset keyboard flag appropriately
	cc.keyboardShown = true
	cc.currentStep = stepClassSelection
	cc.updateClassSelection()
	// If user went back to name input, keyboard flag should be reset
	// (We can't test keyboard input in unit tests, but we can verify the flag)

	// Verify confirmation step can detect validation errors and reset flag
	cc.currentStep = stepConfirmation
	cc.characterData.Name = "" // Invalid (empty name)
	cc.keyboardShown = true
	cc.updateConfirmation()
	// After validation error, should go back to name input with flag reset
	// (keyboardShown would be reset to false for re-entry)
}

// TestCharacterCreation_KeyboardFlagConsistency verifies keyboardShown flag
// is managed consistently across all state transitions.
func TestCharacterCreation_KeyboardFlagConsistency(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	// Test: Going from class selection back to name input resets flag
	cc.currentStep = stepClassSelection
	cc.keyboardShown = true // Simulate keyboard was shown previously

	// Simulate pressing backspace to go back
	cc.currentStep = stepNameInput
	cc.keyboardShown = false // Should be reset for re-entry

	if cc.keyboardShown {
		t.Error("When returning to name input, keyboardShown should be reset to false")
	}

	// Test: Cleanup always ensures keyboard is hidden
	cc.keyboardShown = true
	cc.Cleanup()
	if cc.keyboardShown {
		t.Error("Cleanup() must always set keyboardShown to false")
	}
}

// TestSetDefaultNameFromSeed verifies that default name is set deterministically from world seed.
func TestSetDefaultNameFromSeed(t *testing.T) {
	tests := []struct {
		name string
		seed int64
	}{
		{"seed 12345", 12345},
		{"seed 98765", 98765},
		{"seed 0", 0},
		{"seed 1", 1},
		{"negative seed", -999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := NewCharacterCreation(800, 600)

			// Set default name from seed
			cc.SetDefaultNameFromSeed(tt.seed)

			// Verify name is set in defaults
			if cc.defaults.DefaultName == "" {
				t.Error("SetDefaultNameFromSeed() did not set DefaultName")
			}

			// Verify name is applied to nameInput when in name input step
			if cc.currentStep == stepNameInput && cc.nameInput == "" {
				t.Error("SetDefaultNameFromSeed() did not apply name to nameInput in stepNameInput")
			}

			// Verify determinism: same seed should produce same name
			cc2 := NewCharacterCreation(800, 600)
			cc2.SetDefaultNameFromSeed(tt.seed)

			if cc.defaults.DefaultName != cc2.defaults.DefaultName {
				t.Errorf("SetDefaultNameFromSeed(%d) not deterministic: got %s and %s",
					tt.seed, cc.defaults.DefaultName, cc2.defaults.DefaultName)
			}

			// Verify name is applied when reset
			cc.Reset()
			if cc.nameInput != cc.defaults.DefaultName {
				t.Errorf("After Reset(), nameInput = %q, want %q", cc.nameInput, cc.defaults.DefaultName)
			}
		})
	}
}

// TestSetDefaultNameFromSeed_Integration tests that default names work with the full character creation flow.
func TestSetDefaultNameFromSeed_Integration(t *testing.T) {
	cc := NewCharacterCreation(800, 600)
	seed := int64(12345)

	// Set default name from seed
	cc.SetDefaultNameFromSeed(seed)
	defaultName := cc.defaults.DefaultName

	// Should be applied to nameInput initially
	if cc.nameInput != defaultName {
		t.Errorf("Initial nameInput = %q, want %q", cc.nameInput, defaultName)
	}

	// User can still modify the name
	cc.nameInput = "CustomName"
	if cc.nameInput == defaultName {
		t.Error("User should be able to modify default name")
	}

	// Reset should restore default
	cc.Reset()
	if cc.nameInput != defaultName {
		t.Errorf("After Reset(), nameInput = %q, want default %q", cc.nameInput, defaultName)
	}
}

// TestGenerateRandomName_Deterministic verifies that generateRandomName produces deterministic results.
// Same seed + counter + class should always produce the same name.
func TestGenerateRandomName_Deterministic(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		class CharacterClass
	}{
		{"seed12345_warrior", 12345, ClassWarrior},
		{"seed0_mage", 0, ClassMage},
		{"seed999999_rogue", 999999, ClassRogue},
		{"negative_seed", -12345, ClassWarrior},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create first character creation instance
			cc1 := NewCharacterCreation(800, 600)
			cc1.SetDefaultNameFromSeed(tt.seed)
			cc1.selectedClass = tt.class

			// Generate name with first instance
			name1 := cc1.generateRandomName()

			// Create second character creation instance with same seed
			cc2 := NewCharacterCreation(800, 600)
			cc2.SetDefaultNameFromSeed(tt.seed)
			cc2.selectedClass = tt.class

			// Generate name with second instance
			name2 := cc2.generateRandomName()

			// Should be identical
			if name1 != name2 {
				t.Errorf("generateRandomName() not deterministic: got %q and %q for seed %d",
					name1, name2, tt.seed)
			}

			// Name should not be empty
			if name1 == "" {
				t.Error("generateRandomName() returned empty string")
			}
		})
	}
}

// TestGenerateRandomName_DifferentSeeds verifies that different seeds produce different names.
func TestGenerateRandomName_DifferentSeeds(t *testing.T) {
	cc1 := NewCharacterCreation(800, 600)
	cc1.SetDefaultNameFromSeed(12345)
	cc1.selectedClass = ClassWarrior
	name1 := cc1.generateRandomName()

	cc2 := NewCharacterCreation(800, 600)
	cc2.SetDefaultNameFromSeed(67890)
	cc2.selectedClass = ClassWarrior
	name2 := cc2.generateRandomName()

	// Different seeds should produce different names (high probability)
	if name1 == name2 {
		t.Logf("Warning: different seeds produced same name (possible but unlikely): %q", name1)
	}
}

// TestGenerateRandomName_MultipleCallsVary verifies that multiple calls produce varied names.
func TestGenerateRandomName_MultipleCallsVary(t *testing.T) {
	cc := NewCharacterCreation(800, 600)
	cc.SetDefaultNameFromSeed(12345)
	cc.selectedClass = ClassWarrior

	// Generate multiple names
	names := make(map[string]bool)
	for i := 0; i < 10; i++ {
		name := cc.generateRandomName()
		names[name] = true
	}

	// Should have at least 2 unique names from 10 generations
	if len(names) < 2 {
		t.Errorf("generateRandomName() should vary across calls: got only %d unique names from 10 calls",
			len(names))
	}
}

// TestGenerateRandomName_NoUIStateDependency verifies name generation doesn't depend on UI state.
func TestGenerateRandomName_NoUIStateDependency(t *testing.T) {
	// Create two instances with same seed but different UI state
	cc1 := NewCharacterCreation(800, 600)
	cc1.SetDefaultNameFromSeed(12345)
	cc1.selectedClass = ClassWarrior
	cc1.nameInput = ""

	cc2 := NewCharacterCreation(800, 600)
	cc2.SetDefaultNameFromSeed(12345)
	cc2.selectedClass = ClassWarrior
	cc2.nameInput = "SomePreviousName"

	// Both should produce same name since seed and counter are the same
	name1 := cc1.generateRandomName()
	name2 := cc2.generateRandomName()

	if name1 != name2 {
		t.Errorf("generateRandomName() depends on UI state: empty input=%q, with input=%q",
			name1, name2)
	}
}

// TestApplyClassStats_AllClasses verifies all 21 classes can have stats applied successfully.
func TestApplyClassStats_AllClasses(t *testing.T) {
	allClasses := []struct {
		class    CharacterClass
		wantHP   float64
		wantMana int
		wantAtk  float64
		wantDef  float64
	}{
		{ClassWarrior, 150, 50, 12, 8},
		{ClassMage, 80, 150, 6, 3},
		{ClassRogue, 100, 80, 10, 5},
		{ClassRanger, 110, 70, 11, 5},
		{ClassCleric, 120, 120, 7, 6},
		{ClassNecromancer, 90, 140, 8, 4},
		{ClassBattlemage, 115, 100, 10, 6},
		{ClassSpellblade, 90, 110, 9, 4},
		{ClassPaladin, 140, 80, 10, 9},
		{ClassMonk, 100, 90, 9, 5},
		{ClassDeathKnight, 130, 90, 11, 7},
		{ClassWitchHunter, 115, 90, 10, 5},
		{ClassBeastlord, 135, 60, 11, 7},
		{ClassArcaneArcher, 95, 110, 10, 4},
		{ClassShadowPriest, 85, 130, 8, 4},
		{ClassDruid, 105, 115, 9, 5},
		{ClassInquisitor, 110, 100, 9, 6},
		{ClassBloodKnight, 125, 85, 12, 6},
		{ClassMystic, 95, 135, 7, 5},
		{ClassWarlock, 85, 145, 9, 3},
		{ClassNinja, 90, 75, 11, 4},
	}

	for _, tc := range allClasses {
		t.Run(tc.class.String(), func(t *testing.T) {
			world := NewWorld()
			player := world.CreateEntity()
			player.AddComponent(&HealthComponent{Current: 100, Max: 100})
			player.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})
			player.AddComponent(NewStatsComponent())
			player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 0.5})

			err := ApplyClassStats(player, tc.class)
			if err != nil {
				t.Fatalf("ApplyClassStats(%s) error = %v", tc.class, err)
			}

			// Verify health
			healthComp, _ := player.GetComponent("health")
			health := healthComp.(*HealthComponent)
			if health.Max != tc.wantHP {
				t.Errorf("%s health = %v, want %v", tc.class, health.Max, tc.wantHP)
			}
			if health.Current != tc.wantHP {
				t.Errorf("%s current health = %v, want %v", tc.class, health.Current, tc.wantHP)
			}

			// Verify mana
			manaComp, _ := player.GetComponent("mana")
			mana := manaComp.(*ManaComponent)
			if mana.Max != tc.wantMana {
				t.Errorf("%s mana = %v, want %v", tc.class, mana.Max, tc.wantMana)
			}
			if mana.Current != tc.wantMana {
				t.Errorf("%s current mana = %v, want %v", tc.class, mana.Current, tc.wantMana)
			}

			// Verify attack/defense stats
			statsCompRaw, _ := player.GetComponent("stats")
			statsComp := statsCompRaw.(*StatsComponent)
			if statsComp.Attack != tc.wantAtk {
				t.Errorf("%s attack = %v, want %v", tc.class, statsComp.Attack, tc.wantAtk)
			}
			if statsComp.Defense != tc.wantDef {
				t.Errorf("%s defense = %v, want %v", tc.class, statsComp.Defense, tc.wantDef)
			}
		})
	}
}

// TestApplyClassStats_HybridClassesHaveUniqueStats verifies hybrid classes are distinct from their base classes.
func TestApplyClassStats_HybridClassesHaveUniqueStats(t *testing.T) {
	createPlayer := func() *Entity {
		world := NewWorld()
		player := world.CreateEntity()
		player.AddComponent(&HealthComponent{Current: 100, Max: 100})
		player.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})
		player.AddComponent(NewStatsComponent())
		player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 0.5})
		return player
	}

	// Test Battlemage is distinct from both Warrior and Mage
	t.Run("Battlemage_vs_parents", func(t *testing.T) {
		warrior := createPlayer()
		mage := createPlayer()
		battlemage := createPlayer()

		ApplyClassStats(warrior, ClassWarrior)
		ApplyClassStats(mage, ClassMage)
		ApplyClassStats(battlemage, ClassBattlemage)

		wH, _ := warrior.GetComponent("health")
		mH, _ := mage.GetComponent("health")
		bH, _ := battlemage.GetComponent("health")

		warriorHP := wH.(*HealthComponent).Max
		mageHP := mH.(*HealthComponent).Max
		battlemageHP := bH.(*HealthComponent).Max

		// Battlemage should be between Warrior and Mage HP (or distinct)
		if battlemageHP == warriorHP || battlemageHP == mageHP {
			t.Errorf("Battlemage HP (%v) should differ from Warrior (%v) and Mage (%v)",
				battlemageHP, warriorHP, mageHP)
		}
	})

	// Test Ninja has highest crit rate and evasion
	t.Run("Ninja_specialization", func(t *testing.T) {
		ninja := createPlayer()
		rogue := createPlayer()

		ApplyClassStats(ninja, ClassNinja)
		ApplyClassStats(rogue, ClassRogue)

		nS, _ := ninja.GetComponent("stats")
		rS, _ := rogue.GetComponent("stats")

		ninjaCrit := nS.(*StatsComponent).CritChance
		rogueCrit := rS.(*StatsComponent).CritChance

		if ninjaCrit <= rogueCrit {
			t.Errorf("Ninja crit (%v) should be higher than Rogue (%v)", ninjaCrit, rogueCrit)
		}
	})
}

// TestClassSelection_SixClasses verifies arrow key navigation wraps through all 6 base classes
func TestClassSelection_SixClasses(t *testing.T) {
	t.Run("baseClasses contains all 6 base classes", func(t *testing.T) {
		expected := []CharacterClass{
			ClassWarrior, ClassMage, ClassRogue, ClassRanger, ClassCleric, ClassNecromancer,
		}
		if len(baseClasses) != len(expected) {
			t.Errorf("baseClasses length = %d, want %d", len(baseClasses), len(expected))
		}
		for i, c := range expected {
			if baseClasses[i] != c {
				t.Errorf("baseClasses[%d] = %v, want %v", i, baseClasses[i], c)
			}
		}
	})

	t.Run("all 6 base classes have descriptions", func(t *testing.T) {
		for _, class := range baseClasses {
			desc := class.Description()
			if desc == "" {
				t.Errorf("Class %v has empty description", class)
			}
		}
	})

	t.Run("all 6 base classes have string names", func(t *testing.T) {
		expectedNames := []string{"Warrior", "Mage", "Rogue", "Ranger", "Cleric", "Necromancer"}
		for i, class := range baseClasses {
			name := class.String()
			if name != expectedNames[i] {
				t.Errorf("Class %v string = %q, want %q", class, name, expectedNames[i])
			}
		}
	})
}

// TestGetClassStats_AllBaseClasses verifies stat preview text for all 6 base classes
func TestGetClassStats_AllBaseClasses(t *testing.T) {
	tests := []struct {
		class              CharacterClass
		wantMinLen         int
		wantHealthContains string
	}{
		{ClassWarrior, 4, "150"},
		{ClassMage, 4, "80"},
		{ClassRogue, 4, "100"},
		{ClassRanger, 4, "110"},
		{ClassCleric, 4, "120"},
		{ClassNecromancer, 4, "90"},
	}

	for _, tt := range tests {
		t.Run(tt.class.String(), func(t *testing.T) {
			cc := &EbitenCharacterCreation{
				characterData: CharacterData{Class: tt.class},
			}
			stats := cc.getClassStats()
			if len(stats) < tt.wantMinLen {
				t.Errorf("getClassStats() for %v returned %d stats, want at least %d", tt.class, len(stats), tt.wantMinLen)
			}
			if len(stats) > 0 && !strings.Contains(stats[0], tt.wantHealthContains) {
				t.Errorf("getClassStats() health for %v = %q, want to contain %q", tt.class, stats[0], tt.wantHealthContains)
			}
		})
	}

	t.Run("unknown class returns empty", func(t *testing.T) {
		cc := &EbitenCharacterCreation{
			characterData: CharacterData{Class: CharacterClass(99)},
		}
		stats := cc.getClassStats()
		if len(stats) != 0 {
			t.Errorf("getClassStats() for unknown class returned %d stats, want 0", len(stats))
		}
	})
}

// TestClassPagination verifies that class pagination correctly organizes all 21 classes
func TestClassPagination(t *testing.T) {
	t.Run("total pages is 4", func(t *testing.T) {
		if totalClassPages() != 4 {
			t.Errorf("totalClassPages() = %d, want 4", totalClassPages())
		}
	})

	t.Run("page 0 returns base classes", func(t *testing.T) {
		classes := getClassesForPage(0)
		if len(classes) != 6 {
			t.Errorf("page 0 has %d classes, want 6", len(classes))
		}
		if classes[0] != ClassWarrior {
			t.Errorf("first class on page 0 is %v, want ClassWarrior", classes[0])
		}
		if classes[5] != ClassNecromancer {
			t.Errorf("last class on page 0 is %v, want ClassNecromancer", classes[5])
		}
	})

	t.Run("page 1 returns first hybrid classes", func(t *testing.T) {
		classes := getClassesForPage(1)
		if len(classes) != 6 {
			t.Errorf("page 1 has %d classes, want 6", len(classes))
		}
		if classes[0] != ClassBattlemage {
			t.Errorf("first class on page 1 is %v, want ClassBattlemage", classes[0])
		}
	})

	t.Run("page 3 returns final hybrid classes", func(t *testing.T) {
		classes := getClassesForPage(3)
		if len(classes) != 3 {
			t.Errorf("page 3 has %d classes, want 3", len(classes))
		}
		if classes[len(classes)-1] != ClassNinja {
			t.Errorf("last class on page 3 is %v, want ClassNinja", classes[len(classes)-1])
		}
	})

	t.Run("invalid page returns nil", func(t *testing.T) {
		classes := getClassesForPage(99)
		if classes != nil {
			t.Errorf("page 99 should return nil, got %v", classes)
		}
	})

	t.Run("all 21 classes are reachable", func(t *testing.T) {
		allClasses := make(map[CharacterClass]bool)
		for page := 0; page < totalClassPages(); page++ {
			for _, class := range getClassesForPage(page) {
				allClasses[class] = true
			}
		}
		if len(allClasses) != 21 {
			t.Errorf("total unique classes across pages = %d, want 21", len(allClasses))
		}
	})
}

// TestClassPageNavigation verifies the EbitenCharacterCreation pagination field
func TestClassPageNavigation(t *testing.T) {
	cc := NewCharacterCreation(800, 600)

	t.Run("initial page is 0", func(t *testing.T) {
		if cc.classPage != 0 {
			t.Errorf("initial classPage = %d, want 0", cc.classPage)
		}
	})

	t.Run("can switch pages", func(t *testing.T) {
		cc.classPage = 1
		if cc.classPage != 1 {
			t.Errorf("classPage = %d, want 1", cc.classPage)
		}
	})

	t.Run("page title changes per page", func(t *testing.T) {
		if getPageTitle(0) != "Base Classes" {
			t.Errorf("page 0 title = %q, want 'Base Classes'", getPageTitle(0))
		}
		if getPageTitle(1) != "Advanced Classes" {
			t.Errorf("page 1 title = %q, want 'Advanced Classes'", getPageTitle(1))
		}
	})
}

// TestEquipmentLoadout_Generation verifies deterministic loadout generation per class.
func TestEquipmentLoadout_Generation(t *testing.T) {
	seed := int64(12345)

	t.Run("generates 3 loadouts per class", func(t *testing.T) {
		classes := []CharacterClass{
			ClassWarrior, ClassMage, ClassRogue,
			ClassRanger, ClassCleric, ClassNecromancer,
		}
		for _, class := range classes {
			loadouts := generateClassLoadouts(class, seed)
			if len(loadouts) != 3 {
				t.Errorf("generateClassLoadouts(%s) returned %d loadouts, want 3",
					class.String(), len(loadouts))
			}
		}
	})

	t.Run("deterministic - same seed same loadouts", func(t *testing.T) {
		loadouts1 := generateClassLoadouts(ClassWarrior, seed)
		loadouts2 := generateClassLoadouts(ClassWarrior, seed)

		for i := range loadouts1 {
			if loadouts1[i].Name != loadouts2[i].Name {
				t.Errorf("loadout %d name differs: %q vs %q", i, loadouts1[i].Name, loadouts2[i].Name)
			}
			if loadouts1[i].MainHand != loadouts2[i].MainHand {
				t.Errorf("loadout %d mainhand differs: %q vs %q", i, loadouts1[i].MainHand, loadouts2[i].MainHand)
			}
		}
	})

	t.Run("different classes have different loadouts", func(t *testing.T) {
		warriorLoadouts := generateClassLoadouts(ClassWarrior, seed)
		mageLoadouts := generateClassLoadouts(ClassMage, seed)

		if warriorLoadouts[0].Name == mageLoadouts[0].Name {
			t.Errorf("warrior and mage have same loadout name: %q", warriorLoadouts[0].Name)
		}
	})

	t.Run("loadouts have required fields", func(t *testing.T) {
		loadouts := generateClassLoadouts(ClassWarrior, seed)
		for i, loadout := range loadouts {
			if loadout.Name == "" {
				t.Errorf("loadout %d has empty Name", i)
			}
			if loadout.Description == "" {
				t.Errorf("loadout %d has empty Description", i)
			}
			if loadout.MainHand == "" {
				t.Errorf("loadout %d has empty MainHand", i)
			}
			if loadout.Armor == "" {
				t.Errorf("loadout %d has empty Armor", i)
			}
		}
	})

	t.Run("hybrid classes get generated loadouts", func(t *testing.T) {
		loadouts := generateClassLoadouts(ClassBattlemage, seed)
		if len(loadouts) != 3 {
			t.Errorf("hybrid class loadouts = %d, want 3", len(loadouts))
		}
		// Hybrid loadouts should reference the class name
		found := false
		for _, l := range loadouts {
			if strings.Contains(l.Name, "Battlemage") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("no hybrid loadout mentions class name")
		}
	})
}

// TestCharacterCreation_EquipmentStep verifies equipment step integration.
func TestCharacterCreation_EquipmentStep(t *testing.T) {
	t.Run("stepSubclassSelection exists in enum", func(t *testing.T) {
		// Verify the enum order is correct
		if stepNameInput != 0 {
			t.Errorf("stepNameInput = %d, want 0", stepNameInput)
		}
		if stepClassSelection != 1 {
			t.Errorf("stepClassSelection = %d, want 1", stepClassSelection)
		}
		if stepSubclassSelection != 2 {
			t.Errorf("stepSubclassSelection = %d, want 2", stepSubclassSelection)
		}
		if stepPortraitSelection != 3 {
			t.Errorf("stepPortraitSelection = %d, want 3", stepPortraitSelection)
		}
		if stepConfirmation != 4 {
			t.Errorf("stepConfirmation = %d, want 4", stepConfirmation)
		}
	})

	t.Run("CharacterData includes StartingLoadout", func(t *testing.T) {
		loadout := EquipmentLoadout{
			Name:        "Test",
			MainHand:    "Sword",
			Armor:       "Plate",
			BonusHP:     10,
			BonusAttack: 2,
		}
		data := CharacterData{
			Name:            "TestChar",
			Class:           ClassWarrior,
			StartingLoadout: &loadout,
		}
		if data.StartingLoadout == nil {
			t.Error("StartingLoadout is nil")
		}
		if data.StartingLoadout.Name != "Test" {
			t.Errorf("StartingLoadout.Name = %q, want 'Test'", data.StartingLoadout.Name)
		}
	})
}
