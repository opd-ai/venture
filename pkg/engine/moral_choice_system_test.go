package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

func TestNewMoralChoiceSystem(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()

	sys := NewMoralChoiceSystem(world, logger)

	if sys == nil {
		t.Fatal("NewMoralChoiceSystem returned nil")
	}

	if sys.world != world {
		t.Error("World not set correctly")
	}

	if sys.logger != logger {
		t.Error("Logger not set correctly")
	}
}

func TestNewMoralChoiceSystem_NilLogger(t *testing.T) {
	world := NewWorld()

	sys := NewMoralChoiceSystem(world, nil)

	if sys == nil {
		t.Fatal("NewMoralChoiceSystem returned nil")
	}

	if sys.logger == nil {
		t.Error("Logger should be created if nil passed")
	}
}

func TestMoralChoiceSystem_Update_RemovesExpiredChoices(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	//logger.SetLevel(logrus.FatalLevel) // Suppress logs during test
	sys := NewMoralChoiceSystem(world, logger)

	entity := world.CreateEntity()
	moralChoice := NewMoralChoiceComponent()

	// Add expired choice
	expiredChoice := MoralChoice{
		ID:          "expired",
		Description: "Old choice",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}
	moralChoice.AddChoice(expiredChoice)

	// Add valid choice
	validChoice := MoralChoice{
		ID:          "valid",
		Description: "Current choice",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	moralChoice.AddChoice(validChoice)

	entity.AddComponent(moralChoice)

	// Debug: Check expiry status before update
	// t.Logf("Before update: len=%d, expired=%v, valid=%v",
	//	len(moralChoice.PendingChoices),
	//	moralChoice.PendingChoices[0].IsExpired(),
	//	moralChoice.PendingChoices[1].IsExpired())

	// Update system
	sys.Update(0.016)

	// Debug: Check what remains
	//for i, ch := range moralChoice.PendingChoices {
	//	t.Logf("After update [%d]: ID=%s, Expired=%v", i, ch.ID, ch.IsExpired())
	//}

	// Check that expired choice was removed
	if len(moralChoice.PendingChoices) != 1 {
		// Show what we actually have
		for i, ch := range moralChoice.PendingChoices {
			t.Logf("Choice [%d]: ID=%s, ExpiresAt=%v, IsExpired=%v", i, ch.ID, ch.ExpiresAt, ch.IsExpired())
		}
		t.Errorf("Expected 1 pending choice, got %d", len(moralChoice.PendingChoices))
	}

	if moralChoice.PendingChoices[0].ID != "valid" {
		t.Error("Wrong choice removed")
	}
}

func TestMoralChoiceSystem_Update_CompletesRedemptionArcs(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	sys := NewMoralChoiceSystem(world, logger)

	entity := world.CreateEntity()
	moralChoice := NewMoralChoiceComponent()
	reputation := NewReputationComponent()
	reputation.SetReputation("TestFaction", 50.0)

	// Add completed redemption arc (all actions complete)
	arc := RedemptionArc{
		FactionName:        "TestFaction",
		StartingReputation: -30.0,
		TargetReputation:   10.0,
		CurrentReputation:  50.0,
		RequiredActions: []RedemptionAction{
			{Quantity: 5, Progress: 5}, // Complete
			{Quantity: 3, Progress: 3}, // Complete
		},
		CompletedActions: 2, // All actions complete
		StartTime:        time.Now(),
	}
	moralChoice.StartRedemption(arc)

	entity.AddComponent(moralChoice)
	entity.AddComponent(reputation)

	// Update system
	sys.Update(0.016)

	// Check that completed arc was removed
	if len(moralChoice.ActiveRedemptions) != 0 {
		t.Errorf("Expected 0 active redemptions, got %d", len(moralChoice.ActiveRedemptions))
	}
}

func TestMoralChoiceSystem_MakeChoice_Success(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	sys := NewMoralChoiceSystem(world, logger)

	entity := world.CreateEntity()
	moralChoice := NewMoralChoiceComponent()
	reputation := NewReputationComponent()
	experience := NewExperienceComponent()
	inventory := NewInventoryComponent(20, 100.0)
	position := &PositionComponent{X: 100, Y: 100}

	choice := MoralChoice{
		ID:          "test_choice",
		Description: "Help or harm?",
		Options: []ChoiceOption{
			{
				Label:       "Help",
				Description: "Aid the villagers",
				AlignmentImpact: AlignmentDelta{
					LawDelta:  0.05,
					GoodDelta: 0.1,
				},
				ReputationImpact: map[string]float64{
					"Villagers": 20.0,
				},
				Rewards: &ChoiceRewards{
					XP:   100,
					Gold: 50,
				},
			},
			{
				Label:       "Harm",
				Description: "Attack the villagers",
				AlignmentImpact: AlignmentDelta{
					LawDelta:  -0.1,
					GoodDelta: -0.15,
				},
				ReputationImpact: map[string]float64{
					"Villagers": -30.0,
				},
			},
		},
	}
	moralChoice.AddChoice(choice)

	entity.AddComponent(moralChoice)
	entity.AddComponent(reputation)
	entity.AddComponent(experience)
	entity.AddComponent(inventory)
	entity.AddComponent(position)

	// Make choice (option 0 = Help)
	err := sys.MakeChoice(entity, "test_choice", 0)
	if err != nil {
		t.Fatalf("MakeChoice failed: %v", err)
	}

	// Verify alignment changed
	if reputation.Alignment.LawAxis != 0.05 {
		t.Errorf("Expected LawAxis 0.05, got %.2f", reputation.Alignment.LawAxis)
	}
	if reputation.Alignment.GoodAxis != 0.1 {
		t.Errorf("Expected GoodAxis 0.1, got %.2f", reputation.Alignment.GoodAxis)
	}

	// Verify reputation changed
	if reputation.GetReputation("Villagers") != 20.0 {
		t.Errorf("Expected Villagers reputation 20.0, got %.2f", reputation.GetReputation("Villagers"))
	}

	// Verify XP granted
	if experience.TotalXP != 100 {
		t.Errorf("Expected 100 XP, got %d", experience.TotalXP)
	}

	// Verify gold granted
	if inventory.Gold != 50 {
		t.Errorf("Expected 50 gold, got %d", inventory.Gold)
	}

	// Verify choice recorded in history
	if len(moralChoice.ChoiceHistory) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(moralChoice.ChoiceHistory))
	}

	completed := moralChoice.ChoiceHistory[0]
	if completed.ChoiceID != "test_choice" {
		t.Errorf("Expected ChoiceID 'test_choice', got '%s'", completed.ChoiceID)
	}
	if completed.SelectedOption != 0 {
		t.Errorf("Expected SelectedOption 0, got %d", completed.SelectedOption)
	}
	if completed.OptionLabel != "Help" {
		t.Errorf("Expected OptionLabel 'Help', got '%s'", completed.OptionLabel)
	}

	// Verify choice removed from pending
	if len(moralChoice.PendingChoices) != 0 {
		t.Errorf("Expected 0 pending choices, got %d", len(moralChoice.PendingChoices))
	}

	// Verify deed recorded
	if len(reputation.KarmaDeeds) != 1 {
		t.Errorf("Expected 1 deed, got %d", len(reputation.KarmaDeeds))
	}
}

