//go:build !android && !ios
// +build !android,!ios

// Package main (system_reg.go) consolidates the repeated SetGenre + AddSystem
// registration pattern used for the majority of client-side systems.
package main

import "github.com/opd-ai/venture/pkg/engine"

// genreSystem is a local interface satisfied by every system that exposes a
// SetGenre method. It is intentionally unexported and scoped to this package.
type genreSystem interface {
	engine.System
	SetGenre(genre string)
}

// addGenreSystem configures genre on sys and registers it with world in a
// single call, eliminating the repeated two-line boilerplate:
//
//	sys.SetGenre(genreID)
//	world.AddSystem(sys)
func addGenreSystem(world *engine.World, genreID string, sys genreSystem) {
	sys.SetGenre(genreID)
	world.AddSystem(sys)
}
