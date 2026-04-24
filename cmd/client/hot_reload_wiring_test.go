//go:build !android && !ios
// +build !android,!ios

// Package main contains integration tests for hot-reload wiring.
package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// TestHotReloadSystemRegistered verifies that HotReloadSystem is registered in the
// world after initializeModBrowserWiring is called (AUDIT.md G15).
func TestHotReloadSystemRegistered(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	clientLogger := logger.WithField("test", "hot_reload_wiring")

	game := &engine.EbitenGame{
		World:        engine.NewWorldWithLogger(logger),
		ScreenWidth:  800,
		ScreenHeight: 600,
	}
	game.CameraSystem = engine.NewCameraSystem(800, 600)
	game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

	sys := initializeCoreSystems(game, logger, clientLogger)

	initializeModBrowserWiring(game, sys, clientLogger)

	found := false
	for _, s := range game.World.GetSystems() {
		if _, ok := s.(*engine.HotReloadSystem); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("HotReloadSystem not found in world after initializeModBrowserWiring (AUDIT.md G15)")
	}
}

// TestModBrowserSystemWired verifies that the mod browser system is configured
// after initializeModBrowserWiring is called (AUDIT.md G16).
func TestModBrowserSystemWired(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	clientLogger := logger.WithField("test", "mod_browser_wired")

	game := &engine.EbitenGame{
		World:        engine.NewWorldWithLogger(logger),
		ScreenWidth:  800,
		ScreenHeight: 600,
	}
	game.CameraSystem = engine.NewCameraSystem(800, 600)
	game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

	sys := initializeCoreSystems(game, logger, clientLogger)

	initializeModBrowserWiring(game, sys, clientLogger)

	// Verify the mod browser system exists in the world (AUDIT.md G16).
	if sys.modBrowserSys == nil {
		t.Fatal("modBrowserSys is nil after initializeModBrowserWiring (AUDIT.md G16)")
	}

	found := false
	for _, s := range game.World.GetSystems() {
		if _, ok := s.(*engine.ModBrowserSystem); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("ModBrowserSystem not found in world after initializeModBrowserWiring (AUDIT.md G16)")
	}
}
