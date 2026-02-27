// Voice Chat Integration Example
//
// This example demonstrates how to integrate the voice codec system
// with the existing voice channel and spatial voice systems.

package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/engine"
)

// ExampleVoiceTransport is a simple transport implementation for demonstration.
type ExampleVoiceTransport struct {
	sendQueue    []VoicePacket
	receiveQueue []VoicePacket
}

type VoicePacket struct {
	channelID string
	senderID  string
	data      []byte
}

func NewExampleVoiceTransport() *ExampleVoiceTransport {
	return &ExampleVoiceTransport{
		sendQueue:    make([]VoicePacket, 0),
		receiveQueue: make([]VoicePacket, 0),
	}
}

func (t *ExampleVoiceTransport) SendVoice(channelID string, data []byte) error {
	t.sendQueue = append(t.sendQueue, VoicePacket{
		channelID: channelID,
		senderID:  "local-player",
		data:      data,
	})
	return nil
}

func (t *ExampleVoiceTransport) ReceiveVoice() (string, string, []byte, bool) {
	if len(t.receiveQueue) == 0 {
		return "", "", nil, false
	}
	packet := t.receiveQueue[0]
	t.receiveQueue = t.receiveQueue[1:]
	return packet.channelID, packet.senderID, packet.data, true
}

// SetSpatialParams is a no-op in this simplified example.
//
// In a real implementation, spatial parameters would be used for:
//   - Adjusting voice quality based on distance (lower bitrate for distant players)
//   - Implementing voice priority queues (important speakers get more bandwidth)
//   - Client-side volume/pan adjustment before playback (not during encoding)
//
// This example focuses on the encoding/transmission side only. See doc.go
// "What Real Implementation Needs" section for production requirements.
func (t *ExampleVoiceTransport) SetSpatialParams(volume, pan float64) {
	// Intentionally empty - spatial params are applied client-side during playback,
	// not during encoding/transmission in this simplified example.
}

