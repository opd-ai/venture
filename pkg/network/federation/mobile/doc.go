// Package mobile provides federation support for mobile devices (phones/tablets as servers).
//
// # Overview
//
// Mobile federation enables phones and tablets to act as federated servers, supporting
// battery-aware synchronization, background tasks, and mobile-optimized protocols.
//
// # Key Features
//
//   - Mobile devices as federated servers
//   - Battery-aware federation (reduced sync when low battery)
//   - Background federation sync (iOS/Android background tasks)
//   - Mobile-optimized protocol (reduced bandwidth, longer timeouts)
//   - Wake-on-demand for incoming connections
//
// # Battery Optimization
//
// The system adjusts sync frequency based on battery level:
//   - Normal (>50%): 1-minute intervals, full features
//   - Low (20-50%): 5-minute intervals, reduced features
//   - Critical (<20%): 15-minute intervals, minimal sync
//
// # Background Sync
//
// Background tasks are scheduled based on platform:
//   - iOS: Uses BGTaskScheduler with 15-minute minimum interval
//   - Android: Uses WorkManager with 15-minute minimum interval
//   - Both: Exponential backoff on failures
//
// # Usage Example
//
//	// Create mobile adapter
//	adapter := mobile.NewAdapter(&mobile.Config{
//	    DeviceID:          "phone-123",
//	    Platform:          mobile.PlatformAndroid,
//	    EnableBackground:  true,
//	    BatteryThreshold:  0.5, // Start reducing sync at 50%
//	    SyncInterval:      60 * time.Second,
//	})
//
//	// Start federation
//	if err := adapter.Start(); err != nil {
//	    log.Fatalf("Failed to start: %v", err)
//	}
//	defer adapter.Stop()
//
//	// Register sync handler
//	adapter.RegisterSyncHandler(func(ctx context.Context) error {
//	    // Perform federation sync
//	    return nil
//	})
//
//	// Battery level changes are handled automatically
//	adapter.UpdateBatteryLevel(0.3) // Switches to low-battery mode
//
// # Performance Targets
//
//   - Battery consumption: <5% per hour (idle federation)
//   - Background sync: 5-minute intervals (low battery), 1-minute (normal)
//   - Wake-on-demand latency: <10s
//
// # Thread Safety
//
// All public methods are thread-safe and use sync.RWMutex for concurrent access.
package mobile
