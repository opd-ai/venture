# Runbook: High CPU Usage

**Severity:** P2 (Performance Degradation)  
**Symptoms:** Server CPU usage >80%, slow response times, frame rate drops  
**Owner:** Infrastructure Team  
**Last Updated:** 2026-01-07

---

## Overview

This runbook guides you through diagnosing and resolving high CPU usage in Venture game servers. High CPU can manifest as degraded frame rates, slow player actions, or complete server unresponsiveness.

**Expected State:** CPU usage <40% with 4 players, <60% with 10 players  
**Alert Threshold:** CPU >80% sustained for >5 minutes

---

## Initial Assessment (5 minutes)

### 1. Verify the Issue

Check current CPU usage via metrics endpoint:

```bash
curl http://localhost:8081/status | jq '.runtime.goroutines'
```

Expected: <100 goroutines for healthy server  
Alert if: >500 goroutines (potential goroutine leak)

Check system CPU usage:

```bash
# Linux
top -p $(pgrep venture-server) -n 1 | grep venture

# macOS
top -l 1 -pid $(pgrep venture-server) | grep venture
```

### 2. Check Server Health

```bash
# Verify server is responsive
curl -w '\nResponse Time: %{time_total}s\n' http://localhost:8081/health

# Check readiness
curl http://localhost:8081/ready | jq '.'
```

Expected: <100ms response time  
Alert if: >1s response time or 503 status

### 3. Review Recent Changes

- Check if CPU spike correlates with player count change
- Review deployment history (did new code deploy recently?)
- Check if new mods were enabled

```bash
# Check server uptime and player count
curl http://localhost:8081/status | jq '{uptime: .uptime_seconds, players: .network.connected_players}'
```

---

## Diagnosis (10 minutes)

### Step 1: Profile CPU Usage

Enable CPU profiling if not already active:

```bash
# Send SIGUSR1 to enable profiling (if implemented)
kill -USR1 $(pgrep venture-server)

# Wait 30 seconds for profile data
sleep 30

# Download profile
curl http://localhost:8081/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze with go tool pprof
go tool pprof cpu.prof
# In pprof: type 'top20' to see top CPU consumers
# In pprof: type 'list <function_name>' to see function details
```

**Common Culprits:**
- `rendering.SpriteCache.Get` - Cache thrashing, increase cache size
- `procgen.*Generator.Generate` - Excessive procedural generation, check seed loops
- `network.Server.Update` - Network storm, check packet rate
- `engine.CollisionSystem.Update` - Too many entities, implement spatial partitioning
- `runtime.gcBgMarkWorker` - GC pressure, memory leak suspected (see memory-leak.md)

### Step 2: Check Entity Count

High entity count can cause CPU spikes in collision detection and rendering:

```bash
curl http://localhost:8081/status | jq '.game_state.entity_count'
```

Expected: <2000 entities  
Alert if: >5000 entities (indicates entity leak or generation issue)

**Action:** If entity count is abnormally high, check for:
- Entities not being cleaned up (projectiles, particles)
- Runaway procedural generation loops
- Quest/monster spawning bugs

### Step 3: Check Network Activity

High packet rate can cause CPU spikes:

```bash
# Check packets per second
curl http://localhost:8081/metrics | grep packets_sent_total

# Wait 10 seconds
sleep 10

curl http://localhost:8081/metrics | grep packets_sent_total

# Calculate packets/sec: (second_value - first_value) / 10
```

Expected: <1000 packets/sec per player  
Alert if: >5000 packets/sec (network storm)

**Action:** If packet rate is high:
- Check for broadcast loops in federation
- Verify rate limiting is active
- Look for chat spam or trade spam

### Step 4: Check Goroutine Count

Goroutine leaks can cause CPU spikes:

```bash
curl http://localhost:8081/status | jq '.runtime.goroutines'
```

Expected: <100 goroutines  
Alert if: >500 goroutines (leak suspected)

If goroutines are high, download goroutine stack trace:

```bash
curl http://localhost:8081/debug/pprof/goroutine > goroutine.txt
grep -A 5 "goroutine" goroutine.txt | head -50
```

Look for repeated patterns indicating leaks (e.g., thousands of network receive loops).

---

## Resolution

### Scenario 1: High Entity Count

**Symptom:** entity_count >5000  
**Root Cause:** Entity cleanup failure or generation runaway

```bash
# Restart server to clear entities
systemctl restart venture-server

# Monitor entity count after restart
watch -n 5 'curl -s http://localhost:8081/status | jq .game_state.entity_count'
```

If entity count grows rapidly (>100/sec), investigate procedural generation:
1. Check logs for generation errors: `grep "procgen" /var/log/venture-server.log`
2. Review recent mod installations (mods can override generation)
3. File bug report with reproduction steps

### Scenario 2: Network Storm

**Symptom:** packets_sent >5000/sec  
**Root Cause:** Broadcast loop or spam

