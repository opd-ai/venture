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
	"github.com/opd-ai/venture/pkg/stability"
	"github.com/sirupsen/logrus"
)

// startPerformanceMonitoring initializes all performance monitoring goroutines.
// Only runs when verbose mode is enabled.
func startPerformanceMonitoring(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	if !*verbose {
		return
	}
	perfMonitor, stabilityMonitor := initializeMonitors(game, clientLogger)
	startLegacyMetricsMonitor(perfMonitor, clientLogger)
	startStabilityMonitor(game, stabilityMonitor, clientLogger)
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
func startLegacyMetricsMonitor(perfMonitor *engine.PerformanceMonitor, clientLogger *logrus.Entry) {
	go func() {
		ticker := time.NewTicker(perfMonitorInterval * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			metrics := perfMonitor.GetMetrics()
			clientLogger.WithField("metrics", metrics.String()).Info("performance metrics")
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
