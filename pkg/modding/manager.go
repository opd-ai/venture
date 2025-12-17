package modding

import (
	"fmt"
	"sync"
	"time"
)

// handlerWithOwner associates an event handler with the mod that registered it.
// This enables proper cleanup when a mod is removed.
type handlerWithOwner struct {
	modID   string
	handler EventHandler
}

// Manager manages loaded mods and applies their rules to the game.
type Manager struct {
	mods           map[string]*Mod
	activeRules    map[string]interface{}
	eventHandlers  map[string][]handlerWithOwner
	mu             sync.RWMutex
	config         ModConfig
	ruleChangeLog  []RuleContext
	lastRuleChange time.Time
	changeCount    int
}

// NewManager creates a new mod manager.
func NewManager() *Manager {
	return &Manager{
		mods:          make(map[string]*Mod),
		activeRules:   make(map[string]interface{}),
		eventHandlers: make(map[string][]handlerWithOwner),
		config:        DefaultConfig(),
		ruleChangeLog: make([]RuleContext, 0, 1000),
	}
}

// NewManagerWithConfig creates a new mod manager with custom configuration.
func NewManagerWithConfig(config ModConfig) *Manager {
	return &Manager{
		mods:          make(map[string]*Mod),
		activeRules:   make(map[string]interface{}),
		eventHandlers: make(map[string][]handlerWithOwner),
		config:        config,
		ruleChangeLog: make([]RuleContext, 0, 1000),
	}
}

// AddMod adds a mod to the manager.
func (m *Manager) AddMod(mod *Mod) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate mod
	if err := mod.Validate(); err != nil {
		return err
	}

	// Check if mod already exists
	if _, exists := m.mods[mod.ID]; exists {
		return fmt.Errorf("mod %s already loaded", mod.ID)
	}

	// Check max mods limit
	if len(m.mods) >= m.config.MaxMods {
		return fmt.Errorf("maximum number of mods (%d) reached", m.config.MaxMods)
	}

	// Check dependencies
	for _, depID := range mod.Dependencies {
		if _, exists := m.mods[depID]; !exists {
			return fmt.Errorf("dependency %s not loaded", depID)
		}
	}

	// Store mod
	m.mods[mod.ID] = mod

	// Register event handlers with ownership tracking
	if len(mod.EventHandlers) > 0 {
		for eventType, handler := range mod.EventHandlers {
			m.eventHandlers[eventType] = append(m.eventHandlers[eventType], handlerWithOwner{
				modID:   mod.ID,
				handler: handler,
			})
		}
	}

	return nil
}

// RemoveMod removes a mod from the manager.
func (m *Manager) RemoveMod(modID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if mod exists
	mod, exists := m.mods[modID]
	if !exists {
		return fmt.Errorf("mod %s not found", modID)
	}

	// Check for dependent mods
	for _, otherMod := range m.mods {
		for _, depID := range otherMod.Dependencies {
			if depID == modID {
				return fmt.Errorf("cannot remove %s: mod %s depends on it", modID, otherMod.ID)
			}
		}
	}

	// Remove event handlers belonging to this mod only
	for eventType := range mod.EventHandlers {
		handlers := m.eventHandlers[eventType]
		filtered := handlers[:0]
		for _, h := range handlers {
			if h.modID != modID {
				filtered = append(filtered, h)
			}
		}
		if len(filtered) > 0 {
			m.eventHandlers[eventType] = filtered
		} else {
			delete(m.eventHandlers, eventType)
		}
	}

	// Remove mod
	delete(m.mods, modID)

	return nil
}

// GetMod retrieves a mod by ID.
func (m *Manager) GetMod(modID string) (*Mod, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mod, exists := m.mods[modID]
	return mod, exists
}

// ListMods returns all loaded mods.
func (m *Manager) ListMods() []*Mod {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mods := make([]*Mod, 0, len(m.mods))
	for _, mod := range m.mods {
		mods = append(mods, mod)
	}

	return mods
}

