// File: constants.go
// Purpose: Centralized constants for all building-related enums
//
// This file contains all constant definitions for the building package,
// extracted from types.go during reorganization for improved navigability.
// All enum constants are grouped by their respective types.
package building

// Building type constants
// Originally defined in: types.go
const (
	TypeHouse BuildingType = iota
	TypeWorkshop
	TypeStorage
	TypeTower
	TypeManor
	TypeGuildHall
)

// Architectural style constants - Fantasy styles
// Originally defined in: types.go
const (
	StyleMedieval ArchitecturalStyle = iota
	StyleElven
	StyleDwarven
	StyleWizardTower
	StyleVillage

	// Sci-Fi styles
	StyleModular
	StyleBrutalist
	StyleOrganic
	StyleGeometric
	StyleCrystalline

	// Horror styles
	StyleGothic
	StyleDecayed
	StyleAsylum
	StyleMansion
	StyleCrypt

	// Cyberpunk styles
	StyleNeon
	StyleIndustrial
	StyleCorporate
	StyleUnderground
	StyleMegastructure

	// Post-Apocalyptic styles
	StyleSalvage
	StyleBunker
	StyleRuins
	StyleFortified
	StyleScrapyard
)

// Room type constants
// Originally defined in: types.go
const (
	RoomEntrance RoomType = iota
	RoomLiving
	RoomBedroom
	RoomStorage
	RoomWorkshop
	RoomKitchen
	RoomHallway
	RoomTower
)

// Door type constants
// Originally defined in: types.go
const (
	DoorWooden DoorType = iota
	DoorMetal
	DoorGlass
	DoorSecret
)

// Window type constants
// Originally defined in: types.go
const (
	WindowSmall WindowType = iota
	WindowLarge
	WindowStained
	WindowBroken
)

// Roof type constants
// Originally defined in: types.go
const (
	RoofFlat RoofType = iota
	RoofGabled
	RoofHipped
	RoofDomed
	RoofSpire
)
