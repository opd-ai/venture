package audio

import (
	"testing"
)

func TestSimpleVoiceCodec_Creation(t *testing.T) {
	tests := []struct {
		name        string
		sampleRate  int
		quality     VoiceQuality
		wantFrame   int
		wantBitrate int
	}{
		{
			name:        "48kHz high quality",
			sampleRate:  48000,
			quality:     VoiceQualityHigh,
			wantFrame:   960,
			wantBitrate: 32000,
		},
		{
			name:        "44.1kHz medium quality",
			sampleRate:  44100,
			quality:     VoiceQualityMedium,
			wantFrame:   882,
			wantBitrate: 16000,
		},
		{
			name:        "16kHz low quality",
			sampleRate:  16000,
			quality:     VoiceQualityLow,
			wantFrame:   320,
			wantBitrate: 8000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codec := NewSimpleVoiceCodec(tt.sampleRate, tt.quality)
			if codec == nil {
				t.Fatal("codec is nil")
			}
			if codec.GetSampleRate() != tt.sampleRate {
				t.Errorf("GetSampleRate() = %d, want %d", codec.GetSampleRate(), tt.sampleRate)
			}
			if codec.GetFrameSize() != tt.wantFrame {
				t.Errorf("GetFrameSize() = %d, want %d", codec.GetFrameSize(), tt.wantFrame)
			}
			if codec.GetBitrate() != tt.wantBitrate {
				t.Errorf("GetBitrate() = %d, want %d", codec.GetBitrate(), tt.wantBitrate)
			}
		})
	}
}

func TestSimpleVoiceCodec_EncodeEmpty(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	_, err := codec.Encode([]float64{})
	if err == nil {
		t.Error("expected error encoding empty samples")
	}
}

func TestSimpleVoiceCodec_DecodeEmpty(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	_, err := codec.Decode([]byte{})
	if err == nil {
		t.Error("expected error decoding empty data")
	}
}

func TestSimpleVoiceCodec_EncodeDecode(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityHigh)

	tests := []struct {
		name    string
		samples []float64
	}{
		{
			name:    "silence",
			samples: make([]float64, 100),
		},
		{
			name: "constant tone",
			samples: func() []float64 {
				s := make([]float64, 100)
				for i := range s {
					s[i] = 0.5
				}
				return s
			}(),
		},
		{
			name: "varying amplitude",
			samples: func() []float64 {
				s := make([]float64, 100)
				for i := range s {
					s[i] = float64(i) / 100.0
				}
				return s
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := codec.Encode(tt.samples)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if len(encoded) == 0 {
				t.Fatal("encoded data is empty")
			}

			decoded, err := codec.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(decoded) == 0 {
				t.Fatal("decoded samples are empty")
			}

			// Check length (should be within reasonable bounds)
			if len(decoded) < len(tt.samples)-2 || len(decoded) > len(tt.samples)+2 {
				t.Errorf("decoded length = %d, want ~%d", len(decoded), len(tt.samples))
			}

			// Check for reasonable reconstruction (lossy codec)
			maxError := 0.2 // Allow 20% error for simple codec
			for i := 0; i < len(tt.samples) && i < len(decoded); i++ {
				diff := tt.samples[i] - decoded[i]
				if diff < 0 {
					diff = -diff
				}
				if diff > maxError {
					t.Errorf("sample[%d] error = %f, want < %f", i, diff, maxError)
					break
				}
			}
		})
	}
}

func TestSimpleVoiceCodec_ClampingBehavior(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)

	samples := []float64{-2.0, -1.5, -1.0, 0.0, 1.0, 1.5, 2.0}
	encoded, err := codec.Encode(samples)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// All decoded samples should be clamped to [-1.0, 1.0]
	for i, sample := range decoded {
		if sample < -1.0 || sample > 1.0 {
			t.Errorf("decoded[%d] = %f, should be clamped to [-1.0, 1.0]", i, sample)
		}
	}
}

func TestSimpleVoiceCodec_Compression(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)

	samples := make([]float64, 1000)
	for i := range samples {
		samples[i] = 0.5
	}

	encoded, err := codec.Encode(samples)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Should achieve at least 2:1 compression (8 bytes per sample -> 4 bits)
	expectedMaxSize := len(samples) * 8 / 2
	if len(encoded) > expectedMaxSize {
		t.Errorf("encoded size = %d bytes, want <= %d bytes", len(encoded), expectedMaxSize)
	}
}

type mockTransport struct {
	sentData     map[string][]byte
	receiveQueue []voicePacket
	volume       float64
	pan          float64
}

type voicePacket struct {
	channelID string
	senderID  string
	data      []byte
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		sentData:     make(map[string][]byte),
		receiveQueue: make([]voicePacket, 0),
		volume:       1.0,
		pan:          0.0,
	}
}

