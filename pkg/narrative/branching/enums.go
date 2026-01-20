// Package branching provides branching narrative enumeration types.
// This file defines all enumeration types used for narrative classification.
// Originally from: types.go
package branching

// NodeType represents the type of narrative node
type NodeType int

const (
	NodeTypeStart NodeType = iota
	NodeTypeChoice
	NodeTypeEvent
	NodeTypeConsequence
	NodeTypeEnding
)

// String returns the string representation of a node type.
func (n NodeType) String() string {
	switch n {
	case NodeTypeStart:
		return "Start"
	case NodeTypeChoice:
		return "Choice"
	case NodeTypeEvent:
		return "Event"
	case NodeTypeConsequence:
		return "Consequence"
	case NodeTypeEnding:
		return "Ending"
	default:
		return "Unknown"
	}
}

// EndingType represents the type of story ending
type EndingType int

const (
	EndingTypeHeroic EndingType = iota
	EndingTypeTragic
	EndingTypeNeutral
	EndingTypeMystery
	EndingTypeTriumph
	EndingTypeBetrayal
)

// String returns the string representation of an ending type.
func (e EndingType) String() string {
	switch e {
	case EndingTypeHeroic:
		return "Heroic"
	case EndingTypeTragic:
		return "Tragic"
	case EndingTypeNeutral:
		return "Neutral"
	case EndingTypeMystery:
		return "Mystery"
	case EndingTypeTriumph:
		return "Triumph"
	case EndingTypeBetrayal:
		return "Betrayal"
	default:
		return "Unknown"
	}
}

// AlignmentAxis represents moral alignment axes
type AlignmentAxis string

const (
	AlignmentGoodEvil      AlignmentAxis = "good_evil"
	AlignmentLawChaos      AlignmentAxis = "law_chaos"
	AlignmentHonorDishonor AlignmentAxis = "honor_dishonor"
)
