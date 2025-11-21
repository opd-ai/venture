// Package procgen provides procedural generation utilities.
// This file implements deterministic character name selection.
package procgen

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetReportCaller(true)
	log.WithFields(logrus.Fields{
		"package":    "procgen",
		"subsystem":  "naming",
		"name_count": len(DefaultNames),
	}).Debug("Naming system initialized")
}

// DefaultNames is a list of 100 culturally diverse, inoffensive character names.
// These names are used for deterministic default character naming based on world seed.
var DefaultNames = [100]string{
	// Western names (20)
	"Alexander", "Amelia", "Benjamin", "Charlotte", "Daniel", "Elizabeth", "Frederick", "Grace",
	"Henry", "Isabella", "James", "Katherine", "Lucas", "Margaret", "Nicholas", "Olivia",
	"Patrick", "Rachel", "Samuel", "Sophia",

	// Eastern names (16)
	"Akira", "Mei", "Hiroshi", "Yuki", "Chen", "Lin", "Ryu", "Sakura",
	"Jin", "Hana", "Kenji", "Aiko", "Wei", "Feng", "Taro", "Emi",

	// Middle Eastern names (16)
	"Ali", "Fatima", "Hassan", "Layla", "Omar", "Zara", "Malik", "Amira",
	"Khalid", "Nora", "Tariq", "Salma", "Rashid", "Yasmin", "Karim", "Leila",

	// African names (16)
	"Kwame", "Ama", "Kofi", "Abena", "Nia", "Jabari", "Zuri", "Kendi",
	"Amani", "Bakari", "Imani", "Jafari", "Makena", "Thabo", "Zahara", "Akil",

	// Latin American names (16)
	"Diego", "Sofia", "Miguel", "Lucia", "Carlos", "Carmen", "Pablo", "Elena",
	"Rafael", "Valentina", "Marco", "Gabriela", "Antonio", "Marina", "Roberto", "Rosa",

	// Celtic/Nordic names (16)
	"Alistair", "Fiona", "Magnus", "Astrid", "Ronan", "Freya", "Erik", "Ingrid",
	"Finn", "Brenna", "Bjorn", "Signe", "Declan", "Moira", "Soren", "Isla",
}

// SelectDefaultName deterministically selects a default character name based on the world seed.
// The same seed will always return the same name, ensuring consistency across sessions.
//
// Parameters:
//   - seed: The world generation seed used for deterministic selection
//
// Returns:
//   - A name from the DefaultNames list, selected deterministically based on the seed
//
// Example:
//
//	name := SelectDefaultName(12345) // Always returns the same name for seed 12345
func SelectDefaultName(seed int64) string {
	log.WithFields(logrus.Fields{
		"seed":            seed,
		"operation":       "select_default_name",
		"available_names": len(DefaultNames),
	}).Debug("SelectDefaultName function entry")

	// Create a seeded random number generator for deterministic selection
	rng := rand.New(rand.NewSource(seed))
	log.WithFields(logrus.Fields{
		"seed": seed,
	}).Debug("Created seeded RNG for deterministic name selection")

	// Select an index from the name list
	index := rng.Intn(len(DefaultNames))
	log.WithFields(logrus.Fields{
		"seed":      seed,
		"index":     index,
		"max_index": len(DefaultNames) - 1,
	}).Debug("Generated random index for name selection")

	selectedName := DefaultNames[index]
	log.WithFields(logrus.Fields{
		"seed":          seed,
		"index":         index,
		"selected_name": selectedName,
		"operation":     "select_default_name",
	}).Debug("SelectDefaultName function exit")

	return selectedName
}
