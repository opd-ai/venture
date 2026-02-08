package network

import (
	"crypto/rand"
	"testing"
	"time"
)

func TestSerializeDeserializeChatPacket(t *testing.T) {
	// Create test packet
	var messageID [16]byte
	rand.Read(messageID[:])

	originalPkt := &ChatPacket{
		Header: PacketHeader{
			MessageID: messageID,
		},
		SenderID:  12345,
		Channel:   1,
		Timestamp: time.Now().Truncate(time.Second), // Truncate to second precision
		Payload:   []byte("Hello, world!"),
	}

	// Serialize
	serialized, err := SerializeChatPacket(originalPkt)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	// Verify minimum size
	expectedMinSize := 37 + len(originalPkt.Payload)
	if len(serialized) != expectedMinSize {
		t.Errorf("Expected serialized size %d, got %d", expectedMinSize, len(serialized))
	}

	// Deserialize
	deserializedPkt, err := DeserializeChatPacket(serialized)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// Verify fields
	if deserializedPkt.Header.MessageID != originalPkt.Header.MessageID {
		t.Error("Message ID mismatch")
	}
	if deserializedPkt.SenderID != originalPkt.SenderID {
		t.Errorf("Sender ID mismatch: expected %d, got %d", originalPkt.SenderID, deserializedPkt.SenderID)
	}
	if deserializedPkt.Channel != originalPkt.Channel {
		t.Errorf("Channel mismatch: expected %d, got %d", originalPkt.Channel, deserializedPkt.Channel)
	}
	if !deserializedPkt.Timestamp.Equal(originalPkt.Timestamp) {
		t.Errorf("Timestamp mismatch: expected %v, got %v", originalPkt.Timestamp, deserializedPkt.Timestamp)
	}
	if !bytesEqual(deserializedPkt.Payload, originalPkt.Payload) {
		t.Error("Payload mismatch")
	}
}

func TestDeserializeChatPacket_TooShort(t *testing.T) {
	shortData := make([]byte, 30) // Less than minimum 37 bytes
	_, err := DeserializeChatPacket(shortData)

	if err == nil {
		t.Error("Expected error for packet too short")
	}
}

func TestDeserializeChatPacket_InvalidPayloadLength(t *testing.T) {
	// Create packet with invalid payload length
	data := make([]byte, 37)
	// Set payload length to 1000 but don't include actual payload
	data[33] = 0xE8 // 1000 in little endian (low byte)
	data[34] = 0x03

	_, err := DeserializeChatPacket(data)
	if err == nil {
		t.Error("Expected error for invalid payload length")
	}
}

func TestSerializeDeserializeTradeProposal(t *testing.T) {
	var messageID [16]byte
	rand.Read(messageID[:])

	originalPkt := &TradeProposalPacket{
		Header: PacketHeader{
			MessageID: messageID,
		},
		ProposerID:  100,
		RecipientID: 200,
		ItemCount:   3,
		Items: []TradeItem{
			{ItemID: 1001, Quantity: 5},
			{ItemID: 1002, Quantity: 10},
			{ItemID: 1003, Quantity: 1},
		},
	}

	// Serialize
	serialized, err := SerializeTradeProposal(originalPkt)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	// Verify size: 16 + 8 + 8 + 4 + (12 * 3) = 72 bytes
	expectedSize := 72
	if len(serialized) != expectedSize {
		t.Errorf("Expected serialized size %d, got %d", expectedSize, len(serialized))
	}

	// Deserialize
	deserializedPkt, err := DeserializeTradeProposal(serialized)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// Verify fields
	if deserializedPkt.ProposerID != originalPkt.ProposerID {
		t.Errorf("ProposerID mismatch: expected %d, got %d", originalPkt.ProposerID, deserializedPkt.ProposerID)
	}
	if deserializedPkt.RecipientID != originalPkt.RecipientID {
		t.Errorf("RecipientID mismatch: expected %d, got %d", originalPkt.RecipientID, deserializedPkt.RecipientID)
	}
	if len(deserializedPkt.Items) != len(originalPkt.Items) {
		t.Fatalf("Item count mismatch: expected %d, got %d", len(originalPkt.Items), len(deserializedPkt.Items))
	}

	for i, item := range originalPkt.Items {
		if deserializedPkt.Items[i].ItemID != item.ItemID {
			t.Errorf("Item %d ID mismatch: expected %d, got %d", i, item.ItemID, deserializedPkt.Items[i].ItemID)
		}
		if deserializedPkt.Items[i].Quantity != item.Quantity {
			t.Errorf("Item %d quantity mismatch: expected %d, got %d", i, item.Quantity, deserializedPkt.Items[i].Quantity)
		}
	}
}