```bash
# Enable rate limiting if not active (requires restart)
# Edit /etc/venture/server.conf:
# rate_limit_enabled=true
# rate_limit_packets_per_sec=1000

# Restart server
systemctl restart venture-server

# Verify rate limit is active
grep "rate_limit" /var/log/venture-server.log
```

If caused by federation broadcast loop:
1. Temporarily disable federation: `venture-server -federation=false`
2. Check federation health: See federation-debug.md runbook
3. Re-enable federation after fix confirmation

### Scenario 3: Goroutine Leak

**Symptom:** goroutines >500  
**Root Cause:** Goroutine not terminating (network leak, worker pool leak)

```bash
# Restart server (immediate mitigation)
systemctl restart venture-server

# File bug report with goroutine dump
curl http://localhost:8081/debug/pprof/goroutine > goroutine-$(date +%Y%m%d-%H%M%S).txt
```

**Note:** Goroutine leaks indicate code bugs. Restart is temporary fix. Provide goroutine dump to developers.

### Scenario 4: GC Pressure

**Symptom:** `runtime.gcBgMarkWorker` in top CPU consumers  
**Root Cause:** Memory pressure, see memory-leak.md

```bash
# Increase GOGC to reduce GC frequency (trade CPU for memory)
export GOGC=200  # Default is 100
systemctl restart venture-server

# Monitor memory usage
watch -n 5 'curl -s http://localhost:8081/status | jq .performance.memory_mb'
```

**Warning:** Increasing GOGC reduces CPU but increases memory usage. Monitor both.

### Scenario 5: Procedural Generation Overload

**Symptom:** `procgen.*Generate` in top CPU consumers  
**Root Cause:** Excessive on-demand generation

```bash
# Check generation cache hit rate
grep "cache_hit_rate" /var/log/venture-server.log | tail -20

# If hit rate <80%, increase generation cache size
# Edit /etc/venture/server.conf:
# generation_cache_mb=500  # Default: 100

systemctl restart venture-server
```

---

## Monitoring and Verification (15 minutes)

After applying fixes, monitor for 15 minutes:

```bash
# Monitor CPU, memory, and goroutines every 10 seconds
watch -n 10 'curl -s http://localhost:8081/status | jq "{cpu_goroutines: .runtime.goroutines, memory_mb: .performance.memory_mb, entities: .game_state.entity_count, fps: .performance.fps}"'
```

**Success Criteria:**
- CPU usage <60%
- Goroutines <100
- Entity count stable or growing slowly (<10/min)
- FPS >60

---

## Escalation

If issue persists after all resolution attempts:

1. **Collect Diagnostic Bundle:**
   ```bash
   # Create diagnostic archive
   mkdir venture-diag-$(date +%Y%m%d-%H%M%S)
   cd venture-diag-*
   
   # Collect profiles
   curl http://localhost:8081/debug/pprof/profile?seconds=30 > cpu.prof
   curl http://localhost:8081/debug/pprof/goroutine > goroutine.txt
   curl http://localhost:8081/status > status.json
   
   # Collect logs
   cp /var/log/venture-server.log .
   
   # Create archive
   cd ..
   tar czf venture-diag-$(date +%Y%m%d-%H%M%S).tar.gz venture-diag-*
   ```

2. **File GitHub Issue:**
   - Title: "Production: High CPU usage - <brief description>"
   - Label: `P2`, `performance`, `bug`
   - Attach diagnostic bundle
   - Include: server version, player count, uptime, steps to reproduce

3. **Temporary Mitigation:**
   - Reduce player cap: `-max-players 4` (from 10)
   - Disable non-essential features: `-federation=false -mods=false`
   - Scale horizontally: Add second server instance, split player load

---

## Prevention

### Monitoring Setup

Configure alerts in monitoring system (Prometheus/Grafana):

```yaml
# Prometheus alert rules
- alert: HighCPU
  expr: rate(process_cpu_seconds_total[5m]) > 0.8
  for: 5m
  annotations:
    summary: "Venture server CPU >80% for 5 minutes"
    
- alert: GoroutineLeak
  expr: go_goroutines > 500
  annotations:
    summary: "Goroutine count >500, leak suspected"
    
- alert: EntityOverload
  expr: venture_entities_total > 5000
  annotations:
    summary: "Entity count >5000, cleanup issue suspected"
```

### Capacity Planning

Review capacity metrics weekly:

```bash
# Generate weekly capacity report
curl http://localhost:8081/status > weekly-$(date +%Y%m%d).json

# Compare week-over-week trends
# - Average CPU: should be <40%
# - Peak CPU: should be <60%
# - Goroutine baseline: should be <50
```

If approaching limits (CPU >50% average), scale vertically (more CPU cores) or horizontally (more instances).

---

## Related Runbooks

- [Memory Leak Investigation](memory-leak.md) - If GC pressure is root cause
- [Network Issues](network-issues.md) - If network storm is root cause
- [Federation Debugging](federation-debug.md) - If federation causing high load

---

## Revision History

| Date       | Author           | Changes                          |
|------------|------------------|----------------------------------|
| 2026-01-07 | Infrastructure   | Initial version for v10.0        |
