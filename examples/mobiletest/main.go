package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/mobile"
)

var (
	deviceID     = flag.String("device", "mobile-test", "Device identifier")
	platform     = flag.String("platform", "android", "Platform: ios, android")
	interval     = flag.Duration("interval", 60*time.Second, "Base sync interval")
	batteryLevel = flag.Float64("battery", 1.0, "Initial battery level (0.0-1.0)")
	simulate     = flag.Bool("simulate", false, "Simulate battery drain and sync operations")
	duration     = flag.Duration("duration", 30*time.Second, "Simulation duration")
)

func main() {
	flag.Parse()

	// Parse platform
	var plat mobile.Platform
	switch *platform {
	case "ios":
		plat = mobile.PlatformIOS
	case "android":
		plat = mobile.PlatformAndroid
	default:
		plat = mobile.PlatformUnknown
	}

	// Create adapter configuration
	config := &mobile.Config{
		DeviceID:          *deviceID,
		Platform:          plat,
		EnableBackground:  true,
		BatteryThreshold:  0.5,
		SyncInterval:      *interval,
		MaxBandwidth:      0,
		TimeoutMultiplier: 2.0,
	}

	// Create adapter
	adapter := mobile.NewAdapter(config)

	// Register sync handler
	syncCount := 0
	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		syncCount++
		fmt.Printf("[SYNC] Operation %d completed\n", syncCount)
		return nil
	})

	// Set initial battery level
	adapter.UpdateBatteryLevel(*batteryLevel)

	fmt.Printf("Mobile Federation Test\n")
	fmt.Printf("======================\n")
	fmt.Printf("Device: %s (%s)\n", config.DeviceID, config.Platform)
	fmt.Printf("Battery: %.0f%%\n", *batteryLevel*100)
	fmt.Printf("Sync Interval: %v\n", config.SyncInterval)
	fmt.Printf("\n")

	// Start adapter
	if err := adapter.Start(); err != nil {
		log.Fatalf("Failed to start adapter: %v", err)
	}
	defer adapter.Stop()

	fmt.Println("Adapter started successfully")

	if *simulate {
		runSimulation(adapter, *duration)
	} else {
		runInteractive(adapter)
	}
}

func runSimulation(adapter *mobile.Adapter, duration time.Duration) {
	fmt.Printf("\nRunning simulation for %v\n\n", duration)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	currentBattery := adapter.GetState().BatteryLevel

	for {
		select {
		case <-ticker.C:
			// Simulate battery drain (1% per 5 seconds)
			currentBattery -= 0.01
			if currentBattery < 0 {
				currentBattery = 0
			}
			adapter.UpdateBatteryLevel(currentBattery)

			// Print status
			printStatus(adapter)

			// Check if simulation complete
			if time.Since(startTime) >= duration {
				fmt.Println("\nSimulation complete")
				printFinalStats(adapter)
				return
			}
		}
	}
}

func runInteractive(adapter *mobile.Adapter) {
	fmt.Println("\nInteractive mode - Press Ctrl+C to stop")
	fmt.Println("Commands:")
	fmt.Println("  (Status updates every 10 seconds)")
	fmt.Println()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Print initial status
	printStatus(adapter)

	for {
		select {
		case <-sigChan:
			fmt.Println("\nShutting down...")
			printFinalStats(adapter)
			return
		case <-ticker.C:
			printStatus(adapter)
		}
	}
}

func printStatus(adapter *mobile.Adapter) {
	state := adapter.GetState()

	fmt.Printf("[%s] Battery: %.0f%% (%s) | Status: %s | Syncs: %d | BG: %d | Errors: %d\n",
		time.Now().Format("15:04:05"),
		state.BatteryLevel*100,
		state.BatteryMode,
		state.SyncStatus,
		state.SyncCount,
		state.BackgroundCount,
		state.SyncErrors,
	)
}

func printFinalStats(adapter *mobile.Adapter) {
	state := adapter.GetState()
	bytesSent, bytesReceived, syncCount, bgCount := state.GetStats()

	fmt.Println("\n=== Final Statistics ===")
	fmt.Printf("Battery Level: %.0f%% (%s)\n", state.BatteryLevel*100, state.BatteryMode)
	fmt.Printf("Sync Status: %s\n", state.SyncStatus)
	fmt.Printf("Total Syncs: %d\n", syncCount)
	fmt.Printf("Background Syncs: %d\n", bgCount)
	fmt.Printf("Sync Errors: %d\n", state.SyncErrors)
	fmt.Printf("Bytes Sent: %d\n", bytesSent)
	fmt.Printf("Bytes Received: %d\n", bytesReceived)

	if !state.LastSyncTime.IsZero() {
		fmt.Printf("Last Sync: %s\n", state.LastSyncTime.Format("15:04:05"))
	}

	fmt.Println("========================")
}
