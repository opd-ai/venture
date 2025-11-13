package engine

import "math"

// SimpleAnimationSequence implements AnimationSequence for expression animations.
// Provides basic animation properties for procedural expression effects.
type SimpleAnimationSequence struct {
	frameCount int
	frameTime  float64
	loop       bool
}

// NewSimpleAnimationSequence creates a new animation sequence with specified parameters.
func NewSimpleAnimationSequence(frameCount int, frameTime float64, loop bool) *SimpleAnimationSequence {
	return &SimpleAnimationSequence{
		frameCount: frameCount,
		frameTime:  frameTime,
		loop:       loop,
	}
}

// GetFrameCount returns the number of animation frames.
func (s *SimpleAnimationSequence) GetFrameCount() int {
	return s.frameCount
}

// GetFrameTime returns the duration of each frame in seconds.
func (s *SimpleAnimationSequence) GetFrameTime() float64 {
	return s.frameTime
}

// ShouldLoop returns whether the animation repeats.
func (s *SimpleAnimationSequence) ShouldLoop() bool {
	return s.loop
}

// BaseExpression implements Expression interface for standard player emotes.
// Each expression type has specific animation and sound properties.
type BaseExpression struct {
	expressionType ExpressionType
	animation      AnimationSequence
	soundEffect    string
	duration       float64
}

// NewBaseExpression creates a new expression with the given type.
// Animation and sound are configured based on the expression type.
func NewBaseExpression(expressionType ExpressionType) *BaseExpression {
	expr := &BaseExpression{
		expressionType: expressionType,
	}

	// Configure animation and sound based on type
	switch expressionType {
	case ExpressionWave:
		expr.animation = NewSimpleAnimationSequence(8, 0.1, false)
		expr.soundEffect = "wave"
		expr.duration = 3.0

	case ExpressionCheer:
		expr.animation = NewSimpleAnimationSequence(10, 0.08, false)
		expr.soundEffect = "cheer"
		expr.duration = 3.0

	case ExpressionDance:
		expr.animation = NewSimpleAnimationSequence(16, 0.15, true)
		expr.soundEffect = "music_note"
		expr.duration = 5.0

	case ExpressionLaugh:
		expr.animation = NewSimpleAnimationSequence(6, 0.12, true)
		expr.soundEffect = "laugh"
		expr.duration = 3.0

	case ExpressionCry:
		expr.animation = NewSimpleAnimationSequence(8, 0.2, true)
		expr.soundEffect = "sob"
		expr.duration = 3.0

	case ExpressionSit:
		expr.animation = NewSimpleAnimationSequence(4, 0.15, false)
		expr.soundEffect = ""
		expr.duration = math.Inf(1) // Lasts until canceled

	case ExpressionPoint:
		expr.animation = NewSimpleAnimationSequence(4, 0.1, false)
		expr.soundEffect = ""
		expr.duration = 2.0

	case ExpressionSalute:
		expr.animation = NewSimpleAnimationSequence(6, 0.12, false)
		expr.soundEffect = ""
		expr.duration = 2.5

	case ExpressionShrug:
		expr.animation = NewSimpleAnimationSequence(5, 0.15, false)
		expr.soundEffect = ""
		expr.duration = 2.0

	case ExpressionThumbsUp:
		expr.animation = NewSimpleAnimationSequence(4, 0.1, false)
		expr.soundEffect = "approval"
		expr.duration = 2.0

	case ExpressionFacepalm:
		expr.animation = NewSimpleAnimationSequence(6, 0.15, false)
		expr.soundEffect = "thud"
		expr.duration = 3.0

	case ExpressionSleep:
		expr.animation = NewSimpleAnimationSequence(8, 0.25, true)
		expr.soundEffect = "snore"
		expr.duration = math.Inf(1) // Lasts until canceled

	default:
		// Default animation
		expr.animation = NewSimpleAnimationSequence(4, 0.15, false)
		expr.soundEffect = ""
		expr.duration = 2.0
	}

	return expr
}

// GetAnimation returns the procedural animation sequence for this expression.
func (e *BaseExpression) GetAnimation() AnimationSequence {
	return e.animation
}

// GetSoundEffect returns the audio effect ID to play when expression triggers.
func (e *BaseExpression) GetSoundEffect() string {
	return e.soundEffect
}

// GetDuration returns how long the expression animation lasts in seconds.
func (e *BaseExpression) GetDuration() float64 {
	return e.duration
}