func (m *mockTransport) SendVoice(channelID string, data []byte) error {
	m.sentData[channelID] = append(m.sentData[channelID], data...)
	return nil
}

func (m *mockTransport) ReceiveVoice() (string, string, []byte, bool) {
	if len(m.receiveQueue) == 0 {
		return "", "", nil, false
	}
	packet := m.receiveQueue[0]
	m.receiveQueue = m.receiveQueue[1:]
	return packet.channelID, packet.senderID, packet.data, true
}

func (m *mockTransport) SetSpatialParams(volume, pan float64) {
	m.volume = volume
	m.pan = pan
}

func (m *mockTransport) addReceivePacket(channelID, senderID string, data []byte) {
	m.receiveQueue = append(m.receiveQueue, voicePacket{
		channelID: channelID,
		senderID:  senderID,
		data:      data,
	})
}

func TestVoiceProcessor_Creation(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	transport := newMockTransport()
	processor := NewVoiceProcessor(codec, transport)

	if processor == nil {
		t.Fatal("processor is nil")
	}
	if !processor.IsEnabled() {
		t.Error("processor should be enabled by default")
	}
	if processor.GetCodec() != codec {
		t.Error("codec mismatch")
	}
	if processor.GetTransport() != transport {
		t.Error("transport mismatch")
	}
}

func TestVoiceProcessor_EnableDisable(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	processor := NewVoiceProcessor(codec, nil)

	processor.SetEnabled(false)
	if processor.IsEnabled() {
		t.Error("processor should be disabled")
	}

	processor.SetEnabled(true)
	if !processor.IsEnabled() {
		t.Error("processor should be enabled")
	}
}

func TestVoiceProcessor_ProcessInputDisabled(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	transport := newMockTransport()
	processor := NewVoiceProcessor(codec, transport)

	processor.SetEnabled(false)

	samples := make([]float64, 100)
	err := processor.ProcessInput("test-channel", samples)
	if err != nil {
		t.Errorf("ProcessInput() error = %v", err)
	}

	// Should not send anything when disabled
	if len(transport.sentData) > 0 {
		t.Error("should not send data when disabled")
	}
}

func TestVoiceProcessor_ProcessInputEmpty(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	transport := newMockTransport()
	processor := NewVoiceProcessor(codec, transport)

	err := processor.ProcessInput("test-channel", []float64{})
	if err != nil {
		t.Errorf("ProcessInput() error = %v", err)
	}
}

func TestVoiceProcessor_ProcessInputFullFrame(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	transport := newMockTransport()
	processor := NewVoiceProcessor(codec, transport)

	frameSize := codec.GetFrameSize()
	samples := make([]float64, frameSize)
	for i := range samples {
		samples[i] = 0.5
	}

	err := processor.ProcessInput("test-channel", samples)
	if err != nil {
		t.Errorf("ProcessInput() error = %v", err)
	}

	// Should send exactly one frame
	if len(transport.sentData) == 0 {
		t.Error("should send data")
	}
	if len(transport.sentData["test-channel"]) == 0 {
		t.Error("should send data to channel")
	}
}

func TestVoiceProcessor_ProcessInputPartialFrame(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	transport := newMockTransport()
	processor := NewVoiceProcessor(codec, transport)

	frameSize := codec.GetFrameSize()
	samples := make([]float64, frameSize/2)

	err := processor.ProcessInput("test-channel", samples)
	if err != nil {
		t.Errorf("ProcessInput() error = %v", err)
	}

	// Should buffer but not send yet
	if len(transport.sentData) > 0 {
		t.Error("should not send partial frame")
	}

	// Send remaining samples
	err = processor.ProcessInput("test-channel", samples)
	if err != nil {
		t.Errorf("ProcessInput() error = %v", err)
	}

	// Now should send complete frame
	if len(transport.sentData) == 0 {
		t.Error("should send complete frame")
	}
}

func TestVoiceProcessor_ProcessOutputDisabled(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	transport := newMockTransport()
	processor := NewVoiceProcessor(codec, transport)

	processor.SetEnabled(false)

	output, err := processor.ProcessOutput()
	if err != nil {
		t.Errorf("ProcessOutput() error = %v", err)
	}
	if len(output) > 0 {
		t.Error("should not process output when disabled")
	}
}

func TestVoiceProcessor_ProcessOutputNoTransport(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	processor := NewVoiceProcessor(codec, nil)

	output, err := processor.ProcessOutput()
	if err != nil {
		t.Errorf("ProcessOutput() error = %v", err)
	}
	if output == nil {
		t.Error("output should not be nil")
	}
}

