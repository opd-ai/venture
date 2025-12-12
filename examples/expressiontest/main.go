package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command line flags
	expressionName := flag.String("expression", "wave", "Expression to test: wave, cheer, dance, laugh, cry, sit, point, salute, shrug, thumbsup, facepalm, sleep, or 'all' for all expressions")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	listExprs := flag.Bool("list", false, "List all available expressions")

	flag.Parse()

	// Initialize logger
	logger := logging.TestUtilityLogger("expressiontest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.WithFields(logrus.Fields{
		"expression": *expressionName,
	}).Info("Expression Test Tool started")

	// List expressions if requested
	if *listExprs {
		listExpressions()
		return
	}

	// Map of expression names to types
	expressionMap := map[string]engine.ExpressionType{
		"wave":     engine.ExpressionWave,
		"cheer":    engine.ExpressionCheer,
		"dance":    engine.ExpressionDance,
		"laugh":    engine.ExpressionLaugh,
		"cry":      engine.ExpressionCry,
		"sit":      engine.ExpressionSit,
		"point":    engine.ExpressionPoint,
		"salute":   engine.ExpressionSalute,
		"shrug":    engine.ExpressionShrug,
		"thumbsup": engine.ExpressionThumbsUp,
		"facepalm": engine.ExpressionFacepalm,
		"sleep":    engine.ExpressionSleep,
	}

	// Test all expressions or specific one
	if *expressionName == "all" {
		testAllExpressions(expressionMap, *verbose, logger)
	} else {
		expType, ok := expressionMap[*expressionName]
		if !ok {
			logger.WithField("expression", *expressionName).Error("unknown expression")
			fmt.Fprintf(os.Stderr, "Unknown expression: %s\n", *expressionName)
			fmt.Fprintf(os.Stderr, "Use -list to see available expressions\n")
			os.Exit(1)
		}
		testExpression(expType, *verbose, logger)
	}

	logger.Info("expression test completed")
}

func listExpressions() {
	fmt.Println("=== Available Expressions ===")
	fmt.Println()

	expressions := []struct {
		name   string
		hotkey string
		desc   string
	}{
		{"wave", "Shift+1", "Wave hand greeting"},
		{"cheer", "Shift+2", "Jump for joy"},
		{"dance", "Shift+3", "Dance animation (loops)"},
		{"laugh", "Shift+4", "Laughing animation"},
		{"cry", "Shift+5", "Crying animation"},
		{"sit", "Shift+6", "Sit down (infinite duration)"},
		{"point", "Shift+7", "Point at something"},
		{"salute", "Shift+8", "Military salute"},
		{"shrug", "Shift+9", "Shrug shoulders"},
		{"thumbsup", "Shift+0", "Thumbs up gesture"},
		{"facepalm", "Shift+-", "Facepalm gesture"},
		{"sleep", "Shift+=", "Sleep animation (infinite duration)"},
	}

	for _, expr := range expressions {
		fmt.Printf("%-12s %-10s %s\n", expr.name, expr.hotkey, expr.desc)
	}

	fmt.Println()
	fmt.Println("Usage: expressiontest -expression <name>")
	fmt.Println("       expressiontest -expression all")
	fmt.Println("       expressiontest -list")
}

