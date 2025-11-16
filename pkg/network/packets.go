package network

import (
	"encoding/binary"
	"fmt"
	"time"
)

// PacketType represents the type of network packet.
type PacketType byte

const (
	// PacketTypeChatMessage represents a chat message packet
	PacketTypeChatMessage PacketType = 50
	// PacketTypeImageMetadata represents an image metadata packet
	PacketTypeImageMetadata PacketType = 51
	// PacketTypeImageChunk represents an image data chunk packet
	PacketTypeImageChunk PacketType = 52
	// PacketTypeTradeProposal represents a trade proposal packet
	PacketTypeTradeProposal PacketType = 53
	// PacketTypeTradeAccept represents a trade acceptance packet
	PacketTypeTradeAccept PacketType = 54
	// PacketTypeMessageACK represents a message acknowledgment
	PacketTypeMessageACK PacketType = 55
)

// PacketHeader represents the common header for all packets (16 bytes).
// Layout:
// - Message ID (16 bytes): UUID in binary form
type PacketHeader struct {
	MessageID [16]byte // UUID v4
}

// ChatPacket represents the complete chat message packet structure.
// Total size: 16 (header) + 8 (sender) + 1 (channel) + 8 (timestamp) + variable (payload)
type ChatPacket struct {
	Header    PacketHeader
	SenderID  uint64    // 8 bytes
	Channel   byte      // 1 byte
	Timestamp time.Time // 8 bytes (Unix timestamp)
	Payload   []byte    // Variable (encrypted + optional compression flag)
}

// ImageMetadataPacket represents image metadata (80 bytes total).
// Layout:
// - Header (16 bytes)
// - ImageID (16 bytes): UUID
// - SenderID (8 bytes)
// - Width (4 bytes)
// - Height (4 bytes)
// - Size (4 bytes)
// - Format (4 bytes)
// - ThumbnailOffset (8 bytes)
// - Reserved (16 bytes)
type ImageMetadataPacket struct {
	Header          PacketHeader
	ImageID         [16]byte
	SenderID        uint64
	Width           uint32
	Height          uint32
	Size            uint32
	Format          uint32 // 0=PNG, 1=JPEG
	ThumbnailOffset uint64
	Reserved        [16]byte
}

// TradeProposalPacket represents a trade proposal.
// Header (16 bytes) + ProposerID (8 bytes) + RecipientID (8 bytes) + ItemCount (4 bytes) + Items (variable)
type TradeProposalPacket struct {
	Header      PacketHeader
	ProposerID  uint64
	RecipientID uint64
	ItemCount   uint32
	Items       []TradeItem // Max 20 items
}

// TradeItem represents a single item in a trade (12 bytes).
type TradeItem struct {
	ItemID   uint64 // 8 bytes
	Quantity uint32 // 4 bytes
}

// SerializeChatPacket serializes a chat packet to bytes.
// Format: Header (16) + SenderID (8) + Channel (1) + Timestamp (8) + PayloadLen (4) + Payload (variable)
func SerializeChatPacket(pkt *ChatPacket) ([]byte, error) {
	// Calculate total size
	totalSize := 16 + 8 + 1 + 8 + 4 + len(pkt.Payload)
	buf := make([]byte, totalSize)

	// Write header
	copy(buf[0:16], pkt.Header.MessageID[:])

	// Write sender ID
	binary.LittleEndian.PutUint64(buf[16:24], pkt.SenderID)

	// Write channel
	buf[24] = pkt.Channel

	// Write timestamp (Unix seconds)
	binary.LittleEndian.PutUint64(buf[25:33], uint64(pkt.Timestamp.Unix()))

	// Write payload length
	binary.LittleEndian.PutUint32(buf[33:37], uint32(len(pkt.Payload)))

	// Write payload
	copy(buf[37:], pkt.Payload)

	return buf, nil
}

