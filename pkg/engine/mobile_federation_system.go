package engine

import (
	"context"
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/mobile"
	"github.com/sirupsen/logrus"
)

// MobileFederationSystem manages mobile-optimized federation
type MobileFederationSystem struct {
	adapter *mobile.Adapter
	logger  *logrus.Entry
}

// NewMobileFederationSystem creates a mobile federation system
func NewMobileFederationSystem(config *mobile.Config) *MobileFederationSystem {
	if config == nil {
		config = mobile.DefaultConfig()
	}

	adapter := mobile.NewAdapter(config)

	// Register sync handler for federation operations
	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		// Implement federation sync logic here
		// This would typically sync player data, guild info, etc.
		// For now, this is a placeholder that returns success
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			return nil
		}
	})

	return &MobileFederationSystem{
		adapter: adapter,
		logger:  logrus.WithField("system", "mobile_federation"),
	}
}

// Start initializes the mobile federation system
func (s *MobileFederationSystem) Start() error {
	s.logger.Debug("Starting mobile federation system")
	return s.adapter.Start()
}

// Stop shuts down the mobile federation system
func (s *MobileFederationSystem) Stop() error {
	s.logger.Debug("Stopping mobile federation system")
	return s.adapter.Stop()
}

// Update performs per-frame updates
func (s *MobileFederationSystem) Update(entities []*Entity, deltaTime float64) {
	// Mobile federation system updates are handled asynchronously
	// This method exists for ECS interface compliance
	// No per-frame operations needed
}

// UpdateBatteryLevel notifies the system of battery level changes
func (s *MobileFederationSystem) UpdateBatteryLevel(level float64) {
	s.logger.WithField("battery_level", level).Debug("Updating battery level")
	s.adapter.UpdateBatteryLevel(level)
}

// GetAdapter returns the underlying mobile adapter
func (s *MobileFederationSystem) GetAdapter() *mobile.Adapter {
	return s.adapter
}

// GetState returns current federation state
func (s *MobileFederationSystem) GetState() mobile.State {
	return s.adapter.GetState()
}

// PauseSync pauses federation sync operations
func (s *MobileFederationSystem) PauseSync() {
	s.logger.Info("Pausing mobile federation sync")
	s.adapter.PauseSync()
}

// ResumeSync resumes federation sync operations
func (s *MobileFederationSystem) ResumeSync() {
	s.logger.Info("Resuming mobile federation sync")
	s.adapter.ResumeSync()
}
