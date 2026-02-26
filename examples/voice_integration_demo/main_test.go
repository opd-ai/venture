package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/engine"
)

// TestExampleRuns verifies the example code doesn't panic on setup.
// This catches API drift when voice systems change.
func TestExampleRuns(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "audio manager initialization",
			fn: func(t *testing.T) {
				manager := audio.NewManager(44100, 12345)
				manager.SetMasterVolume(1.0)
				manager.SetVoiceVolume(0.8)
				if manager == nil {
					t.Fatal("audio manager should not be nil")
				}
			},
		},
		{
			name: "voice codec initialization",
			fn: func(t *testing.T) {
				manager := audio.NewManager(44100, 12345)
				transport := NewExampleVoiceTransport()
				err := manager.InitializeVoice(audio.VoiceQualityMedium, transport)
				if err != nil {
					t.Fatalf("InitializeVoice failed: %v", err)
				}
				if !manager.IsVoiceEnabled() {
					t.Fatal("voice should be enabled after initialization")
				}
			},
		},
		{
			name: "voice systems creation",
			fn: func(t *testing.T) {
				world := engine.NewWorld()
				voiceChannelSys := engine.NewVoiceChannelSystem(world)
				spatialVoiceSys := engine.NewSpatialVoiceSystem(world)
				voiceAudioSys := engine.NewVoiceAudioSystem(world)

				if voiceChannelSys == nil {
					t.Fatal("voiceChannelSystem should not be nil")
				}
				if spatialVoiceSys == nil {
					t.Fatal("spatialVoiceSystem should not be nil")
				}
				if voiceAudioSys == nil {
					t.Fatal("voiceAudioSystem should not be nil")
				}
			},
		},
		{
			name: "voice components creation",
			fn: func(t *testing.T) {
				world := engine.NewWorld()
				player := world.CreateEntity()

				// VoiceAudioComponent
				voiceAudioComp := engine.NewVoiceAudioComponent()
				voiceAudioComp.InputMode = engine.VoiceInputVoiceActivity
				voiceAudioComp.VoiceThreshold = 0.1
				player.AddComponent(voiceAudioComp)

				// SpatialVoiceComponent
				spatialComp := engine.NewSpatialVoiceComponent()
				spatialComp.SetRange(50.0, 500.0)
				spatialComp.SetFalloffCurve(engine.VoiceFalloffLogarithmic)
				player.AddComponent(spatialComp)

				if !player.HasComponent("voice_audio") {
					t.Fatal("player should have voice_audio component")
				}
				if !player.HasComponent("spatial_voice") {
					t.Fatal("player should have spatial_voice component")
				}
			},
		},
		{
			name: "voice channel join",
			fn: func(t *testing.T) {
				world := engine.NewWorld()
				voiceChannelSys := engine.NewVoiceChannelSystem(world)
				player := world.CreateEntity()

				err := voiceChannelSys.JoinPartyChannel(player, "party-123", true)
				if err != nil {
					t.Fatalf("JoinPartyChannel failed: %v", err)
				}

				channel, ok := voiceChannelSys.GetChannel("party:party-123")
				if !ok {
					t.Fatal("channel should exist after join")
				}
				if len(channel.Participants) != 1 {
					t.Fatalf("expected 1 participant, got %d", len(channel.Participants))
				}
			},
		},
		{
			name: "voice input simulation",
			fn: func(t *testing.T) {
				world := engine.NewWorld()
				voiceAudioSys := engine.NewVoiceAudioSystem(world)
				player := world.CreateEntity()

				// Add voice audio component with voice activity mode
				voiceAudioComp := engine.NewVoiceAudioComponent()
				voiceAudioComp.InputMode = engine.VoiceInputVoiceActivity
				voiceAudioComp.VoiceThreshold = 0.1
				player.AddComponent(voiceAudioComp)

				// Simulate input
				err := voiceAudioSys.SimulateInput(player, 0.3)
				if err != nil {
					t.Fatalf("SimulateInput failed: %v", err)
				}

				// Update system to process voice activity
				entities := []*engine.Entity{player}
				voiceAudioSys.Update(entities, 0.016)

				if !voiceAudioSys.IsTransmitting(player) {
					t.Fatal("player should be transmitting after simulated input")
				}
			},
		},
		{
			name: "spatial audio calculation",
			fn: func(t *testing.T) {
				world := engine.NewWorld()
				spatialVoiceSys := engine.NewSpatialVoiceSystem(world)

				listener := world.CreateEntity()
				speaker := world.CreateEntity()

				listener.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
				speaker.AddComponent(&engine.PositionComponent{X: 150, Y: 120})

				spatialComp := engine.NewSpatialVoiceComponent()
				speaker.AddComponent(spatialComp)

				spatialVoiceSys.SetListener(listener)

				entities := []*engine.Entity{listener, speaker}
				spatialVoiceSys.Update(entities, 0.016)

				volume := spatialVoiceSys.GetVolumeForEntity(speaker)
				pan := spatialVoiceSys.GetPanForEntity(speaker)
				distance := spatialVoiceSys.GetDistanceToEntity(speaker)

				// Basic sanity checks
				if volume < 0 || volume > 1 {
					t.Fatalf("volume should be in [0,1], got %.2f", volume)
				}
				if pan < -1 || pan > 1 {
					t.Fatalf("pan should be in [-1,1], got %.2f", pan)
				}
				if distance < 0 {
					t.Fatalf("distance should be non-negative, got %.2f", distance)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each test should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("test panicked: %v", r)
				}
			}()
			tt.fn(t)
		})
	}
}

// TestExampleVoiceTransport tests the example transport implementation.
func TestExampleVoiceTransport(t *testing.T) {
	transport := NewExampleVoiceTransport()

	// Test send
	testData := []byte{1, 2, 3, 4}
	err := transport.SendVoice("test-channel", testData)
	if err != nil {
		t.Fatalf("SendVoice failed: %v", err)
	}

	if len(transport.sendQueue) != 1 {
		t.Fatalf("expected 1 packet in queue, got %d", len(transport.sendQueue))
	}

	// Test receive (should be empty initially)
	_, _, _, ok := transport.ReceiveVoice()
	if ok {
		t.Fatal("should have no packets to receive")
	}

	// Add packet to receive queue
	transport.receiveQueue = append(transport.receiveQueue, VoicePacket{
		channelID: "test-channel",
		senderID:  "remote-player",
		data:      testData,
	})

	// Test receive
	channelID, senderID, data, ok := transport.ReceiveVoice()
	if !ok {
		t.Fatal("should have packet to receive")
	}
	if channelID != "test-channel" {
		t.Fatalf("expected channelID 'test-channel', got '%s'", channelID)
	}
	if senderID != "remote-player" {
		t.Fatalf("expected senderID 'remote-player', got '%s'", senderID)
	}
	if len(data) != len(testData) {
		t.Fatalf("expected data length %d, got %d", len(testData), len(data))
	}

	// Queue should be empty now
	_, _, _, ok = transport.ReceiveVoice()
	if ok {
		t.Fatal("queue should be empty after receiving")
	}
}