func TestVoiceProcessor_ProcessOutput(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	transport := newMockTransport()
	processor := NewVoiceProcessor(codec, transport)

	// Create test samples
	samples := make([]float64, 100)
	for i := range samples {
		samples[i] = 0.5
	}

	// Encode samples
	encoded, err := codec.Encode(samples)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Add to receive queue
	transport.addReceivePacket("test-channel", "sender-1", encoded)

	// Process output
	output, err := processor.ProcessOutput()
	if err != nil {
		t.Errorf("ProcessOutput() error = %v", err)
	}

	// Should have decoded samples
	if len(output) == 0 {
		t.Fatal("output is empty")
	}

	key := "test-channel:sender-1"
	if _, ok := output[key]; !ok {
		t.Errorf("expected output key %s", key)
	}
}

func TestVoiceProcessor_ProcessOutputMultipleSenders(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	transport := newMockTransport()
	processor := NewVoiceProcessor(codec, transport)

	samples := make([]float64, 100)
	encoded, _ := codec.Encode(samples)

	// Add packets from multiple senders
	transport.addReceivePacket("channel-1", "sender-1", encoded)
	transport.addReceivePacket("channel-1", "sender-2", encoded)
	transport.addReceivePacket("channel-2", "sender-3", encoded)

	output, err := processor.ProcessOutput()
	if err != nil {
		t.Errorf("ProcessOutput() error = %v", err)
	}

	// Should have three separate outputs
	if len(output) != 3 {
		t.Errorf("output count = %d, want 3", len(output))
	}
}

func TestVoiceProcessor_ClearBuffers(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	transport := newMockTransport()
	processor := NewVoiceProcessor(codec, transport)

	// Add partial frame
	frameSize := codec.GetFrameSize()
	samples := make([]float64, frameSize/2)
	processor.ProcessInput("test-channel", samples)

	// Add output
	encoded, _ := codec.Encode(make([]float64, 100))
	transport.addReceivePacket("test-channel", "sender-1", encoded)
	processor.ProcessOutput()

	// Clear buffers
	processor.ClearBuffers()

	// Output should be empty
	output, _ := processor.ProcessOutput()
	if len(output) > 0 {
		t.Error("output should be empty after clear")
	}

	// Input buffer should be cleared (next input won't trigger send)
	processor.ProcessInput("test-channel", samples)
	if len(transport.sentData) > 0 {
		t.Error("should not send after clear and partial input")
	}
}

func BenchmarkSimpleVoiceCodec_Encode(b *testing.B) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	samples := make([]float64, codec.GetFrameSize())
	for i := range samples {
		samples[i] = 0.5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = codec.Encode(samples)
	}
}

func BenchmarkSimpleVoiceCodec_Decode(b *testing.B) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	samples := make([]float64, codec.GetFrameSize())
	encoded, _ := codec.Encode(samples)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = codec.Decode(encoded)
	}
}

func BenchmarkVoiceProcessor_ProcessInput(b *testing.B) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)
	transport := newMockTransport()
	processor := NewVoiceProcessor(codec, transport)

	samples := make([]float64, codec.GetFrameSize())
	for i := range samples {
		samples[i] = 0.5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.ProcessInput("test-channel", samples)
	}
}

