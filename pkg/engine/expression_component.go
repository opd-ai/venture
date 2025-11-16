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

// Serialize converts ExpressionComponent to bytes for network transmission.
// Format: [ActiveExpression:1][ExpressionTime:8][Cooldown:8] = 17 bytes
func (e *ExpressionComponent) Serialize() []byte {
	buf := make([]byte, 17)
	buf[0] = byte(e.ActiveExpression)
	writeFloat64(buf[1:9], e.ExpressionTime)
	writeFloat64(buf[9:17], e.Cooldown)
	return buf
}

// Deserialize reads ExpressionComponent from bytes.
func (e *ExpressionComponent) Deserialize(data []byte) error {
	if len(data) < 17 {
		return ErrInvalidComponentData
	}
	e.ActiveExpression = ExpressionType(data[0])
	e.ExpressionTime = readFloat64(data[1:9])
	e.Cooldown = readFloat64(data[9:17])
	return nil
}
