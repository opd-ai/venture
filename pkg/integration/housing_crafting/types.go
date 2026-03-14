package housing_crafting

// Shared type definitions for the housing_crafting package.
// This file contains StationType and QualityTier enumerations used
// throughout the package for categorizing crafting stations and
// determining their effectiveness multipliers.

// StationType represents the type of crafting station
type StationType int

const (
	StationTypeForge StationType = iota
	StationTypeAlchemy
	StationTypeEnchanting
	StationTypeCooking
	StationTypeTailoring
	StationTypeWoodworking
	StationTypeInscription
	StationTypeEngineering
)

// String returns the human-readable name of the station type.
// Returns "unknown" for invalid or unrecognized values.
func (st StationType) String() string {
	switch st {
	case StationTypeForge:
		return "Forge"
	case StationTypeAlchemy:
		return "Alchemy"
	case StationTypeEnchanting:
		return "Enchanting"
	case StationTypeCooking:
		return "Cooking"
	case StationTypeTailoring:
		return "Tailoring"
	case StationTypeWoodworking:
		return "Woodworking"
	case StationTypeInscription:
		return "Inscription"
	case StationTypeEngineering:
		return "Engineering"
	default:
		return "Unknown"
	}
}

// QualityTier represents the quality level of a crafting station
type QualityTier int

const (
	QualityBasic QualityTier = iota
	QualityStandard
	QualityAdvanced
	QualityMaster
)

// String returns the human-readable name of the quality tier.
// Returns "unknown" for invalid or unrecognized values.
func (qt QualityTier) String() string {
	switch qt {
	case QualityBasic:
		return "Basic"
	case QualityStandard:
		return "Standard"
	case QualityAdvanced:
		return "Advanced"
	case QualityMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

// Multiplier returns the crafting bonus multiplier for this quality tier
func (qt QualityTier) Multiplier() float64 {
	switch qt {
	case QualityBasic:
		return 1.0
	case QualityStandard:
		return 1.2
	case QualityAdvanced:
		return 1.5
	case QualityMaster:
		return 2.0
	default:
		return 1.0
	}
}