// DeserializeChatPacket deserializes a chat packet from bytes.
func DeserializeChatPacket(data []byte) (*ChatPacket, error) {
	if len(data) < 37 {
		return nil, fmt.Errorf("packet too short: %d bytes", len(data))
	}

	pkt := &ChatPacket{}

	// Read header
	copy(pkt.Header.MessageID[:], data[0:16])

	// Read sender ID
	pkt.SenderID = binary.LittleEndian.Uint64(data[16:24])

	// Read channel
	pkt.Channel = data[24]

	// Read timestamp
	timestamp := int64(binary.LittleEndian.Uint64(data[25:33]))
	pkt.Timestamp = time.Unix(timestamp, 0)

	// Read payload length
	payloadLen := binary.LittleEndian.Uint32(data[33:37])

	// Validate payload length
	if len(data) < int(37+payloadLen) {
		return nil, fmt.Errorf("invalid payload length: expected %d, got %d", payloadLen, len(data)-37)
	}

	// Read payload
	pkt.Payload = make([]byte, payloadLen)
	copy(pkt.Payload, data[37:37+payloadLen])

	return pkt, nil
}

// SerializeTradeProposal serializes a trade proposal packet to bytes.
func SerializeTradeProposal(pkt *TradeProposalPacket) ([]byte, error) {
	if len(pkt.Items) > 20 {
		return nil, fmt.Errorf("too many items: %d (max 20)", len(pkt.Items))
	}

	// Calculate size: Header (16) + ProposerID (8) + RecipientID (8) + ItemCount (4) + Items (12 * count)
	totalSize := 16 + 8 + 8 + 4 + (12 * len(pkt.Items))
	buf := make([]byte, totalSize)

	// Write header
	copy(buf[0:16], pkt.Header.MessageID[:])

	// Write proposer ID
	binary.LittleEndian.PutUint64(buf[16:24], pkt.ProposerID)

	// Write recipient ID
	binary.LittleEndian.PutUint64(buf[24:32], pkt.RecipientID)

	// Write item count
	binary.LittleEndian.PutUint32(buf[32:36], uint32(len(pkt.Items)))

	// Write items
	offset := 36
	for _, item := range pkt.Items {
		binary.LittleEndian.PutUint64(buf[offset:offset+8], item.ItemID)
		binary.LittleEndian.PutUint32(buf[offset+8:offset+12], item.Quantity)
		offset += 12
	}

	return buf, nil
}

// DeserializeTradeProposal deserializes a trade proposal packet from bytes.
func DeserializeTradeProposal(data []byte) (*TradeProposalPacket, error) {
	if len(data) < 36 {
		return nil, fmt.Errorf("packet too short: %d bytes", len(data))
	}

	pkt := &TradeProposalPacket{}

	// Read header
	copy(pkt.Header.MessageID[:], data[0:16])

	// Read proposer ID
	pkt.ProposerID = binary.LittleEndian.Uint64(data[16:24])

	// Read recipient ID
	pkt.RecipientID = binary.LittleEndian.Uint64(data[24:32])

	// Read item count
	itemCount := binary.LittleEndian.Uint32(data[32:36])
	if itemCount > 20 {
		return nil, fmt.Errorf("invalid item count: %d (max 20)", itemCount)
	}

	// Validate total size
	expectedSize := 36 + (12 * int(itemCount))
	if len(data) < expectedSize {
		return nil, fmt.Errorf("invalid packet size: expected %d, got %d", expectedSize, len(data))
	}

	// Read items
	pkt.Items = make([]TradeItem, itemCount)
	offset := 36
	for i := uint32(0); i < itemCount; i++ {
		pkt.Items[i].ItemID = binary.LittleEndian.Uint64(data[offset : offset+8])
		pkt.Items[i].Quantity = binary.LittleEndian.Uint32(data[offset+8 : offset+12])
		offset += 12
	}

	return pkt, nil
}

// EstimatePacketSize estimates the total packet size for a chat message.
func EstimatePacketSize(messageLen int, compressed bool) int {
	// Header (16) + SenderID (8) + Channel (1) + Timestamp (8) + PayloadLen (4) + Payload
	baseSize := 37

	payloadSize := messageLen
	if compressed {
		// Estimate 50% compression ratio for typical text
		payloadSize = messageLen / 2
	}

	// Add encryption overhead (IV + tag for AES-GCM)
	payloadSize += 12 + 16

	return baseSize + payloadSize
}