func main() {
	fmt.Println("Voice Chat Integration Example")
	fmt.Println("==============================")

	// Step 1: Initialize audio manager
	fmt.Println("\n1. Initializing audio manager...")
	// Sample rate: 44100 Hz (CD quality, standard for voice chat)
	// Seed: 12345 (fixed seed for reproducible audio synthesis demos)
	audioManager := audio.NewManager(44100, 12345)
	audioManager.SetMasterVolume(1.0)
	audioManager.SetVoiceVolume(0.8)

	// Step 2: Initialize voice codec with transport
	fmt.Println("2. Setting up voice codec...")
	transport := NewExampleVoiceTransport()
	err := audioManager.InitializeVoice(audio.VoiceQualityMedium, transport)
	if err != nil {
		log.Fatal(err)
	}

	if !audioManager.IsVoiceEnabled() {
		log.Fatal("Voice chat not enabled")
	}
	fmt.Printf("   Voice codec: %d Hz, %d kbps\n",
		audioManager.GetVoiceCodec().GetSampleRate(),
		audioManager.GetVoiceCodec().GetBitrate()/1000)

	// Step 3: Create ECS world and voice systems
	fmt.Println("3. Creating voice systems...")
	world := engine.NewWorld()
	voiceChannelSys := engine.NewVoiceChannelSystem(world)
	spatialVoiceSys := engine.NewSpatialVoiceSystem(world)
	voiceAudioSys := engine.NewVoiceAudioSystem(world)

	// Step 4: Create player entities
	fmt.Println("4. Creating player entities...")
	localPlayer := world.CreateEntity()
	remotePlayer := world.CreateEntity()

	// Add position components
	localPlayer.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	remotePlayer.AddComponent(&engine.PositionComponent{X: 150, Y: 120})

	// Step 5: Set up voice components for local player
	fmt.Println("5. Setting up local player voice...")

	// Voice audio component (input/output settings)
	voiceAudioComp := engine.NewVoiceAudioComponent()
	voiceAudioComp.InputMode = engine.VoiceInputVoiceActivity
	voiceAudioComp.VoiceThreshold = 0.1 // 10% input level triggers voice activity
	localPlayer.AddComponent(voiceAudioComp)

	// Spatial voice component (for proximity chat)
	spatialComp := engine.NewSpatialVoiceComponent()
	// Min range: 50 units (clear audio), Max range: 500 units (barely audible)
	spatialComp.SetRange(50.0, 500.0)
	spatialComp.SetFalloffCurve(engine.VoiceFalloffLogarithmic)
	localPlayer.AddComponent(spatialComp)

	// Set local player as listener
	spatialVoiceSys.SetListener(localPlayer)

	// Step 6: Join voice channel
	fmt.Println("6. Joining party voice channel...")
	err = voiceChannelSys.JoinPartyChannel(localPlayer, "party-123", true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("   Joined channel: party:party-123")

	// Step 7: Simulate voice input
	fmt.Println("7. Simulating voice input...")
	processor := audioManager.GetVoiceProcessor()

	// Generate test audio samples (sine wave)
	frameSize := audioManager.GetVoiceCodec().GetFrameSize()
	testSamples := make([]float64, frameSize)
	for i := range testSamples {
		testSamples[i] = 0.3 // 30% input level (above 10% threshold)
	}

	// Simulate voice activity detected
	err = voiceAudioSys.SimulateInput(localPlayer, 0.3)
	if err != nil {
		log.Fatal(err)
	}

	// Process input (encodes and transmits)
	err = processor.ProcessInput("party:party-123", testSamples)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Encoded and queued %d samples\n", len(testSamples))

	// Check if transmitting
	if voiceAudioSys.IsTransmitting(localPlayer) {
		fmt.Println("   ✓ Local player is transmitting")
	}

	// Step 7b: Demonstrate receiving and decoding voice (output side)
	fmt.Println("   Simulating voice reception...")

	// In a real implementation, this would be called in your game loop
	// to process incoming voice packets from the network and decode them
	outputSamples, err := processor.ProcessOutput()
	if err != nil {
		log.Fatal(err)
	}

	// Display received audio for each channel
	for channelID, samples := range outputSamples {
		fmt.Printf("   Received %d samples from channel: %s\n", len(samples), channelID)
		// In production: mix these samples with spatial audio parameters (volume, pan)
		// and write to audio output device or Web Audio API
	}

	if len(outputSamples) == 0 {
		fmt.Println("   No incoming voice packets (expected in this demo)")
		fmt.Println("   In production: ProcessOutput() retrieves decoded samples from network")
	}

	// Step 8: Demonstrate spatial audio
	fmt.Println("8. Testing spatial audio...")

	// Add spatial component to remote player
	remoteSpatial := engine.NewSpatialVoiceComponent()
	remotePlayer.AddComponent(remoteSpatial)

	// Update spatial system
	entities := []*engine.Entity{localPlayer, remotePlayer}
	spatialVoiceSys.Update(entities, 0.016) // 16ms frame

	volume := spatialVoiceSys.GetVolumeForEntity(remotePlayer)
	pan := spatialVoiceSys.GetPanForEntity(remotePlayer)
	distance := spatialVoiceSys.GetDistanceToEntity(remotePlayer)

	fmt.Printf("   Distance: %.1f units\n", distance)
	fmt.Printf("   Volume: %.2f\n", volume)
	fmt.Printf("   Pan: %.2f (%.2f = left, 0 = center, %.2f = right)\n", pan, -1.0, 1.0)
	fmt.Printf("   Audible: %v\n", spatialVoiceSys.IsEntityAudible(remotePlayer))

	// Step 9: Display channel information
	fmt.Println("9. Voice channel information...")
	if channel, ok := voiceChannelSys.GetChannel("party:party-123"); ok {
		fmt.Printf("   Channel type: %s\n", channel.Type)
		fmt.Printf("   Participants: %d\n", len(channel.Participants))
	}

	fmt.Println("\n✓ Voice chat integration successful!")
	fmt.Println("\nNext steps:")
	fmt.Println("- Implement actual microphone input")
	fmt.Println("- Connect network transport to real sockets")
	fmt.Println("- Add audio playback for received voice")
	fmt.Println("- Implement push-to-talk keybinding")
	fmt.Println("- Add voice activity visualization")
}
