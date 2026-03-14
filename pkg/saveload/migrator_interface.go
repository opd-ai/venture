// Package saveload provides the Migrator interface shared across all build targets.
// This file has no build constraints so that both desktop and WASM builds use
// the same Migrator definition, preventing drift between the two copies.
package saveload

// Migrator handles save file version migrations.
// Implementations transform older save formats to the current version.
type Migrator interface {
	// CanMigrate returns true if the migrator can handle the given source version.
	CanMigrate(sourceVersion string) bool

	// Migrate transforms a save from sourceVersion to the current SaveVersion.
	// Returns the migrated save or an error if migration fails.
	Migrate(save *GameSave, sourceVersion string) (*GameSave, error)

	// SupportedVersions returns the list of versions this migrator can upgrade from.
	SupportedVersions() []string
}
