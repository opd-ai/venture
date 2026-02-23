// Package stability provides long-running stability testing for server uptime validation.
//
// # Overview
//
// The stability package implements automated 72-hour uptime testing as required by
// Phase 66 of V10.0 Production Readiness. It monitors server health, detects crashes,
// memory leaks, and performance degradation over extended run periods.
//
// # Features
//
// - Continuous health monitoring with configurable intervals
// - Memory leak detection via heap profiling
// - Performance regression detection (FPS, frame time)
// - Crash detection and recovery logging
// - Detailed stability reports with metrics
//
// # Example Usage
//
//	// Start a 72-hour stability test
//	monitor := stability.NewMonitor(stability.Config{
//	    Duration:       72 * time.Hour,
//	    CheckInterval:  30 * time.Second,
//	    MemoryLimit:    500 * 1024 * 1024, // 500MB
//	    MinFPS:         60,
//	    ReportPath:     "stability_report.json",
//	})
//
//	report, err := monitor.Run(serverInstance)
//	if err != nil {
//	    logrus.WithError(err).Fatal("stability test failed")
//	}
//
//	logrus.WithFields(logrus.Fields{
//	    "uptime":        report.TotalUptime,
//	    "crash_count":   report.CrashCount,
//	    "memory_leaks":  report.MemoryLeakCount,
//	}).Info("stability test completed")
package stability
