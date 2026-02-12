# Voice Chat System Documentation

## Overview

The Venture voice chat system provides real-time voice communication with spatial audio support, integrated voice codec, and network transport abstraction. It supports party chat, guild chat, proximity-based voice, and private channels.

## Architecture

### Components

1. **Voice Channel System** (`pkg/engine/voice_channel_system.go`)
   - Manages voice channel lifecycle (party, guild, proximity, private)
   - Handles participant synchronization
   - Provides moderation features (mute, deafen, kick)

2. **Spatial Voice System** (`pkg/engine/spatial_voice_system.go`)
   - Calculates distance-based volume attenuation
   - Provides stereo panning based on relative position
   - Supports multiple falloff curves (linear, logarithmic, exponential)

3. **Voice Audio System** (`pkg/engine/voice_audio_system.go`)
   - Processes audio input/output
   - Supports push-to-talk and voice activity detection
   - Manages transmission state with cooldowns

4. **Voice Codec** (`pkg/audio/voice.go`)
   - Encodes/decodes voice data for network transmission
   - Simple ADPCM codec with 2:1 compression (no external dependencies)
   - Pluggable interface for advanced codecs (Opus, etc.)

### Quality Presets

- **Low Quality (8 kbps)**: Minimal bandwidth, suitable for high-latency connections
- **Medium Quality (16 kbps)**: Balanced quality and bandwidth
- **High Quality (32 kbps)**: Best quality for low-latency connections

## Usage

### Initialization

```go
import (
    "github.com/opd-ai/venture/pkg/audio"
    "github.com/opd-ai/venture/pkg/engine"
)

// Initialize audio manager with voice support
audioManager := audio.NewManager(44100, 12345)

// Create transport (implement VoiceTransport interface)
transport := NewMyVoiceTransport()

// Initialize voice codec
err := audioManager.InitializeVoice(audio.VoiceQualityMedium, transport)
if err != nil {
    log.Fatal(err)
}

// Create voice systems
world := engine.NewWorld()
voiceChannel := engine.NewVoiceChannelSystem(world)
spatialVoice := engine.NewSpatialVoiceSystem(world)
voiceAudio := engine.NewVoiceAudioSystem(world)
```

### Joining a Voice Channel

```go
// Join party voice channel
err := voiceChannel.JoinPartyChannel(playerEntity, "party-123", isLeader)

// Join guild voice channel
err := voiceChannel.JoinGuildChannel(playerEntity, "guild-456", "Member")

// Join proximity voice (automatic for nearby players)
proximityComp := engine.NewProximityVoiceChannelComponent()
playerEntity.AddComponent(proximityComp)
```

### Voice Input Processing

```go
// Push-to-talk mode
voiceAudio.SetInputMode(playerEntity, engine.VoiceInputPushToTalk)
voiceAudio.SetPushToTalkKey(playerEntity, "V")

// Activate push-to-talk
voiceAudio.SetPushToTalk(playerEntity, true)

// Voice activity detection mode
voiceAudio.SetInputMode(playerEntity, engine.VoiceInputVoiceActivity)
voiceAudio.SetVoiceThreshold(playerEntity, 0.1) // 10% threshold

// Simulate audio input (in real implementation, from microphone)
voiceAudio.SimulateInput(playerEntity, 0.5) // 50% input level
```

### Spatial Audio Configuration

```go
// Set up spatial voice for proximity chat
spatialComp := engine.NewSpatialVoiceComponent()
spatialComp.SetRange(50.0, 500.0) // Min 50 units, max 500 units
spatialComp.SetFalloffCurve(engine.VoiceFalloffLogarithmic)
playerEntity.AddComponent(spatialComp)

// Set listener (local player)
spatialVoice.SetListener(localPlayerEntity)

// Get spatial audio parameters
volume := spatialVoice.GetVolumeForEntity(remotePlayerEntity)
pan := spatialVoice.GetPanForEntity(remotePlayerEntity)
```

### Voice Encoding/Decoding

```go
// Get voice processor
processor := audioManager.GetVoiceProcessor()

// Process input audio (encodes and sends)
inputSamples := []float64{0.1, 0.2, 0.3, ...}
err := processor.ProcessInput("channel-id", inputSamples)

// Process output audio (receives and decodes)
outputMap, err := processor.ProcessOutput()
for channelSenderKey, samples := range outputMap {
    // channelSenderKey format: "channel-id:sender-id"
    // samples contains decoded audio for playback
}
```

### Volume Control

```go
// Master volume (affects all audio)
audioManager.SetMasterVolume(0.8)

// Voice chat volume
audioManager.SetVoiceVolume(0.7)

// Per-entity output volume
voiceAudio.SetOutputVolume(remotePlayerEntity, 0.9)
```

