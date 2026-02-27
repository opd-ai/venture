// Package modding provides mod system adapter methods for ECS integration.
// This file contains methods that allow modding.Manager to implement
// the engine.ModRuleProvider interface without creating circular imports.
package modding

import (
	"time"
)

// ProviderAdapter wraps a Manager to implement engine.ModRuleProvider interface.
// This is used to bridge the modding package with the engine package without
// requiring the engine package to import modding (avoiding circular imports).
//
// Usage:
//
//	manager := modding.NewManager()
//	adapter := modding.NewProviderAdapter(manager)
//	world.SetModRules(adapter)
type ProviderAdapter struct {
	manager *Manager
}

// NewProviderAdapter creates a new adapter that wraps a Manager.
func NewProviderAdapter(manager *Manager) *ProviderAdapter {
	return &ProviderAdapter{manager: manager}
}

// GetRule retrieves the current value of a rule.
func (p *ProviderAdapter) GetRule(ruleName string) (interface{}, bool) {
	return p.manager.GetRule(ruleName)
}

// GetRuleFloat64 retrieves a rule as float64 with a default value.
func (p *ProviderAdapter) GetRuleFloat64(ruleName string, defaultValue float64) float64 {
	return p.manager.GetRuleFloat64(ruleName, defaultValue)
}

// GetRuleBool retrieves a rule as bool with a default value.
func (p *ProviderAdapter) GetRuleBool(ruleName string, defaultValue bool) bool {
	return p.manager.GetRuleBool(ruleName, defaultValue)
}

// TriggerEvent triggers a mod event. This converts the simple parameters
// to the internal modding.Event type.
// Phase 6.3 (PLAN.md): Modding System Integration
func (p *ProviderAdapter) TriggerEvent(eventType string, eventData map[string]interface{}) error {
	event := Event{
		Type: eventType,
		Data: eventData,
		// INTENTIONAL time.Now() EXCEPTION: Event timestamp for audit trail only.
		// Does not affect procedural content generation. See doc.go:113-120.
		Timestamp: time.Now(),
	}
	return p.manager.TriggerEvent(event)
}