// TestSimpleVoiceCodec_OddLengthEncoding tests the edge case behavior documented in Encode():
// "For odd-length sample arrays, the output rounds up to ceil(n/2) bytes.
// During decoding, this produces an extra sample (decode always produces even count)."
func TestSimpleVoiceCodec_OddLengthEncoding(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)

	tests := []struct {
		name            string
		inputLen        int
		wantEncodedLen  int
		wantDecodedLen  int
		wantExtraSample bool
	}{
		{
			name:            "odd length 1",
			inputLen:        1,
			wantEncodedLen:  1, // ceil(1/2) = 1
			wantDecodedLen:  2, // always even
			wantExtraSample: true,
		},
		{
			name:            "odd length 3",
			inputLen:        3,
			wantEncodedLen:  2, // ceil(3/2) = 2
			wantDecodedLen:  4, // always even
			wantExtraSample: true,
		},
		{
			name:            "odd length 5",
			inputLen:        5,
			wantEncodedLen:  3, // ceil(5/2) = 3
			wantDecodedLen:  6, // always even
			wantExtraSample: true,
		},
		{
			name:            "odd length 101",
			inputLen:        101,
			wantEncodedLen:  51,  // ceil(101/2) = 51
			wantDecodedLen:  102, // always even
			wantExtraSample: true,
		},
		{
			name:            "even length 2",
			inputLen:        2,
			wantEncodedLen:  1, // 2/2 = 1
			wantDecodedLen:  2, // always even
			wantExtraSample: false,
		},
		{
			name:            "even length 100",
			inputLen:        100,
			wantEncodedLen:  50,  // 100/2 = 50
			wantDecodedLen:  100, // always even
			wantExtraSample: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create input samples with distinct values
			samples := make([]float64, tt.inputLen)
			for i := range samples {
				samples[i] = float64(i%10) / 10.0 // Values 0.0, 0.1, ..., 0.9
			}

			// Test encoding
			encoded, err := codec.Encode(samples)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if len(encoded) != tt.wantEncodedLen {
				t.Errorf("Encode() len = %d, want %d (ceil(%d/2) = %d)",
					len(encoded), tt.wantEncodedLen, tt.inputLen, tt.wantEncodedLen)
			}

			// Test decoding
			decoded, err := codec.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(decoded) != tt.wantDecodedLen {
				t.Errorf("Decode() len = %d, want %d (always even, 2 samples per byte)",
					len(decoded), tt.wantDecodedLen)
			}

			// Verify decode always produces even count
			if len(decoded)%2 != 0 {
				t.Errorf("Decode() len = %d, should always be even", len(decoded))
			}

			// For odd-length inputs, verify extra sample exists
			if tt.wantExtraSample {
				if len(decoded) != tt.inputLen+1 {
					t.Errorf("Decode() len = %d, want %d (input + 1 extra sample)",
						len(decoded), tt.inputLen+1)
				}
				// Extra sample should be valid (clamped to [-1.0, 1.0])
				extraSample := decoded[len(decoded)-1]
				if extraSample < -1.0 || extraSample > 1.0 {
					t.Errorf("extra sample = %f, should be in [-1.0, 1.0]", extraSample)
				}
			}

			// Verify original samples are reasonably reconstructed (lossy codec)
			maxError := 0.2 // 20% tolerance for simple codec
			for i := 0; i < tt.inputLen; i++ {
				diff := samples[i] - decoded[i]
				if diff < 0 {
					diff = -diff
				}
				if diff > maxError {
					t.Errorf("sample[%d] error = %f, want < %f (input=%f, decoded=%f)",
						i, diff, maxError, samples[i], decoded[i])
					break
				}
			}
		})
	}
}

// TestSimpleVoiceCodec_OddLengthRoundTrip tests that odd-length encoding/decoding maintains
// reasonable data integrity despite the extra sample.
func TestSimpleVoiceCodec_OddLengthRoundTrip(t *testing.T) {
	codec := NewSimpleVoiceCodec(48000, VoiceQualityMedium)

	// Test various odd-length sample patterns
	patterns := []struct {
		name    string
		samples []float64
	}{
		{
			name:    "single sample silence",
			samples: []float64{0.0},
		},
		{
			name:    "single sample tone",
			samples: []float64{0.5},
		},
		{
			name:    "three samples varying",
			samples: []float64{0.0, 0.5, 1.0},
		},
		{
			name:    "five samples sawtooth",
			samples: []float64{0.0, 0.25, 0.5, 0.75, 1.0},
		},
		{
			name:    "seven samples alternating",
			samples: []float64{-0.5, 0.5, -0.5, 0.5, -0.5, 0.5, -0.5},
		},
	}

	for _, p := range patterns {
		t.Run(p.name, func(t *testing.T) {
			encoded, err := codec.Encode(p.samples)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			decoded, err := codec.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			// Verify decoded length is always even and input + 1
			if len(decoded)%2 != 0 {
				t.Errorf("decoded len = %d, should be even", len(decoded))
			}
			if len(decoded) != len(p.samples)+1 {
				t.Errorf("decoded len = %d, want %d (input + 1)", len(decoded), len(p.samples)+1)
			}

			// Verify original samples are reconstructed within tolerance
			maxError := 0.2
			for i := 0; i < len(p.samples); i++ {
				diff := p.samples[i] - decoded[i]
				if diff < 0 {
					diff = -diff
				}
				if diff > maxError {
					t.Errorf("sample[%d] error = %f, input=%f, decoded=%f",
						i, diff, p.samples[i], decoded[i])
				}
			}
		})
	}
}

func TestVoiceProcessor_ProcessInputEmptyChannelID(t *testing.T) {
	tests := []struct {
		name    string
		samples []float64
		wantErr bool
	}{
		{
			name:    "empty channelID with samples",
			samples: []float64{0.1, 0.2, 0.3},
			wantErr: true,
		},
		{
			name:    "empty channelID with no samples",
			samples: []float64{},
			wantErr: false, // No error because len(samples) == 0 returns early
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codec := NewSimpleVoiceCodec(48000, VoiceQualityHigh)
			processor := NewVoiceProcessor(codec, nil)

			err := processor.ProcessInput("", tt.samples)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessInput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && err.Error() != "channelID cannot be empty" {
				t.Errorf("ProcessInput() error = %v, want 'channelID cannot be empty'", err)
			}
		})
	}
}
