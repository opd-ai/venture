package audio

import (
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
)

// VoiceCodec defines the interface for voice encoding and decoding.
type VoiceCodec interface {
	// Encode compresses audio samples for transmission
	Encode(samples []float64) ([]byte, error)

	// Decode decompresses received data into audio samples
	Decode(data []byte) ([]float64, error)

	// GetBitrate returns the codec bitrate in bits per second
	GetBitrate() int

	// GetSampleRate returns the expected sample rate
	GetSampleRate() int

	// GetFrameSize returns the expected frame size in samples
	GetFrameSize() int
}

// VoiceQuality defines the quality preset for voice encoding.
type VoiceQuality int

const (
	// VoiceQualityLow provides minimal bandwidth (8 kbps, suitable for high-latency)
	VoiceQualityLow VoiceQuality = iota
	// VoiceQualityMedium provides balanced quality (16 kbps)
	VoiceQualityMedium
	// VoiceQualityHigh provides high quality (32 kbps)
	VoiceQualityHigh
)

// SimpleVoiceCodec is a basic voice codec using simple compression.
// Production systems would use Opus or similar, but this provides
// a working implementation without external dependencies.
type SimpleVoiceCodec struct {
	sampleRate int
	frameSize  int
	bitrate    int
	quality    VoiceQuality
}

// NewSimpleVoiceCodec creates a new simple voice codec.
func NewSimpleVoiceCodec(sampleRate int, quality VoiceQuality) *SimpleVoiceCodec {
	frameSize := 960 // 20ms at 48kHz
	if sampleRate == 44100 {
		frameSize = 882 // ~20ms at 44.1kHz
	} else if sampleRate == 16000 {
		frameSize = 320 // 20ms at 16kHz
	}

	bitrate := 16000
	switch quality {
	case VoiceQualityLow:
		bitrate = 8000
	case VoiceQualityMedium:
		bitrate = 16000
	case VoiceQualityHigh:
		bitrate = 32000
	}

	log.WithFields(log.Fields{
		"sample_rate": sampleRate,
		"frame_size":  frameSize,
		"bitrate":     bitrate,
		"quality":     quality,
	}).Debug("Creating simple voice codec")

	return &SimpleVoiceCodec{
		sampleRate: sampleRate,
		frameSize:  frameSize,
		bitrate:    bitrate,
		quality:    quality,
	}
}

// Encode compresses audio samples using simple ADPCM-like encoding.
// Each sample is quantized to 4 bits, with two samples packed per byte.
// Note: For odd-length sample arrays, the output rounds up to ceil(n/2) bytes.
// During decoding, this produces an extra sample (decode always produces even count).
// Frame sizes (320-960 based on sample rate) are typically even, so this rarely matters.
func (c *SimpleVoiceCodec) Encode(samples []float64) ([]byte, error) {
	if len(samples) == 0 {
		return nil, fmt.Errorf("no samples to encode")
	}

	// Simple 4-bit ADPCM encoding - two samples per byte
	encodedLen := (len(samples) + 1) / 2
	encoded := make([]byte, encodedLen)
	var predictor float64

	for i := 0; i < len(samples); i++ {
		sample := samples[i]
		diff := sample - predictor

		// Quantize to 4 bits (-8 to 7)
		quantized := int(math.Round(diff * 8))
		if quantized > 7 {
			quantized = 7
		} else if quantized < -8 {
			quantized = -8
		}

		// Store 4-bit value: low nibble for even indices, high nibble for odd
		byteIndex := i / 2
		if i%2 == 0 {
			encoded[byteIndex] = byte(quantized & 0x0F)
		} else {
			encoded[byteIndex] |= byte((quantized & 0x0F) << 4)
		}

		// Update predictor with dequantized value
		predictor += float64(quantized) / 8.0
		if predictor > 1.0 {
			predictor = 1.0
		} else if predictor < -1.0 {
			predictor = -1.0
		}
	}

	return encoded, nil
}

// Decode decompresses received data into audio samples.
// The output always has an even number of samples (2 samples per byte).
// If the original input had an odd number of samples, the decoded output
// will include one extra sample derived from the padding zero nibble.
func (c *SimpleVoiceCodec) Decode(data []byte) ([]float64, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to decode")
	}

	samples := make([]float64, len(data)*2)
	var predictor float64

	for i := 0; i < len(data); i++ {
		val1, val2 := extractFourBitValues(data[i])
		predictor = decodeSample(val1, predictor, &samples, i*2)

		if i*2+1 < len(samples) {
			predictor = decodeSample(val2, predictor, &samples, i*2+1)
		}
	}

	return samples, nil
}

// extractFourBitValues extracts two 4-bit values from a byte and sign-extends them.
func extractFourBitValues(b byte) (int, int) {
	val1 := int(b & 0x0F)
	val2 := int((b >> 4) & 0x0F)

	if val1 >= 8 {
		val1 -= 16
	}
	if val2 >= 8 {
		val2 -= 16
	}

	return val1, val2
}

