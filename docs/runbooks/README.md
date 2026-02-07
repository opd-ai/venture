# Operational Runbooks

This directory contains step-by-step operational runbooks for diagnosing and resolving production issues in Venture game servers. These runbooks are designed for server operators, system administrators, and infrastructure teams.

## Available Runbooks

| Runbook | Severity | Description | Estimated Resolution Time |
|---------|----------|-------------|---------------------------|
| [high-cpu.md](high-cpu.md) | P2 | High CPU usage troubleshooting | 15-30 minutes |
| [memory-leak.md](memory-leak.md) | P1 | Memory leak investigation and recovery | 20-45 minutes |
| [network-issues.md](network-issues.md) | P1 | Network connectivity problems | 10-30 minutes |
| [save-corruption.md](save-corruption.md) | P2 | Save file corruption recovery | 5-30 minutes |
| [federation-debug.md](federation-debug.md) | P2 | Federation connectivity debugging | 15-45 minutes |

## Severity Levels

- **P0 (Critical):** Complete service outage, immediate action required
- **P1 (High):** Service degradation affecting users, resolve within 1 hour
- **P2 (Medium):** Feature degradation or performance issue, resolve within 4 hours
- **P3 (Low):** Minor issue or cosmetic problem, resolve within 24 hours

## Using These Runbooks

### Structure

Each runbook follows a consistent structure:

1. **Overview** - Summary, symptoms, expected state
2. **Initial Assessment** - Quick checks to verify the issue (5 minutes)
3. **Diagnosis** - Detailed troubleshooting steps (10-15 minutes)
4. **Resolution** - Step-by-step fixes for common scenarios
5. **Monitoring** - Verification and post-recovery monitoring
6. **Prevention** - Long-term measures to prevent recurrence
7. **Escalation** - When and how to escalate if issue persists

### Prerequisites

Before using these runbooks, ensure you have:

- **Access to server:** SSH access with appropriate permissions
- **Monitoring endpoints:** Health/metrics endpoints accessible at `http://localhost:8081/`
- **Tools installed:** `curl`, `jq`, `netstat`/`ss`, `tcpdump` (optional), `go` toolchain (for profiling)
- **Log access:** Read access to `/var/log/venture-server.log`

### Quick Start

1. **Identify the symptom** from the table above
2. **Open the relevant runbook**
3. **Start with Initial Assessment** - verify you have the right issue
4. **Follow Diagnosis steps** in order - don't skip steps
5. **Apply Resolution** for your specific scenario
6. **Monitor** to verify the fix worked
7. **Escalate** if issue persists after all resolution attempts

## Common Diagnostic Commands

### Server Health Check

```bash
# Quick health check
curl http://localhost:8081/health

# Detailed status (requires jq)
curl http://localhost:8081/status | jq '.'

# Readiness check
curl http://localhost:8081/ready | jq '.'
```

### Metrics Collection

```bash
# Prometheus metrics
curl http://localhost:8081/metrics

# Specific metric extraction
curl -s http://localhost:8081/metrics | grep venture_fps
```

### Log Analysis

```bash
# Recent errors
grep -i "error\|fatal" /var/log/venture-server.log | tail -20

# Errors in last hour
journalctl -u venture-server --since "1 hour ago" | grep -i error

# Follow logs in real-time
tail -f /var/log/venture-server.log
```

### Performance Profiling

```bash
# CPU profile (30 seconds)
curl http://localhost:8081/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Heap profile
curl http://localhost:8081/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Goroutine dump
curl http://localhost:8081/debug/pprof/goroutine > goroutine.txt
```

## Monitoring Setup

For production deployments, set up monitoring with alerts:

### Prometheus Metrics

Add Venture server as a scrape target:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'venture'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Alert Rules

Example Prometheus alert rules (see individual runbooks for specific thresholds):

```yaml
groups:
  - name: venture-alerts
    interval: 30s
    rules:
      - alert: VentureServerDown
        expr: up{job="venture"} == 0
        for: 1m
        annotations:
          summary: "Venture server is down"
          
      - alert: HighCPU
        expr: rate(process_cpu_seconds_total{job="venture"}[5m]) > 0.8
        for: 5m
        annotations:
          summary: "Venture server CPU >80%"
          
      - alert: HighMemory
        expr: venture_memory_usage_mb > 800
        annotations:
          summary: "Venture server memory >800MB"
```

### Grafana Dashboard

Import pre-built dashboards (if available) or create custom:

**Key Panels:**
- FPS over time
- Memory usage over time
- Connected players
- Entity count
- Network bandwidth (bytes sent/received)
- Goroutine count
- GC runs per minute

## Integration with Incident Management

### Creating Incidents

When escalating (see individual runbooks), create incidents following this template:

```
Title: [P1] Venture Production - <Brief Description>

Description:
- Symptom: <What users are experiencing>
- Affected servers: <Server names/IPs>
- Started at: <Timestamp>
- Player impact: <Number of affected players>
- Runbook attempted: <Which runbook(s)>
- Steps taken: <Summary of troubleshooting>
- Current state: <Is issue ongoing or mitigated?>

Attachments:
- Diagnostic bundle (see runbook Escalation section)
- Relevant log excerpts
- Screenshots of monitoring dashboards
```

### Communication

During incidents, follow communication protocol:

1. **Acknowledge** - Confirm receipt within 5 minutes
2. **Assess** - Provide initial assessment within 15 minutes
3. **Update** - Status updates every 30 minutes until resolved
4. **Resolve** - Confirm resolution and document root cause
5. **Postmortem** - Document lessons learned within 24 hours

## Maintenance Windows

Some resolutions require server restarts. Schedule during low-traffic periods:

**Recommended maintenance windows:**
- **North America:** Tuesday/Wednesday 2-6 AM PST
- **Europe:** Tuesday/Wednesday 1-5 AM GMT
- **Asia:** Tuesday/Wednesday 2-6 AM JST

**Pre-maintenance checklist:**
1. Announce maintenance 24 hours in advance
2. Verify backups are current
3. Prepare rollback plan
4. Have staff on-call during window
5. Monitor health checks post-maintenance

## Related Documentation

- [TROUBLESHOOTING.md](../TROUBLESHOOTING.md) - User-facing troubleshooting guide
- [PRODUCTION_DEPLOYMENT.md](../PRODUCTION_DEPLOYMENT.md) - Production deployment guide
- [HEALTH_ENDPOINTS.md](../HEALTH_ENDPOINTS.md) - Health endpoint documentation
- [ERROR_HANDLING.md](../ERROR_HANDLING.md) - Error handling architecture
- [SECURITY.md](../SECURITY.md) - Security policies and incident response

## Contributing

Found an issue not covered by existing runbooks? Follow this process:

1. **Document the incident** thoroughly during resolution
2. **Create a new runbook** following the existing structure
3. **Submit a pull request** with:
   - New runbook markdown file
   - Update to this README
   - Real-world validation (runbook tested in production or staging)

## Support

If runbooks don't resolve your issue:

- **GitHub Issues:** [opd-ai/venture/issues](https://github.com/opd-ai/venture/issues)
- **Emergency:** Page on-call engineer via PagerDuty/Opsgenie
- **Questions:** #infrastructure channel on Discord/Slack

---

**Last Updated:** 2026-02-07  
**Maintained By:** Infrastructure Team  
**Version:** 1.0.0
