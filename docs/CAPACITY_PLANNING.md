# Capacity Planning Guide

This document provides capacity planning guidelines for Venture multiplayer servers based on performance benchmarks and load testing results.

## Performance Targets

| Metric | Target | Measured |
|--------|--------|----------|
| Frame Rate | 60 FPS | 89 FPS |
| Memory (Client) | <500 MB | ~120 MB |
| Memory (Server, 4 players) | <1 GB | ~200 MB |
| Network Bandwidth | <100 KB/s per player | ~15-30 KB/s |
| Entity Count | 2000 entities at 60 FPS | ✅ Achieved |

## Server Capacity Estimates

Based on benchmarks from `docs/PERFORMANCE.md` and load testing with `scripts/load-test.sh`:

### Recommended Player Limits

| Server Type | CPU Cores | RAM | Max Players | Notes |
|-------------|-----------|-----|-------------|-------|
| Minimal | 1 | 512 MB | 4 | Basic gameplay, limited entities |
| Standard | 2 | 2 GB | 16 | Full features, moderate entity count |
| High Capacity | 4+ | 4 GB | 32 | Full features, high entity density |
| Dedicated | 8+ | 8 GB | 64+ | Multiple zones, federation support |

### Resource Usage Per Player

| Resource | Idle Player | Active Player | Combat |
|----------|-------------|---------------|--------|
| CPU | ~1% | ~2-3% | ~5% |
| Memory | ~20 MB | ~30 MB | ~50 MB |
| Network TX | ~5 KB/s | ~15 KB/s | ~30 KB/s |
| Network RX | ~2 KB/s | ~5 KB/s | ~10 KB/s |

## Network Considerations

### Bandwidth Requirements

```
Total Bandwidth = Players × (TX + RX) × Update Rate Factor

Example (16 players, standard activity):
  16 × (15 + 5) KB/s × 1.0 = 320 KB/s = ~2.5 Mbps
```

### Latency Tolerance

| Network Type | Latency | Client Config | Server Config |
|--------------|---------|---------------|---------------|
| LAN | <10ms | Default | Default |
| Internet | 50-200ms | Default | Default |
| High Latency | 200-1000ms | Default | HighLatency |
| Tor/Onion | 1000-5000ms | Tor | HighLatency |

## Memory Planning

### Base Memory Footprint

```
Server Memory = Base + (Players × Per-Player) + (Entities × Per-Entity)

Base: ~50 MB
Per-Player: ~30 MB
Per-Entity: ~1 KB

Example (16 players, 2000 entities):
  50 + (16 × 30) + (2000 × 0.001) = 532 MB
```

### Memory Hotspots

| Component | Memory Impact | Mitigation |
|-----------|--------------|------------|
| Sprite Cache | 10-50 MB | LRU eviction (automatic) |
| Entity Components | 1 KB/entity | Spatial partitioning limits active |
| Network Buffers | 256 KB/client | Buffer pool reuse |
| Procedural Gen Cache | 5-20 MB | On-demand generation |

## CPU Planning

### System CPU Budget (per tick at 20 Hz)

| System | Budget | Typical | Heavy Load |
|--------|--------|---------|------------|
| Physics | 15ms | 2-5ms | 10ms |
| Collision | 10ms | 1-3ms | 8ms |
| AI/Behavior | 10ms | 2-4ms | 8ms |
| Rendering (server) | N/A | N/A | N/A |
| Network Sync | 5ms | 1-2ms | 4ms |
| **Total Available** | **50ms** | 6-14ms | 30ms |

### CPU Scaling

```
Entities at 60 FPS ≈ 2000 baseline
Each additional CPU core adds ~40% capacity
Multi-zone servers can run zones on separate cores
```

## Load Testing

### Running Load Tests

```bash
# Basic test (10 clients, 5 minutes)
./scripts/load-test.sh

# Extended test (20 clients, 20 minutes)
./scripts/load-test.sh --clients 20 --duration 20m

# High-latency simulation
./scripts/load-test.sh --min-latency 100ms --max-latency 1000ms

# Test specific server
./scripts/load-test.sh --server gameserver.example.com:8080
```

### Interpreting Results

**Success Criteria:**
- All clients maintain 90%+ uptime
- CPU usage stays below 80%
- Memory usage stays below target
- Zero message queue overflow

**Warning Signs:**
- Frequent reconnections indicate network issues
- CPU >80% suggests scaling needed
- Memory growth over time indicates leaks
- High error count suggests protocol issues

## Scaling Strategies

### Vertical Scaling

1. **Add CPU cores** - Enables more concurrent systems
2. **Add RAM** - Supports more entities and players
3. **Faster storage** - Improves save/load times
4. **Better network** - Reduces latency, increases throughput

### Horizontal Scaling (Federation)

For player counts beyond single-server capacity:

1. **Zone Sharding** - Split world into zones on separate servers
2. **Instance Sharding** - Multiple instances of same zone
3. **Federation** - Cross-server travel and communication

Federation is enabled via WebRTC peer connections:

```bash
# Primary server
./venture-server --port 8080 --federation-enabled

# Secondary server (federated)
./venture-server --port 8081 --federation-peer localhost:8080
```

## Monitoring

### Key Metrics

| Metric | Source | Warning Threshold | Critical |
|--------|--------|-------------------|----------|
| CPU Usage | /proc/stat | >70% | >90% |
| Memory Usage | /proc/meminfo | >70% | >90% |
| Active Connections | Server stats | >80% max | 100% |
| Tick Duration | Server logs | >40ms | >50ms |
| Message Queue | Buffer stats | >80% | Full |

### Prometheus Metrics

When observability is enabled (`/metrics` endpoint):

```prometheus
# Player count
venture_server_players_active

# Memory usage
process_resident_memory_bytes

# Network stats
venture_network_bytes_sent_total
venture_network_bytes_received_total

# Game loop timing
venture_tick_duration_seconds
```

## Capacity Checklist

Before deploying to production:

- [ ] Run load test with expected player count
- [ ] Verify memory stays within budget over 1+ hour
- [ ] Confirm CPU headroom of at least 20%
- [ ] Test with high-latency client simulation
- [ ] Verify graceful degradation under overload
- [ ] Set up monitoring and alerting
- [ ] Document rollback procedures
- [ ] Test save/load under load

## Troubleshooting

### Server Overload

1. Check active player count vs capacity
2. Review entity count (spatial partitioning stats)
3. Check for memory leaks (steadily increasing usage)
4. Review tick timing logs for slow systems

### Client Disconnections

1. Check network latency (ping test)
2. Review server logs for timeout errors
3. Verify client config matches network conditions
4. Check for packet loss (network diagnostics)

### Memory Growth

1. Enable pprof profiling
2. Take heap snapshots before/after load
3. Check sprite cache hit rate
4. Review entity cleanup on player disconnect

---

*See also: [docs/PERFORMANCE.md](PERFORMANCE.md) for benchmark details, [docs/runbooks/](runbooks/) for operational procedures.*
