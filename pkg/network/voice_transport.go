// Package network provides voice transport for multiplayer voice chat.
// This file implements VoiceTransport which handles voice packet transmission
// over TCP connections with jitter buffering for smooth playback.
package network

import (
	"container/heap"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/opd-ai/venture/pkg/audio"
	"github.com/sirupsen/logrus"
)

// VoiceTransportConfig configures voice transport behavior.
type VoiceTransportConfig struct {
	// JitterBufferSize is the number of packets to buffer for reordering.
	// Larger values handle more jitter but add latency.
	JitterBufferSize int

	// JitterBufferDelayMs is the target delay for the jitter buffer in milliseconds.
	// Packets are held for this duration before being delivered.
	JitterBufferDelayMs int

	// MaxPacketsPerSecond limits outgoing voice packets per second.
	MaxPacketsPerSecond int

	// DropOldPackets drops packets older than this many sequence numbers.
	DropOldPackets uint32
}

// DefaultVoiceTransportConfig returns sensible defaults for voice transport.
func DefaultVoiceTransportConfig() VoiceTransportConfig {
	return VoiceTransportConfig{
		JitterBufferSize:    32,
		JitterBufferDelayMs: 60,
		MaxPacketsPerSecond: 50,
		DropOldPackets:      100,
	}
}

// HighLatencyVoiceTransportConfig returns configuration for high-latency networks.
func HighLatencyVoiceTransportConfig() VoiceTransportConfig {
	return VoiceTransportConfig{
		JitterBufferSize:    64,
		JitterBufferDelayMs: 200,
		MaxPacketsPerSecond: 50,
		DropOldPackets:      200,
	}
}

// voicePacketHeapItem is a heap item for jitter buffer ordering.
type voicePacketHeapItem struct {
	packet    *VoicePacket
	timestamp time.Time
	index     int
}

// voicePacketHeap implements heap.Interface for voice packet ordering.
type voicePacketHeap []*voicePacketHeapItem

func (h voicePacketHeap) Len() int { return len(h) }
func (h voicePacketHeap) Less(i, j int) bool {
	return h[i].packet.SequenceNumber < h[j].packet.SequenceNumber
}
func (h voicePacketHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *voicePacketHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*voicePacketHeapItem)
	item.index = n
	*h = append(*h, item)
}

