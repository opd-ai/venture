package modding

import (
	"testing"
)

func TestProviderAdapter_GetRule(t *testing.T) {
	manager := NewManager()
	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules:   map[string]interface{}{"difficulty": 2.0},
	}
	if err := manager.AddMod(mod); err != nil {
		t.Fatalf("AddMod failed: %v", err)
	}
	if err := manager.EnableMod("test-mod"); err != nil {
		t.Fatalf("EnableMod failed: %v", err)
	}
	if err := manager.ApplyRules(); err != nil {
		t.Fatalf("ApplyRules failed: %v", err)
	}

	adapter := NewProviderAdapter(manager)

	// Test GetRule
	val, exists := adapter.GetRule("difficulty")
	if !exists {
		t.Error("GetRule should find difficulty")
	}
	if val != 2.0 {
		t.Errorf("GetRule(difficulty) = %v, want 2.0", val)
	}

	// Test GetRule for non-existent
	_, exists = adapter.GetRule("nonexistent")
	if exists {
		t.Error("GetRule should not find nonexistent")
	}
}

func TestProviderAdapter_GetRuleFloat64(t *testing.T) {
	manager := NewManager()
	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules:   map[string]interface{}{"damage_multiplier": 1.5},
	}
	if err := manager.AddMod(mod); err != nil {
		t.Fatalf("AddMod failed: %v", err)
	}
	if err := manager.EnableMod("test-mod"); err != nil {
		t.Fatalf("EnableMod failed: %v", err)
	}
	if err := manager.ApplyRules(); err != nil {
		t.Fatalf("ApplyRules failed: %v", err)
	}

	adapter := NewProviderAdapter(manager)

	// Test existing rule
	val := adapter.GetRuleFloat64("damage_multiplier", 1.0)
	if val != 1.5 {
		t.Errorf("GetRuleFloat64(damage_multiplier) = %v, want 1.5", val)
	}

	// Test non-existent rule with default
	val = adapter.GetRuleFloat64("nonexistent", 1.0)
	if val != 1.0 {
		t.Errorf("GetRuleFloat64(nonexistent) = %v, want 1.0", val)
	}
}

func TestProviderAdapter_GetRuleBool(t *testing.T) {
	manager := NewManager()
	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules:   map[string]interface{}{"permadeath": true},
	}
	if err := manager.AddMod(mod); err != nil {
		t.Fatalf("AddMod failed: %v", err)
	}
	if err := manager.EnableMod("test-mod"); err != nil {
		t.Fatalf("EnableMod failed: %v", err)
	}
	if err := manager.ApplyRules(); err != nil {
		t.Fatalf("ApplyRules failed: %v", err)
	}

	adapter := NewProviderAdapter(manager)

	// Test existing rule
	val := adapter.GetRuleBool("permadeath", false)
	if !val {
		t.Errorf("GetRuleBool(permadeath) = %v, want true", val)
	}

	// Test non-existent rule with default
	val = adapter.GetRuleBool("nonexistent", false)
	if val {
		t.Errorf("GetRuleBool(nonexistent) = %v, want false", val)
	}
}

func TestProviderAdapter_TriggerEvent(t *testing.T) {
	manager := NewManager()

	eventCalled := false
	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeEvent,
		EventHandlers: map[string]EventHandler{
			"test_event": func(e Event) error {
				eventCalled = true
				if e.Type != "test_event" {
					t.Errorf("Event type = %s, want test_event", e.Type)
				}
				if e.Data["key"] != "value" {
					t.Errorf("Event data key = %v, want value", e.Data["key"])
				}
				return nil
			},
		},
	}
	if err := manager.AddMod(mod); err != nil {
		t.Fatalf("AddMod failed: %v", err)
	}
	if err := manager.EnableMod("test-mod"); err != nil {
		t.Fatalf("EnableMod failed: %v", err)
	}

	adapter := NewProviderAdapter(manager)

	// Trigger event
	err := adapter.TriggerEvent("test_event", map[string]interface{}{"key": "value"})
	if err != nil {
		t.Errorf("TriggerEvent failed: %v", err)
	}

	if !eventCalled {
		t.Error("Event handler was not called")
	}
}

func TestProviderAdapter_NilManager(t *testing.T) {
	adapter := NewProviderAdapter(nil)

	// These should panic with nil pointer - this tests that users should
	// only create adapters with valid managers
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil manager")
		}
	}()

	adapter.GetRule("test")
}

func TestProviderAdapter_ImplementsInterface(t *testing.T) {
	// This test verifies at compile time that ProviderAdapter implements
	// an interface equivalent to engine.ModRuleProvider
	type ModRuleProvider interface {
		GetRule(ruleName string) (interface{}, bool)
		GetRuleFloat64(ruleName string, defaultValue float64) float64
		GetRuleBool(ruleName string, defaultValue bool) bool
		TriggerEvent(eventType string, eventData map[string]interface{}) error
	}

	var _ ModRuleProvider = (*ProviderAdapter)(nil)
}
