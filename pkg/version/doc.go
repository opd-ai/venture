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
// # Protocol Versioning
//
// The package also provides protocol version constants for network federation:
//   - ProtocolVersion: federation protocol version (e.g., "6.0.0")
//   - ProtocolMajor, ProtocolMinor, ProtocolPatch: individual components
//
// Protocol versions follow semantic versioning independently of the application
// version. Two protocol versions are compatible if they share the same major
// version number. Use IsCompatible() for version negotiation.
//
// # Version Comparison
//
// The package provides utilities for parsing and comparing semantic versions:
//   - ParseVersion: parse "1.2.3" into (1, 2, 3) components
//   - Compare: compare two versions (-1, 0, 1)
//   - IsCompatible: check if major versions match (for protocol negotiation)
//
// # Usage
//
// Import and use the version constants in startup messages:
//
//	import "github.com/opd-ai/venture/pkg/version"
//
//	logger.Infof("Starting Venture %s", version.FullVersion)
//
// For network protocol negotiation:
//
//	compatible, err := version.IsCompatible(ourVersion, theirVersion)
//	if err != nil {
//	    logger.WithError(err).Error("invalid version format")
//	    return
//	}
//	if compatible {
//	    // Versions are compatible, proceed with connection
//	}
package version
