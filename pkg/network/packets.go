package network

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
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
	// PacketTypeVoice represents a voice data packet
	PacketTypeVoice PacketType = 56
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

// ImageChunkPacket represents an image data chunk for transfer.
// Layout:
// - Header (16 bytes)
// - ImageID (16 bytes): UUID
// - ChunkIndex (4 bytes)
// - TotalChunks (4 bytes)
// - IsResume (1 byte)
// - LastChunkIdx (4 bytes)
// - DataLen (4 bytes)
// - Data (variable)
type ImageChunkPacket struct {
	Header       PacketHeader
	ImageID      [16]byte
	ChunkIndex   uint32
	TotalChunks  uint32
	IsResume     bool
	LastChunkIdx int32
	Data         []byte
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
		logrus.WithFields(logrus.Fields{
			"packet_type": "chat",
			"data_length": len(data),
			"min_size":    37,
		}).Warn("packet too short")
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
		logrus.WithFields(logrus.Fields{
			"packet_type":      "chat",
			"expected_payload": payloadLen,
			"actual_payload":   len(data) - 37,
		}).Warn("invalid payload length")
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
		logrus.WithFields(logrus.Fields{
			"packet_type": "trade_proposal",
			"item_count":  len(pkt.Items),
			"max_items":   20,
		}).Warn("too many items in trade proposal")
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

// SerializeImageMetadata serializes an image metadata packet to bytes.
func SerializeImageMetadata(pkt *ImageMetadataPacket) ([]byte, error) {
	// Fixed size: 80 bytes
	buf := make([]byte, 80)

	// Write header
	copy(buf[0:16], pkt.Header.MessageID[:])

	// Write image ID
	copy(buf[16:32], pkt.ImageID[:])

	// Write sender ID
	binary.LittleEndian.PutUint64(buf[32:40], pkt.SenderID)

	// Write dimensions and size
	binary.LittleEndian.PutUint32(buf[40:44], pkt.Width)
	binary.LittleEndian.PutUint32(buf[44:48], pkt.Height)
	binary.LittleEndian.PutUint32(buf[48:52], pkt.Size)

	// Write format
	binary.LittleEndian.PutUint32(buf[52:56], pkt.Format)

	// Write thumbnail offset
	binary.LittleEndian.PutUint64(buf[56:64], pkt.ThumbnailOffset)

	// Write reserved bytes (already zeroed)
	copy(buf[64:80], pkt.Reserved[:])

	return buf, nil
}

// DeserializeImageMetadata deserializes an image metadata packet from bytes.
func DeserializeImageMetadata(data []byte) (*ImageMetadataPacket, error) {
	if len(data) < 80 {
		return nil, fmt.Errorf("packet too short: %d bytes (expected 80)", len(data))
	}

	pkt := &ImageMetadataPacket{}

	// Read header
	copy(pkt.Header.MessageID[:], data[0:16])

	// Read image ID
	copy(pkt.ImageID[:], data[16:32])

	// Read sender ID
	pkt.SenderID = binary.LittleEndian.Uint64(data[32:40])

	// Read dimensions and size
	pkt.Width = binary.LittleEndian.Uint32(data[40:44])
	pkt.Height = binary.LittleEndian.Uint32(data[44:48])
	pkt.Size = binary.LittleEndian.Uint32(data[48:52])

	// Read format
	pkt.Format = binary.LittleEndian.Uint32(data[52:56])

	// Read thumbnail offset
	pkt.ThumbnailOffset = binary.LittleEndian.Uint64(data[56:64])

	// Read reserved bytes
	copy(pkt.Reserved[:], data[64:80])

	return pkt, nil
}

// SerializeImageChunk serializes an image chunk packet to bytes.
// Format: Header (16) + ImageID (16) + ChunkIndex (4) + TotalChunks (4) + IsResume (1) + LastChunkIdx (4) + DataLen (4) + Data (variable)
func SerializeImageChunk(pkt *ImageChunkPacket) ([]byte, error) {
	// Calculate total size
	totalSize := 16 + 16 + 4 + 4 + 1 + 4 + 4 + len(pkt.Data)
	buf := make([]byte, totalSize)

	// Write header
	copy(buf[0:16], pkt.Header.MessageID[:])

	// Write image ID
	copy(buf[16:32], pkt.ImageID[:])

	// Write chunk index
	binary.LittleEndian.PutUint32(buf[32:36], pkt.ChunkIndex)

	// Write total chunks
	binary.LittleEndian.PutUint32(buf[36:40], pkt.TotalChunks)

	// Write IsResume flag
	if pkt.IsResume {
		buf[40] = 1
	} else {
		buf[40] = 0
	}

	// Write last chunk index
	binary.LittleEndian.PutUint32(buf[41:45], uint32(pkt.LastChunkIdx))

	// Write data length
	binary.LittleEndian.PutUint32(buf[45:49], uint32(len(pkt.Data)))

	// Write data
	copy(buf[49:], pkt.Data)

	return buf, nil
}

// DeserializeImageChunk deserializes an image chunk packet from bytes.
func DeserializeImageChunk(data []byte) (*ImageChunkPacket, error) {
	if len(data) < 49 {
		return nil, fmt.Errorf("packet too short: %d bytes (expected at least 49)", len(data))
	}

	pkt := &ImageChunkPacket{}

	// Read header
	copy(pkt.Header.MessageID[:], data[0:16])

	// Read image ID
	copy(pkt.ImageID[:], data[16:32])

	// Read chunk index
	pkt.ChunkIndex = binary.LittleEndian.Uint32(data[32:36])

	// Read total chunks
	pkt.TotalChunks = binary.LittleEndian.Uint32(data[36:40])

	// Read IsResume flag
	pkt.IsResume = data[40] != 0

	// Read last chunk index
	pkt.LastChunkIdx = int32(binary.LittleEndian.Uint32(data[41:45]))

	// Read data length
	dataLen := binary.LittleEndian.Uint32(data[45:49])

	// Validate data length
	if len(data) < int(49+dataLen) {
		return nil, fmt.Errorf("invalid data length: expected %d, got %d", dataLen, len(data)-49)
	}

	// Read data
	pkt.Data = make([]byte, dataLen)
	copy(pkt.Data, data[49:49+dataLen])

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

// VoicePacket represents voice data for transmission.
// Layout:
// - Header (16 bytes): Message ID
// - SenderID (8 bytes): Entity that sent the voice data
// - ChannelID length (1 byte): Length of channel ID string
// - ChannelID (variable): Voice channel identifier
// - SequenceNumber (4 bytes): Monotonically increasing for ordering
// - Timestamp (8 bytes): When the voice frame was captured
// - DataLen (2 bytes): Length of voice data
// - Data (variable): Encoded voice data (ADPCM)
type VoicePacket struct {
	Header         PacketHeader
	SenderID       uint64
	ChannelID      string
	SequenceNumber uint32
	Timestamp      uint64
	Data           []byte
}

// VoicePacketHeaderSize is the minimum size of a voice packet (excluding variable fields).
const VoicePacketHeaderSize = 16 + 8 + 1 + 4 + 8 + 2 // 39 bytes minimum

// SerializeVoicePacket serializes a voice packet to bytes.
func SerializeVoicePacket(pkt *VoicePacket) ([]byte, error) {
	if len(pkt.ChannelID) > 255 {
		return nil, fmt.Errorf("channel ID too long: %d bytes (max 255)", len(pkt.ChannelID))
	}
	if len(pkt.Data) > 65535 {
		return nil, fmt.Errorf("voice data too large: %d bytes (max 65535)", len(pkt.Data))
	}

	// Calculate total size
	totalSize := VoicePacketHeaderSize + len(pkt.ChannelID) + len(pkt.Data)
	buf := make([]byte, totalSize)

	offset := 0

	// Write header (message ID)
	copy(buf[offset:offset+16], pkt.Header.MessageID[:])
	offset += 16

	// Write sender ID
	binary.LittleEndian.PutUint64(buf[offset:offset+8], pkt.SenderID)
	offset += 8

	// Write channel ID length and data
	buf[offset] = byte(len(pkt.ChannelID))
	offset++
	copy(buf[offset:offset+len(pkt.ChannelID)], pkt.ChannelID)
	offset += len(pkt.ChannelID)

	// Write sequence number
	binary.LittleEndian.PutUint32(buf[offset:offset+4], pkt.SequenceNumber)
	offset += 4

	// Write timestamp
	binary.LittleEndian.PutUint64(buf[offset:offset+8], pkt.Timestamp)
	offset += 8

	// Write data length and data
	binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(len(pkt.Data)))
	offset += 2
	copy(buf[offset:], pkt.Data)

	logrus.WithFields(logrus.Fields{
		"packet_type":     "voice",
		"sender_id":       pkt.SenderID,
		"channel_id":      pkt.ChannelID,
		"sequence_number": pkt.SequenceNumber,
		"data_len":        len(pkt.Data),
		"total_size":      totalSize,
	}).Debug("Serialized voice packet")

	return buf, nil
}

// DeserializeVoicePacket deserializes a voice packet from bytes.
func DeserializeVoicePacket(data []byte) (*VoicePacket, error) {
	if len(data) < VoicePacketHeaderSize {
		return nil, fmt.Errorf("voice packet too short: %d bytes (min %d)", len(data), VoicePacketHeaderSize)
	}

	pkt := &VoicePacket{}
	offset := 0

	// Read header
	copy(pkt.Header.MessageID[:], data[offset:offset+16])
	offset += 16

	// Read sender ID
	pkt.SenderID = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Read channel ID length and data
	channelIDLen := int(data[offset])
	offset++

	if len(data) < offset+channelIDLen+4+8+2 {
		return nil, fmt.Errorf("voice packet truncated: expected %d bytes for channel ID", channelIDLen)
	}

	pkt.ChannelID = string(data[offset : offset+channelIDLen])
	offset += channelIDLen

	// Read sequence number
	pkt.SequenceNumber = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Read timestamp
	pkt.Timestamp = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Read data length and data
	dataLen := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	if len(data) < offset+dataLen {
		return nil, fmt.Errorf("voice packet data truncated: expected %d bytes, got %d", dataLen, len(data)-offset)
	}

	pkt.Data = make([]byte, dataLen)
	copy(pkt.Data, data[offset:offset+dataLen])

	logrus.WithFields(logrus.Fields{
		"packet_type":     "voice",
		"sender_id":       pkt.SenderID,
		"channel_id":      pkt.ChannelID,
		"sequence_number": pkt.SequenceNumber,
		"data_len":        dataLen,
	}).Debug("Deserialized voice packet")

	return pkt, nil
}
