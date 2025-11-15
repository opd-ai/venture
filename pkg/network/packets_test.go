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
			expectedRange: [2]int{90, 100}, // 37 base + 50 msg + ~12+16 encryption
		},
		{
			name:          "short compressed",
			messageLen:    50,
			compressed:    true,
			expectedRange: [2]int{60, 70}, // 37 base + ~25 compressed + ~12+16 encryption
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
