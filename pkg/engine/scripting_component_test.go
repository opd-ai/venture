package engine

import (
	"encoding/json"
	"testing"
)

func TestScriptingComponent_Type(t *testing.T) {
	c := NewScriptingComponent()
	if c.Type() != "scripting" {
		t.Errorf("Type() = %s, want scripting", c.Type())
	}
}

func TestScriptingComponent_AddScript(t *testing.T) {
	tests := []struct {
		name    string
		script  *Script
		wantErr bool
	}{
		{
			name: "valid script",
			script: &Script{
				ID:           "test_script",
				ModID:        "test_mod",
				Source:       "return 42",
				TriggerEvent: "on_tick",
			},
			wantErr: false,
		},
		{
			name:    "nil script",
			script:  nil,
			wantErr: true,
		},
		{
			name: "empty ID",
			script: &Script{
				ID:     "",
				Source: "return 1",
			},
			wantErr: true,
		},
		{
			name: "empty source",
			script: &Script{
				ID:     "test",
				Source: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewScriptingComponent()
			err := c.AddScript(tt.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddScript() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScriptingComponent_AddScript_Duplicate(t *testing.T) {
	c := NewScriptingComponent()
	script := &Script{ID: "dup", Source: "test"}

	if err := c.AddScript(script); err != nil {
		t.Fatalf("first AddScript() failed: %v", err)
	}

	err := c.AddScript(&Script{ID: "dup", Source: "test2"})
	if err == nil {
		t.Error("AddScript() should fail for duplicate ID")
	}
}

func TestScriptingComponent_RemoveScript(t *testing.T) {
	c := NewScriptingComponent()
	script := &Script{ID: "remove_me", Source: "test"}
	_ = c.AddScript(script)

	if err := c.RemoveScript("remove_me"); err != nil {
		t.Errorf("RemoveScript() error = %v", err)
	}

	if c.GetScriptCount() != 0 {
		t.Error("script not removed")
	}

	// Remove non-existent
	if err := c.RemoveScript("nonexistent"); err == nil {
		t.Error("RemoveScript() should fail for non-existent script")
	}
}

func TestScriptingComponent_GetScript(t *testing.T) {
	c := NewScriptingComponent()
	script := &Script{ID: "get_test", ModID: "mod1", Source: "test"}
	_ = c.AddScript(script)

	got, exists := c.GetScript("get_test")
	if !exists {
		t.Error("GetScript() should find script")
	}
	if got.ModID != "mod1" {
		t.Errorf("GetScript() ModID = %s, want mod1", got.ModID)
	}

	_, exists = c.GetScript("nonexistent")
	if exists {
		t.Error("GetScript() should not find nonexistent script")
	}
}

func TestScriptingComponent_GetScriptsByEvent(t *testing.T) {
	c := NewScriptingComponent()
	_ = c.AddScript(&Script{ID: "s1", Source: "1", TriggerEvent: "on_tick"})
	_ = c.AddScript(&Script{ID: "s2", Source: "2", TriggerEvent: "on_tick"})
	_ = c.AddScript(&Script{ID: "s3", Source: "3", TriggerEvent: "on_spawn"})

	scripts := c.GetScriptsByEvent("on_tick")
	if len(scripts) != 2 {
		t.Errorf("GetScriptsByEvent() = %d scripts, want 2", len(scripts))
	}

	scripts = c.GetScriptsByEvent("on_spawn")
	if len(scripts) != 1 {
		t.Errorf("GetScriptsByEvent() = %d scripts, want 1", len(scripts))
	}

	scripts = c.GetScriptsByEvent("nonexistent")
	if len(scripts) != 0 {
		t.Errorf("GetScriptsByEvent() = %d scripts, want 0", len(scripts))
	}
}

func TestScriptingComponent_Variables(t *testing.T) {
	c := NewScriptingComponent()

	c.SetVariable("counter", 42)
	c.SetVariable("name", "test")

	val, exists := c.GetVariable("counter")
	if !exists || val != 42 {
		t.Errorf("GetVariable(counter) = %v, %v, want 42, true", val, exists)
	}

	val, exists = c.GetVariable("name")
	if !exists || val != "test" {
		t.Errorf("GetVariable(name) = %v, %v, want test, true", val, exists)
	}

	_, exists = c.GetVariable("nonexistent")
	if exists {
		t.Error("GetVariable() should not find nonexistent variable")
	}

	c.ClearVariables()
	if len(c.Variables) != 0 {
		t.Error("ClearVariables() should clear all variables")
	}
}

func TestScriptingComponent_ScriptCounts(t *testing.T) {
	c := NewScriptingComponent()

	if c.GetScriptCount() != 0 {
		t.Error("initial count should be 0")
	}

	_ = c.AddScript(&Script{ID: "s1", Source: "1"})
	_ = c.AddScript(&Script{ID: "s2", Source: "2"})

	if c.GetScriptCount() != 2 {
		t.Errorf("GetScriptCount() = %d, want 2", c.GetScriptCount())
	}

	if c.GetActiveScriptCount() != 2 {
		t.Errorf("GetActiveScriptCount() = %d, want 2", c.GetActiveScriptCount())
	}

	// Disable one
	_ = c.SetScriptEnabled("s1", false)
	if c.GetActiveScriptCount() != 1 {
		t.Errorf("GetActiveScriptCount() = %d, want 1", c.GetActiveScriptCount())
	}
}

func TestScriptingComponent_SetScriptEnabled(t *testing.T) {
	c := NewScriptingComponent()
	_ = c.AddScript(&Script{ID: "s1", Source: "1"})

	if err := c.SetScriptEnabled("s1", false); err != nil {
		t.Errorf("SetScriptEnabled() error = %v", err)
	}

	script, _ := c.GetScript("s1")
	if script.Enabled {
		t.Error("script should be disabled")
	}

	if err := c.SetScriptEnabled("nonexistent", true); err == nil {
		t.Error("SetScriptEnabled() should fail for nonexistent script")
	}
}

func TestScriptingComponent_GetScriptsByEvent_DisabledScript(t *testing.T) {
	c := NewScriptingComponent()
	_ = c.AddScript(&Script{ID: "s1", Source: "1", TriggerEvent: "tick"})
	_ = c.SetScriptEnabled("s1", false)

	scripts := c.GetScriptsByEvent("tick")
	if len(scripts) != 0 {
		t.Error("disabled scripts should not be returned")
	}
}

func TestScriptingComponent_Serialize(t *testing.T) {
	c := NewScriptingComponent()
	_ = c.AddScript(&Script{ID: "s1", ModID: "m1", Source: "test"})
	c.SetVariable("x", 123)
	c.LastError = "test error"

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	var parsed ScriptingData
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(parsed.Scripts) != 1 {
		t.Error("serialized scripts missing")
	}
	if parsed.Variables["x"] != float64(123) { // JSON numbers are float64
		t.Error("serialized variables missing")
	}
	if parsed.LastError != "test error" {
		t.Error("serialized last error missing")
	}
}

func TestScriptingComponent_Deserialize(t *testing.T) {
	original := NewScriptingComponent()
	_ = original.AddScript(&Script{ID: "s1", ModID: "m1", Source: "test"})
	original.SetVariable("key", "value")
	original.Enabled = true

	data, _ := original.Serialize()

	restored := NewScriptingComponent()
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if restored.GetScriptCount() != 1 {
		t.Error("script not restored")
	}

	script, exists := restored.GetScript("s1")
	if !exists || script.ModID != "m1" {
		t.Error("script data not correctly restored")
	}

	val, exists := restored.GetVariable("key")
	if !exists || val != "value" {
		t.Error("variable not restored")
	}
}

func TestScriptingComponent_Deserialize_EmptyData(t *testing.T) {
	c := NewScriptingComponent()

	data := []byte(`{"enabled": true}`)
	if err := c.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if c.Scripts == nil {
		t.Error("Scripts map should be initialized")
	}
	if c.Variables == nil {
		t.Error("Variables map should be initialized")
	}
}

func TestScriptingComponent_Deserialize_InvalidJSON(t *testing.T) {
	c := NewScriptingComponent()
	if err := c.Deserialize([]byte("not json")); err == nil {
		t.Error("Deserialize() should fail for invalid JSON")
	}
}

func TestScript_Fields(t *testing.T) {
	script := &Script{
		ID:           "test",
		ModID:        "mod",
		Source:       "return true",
		TriggerEvent: "tick",
		Priority:     10,
		LastRun:      1234567890,
		RunCount:     5,
		Enabled:      true,
		Metadata:     map[string]any{"author": "test"},
	}

	if script.ID != "test" {
		t.Error("ID field incorrect")
	}
	if script.Priority != 10 {
		t.Error("Priority field incorrect")
	}
	if script.Metadata["author"] != "test" {
		t.Error("Metadata field incorrect")
	}
}

func BenchmarkScriptingComponent_AddScript(b *testing.B) {
	c := NewScriptingComponent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := &Script{
			ID:     "script_" + string(rune(i%26+65)),
			Source: "test",
		}
		c.Scripts = make(map[string]*Script) // Reset for each iteration
		_ = c.AddScript(script)
	}
}

func BenchmarkScriptingComponent_GetScriptsByEvent(b *testing.B) {
	c := NewScriptingComponent()
	for i := 0; i < 100; i++ {
		_ = c.AddScript(&Script{
			ID:           "script_" + string(rune(i)),
			Source:       "test",
			TriggerEvent: "tick",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.GetScriptsByEvent("tick")
	}
}
