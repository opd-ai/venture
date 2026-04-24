//go:build !android && !ios
// +build !android,!ios

// Package main contains the client entry point and game initialization.
// This file contains performance and stability monitoring initialization.

package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/observability"
	"github.com/opd-ai/venture/pkg/stability"
	"github.com/sirupsen/logrus"
)

// startPerformanceMonitoring initializes all performance monitoring goroutines.
// Only runs when verbose mode is enabled. Returns a cancel function that stops
// the legacy metrics goroutine on game exit (AUDIT.md G14-5).
func startPerformanceMonitoring(game *engine.EbitenGame, clientLogger *logrus.Entry) context.CancelFunc {
	if !*verbose {
		return func() {}
	}
	ctx, cancel := context.WithCancel(context.Background())
	perfMonitor, stabilityMonitor := initializeMonitors(game, clientLogger)
	startLegacyMetricsMonitor(ctx, perfMonitor, clientLogger)
	startStabilityMonitor(game, stabilityMonitor, clientLogger)
	return cancel
}

// initializeMonitors creates and configures performance and stability monitors.
func initializeMonitors(game *engine.EbitenGame, clientLogger *logrus.Entry) (*engine.PerformanceMonitor, *stability.Monitor) {
	perfMonitor := engine.NewPerformanceMonitor(game.World)
	stabilityConfig := stability.Config{
		Duration:      0,
		CheckInterval: 30 * time.Second,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
		ReportPath:    "",
	}
	stabilityMonitor := stability.NewMonitor(stabilityConfig)
	stabilityMonitor.SetFPSProvider(game)
	clientLogger.WithFields(logrus.Fields{
		"min_fps":      stabilityConfig.MinFPS,
		"memory_limit": stabilityConfig.MemoryLimit / (1024 * 1024),
	}).Info("performance monitoring initialized with stability enforcement")
	return perfMonitor, stabilityMonitor
}

// startLegacyMetricsMonitor starts background goroutine for legacy performance metrics.
// The goroutine shuts down cleanly when ctx is cancelled (AUDIT.md G14-5).
func startLegacyMetricsMonitor(ctx context.Context, perfMonitor *engine.PerformanceMonitor, clientLogger *logrus.Entry) {
	go func() {
		ticker := time.NewTicker(perfMonitorInterval * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metrics := perfMonitor.GetMetrics()
				clientLogger.WithField("metrics", metrics.String()).Info("performance metrics")
			}
		}
	}()
}

// startStabilityMonitor starts background goroutine for stability monitoring.
func startStabilityMonitor(game *engine.EbitenGame, monitor *stability.Monitor, clientLogger *logrus.Entry) {
	go func() {
		ctx := context.Background()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				checkAndLogStability(game, clientLogger)
			}
		}
	}()
}

// checkAndLogStability performs health check and logs warnings when targets are exceeded.
func checkAndLogStability(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	fps, currentMem := collectPerformanceMetrics(game)
	fields := buildPerformanceFields(fps, currentMem)
	logPerformanceStatus(fps, currentMem, fields, clientLogger)
}

// collectPerformanceMetrics gathers current FPS and memory usage.
func collectPerformanceMetrics(game *engine.EbitenGame) (float64, uint64) {
	fps := game.CurrentFPS()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return fps, memStats.HeapAlloc
}

// buildPerformanceFields constructs log fields for performance metrics.
func buildPerformanceFields(fps float64, currentMem uint64) logrus.Fields {
	return logrus.Fields{
		"fps":        fmt.Sprintf("%.1f", fps),
		"memory_mb":  fmt.Sprintf("%.1f", float64(currentMem)/(1024*1024)),
		"goroutines": runtime.NumGoroutine(),
	}
}

// logPerformanceStatus logs warnings for threshold violations or debug for passing checks.
func logPerformanceStatus(fps float64, currentMem uint64, fields logrus.Fields, clientLogger *logrus.Entry) {
	const minFPS = 60.0
	const memoryLimit = 500 * 1024 * 1024
	if fps < minFPS {
		fields["target_fps"] = minFPS
		clientLogger.WithFields(fields).Warn("FPS below target")
	}
	if currentMem > memoryLimit {
		fields["target_memory_mb"] = memoryLimit / (1024 * 1024)
		clientLogger.WithFields(fields).Warn("memory usage above target")
	}
	if fps >= minFPS && currentMem <= memoryLimit {
		clientLogger.WithFields(fields).Debug("stability check passed")
	}
}

// initializeClientMetrics starts the opt-in Prometheus metrics HTTP endpoint on
// the client process.  It is only started when --enable-metrics is set.
// The endpoint is bound to localhost only to avoid accidental exposure.
//
// Satisfies AUDIT.md G7: client observability surface for host-and-play.
func initializeClientMetrics(game *engine.EbitenGame, clientLogger *logrus.Entry) func() {
	if !*clientEnableMetrics {
		return func() {}
	}

	addr := "127.0.0.1:" + *clientMetricsPort
	exporter := initObservabilityExporter(game, addr, clientLogger)
	if exporter == nil {
		return func() {}
	}

	if err := exporter.Start(); err != nil {
		clientLogger.WithFields(logrus.Fields{
			"addr":  addr,
			"error": err,
		}).Warn("client metrics: failed to start exporter — continuing without it")
		return func() {}
	}

	clientLogger.WithField("addr", addr).Info("client metrics endpoint started")

	return func() {
		if err := exporter.Stop(); err != nil {
			clientLogger.WithError(err).Warn("client metrics: failed to stop exporter")
		}
	}
}

// initObservabilityExporter creates a MetricsExporter, registers the world and
// the PerformanceMonitoringSystem from the game, and returns it ready to Start().
// Returns nil if addr is empty or the exporter cannot be constructed.
func initObservabilityExporter(game *engine.EbitenGame, addr string, clientLogger *logrus.Entry) *observability.MetricsExporter {
	exporter := observability.NewMetricsExporterWithLogger(addr, clientLogger.Logger)

	// Locate the PerformanceMonitoringSystem already registered in the world.
	var perfMon *engine.PerformanceMonitoringSystem
	for _, sys := range game.World.GetSystems() {
		if pms, ok := sys.(*engine.PerformanceMonitoringSystem); ok {
			perfMon = pms
			break
		}
	}
	if perfMon == nil {
		perfMon = engine.NewPerformanceMonitoringSystem()
		game.World.AddSystem(perfMon)
	}

	exporter.RegisterPerformanceMonitor(perfMon)
	exporter.RegisterWorld(game.World)
	return exporter
}