### Moderation

```go
// Mute a participant (moderator only)
err := voiceChannel.MuteParticipant(moderatorEntity, targetEntity, true)

// Deafen a participant (moderator only)
err := voiceChannel.DeafenParticipant(moderatorEntity, targetEntity, true)

// Kick from channel (moderator only)
err := voiceChannel.KickParticipant(moderatorEntity, targetEntity)
```

## Transport Implementation

Implement the `VoiceTransport` interface to integrate with your networking layer:

```go
type MyVoiceTransport struct {
    // Your network connection
    conn net.Conn
    receiveQueue chan VoicePacket
}

func (t *MyVoiceTransport) SendVoice(channelID string, data []byte) error {
    // Send encoded voice data over network
    packet := VoicePacket{
        ChannelID: channelID,
        Data:      data,
    }
    return t.sendPacket(packet)
}

func (t *MyVoiceTransport) ReceiveVoice() (string, string, []byte, bool) {
    // Receive voice data from network
    select {
    case packet := <-t.receiveQueue:
        return packet.ChannelID, packet.SenderID, packet.Data, true
    default:
        return "", "", nil, false
    }
}

func (t *MyVoiceTransport) SetSpatialParams(volume, pan float64) {
    // Optional: adjust transmission based on spatial parameters
}
```

## Performance Characteristics

### Codec Performance

- **Encoding Speed**: ~5-10 µs per frame (960 samples @ 48kHz)
- **Decoding Speed**: ~3-5 µs per frame
- **Compression Ratio**: 2:1 (4 bits per sample)
- **Memory Usage**: Minimal buffering (frame size × 8 bytes)

### Bandwidth Requirements

| Quality | Bitrate | Frame Size | Bandwidth (with overhead) |
|---------|---------|-----------|---------------------------|
| Low     | 8 kbps  | 320-960   | ~10 kbps                  |
| Medium  | 16 kbps | 320-960   | ~18 kbps                  |
| High    | 32 kbps | 320-960   | ~35 kbps                  |

### Latency

- **Encoding Latency**: <1ms
- **Frame Duration**: 20ms (buffering)
- **Total Pipeline**: Network latency + 20-40ms processing

## Advanced Usage

### Custom Codec Integration

Replace `SimpleVoiceCodec` with a production codec (e.g., Opus):

```go
type OpusCodec struct {
    encoder *opus.Encoder
    decoder *opus.Decoder
}

func (c *OpusCodec) Encode(samples []float64) ([]byte, error) {
    // Convert to int16 and encode with Opus
}

func (c *OpusCodec) Decode(data []byte) ([]float64, error) {
    // Decode with Opus and convert to float64
}

// Register with audio manager
audioManager.SetVoiceCodec(NewOpusCodec(48000, audio.VoiceQualityHigh))
```

### Noise Suppression

```go
// Set noise gate to filter background noise
voiceAudio.SetNoiseGateLevel(playerEntity, 0.05) // 5% threshold

// Adjust input gain for quiet microphones
voiceAudio.SetInputGain(playerEntity, 1.5) // 150% gain
```

### Voice Activity Detection Tuning

```go
// Sensitive (picks up quiet speech)
voiceAudio.SetVoiceThreshold(playerEntity, 0.05)

// Normal
voiceAudio.SetVoiceThreshold(playerEntity, 0.1)

// Aggressive (only loud speech)
voiceAudio.SetVoiceThreshold(playerEntity, 0.3)
```

## Testing

The voice system includes comprehensive tests:

```bash
# Run all voice tests
go test ./pkg/audio/... -v -run Voice

# Run with coverage
go test ./pkg/audio/... -cover

# Benchmarks
go test ./pkg/audio/... -bench=Codec -benchmem
```

## Troubleshooting

### No voice output

1. Check `audioManager.IsVoiceEnabled()` returns true
2. Verify codec and processor are initialized
3. Ensure voice volume > 0
4. Check transport is receiving data

### Poor quality

1. Increase quality preset
2. Check network bandwidth
3. Verify sample rate matches across client/server
4. Monitor packet loss

### High latency

1. Reduce quality preset for lower bitrate
2. Enable network prediction/jitter buffer
3. Check frame size (smaller = lower latency, more overhead)

### Echo/feedback

1. Implement acoustic echo cancellation
2. Use push-to-talk instead of voice activity
3. Lower output volume
4. Enable self-mute when speaking

## See Also

- `pkg/engine/voice_channel_system.go` - Voice channel management
- `pkg/engine/spatial_voice_system.go` - Spatial audio calculations
- `pkg/engine/voice_audio_system.go` - Audio processing
- `pkg/audio/voice.go` - Codec implementation
- `pkg/network/` - Network transport layer
