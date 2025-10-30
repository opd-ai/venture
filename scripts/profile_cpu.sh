#!/bin/bash
# CPU Profiling Script for Venture Engine
# Based on PLAN.md Section 1.1

set -e

PROF_DIR="docs/profiling"
mkdir -p "$PROF_DIR"

echo "=== Venture CPU Profiling ==="
echo "Starting comprehensive CPU profiling..."
echo ""

# 1. Profile ECS and Entity Queries
echo "[1/5] Profiling ECS system and entity queries..."
go test ./pkg/engine -bench="BenchmarkGetEntities|BenchmarkComponentAccess" \
    -cpuprofile="$PROF_DIR/cpu_ecs.prof" \
    -benchtime=3s \
    -run=^$ 2>&1 | tee "$PROF_DIR/cpu_ecs_bench.txt"

# 2. Profile Rendering Pipeline
echo ""
echo "[2/5] Profiling rendering pipeline..."
go test ./pkg/engine -bench="BenchmarkRender|BenchmarkDraw" \
    -cpuprofile="$PROF_DIR/cpu_render.prof" \
    -benchtime=3s \
    -run=^$ 2>&1 | tee "$PROF_DIR/cpu_render_bench.txt"

# 3. Profile Collision System
echo ""
echo "[3/5] Profiling collision system..."
go test ./pkg/engine -bench="BenchmarkCollision|BenchmarkSpatial" \
    -cpuprofile="$PROF_DIR/cpu_collision.prof" \
    -benchtime=3s \
    -run=^$ 2>&1 | tee "$PROF_DIR/cpu_collision_bench.txt"

# 4. Profile AI System
echo ""
echo "[4/5] Profiling AI system..."
go test ./pkg/engine -bench="BenchmarkAI" \
    -cpuprofile="$PROF_DIR/cpu_ai.prof" \
    -benchtime=3s \
    -run=^$ 2>&1 | tee "$PROF_DIR/cpu_ai_bench.txt"

# 5. Profile Animation System
echo ""
echo "[5/5] Profiling animation system..."
go test ./pkg/engine -bench="BenchmarkAnimation" \
    -cpuprofile="$PROF_DIR/cpu_animation.prof" \
    -benchtime=3s \
    -run=^$ 2>&1 | tee "$PROF_DIR/cpu_animation_bench.txt"

# 6. Profile Generation Systems
echo ""
echo "[6/6] Profiling generation systems..."
go test ./pkg/procgen/... -bench=. \
    -cpuprofile="$PROF_DIR/cpu_procgen.prof" \
    -benchtime=2s \
    -run=^$ 2>&1 | tee "$PROF_DIR/cpu_procgen_bench.txt"

echo ""
echo "=== CPU Profiling Complete ==="
echo "Profile files saved to: $PROF_DIR/"
echo ""
echo "Analyze with:"
echo "  go tool pprof $PROF_DIR/cpu_ecs.prof"
echo "  go tool pprof $PROF_DIR/cpu_render.prof"
echo "  go tool pprof $PROF_DIR/cpu_collision.prof"
echo "  go tool pprof $PROF_DIR/cpu_ai.prof"
echo "  go tool pprof $PROF_DIR/cpu_animation.prof"
echo "  go tool pprof $PROF_DIR/cpu_procgen.prof"
echo ""
echo "Interactive commands:"
echo "  (pprof) top20        # Top 20 functions by CPU time"
echo "  (pprof) top20 -cum   # Top 20 by cumulative time"
echo "  (pprof) list FuncName # Annotated source"
echo "  (pprof) web          # Call graph (requires graphviz)"
