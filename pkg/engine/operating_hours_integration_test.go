package engine

import (
	"testing"
	"time"
)

// TestOperatingHoursIntegration tests the full integration of operating hours
// with dialog and commerce systems.
func TestOperatingHoursIntegration(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC)) // Monday 10 AM

	availSystem := NewAvailabilitySystem(world, clock)

	// Create a merchant with operating hours
	merchant := world.CreateEntity()
	merchant.AddComponent(&PositionComponent{X: 100, Y: 100})
	merchant.AddComponent(NewMerchantComponent(20, MerchantFixed, 1.5))

	// Add operating hours: open 8 AM - 6 PM, Mon-Sat
	hours := NewOperatingHoursComponent()
	hours.ServiceType = "shop"
	hours.ClosedMessage = "The shop is closed. Come back during business hours."
	merchant.AddComponent(hours)

	// Add dialog with shop option
	dialog := &DialogComponent{
		Options: []DialogOption{
			{Text: "Browse your wares", Action: ActionOpenShop, Enabled: true},
			{Text: "Goodbye", Action: ActionCloseDialog, Enabled: true},
		},
	}
	merchant.AddComponent(dialog)

	// Flush entities
	world.Update(0)

	// At 10 AM Monday, shop should be open
	t.Run("shop open during business hours", func(t *testing.T) {
		if !availSystem.IsEntityOpen(merchant) {
			t.Error("Merchant should be open at 10 AM Monday")
		}

		allowed, msg := availSystem.CheckInteractionAllowed(merchant.ID)
		if !allowed {
			t.Errorf("Interaction should be allowed, got: %s", msg)
		}

		// Update should keep shop option enabled
		availSystem.Update([]*Entity{merchant}, 0.016)
		if !dialog.Options[0].Enabled {
			t.Error("Shop option should be enabled during business hours")
		}
		if dialog.Options[0].Text != "Browse your wares" {
			t.Errorf("Shop option text = %q, want %q", dialog.Options[0].Text, "Browse your wares")
		}
	})

	// At 8 PM Monday, shop should be closed
	t.Run("shop closed after hours", func(t *testing.T) {
		clock.Reset(time.Date(2025, 1, 6, 20, 0, 0, 0, time.UTC)) // 8 PM

		if availSystem.IsEntityOpen(merchant) {
			t.Error("Merchant should be closed at 8 PM")
		}

		allowed, msg := availSystem.CheckInteractionAllowed(merchant.ID)
		if allowed {
			t.Error("Interaction should not be allowed after hours")
		}
		if msg == "" {
			t.Error("Should provide closed message")
		}

		// Update should disable shop option
		availSystem.Update([]*Entity{merchant}, 0.016)
		if dialog.Options[0].Enabled {
			t.Error("Shop option should be disabled after hours")
		}
		if dialog.Options[0].Text != "Shop is closed" {
			t.Errorf("Shop option text = %q, want %q", dialog.Options[0].Text, "Shop is closed")
		}
	})

	// On Sunday, shop should be closed
	t.Run("shop closed on Sunday", func(t *testing.T) {
		clock.Reset(time.Date(2025, 1, 5, 12, 0, 0, 0, time.UTC)) // Sunday noon

		if availSystem.IsEntityOpen(merchant) {
			t.Error("Merchant should be closed on Sunday")
		}

		status := availSystem.GetEntityStatus(merchant)
		if status != "The shop is closed. Come back during business hours." {
			t.Errorf("Status = %q, want closed message", status)
		}
	})

	// Next open time should be Monday
	t.Run("next open time", func(t *testing.T) {
		clock.Reset(time.Date(2025, 1, 5, 12, 0, 0, 0, time.UTC)) // Sunday noon

		hour, day, hasHours := availSystem.GetNextOpenTime(merchant)
		if !hasHours {
			t.Error("Should have operating hours")
		}
		if hour != 8 {
			t.Errorf("Next open hour = %d, want 8", hour)
		}
		if day != Monday {
			t.Errorf("Next open day = %s, want Monday", day.String())
		}
	})
}

// TestAlwaysOpenServices tests services that operate 24/7.
func TestAlwaysOpenServices(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 5, 3, 0, 0, 0, time.UTC)) // Sunday 3 AM
	availSystem := NewAvailabilitySystem(world, clock)

	// Create an inn (always open)
	inn := world.CreateEntity()
	inn.AddComponent(&PositionComponent{X: 200, Y: 200})
	inn.AddComponent(NewAlwaysOpenComponent("inn"))

	// Flush entities
	world.Update(0)

	// Should be open at any time
	if !availSystem.IsEntityOpen(inn) {
		t.Error("Inn should be open at 3 AM Sunday")
	}

	allowed, _ := availSystem.CheckInteractionAllowed(inn.ID)
	if !allowed {
		t.Error("Interaction with inn should always be allowed")
	}

	status := availSystem.GetEntityStatus(inn)
	if status != "Open" {
		t.Errorf("Inn status = %q, want %q", status, "Open")
	}
}

// TestOperatingHoursDeterminism verifies schedule behavior is deterministic.
func TestOperatingHoursDeterminism(t *testing.T) {
	// Same setup should produce identical results
	for i := 0; i < 3; i++ {
		world := NewWorld()
		clock := NewSimulationClock(12345)
		clock.Reset(time.Date(2025, 1, 6, 12, 0, 0, 0, time.UTC))
		availSystem := NewAvailabilitySystem(world, clock)

		shop := world.CreateEntity()
		hours := NewOperatingHoursComponent()
		hours.SetHours(8, 18)
		shop.AddComponent(hours)

		world.Update(0)

		isOpen := availSystem.IsEntityOpen(shop)
		if !isOpen {
			t.Errorf("Run %d: Shop should be open at noon on Monday", i)
		}
	}
}