// EnableMod enables a mod.
func (m *Manager) EnableMod(modID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mod, exists := m.mods[modID]
	if !exists {
		return fmt.Errorf("mod %s not found", modID)
	}

	mod.Enabled = true
	return nil
}

// DisableMod disables a mod.
func (m *Manager) DisableMod(modID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mod, exists := m.mods[modID]
	if !exists {
		return fmt.Errorf("mod %s not found", modID)
	}

	mod.Enabled = false
	return nil
}

// ApplyRules applies all enabled mod rules.
func (m *Manager) ApplyRules() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check rate limit
	if err := m.checkRateLimit(); err != nil {
		return err
	}

	// Clear active rules
	m.activeRules = make(map[string]interface{})

	// Apply rules from enabled mods
	for _, mod := range m.mods {
		if !mod.Enabled {
			continue
		}

		for ruleName, value := range mod.Rules {
			// Log rule change
			ctx := RuleContext{
				ModID:     mod.ID,
				RuleName:  ruleName,
				OldValue:  m.activeRules[ruleName],
				NewValue:  value,
				AppliedAt: time.Now(),
			}
			m.ruleChangeLog = append(m.ruleChangeLog, ctx)

			// Apply rule (later mods override earlier ones)
			m.activeRules[ruleName] = value
		}
	}

	return nil
}

// GetRule retrieves the current value of a rule.
func (m *Manager) GetRule(ruleName string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, exists := m.activeRules[ruleName]
	return value, exists
}

// GetRuleFloat64 retrieves a rule as float64 with a default value.
func (m *Manager) GetRuleFloat64(ruleName string, defaultValue float64) float64 {
	value, exists := m.GetRule(ruleName)
	if !exists {
		return defaultValue
	}

	// Try to convert to float64
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return defaultValue
	}
}

// GetRuleBool retrieves a rule as bool with a default value.
func (m *Manager) GetRuleBool(ruleName string, defaultValue bool) bool {
	value, exists := m.GetRule(ruleName)
	if !exists {
		return defaultValue
	}

	if boolVal, ok := value.(bool); ok {
		return boolVal
	}

	return defaultValue
}

// TriggerEvent triggers an event and calls registered handlers.
func (m *Manager) TriggerEvent(event Event) error {
	m.mu.RLock()
	handlers, exists := m.eventHandlers[event.Type]
	m.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		return nil // No handlers for this event type
	}

	// Call handlers (in order of registration)
	for _, h := range handlers {
		if err := h.handler(event); err != nil {
			return fmt.Errorf("event handler failed for %s: %w", event.Type, err)
		}
	}

	return nil
}

// GetRuleChangeLog returns the rule change history.
func (m *Manager) GetRuleChangeLog() []RuleContext {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent modification
	log := make([]RuleContext, len(m.ruleChangeLog))
	copy(log, m.ruleChangeLog)

	return log
}

// checkRateLimit checks if the rate limit for rule changes has been exceeded.
func (m *Manager) checkRateLimit() error {
	now := time.Now()

	// Reset counter if more than 1 second has passed
	if now.Sub(m.lastRuleChange) > time.Second {
		m.changeCount = 0
		m.lastRuleChange = now
	}

	m.changeCount++

	// Check if rate limit exceeded
	if float64(m.changeCount) > m.config.RuleChangeRateLimit {
		return fmt.Errorf("rate limit exceeded: %d changes per second (max %.0f)",
			m.changeCount, m.config.RuleChangeRateLimit)
	}

	return nil
}

// GetStats returns statistics about the mod manager.
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	enabledCount := 0
	for _, mod := range m.mods {
		if mod.Enabled {
			enabledCount++
		}
	}

	return map[string]interface{}{
		"total_mods":     len(m.mods),
		"enabled_mods":   enabledCount,
		"active_rules":   len(m.activeRules),
		"event_handlers": len(m.eventHandlers),
		"rule_changes":   len(m.ruleChangeLog),
	}
}
