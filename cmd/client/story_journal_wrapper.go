package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/rendering/ui"
)

// storyJournalWrapper wraps *ui.StoryJournalUI and implements engine.StoryJournalProvider.
// This bridges the rendering/ui package (which imports engine) to EbitenGame without
// creating a circular import.
//
// Rendering strategy: Render() on *ui.StoryJournalUI allocates a fresh *image.RGBA
// on every call, so pointer-equality caching would never hit. Instead, we track a
// dirty flag that is set whenever the journal is first shown (Toggle) or when
// MarkDirty is called externally (e.g., after navigation key input). Draw only
// re-uploads to the GPU when dirty is true, keeping per-frame cost to a blitI.
type storyJournalWrapper struct {
	inner     *ui.StoryJournalUI
	visible   bool
	cachedTex *ebiten.Image // GPU-resident texture; nil until first render
	dirty     bool          // true when cachedTex must be rebuilt before the next blit
}

// Toggle shows or hides the journal, refreshing data from the player's journal component.
// Marks the wrapper dirty so Draw re-renders on the next call.
func (w *storyJournalWrapper) Toggle(journal *engine.StoryJournalComponent, world *engine.World) {
	if w.visible {
		w.visible = false
		return
	}
	if journal != nil {
		w.inner.LoadFromJournal(journal, world)
	}
	w.dirty = true
	w.visible = true
}

// IsVisible reports whether the journal is currently shown.
func (w *storyJournalWrapper) IsVisible() bool { return w.visible }

// MarkDirty signals that the journal content has changed and the cached texture
// must be rebuilt on the next Draw call. Call this after navigation key events
// (NavigateUp, NavigateDown, Back) to keep the display in sync.
func (w *storyJournalWrapper) MarkDirty() { w.dirty = true }

// Draw renders the journal. screen must be *ebiten.Image; mismatched types are silently ignored.
// Only re-renders from CPU when dirty is true; otherwise blits the cached texture, keeping
// the GPU upload cost out of the steady-state frame loop.
func (w *storyJournalWrapper) Draw(screen interface{}) {
	ebitenScreen, ok := screen.(*ebiten.Image)
	if !ok || ebitenScreen == nil {
		return
	}
	if w.dirty || w.cachedTex == nil {
		img := w.inner.Render()
		if img == nil {
			return
		}
		w.cachedTex = ebiten.NewImageFromImage(img)
		w.dirty = false
	}
	op := &ebiten.DrawImageOptions{}
	ebitenScreen.DrawImage(w.cachedTex, op)
}
