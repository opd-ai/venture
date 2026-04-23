package audit

import (
	"github.com/opd-ai/venture/pkg/procgen"

	// Blank imports trigger the init() registration in each generator package.
	// To add a new generator to audit tests, add audit_entry.go in the generator
	// package and add a blank import here.
	_ "github.com/opd-ai/venture/pkg/procgen/book"
	_ "github.com/opd-ai/venture/pkg/procgen/building"
	_ "github.com/opd-ai/venture/pkg/procgen/companion"
	_ "github.com/opd-ai/venture/pkg/procgen/entity"
	_ "github.com/opd-ai/venture/pkg/procgen/furniture"
	_ "github.com/opd-ai/venture/pkg/procgen/item"
	_ "github.com/opd-ai/venture/pkg/procgen/legendary"
	_ "github.com/opd-ai/venture/pkg/procgen/magic"
	_ "github.com/opd-ai/venture/pkg/procgen/quest"
	_ "github.com/opd-ai/venture/pkg/procgen/recipe"
	_ "github.com/opd-ai/venture/pkg/procgen/skills"
	_ "github.com/opd-ai/venture/pkg/procgen/station"
	_ "github.com/opd-ai/venture/pkg/procgen/terrain"
	_ "github.com/opd-ai/venture/pkg/procgen/vehicle"
)

// GeneratorEntry is an alias for procgen.AuditEntry for backwards compatibility.
// Each generator package self-registers via init(); this function returns all
// registered entries. Add new generators by creating audit_entry.go in the
// generator package and adding a blank import above.
type GeneratorEntry = procgen.AuditEntry

// GetAllGenerators returns all registered audit generator entries.
// The list is populated by init() functions in each generator package
// (see each package's audit_entry.go).
func GetAllGenerators() []GeneratorEntry {
	return procgen.GetAuditEntries()
}
