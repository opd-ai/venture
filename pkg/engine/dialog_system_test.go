package engine

import (
	"testing"
)

func TestNewDialogSystem(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	if system == nil {
		t.Fatal("NewDialogSystem returned nil")
	}
	if system.world != world {
		t.Error("DialogSystem world not set correctly")
	}
	if system.activeDialogEntity != 0 {
		t.Errorf("activeDialogEntity = %d, want 0", system.activeDialogEntity)
	}
}

func TestDialogSystem_StartDialog(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	// Create player entity
	player := world.CreateEntity()
	playerID := player.ID

	// Create NPC entity with dialog component
	npc := world.CreateEntity()
	npcID := npc.ID
	provider := NewMerchantDialogProvider("Test Merchant")
	dialogComp := NewDialogComponent(provider)
	npc.AddComponent(dialogComp)

	// Process pending entities
	world.Update(0)

	tests := []struct {
		name     string
		playerID uint64
		npcID    uint64
		wantOk   bool
		wantErr  bool
	}{
		{"valid dialog start", playerID, npcID, true, false},
		{"nonexistent npc", playerID, 9999, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset system state
			system.activeDialogEntity = 0

			ok, err := system.StartDialog(tt.playerID, tt.npcID)

			if ok != tt.wantOk {
				t.Errorf("StartDialog() ok = %v, want %v", ok, tt.wantOk)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("StartDialog() error = %v, wantErr %v", err, tt.wantErr)
			}

			if ok && system.activeDialogEntity != tt.npcID {
				t.Errorf("activeDialogEntity = %d, want %d", system.activeDialogEntity, tt.npcID)
			}
		})
	}
}

func TestDialogSystem_StartDialog_AlreadyActive(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	player := world.CreateEntity()
	playerID := player.ID
	npc1 := world.CreateEntity()
	npc1ID := npc1.ID
	npc2 := world.CreateEntity()
	npc2ID := npc2.ID

	provider := NewMerchantDialogProvider("Merchant 1")
	npc1.AddComponent(NewDialogComponent(provider))
	npc2.AddComponent(NewDialogComponent(provider))

	// Process pending entities
	world.Update(0)

	// Start dialog with first NPC
	ok, err := system.StartDialog(playerID, npc1ID)
	if !ok || err != nil {
		t.Fatalf("Failed to start first dialog: %v", err)
	}

	// Try to start dialog with second NPC (should fail)
	ok, err = system.StartDialog(playerID, npc2ID)
	if ok {
		t.Error("StartDialog() succeeded when dialog already active, want failure")
	}
	if err == nil {
		t.Error("StartDialog() error = nil, want error when dialog already active")
	}
	if system.activeDialogEntity != npc1ID {
		t.Errorf("activeDialogEntity changed to %d, want %d", system.activeDialogEntity, npc1ID)
	}
}

func TestDialogSystem_StartDialog_NoDialogComponent(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	player := world.CreateEntity()
	playerID := player.ID
	npc := world.CreateEntity()
	npcID := npc.ID
	// Don't add dialog component

	// Process pending entities
	world.Update(0)

	ok, err := system.StartDialog(playerID, npcID)
	if ok {
		t.Error("StartDialog() ok = true, want false for entity without dialog component")
	}
	if err == nil {
		t.Error("StartDialog() error = nil, want error for entity without dialog component")
	}
}

