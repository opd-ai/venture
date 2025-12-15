package engine

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ScriptingSystem executes mod scripts safely within a sandbox.
type ScriptingSystem struct {
	world     *World
	evaluator *ExpressionEvaluator

	// Built-in functions available to scripts
	builtins map[string]BuiltinFunc

	// Execution limits
	maxExecutionTime time.Duration
	maxMemoryBytes   int64

	// Statistics
	totalExecutions int64
	totalErrors     int64

	mu sync.RWMutex
}

// BuiltinFunc represents a function callable from scripts.
type BuiltinFunc func(args []any, ctx *ScriptContext) (any, error)

// ScriptContext provides execution context for a script.
type ScriptContext struct {
	ScriptID    string
	ModID       string
	EntityID    uint64
	Variables   map[string]any
	EventData   map[string]any
	StartTime   time.Time
	Evaluator   *ExpressionEvaluator
	World       *World
	ResultValue any
}

// NewScriptingSystem creates a new scripting system.
func NewScriptingSystem(world *World) *ScriptingSystem {
	s := &ScriptingSystem{
		world:            world,
		evaluator:        NewExpressionEvaluator(),
		builtins:         make(map[string]BuiltinFunc),
		maxExecutionTime: 10 * time.Millisecond,
		maxMemoryBytes:   1024 * 1024, // 1 MB
	}

	s.registerBuiltins()

	logrus.WithFields(logrus.Fields{
		"system_name": "scripting",
	}).Debug("Created scripting system")

	return s
}

// registerBuiltins registers default built-in functions.
func (s *ScriptingSystem) registerBuiltins() {
	// Math functions
	s.builtins["abs"] = builtinAbs
	s.builtins["min"] = builtinMin
	s.builtins["max"] = builtinMax
	s.builtins["floor"] = builtinFloor
	s.builtins["ceil"] = builtinCeil
	s.builtins["round"] = builtinRound
	s.builtins["clamp"] = builtinClamp

	// Logic functions
	s.builtins["if"] = builtinIf
	s.builtins["not"] = builtinNot
	s.builtins["and"] = builtinAnd
	s.builtins["or"] = builtinOr

	// String functions
	s.builtins["len"] = builtinLen
	s.builtins["concat"] = builtinConcat
	s.builtins["upper"] = builtinUpper
	s.builtins["lower"] = builtinLower

	// Entity functions
	s.builtins["get_component"] = s.builtinGetComponent
	s.builtins["has_component"] = s.builtinHasComponent
	s.builtins["entity_count"] = s.builtinEntityCount

	// Variable functions
	s.builtins["get"] = builtinGet
	s.builtins["set"] = builtinSet
}

// Update processes scripts for the given entities.
func (s *ScriptingSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		compRaw, ok := entity.GetComponent("scripting")
		if !ok {
			continue
		}
		comp, ok := compRaw.(*ScriptingComponent)
		if !ok || !comp.Enabled {
			continue
		}

		// Execute tick scripts
		scripts := comp.GetScriptsByEvent("on_tick")
		s.executeScripts(scripts, comp, entity.ID, map[string]any{
			"delta_time": deltaTime,
		})
	}
}

// ExecuteEvent runs all scripts triggered by the given event.
func (s *ScriptingSystem) ExecuteEvent(entityID uint64, eventType string, eventData map[string]any) error {
	if s.world == nil {
		return fmt.Errorf("world not set")
	}

	entity, found := s.world.GetEntity(entityID)
	if !found {
		return fmt.Errorf("entity not found: %d", entityID)
	}

	compRaw, ok := entity.GetComponent("scripting")
	if !ok {
		return nil
	}
	comp, ok := compRaw.(*ScriptingComponent)
	if !ok || !comp.Enabled {
		return nil
	}

	scripts := comp.GetScriptsByEvent(eventType)
	return s.executeScripts(scripts, comp, entityID, eventData)
}