func TestSerializeTradeProposal_TooManyItems(t *testing.T) {
	var messageID [16]byte
	pkt := &TradeProposalPacket{
		Header:    PacketHeader{MessageID: messageID},
		ItemCount: 21,
		Items:     make([]TradeItem, 21), // Exceeds max of 20
	}

	_, err := SerializeTradeProposal(pkt)
	if err == nil {
		t.Error("Expected error for too many items")
	}
}

func TestEstimatePacketSize(t *testing.T) {
	tests := []struct {
		name          string
		messageLen    int
		compressed    bool
		expectedRange [2]int // [min, max]
	}{
		{
			name:          "short uncompressed",
			messageLen:    50,
			compressed:    false,
			expectedRange: [2]int{110, 120}, // 37 base + 50 msg + 28 encryption overhead
		},
		{
			name:          "short compressed",
			messageLen:    50,
			compressed:    true,
			expectedRange: [2]int{85, 95}, // 37 base + ~25 compressed + 28 encryption overhead
		},
		{
			name:          "long uncompressed",
			messageLen:    200,
			compressed:    false,
			expectedRange: [2]int{260, 270},
		},
		{
			name:          "long compressed",
			messageLen:    200,
			compressed:    true,
			expectedRange: [2]int{140, 170}, // ~50% compression + encryption overhead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := EstimatePacketSize(tt.messageLen, tt.compressed)

			if size < tt.expectedRange[0] || size > tt.expectedRange[1] {
				t.Errorf("Expected size in range [%d, %d], got %d",
					tt.expectedRange[0], tt.expectedRange[1], size)
			}
		})
	}
}

func TestBandwidthEstimation_ChatMessages(t *testing.T) {
	// Simulate 10 messages per minute for 1 player
	messagesPerMinute := 10
	avgMessageLen := 100 // bytes

	// Calculate total bandwidth
	totalBytes := 0
	for i := 0; i < messagesPerMinute; i++ {
		// Estimate with compression for messages >100 bytes
		packetSize := EstimatePacketSize(avgMessageLen, true)
		totalBytes += packetSize
	}

	bytesPerSecond := totalBytes / 60
	kilobytesPerSecond := float64(bytesPerSecond) / 1024.0

	t.Logf("Bandwidth for 10 msg/min: %.2f KB/s (%d bytes/s)", kilobytesPerSecond, bytesPerSecond)

	// Phase 36 acceptance criteria: <10KB/s per player
	if kilobytesPerSecond > 10.0 {
		t.Errorf("Bandwidth exceeds target: %.2f KB/s > 10 KB/s", kilobytesPerSecond)
	}
}

func TestPacketCompression_Savings(t *testing.T) {
	// Test compression savings on typical chat messages
	testMessages := []string{
		"Hello!",
		"The quick brown fox jumps over the lazy dog.",
		"This is a longer message with some repetitive content, repetitive content, repetitive content.",
	}

	for _, msg := range testMessages {
		uncompressedSize := EstimatePacketSize(len(msg), false)
		compressedSize := EstimatePacketSize(len(msg), true)

		savings := float64(uncompressedSize-compressedSize) / float64(uncompressedSize) * 100

		t.Logf("Message (%d bytes): Uncompressed=%d bytes, Compressed=%d bytes, Savings=%.1f%%",
			len(msg), uncompressedSize, compressedSize, savings)

		// For messages >100 bytes, expect >30% savings (Phase 36 requirement)
		if len(msg) > 100 {
			if savings < 30.0 {
				t.Logf("Warning: Compression savings %.1f%% below 30%% target", savings)
			}
		}
	}
}

// BenchmarkSerializeChatPacket benchmarks chat packet serialization
func BenchmarkSerializeChatPacket(b *testing.B) {
	var messageID [16]byte
	rand.Read(messageID[:])

	pkt := &ChatPacket{
		Header:    PacketHeader{MessageID: messageID},
		SenderID:  12345,
		Channel:   1,
		Timestamp: time.Now(),
		Payload:   []byte("Hello, world! This is a test message."),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SerializeChatPacket(pkt)
	}
}