func TestMoralChoiceSystem_MakeChoice_Errors(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	sys := NewMoralChoiceSystem(world, logger)

	tests := []struct {
		name          string
		setupEntity   func() *Entity
		choiceID      string
		optionIndex   int
		expectedError string
	}{
		{
			name: "No moral choice component",
			setupEntity: func() *Entity {
				return world.CreateEntity()
			},
			choiceID:      "test",
			optionIndex:   0,
			expectedError: "has no moral choice component",
		},
		{
			name: "Choice not found",
			setupEntity: func() *Entity {
				entity := world.CreateEntity()
				moralChoice := NewMoralChoiceComponent()
				entity.AddComponent(moralChoice)
				entity.AddComponent(NewReputationComponent())
				return entity
			},
			choiceID:      "nonexistent",
			optionIndex:   0,
			expectedError: "not found",
		},
		{
			name: "Invalid option index (negative)",
			setupEntity: func() *Entity {
				entity := world.CreateEntity()
				moralChoice := NewMoralChoiceComponent()
				moralChoice.AddChoice(MoralChoice{
					ID:      "test",
					Options: []ChoiceOption{{Label: "A"}, {Label: "B"}},
				})
				entity.AddComponent(moralChoice)
				entity.AddComponent(NewReputationComponent())
				return entity
			},
			choiceID:      "test",
			optionIndex:   -1,
			expectedError: "invalid option index",
		},
		{
			name: "Invalid option index (too high)",
			setupEntity: func() *Entity {
				entity := world.CreateEntity()
				moralChoice := NewMoralChoiceComponent()
				moralChoice.AddChoice(MoralChoice{
					ID:      "test",
					Options: []ChoiceOption{{Label: "A"}, {Label: "B"}},
				})
				entity.AddComponent(moralChoice)
				entity.AddComponent(NewReputationComponent())
				return entity
			},
			choiceID:      "test",
			optionIndex:   5,
			expectedError: "invalid option index",
		},
		{
			name: "No reputation component",
			setupEntity: func() *Entity {
				entity := world.CreateEntity()
				moralChoice := NewMoralChoiceComponent()
				moralChoice.AddChoice(MoralChoice{
					ID:      "test",
					Options: []ChoiceOption{{Label: "A"}},
				})
				entity.AddComponent(moralChoice)
				return entity
			},
			choiceID:      "test",
			optionIndex:   0,
			expectedError: "has no reputation component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := tt.setupEntity()
			err := sys.MakeChoice(entity, tt.choiceID, tt.optionIndex)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			if tt.expectedError != "" && !contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

func TestMoralChoiceSystem_StartRedemption(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	sys := NewMoralChoiceSystem(world, logger)

	entity := world.CreateEntity()
	reputation := NewReputationComponent()
	reputation.SetReputation("Bandits", -60.0)
	entity.AddComponent(reputation)

	actions := []RedemptionAction{
		{
			Type:           "Kill",
			Description:    "Kill rival bandits",
			Target:         "Rival_Bandit",
			Quantity:       10,
			ReputationGain: 5.0,
		},
		{
			Type:           "Deliver",
			Description:    "Deliver stolen goods",
			Target:         "Stolen_Goods",
			Quantity:       5,
			ReputationGain: 10.0,
		},
	}

	err := sys.StartRedemption(entity, "Bandits", -20.0, actions)
	if err != nil {
		t.Fatalf("StartRedemption failed: %v", err)
	}

	// Verify moral choice component was created
	comp, ok := entity.GetComponent("moral_choice")
	if !ok {
		t.Fatal("Moral choice component should be created")
	}

	moralChoice, ok := comp.(*MoralChoiceComponent)
	if !ok {
		t.Fatal("Invalid component type")
	}

	// Verify redemption arc was added
	if len(moralChoice.ActiveRedemptions) != 1 {
		t.Fatalf("Expected 1 redemption, got %d", len(moralChoice.ActiveRedemptions))
	}

	arc := moralChoice.ActiveRedemptions[0]
	if arc.FactionName != "Bandits" {
		t.Errorf("Expected faction 'Bandits', got '%s'", arc.FactionName)
	}
	if arc.StartingReputation != -60.0 {
		t.Errorf("Expected starting reputation -60.0, got %.2f", arc.StartingReputation)
	}
	if arc.TargetReputation != -20.0 {
		t.Errorf("Expected target reputation -20.0, got %.2f", arc.TargetReputation)
	}
	if len(arc.RequiredActions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(arc.RequiredActions))
	}
}

func TestMoralChoiceSystem_StartRedemption_Errors(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	sys := NewMoralChoiceSystem(world, logger)

	t.Run("Already active", func(t *testing.T) {
		entity := world.CreateEntity()
		moralChoice := NewMoralChoiceComponent()
		reputation := NewReputationComponent()

		// Start first redemption
		moralChoice.StartRedemption(RedemptionArc{FactionName: "TestFaction"})

		entity.AddComponent(moralChoice)
		entity.AddComponent(reputation)

		// Try to start second redemption for same faction
		err := sys.StartRedemption(entity, "TestFaction", 10.0, []RedemptionAction{})
		if err == nil {
			t.Fatal("Expected error for duplicate redemption")
		}
		if !contains(err.Error(), "already active") {
			t.Errorf("Expected 'already active' error, got: %v", err)
		}
	})

	t.Run("No reputation component", func(t *testing.T) {
		entity := world.CreateEntity()
		moralChoice := NewMoralChoiceComponent()
		entity.AddComponent(moralChoice)

		err := sys.StartRedemption(entity, "TestFaction", 10.0, []RedemptionAction{})
		if err == nil {
			t.Fatal("Expected error for missing reputation component")
		}
		if !contains(err.Error(), "has no reputation component") {
			t.Errorf("Expected 'has no reputation component' error, got: %v", err)
		}
	})
}

func TestMoralChoiceSystem_UpdateRedemptionProgress(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	sys := NewMoralChoiceSystem(world, logger)

	entity := world.CreateEntity()
	moralChoice := NewMoralChoiceComponent()
	reputation := NewReputationComponent()
	reputation.SetReputation("TestFaction", -50.0)

	arc := RedemptionArc{
		FactionName:        "TestFaction",
		StartingReputation: -50.0,
		TargetReputation:   10.0,
		RequiredActions: []RedemptionAction{
			{
				Type:           "Kill",
				Quantity:       10,
				Progress:       5,
				ReputationGain: 20.0,
			},
		},
		CompletedActions: 0,
	}
	moralChoice.StartRedemption(arc)

	entity.AddComponent(moralChoice)
	entity.AddComponent(reputation)

	// Update progress (complete the action)
	err := sys.UpdateRedemptionProgress(entity, "TestFaction", 0, 5)
	if err != nil {
		t.Fatalf("UpdateRedemptionProgress failed: %v", err)
	}

	// Verify progress updated
	updatedArc := moralChoice.GetRedemptionArc("TestFaction")
	if updatedArc.RequiredActions[0].Progress != 10 {
		t.Errorf("Expected progress 10, got %d", updatedArc.RequiredActions[0].Progress)
	}

	// Verify completed actions incremented
	if updatedArc.CompletedActions != 1 {
		t.Errorf("Expected 1 completed action, got %d", updatedArc.CompletedActions)
	}

	// Verify reputation increased
	if reputation.GetReputation("TestFaction") != -30.0 {
		t.Errorf("Expected reputation -30.0, got %.2f", reputation.GetReputation("TestFaction"))
	}
}

func TestMoralChoiceSystem_UpdateRedemptionProgress_Errors(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	sys := NewMoralChoiceSystem(world, logger)

	tests := []struct {
		name          string
		setupEntity   func() *Entity
		faction       string
		actionIndex   int
		expectedError string
	}{
		{
			name: "No moral choice component",
			setupEntity: func() *Entity {
				return world.CreateEntity()
			},
			faction:       "Test",
			actionIndex:   0,
			expectedError: "has no moral choice component",
		},
		{
			name: "No redemption arc",
			setupEntity: func() *Entity {
				entity := world.CreateEntity()
				entity.AddComponent(NewMoralChoiceComponent())
				return entity
			},
			faction:       "NonexistentFaction",
			actionIndex:   0,
			expectedError: "no redemption arc found",
		},
		{
			name: "Invalid action index",
			setupEntity: func() *Entity {
				entity := world.CreateEntity()
				moralChoice := NewMoralChoiceComponent()
				moralChoice.StartRedemption(RedemptionArc{
					FactionName:     "Test",
					RequiredActions: []RedemptionAction{{Quantity: 5}},
				})
				entity.AddComponent(moralChoice)
				return entity
			},
			faction:       "Test",
			actionIndex:   5,
			expectedError: "invalid action index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := tt.setupEntity()
			err := sys.UpdateRedemptionProgress(entity, tt.faction, tt.actionIndex, 1)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			if !contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

func TestMoralChoiceSystem_OfferFactionConflictChoice(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	sys := NewMoralChoiceSystem(world, logger)

	entity := world.CreateEntity()

	err := sys.OfferFactionConflictChoice(entity, "Merchants", "Thieves", "The merchants accuse the thieves of stealing")
	if err != nil {
		t.Fatalf("OfferFactionConflictChoice failed: %v", err)
	}

	// Verify moral choice component was created
	comp, ok := entity.GetComponent("moral_choice")
	if !ok {
		t.Fatal("Moral choice component should be created")
	}

	moralChoice, ok := comp.(*MoralChoiceComponent)
	if !ok {
		t.Fatal("Invalid component type")
	}

	// Verify choice was added
	if len(moralChoice.PendingChoices) != 1 {
		t.Fatalf("Expected 1 pending choice, got %d", len(moralChoice.PendingChoices))
	}

	choice := moralChoice.PendingChoices[0]

	// Verify 3 options (support faction1, support faction2, stay neutral)
	if len(choice.Options) != 3 {
		t.Fatalf("Expected 3 options, got %d", len(choice.Options))
	}

	// Verify option 0 supports Merchants
	opt0 := choice.Options[0]
	if opt0.ReputationImpact["Merchants"] != 20.0 {
		t.Errorf("Expected Merchants +20, got %.2f", opt0.ReputationImpact["Merchants"])
	}
	if opt0.ReputationImpact["Thieves"] != -30.0 {
		t.Errorf("Expected Thieves -30, got %.2f", opt0.ReputationImpact["Thieves"])
	}

	// Verify option 1 supports Thieves
	opt1 := choice.Options[1]
	if opt1.ReputationImpact["Merchants"] != -30.0 {
		t.Errorf("Expected Merchants -30, got %.2f", opt1.ReputationImpact["Merchants"])
	}
	if opt1.ReputationImpact["Thieves"] != 20.0 {
		t.Errorf("Expected Thieves +20, got %.2f", opt1.ReputationImpact["Thieves"])
	}

	// Verify option 2 is neutral
	opt2 := choice.Options[2]
	if opt2.Label != "Stay neutral" {
		t.Errorf("Expected 'Stay neutral', got '%s'", opt2.Label)
	}
	if opt2.ReputationImpact["Merchants"] != -5.0 {
		t.Errorf("Expected Merchants -5, got %.2f", opt2.ReputationImpact["Merchants"])
	}
	if opt2.ReputationImpact["Thieves"] != -5.0 {
		t.Errorf("Expected Thieves -5, got %.2f", opt2.ReputationImpact["Thieves"])
	}
}

func TestMoralChoiceSystem_ApplyConsequences(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	sys := NewMoralChoiceSystem(world, logger)

	entity := world.CreateEntity()
	reputation := NewReputationComponent()
	inventory := NewInventoryComponent(20, 100.0)
	inventory.Items = []*item.Item{
		{ID: "item1", Name: "Keep This"},
		{ID: "item2", Name: "Lose This"},
		{ID: "item3", Name: "Keep This Too"},
	}
	position := &PositionComponent{X: 50, Y: 50}

	entity.AddComponent(reputation)
	entity.AddComponent(inventory)
	entity.AddComponent(position)

	consequences := &ChoiceConsequences{
		HostileFactions: []string{"Guards", "Merchants"},
		LoseItems:       []string{"item2"},
		LoseQuests:      []string{"quest1"},
		SpawnEnemies:    5,
	}

	err := sys.applyConsequences(entity, consequences)
	if err != nil {
		t.Fatalf("applyConsequences failed: %v", err)
	}

	// Verify factions are hostile
	if reputation.GetReputation("Guards") != -50.0 {
		t.Errorf("Expected Guards reputation -50.0, got %.2f", reputation.GetReputation("Guards"))
	}
	if reputation.GetReputation("Merchants") != -50.0 {
		t.Errorf("Expected Merchants reputation -50.0, got %.2f", reputation.GetReputation("Merchants"))
	}

	// Verify item was removed
	if len(inventory.Items) != 2 {
		t.Errorf("Expected 2 items remaining, got %d", len(inventory.Items))
	}
	for _, item := range inventory.Items {
		if item.ID == "item2" {
			t.Error("item2 should have been removed")
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
