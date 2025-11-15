# Development Roadmap - Version 7.0: Advanced Visual Improvements

## Current Status

**Status:** PLANNING - Not Yet Started  
**Prerequisites:** V6.0 completion  
**Projected Start:** Post-2027  

This is a future planning document. No V7.0 features have been implemented yet.

## Overview

**Version:** 7.0 (Previous: 6.0 - Planning)  
**Focus:** Display scaling, enhanced sprites, smoothed walls, pixel-perfect collision

## Deliverables

- 1920x1080 default resolution (from 1280x720) with scalable UI
- 64x64 sprites with 0.85+ silhouette recognition (from 32x32, 0.75)
- Anti-aliased wall rendering with seamless corner blending
- Sub-pixel collision detection (0.1px precision)

## Targets

**Performance:** 60 FPS at 1920x1080, <500MB memory  
**Quality:** ≥65% test coverage, backward compatible with v6.0  
**Cache:** ≥90% sprite hit rate (current: 95.9%)

## Architecture

**Display Scaling:** `pkg/rendering/display/` (manager, scaler) → Viewport/UI scaling → Font rendering  
**Sprites:** Existing Phase 15.1 templates (anatomy, silhouette, composite) + 64x64 size + genre variations  
**Walls:** `pkg/rendering/tiles/` + patterns for anti-aliasing + corner blending (L/T/cross joints)  
**Collision:** `pkg/engine/collision.go` + sub-pixel precision + quadtree optimization

## Phase 43: Display Foundation

**Deliverables:**
- [ ] Create `pkg/rendering/display/` (manager, scaler, config)
- [ ] Dynamic resolution detection, CLI flags (`-width`, `-height`, `-fullscreen`)
- [ ] Support 1280x720, 1920x1080, 2560x1440, 3840x2160
- [ ] UI scaling in `pkg/rendering/ui/scaler.go` (fonts, buttons, panels)
- [ ] Update all 7 menus (inventory, character, skills, quests, map, crafting, shop)

**Metrics:** All resolutions render without distortion, <50ms resolution switch, <100ms UI init per menu

---

## Phase 44: Viewport Optimization

**Deliverables:**
- [ ] Update viewport culling for 2.25x larger screen (1920x1080)
- [ ] Enhance quadtree queries, frustum culling with 1-tile margin
- [ ] Sprite cache for 64x64 sprites with LRU eviction (max 1000 sprites, <300MB)
- [ ] Memory monitoring and batch pre-generation

**Metrics:** <5% entities rendered off-screen, <0.1ms quadtree queries, ≥90% cache hit rate

---

## Phase 45: Enhanced Sprites

**Deliverables:**
- [ ] 64x64 sprite templates (head 12%, torso 40%, legs 48%)
- [ ] Secondary details (shoulders, neck, feet), genre variations
- [ ] Multi-layer composition (skin → clothing → armor), shading gradients
- [ ] Genre palettes and texture patterns

**Metrics:** 0.85+ silhouette score, <2ms generation time, 90%+ genre recognition

---

## Phase 46: Animation Fluidity

**Deliverables:**
- [ ] 8-frame animations (from 4), 8-direction support
- [ ] Body part articulation (arms ±3px, legs ±4px)
- [ ] Animation caching and sprite pooling
- [ ] Pre-compute common animations

**Metrics:** 8 frames/cycle, ≥85% cache hit rate, <1ms per frame, 60 FPS with 100 animated entities

---

## Phase 47: Wall Rendering

**Deliverables:**
- [ ] Edge anti-aliasing with 2x2 super-sampling
- [ ] Wall/floor boundary blending (50/50 color)
- [ ] Corner detection (L/T/cross joints) and 4px blending radius
- [ ] Diagonal wall support, shadow integration

**Metrics:** No jagged pixels at 1920x1080, <0.5ms per tile, seamless corners

---

## Phase 48: Pixel-Perfect Collision

**Deliverables:**
- [ ] 0.1-pixel precision (Float64 positions)
- [ ] Smooth wall sliding with edge normals
- [ ] Collision shapes for rounded corners
- [ ] Debug mode (`-debug-collision` flag)

**Metrics:** ±0.1px accuracy, <0.5px visual/collision alignment, 60 FPS with 2000 entities

## Success Criteria

**Visual Quality:**
- [ ] Silhouette recognition ≥0.85, genre ID 90%+, 8 frames/cycle at 60 FPS
- [ ] No jagged pixels, seamless corners, proper shadow casting
- [ ] Text legible at all resolutions, consistent proportions, no clipping

**Performance:**
- [ ] 60 FPS at 1920x1080 (2000 entities), 30 FPS at 3840x2160 (1000 entities)
- [ ] <500MB total, <300MB sprite cache (≥90% hit), <50MB animation cache
- [ ] <50ms resolution change, <2s cache warm-up, <100ms UI init

**Compatibility:**
- [ ] v6.0 saves load, legacy resolutions work, multiplayer sync across resolutions
- [ ] Desktop (Linux/macOS/Windows), WebAssembly, mobile (iOS/Android)
- [ ] Same seed = identical sprites at all resolutions

---

## Risk Mitigation

**Performance:** Early viewport culling, CPU/GPU profiling, quality settings (Low/Med/High), auto-reduce if FPS <60  
**Memory:** LRU eviction, monitoring + trimming, lazy generation, LOD for distant entities  
**Visual Quality:** User testing at week 11, A/B comparison, iterate on templates, keep 32x32 fallback  
**API Changes:** Maintain v6.0 compatibility via wrappers, version methods (GenerateV6/V7), gradual migration

---

## Dependencies

**Existing:** Phase 3 (culling/batching), Phase 15 (sprites/silhouette), Phase 17 (lighting), Phase 11 (diagonal walls), quadtree  
**New:** `pkg/rendering/display/`, `pkg/rendering/ui/scaler.go`, enhanced anatomy/tiles/collision

**Build:** Go 1.24.5+, Ebiten v2.9.3+, 16GB RAM, 1920x1080 display  
**Testing:** Visual regression, profiling (`-cpuprofile`/`-memprofile`), multi-resolution suite, silhouette scoring in CI

---


## Completion Checklist

**Features:** [ ] 1920x1080 scalable UI [ ] 64x64 sprites (≥0.85) [ ] Smoothed walls [ ] Sub-pixel collision [ ] 60 FPS, <500MB  
**Quality:** [ ] ≥65% coverage [ ] Tests pass [ ] Cross-platform [ ] v6.0 compatible [ ] Multiplayer sync  
**Docs:** [ ] Update ARCHITECTURE/TECHNICAL_SPEC/USER_MANUAL/API_REFERENCE [ ] Create RELEASE_NOTES_V7.0.md  
**Release:** [ ] Version bump 6.0→7.0 [ ] Build all platforms [ ] WebAssembly deploy [ ] Mobile builds [ ] Tag v7.0.0

---

**Document Status:** Draft  
**Last Updated:** November 14, 2025  
**Next Review:** Upon Phase 36 completion (V6.0 release)
