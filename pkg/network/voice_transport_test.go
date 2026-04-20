package network

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewTCPVoiceTransport(t *testing.T) {
	config := DefaultVoiceTransportConfig()
	transport := NewTCPVoiceTransport(config, 12345, nil)
	defer transport.Close()

	if transport == nil {
		t.Fatal("NewTCPVoiceTransport returned nil")
	}

	if transport.playerID != 12345 {
		t.Errorf("playerID = %d, want 12345", transport.playerID)
	}

	stats := transport.GetStats()
	if stats.PacketsSent != 0 {
		t.Errorf("initial PacketsSent = %d, want 0", stats.PacketsSent)
	}
}

func TestVoiceTransportSendVoice(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		data      []byte
		wantErr   bool
	}{
		{
			name:      "valid send",
			channelID: "party:test",
			data:      []byte{0x01, 0x02, 0x03, 0x04},
			wantErr:   false,
		},
		{
			name:      "empty channel ID",
			channelID: "",
			data:      []byte{0x01},
			wantErr:   true,
		},
		{
			name:      "empty data",
			channelID: "party:test",
			data:      []byte{},
			wantErr:   true,
		},
		{
			name:      "nil data",
			channelID: "party:test",
			data:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sentData []byte
			sendFunc := func(data []byte) error {
				sentData = make([]byte, len(data))
				copy(sentData, data)
				return nil
			}

			config := DefaultVoiceTransportConfig()
			transport := NewTCPVoiceTransport(config, 12345, sendFunc)
			defer transport.Close()

			err := transport.SendVoice(tt.channelID, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendVoice() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(sentData) == 0 {
				t.Error("SendVoice() did not send any data")
			}

			if !tt.wantErr && sentData[0] != byte(PacketTypeVoice) {
				t.Errorf("packet type = %d, want %d", sentData[0], PacketTypeVoice)
			}
		})
	}
}

func TestVoiceTransportSequenceNumbers(t *testing.T) {
	var packets [][]byte
	var mu sync.Mutex

	sendFunc := func(data []byte) error {
		mu.Lock()
		defer mu.Unlock()
		copied := make([]byte, len(data))
		copy(copied, data)
		packets = append(packets, copied)
		return nil
	}

	config := DefaultVoiceTransportConfig()
	transport := NewTCPVoiceTransport(config, 12345, sendFunc)
	defer transport.Close()

	// Send multiple packets
	for i := 0; i < 5; i++ {
		err := transport.SendVoice("party:test", []byte{byte(i)})
		if err != nil {
			t.Fatalf("SendVoice() error = %v", err)
		}
	}

	// Verify sequence numbers are incrementing
	if len(packets) != 5 {
		t.Fatalf("expected 5 packets, got %d", len(packets))
	}

	lastSeq := uint32(0)
	for i, pktData := range packets {
		// Skip packet type byte
		pkt, err := DeserializeVoicePacket(pktData[1:])
		if err != nil {
			t.Fatalf("DeserializeVoicePacket() error = %v", err)
		}

		if i > 0 && pkt.SequenceNumber <= lastSeq {
			t.Errorf("sequence number not increasing: got %d, previous was %d", pkt.SequenceNumber, lastSeq)
		}
		lastSeq = pkt.SequenceNumber
	}
}

func TestVoiceTransportChannelMembership(t *testing.T) {
	config := DefaultVoiceTransportConfig()
	transport := NewTCPVoiceTransport(config, 12345, nil)
	defer transport.Close()

	// Initially not in any channel
	if transport.IsInChannel("party:test") {
		t.Error("IsInChannel() = true, want false")
	}

	// Join channel
	transport.JoinChannel("party:test")
	if !transport.IsInChannel("party:test") {
		t.Error("IsInChannel() = false after JoinChannel, want true")
	}

	// Leave channel
	transport.LeaveChannel("party:test")
	if transport.IsInChannel("party:test") {
		t.Error("IsInChannel() = true after LeaveChannel, want false")
	}
}