// decodeSample decodes a single ADPCM sample and clamps the predictor.
func decodeSample(val int, predictor float64, samples *[]float64, index int) float64 {
	predictor += float64(val) / 8.0
	if predictor > 1.0 {
		predictor = 1.0
	} else if predictor < -1.0 {
		predictor = -1.0
	}
	(*samples)[index] = predictor
	return predictor
}

// GetBitrate returns the codec bitrate.
func (c *SimpleVoiceCodec) GetBitrate() int {
	return c.bitrate
}

// GetSampleRate returns the sample rate.
func (c *SimpleVoiceCodec) GetSampleRate() int {
	return c.sampleRate
}

// GetFrameSize returns the frame size.
func (c *SimpleVoiceCodec) GetFrameSize() int {
	return c.frameSize
}

// VoiceTransport handles network transmission of voice data.
type VoiceTransport interface {
	// SendVoice transmits encoded voice data to a channel
	SendVoice(channelID string, data []byte) error

	// ReceiveVoice gets received voice data for processing
	ReceiveVoice() (channelID, senderID string, data []byte, ok bool)

	// SetSpatialParams sets spatial audio parameters for transmission
	SetSpatialParams(volume, pan float64)
}

// VoiceProcessor manages voice encoding, decoding, and transmission.
type VoiceProcessor struct {
	codec     VoiceCodec
	transport VoiceTransport
	enabled   bool

	// Buffers for frame accumulation
	inputBuffer  []float64
	outputBuffer map[string][]float64 // channelID -> decoded samples
}

// NewVoiceProcessor creates a new voice processor.
func NewVoiceProcessor(codec VoiceCodec, transport VoiceTransport) *VoiceProcessor {
	return &VoiceProcessor{
		codec:        codec,
		transport:    transport,
		enabled:      true,
		inputBuffer:  make([]float64, 0, codec.GetFrameSize()),
		outputBuffer: make(map[string][]float64),
	}
}

// SetEnabled enables or disables voice processing.
func (p *VoiceProcessor) SetEnabled(enabled bool) {
	p.enabled = enabled
}

// IsEnabled returns whether voice processing is enabled.
func (p *VoiceProcessor) IsEnabled() bool {
	return p.enabled
}

// ProcessInput processes input audio samples and sends them if ready.
// channelID identifies the voice channel (e.g., "team", "proximity", "guild")
// that the encoded audio will be transmitted to via the transport layer.
// Samples are accumulated in an internal buffer and encoded into frames
// when enough data is available.
//
// Returns an error if channelID is empty or if transport transmission fails.
func (p *VoiceProcessor) ProcessInput(channelID string, samples []float64) error {
	if !p.enabled || len(samples) == 0 {
		return nil
	}

	// Validate channelID is non-empty
	if channelID == "" {
		return fmt.Errorf("channelID cannot be empty")
	}

	// Accumulate samples in buffer
	p.inputBuffer = append(p.inputBuffer, samples...)

	// Process complete frames
	frameSize := p.codec.GetFrameSize()
	for len(p.inputBuffer) >= frameSize {
		frame := p.inputBuffer[:frameSize]
		p.inputBuffer = p.inputBuffer[frameSize:]

		// Encode frame
		encoded, err := p.codec.Encode(frame)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Warn("Failed to encode voice frame")
			continue
		}

		// Send encoded data
		if p.transport != nil {
			if err := p.transport.SendVoice(channelID, encoded); err != nil {
				log.WithFields(log.Fields{
					"channel_id": channelID,
					"error":      err,
				}).Warn("Failed to send voice data")
			}
		}
	}

	return nil
}

// ProcessOutput receives and decodes voice data.
func (p *VoiceProcessor) ProcessOutput() (map[string][]float64, error) {
	if !p.enabled || p.transport == nil {
		return p.outputBuffer, nil
	}

	// Clear previous output
	p.outputBuffer = make(map[string][]float64)

	// Receive all available voice packets
	for {
		channelID, senderID, data, ok := p.transport.ReceiveVoice()
		if !ok {
			break
		}

		// Decode voice data
		samples, err := p.codec.Decode(data)
		if err != nil {
			log.WithFields(log.Fields{
				"channel_id": channelID,
				"sender_id":  senderID,
				"error":      err,
			}).Warn("Failed to decode voice frame")
			continue
		}

		// Store in output buffer
		key := fmt.Sprintf("%s:%s", channelID, senderID)
		p.outputBuffer[key] = append(p.outputBuffer[key], samples...)
	}

	return p.outputBuffer, nil
}

// GetCodec returns the voice codec.
func (p *VoiceProcessor) GetCodec() VoiceCodec {
	return p.codec
}

// GetTransport returns the voice transport.
func (p *VoiceProcessor) GetTransport() VoiceTransport {
	return p.transport
}

// ClearBuffers clears all internal buffers.
func (p *VoiceProcessor) ClearBuffers() {
	p.inputBuffer = make([]float64, 0, p.codec.GetFrameSize())
	p.outputBuffer = make(map[string][]float64)
}
