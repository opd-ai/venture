// Package observability provides monitoring and observability infrastructure for Venture.
// It includes Prometheus metrics export, health checks, and distributed tracing support.
//
// Metrics Export:
// The package exports key performance metrics in Prometheus format via HTTP endpoint.
// Core metrics include FPS, entity count, memory usage, and network traffic.
// Game-specific metrics include player count, active quests, and trade volume.
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
package observability