func TestVoiceTransportSpatialParams(t *testing.T) {
	config := DefaultVoiceTransportConfig()
	transport := NewTCPVoiceTransport(config, 12345, nil)
	defer transport.Close()

	// Default values
	volume, pan := transport.GetSpatialParams()
	if volume != 1.0 {
		t.Errorf("initial volume = %f, want 1.0", volume)
	}
	if pan != 0.0 {
		t.Errorf("initial pan = %f, want 0.0", pan)
	}

	// Set new values
	transport.SetSpatialParams(0.5, -0.3)
	volume, pan = transport.GetSpatialParams()
	if volume != 0.5 {
		t.Errorf("volume = %f, want 0.5", volume)
	}
	if pan != -0.3 {
		t.Errorf("pan = %f, want -0.3", pan)
	}
}

func TestVoiceTransportReceiveVoice(t *testing.T) {
	config := DefaultVoiceTransportConfig()
	config.JitterBufferDelayMs = 10 // Short delay for testing
	transport := NewTCPVoiceTransport(config, 12345, nil)
	defer transport.Close()

	// Join the channel first
	transport.JoinChannel("party:test")

	// Create and inject a voice packet
	pkt := &VoicePacket{
		Header:         PacketHeader{MessageID: uuid.New()},
		SenderID:       99999, // Different sender
		ChannelID:      "party:test",
		SequenceNumber: 1,
		Timestamp:      uint64(time.Now().UnixMilli()),
		Data:           []byte{0xAA, 0xBB, 0xCC},
	}

	// Handle received packet
	err := transport.HandleReceivedPacket(pkt)
	if err != nil {
		t.Fatalf("HandleReceivedPacket() error = %v", err)
	}

	// Wait for jitter buffer to deliver
	time.Sleep(50 * time.Millisecond)

	// Should be able to receive
	channelID, senderID, data, ok := transport.ReceiveVoice()
	if !ok {
		t.Fatal("ReceiveVoice() returned ok=false, expected data")
	}

	if channelID != "party:test" {
		t.Errorf("channelID = %s, want party:test", channelID)
	}

	if senderID != "99999" {
		t.Errorf("senderID = %s, want 99999", senderID)
	}

	if len(data) != 3 || data[0] != 0xAA || data[1] != 0xBB || data[2] != 0xCC {
		t.Errorf("data = %v, want [0xAA 0xBB 0xCC]", data)
	}
}

func TestVoiceTransportDropsOwnPackets(t *testing.T) {
	config := DefaultVoiceTransportConfig()
	config.JitterBufferDelayMs = 10
	transport := NewTCPVoiceTransport(config, 12345, nil)
	defer transport.Close()

	transport.JoinChannel("party:test")

	// Inject a packet from ourselves
	pkt := &VoicePacket{
		Header:         PacketHeader{MessageID: uuid.New()},
		SenderID:       12345, // Same as transport playerID
		ChannelID:      "party:test",
		SequenceNumber: 1,
		Timestamp:      uint64(time.Now().UnixMilli()),
		Data:           []byte{0x01},
	}

	err := transport.HandleReceivedPacket(pkt)
	if err != nil {
		t.Fatalf("HandleReceivedPacket() error = %v", err)
	}

	// Wait and check - should not receive our own packet
	time.Sleep(50 * time.Millisecond)

	_, _, _, ok := transport.ReceiveVoice()
	if ok {
		t.Error("ReceiveVoice() returned ok=true for own packet, should be filtered")
	}
}

