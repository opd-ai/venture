// Package puzzle provides procedural puzzle generation for Phase 11.2.
// This package contains the constraint solver and puzzle generator.
package puzzle

// Package documentation for puzzle generation.
//
// The puzzle generation system creates solvable constraint-based puzzles
// using a Constraint Satisfaction Problem (CSP) approach with backtracking search.
//
// Supported puzzle types:
// - Pressure Plate: Step on plates in correct sequence
// - Lever Sequence: Activate levers in correct order
// - Block Pushing: Push blocks to target positions
// - Timed Challenge: Complete within time limit
// - Memory Pattern: Repeat shown pattern
// - Color Matching: Match colors/symbols
//
// All puzzles are generated deterministically from a seed and guarantee solvability
// through constraint solving and validation.