// executeScripts runs a list of scripts in priority order.
func (s *ScriptingSystem) executeScripts(scripts []*Script, comp *ScriptingComponent, entityID uint64, eventData map[string]any) error {
	if len(scripts) == 0 {
		return nil
	}

	// Sort by priority
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].Priority < scripts[j].Priority
	})

	var lastErr error
	for _, script := range scripts {
		if !script.Enabled {
			continue
		}

		ctx := &ScriptContext{
			ScriptID:  script.ID,
			ModID:     script.ModID,
			EntityID:  entityID,
			Variables: comp.Variables,
			EventData: eventData,
			StartTime: time.Now(),
			Evaluator: s.evaluator,
			World:     s.world,
		}

		result, err := s.executeScript(script, ctx)

		s.mu.Lock()
		s.totalExecutions++
		if err != nil {
			s.totalErrors++
			comp.LastError = err.Error()
			lastErr = err
		}
		s.mu.Unlock()

		// Update script stats
		script.LastRun = time.Now().Unix()
		script.RunCount++
		ctx.ResultValue = result
	}

	return lastErr
}

// executeScript runs a single script and returns the result.
func (s *ScriptingSystem) executeScript(script *Script, ctx *ScriptContext) (any, error) {
	// Check execution time limit
	deadline := ctx.StartTime.Add(s.maxExecutionTime)

	// Parse and evaluate the script source
	result, err := s.evaluateSource(script.Source, ctx)
	if err != nil {
		return nil, fmt.Errorf("script %s: %w", script.ID, err)
	}

	// Check if we exceeded time limit
	if time.Now().After(deadline) {
		return nil, fmt.Errorf("script %s: execution time limit exceeded", script.ID)
	}

	return result, nil
}

// evaluateSource evaluates script source code.
func (s *ScriptingSystem) evaluateSource(source string, ctx *ScriptContext) (any, error) {
	// Split into statements
	statements := strings.Split(source, ";")

	var result any
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Check for function calls
		if strings.Contains(stmt, "(") {
			r, err := s.evaluateFunctionCall(stmt, ctx)
			if err != nil {
				return nil, err
			}
			result = r
			continue
		}

		// Check for variable assignment
		if strings.Contains(stmt, "=") && !strings.Contains(stmt, "==") {
			parts := strings.SplitN(stmt, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				value, err := s.evaluator.Evaluate(strings.TrimSpace(parts[1]), ctx.Variables)
				if err != nil {
					return nil, err
				}
				ctx.Variables[name] = value
				result = value
				continue
			}
		}

		// Otherwise, evaluate as expression
		r, err := s.evaluator.Evaluate(stmt, ctx.Variables)
		if err != nil {
			return nil, err
		}
		result = r
	}

	return result, nil
}

// evaluateFunctionCall parses and executes a function call.
func (s *ScriptingSystem) evaluateFunctionCall(stmt string, ctx *ScriptContext) (any, error) {
	// Parse function name and args
	re := regexp.MustCompile(`^(\w+)\s*\((.*)\)$`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid function call: %s", stmt)
	}

	funcName := matches[1]
	argsStr := matches[2]

	// Parse arguments
	args, err := s.parseArguments(argsStr, ctx)
	if err != nil {
		return nil, err
	}

	// Look up builtin
	builtin, exists := s.builtins[funcName]
	if !exists {
		return nil, fmt.Errorf("unknown function: %s", funcName)
	}

	return builtin(args, ctx)
}

// parseArguments parses comma-separated function arguments.
func (s *ScriptingSystem) parseArguments(argsStr string, ctx *ScriptContext) ([]any, error) {
	if strings.TrimSpace(argsStr) == "" {
		return nil, nil
	}

	var args []any
	parts := strings.Split(argsStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Evaluate the argument
		val, err := s.evaluator.Evaluate(part, ctx.Variables)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}

	return args, nil
}

// RegisterBuiltin adds a custom built-in function.
func (s *ScriptingSystem) RegisterBuiltin(name string, fn BuiltinFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.builtins[name] = fn
}

// GetStatistics returns execution statistics.
func (s *ScriptingSystem) GetStatistics() (executions, errors int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.totalExecutions, s.totalErrors
}

// SetExecutionLimit sets the maximum execution time per script.
func (s *ScriptingSystem) SetExecutionLimit(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.maxExecutionTime = d
}

// ExpressionEvaluator evaluates simple expressions.
type ExpressionEvaluator struct{}

// NewExpressionEvaluator creates a new expression evaluator.
func NewExpressionEvaluator() *ExpressionEvaluator {
	return &ExpressionEvaluator{}
}