func testExpression(expType engine.ExpressionType, verbose bool, logger *logrus.Logger) {
	fmt.Printf("=== Testing Expression: %s ===\n", expType.String())
	fmt.Println()

	// Create expression instance
	expr := engine.NewBaseExpression(expType)

	// Display expression properties
	fmt.Printf("Expression Type: %s\n", expType.String())
	fmt.Printf("Duration: %.2f seconds", expr.GetDuration())
	if expr.GetDuration() > 1000000 {
		fmt.Printf(" (infinite)")
	}
	fmt.Println()
	fmt.Printf("Sound Effect: %s\n", expr.GetSoundEffect())
	if expr.GetSoundEffect() == "" {
		fmt.Printf("  (no sound)\n")
	}
	fmt.Println()

	// Display animation properties
	anim := expr.GetAnimation()
	if anim != nil {
		fmt.Printf("Animation:\n")
		fmt.Printf("  Frame Count: %d\n", anim.GetFrameCount())
		fmt.Printf("  Frame Time: %.3f seconds\n", anim.GetFrameTime())
		fmt.Printf("  Loops: %t\n", anim.ShouldLoop())
		totalDuration := float64(anim.GetFrameCount()) * anim.GetFrameTime()
		fmt.Printf("  Total Animation Duration: %.2f seconds\n", totalDuration)
	}
	fmt.Println()

	// Test expression system integration
	world := engine.NewWorld()
	audioMgr := engine.NewAudioManager(44100, 12345)
	exprSys := engine.NewExpressionSystem(world, audioMgr)

	// Create test entity
	entity := world.CreateEntity()
	world.Update(0)

	// Trigger expression
	success := exprSys.TriggerExpression(entity.ID, expType)
	if success {
		fmt.Printf("✓ Expression triggered successfully\n")
	} else {
		fmt.Printf("✗ Expression trigger failed\n")
		logger.Error("expression trigger failed")
		os.Exit(1)
	}

	// Verify component
	expComp, ok := entity.GetComponent("expression")
	if !ok {
		fmt.Printf("✗ Expression component not found\n")
		logger.Error("expression component not found")
		os.Exit(1)
	}

	exprCompTyped := expComp.(*engine.ExpressionComponent)
	fmt.Printf("✓ Expression component verified\n")
	fmt.Printf("  Active Expression: %s\n", exprCompTyped.ActiveExpression.String())
	fmt.Printf("  Time Remaining: %.2f seconds\n", exprCompTyped.ExpressionTime)
	fmt.Printf("  Cooldown: %.2f seconds\n", exprCompTyped.Cooldown)
	fmt.Println()

	// Test cooldown
	success = exprSys.TriggerExpression(entity.ID, engine.ExpressionWave)
	if !success {
		fmt.Printf("✓ Cooldown working (prevented immediate re-trigger)\n")
	} else {
		fmt.Printf("✗ Cooldown not working (allowed immediate re-trigger)\n")
		logger.Error("cooldown not working")
	}
	fmt.Println()

	// Simulate update
	fmt.Printf("Simulating 1 second update...\n")
	exprSys.Update(1.0)
	fmt.Printf("  Time Remaining: %.2f seconds\n", exprCompTyped.ExpressionTime)
	fmt.Printf("  Cooldown: %.2f seconds\n", exprCompTyped.Cooldown)
	fmt.Println()

	if verbose {
		logger.WithFields(logrus.Fields{
			"expression": expType.String(),
			"duration":   expr.GetDuration(),
			"sound":      expr.GetSoundEffect(),
			"frameCount": anim.GetFrameCount(),
			"frameTime":  anim.GetFrameTime(),
			"loops":      anim.ShouldLoop(),
		}).Debug("expression details")
	}

	fmt.Printf("=== Test Passed ===\n")
}

func testAllExpressions(expressionMap map[string]engine.ExpressionType, verbose bool, logger *logrus.Logger) {
	fmt.Println("=== Testing All Expressions ===")
	fmt.Println()

	// Test each expression
	expressionNames := []string{
		"wave", "cheer", "dance", "laugh", "cry", "sit",
		"point", "salute", "shrug", "thumbsup", "facepalm", "sleep",
	}

	passed := 0
	failed := 0

	for _, name := range expressionNames {
		expType := expressionMap[name]
		expr := engine.NewBaseExpression(expType)

		fmt.Printf("%-12s ", name)

		// Quick validation
		if expr == nil {
			fmt.Printf("✗ FAILED (nil expression)\n")
			failed++
			continue
		}

		anim := expr.GetAnimation()
		if anim == nil {
			fmt.Printf("✗ FAILED (nil animation)\n")
			failed++
			continue
		}

		if anim.GetFrameCount() <= 0 {
			fmt.Printf("✗ FAILED (invalid frame count)\n")
			failed++
			continue
		}

		if anim.GetFrameTime() <= 0 {
			fmt.Printf("✗ FAILED (invalid frame time)\n")
			failed++
			continue
		}

		if expr.GetDuration() <= 0 && expr.GetDuration() != 1.7976931348623157e+308 { // not infinite
			fmt.Printf("✗ FAILED (invalid duration)\n")
			failed++
			continue
		}

		fmt.Printf("✓ PASSED\n")
		passed++

		if verbose {
			fmt.Printf("  Duration: %.2fs, Frames: %d, Sound: %s\n",
				expr.GetDuration(), anim.GetFrameCount(), expr.GetSoundEffect())
		}
	}

	fmt.Println()
	fmt.Printf("=== Results ===\n")
	fmt.Printf("Passed: %d/%d\n", passed, len(expressionNames))
	fmt.Printf("Failed: %d/%d\n", failed, len(expressionNames))
	fmt.Println()

	if failed > 0 {
		logger.WithField("failed", failed).Error("some expressions failed")
		os.Exit(1)
	}

	fmt.Printf("=== All Tests Passed ===\n")
}