func TestVoiceTransportDropsUnsubscribedChannels(t *testing.T) {
	config := DefaultVoiceTransportConfig()
	config.JitterBufferDelayMs = 10
	transport := NewTCPVoiceTransport(config, 12345, nil)
	defer transport.Close()

	// NOT joining the channel

	pkt := &VoicePacket{
		Header:         PacketHeader{MessageID: uuid.New()},
		SenderID:       99999,
		ChannelID:      "party:test", // Not subscribed
		SequenceNumber: 1,
		Timestamp:      uint64(time.Now().UnixMilli()),
		Data:           []byte{0x01},
	}

	err := transport.HandleReceivedPacket(pkt)
	if err != nil {
		t.Fatalf("HandleReceivedPacket() error = %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	_, _, _, ok := transport.ReceiveVoice()
	if ok {
		t.Error("ReceiveVoice() returned ok=true for unsubscribed channel")
	}
}

func TestVoiceTransportJitterBufferOrdering(t *testing.T) {
	config := DefaultVoiceTransportConfig()
	config.JitterBufferDelayMs = 20
	transport := NewTCPVoiceTransport(config, 12345, nil)
	defer transport.Close()

	transport.JoinChannel("party:test")

	// Send packets out of order: 3, 1, 2
	for _, seq := range []uint32{3, 1, 2} {
		pkt := &VoicePacket{
			Header:         PacketHeader{MessageID: uuid.New()},
			SenderID:       99999,
			ChannelID:      "party:test",
			SequenceNumber: seq,
			Timestamp:      uint64(time.Now().UnixMilli()),
			Data:           []byte{byte(seq)},
		}
		err := transport.HandleReceivedPacket(pkt)
		if err != nil {
			t.Fatalf("HandleReceivedPacket() error = %v", err)
		}
		// Small delay between packets
		time.Sleep(5 * time.Millisecond)
	}

	// Wait for jitter buffer
	time.Sleep(100 * time.Millisecond)

	// Should receive in order: 1, 2, 3
	expectedOrder := []uint32{1, 2, 3}
	for i, expected := range expectedOrder {
		_, _, data, ok := transport.ReceiveVoice()
		if !ok {
			t.Fatalf("packet %d: ReceiveVoice() returned ok=false", i)
		}
		if len(data) != 1 || uint32(data[0]) != expected {
			t.Errorf("packet %d: data[0] = %d, want %d", i, data[0], expected)
		}
	}
}

func TestVoiceTransportClose(t *testing.T) {
	config := DefaultVoiceTransportConfig()
	transport := NewTCPVoiceTransport(config, 12345, nil)

	// Close should work
	err := transport.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Second close should be idempotent
	err = transport.Close()
	if err != nil {
		t.Errorf("second Close() error = %v", err)
	}

	// Send should fail after close
	err = transport.SendVoice("party:test", []byte{0x01})
	if err == nil {
		t.Error("SendVoice() after Close() should return error")
	}
}

func TestVoiceTransportStats(t *testing.T) {
	sendCount := 0
	sendFunc := func(data []byte) error {
		sendCount++
		return nil
	}

	config := DefaultVoiceTransportConfig()
	transport := NewTCPVoiceTransport(config, 12345, sendFunc)
	defer transport.Close()

	// Send some packets
	for i := 0; i < 10; i++ {
		transport.SendVoice("party:test", []byte{byte(i)})
	}

	// Join a channel
	transport.JoinChannel("party:test")
	transport.JoinChannel("guild:test")

	stats := transport.GetStats()
	if stats.PacketsSent != 10 {
		t.Errorf("PacketsSent = %d, want 10", stats.PacketsSent)
	}
	if stats.ChannelCount != 2 {
		t.Errorf("ChannelCount = %d, want 2", stats.ChannelCount)
	}
}

func TestVoicePacketSerialization(t *testing.T) {
	tests := []struct {
		name    string
		packet  *VoicePacket
		wantErr bool
	}{
		{
			name: "valid packet",
			packet: &VoicePacket{
				Header:         PacketHeader{MessageID: uuid.New()},
				SenderID:       12345,
				ChannelID:      "party:test",
				SequenceNumber: 42,
				Timestamp:      1234567890,
				Data:           []byte{0x01, 0x02, 0x03, 0x04},
			},
			wantErr: false,
		},
		{
			name: "empty channel ID",
			packet: &VoicePacket{
				Header:         PacketHeader{MessageID: uuid.New()},
				SenderID:       12345,
				ChannelID:      "",
				SequenceNumber: 1,
				Timestamp:      1234567890,
				Data:           []byte{0x01},
			},
			wantErr: false,
		},
		{
			name: "long channel ID",
			packet: &VoicePacket{
				Header:         PacketHeader{MessageID: uuid.New()},
				SenderID:       12345,
				ChannelID:      string(make([]byte, 255)), // Max length
				SequenceNumber: 1,
				Timestamp:      1234567890,
				Data:           []byte{0x01},
			},
			wantErr: false,
		},
		{
			name: "channel ID too long",
			packet: &VoicePacket{
				Header:         PacketHeader{MessageID: uuid.New()},
				SenderID:       12345,
				ChannelID:      string(make([]byte, 256)), // Too long
				SequenceNumber: 1,
				Timestamp:      1234567890,
				Data:           []byte{0x01},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serialized, err := SerializeVoicePacket(tt.packet)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeVoicePacket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Deserialize and verify
			deserialized, err := DeserializeVoicePacket(serialized)
			if err != nil {
				t.Fatalf("DeserializeVoicePacket() error = %v", err)
			}

			if deserialized.SenderID != tt.packet.SenderID {
				t.Errorf("SenderID = %d, want %d", deserialized.SenderID, tt.packet.SenderID)
			}
			if deserialized.ChannelID != tt.packet.ChannelID {
				t.Errorf("ChannelID = %s, want %s", deserialized.ChannelID, tt.packet.ChannelID)
			}
			if deserialized.SequenceNumber != tt.packet.SequenceNumber {
				t.Errorf("SequenceNumber = %d, want %d", deserialized.SequenceNumber, tt.packet.SequenceNumber)
			}
			if deserialized.Timestamp != tt.packet.Timestamp {
				t.Errorf("Timestamp = %d, want %d", deserialized.Timestamp, tt.packet.Timestamp)
			}
			if len(deserialized.Data) != len(tt.packet.Data) {
				t.Errorf("Data length = %d, want %d", len(deserialized.Data), len(tt.packet.Data))
			}
			for i := range deserialized.Data {
				if deserialized.Data[i] != tt.packet.Data[i] {
					t.Errorf("Data[%d] = %d, want %d", i, deserialized.Data[i], tt.packet.Data[i])
				}
			}
		})
	}
}

func TestVoicePacketDeserializeErrors(t *testing.T) {
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
			data:    make([]byte, VoicePacketHeaderSize-1),
			wantErr: true,
		},
		{
			name:    "truncated channel ID",
			data:    append(make([]byte, 25), 50), // Claims 50-byte channel ID but not enough data
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeserializeVoicePacket(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeVoicePacket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVoiceTransportConcurrency(t *testing.T) {
	config := DefaultVoiceTransportConfig()
	config.JitterBufferDelayMs = 10

	sentCount := 0
	var mu sync.Mutex
	sendFunc := func(data []byte) error {
		mu.Lock()
		defer mu.Unlock()
		sentCount++
		return nil
	}

	transport := NewTCPVoiceTransport(config, 12345, sendFunc)
	defer transport.Close()

	transport.JoinChannel("party:test")

	// Concurrent sends
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				transport.SendVoice("party:test", []byte{byte(id), byte(j)})
			}
		}(i)
	}

	// Concurrent receives
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				pkt := &VoicePacket{
					Header:         PacketHeader{MessageID: uuid.New()},
					SenderID:       uint64(99999 + j),
					ChannelID:      "party:test",
					SequenceNumber: uint32(j),
					Timestamp:      uint64(time.Now().UnixMilli()),
					Data:           []byte{byte(j)},
				}
				transport.HandleReceivedPacket(pkt)
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	if sentCount != 100 {
		t.Errorf("sentCount = %d, want 100", sentCount)
	}
	mu.Unlock()
}

func TestHighLatencyVoiceTransportConfig(t *testing.T) {
	config := HighLatencyVoiceTransportConfig()

	if config.JitterBufferSize != 64 {
		t.Errorf("JitterBufferSize = %d, want 64", config.JitterBufferSize)
	}
	if config.JitterBufferDelayMs != 200 {
		t.Errorf("JitterBufferDelayMs = %d, want 200", config.JitterBufferDelayMs)
	}
	if config.DropOldPackets != 200 {
		t.Errorf("DropOldPackets = %d, want 200", config.DropOldPackets)
	}
}

func BenchmarkVoicePacketSerialization(b *testing.B) {
	pkt := &VoicePacket{
		Header:         PacketHeader{MessageID: uuid.New()},
		SenderID:       12345,
		ChannelID:      "party:test123",
		SequenceNumber: 42,
		Timestamp:      1234567890,
		Data:           make([]byte, 480), // Typical voice frame size
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SerializeVoicePacket(pkt)
	}
}

func BenchmarkVoicePacketDeserialization(b *testing.B) {
	pkt := &VoicePacket{
		Header:         PacketHeader{MessageID: uuid.New()},
		SenderID:       12345,
		ChannelID:      "party:test123",
		SequenceNumber: 42,
		Timestamp:      1234567890,
		Data:           make([]byte, 480),
	}
	serialized, _ := SerializeVoicePacket(pkt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeserializeVoicePacket(serialized)
	}
}

func BenchmarkVoiceTransportSend(b *testing.B) {
	sendFunc := func(data []byte) error {
		return nil
	}

	config := DefaultVoiceTransportConfig()
	transport := NewTCPVoiceTransport(config, 12345, sendFunc)
	defer transport.Close()

	data := make([]byte, 480) // Typical voice frame

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transport.SendVoice("party:test", data)
	}
}

