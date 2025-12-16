package engine

import (
	"testing"
)

// BenchmarkVisualFeedbackCached measures performance of cached GetVisualFeedback() getter
func BenchmarkVisualFeedbackCached(b *testing.B) {
	// Create entities with visual feedback component
	entities := make([]*Entity, 1000)
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(NewVisualFeedbackComponent())
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, entity := range entities {
			// Use cached getter (new optimized path)
			feedback := entity.GetVisualFeedback()
			if feedback != nil {
				_ = feedback.GetFlashAlpha()
			}
		}
	}
}

// BenchmarkVisualFeedbackGeneric measures performance of generic GetComponent() path
func BenchmarkVisualFeedbackGeneric(b *testing.B) {
	// Create entities with visual feedback component
	entities := make([]*Entity, 1000)
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(NewVisualFeedbackComponent())
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, entity := range entities {
			// Use generic GetComponent path (old non-optimized path)
			feedbackComp, ok := entity.GetComponent("visual_feedback")
			if ok {
				if feedback, ok := feedbackComp.(*VisualFeedbackComponent); ok {
					_ = feedback.GetFlashAlpha()
				}
			}
		}
	}
}

// BenchmarkExtractVisualFeedback measures the extractVisualFeedback function performance
func BenchmarkExtractVisualFeedback(b *testing.B) {
	// Create camera system
	cameraSys := NewCameraSystem(800, 600)

	// Create render system
	renderSys := NewRenderSystem(cameraSys)

	// Create entities with visual feedback component
	entities := make([]*Entity, 1000)
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(NewVisualFeedbackComponent())
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, entity := range entities {
			_, _, _, _, _ = renderSys.extractVisualFeedback(entity)
		}
	}
}
