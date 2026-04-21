package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/rendering/ui"
)

// storyJournalWrapper wraps *ui.StoryJournalUI and implements engine.StoryJournalProvider.
// This bridges the rendering/ui package (which imports engine) to EbitenGame without
// creating a circular import.
type storyJournalWrapper struct {
	inner   *ui.StoryJournalUI
	visible bool
}

// Toggle shows or hides the journal, refreshing data from the player's journal component.
func (w *storyJournalWrapper) Toggle(journal *engine.StoryJournalComponent, world *engine.World) {
	if w.visible {
		w.visible = false
		return
	}
	if journal != nil {
		w.inner.LoadFromJournal(journal, world)
	}
	w.visible = true
}

// IsVisible reports whether the journal is currently shown.
func (w *storyJournalWrapper) IsVisible() bool { return w.visible }

// Draw renders the journal. screen must be *ebiten.Image; mismatched types are silently ignored.
func (w *storyJournalWrapper) Draw(screen interface{}) {
	ebitenScreen, ok := screen.(*ebiten.Image)
	if !ok || ebitenScreen == nil {
		return
	}
	img := w.inner.Render()
	if img == nil {
		return
	}
	jTex := ebiten.NewImageFromImage(img)
	op := &ebiten.DrawImageOptions{}
	ebitenScreen.DrawImage(jTex, op)
}