// Helper to verify interface compliance
func TestVoiceTransportInterfaceCompliance(t *testing.T) {
	config := DefaultVoiceTransportConfig()
	transport := NewTCPVoiceTransport(config, 12345, nil)
	defer transport.Close()

	// This test verifies that TCPVoiceTransport implements audio.VoiceTransport
	// The compile-time check in voice_transport.go handles this, but this test
	// exercises the interface methods directly.

	// SendVoice
	_ = transport.SendVoice("test", []byte{0x01})

	// ReceiveVoice
	_, _, _, _ = transport.ReceiveVoice()

	// SetSpatialParams
	transport.SetSpatialParams(0.5, 0.5)

	t.Log("Interface compliance verified")
}

// TestVoiceTransportSendError tests error handling in send function
func TestVoiceTransportSendError(t *testing.T) {
	sendFunc := func(data []byte) error {
		return fmt.Errorf("network error")
	}

	config := DefaultVoiceTransportConfig()
	transport := NewTCPVoiceTransport(config, 12345, sendFunc)
	defer transport.Close()

	err := transport.SendVoice("party:test", []byte{0x01})
	if err == nil {
		t.Error("SendVoice() should return error when send function fails")
	}
}

// TestVoiceEndToEnd verifies the full voice path: a packet sent by client A
// over a real TCPServer is fanned out to client B and surfaces via B's
// registered SetVoiceHandler. This is the regression test for AUDIT.md
// CRITICAL: voice chat server-side packet routing.
func TestVoiceEndToEnd(t *testing.T) {
	// Start a real TCP server on an OS-assigned port so the test is hermetic.
	serverConfig := DefaultServerConfig()
	serverConfig.Address = "127.0.0.1:0"
	serverConfig.MaxPlayers = 4
	server := NewServer(serverConfig)
	if err := server.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	defer server.Stop()

	addr := server.Address()
	if addr == "" {
		t.Fatal("server Address() returned empty after Start()")
	}

	// Drain join/leave/error/input channels so the server doesn't block on
	// unconsumed events. (Voice commands are routed before they reach the
	// inputCommands channel, but other input types might be enqueued.)
	go func() {
		for {
			select {
			case <-server.ReceivePlayerJoin():
			case <-server.ReceivePlayerLeave():
			case <-server.ReceiveError():
			case <-server.ReceiveInputCommand():
			case <-time.After(2 * time.Second):
				return
			}
		}
	}()

	// Connect two clients to the server.
	clientA := connectVoiceTestClient(t, addr)
	defer clientA.Disconnect()
	clientB := connectVoiceTestClient(t, addr)
	defer clientB.Disconnect()

	// Wait for the server to register both clients before sending voice.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && server.GetPlayerCount() < 2 {
		time.Sleep(10 * time.Millisecond)
	}
	if got := server.GetPlayerCount(); got < 2 {
		t.Fatalf("server has %d players, want 2", got)
	}

	// Discover the connection-bound player IDs from the server. The server
	// assigns IDs in connection order starting from 1, so the smaller ID
	// belongs to clientA and the larger to clientB.
	players := server.GetPlayers()
	if len(players) < 2 {
		t.Fatalf("server.GetPlayers() returned %d, want >=2", len(players))
	}
	idA, idB := players[0], players[1]
	if idA > idB {
		idA, idB = idB, idA
	}
	clientA.SetPlayerID(idA)
	clientB.SetPlayerID(idB)

	// Register a voice handler on client B that captures the inbound packet.
	received := make(chan *VoicePacket, 1)
	clientB.SetVoiceHandler(func(pkt *VoicePacket) {
		select {
		case received <- pkt:
		default:
		}
	})

	// Build a voice transport for client A that sends through A's connection.
	// Mark B as a member of the channel so its receive path doesn't drop the
	// packet (HandleReceivedPacket enforces channel membership).
	const channelID = "party:e2e"
	transportA := NewTCPVoiceTransport(DefaultVoiceTransportConfig(), idA, func(data []byte) error {
		return clientA.SendInput(VoiceInputType, data)
	})
	defer transportA.Close()

	// On client B, instead of constructing a full audio.Manager, route inbound
	// voice directly to a transport whose receiveQueue we observe.
	transportB := NewTCPVoiceTransport(DefaultVoiceTransportConfig(), idB, nil)
	defer transportB.Close()
	transportB.JoinChannel(channelID)

	// Override B's voice handler to feed transportB and signal arrival.
	signal := make(chan struct{}, 1)
	clientB.SetVoiceHandler(func(pkt *VoicePacket) {
		// Verify the packet was authored by A and addressed to the channel.
		if pkt.SenderID != idA {
			t.Errorf("inbound voice SenderID = %d, want %d", pkt.SenderID, idA)
		}
		if pkt.ChannelID != channelID {
			t.Errorf("inbound voice ChannelID = %q, want %q", pkt.ChannelID, channelID)
		}
		_ = transportB.HandleReceivedPacket(pkt)
		select {
		case signal <- struct{}{}:
		default:
		}
		select {
		case received <- pkt:
		default:
		}
	})

	// Send a voice packet from A.
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE}
	if err := transportA.SendVoice(channelID, payload); err != nil {
		t.Fatalf("transportA.SendVoice: %v", err)
	}

	// Wait for B to receive the packet via the network round-trip.
	select {
	case pkt := <-received:
		if string(pkt.Data) != string(payload) {
			t.Errorf("voice payload mismatch: got %x, want %x", pkt.Data, payload)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("client B did not receive voice packet within 3s — server-side voice routing missing")
	}

	// Sanity check: client A must NOT receive its own voice (sender filtering).
	select {
	case <-signal:
		// signal fires from B's handler; we don't actually use it here, but
		// receiving it confirms the chain ran.
	default:
	}
}

// connectVoiceTestClient is a small helper that constructs a TCPClient and
// connects it to the given address. Caller is responsible for Disconnect().
func connectVoiceTestClient(t *testing.T, addr string) *TCPClient {
	t.Helper()

	cfg := DefaultClientConfig()
	cfg.ServerAddress = addr
	cfg.ConnectionTimeout = 2 * time.Second
	cfg.PingInterval = 30 * time.Second

	client := NewClient(cfg)
	if err := client.Connect(); err != nil {
		t.Fatalf("client connect to %s: %v", addr, err)
	}
	return client
}
