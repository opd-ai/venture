//go:build android || ios
// +build android ios

// Package main provides a stub for mobile platforms.
// Servers are not supported on mobile platforms.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "ERROR: cmd/server is not supported on mobile platforms")
	fmt.Fprintln(os.Stderr, "Servers can only run on desktop platforms (Linux, macOS, Windows) or WASM")
	os.Exit(1)
}
