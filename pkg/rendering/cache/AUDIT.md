# Audit: pkg/rendering/cache

**Date**: 2026-02-16
**Auditor**: Copilot
**Coverage**: Unable to measure (Ebiten/GLFW requires display)

## Summary

The cache package provides LRU sprite caching with memory monitoring, predictive warming, and batch pre-generation. Three issues were found and fixed.

## Files Reviewed

| File | Lines | Purpose |
|------|-------|---------|
| sprite_cache.go | 289 | LRU cache for Ebiten images with size-based eviction |
| memory_monitor.go | 188 | Background memory monitoring with soft/hard limits and cleanup triggers |
| predictive_warmer.go | 321 | Access pattern analysis for proactive cache warming |
| pregenerator.go | 190 | Batch sprite pre-generation queue |
| doc.go | — | Package documentation |

## Issues Found

### Fixed

| # | Severity | File | Issue | Fix |
|---|----------|------|-------|-----|
| 1 | **High** | predictive_warmer.go | `Stats()` called `PredictNext()` which re-acquires `RLock`, causing potential deadlock when a writer is waiting | Extracted lock-free `predictNextLocked()` used by both `Stats()` and `PredictNext()` |
| 2 | **Medium** | memory_monitor.go | `Stop()` panics if called twice (double close of channel) | Added `sync.Once` to guard channel close |
| 3 | **Medium** | memory_monitor.go | `SetLimits()` accepted softLimit > hardLimit, which would cause the soft cleanup to never trigger correctly | Added validation: clamp softLimit to hardLimit |

### Remaining (Low)

| # | Severity | File | Issue |
|---|----------|------|-------|
| 1 | Low | predictive_warmer.go | `patterns` map grows unbounded over session lifetime; no cleanup mechanism |
| 2 | Low | memory_monitor.go | `interval` field changes via `SetInterval()` don't take effect until monitor restart |

## Verdict

**1 High, 2 Medium fixed; 2 Low remaining.** Core cache logic (LRU, eviction, pre-generation) is well-implemented with proper mutex usage.
