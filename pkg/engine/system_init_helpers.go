// Package engine system_init_helpers.go provides helper functions that reduce
// boilerplate in InitializeGameSystems.
//
// These helpers are intentionally unexported and kept in pkg/engine to avoid
// any import-cycle hazard.
package engine

// genreConfigurer is the minimal interface required by regGenre. It is
// satisfied by every system that exposes a SetGenre method.
//
// Note: cmd/client defines an identical unexported interface (genreSystem) for
// the same purpose. Both remain separate because pkg/engine cannot import
// cmd/client and vice-versa; a shared location would require a new library
// package. Add one only when a third call-site appears.
type genreConfigurer interface {
	System
	SetGenre(genre string)
}

// regGenre configures genre on s, registers it with world, and returns s.
// It replaces the repetitive four-line pattern used across InitializeGameSystems:
//
//	s := NewXxx(world, seed)
//	s.SetGenre(genreID)
//	result.Xxx = s
//	world.AddSystem(s)
//
// with a single expression:
//
//	result.Xxx = regGenre(world, NewXxx(world, seed), genreID)
func regGenre[T genreConfigurer](world *World, s T, genreID string) T {
	s.SetGenre(genreID)
	world.AddSystem(s)
	return s
}