func (h *voicePacketHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

// TCPVoiceTransport implements VoiceTransport over TCP connections.
// It provides reliable voice data transmission with jitter buffering.
type TCPVoiceTransport struct {
	config   VoiceTransportConfig
	playerID uint64

	// Connection for sending voice data
	sendFunc func(data []byte) error

	// Sequence tracking
	sendSeq atomic.Uint32

	// Jitter buffer for received packets (per channel:sender)
	jitterBuffers   map[string]*jitterBuffer
	jitterBuffersMu sync.RWMutex

	// Received packets ready for delivery
	receiveQueue chan *VoicePacket
	receiveMu    sync.Mutex

	// Spatial audio parameters
	spatialVolume float64
	spatialPan    float64
	spatialMu     sync.RWMutex

	// Channel membership verification
	channelMembership   map[string]bool
	channelMembershipMu sync.RWMutex

	// Logger
	logger *logrus.Entry

	// Shutdown
	done   chan struct{}
	closed atomic.Bool
}

// jitterBuffer manages packet ordering for a single sender.
type jitterBuffer struct {
	heap          voicePacketHeap
	lastDelivered uint32
	mu            sync.Mutex
}

// NewTCPVoiceTransport creates a new voice transport.
func NewTCPVoiceTransport(config VoiceTransportConfig, playerID uint64, sendFunc func(data []byte) error) *TCPVoiceTransport {
	logger := logrus.WithFields(logrus.Fields{
		"component": "voice_transport",
		"player_id": playerID,
	})

	transport := &TCPVoiceTransport{
		config:            config,
		playerID:          playerID,
		sendFunc:          sendFunc,
		jitterBuffers:     make(map[string]*jitterBuffer),
		receiveQueue:      make(chan *VoicePacket, config.JitterBufferSize*4),
		spatialVolume:     1.0,
		spatialPan:        0.0,
		channelMembership: make(map[string]bool),
		logger:            logger,
		done:              make(chan struct{}),
	}

	// Start jitter buffer processing
	go transport.processJitterBuffers()

	logger.WithFields(logrus.Fields{
		"jitter_buffer_size": config.JitterBufferSize,
		"jitter_delay_ms":    config.JitterBufferDelayMs,
	}).Info("Voice transport initialized")

	return transport
}

// SendVoice transmits encoded voice data to a channel.
// Implements audio.VoiceTransport interface.
func (t *TCPVoiceTransport) SendVoice(channelID string, data []byte) error {
	if t.closed.Load() {
		return fmt.Errorf("voice transport closed")
	}

	if channelID == "" {
		return fmt.Errorf("empty channel ID")
	}

	if len(data) == 0 {
		return fmt.Errorf("empty voice data")
	}

	// Generate message ID
	msgID := uuid.New()

	// Create voice packet
	pkt := &VoicePacket{
		Header:         PacketHeader{MessageID: msgID},
		SenderID:       t.playerID,
		ChannelID:      channelID,
		SequenceNumber: t.sendSeq.Add(1),
		Timestamp:      uint64(time.Now().UnixMilli()),
		Data:           data,
	}

	// Serialize packet
	serialized, err := SerializeVoicePacket(pkt)
	if err != nil {
		t.logger.WithError(err).Error("Failed to serialize voice packet")
		return fmt.Errorf("serialize voice packet: %w", err)
	}

	// Add packet type prefix
	fullPacket := make([]byte, 1+len(serialized))
	fullPacket[0] = byte(PacketTypeVoice)
	copy(fullPacket[1:], serialized)

	// Send via the provided send function
	if t.sendFunc != nil {
		if err := t.sendFunc(fullPacket); err != nil {
			t.logger.WithError(err).WithField("channel_id", channelID).Warn("Failed to send voice packet")
			return fmt.Errorf("send voice packet: %w", err)
		}
	}

	t.logger.WithFields(logrus.Fields{
		"channel_id":      channelID,
		"sequence_number": pkt.SequenceNumber,
		"data_len":        len(data),
	}).Debug("Voice packet sent")

	return nil
}

// ReceiveVoice gets received voice data for processing.
// Implements audio.VoiceTransport interface.
// Returns channelID, senderID, data, and ok (false if no data available).
func (t *TCPVoiceTransport) ReceiveVoice() (channelID, senderID string, data []byte, ok bool) {
	select {
	case pkt := <-t.receiveQueue:
		return pkt.ChannelID, fmt.Sprintf("%d", pkt.SenderID), pkt.Data, true
	default:
		return "", "", nil, false
	}
}

// SetSpatialParams sets spatial audio parameters for transmission.
// Implements audio.VoiceTransport interface.
func (t *TCPVoiceTransport) SetSpatialParams(volume, pan float64) {
	t.spatialMu.Lock()
	defer t.spatialMu.Unlock()
	t.spatialVolume = volume
	t.spatialPan = pan
}

// GetSpatialParams returns current spatial audio parameters.
func (t *TCPVoiceTransport) GetSpatialParams() (volume, pan float64) {
	t.spatialMu.RLock()
	defer t.spatialMu.RUnlock()
	return t.spatialVolume, t.spatialPan
}

// JoinChannel marks this transport as a member of a voice channel.
func (t *TCPVoiceTransport) JoinChannel(channelID string) {
	t.channelMembershipMu.Lock()
	defer t.channelMembershipMu.Unlock()
	t.channelMembership[channelID] = true

	t.logger.WithField("channel_id", channelID).Debug("Joined voice channel")
}

// LeaveChannel removes this transport from a voice channel.
func (t *TCPVoiceTransport) LeaveChannel(channelID string) {
	t.channelMembershipMu.Lock()
	defer t.channelMembershipMu.Unlock()
	delete(t.channelMembership, channelID)

	t.logger.WithField("channel_id", channelID).Debug("Left voice channel")
}

// IsInChannel returns whether this transport is a member of a channel.
func (t *TCPVoiceTransport) IsInChannel(channelID string) bool {
	t.channelMembershipMu.RLock()
	defer t.channelMembershipMu.RUnlock()
	return t.channelMembership[channelID]
}

// HandleReceivedPacket processes a received voice packet.
// Called by the network layer when a voice packet arrives.
func (t *TCPVoiceTransport) HandleReceivedPacket(pkt *VoicePacket) error {
	if t.closed.Load() {
		return fmt.Errorf("voice transport closed")
	}

	// Verify channel membership (only receive from channels we're in)
	if !t.IsInChannel(pkt.ChannelID) {
		t.logger.WithFields(logrus.Fields{
			"channel_id": pkt.ChannelID,
			"sender_id":  pkt.SenderID,
		}).Debug("Dropping voice packet: not in channel")
		return nil
	}

	// Don't process our own voice
	if pkt.SenderID == t.playerID {
		return nil
	}

	// Add to jitter buffer
	bufferKey := fmt.Sprintf("%s:%d", pkt.ChannelID, pkt.SenderID)

	t.jitterBuffersMu.Lock()
	jb, exists := t.jitterBuffers[bufferKey]
	if !exists {
		jb = &jitterBuffer{
			heap:          make(voicePacketHeap, 0, t.config.JitterBufferSize),
			lastDelivered: 0,
		}
		heap.Init(&jb.heap)
		t.jitterBuffers[bufferKey] = jb
	}
	t.jitterBuffersMu.Unlock()

	jb.mu.Lock()
	defer jb.mu.Unlock()

	// Drop packets that are too old
	if pkt.SequenceNumber <= jb.lastDelivered && jb.lastDelivered-pkt.SequenceNumber < t.config.DropOldPackets {
		t.logger.WithFields(logrus.Fields{
			"channel_id":      pkt.ChannelID,
			"sender_id":       pkt.SenderID,
			"sequence_number": pkt.SequenceNumber,
			"last_delivered":  jb.lastDelivered,
		}).Debug("Dropping old voice packet")
		return nil
	}

	// Check buffer capacity
	if jb.heap.Len() >= t.config.JitterBufferSize {
		// Remove oldest packet to make room
		heap.Pop(&jb.heap)
		t.logger.Debug("Jitter buffer overflow, dropping oldest packet")
	}

	// Add to jitter buffer
	item := &voicePacketHeapItem{
		packet:    pkt,
		timestamp: time.Now(),
	}
	heap.Push(&jb.heap, item)

	t.logger.WithFields(logrus.Fields{
		"channel_id":      pkt.ChannelID,
		"sender_id":       pkt.SenderID,
		"sequence_number": pkt.SequenceNumber,
		"buffer_size":     jb.heap.Len(),
	}).Debug("Voice packet added to jitter buffer")

	return nil
}

// processJitterBuffers periodically delivers packets from jitter buffers.
func (t *TCPVoiceTransport) processJitterBuffers() {
	ticker := time.NewTicker(time.Duration(t.config.JitterBufferDelayMs/4) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-t.done:
			return
		case <-ticker.C:
			t.deliverBufferedPackets()
		}
	}
}

