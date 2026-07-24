// Package webrtc STUN client implementation.
// This file implements STUN (Session Traversal Utilities for NAT) protocol client
// for public IP discovery and NAT type detection per RFC 5389.
package webrtc

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// STUNClient handles STUN protocol for NAT traversal.
type STUNClient struct {
	mu      sync.RWMutex
	servers []string

	// Cached public address
	cachedPublicIP   net.IP
	cachedPublicPort int
	cacheExpiry      time.Time

	// Statistics
	totalRequests      int
	successfulRequests int
	failedRequests     int

	// timeProvider abstracts time access for deterministic testing.
	timeProvider TimeProvider
}

// NewSTUNClient creates a new STUN client with the given servers.
func NewSTUNClient(servers []string) *STUNClient {
	if len(servers) == 0 {
		servers = []string{"stun:stun.l.google.com:19302"}
	}

	return &STUNClient{
		servers:      servers,
		timeProvider: DefaultTimeProvider(),
	}
}

// STUNResponse represents the response from a STUN server.
type STUNResponse struct {
	// PublicIP is the public IP address as seen by the STUN server.
	PublicIP net.IP
	// PublicPort is the public port as seen by the STUN server.
	PublicPort int
	// Server is the STUN server that responded.
	Server string
	// RTT is the round-trip time to the STUN server.
	RTT time.Duration
	// NATType is the detected NAT type.
	NATType NATType
}

// NATType represents the type of NAT detected.
type NATType int

// String returns the string representation of NATType.
func (t NATType) String() string {
	switch t {
	case NATTypeUnknown:
		return "Unknown"
	case NATTypeNone:
		return "None"
	case NATTypeFullCone:
		return "FullCone"
	case NATTypeRestrictedCone:
		return "RestrictedCone"
	case NATTypePortRestrictedCone:
		return "PortRestrictedCone"
	case NATTypeSymmetric:
		return "Symmetric"
	default:
		return "Unknown"
	}
}

// GetPublicAddress discovers the public IP and port using STUN.
func (s *STUNClient) GetPublicAddress(ctx context.Context) (*STUNResponse, error) {
	s.mu.RLock()
	if s.cachedPublicIP != nil && s.timeProvider.Now().Before(s.cacheExpiry) {
		cached := &STUNResponse{
			PublicIP:   s.cachedPublicIP,
			PublicPort: s.cachedPublicPort,
			NATType:    NATTypeUnknown,
		}
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	s.totalRequests++
	s.mu.Unlock()

	for _, server := range s.servers {
		resp, err := s.querySTUNServer(ctx, server)
		if err != nil {
			continue
		}

		s.mu.Lock()
		s.successfulRequests++
		s.cachedPublicIP = resp.PublicIP
		s.cachedPublicPort = resp.PublicPort
		s.cacheExpiry = s.timeProvider.Now().Add(5 * time.Minute)
		s.mu.Unlock()

		log.WithFields(log.Fields{
			"server":      server,
			"public_ip":   resp.PublicIP,
			"public_port": resp.PublicPort,
			"rtt":         resp.RTT,
		}).Debug("STUN public address discovered")

		return resp, nil
	}

	s.mu.Lock()
	s.failedRequests++
	s.mu.Unlock()

	return nil, ErrSTUNServerUnreachable
}

// querySTUNServer sends a STUN Binding Request to a specific server per RFC 5389.
func (s *STUNClient) querySTUNServer(ctx context.Context, server string) (*STUNResponse, error) {
	start := s.timeProvider.Now()

	// Parse STUN URL: stun:host:port
	if len(server) < 6 || server[:5] != "stun:" {
		return nil, fmt.Errorf("invalid STUN server URL: %s", server)
	}

	hostPort := server[5:]

	// Create UDP connection
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp", hostPort)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to STUN server: %w", err)
	}
	defer conn.Close()

	// Generate transaction ID (12 bytes per RFC 5389)
	var transactionID [12]byte
	if _, err := rand.Read(transactionID[:]); err != nil {
		return nil, fmt.Errorf("failed to generate transaction ID: %w", err)
	}

	// Build STUN Binding Request (RFC 5389)
	// Message Type: 0x0001 (Binding Request)
	// Message Length: 0 (no attributes)
	// Magic Cookie: 0x2112A442
	// Transaction ID: 12 bytes
	request := make([]byte, 20)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)     // Binding Request
	binary.BigEndian.PutUint16(request[2:4], 0)          // Message Length
	binary.BigEndian.PutUint32(request[4:8], 0x2112A442) // Magic Cookie
	copy(request[8:20], transactionID[:])

	// Send request
	if _, err := conn.Write(request); err != nil {
		return nil, fmt.Errorf("failed to send STUN request: %w", err)
	}

	// Set read deadline based on context
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetReadDeadline(deadline)
	} else {
		_ = conn.SetReadDeadline(s.timeProvider.Now().Add(3 * time.Second))
	}

	// Read response
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read STUN response: %w", err)
	}
	response = response[:n]

	// Parse STUN response
	resp, err := parseSTUNResponse(response, transactionID, hostPort, start, s.timeProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to parse STUN response: %w", err)
	}

	return resp, nil
}

