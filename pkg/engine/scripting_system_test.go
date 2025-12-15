package engine

import (
	"testing"
	"time"
)

func TestNewScriptingSystem(t *testing.T) {
	s := NewScriptingSystem(nil)
	if s == nil {
		t.Fatal("NewScriptingSystem returned nil")
	}
	if s.evaluator == nil {
		t.Error("evaluator not initialized")
	}
	if len(s.builtins) == 0 {
		t.Error("builtins not registered")
	}
}

func TestScriptingSystem_RegisterBuiltin(t *testing.T) {
	s := NewScriptingSystem(nil)

	called := false
	s.RegisterBuiltin("test_func", func(args []any, ctx *ScriptContext) (any, error) {
		called = true
		return "ok", nil
	})

	ctx := &ScriptContext{Variables: make(map[string]any)}
	result, err := s.evaluateFunctionCall("test_func()", ctx)
	if err != nil {
		t.Errorf("function call error: %v", err)
	}
	if !called {
		t.Error("custom builtin was not called")
	}
	if result != "ok" {
		t.Errorf("result = %v, want ok", result)
	}
}

func TestScriptingSystem_ExecuteEvent(t *testing.T) {
	world := NewWorld()
	s := NewScriptingSystem(world)

	entity := world.CreateEntity()
	comp := NewScriptingComponent()
	_ = comp.AddScript(&Script{
		ID:           "test",
		Source:       "x = 42",
		TriggerEvent: "on_spawn",
	})
	entity.AddComponent(comp)
	world.FlushPendingEntities()

	err := s.ExecuteEvent(entity.ID, "on_spawn", nil)
	if err != nil {
		t.Errorf("ExecuteEvent() error = %v", err)
	}

	if comp.Variables["x"] != float64(42) {
		t.Errorf("variable x = %v, want 42", comp.Variables["x"])
	}
}

func TestScriptingSystem_ExecuteEvent_NoEntity(t *testing.T) {
	world := NewWorld()
	s := NewScriptingSystem(world)

	err := s.ExecuteEvent(99999, "test", nil)
	if err == nil {
		t.Error("ExecuteEvent() should error for nonexistent entity")
	}
}

func TestScriptingSystem_ExecuteEvent_NoWorld(t *testing.T) {
	s := NewScriptingSystem(nil)

	err := s.ExecuteEvent(1, "test", nil)
	if err == nil {
		t.Error("ExecuteEvent() should error when world is nil")
	}
}

func TestScriptingSystem_Update(t *testing.T) {
	world := NewWorld()
	s := NewScriptingSystem(world)

	entity := world.CreateEntity()
	comp := NewScriptingComponent()
	_ = comp.AddScript(&Script{
		ID:           "tick_script",
		Source:       "counter = 1",
		TriggerEvent: "on_tick",
	})
	entity.AddComponent(comp)

	s.Update([]*Entity{entity}, 0.016)

	if comp.Variables["counter"] != float64(1) {
		t.Error("tick script did not execute")
	}
}

func TestScriptingSystem_Update_DisabledComponent(t *testing.T) {
	world := NewWorld()
	s := NewScriptingSystem(world)

	entity := world.CreateEntity()
	comp := NewScriptingComponent()
	comp.Enabled = false
	_ = comp.AddScript(&Script{
		ID:           "tick",
		Source:       "x = 1",
		TriggerEvent: "on_tick",
	})
	entity.AddComponent(comp)

	s.Update([]*Entity{entity}, 0.016)

	if _, exists := comp.Variables["x"]; exists {
		t.Error("disabled component should not execute scripts")
	}
}

func TestScriptingSystem_GetStatistics(t *testing.T) {
	world := NewWorld()
	s := NewScriptingSystem(world)

	entity := world.CreateEntity()
	comp := NewScriptingComponent()
	_ = comp.AddScript(&Script{
		ID:           "test",
		Source:       "x = 1",
		TriggerEvent: "tick",
	})
	entity.AddComponent(comp)
	world.FlushPendingEntities()

	_ = s.ExecuteEvent(entity.ID, "tick", nil)
	_ = s.ExecuteEvent(entity.ID, "tick", nil)

	execs, errors := s.GetStatistics()
	if execs != 2 {
		t.Errorf("executions = %d, want 2", execs)
	}
	if errors != 0 {
		t.Errorf("errors = %d, want 0", errors)
	}
}

