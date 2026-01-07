# Runbook: Memory Leak Investigation

**Severity:** P1 (Service Degrading)  
**Symptoms:** Increasing memory usage over time, eventual OOM crash, high GC activity  
**Owner:** Infrastructure Team  
**Last Updated:** 2026-01-07

---

## Overview

This runbook helps diagnose and resolve memory leaks in Venture game servers. Memory leaks manifest as continuously increasing memory usage that doesn't stabilize, eventually leading to out-of-memory (OOM) crashes.

**Expected State:** Memory usage <500MB with 4 players, stable over time  
**Alert Threshold:** Memory >800MB or growth >10MB/hour

---

## Initial Assessment (5 minutes)

### 1. Verify Memory Growth

Check current memory usage:

```bash
# Get current memory stats
curl http://localhost:8081/status | jq '{memory_mb: .performance.memory_mb, heap_alloc: .runtime.heap_alloc_bytes, goroutines: .runtime.goroutines}'
```

Record baseline, wait 5 minutes, check again:

```bash
sleep 300
curl http://localhost:8081/status | jq '{memory_mb: .performance.memory_mb, heap_alloc: .runtime.heap_alloc_bytes, goroutines: .runtime.goroutines}'
```

**Calculate growth rate:**
- Growth >50MB in 5 minutes = Critical leak (1GB/hour)
- Growth 10-50MB in 5 minutes = Moderate leak (120-600MB/hour)
- Growth <10MB in 5 minutes = Slow leak or normal variance

### 2. Check System Memory

Verify OS-level memory usage:

```bash
# Linux
free -m
ps aux | grep venture-server | awk '{print $6/1024 " MB"}'

# macOS
vm_stat
ps -p $(pgrep venture-server) -o rss | tail -1 | awk '{print $1/1024 " MB"}'
```

Compare process RSS to heap_alloc:
- RSS > 2× heap_alloc: Memory fragmentation or off-heap allocation
- RSS ≈ heap_alloc: Normal (Go heap dominates)

### 3. Review GC Activity

Check garbage collection frequency:

```bash
curl http://localhost:8081/status | jq '.runtime.gc_runs'
# Wait 60 seconds
sleep 60
curl http://localhost:8081/status | jq '.runtime.gc_runs'
```

Calculate GC rate: (second_value - first_value) / 60 seconds

**Expected:** <1 GC/sec with stable memory  
**Alert if:** >5 GC/sec (GC thrashing, memory pressure)

---

## Diagnosis (15 minutes)

### Step 1: Capture Memory Profile

Get heap allocation snapshot:

```bash
# Capture heap profile
curl http://localhost:8081/debug/pprof/heap > heap-$(date +%Y%m%d-%H%M%S).prof

# Analyze with pprof
go tool pprof heap-*.prof

# In pprof interactive mode:
# (pprof) top20              # Show top 20 allocators
# (pprof) list <function>    # Show source code of function
# (pprof) web                # Generate visual graph (requires graphviz)
```

**Common Leak Patterns:**

1. **Sprite Cache Growth**
   - Symptom: `rendering.SpriteCache` in top allocations
   - Cause: Cache not respecting size limits
   - Fix: Reduce cache size or fix eviction logic

2. **Entity Accumulation**
   - Symptom: `engine.Entity` or component allocations growing
   - Cause: Entities not being removed from world
   - Fix: Fix entity cleanup logic

3. **Network Buffers**
   - Symptom: `network.PacketBuffer` or `[]byte` allocations
   - Cause: Buffers not being released after use
   - Fix: Implement buffer pooling or fix release logic

4. **Goroutine Stacks**
   - Symptom: High goroutine count + memory growth
   - Cause: Goroutines leaking (see high-cpu.md)
   - Fix: Fix goroutine termination

### Step 2: Compare Two Heap Snapshots

Capture second snapshot after 10 minutes to see growth:

```bash
# Wait 10 minutes
sleep 600

# Capture second snapshot
curl http://localhost:8081/debug/pprof/heap > heap-second-$(date +%Y%m%d-%H%M%S).prof

# Compare snapshots to see what grew
go tool pprof -base heap-first-*.prof heap-second-*.prof

# In pprof:
# (pprof) top20    # Shows allocations that GREW between snapshots
```

This reveals the leak source directly.

### Step 3: Check Entity and Resource Counts

```bash
# Check entity count trend
curl http://localhost:8081/status | jq '.game_state.entity_count'

# Check active quest count
curl http://localhost:8081/status | jq '.game_state.active_quests'

# Check goroutine count
curl http://localhost:8081/status | jq '.runtime.goroutines'
```

