# Health and Readiness Endpoints

This document describes the health, readiness, and status endpoints provided by the Venture game server's observability system.

## Overview

The observability package (`pkg/observability`) provides three HTTP endpoints for monitoring server health and operational status:

- **`/health`** - Basic liveness check
- **`/ready`** - Readiness check with component validation
- **`/status`** - Detailed operational status with metrics

These endpoints are served by the `MetricsExporter` on a configurable port (default: 9090).

## Endpoints

### GET /health

**Purpose:** Basic liveness check to verify the server process is running.

**Response:**
- **Status Code:** `200 OK`
- **Content-Type:** `text/plain`
- **Body:** `OK\n`

**Use Case:** Load balancer health checks, basic uptime monitoring.

**Example:**
```bash
$ curl http://localhost:9090/health
OK
```

### GET /ready

**Purpose:** Readiness check to verify all critical components are operational.

**Response (All Checks Pass):**
- **Status Code:** `200 OK`
- **Content-Type:** `application/json`
- **Body:** `{"status":"ready"}`

**Response (One or More Checks Fail):**
- **Status Code:** `503 Service Unavailable`
- **Content-Type:** `application/json`
- **Body:** `{"status":"not_ready","failed_checks":["component: error message"]}`

**Use Case:** Kubernetes readiness probes, traffic routing decisions, determining if server should accept new connections.

**Example (Ready):**
```bash
$ curl http://localhost:9090/ready
{"status":"ready"}
```

**Example (Not Ready):**
```bash
$ curl http://localhost:9090/ready
{"status":"not_ready","failed_checks":["federation: connection timeout","database: connection refused"]}
```

**Implementation:**

Register custom readiness checkers using the `ReadinessChecker` interface:

```go
import "database/sql"

type ReadinessChecker interface {
    Check() (componentName string, err error)
}

// Example checker
type DatabaseChecker struct {
    db *sql.DB
}

func (c *DatabaseChecker) Check() (string, error) {
    if err := c.db.Ping(); err != nil {
        return "database", err
    }
    return "database", nil
}

// Register with exporter
exporter.RegisterReadinessChecker(&DatabaseChecker{db: myDB})
```

### GET /status

**Purpose:** Detailed operational status with performance metrics and resource usage.

**Response:**
- **Status Code:** `200 OK`
- **Content-Type:** `application/json`
- **Body:** JSON object with comprehensive server status

**Fields:**
- `status` - Overall status ("ok")
- `uptime_seconds` - Server uptime in seconds
- `started_at` - Server start time (RFC3339 format)
- `performance` - Performance metrics (FPS, frame time, memory)
- `network` - Network metrics (connected players, bytes sent/received, packets)
- `game_state` - Game state metrics (entity count, active quests, trade volume)
- `runtime` - Go runtime metrics (goroutines, heap allocation, GC runs)

**Use Case:** Operational dashboards, debugging, performance monitoring.

**Example:**
```bash
$ curl http://localhost:9090/status | jq
{
  "status": "ok",
  "uptime_seconds": 3600.42,
  "started_at": "2026-01-07T12:00:00Z",
  "performance": {
    "fps": 60.0,
    "frame_time_ms": 16.67,
    "memory_mb": 120
  },
  "network": {
    "connected_players": 4,
    "bytes_sent": 1024000,
    "bytes_received": 2048000,
    "packets_sent": 10000,
    "packets_received": 20000
  },
  "game_state": {
    "entity_count": 1500,
    "active_quests": 25,
    "trade_volume": 50000
  },
  "runtime": {
    "goroutines": 15,
    "heap_alloc_bytes": 125829120,
    "gc_runs": 42
  }
}
```

## Configuration

### Server Setup

Enable metrics endpoints when starting the server:

```bash
./venture-server --enable-metrics --metrics-port 9090
```

### Programmatic Usage

```go
import "github.com/opd-ai/venture/pkg/observability"

// Create exporter
exporter := observability.NewMetricsExporter(":9090")

// Register metrics sources (optional)
exporter.RegisterPerformanceMonitor(perfMonitor)
exporter.RegisterNetworkServer(netServer)
exporter.RegisterWorld(gameWorld)

// Register readiness checkers (optional)
exporter.RegisterReadinessChecker(databaseChecker)
exporter.RegisterReadinessChecker(federationChecker)

// Start serving endpoints
if err := exporter.Start(); err != nil {
    log.Fatal(err)
}

// Graceful shutdown
defer exporter.Stop()
```

## Response Time Requirements

As per PLAN.md Phase 3 success criteria:

- **Target:** All endpoints should respond in **<100ms**
- **Implementation:** Endpoints perform minimal work and use read locks for thread-safe access
- **Validation:** Response times are well under 100ms in testing (typically <5ms)

## Integration Examples

### Kubernetes Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 9090
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 2
  failureThreshold: 3
```

### Kubernetes Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 9090
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 2
  failureThreshold: 3
```

### Monitoring Dashboard

The `/status` endpoint provides JSON data suitable for custom dashboards. For Prometheus metrics, use the `/metrics` endpoint instead.

### Example Script

See `examples/health_endpoints_demo/main.go` for a complete working example demonstrating all three endpoints with custom readiness checkers.

## Testing

Run tests with:

```bash
go test -v ./pkg/observability
```

Test coverage: **98.2%**

## See Also

- [PLAN.md Phase 3](../../PLAN.md) - Production readiness plan
- [pkg/observability/metrics.go](../../pkg/observability/metrics.go) - Implementation
- [examples/health_endpoints_demo](../../examples/health_endpoints_demo/) - Working example