// DetectNATType performs NAT type detection using STUN.
func (s *STUNClient) DetectNATType(ctx context.Context) (NATType, error) {
	if len(s.servers) < 2 {
		return NATTypeUnknown, fmt.Errorf("need at least 2 STUN servers for NAT detection")
	}

	// Query first server
	resp1, err := s.querySTUNServer(ctx, s.servers[0])
	if err != nil {
		return NATTypeUnknown, err
	}

	// Query second server
	resp2, err := s.querySTUNServer(ctx, s.servers[1])
	if err != nil {
		return NATTypeUnknown, err
	}

	// Compare responses to determine NAT type
	if resp1.PublicIP.Equal(resp2.PublicIP) && resp1.PublicPort == resp2.PublicPort {
		// Same public address from different servers
		// Likely Full Cone or Restricted Cone NAT
		return NATTypeFullCone, nil
	}

	// Different mappings for different destinations
	// Indicates Symmetric NAT
	return NATTypeSymmetric, nil
}

// GetStats returns STUN client statistics.
func (s *STUNClient) GetStats() STUNStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var successRate float64
	if s.totalRequests > 0 {
		successRate = float64(s.successfulRequests) / float64(s.totalRequests)
	}

	return STUNStats{
		TotalRequests:      s.totalRequests,
		SuccessfulRequests: s.successfulRequests,
		FailedRequests:     s.failedRequests,
		SuccessRate:        successRate,
		CachedPublicIP:     s.cachedPublicIP,
		CacheValid:         s.timeProvider.Now().Before(s.cacheExpiry),
	}
}

// STUNStats holds STUN client statistics.
type STUNStats struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	SuccessRate        float64
	CachedPublicIP     net.IP
	CacheValid         bool
}