If these are growing unbounded, you've found the leak:
- Entities: Fix entity removal logic
- Quests: Fix quest completion/cleanup
- Goroutines: Fix goroutine termination (see high-cpu.md)

### Step 4: Check Cache Statistics

Review sprite cache performance:

```bash
# Check cache hit rate (should be in logs)
grep "sprite_cache_hit_rate" /var/log/venture-server.log | tail -10

# Expected: >95% hit rate
# If hit rate is high but memory growing: cache size limit not working
```

Check if cache is respecting size limits:

```bash
# Expected cache size from config
grep "cache.*mb" /etc/venture/server.conf

# Actual memory used (from heap profile)
go tool pprof -top heap-*.prof | grep -i cache
```

### Step 5: Enable Memory Leak Detection

If available, enable Go's memory leak detector:

```bash
# Set GODEBUG environment variable
export GODEBUG=gctrace=1

# Restart server
systemctl restart venture-server

# Monitor GC logs in /var/log/venture-server.log
tail -f /var/log/venture-server.log | grep "gc "

# Look for:
# - Increasing heap size after each GC (indicates leak)
# - Frequent GCs with little memory reclaimed (indicates retention)
```

---

## Resolution

### Scenario 1: Sprite Cache Leak

**Symptom:** `SpriteCache` allocations growing, cache_hit_rate >95%  
**Root Cause:** Cache not evicting old entries

```bash
# Immediate: Reduce cache size to force eviction
# Edit /etc/venture/server.conf:
cache_size_mb=50  # Reduce from 150

# Restart server
systemctl restart venture-server

# Monitor memory usage
watch -n 10 'curl -s http://localhost:8081/status | jq .performance.memory_mb'
```

Expected: Memory stabilizes at lower level.

If memory still grows, cache eviction is broken - file bug report.

### Scenario 2: Entity Leak

**Symptom:** entity_count growing continuously (>100 entities/min)  
**Root Cause:** Entities not being removed (projectiles, particles, temporary objects)

```bash
# Immediate: Restart server to clear entities
systemctl restart venture-server

# Monitor entity count
watch -n 30 'curl -s http://localhost:8081/status | jq .game_state.entity_count'
```

If entity count grows rapidly after restart:
1. Check logs for entity creation errors: `grep "CreateEntity" /var/log/venture-server.log`
2. Identify entity type growing: Enable debug logging and check entity types
3. File bug report with entity type and creation rate

**Workaround:** Reduce entity lifespan or limit max entities (if configurable).

### Scenario 3: Network Buffer Leak

**Symptom:** `network.PacketBuffer` or `[]byte` in top allocations  
**Root Cause:** Buffers not returned to pool after use

```bash
# Check network stats for abnormal patterns
curl http://localhost:8081/status | jq '.network'

# Look for:
# - Very high packet counts (>1M packets)
# - Unusual packet size patterns
```

**Immediate mitigation:**
```bash
# Restart server to release buffers
systemctl restart venture-server

# Reduce network buffer size if configurable
# Edit /etc/venture/server.conf:
network_buffer_kb=64  # Reduce from 256
```

If leak persists, file bug report with network statistics.

### Scenario 4: Goroutine Stack Leak

**Symptom:** goroutine count >500, memory growing  
**Root Cause:** Each goroutine has a stack (~8KB), leaking goroutines leaks memory

See [high-cpu.md](high-cpu.md) for goroutine leak resolution.

```bash
# Immediate: Restart to clear goroutines
systemctl restart venture-server

# Monitor goroutine count
watch -n 10 'curl -s http://localhost:8081/status | jq .runtime.goroutines'
```

### Scenario 5: Memory Fragmentation

**Symptom:** RSS >> heap_alloc (2×+ difference), GC running but memory not decreasing  
**Root Cause:** OS-level memory fragmentation

```bash
# Force GC and memory release
curl http://localhost:8081/debug/pprof/heap?gc=1

# Wait 30 seconds for OS to reclaim
sleep 30

# Check if RSS decreased
ps aux | grep venture-server | awk '{print $6/1024 " MB"}'
```

If RSS doesn't decrease significantly (>10%):
1. Fragmentation is severe
2. Consider restarting server during low-traffic period
3. Plan for scheduled restarts (e.g., daily at 4 AM)

**Long-term:** Enable jemalloc or tcmalloc for better memory allocator (Go 1.22+ experimental feature).

---

## Emergency Response (OOM Imminent)

If memory >90% of available RAM or server is unresponsive:

### Option 1: Graceful Restart (Preferred)

