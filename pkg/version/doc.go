// Package version provides centralized version information for Venture.
//
// This package contains the application version constants used throughout
// the project. It provides a single source of truth for version information
// that appears in logs, UI, and documentation.
//
// # Version Format
//
// Venture uses semantic versioning (MAJOR.MINOR.PATCH) with a release status:
//   - Version: semantic version string (e.g., "1.2.3")
//   - Release: release status (e.g., "Production", "Beta", "Alpha")
//   - FullVersion: combined version string (e.g., "1.2.3 Production")
//
// # Usage
//
// Import and use the version constants in startup messages:
//
//	import "github.com/opd-ai/venture/pkg/version"
//
//	logger.Infof("Starting Venture %s", version.FullVersion)
package version