// parseSTUNResponse parses a STUN Binding Response per RFC 5389.
func parseSTUNResponse(data []byte, expectedTransactionID [12]byte, server string, start time.Time, tp TimeProvider) (*STUNResponse, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	// Parse header
	msgType := binary.BigEndian.Uint16(data[0:2])
	msgLength := binary.BigEndian.Uint16(data[2:4])
	magicCookie := binary.BigEndian.Uint32(data[4:8])
	var transactionID [12]byte
	copy(transactionID[:], data[8:20])

	// Validate message type: Binding Response (0x0101) or Binding Error Response (0x0111)
	if msgType != 0x0101 && msgType != 0x0111 {
		return nil, fmt.Errorf("unexpected message type: 0x%04x", msgType)
	}

	// Validate magic cookie
	if magicCookie != 0x2112A442 {
		return nil, fmt.Errorf("invalid magic cookie: 0x%08x", magicCookie)
	}

	// Validate transaction ID
	if transactionID != expectedTransactionID {
		return nil, fmt.Errorf("transaction ID mismatch")
	}

	// Validate message length
	if len(data) < 20+int(msgLength) {
		return nil, fmt.Errorf("message length exceeds data: %d vs %d", msgLength, len(data)-20)
	}

	// Check for error response
	if msgType == 0x0111 {
		return nil, parseSTUNError(data[20 : 20+msgLength])
	}

	// Parse attributes to find XOR-MAPPED-ADDRESS (0x0020) or MAPPED-ADDRESS (0x0001)
	attrs := data[20 : 20+msgLength]
	var publicIP net.IP
	var publicPort int

	for len(attrs) >= 4 {
		attrType := binary.BigEndian.Uint16(attrs[0:2])
		attrLength := binary.BigEndian.Uint16(attrs[2:4])

		if len(attrs) < 4+int(attrLength) {
			return nil, fmt.Errorf("attribute length exceeds remaining data")
		}

		attrValue := attrs[4 : 4+attrLength]

		// XOR-MAPPED-ADDRESS (0x0020) or MAPPED-ADDRESS (0x0001)
		if attrType == 0x0020 || attrType == 0x0001 {
			if len(attrValue) >= 8 {
				family := binary.BigEndian.Uint16(attrValue[0:2])
				port := binary.BigEndian.Uint16(attrValue[2:4])

				if family == 0x01 { // IPv4
					if len(attrValue) >= 8 {
						ip := net.IPv4(attrValue[4], attrValue[5], attrValue[6], attrValue[7])
						if attrType == 0x0020 {
							// XOR-MAPPED-ADDRESS: XOR with magic cookie and transaction ID
							publicIP = xorIPv4(ip, magicCookie, transactionID[:4])
							publicPort = int(port ^ uint16(magicCookie>>16))
						} else {
							// MAPPED-ADDRESS: no XOR
							publicIP = ip
							publicPort = int(port)
						}
					}
				} else if family == 0x02 { // IPv6
					if len(attrValue) >= 20 {
						ip := net.IP(attrValue[4:20])
						if attrType == 0x0020 {
							publicIP = xorIPv6(ip, magicCookie, transactionID[:])
							publicPort = int(port ^ uint16(magicCookie>>16))
						} else {
							publicIP = ip
							publicPort = int(port)
						}
					}
				}
				break
			}
		}

		// Move to next attribute (padded to 4-byte boundary)
		padding := int((4 - (attrLength % 4)) % 4)
		attrs = attrs[4+int(attrLength)+padding:]
	}

	if publicIP == nil {
		return nil, fmt.Errorf("XOR-MAPPED-ADDRESS not found in response")
	}

	rtt := tp.Now().Sub(start)

	return &STUNResponse{
		PublicIP:   publicIP,
		PublicPort: publicPort,
		Server:     server,
		RTT:        rtt,
		NATType:    NATTypeUnknown,
	}, nil
}

// xorIPv4 applies XOR decoding for XOR-MAPPED-ADDRESS (IPv4) per RFC 5389.
func xorIPv4(ip net.IP, magicCookie uint32, transactionIDPrefix []byte) net.IP {
	result := make(net.IP, 4)
	cookieBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(cookieBytes, magicCookie)
	for i := 0; i < 4; i++ {
		result[i] = ip[i] ^ cookieBytes[i] ^ transactionIDPrefix[i]
	}
	return result
}

// xorIPv6 applies XOR decoding for XOR-MAPPED-ADDRESS (IPv6) per RFC 5389.
func xorIPv6(ip net.IP, magicCookie uint32, transactionID []byte) net.IP {
	result := make(net.IP, 16)
	cookieBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(cookieBytes, magicCookie)
	// First 4 bytes XOR with magic cookie
	for i := 0; i < 4; i++ {
		result[i] = ip[i] ^ cookieBytes[i]
	}
	// Remaining 12 bytes XOR with transaction ID
	for i := 4; i < 16; i++ {
		result[i] = ip[i] ^ transactionID[i-4]
	}
	return result
}

// parseSTUNError parses ERROR-CODE attribute from STUN error response.
func parseSTUNError(attrs []byte) error {
	for len(attrs) >= 4 {
		attrType := binary.BigEndian.Uint16(attrs[0:2])
		attrLength := binary.BigEndian.Uint16(attrs[2:4])

		if len(attrs) < 4+int(attrLength) {
			break
		}

		attrValue := attrs[4 : 4+attrLength]

		// ERROR-CODE attribute (0x0009)
		if attrType == 0x0009 && len(attrValue) >= 4 {
			class := attrValue[0]
			number := attrValue[1]
			reason := string(attrValue[4:])
			return fmt.Errorf("STUN error %d%d: %s", class, number, reason)
		}

		padding := int((4 - (attrLength % 4)) % 4)
		attrs = attrs[4+int(attrLength)+padding:]
	}
	return fmt.Errorf("STUN error response without ERROR-CODE attribute")
}
