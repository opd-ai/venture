package engine

// ExpressionType represents emotes and gestures
type ExpressionType int

const (
	ExpressionWave ExpressionType = iota
	ExpressionCheer
	ExpressionDance
	ExpressionLaugh
	ExpressionCry
	ExpressionSit
	ExpressionPoint
	ExpressionSalute
	ExpressionShrug
	ExpressionThumbsUp
	ExpressionFacepalm
	ExpressionSleep
)

// ExpressionComponent tracks active expressions
type ExpressionComponent struct {
	ActiveExpression ExpressionType
	ExpressionTime   float64 // Time remaining
	Cooldown         float64 // Prevent spam
}

// Type returns the component type
func (e ExpressionComponent) Type() string {
	return "expression"
}

// String returns the expression name
func (e ExpressionType) String() string {
	names := []string{
		"Wave", "Cheer", "Dance", "Laugh", "Cry", "Sit",
		"Point", "Salute", "Shrug", "ThumbsUp", "Facepalm", "Sleep",
	}
	if int(e) < len(names) {
		return names[e]
	}
	return "Unknown"
}