func TestScriptingSystem_SetExecutionLimit(t *testing.T) {
	s := NewScriptingSystem(nil)
	s.SetExecutionLimit(5 * time.Millisecond)

	if s.maxExecutionTime != 5*time.Millisecond {
		t.Error("execution limit not set")
	}
}

func TestExpressionEvaluator_Numbers(t *testing.T) {
	e := NewExpressionEvaluator()

	tests := []struct {
		expr string
		want float64
	}{
		{"42", 42},
		{"3.14", 3.14},
		{"-5", -5},
		{"0", 0},
	}

	for _, tt := range tests {
		result, err := e.Evaluate(tt.expr, nil)
		if err != nil {
			t.Errorf("Evaluate(%s) error = %v", tt.expr, err)
			continue
		}
		if result != tt.want {
			t.Errorf("Evaluate(%s) = %v, want %v", tt.expr, result, tt.want)
		}
	}
}

func TestExpressionEvaluator_Strings(t *testing.T) {
	e := NewExpressionEvaluator()

	result, err := e.Evaluate(`"hello"`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != "hello" {
		t.Errorf("result = %v, want hello", result)
	}
}

func TestExpressionEvaluator_Booleans(t *testing.T) {
	e := NewExpressionEvaluator()

	result, _ := e.Evaluate("true", nil)
	if result != true {
		t.Error("true should evaluate to true")
	}

	result, _ = e.Evaluate("false", nil)
	if result != false {
		t.Error("false should evaluate to false")
	}
}

func TestExpressionEvaluator_Variables(t *testing.T) {
	e := NewExpressionEvaluator()
	vars := map[string]any{
		"x": float64(10),
		"y": float64(20),
	}

	result, err := e.Evaluate("x", vars)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != float64(10) {
		t.Errorf("result = %v, want 10", result)
	}
}

func TestExpressionEvaluator_Arithmetic(t *testing.T) {
	e := NewExpressionEvaluator()

	tests := []struct {
		expr string
		want float64
	}{
		{"10+5", 15},
		{"10-3", 7},
		{"4*3", 12},
		{"15/3", 5},
		{"10%3", 1},
	}

	for _, tt := range tests {
		result, err := e.Evaluate(tt.expr, nil)
		if err != nil {
			t.Errorf("Evaluate(%s) error = %v", tt.expr, err)
			continue
		}
		if result != tt.want {
			t.Errorf("Evaluate(%s) = %v, want %v", tt.expr, result, tt.want)
		}
	}
}

func TestExpressionEvaluator_Comparison(t *testing.T) {
	e := NewExpressionEvaluator()

	tests := []struct {
		expr string
		want bool
	}{
		{"5>3", true},
		{"3>5", false},
		{"5<10", true},
		{"5>=5", true},
		{"5<=5", true},
		{"5==5", true},
		{"5!=3", true},
	}

	for _, tt := range tests {
		result, err := e.Evaluate(tt.expr, nil)
		if err != nil {
			t.Errorf("Evaluate(%s) error = %v", tt.expr, err)
			continue
		}
		if result != tt.want {
			t.Errorf("Evaluate(%s) = %v, want %v", tt.expr, result, tt.want)
		}
	}
}

func TestExpressionEvaluator_DivisionByZero(t *testing.T) {
	e := NewExpressionEvaluator()

	_, err := e.Evaluate("10/0", nil)
	if err == nil {
		t.Error("division by zero should error")
	}

	_, err = e.Evaluate("10%0", nil)
	if err == nil {
		t.Error("modulo by zero should error")
	}
}

func TestBuiltinAbs(t *testing.T) {
	ctx := &ScriptContext{}

	result, err := builtinAbs([]any{float64(-5)}, ctx)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != float64(5) {
		t.Errorf("abs(-5) = %v, want 5", result)
	}

	_, err = builtinAbs([]any{}, ctx)
	if err == nil {
		t.Error("abs() with no args should error")
	}

	_, err = builtinAbs([]any{"string"}, ctx)
	if err == nil {
		t.Error("abs(string) should error")
	}
}

func TestBuiltinMinMax(t *testing.T) {
	ctx := &ScriptContext{}

	result, err := builtinMin([]any{float64(5), float64(3), float64(8)}, ctx)
	if err != nil {
		t.Fatalf("min error: %v", err)
	}
	if result != float64(3) {
		t.Errorf("min(5,3,8) = %v, want 3", result)
	}

	result, err = builtinMax([]any{float64(5), float64(3), float64(8)}, ctx)
	if err != nil {
		t.Fatalf("max error: %v", err)
	}
	if result != float64(8) {
		t.Errorf("max(5,3,8) = %v, want 8", result)
	}

	_, err = builtinMin([]any{float64(1)}, ctx)
	if err == nil {
		t.Error("min with 1 arg should error")
	}
}

func TestBuiltinFloorCeilRound(t *testing.T) {
	ctx := &ScriptContext{}

	result, _ := builtinFloor([]any{float64(3.7)}, ctx)
	if result != float64(3) {
		t.Errorf("floor(3.7) = %v, want 3", result)
	}

	result, _ = builtinCeil([]any{float64(3.2)}, ctx)
	if result != float64(4) {
		t.Errorf("ceil(3.2) = %v, want 4", result)
	}

	result, _ = builtinRound([]any{float64(3.5)}, ctx)
	if result != float64(4) {
		t.Errorf("round(3.5) = %v, want 4", result)
	}
}

func TestBuiltinClamp(t *testing.T) {
	ctx := &ScriptContext{}

	tests := []struct {
		args []any
		want float64
	}{
		{[]any{float64(5), float64(0), float64(10)}, 5},
		{[]any{float64(-5), float64(0), float64(10)}, 0},
		{[]any{float64(15), float64(0), float64(10)}, 10},
	}

	for _, tt := range tests {
		result, err := builtinClamp(tt.args, ctx)
		if err != nil {
			t.Errorf("clamp error: %v", err)
			continue
		}
		if result != tt.want {
			t.Errorf("clamp(%v) = %v, want %v", tt.args, result, tt.want)
		}
	}

	_, err := builtinClamp([]any{float64(1), float64(2)}, ctx)
	if err == nil {
		t.Error("clamp with 2 args should error")
	}
}

func TestBuiltinLogic(t *testing.T) {
	ctx := &ScriptContext{}

	// if
	result, _ := builtinIf([]any{true, "yes", "no"}, ctx)
	if result != "yes" {
		t.Errorf("if(true) = %v, want yes", result)
	}

	result, _ = builtinIf([]any{false, "yes", "no"}, ctx)
	if result != "no" {
		t.Errorf("if(false) = %v, want no", result)
	}

	// not
	result, _ = builtinNot([]any{true}, ctx)
	if result != false {
		t.Error("not(true) should be false")
	}

	// and
	result, _ = builtinAnd([]any{true, true}, ctx)
	if result != true {
		t.Error("and(true, true) should be true")
	}

	result, _ = builtinAnd([]any{true, false}, ctx)
	if result != false {
		t.Error("and(true, false) should be false")
	}

	// or
	result, _ = builtinOr([]any{false, true}, ctx)
	if result != true {
		t.Error("or(false, true) should be true")
	}

	result, _ = builtinOr([]any{false, false}, ctx)
	if result != false {
		t.Error("or(false, false) should be false")
	}
}

func TestBuiltinString(t *testing.T) {
	ctx := &ScriptContext{}

	// len
	result, _ := builtinLen([]any{"hello"}, ctx)
	if result != float64(5) {
		t.Errorf("len(hello) = %v, want 5", result)
	}

	// concat
	result, _ = builtinConcat([]any{"hello", " ", "world"}, ctx)
	if result != "hello world" {
		t.Errorf("concat = %v, want 'hello world'", result)
	}

	// upper
	result, _ = builtinUpper([]any{"hello"}, ctx)
	if result != "HELLO" {
		t.Errorf("upper(hello) = %v, want HELLO", result)
	}

	// lower
	result, _ = builtinLower([]any{"HELLO"}, ctx)
	if result != "hello" {
		t.Errorf("lower(HELLO) = %v, want hello", result)
	}
}

func TestBuiltinGetSet(t *testing.T) {
	ctx := &ScriptContext{
		Variables: map[string]any{"x": float64(42)},
	}

	result, _ := builtinGet([]any{"x"}, ctx)
	if result != float64(42) {
		t.Errorf("get(x) = %v, want 42", result)
	}

	_, _ = builtinSet([]any{"y", float64(100)}, ctx)
	if ctx.Variables["y"] != float64(100) {
		t.Error("set did not set variable")
	}
}

func TestScriptingSystem_EntityBuiltins(t *testing.T) {
	world := NewWorld()
	s := NewScriptingSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	world.FlushPendingEntities()

	ctx := &ScriptContext{
		EntityID:  entity.ID,
		Variables: make(map[string]any),
		World:     world,
	}

	// has_component
	result, err := s.builtinHasComponent([]any{"position"}, ctx)
	if err != nil {
		t.Fatalf("has_component error: %v", err)
	}
	if result != true {
		t.Error("has_component(position) should be true")
	}

	result, _ = s.builtinHasComponent([]any{"nonexistent"}, ctx)
	if result != false {
		t.Error("has_component(nonexistent) should be false")
	}

	// get_component
	result, err = s.builtinGetComponent([]any{"position"}, ctx)
	if err != nil {
		t.Fatalf("get_component error: %v", err)
	}
	if result == nil {
		t.Error("get_component(position) should return component")
	}

	// entity_count
	result, err = s.builtinEntityCount(nil, ctx)
	if err != nil {
		t.Fatalf("entity_count error: %v", err)
	}
	if result != float64(1) {
		t.Errorf("entity_count = %v, want 1", result)
	}
}

func TestScriptingSystem_FunctionCall(t *testing.T) {
	s := NewScriptingSystem(nil)
	ctx := &ScriptContext{
		Variables: make(map[string]any),
	}

	// Test function call parsing
	result, err := s.evaluateFunctionCall("abs(-10)", ctx)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != float64(10) {
		t.Errorf("abs(-10) = %v, want 10", result)
	}

	// Unknown function
	_, err = s.evaluateFunctionCall("unknown_func()", ctx)
	if err == nil {
		t.Error("unknown function should error")
	}

	// Invalid syntax
	_, err = s.evaluateFunctionCall("invalid", ctx)
	if err == nil {
		t.Error("invalid function call should error")
	}
}

func TestScriptingSystem_EvaluateSource(t *testing.T) {
	s := NewScriptingSystem(nil)
	ctx := &ScriptContext{
		Variables: make(map[string]any),
	}

	// Multiple statements
	result, err := s.evaluateSource("x = 10; y = 20; x + y", ctx)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != float64(30) {
		t.Errorf("result = %v, want 30", result)
	}
	if ctx.Variables["x"] != float64(10) {
		t.Error("x variable not set")
	}
	if ctx.Variables["y"] != float64(20) {
		t.Error("y variable not set")
	}
}

func TestScriptingSystem_ScriptPriority(t *testing.T) {
	world := NewWorld()
	s := NewScriptingSystem(world)

	entity := world.CreateEntity()
	comp := NewScriptingComponent()

	// Add scripts in reverse priority order
	_ = comp.AddScript(&Script{
		ID:           "high_priority",
		Source:       "result = 1",
		TriggerEvent: "test",
		Priority:     100,
	})
	_ = comp.AddScript(&Script{
		ID:           "low_priority",
		Source:       "result = 2",
		TriggerEvent: "test",
		Priority:     1,
	})

	entity.AddComponent(comp)
	world.FlushPendingEntities()

	_ = s.ExecuteEvent(entity.ID, "test", nil)

	// Low priority (1) runs first, then high priority (100) overwrites
	if comp.Variables["result"] != float64(1) {
		t.Errorf("result = %v, want 1 (high priority ran last)", comp.Variables["result"])
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input any
		ok    bool
	}{
		{float64(42), true},
		{float32(3.14), true}, // Note: float32 to float64 may have precision differences
		{int(10), true},
		{int64(100), true},
		{int32(50), true},
		{"string", false},
		{nil, false},
	}

	for _, tt := range tests {
		result, ok := toFloat64(tt.input)
		if ok != tt.ok {
			t.Errorf("toFloat64(%v) ok = %v, want %v", tt.input, ok, tt.ok)
		}
		// Just verify we got a result for valid cases
		if ok && result == 0 {
			// Special case: 0 is valid for int(0) inputs
			if _, isInt := tt.input.(int); !isInt || tt.input.(int) != 0 {
				t.Errorf("toFloat64(%v) returned 0 unexpectedly", tt.input)
			}
		}
	}
}

func BenchmarkExpressionEvaluator_Evaluate(b *testing.B) {
	e := NewExpressionEvaluator()
	vars := map[string]any{"x": float64(10), "y": float64(20)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Evaluate("x+y*2", vars)
	}
}

func BenchmarkScriptingSystem_ExecuteScript(b *testing.B) {
	world := NewWorld()
	s := NewScriptingSystem(world)

	entity := world.CreateEntity()
	comp := NewScriptingComponent()
	_ = comp.AddScript(&Script{
		ID:           "bench",
		Source:       "x = 42; y = x * 2",
		TriggerEvent: "tick",
	})
	entity.AddComponent(comp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ExecuteEvent(entity.ID, "tick", nil)
	}
}
