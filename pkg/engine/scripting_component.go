package engine

import (
	"encoding/json"
	"fmt"
	"sync"
)

// ScriptingComponent tracks script execution state for mod scripting.
type ScriptingComponent struct {
	Scripts   map[string]*Script // scriptID -> script
	Variables map[string]any     // shared script variables
	LastError string             // last execution error
	Enabled   bool               // scripting active

	mu sync.RWMutex
}

// Script represents an individual mod script.
type Script struct {
	ID           string         // unique script identifier
	ModID        string         // parent mod identifier
	Source       string         // script source code
	TriggerEvent string         // event that runs this script
	Priority     int            // execution priority (lower = earlier)
	LastRun      int64          // unix timestamp of last execution
	RunCount     int            // total execution count
	Enabled      bool           // script is active
	Metadata     map[string]any // script metadata
}

// NewScriptingComponent creates a new scripting component.
func NewScriptingComponent() *ScriptingComponent {
	return &ScriptingComponent{
		Scripts:   make(map[string]*Script),
		Variables: make(map[string]any),
		Enabled:   true,
	}
}

// Type returns the component type identifier.
func (s *ScriptingComponent) Type() string {
	return "scripting"
}

// AddScript registers a new script.
func (s *ScriptingComponent) AddScript(script *Script) error {
	if script == nil {
		return fmt.Errorf("script cannot be nil")
	}
	if script.ID == "" {
		return fmt.Errorf("script ID cannot be empty")
	}
	if script.Source == "" {
		return fmt.Errorf("script source cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Scripts[script.ID]; exists {
		return fmt.Errorf("script %s already exists", script.ID)
	}

	script.Enabled = true
	s.Scripts[script.ID] = script
	return nil
}

// RemoveScript removes a script by ID.
func (s *ScriptingComponent) RemoveScript(scriptID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Scripts[scriptID]; !exists {
		return fmt.Errorf("script %s not found", scriptID)
	}

	delete(s.Scripts, scriptID)
	return nil
}

// GetScript returns a script by ID.
func (s *ScriptingComponent) GetScript(scriptID string) (*Script, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	script, exists := s.Scripts[scriptID]
	return script, exists
}

// GetScriptsByEvent returns all scripts triggered by the given event.
func (s *ScriptingComponent) GetScriptsByEvent(eventType string) []*Script {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var scripts []*Script
	for _, script := range s.Scripts {
		if script.TriggerEvent == eventType && script.Enabled {
			scripts = append(scripts, script)
		}
	}
	return scripts
}

// SetVariable sets a shared variable.
func (s *ScriptingComponent) SetVariable(name string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Variables[name] = value
}

// GetVariable gets a shared variable.
func (s *ScriptingComponent) GetVariable(name string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.Variables[name]
	return val, exists
}

// ClearVariables clears all shared variables.
func (s *ScriptingComponent) ClearVariables() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Variables = make(map[string]any)
}

// GetScriptCount returns the number of registered scripts.
func (s *ScriptingComponent) GetScriptCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.Scripts)
}

// GetActiveScriptCount returns the number of enabled scripts.
func (s *ScriptingComponent) GetActiveScriptCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, script := range s.Scripts {
		if script.Enabled {
			count++
		}
	}
	return count
}

// SetScriptEnabled enables or disables a script.
func (s *ScriptingComponent) SetScriptEnabled(scriptID string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	script, exists := s.Scripts[scriptID]
	if !exists {
		return fmt.Errorf("script %s not found", scriptID)
	}

	script.Enabled = enabled
	return nil
}

// ScriptingData represents serializable scripting data.
type ScriptingData struct {
	Scripts   map[string]*Script `json:"scripts"`
	Variables map[string]any     `json:"variables"`
	Enabled   bool               `json:"enabled"`
	LastError string             `json:"last_error,omitempty"`
}

// Serialize converts the component to JSON bytes.
func (s *ScriptingComponent) Serialize() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := ScriptingData{
		Scripts:   s.Scripts,
		Variables: s.Variables,
		Enabled:   s.Enabled,
		LastError: s.LastError,
	}

	return json.Marshal(data)
}

// Deserialize loads the component from JSON bytes.
func (s *ScriptingComponent) Deserialize(data []byte) error {
	var scriptData ScriptingData
	if err := json.Unmarshal(data, &scriptData); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if scriptData.Scripts != nil {
		s.Scripts = scriptData.Scripts
	} else {
		s.Scripts = make(map[string]*Script)
	}

	if scriptData.Variables != nil {
		s.Variables = scriptData.Variables
	} else {
		s.Variables = make(map[string]any)
	}

	s.Enabled = scriptData.Enabled
	s.LastError = scriptData.LastError

	return nil
}
