// Package version provides centralized version information for Venture.
//
// This package contains the application version constants used throughout
// the project. It provides a single source of truth for version information
// that appears in logs, UI, and documentation.
//
// # Version Format
//
// Venture uses semantic versioning (MAJOR.MINOR.PATCH) with a release status:
//   - Version: "3.0.0"
//   - Release: "Production" (or "Beta", "Alpha")
//   - FullVersion: "3.0.0 Production"
//
// # Usage
//
// Import and use the version constants in startup messages:
//
//	import "github.com/opd-ai/venture/pkg/version"
//
//	logger.Infof("Starting Venture %s", version.FullVersion)
package version