```bash
# Save current state if possible
curl -X POST http://localhost:8081/admin/save-world

# Gracefully shutdown
systemctl stop venture-server

# Verify shutdown
ps aux | grep venture-server

# Start fresh instance
systemctl start venture-server

# Verify startup
curl http://localhost:8081/health
```

### Option 2: Force Kill (If Unresponsive)

```bash
# Kill server process
pkill -9 venture-server

# Start fresh instance
systemctl start venture-server

# Verify startup
curl http://localhost:8081/health

# Check for corrupted save files
ls -lh /var/lib/venture/saves/*.bak

# Restore from backup if needed (see save-corruption.md)
```

---

## Monitoring and Verification (30 minutes)

After resolution, monitor memory for 30 minutes to confirm stability:

```bash
# Log memory every minute for 30 minutes
for i in {1..30}; do
  echo "$(date): $(curl -s http://localhost:8081/status | jq '.performance.memory_mb') MB"
  sleep 60
done > memory-trend.log

# Plot trend (requires gnuplot)
# If memory growth <5MB over 30 min, issue resolved
# If memory growth >10MB over 30 min, leak persists
```

**Success Criteria:**
- Memory growth <10MB per hour
- GC rate <1/sec
- Entity/goroutine counts stable
- No RSS growth beyond heap growth

---

## Prevention

### Set Resource Limits

Prevent single process from consuming all system memory:

```bash
# Edit /etc/systemd/system/venture-server.service
[Service]
MemoryLimit=1G
MemoryMax=1.5G

# Reload systemd
systemctl daemon-reload
systemctl restart venture-server
```

Server will be terminated by kernel if exceeding limit (better than full system OOM).

### Enable Memory Monitoring

Set up alerts for memory growth:

```yaml
# Prometheus alert rules
- alert: MemoryLeak
  expr: rate(venture_memory_usage_mb[1h]) > 10
  for: 1h
  annotations:
    summary: "Memory growing >10MB/hour, leak suspected"
    
- alert: MemoryHigh
  expr: venture_memory_usage_mb > 800
  annotations:
    summary: "Memory usage >800MB (limit: 1GB)"
    
- alert: GCThrashing
  expr: rate(venture_go_gc_runs_total[1m]) > 5
  annotations:
    summary: "GC running >5 times/sec, memory pressure"
```

### Scheduled Restarts

If leak is slow and fix is not urgent, schedule daily restarts:

```bash
# Add cron job for daily restart at 4 AM
crontab -e

# Add line:
0 4 * * * systemctl restart venture-server
```

This prevents gradual memory accumulation from causing outages.

### Capacity Planning

Review memory trends weekly:

```bash
# Check weekly peak memory
grep "memory_mb" /var/log/venture-metrics/weekly-*.json | \
  jq -s 'max_by(.performance.memory_mb)'

# If peak >600MB (60% of 1GB limit), investigate or increase capacity
```

---

## Escalation

If leak persists after all resolution attempts:

1. **Collect Diagnostic Bundle:**
   ```bash
   # Create diagnostic directory
   mkdir venture-memory-diag-$(date +%Y%m%d-%H%M%S)
   cd venture-memory-diag-*
   
   # Capture multiple heap snapshots over 30 minutes
   for i in {1..6}; do
     curl http://localhost:8081/debug/pprof/heap > heap-$i-$(date +%H%M).prof
     sleep 300  # 5 minutes between snapshots
   done
   
   # Capture goroutine profile
   curl http://localhost:8081/debug/pprof/goroutine > goroutine.txt
   
   # Capture status snapshots
   curl http://localhost:8081/status > status-start.json
   sleep 1800  # 30 minutes
   curl http://localhost:8081/status > status-end.json
   
   # Capture logs
   cp /var/log/venture-server.log .
   
   # Create archive
   cd ..
   tar czf venture-memory-diag-$(date +%Y%m%d-%H%M%S).tar.gz venture-memory-diag-*
   ```

2. **File GitHub Issue:**
   - Title: "Production: Memory leak - <brief description>"
   - Label: `P1`, `memory-leak`, `bug`
   - Attach diagnostic bundle
   - Include: server version, player count, uptime, memory growth rate

3. **Temporary Mitigation:**
   - Reduce cache sizes: `-cache-size 50`
   - Reduce player cap: `-max-players 4`
   - Enable scheduled restarts (daily)
   - Scale horizontally: Multiple smaller instances instead of one large

---

## Related Runbooks

- [High CPU Usage](high-cpu.md) - If GC pressure causing high CPU
- [Save Corruption](save-corruption.md) - If OOM crash corrupted save files

---

## Revision History

| Date       | Author           | Changes                          |
|------------|------------------|----------------------------------|
| 2026-01-07 | Infrastructure   | Initial version for v10.0        |
