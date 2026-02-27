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
//	type TerrainLoadChecker struct {
//	    world *engine.World
//	}
//
//	func (t *TerrainLoadChecker) Check() (componentName string, err error) {
//	    if t.world == nil {
//	        return "terrain", fmt.Errorf("world not initialized")
//	    }
//	    // Check if initial terrain chunks are loaded
//	    if t.world.GetEntityCount() == 0 {
//	        return "terrain", fmt.Errorf("no entities loaded")
//	    }
//	    return "terrain", nil
//	}
//
//	// Register with the exporter
//	exporter.RegisterReadinessChecker(&TerrainLoadChecker{world: world})
//
// The /ready endpoint will include all registered checkers in its response.
package observability