// Evaluate evaluates an expression with the given variables.
func (e *ExpressionEvaluator) Evaluate(expr string, vars map[string]any) (any, error) {
	expr = strings.TrimSpace(expr)

	// Check for string literal
	if strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`) {
		return expr[1 : len(expr)-1], nil
	}

	// Check for boolean
	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}

	// Check for number
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return num, nil
	}

	// Check for variable
	if val, exists := vars[expr]; exists {
		return val, nil
	}

	// Check for comparison operators
	if result, ok, err := e.evaluateComparison(expr, vars); ok {
		return result, err
	}

	// Check for arithmetic
	if result, ok, err := e.evaluateArithmetic(expr, vars); ok {
		return result, err
	}

	return nil, fmt.Errorf("cannot evaluate: %s", expr)
}

// evaluateComparison handles comparison operators.
func (e *ExpressionEvaluator) evaluateComparison(expr string, vars map[string]any) (any, bool, error) {
	operators := []string{">=", "<=", "==", "!=", ">", "<"}

	for _, op := range operators {
		if strings.Contains(expr, op) {
			parts := strings.SplitN(expr, op, 2)
			if len(parts) == 2 {
				left, err := e.Evaluate(strings.TrimSpace(parts[0]), vars)
				if err != nil {
					return nil, true, err
				}
				right, err := e.Evaluate(strings.TrimSpace(parts[1]), vars)
				if err != nil {
					return nil, true, err
				}

				leftNum, lok := toFloat64(left)
				rightNum, rok := toFloat64(right)

				if lok && rok {
					switch op {
					case ">=":
						return leftNum >= rightNum, true, nil
					case "<=":
						return leftNum <= rightNum, true, nil
					case "==":
						return leftNum == rightNum, true, nil
					case "!=":
						return leftNum != rightNum, true, nil
					case ">":
						return leftNum > rightNum, true, nil
					case "<":
						return leftNum < rightNum, true, nil
					}
				}
			}
		}
	}

	return nil, false, nil
}

// evaluateArithmetic handles arithmetic operators.
func (e *ExpressionEvaluator) evaluateArithmetic(expr string, vars map[string]any) (any, bool, error) {
	// Handle operators in order of precedence (low to high)
	for _, op := range []string{"+", "-", "*", "/", "%"} {
		idx := strings.LastIndex(expr, op)
		if idx > 0 && idx < len(expr)-1 {
			left, err := e.Evaluate(expr[:idx], vars)
			if err != nil {
				continue
			}
			right, err := e.Evaluate(expr[idx+1:], vars)
			if err != nil {
				continue
			}

			leftNum, lok := toFloat64(left)
			rightNum, rok := toFloat64(right)

			if lok && rok {
				switch op {
				case "+":
					return leftNum + rightNum, true, nil
				case "-":
					return leftNum - rightNum, true, nil
				case "*":
					return leftNum * rightNum, true, nil
				case "/":
					if rightNum == 0 {
						return nil, true, fmt.Errorf("division by zero")
					}
					return leftNum / rightNum, true, nil
				case "%":
					if rightNum == 0 {
						return nil, true, fmt.Errorf("modulo by zero")
					}
					return math.Mod(leftNum, rightNum), true, nil
				}
			}
		}
	}

	return nil, false, nil
}

// toFloat64 converts a value to float64 if possible.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	}
	return 0, false
}

// Built-in function implementations

func builtinAbs(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("abs requires 1 argument")
	}
	num, ok := toFloat64(args[0])
	if !ok {
		return nil, fmt.Errorf("abs requires numeric argument")
	}
	return math.Abs(num), nil
}

func builtinMin(args []any, ctx *ScriptContext) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("min requires at least 2 arguments")
	}
	result, ok := toFloat64(args[0])
	if !ok {
		return nil, fmt.Errorf("min requires numeric arguments")
	}
	for i := 1; i < len(args); i++ {
		num, ok := toFloat64(args[i])
		if !ok {
			return nil, fmt.Errorf("min requires numeric arguments")
		}
		if num < result {
			result = num
		}
	}
	return result, nil
}

func builtinMax(args []any, ctx *ScriptContext) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("max requires at least 2 arguments")
	}
	result, ok := toFloat64(args[0])
	if !ok {
		return nil, fmt.Errorf("max requires numeric arguments")
	}
	for i := 1; i < len(args); i++ {
		num, ok := toFloat64(args[i])
		if !ok {
			return nil, fmt.Errorf("max requires numeric arguments")
		}
		if num > result {
			result = num
		}
	}
	return result, nil
}

func builtinFloor(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("floor requires 1 argument")
	}
	num, ok := toFloat64(args[0])
	if !ok {
		return nil, fmt.Errorf("floor requires numeric argument")
	}
	return math.Floor(num), nil
}

func builtinCeil(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ceil requires 1 argument")
	}
	num, ok := toFloat64(args[0])
	if !ok {
		return nil, fmt.Errorf("ceil requires numeric argument")
	}
	return math.Ceil(num), nil
}

func builtinRound(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("round requires 1 argument")
	}
	num, ok := toFloat64(args[0])
	if !ok {
		return nil, fmt.Errorf("round requires numeric argument")
	}
	return math.Round(num), nil
}

func builtinClamp(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("clamp requires 3 arguments (value, min, max)")
	}
	val, ok1 := toFloat64(args[0])
	minVal, ok2 := toFloat64(args[1])
	maxVal, ok3 := toFloat64(args[2])
	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("clamp requires numeric arguments")
	}
	if val < minVal {
		return minVal, nil
	}
	if val > maxVal {
		return maxVal, nil
	}
	return val, nil
}

func builtinIf(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("if requires 3 arguments (condition, then, else)")
	}
	cond, ok := args[0].(bool)
	if !ok {
		return nil, fmt.Errorf("if requires boolean condition")
	}
	if cond {
		return args[1], nil
	}
	return args[2], nil
}

func builtinNot(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("not requires 1 argument")
	}
	val, ok := args[0].(bool)
	if !ok {
		return nil, fmt.Errorf("not requires boolean argument")
	}
	return !val, nil
}

func builtinAnd(args []any, ctx *ScriptContext) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("and requires at least 2 arguments")
	}
	for _, arg := range args {
		val, ok := arg.(bool)
		if !ok {
			return nil, fmt.Errorf("and requires boolean arguments")
		}
		if !val {
			return false, nil
		}
	}
	return true, nil
}

func builtinOr(args []any, ctx *ScriptContext) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("or requires at least 2 arguments")
	}
	for _, arg := range args {
		val, ok := arg.(bool)
		if !ok {
			return nil, fmt.Errorf("or requires boolean arguments")
		}
		if val {
			return true, nil
		}
	}
	return false, nil
}

func builtinLen(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("len requires 1 argument")
	}
	switch v := args[0].(type) {
	case string:
		return float64(len(v)), nil
	case []any:
		return float64(len(v)), nil
	}
	return nil, fmt.Errorf("len requires string or array argument")
}

func builtinConcat(args []any, ctx *ScriptContext) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("concat requires at least 2 arguments")
	}
	var result strings.Builder
	for _, arg := range args {
		result.WriteString(fmt.Sprint(arg))
	}
	return result.String(), nil
}

func builtinUpper(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("upper requires 1 argument")
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("upper requires string argument")
	}
	return strings.ToUpper(str), nil
}

func builtinLower(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("lower requires 1 argument")
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("lower requires string argument")
	}
	return strings.ToLower(str), nil
}

func builtinGet(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("get requires 1 argument")
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("get requires string argument")
	}
	return ctx.Variables[name], nil
}

func builtinSet(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("set requires 2 arguments (name, value)")
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("set requires string name")
	}
	ctx.Variables[name] = args[1]
	return args[1], nil
}

// Entity-related builtins (methods on ScriptingSystem)

func (s *ScriptingSystem) builtinGetComponent(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("get_component requires 1 argument (component_type)")
	}
	compType, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("get_component requires string argument")
	}

	if s.world == nil || ctx.EntityID == 0 {
		return nil, nil
	}

	entity, found := s.world.GetEntity(ctx.EntityID)
	if !found {
		return nil, nil
	}

	comp, exists := entity.GetComponent(compType)
	if !exists {
		return nil, nil
	}

	return comp, nil
}

func (s *ScriptingSystem) builtinHasComponent(args []any, ctx *ScriptContext) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("has_component requires 1 argument (component_type)")
	}
	compType, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("has_component requires string argument")
	}

	if s.world == nil || ctx.EntityID == 0 {
		return false, nil
	}

	entity, found := s.world.GetEntity(ctx.EntityID)
	if !found {
		return false, nil
	}

	return entity.HasComponent(compType), nil
}

func (s *ScriptingSystem) builtinEntityCount(args []any, ctx *ScriptContext) (any, error) {
	if s.world == nil {
		return float64(0), nil
	}
	return float64(len(s.world.GetEntities())), nil
}