func TestDialogSystem_EndDialog(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	player := world.CreateEntity()
	playerID := player.ID
	npc := world.CreateEntity()
	npcID := npc.ID
	provider := NewMerchantDialogProvider("Test Merchant")
	dialogComp := NewDialogComponent(provider)
	npc.AddComponent(dialogComp)

	// Process pending entities
	world.Update(0)

	t.Run("end when no dialog active", func(t *testing.T) {
		err := system.EndDialog()
		if err == nil {
			t.Error("EndDialog() error = nil, want error when no dialog active")
		}
	})

	t.Run("end active dialog", func(t *testing.T) {
		// Start dialog
		system.StartDialog(playerID, npcID)

		err := system.EndDialog()
		if err != nil {
			t.Errorf("EndDialog() error = %v, want nil", err)
		}
		if system.activeDialogEntity != 0 {
			t.Errorf("activeDialogEntity = %d, want 0 after EndDialog", system.activeDialogEntity)
		}
		if dialogComp.IsActive {
			t.Error("DialogComponent.IsActive = true, want false after EndDialog")
		}
	})

	t.Run("end dialog with destroyed entity", func(t *testing.T) {
		// Start dialog
		system.StartDialog(playerID, npcID)

		// Remove NPC entity
		world.RemoveEntity(npcID)
		world.Update(0) // Process removal

		err := system.EndDialog()
		if err != nil {
			t.Errorf("EndDialog() error = %v, want nil when entity destroyed", err)
		}
		if system.activeDialogEntity != 0 {
			t.Errorf("activeDialogEntity = %d, want 0 after EndDialog", system.activeDialogEntity)
		}
	})
}

func TestDialogSystem_SelectDialogOption(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	player := world.CreateEntity()
	playerID := player.ID
	npc := world.CreateEntity()
	npcID := npc.ID
	provider := NewMerchantDialogProvider("Test Merchant")
	dialogComp := NewDialogComponent(provider)
	npc.AddComponent(dialogComp)

	// Process pending entities
	world.Update(0)

	t.Run("select when no dialog active", func(t *testing.T) {
		_, err := system.SelectDialogOption(0)
		if err == nil {
			t.Error("SelectDialogOption() error = nil, want error when no dialog active")
		}
	})

	// Start dialog for remaining tests
	system.StartDialog(playerID, npcID)

	t.Run("select valid option", func(t *testing.T) {
		action, err := system.SelectDialogOption(0)
		if err != nil {
			t.Errorf("SelectDialogOption(0) error = %v, want nil", err)
		}
		if action == ActionNone {
			t.Error("SelectDialogOption(0) returned ActionNone, want valid action")
		}
	})

	t.Run("select invalid index negative", func(t *testing.T) {
		_, err := system.SelectDialogOption(-1)
		if err == nil {
			t.Error("SelectDialogOption(-1) error = nil, want error")
		}
	})

	t.Run("select invalid index too high", func(t *testing.T) {
		_, err := system.SelectDialogOption(999)
		if err == nil {
			t.Error("SelectDialogOption(999) error = nil, want error")
		}
	})
}

func TestDialogSystem_SelectDialogOption_DisabledOption(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	player := world.CreateEntity()
	playerID := player.ID
	npc := world.CreateEntity()
	npcID := npc.ID

	// Create custom dialog with disabled option
	dialogComp := NewDialogComponent(nil)
	dialogComp.CurrentDialog = "Test"
	dialogComp.Options = []DialogOption{
		{Text: "Option 1", Action: ActionOpenShop, Enabled: true},
		{Text: "Option 2", Action: ActionNone, Enabled: false},
	}
	dialogComp.IsActive = true
	npc.AddComponent(dialogComp)

	// Process pending entities
	world.Update(0)

	system.StartDialog(playerID, npcID)

	// Try to select disabled option
	_, err := system.SelectDialogOption(1)
	if err == nil {
		t.Error("SelectDialogOption() error = nil, want error for disabled option")
	}
}

func TestDialogSystem_SelectDialogOption_CloseAction(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	player := world.CreateEntity()
	playerID := player.ID
	npc := world.CreateEntity()
	npcID := npc.ID
	provider := NewMerchantDialogProvider("Test Merchant")
	dialogComp := NewDialogComponent(provider)
	npc.AddComponent(dialogComp)

	// Process pending entities
	world.Update(0)

	system.StartDialog(playerID, npcID)

	// Select "Never mind" option (index 1, which is ActionCloseDialog)
	action, err := system.SelectDialogOption(1)
	if err != nil {
		t.Errorf("SelectDialogOption(1) error = %v, want nil", err)
	}
	if action != ActionCloseDialog {
		t.Errorf("SelectDialogOption(1) action = %v, want ActionCloseDialog", action)
	}

	// Dialog should be automatically closed
	if system.activeDialogEntity != 0 {
		t.Errorf("activeDialogEntity = %d, want 0 after ActionCloseDialog", system.activeDialogEntity)
	}
}

