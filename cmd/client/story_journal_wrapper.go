package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/rendering/ui"
)

// storyJournalWrapper wraps *ui.StoryJournalUI and implements engine.StoryJournalProvider.
// This bridges the rendering/ui package (which imports engine) to EbitenGame without
// creating a circular import.
type storyJournalWrapper struct {
	inner      *ui.StoryJournalUI
	visible    bool
	cachedTex  *ebiten.Image // GPU-resident texture for the last rendered frame
	cachedRGBA *image.RGBA   // Raw pixels corresponding to cachedTex (identity comparison)
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
	w.cachedTex = nil  // Invalidate cache so the next Draw re-renders
	w.cachedRGBA = nil
	w.visible = true
}

// IsVisible reports whether the journal is currently shown.
func (w *storyJournalWrapper) IsVisible() bool { return w.visible }

// Draw renders the journal. screen must be *ebiten.Image; mismatched types are silently ignored.
// The rendered texture is cached so repeated calls within the same frame or across frames
// where content did not change avoid repeated CPU→GPU uploads.
func (w *storyJournalWrapper) Draw(screen interface{}) {
	ebitenScreen, ok := screen.(*ebiten.Image)
	if !ok || ebitenScreen == nil {
		return
	}
	img := w.inner.Render()
	if img == nil {
		return
	}
	// Only upload a new texture when the pixel buffer pointer changes (i.e. the
	// journal has produced a new frame).
	if img != w.cachedRGBA || w.cachedTex == nil {
		w.cachedTex = ebiten.NewImageFromImage(img)
		w.cachedRGBA = img
	}
	op := &ebiten.DrawImageOptions{}
	ebitenScreen.DrawImage(w.cachedTex, op)
}
