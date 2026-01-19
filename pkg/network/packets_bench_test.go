// Package network provides benchmark tests for packet processing.
// Target: 1000 packets/s processing capability.
package network

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// BenchmarkChatPacketSerialize benchmarks chat packet serialization.
func BenchmarkChatPacketSerialize(b *testing.B) {
	msgID, _ := uuid.NewRandom()
	pkt := &ChatPacket{
		Header:    PacketHeader{MessageID: msgID},
		SenderID:  12345,
		Channel:   1,
		Timestamp: time.Now(),
		Payload:   []byte("Hello, this is a test message for benchmarking!"),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := SerializeChatPacket(pkt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkChatPacketDeserialize benchmarks chat packet deserialization.
func BenchmarkChatPacketDeserialize(b *testing.B) {
	msgID, _ := uuid.NewRandom()
	pkt := &ChatPacket{
		Header:    PacketHeader{MessageID: msgID},
		SenderID:  12345,
		Channel:   1,
		Timestamp: time.Now(),
		Payload:   []byte("Hello, this is a test message for benchmarking!"),
	}

	data, err := SerializeChatPacket(pkt)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := DeserializeChatPacket(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTradeProposalSerialize benchmarks trade proposal serialization.
func BenchmarkTradeProposalSerialize(b *testing.B) {
	msgID, _ := uuid.NewRandom()
	pkt := &TradeProposalPacket{
		Header:      PacketHeader{MessageID: msgID},
		ProposerID:  12345,
		RecipientID: 67890,
		ItemCount:   5,
		Items: []TradeItem{
			{ItemID: 1, Quantity: 10},
			{ItemID: 2, Quantity: 5},
			{ItemID: 3, Quantity: 1},
			{ItemID: 4, Quantity: 20},
			{ItemID: 5, Quantity: 3},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := SerializeTradeProposal(pkt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTradeProposalDeserialize benchmarks trade proposal deserialization.
func BenchmarkTradeProposalDeserialize(b *testing.B) {
	msgID, _ := uuid.NewRandom()
	pkt := &TradeProposalPacket{
		Header:      PacketHeader{MessageID: msgID},
		ProposerID:  12345,
		RecipientID: 67890,
		ItemCount:   5,
		Items: []TradeItem{
			{ItemID: 1, Quantity: 10},
			{ItemID: 2, Quantity: 5},
			{ItemID: 3, Quantity: 1},
			{ItemID: 4, Quantity: 20},
			{ItemID: 5, Quantity: 3},
		},
	}

	data, err := SerializeTradeProposal(pkt)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := DeserializeTradeProposal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTradeProposalSerializeMaxItems benchmarks max item trade serialization.
func BenchmarkTradeProposalSerializeMaxItems(b *testing.B) {
	msgID, _ := uuid.NewRandom()
	items := make([]TradeItem, 20) // Maximum allowed items
	for i := range items {
		items[i] = TradeItem{ItemID: uint64(i + 1), Quantity: uint32(i + 1)}
	}

	pkt := &TradeProposalPacket{
		Header:      PacketHeader{MessageID: msgID},
		ProposerID:  12345,
		RecipientID: 67890,
		ItemCount:   20,
		Items:       items,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := SerializeTradeProposal(pkt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkChatPacketSerializeLargePayload benchmarks large message serialization.
func BenchmarkChatPacketSerializeLargePayload(b *testing.B) {
	msgID, _ := uuid.NewRandom()
	// 1KB payload
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	pkt := &ChatPacket{
		Header:    PacketHeader{MessageID: msgID},
		SenderID:  12345,
		Channel:   1,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := SerializeChatPacket(pkt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEstimatePacketSize benchmarks packet size estimation.
func BenchmarkEstimatePacketSize(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = EstimatePacketSize(100, true)
	}
}

// BenchmarkChatPacketRoundTrip benchmarks full serialize/deserialize cycle.
func BenchmarkChatPacketRoundTrip(b *testing.B) {
	msgID, _ := uuid.NewRandom()
	pkt := &ChatPacket{
		Header:    PacketHeader{MessageID: msgID},
		SenderID:  12345,
		Channel:   1,
		Timestamp: time.Now(),
		Payload:   []byte("Hello, this is a test message for benchmarking!"),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data, err := SerializeChatPacket(pkt)
		if err != nil {
			b.Fatal(err)
		}
		_, err = DeserializeChatPacket(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTradePacketRoundTrip benchmarks full trade serialize/deserialize cycle.
func BenchmarkTradePacketRoundTrip(b *testing.B) {
	msgID, _ := uuid.NewRandom()
	pkt := &TradeProposalPacket{
		Header:      PacketHeader{MessageID: msgID},
		ProposerID:  12345,
		RecipientID: 67890,
		ItemCount:   5,
		Items: []TradeItem{
			{ItemID: 1, Quantity: 10},
			{ItemID: 2, Quantity: 5},
			{ItemID: 3, Quantity: 1},
			{ItemID: 4, Quantity: 20},
			{ItemID: 5, Quantity: 3},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data, err := SerializeTradeProposal(pkt)
		if err != nil {
			b.Fatal(err)
		}
		_, err = DeserializeTradeProposal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPacketProcessingParallel benchmarks parallel packet processing.
func BenchmarkPacketProcessingParallel(b *testing.B) {
	msgID, _ := uuid.NewRandom()
	pkt := &ChatPacket{
		Header:    PacketHeader{MessageID: msgID},
		SenderID:  12345,
		Channel:   1,
		Timestamp: time.Now(),
		Payload:   []byte("Hello, this is a test message for benchmarking!"),
	}

	data, err := SerializeChatPacket(pkt)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := DeserializeChatPacket(data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