func TestDialogSystem_GetActiveDialog(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	t.Run("no active dialog", func(t *testing.T) {
		text, options, npcID := system.GetActiveDialog()
		if text != "" {
			t.Errorf("text = %q, want empty string when no dialog active", text)
		}
		if options != nil {
			t.Errorf("options = %v, want nil when no dialog active", options)
		}
		if npcID != 0 {
			t.Errorf("npcID = %d, want 0 when no dialog active", npcID)
		}
	})

	player := world.CreateEntity()
	playerID := player.ID
	npc := world.CreateEntity()
	npcID := npc.ID
	provider := NewMerchantDialogProvider("Test Merchant")
	dialogComp := NewDialogComponent(provider)
	npc.AddComponent(dialogComp)

	// Process pending entities
	world.Update(0)

	t.Run("active dialog", func(t *testing.T) {
		system.StartDialog(playerID, npcID)

		text, options, returnedNpcID := system.GetActiveDialog()
		if text == "" {
			t.Error("text is empty, want dialog text")
		}
		if options == nil || len(options) == 0 {
			t.Error("options is nil or empty, want dialog options")
		}
		if returnedNpcID != npcID {
			t.Errorf("npcID = %d, want %d", returnedNpcID, npcID)
		}
	})

	t.Run("dialog ended", func(t *testing.T) {
		system.EndDialog()

		text, options, returnedNpcID := system.GetActiveDialog()
		if text != "" {
			t.Errorf("text = %q, want empty string after dialog ended", text)
		}
		if options != nil {
			t.Errorf("options = %v, want nil after dialog ended", options)
		}
		if returnedNpcID != 0 {
			t.Errorf("npcID = %d, want 0 after dialog ended", returnedNpcID)
		}
	})
}

func TestDialogSystem_IsDialogActive(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	if system.IsDialogActive() {
		t.Error("IsDialogActive() = true, want false initially")
	}

	player := world.CreateEntity()
	playerID := player.ID
	npc := world.CreateEntity()
	npcID := npc.ID
	provider := NewMerchantDialogProvider("Test Merchant")
	npc.AddComponent(NewDialogComponent(provider))

	// Process pending entities
	world.Update(0)

	system.StartDialog(playerID, npcID)
	if !system.IsDialogActive() {
		t.Error("IsDialogActive() = false, want true after StartDialog")
	}

	system.EndDialog()
	if system.IsDialogActive() {
		t.Error("IsDialogActive() = true, want false after EndDialog")
	}
}

func TestDialogSystem_GetActiveDialogEntity(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	if system.GetActiveDialogEntity() != 0 {
		t.Errorf("GetActiveDialogEntity() = %d, want 0 initially", system.GetActiveDialogEntity())
	}

	player := world.CreateEntity()
	playerID := player.ID
	npc := world.CreateEntity()
	npcID := npc.ID
	provider := NewMerchantDialogProvider("Test Merchant")
	npc.AddComponent(NewDialogComponent(provider))

	// Process pending entities
	world.Update(0)

	system.StartDialog(playerID, npcID)
	if system.GetActiveDialogEntity() != npcID {
		t.Errorf("GetActiveDialogEntity() = %d, want %d", system.GetActiveDialogEntity(), npcID)
	}

	system.EndDialog()
	if system.GetActiveDialogEntity() != 0 {
		t.Errorf("GetActiveDialogEntity() = %d, want 0 after EndDialog", system.GetActiveDialogEntity())
	}
}

func TestDialogSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewDialogSystem(world)

	// Update should not panic (currently no-op, but reserved for future)
	system.Update(nil, 0.016)
	system.Update(nil, 1.0)
	system.Update(nil, 0.0)
}
