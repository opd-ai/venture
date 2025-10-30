//go:build android || ios
// +build android ios

// Package main provides a stub for mobile platforms.
// For mobile platforms (Android/iOS), use cmd/mobile with ebitenmobile build tool instead.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "ERROR: cmd/client is not supported on mobile platforms")
	fmt.Fprintln(os.Stderr, "For Android/iOS builds, use cmd/mobile with ebitenmobile:")
	fmt.Fprintln(os.Stderr, "  ebitenmobile bind -target android -o mobile.aar ./cmd/mobile")
	fmt.Fprintln(os.Stderr, "  ebitenmobile bind -target ios -o Mobile.xcframework ./cmd/mobile")
	os.Exit(1)
}
