// Package observability provides monitoring and observability infrastructure for Venture.
// It includes Prometheus metrics export, health checks, and readiness endpoints.
//
// # Metrics Export
//
// The package exports key performance metrics in Prometheus format via HTTP endpoint.
// Core metrics include FPS, entity count, memory usage, and network traffic.
// Game-specific metrics include connected players, active quests, and trade volume.
//
// Usage:
//
//	// Create metrics exporter
//	exporter := observability.NewMetricsExporter(":9090")
//
//	// Register metrics sources
//	exporter.RegisterPerformanceMonitor(perfMon)
//	exporter.RegisterNetworkServer(server)
//	exporter.RegisterWorld(world)
//
//	// Start metrics HTTP server
//	exporter.Start()
//	defer exporter.Stop()
//
// The /metrics endpoint returns metrics in Prometheus exposition format,
// suitable for scraping by Prometheus or compatible monitoring systems.
//
// # Readiness Checks
//
// Custom readiness checks can be added by implementing the ReadinessChecker interface:
//
//	type DatabaseChecker struct {
//	    db *sql.DB
//	}
//
//	func (d *DatabaseChecker) Check() (componentName string, err error) {
//	    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	    defer cancel()
//	    if err := d.db.PingContext(ctx); err != nil {
//	        return "database", fmt.Errorf("database ping failed: %w", err)
//	    }
//	    return "database", nil
//	}
//
//	// Register with the exporter
//	exporter.RegisterReadinessChecker(&DatabaseChecker{db: db})
//
// The /ready endpoint will include all registered checkers in its response.
package observability