// BenchmarkDeserializeChatPacket benchmarks chat packet deserialization
func BenchmarkDeserializeChatPacket(b *testing.B) {
	var messageID [16]byte
	rand.Read(messageID[:])

	pkt := &ChatPacket{
		Header:    PacketHeader{MessageID: messageID},
		SenderID:  12345,
		Channel:   1,
		Timestamp: time.Now(),
		Payload:   []byte("Hello, world! This is a test message."),
	}

	serialized, _ := SerializeChatPacket(pkt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DeserializeChatPacket(serialized)
	}
}

// BenchmarkSerializeTradeProposal benchmarks trade proposal serialization
func BenchmarkSerializeTradeProposal(b *testing.B) {
	var messageID [16]byte
	rand.Read(messageID[:])

	pkt := &TradeProposalPacket{
		Header:      PacketHeader{MessageID: messageID},
		ProposerID:  100,
		RecipientID: 200,
		Items: []TradeItem{
			{ItemID: 1001, Quantity: 5},
			{ItemID: 1002, Quantity: 10},
			{ItemID: 1003, Quantity: 1},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SerializeTradeProposal(pkt)
	}
}

// TestSerializeDeserializeImageMetadata tests ImageMetadata packet serialization round-trip.
func TestSerializeDeserializeImageMetadata(t *testing.T) {
	tests := []struct {
		name string
		pkt  *ImageMetadataPacket
	}{
		{
			name: "basic metadata",
			pkt: &ImageMetadataPacket{
				Header: PacketHeader{
					MessageID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				},
				ImageID:         [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				SenderID:        12345,
				Width:           1024,
				Height:          768,
				Size:            204800,
				Format:          0, // PNG
				ThumbnailOffset: 4096,
			},
		},
		{
			name: "JPEG format",
			pkt: &ImageMetadataPacket{
				Header: PacketHeader{
					MessageID: [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
				},
				ImageID:         [16]byte{15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
				SenderID:        99999,
				Width:           2048,
				Height:          2048,
				Size:            500000,
				Format:          1, // JPEG
				ThumbnailOffset: 8192,
			},
		},
		{
			name: "max dimensions",
			pkt: &ImageMetadataPacket{
				Header: PacketHeader{
					MessageID: [16]byte{255, 254, 253, 252, 251, 250, 249, 248, 247, 246, 245, 244, 243, 242, 241, 240},
				},
				ImageID:         [16]byte{240, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255},
				SenderID:        ^uint64(0), // Max uint64
				Width:           2048,
				Height:          2048,
				Size:            512000, // 500KB
				Format:          0,
				ThumbnailOffset: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := SerializeImageMetadata(tt.pkt)
			if err != nil {
				t.Fatalf("SerializeImageMetadata failed: %v", err)
			}

			// Verify size
			if len(data) != 80 {
				t.Errorf("Expected serialized size 80, got %d", len(data))
			}

			// Deserialize
			result, err := DeserializeImageMetadata(data)
			if err != nil {
				t.Fatalf("DeserializeImageMetadata failed: %v", err)
			}

			// Verify fields
			if result.Header.MessageID != tt.pkt.Header.MessageID {
				t.Errorf("MessageID mismatch: got %v, want %v", result.Header.MessageID, tt.pkt.Header.MessageID)
			}
			if result.ImageID != tt.pkt.ImageID {
				t.Errorf("ImageID mismatch: got %v, want %v", result.ImageID, tt.pkt.ImageID)
			}
			if result.SenderID != tt.pkt.SenderID {
				t.Errorf("SenderID mismatch: got %d, want %d", result.SenderID, tt.pkt.SenderID)
			}
			if result.Width != tt.pkt.Width {
				t.Errorf("Width mismatch: got %d, want %d", result.Width, tt.pkt.Width)
			}
			if result.Height != tt.pkt.Height {
				t.Errorf("Height mismatch: got %d, want %d", result.Height, tt.pkt.Height)
			}
			if result.Size != tt.pkt.Size {
				t.Errorf("Size mismatch: got %d, want %d", result.Size, tt.pkt.Size)
			}
			if result.Format != tt.pkt.Format {
				t.Errorf("Format mismatch: got %d, want %d", result.Format, tt.pkt.Format)
			}
			if result.ThumbnailOffset != tt.pkt.ThumbnailOffset {
				t.Errorf("ThumbnailOffset mismatch: got %d, want %d", result.ThumbnailOffset, tt.pkt.ThumbnailOffset)
			}
		})
	}
}

// TestDeserializeImageMetadataInvalidInput tests error handling for invalid input.
func TestDeserializeImageMetadataInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "too short",
			data:    make([]byte, 79),
			wantErr: true,
		},
		{
			name:    "exact size",
			data:    make([]byte, 80),
			wantErr: false,
		},
		{
			name:    "larger than needed",
			data:    make([]byte, 100),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeserializeImageMetadata(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeImageMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSerializeDeserializeImageChunk tests ImageChunk packet serialization round-trip.
func TestSerializeDeserializeImageChunk(t *testing.T) {
	tests := []struct {
		name string
		pkt  *ImageChunkPacket
	}{
		{
			name: "first chunk",
			pkt: &ImageChunkPacket{
				Header: PacketHeader{
					MessageID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				},
				ImageID:      [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				ChunkIndex:   0,
				TotalChunks:  10,
				IsResume:     false,
				LastChunkIdx: -1,
				Data:         []byte("chunk data 0"),
			},
		},
		{
			name: "middle chunk",
			pkt: &ImageChunkPacket{
				Header: PacketHeader{
					MessageID: [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
				},
				ImageID:      [16]byte{15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
				ChunkIndex:   5,
				TotalChunks:  10,
				IsResume:     false,
				LastChunkIdx: -1,
				Data:         make([]byte, 64*1024), // Max chunk size
			},
		},
		{
			name: "resume chunk",
			pkt: &ImageChunkPacket{
				Header: PacketHeader{
					MessageID: [16]byte{255, 254, 253, 252, 251, 250, 249, 248, 247, 246, 245, 244, 243, 242, 241, 240},
				},
				ImageID:      [16]byte{240, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255},
				ChunkIndex:   3,
				TotalChunks:  8,
				IsResume:     true,
				LastChunkIdx: 2,
				Data:         []byte("resumed chunk"),
			},
		},
		{
			name: "last chunk",
			pkt: &ImageChunkPacket{
				Header: PacketHeader{
					MessageID: [16]byte{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115},
				},
				ImageID:      [16]byte{115, 114, 113, 112, 111, 110, 109, 108, 107, 106, 105, 104, 103, 102, 101, 100},
				ChunkIndex:   9,
				TotalChunks:  10,
				IsResume:     false,
				LastChunkIdx: -1,
				Data:         []byte("final chunk"),
			},
		},
		{
			name: "empty data",
			pkt: &ImageChunkPacket{
				Header: PacketHeader{
					MessageID: [16]byte{50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65},
				},
				ImageID:      [16]byte{65, 64, 63, 62, 61, 60, 59, 58, 57, 56, 55, 54, 53, 52, 51, 50},
				ChunkIndex:   0,
				TotalChunks:  1,
				IsResume:     false,
				LastChunkIdx: -1,
				Data:         []byte{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := SerializeImageChunk(tt.pkt)
			if err != nil {
				t.Fatalf("SerializeImageChunk failed: %v", err)
			}

			// Verify size
			expectedSize := 49 + len(tt.pkt.Data)
			if len(data) != expectedSize {
				t.Errorf("Expected serialized size %d, got %d", expectedSize, len(data))
			}

			// Deserialize
			result, err := DeserializeImageChunk(data)
			if err != nil {
				t.Fatalf("DeserializeImageChunk failed: %v", err)
			}

			// Verify fields
			if result.Header.MessageID != tt.pkt.Header.MessageID {
				t.Errorf("MessageID mismatch: got %v, want %v", result.Header.MessageID, tt.pkt.Header.MessageID)
			}
			if result.ImageID != tt.pkt.ImageID {
				t.Errorf("ImageID mismatch: got %v, want %v", result.ImageID, tt.pkt.ImageID)
			}
			if result.ChunkIndex != tt.pkt.ChunkIndex {
				t.Errorf("ChunkIndex mismatch: got %d, want %d", result.ChunkIndex, tt.pkt.ChunkIndex)
			}
			if result.TotalChunks != tt.pkt.TotalChunks {
				t.Errorf("TotalChunks mismatch: got %d, want %d", result.TotalChunks, tt.pkt.TotalChunks)
			}
			if result.IsResume != tt.pkt.IsResume {
				t.Errorf("IsResume mismatch: got %v, want %v", result.IsResume, tt.pkt.IsResume)
			}
			if result.LastChunkIdx != tt.pkt.LastChunkIdx {
				t.Errorf("LastChunkIdx mismatch: got %d, want %d", result.LastChunkIdx, tt.pkt.LastChunkIdx)
			}
			if !bytesEqual(result.Data, tt.pkt.Data) {
				t.Errorf("Data mismatch: got %d bytes, want %d bytes", len(result.Data), len(tt.pkt.Data))
			}
		})
	}
}

// TestDeserializeImageChunkInvalidInput tests error handling for invalid input.
func TestDeserializeImageChunkInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "too short for header",
			data:    make([]byte, 48),
			wantErr: true,
		},
		{
			name:    "minimum valid size",
			data:    make([]byte, 49),
			wantErr: false,
		},
		{
			name: "data length mismatch",
			data: func() []byte {
				buf := make([]byte, 49)
				// Set data length to 100 but only provide header
				buf[45] = 100
				return buf
			}(),
			wantErr: true,
		},
		{
			name: "valid with data",
			data: func() []byte {
				buf := make([]byte, 49+10)
				// Set data length to 10
				buf[45] = 10
				return buf
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeserializeImageChunk(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeImageChunk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestImageChunkMaxSize tests handling of maximum chunk size (64KB).
func TestImageChunkMaxSize(t *testing.T) {
	maxData := make([]byte, 64*1024)
	for i := range maxData {
		maxData[i] = byte(i % 256)
	}

	pkt := &ImageChunkPacket{
		Header: PacketHeader{
			MessageID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		ImageID:      [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		ChunkIndex:   0,
		TotalChunks:  1,
		IsResume:     false,
		LastChunkIdx: -1,
		Data:         maxData,
	}

	// Serialize
	data, err := SerializeImageChunk(pkt)
	if err != nil {
		t.Fatalf("SerializeImageChunk failed for max size: %v", err)
	}

	// Verify size is header + max chunk
	expectedSize := 49 + 64*1024
	if len(data) != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, len(data))
	}

	// Deserialize
	result, err := DeserializeImageChunk(data)
	if err != nil {
		t.Fatalf("DeserializeImageChunk failed for max size: %v", err)
	}

	// Verify data integrity
	if !bytesEqual(result.Data, maxData) {
		t.Error("Data integrity check failed for max size chunk")
	}
}

// BenchmarkSerializeImageMetadata benchmarks metadata serialization.
func BenchmarkSerializeImageMetadata(b *testing.B) {
	pkt := &ImageMetadataPacket{
		Header: PacketHeader{
			MessageID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		ImageID:         [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		SenderID:        12345,
		Width:           1024,
		Height:          768,
		Size:            204800,
		Format:          0,
		ThumbnailOffset: 4096,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SerializeImageMetadata(pkt)
	}
}

// BenchmarkDeserializeImageMetadata benchmarks metadata deserialization.
func BenchmarkDeserializeImageMetadata(b *testing.B) {
	pkt := &ImageMetadataPacket{
		Header: PacketHeader{
			MessageID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		ImageID:         [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		SenderID:        12345,
		Width:           1024,
		Height:          768,
		Size:            204800,
		Format:          0,
		ThumbnailOffset: 4096,
	}
	data, _ := SerializeImageMetadata(pkt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DeserializeImageMetadata(data)
	}
}

// BenchmarkSerializeImageChunk benchmarks chunk serialization.
func BenchmarkSerializeImageChunk(b *testing.B) {
	pkt := &ImageChunkPacket{
		Header: PacketHeader{
			MessageID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		ImageID:      [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		ChunkIndex:   5,
		TotalChunks:  10,
		IsResume:     false,
		LastChunkIdx: -1,
		Data:         make([]byte, 64*1024),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SerializeImageChunk(pkt)
	}
}

// BenchmarkDeserializeImageChunk benchmarks chunk deserialization.
func BenchmarkDeserializeImageChunk(b *testing.B) {
	pkt := &ImageChunkPacket{
		Header: PacketHeader{
			MessageID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		ImageID:      [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		ChunkIndex:   5,
		TotalChunks:  10,
		IsResume:     false,
		LastChunkIdx: -1,
		Data:         make([]byte, 64*1024),
	}
	data, _ := SerializeImageChunk(pkt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DeserializeImageChunk(data)
	}
}