// deliverBufferedPackets delivers packets that have been buffered long enough.
func (t *TCPVoiceTransport) deliverBufferedPackets() {
	now := time.Now()
	minAge := time.Duration(t.config.JitterBufferDelayMs) * time.Millisecond

	t.jitterBuffersMu.RLock()
	buffers := make(map[string]*jitterBuffer, len(t.jitterBuffers))
	for k, v := range t.jitterBuffers {
		buffers[k] = v
	}
	t.jitterBuffersMu.RUnlock()

	for _, jb := range buffers {
		jb.mu.Lock()
		for jb.heap.Len() > 0 {
			// Peek at the oldest packet
			oldest := jb.heap[0]

			// Check if it's been buffered long enough
			age := now.Sub(oldest.timestamp)
			if age < minAge {
				break
			}

			// Pop and deliver
			item := heap.Pop(&jb.heap).(*voicePacketHeapItem)
			jb.lastDelivered = item.packet.SequenceNumber

			// Non-blocking send to receive queue
			select {
			case t.receiveQueue <- item.packet:
				t.logger.WithFields(logrus.Fields{
					"channel_id":      item.packet.ChannelID,
					"sender_id":       item.packet.SenderID,
					"sequence_number": item.packet.SequenceNumber,
				}).Debug("Voice packet delivered from jitter buffer")
			default:
				t.logger.Warn("Voice receive queue full, dropping packet")
			}
		}
		jb.mu.Unlock()
	}
}

// Close shuts down the voice transport.
func (t *TCPVoiceTransport) Close() error {
	if t.closed.Swap(true) {
		return nil // Already closed
	}

	close(t.done)

	// Clear jitter buffers
	t.jitterBuffersMu.Lock()
	t.jitterBuffers = make(map[string]*jitterBuffer)
	t.jitterBuffersMu.Unlock()

	t.logger.Info("Voice transport closed")
	return nil
}

// GetStats returns statistics about the voice transport.
func (t *TCPVoiceTransport) GetStats() VoiceTransportStats {
	t.jitterBuffersMu.RLock()
	defer t.jitterBuffersMu.RUnlock()

	totalBuffered := 0
	for _, jb := range t.jitterBuffers {
		jb.mu.Lock()
		totalBuffered += jb.heap.Len()
		jb.mu.Unlock()
	}

	return VoiceTransportStats{
		PacketsSent:     t.sendSeq.Load(),
		JitterBuffers:   len(t.jitterBuffers),
		TotalBuffered:   totalBuffered,
		ChannelCount:    len(t.channelMembership),
		ReceiveQueueLen: len(t.receiveQueue),
	}
}

// VoiceTransportStats contains statistics about voice transport performance.
type VoiceTransportStats struct {
	PacketsSent     uint32
	JitterBuffers   int
	TotalBuffered   int
	ChannelCount    int
	ReceiveQueueLen int
}

// Compile-time interface check
var _ audio.VoiceTransport = (*TCPVoiceTransport)(nil)
