//go:build js && wasm

// Package webrtc — WASM real-WebRTC connection implementation.
// This file provides real browser-to-browser P2P connections in WASM builds
// using github.com/pion/webrtc/v3, which delegates to the browser's native
// RTCPeerConnection API when compiled for js/wasm.
//
// Connection flow:
//  1. Create pion PeerConnection with STUN config from p.Config.
//  2. Open a data channel for federation messages.
//  3. Generate SDP offer; wait for ICE gathering to complete.
//  4. Exchange SDP with remote peer via the SignalingClient.
//  5. Set remote description; pion fires StateConnected when ICE completes.
//  6. processMessages loop forwards p.sendChan → DataChannel.Send().
package webrtc

import (
	"context"
	"fmt"

	pionwebrtc "github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/recovery"
)

// launchConnectionAttempt starts the real pion-based WebRTC connection in a
// background goroutine.  On WASM the goroutine calls into the browser's
// RTCPeerConnection API via pion's WASM shim.
func (p *Peer) launchConnectionAttempt(ctx context.Context, remotePeerID string) {
	go func() {
		defer recovery.RecoverPanicWithLogger("webrtc_peer", "pion connection", nil)()
		if err := p.pionConnect(ctx, remotePeerID); err != nil {
			log.WithFields(log.Fields{
				"peer_id":        p.ID,
				"remote_peer_id": remotePeerID,
				"error":          err,
			}).Error("pion WebRTC connection failed")
		}
	}()
}

// pionConnect establishes a real WebRTC peer connection using pion/webrtc/v3.
// Steps mirror the standard WebRTC offer/answer handshake.
func (p *Peer) pionConnect(ctx context.Context, remotePeerID string) error {
	// 1. Build pion ICE/STUN configuration.
	iceServers := make([]pionwebrtc.ICEServer, 0, len(p.Config.STUNServers)+len(p.Config.TURNServers))
	iceServers = append(iceServers, pionwebrtc.ICEServer{URLs: p.Config.STUNServers})
	for _, turn := range p.Config.TURNServers {
		iceServers = append(iceServers, pionwebrtc.ICEServer{
			URLs:       turn.URLs,
			Username:   turn.Username,
			Credential: turn.Credential,
		})
	}

	pc, err := pionwebrtc.NewPeerConnection(pionwebrtc.Configuration{
		ICEServers: iceServers,
	})
	if err != nil {
		p.setState(StateFailed)
		return fmt.Errorf("create pion PeerConnection: %w", err)
	}

	// 2. Open a data channel for federation messages.
	dc, err := pc.CreateDataChannel(p.Config.DataChannelLabel, nil)
	if err != nil {
		pc.Close() //nolint:errcheck
		p.setState(StateFailed)
		return fmt.Errorf("create data channel: %w", err)
	}

	// 3. Route incoming data-channel messages into recvChan.
	dc.OnMessage(func(msg pionwebrtc.DataChannelMessage) {
		payload := make([]byte, len(msg.Data))
		copy(payload, msg.Data)
		select {
		case p.recvChan <- payload:
		default:
			log.WithField("peer_id", p.ID).Warn("WebRTC recv channel full, dropping message")
		}
	})

	// 4. Mirror pion connection-state changes to our ConnectionState enum.
	pc.OnConnectionStateChange(func(state pionwebrtc.PeerConnectionState) {
		switch state {
		case pionwebrtc.PeerConnectionStateConnected:
			p.mu.Lock()
			p.peerConn = pc
			p.dataChannel = dc
			p.mu.Unlock()
			p.setState(StateConnected)
			go func() {
				defer recovery.RecoverPanicWithLogger("webrtc_peer", "process messages", nil)()
				p.processMessages()
			}()
		case pionwebrtc.PeerConnectionStateDisconnected,
			pionwebrtc.PeerConnectionStateFailed:
			p.setState(StateDisconnected)
		case pionwebrtc.PeerConnectionStateClosed:
			p.setState(StateClosed)
		}
	})

	// 5. Create SDP offer and wait for ICE gathering to finish.
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		pc.Close() //nolint:errcheck
		p.setState(StateFailed)
		return fmt.Errorf("create SDP offer: %w", err)
	}

	gatherDone := pionwebrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(offer); err != nil {
		pc.Close() //nolint:errcheck
		p.setState(StateFailed)
		return fmt.Errorf("set local description: %w", err)
	}

	select {
	case <-gatherDone:
	case <-ctx.Done():
		pc.Close() //nolint:errcheck
		p.setState(StateFailed)
		return fmt.Errorf("ICE gathering cancelled: %w", ctx.Err())
	}

	// 6. Exchange SDP via the signaling server.
	sigClient := NewSignalingClient(p.Config.SignalingURL, p.ID)
	if err := sigClient.Connect(); err != nil {
		pc.Close() //nolint:errcheck
		p.setState(StateFailed)
		return fmt.Errorf("connect signaling: %w", err)
	}
	defer sigClient.Close() //nolint:errcheck

	localDesc := pc.LocalDescription()
	sdpOffer := &SDPOffer{Type: localDesc.Type.String(), SDP: localDesc.SDP}
	if err := sigClient.SendOffer(remotePeerID, sdpOffer); err != nil {
		pc.Close() //nolint:errcheck
		p.setState(StateFailed)
		return fmt.Errorf("send SDP offer: %w", err)
	}

	// 7. Wait for the SDP answer from the remote peer.
	sigRecv := sigClient.Receive()
	for {
		select {
		case msg, ok := <-sigRecv:
			if !ok {
				pc.Close() //nolint:errcheck
				p.setState(StateFailed)
				return fmt.Errorf("signaling channel closed before answer")
			}
			if msg == nil || msg.Type != "answer" || msg.Answer == nil {
				continue
			}
			remoteDesc := pionwebrtc.SessionDescription{
				Type: pionwebrtc.SDPTypeAnswer,
				SDP:  msg.Answer.SDP,
			}
			if err := pc.SetRemoteDescription(remoteDesc); err != nil {
				pc.Close() //nolint:errcheck
				p.setState(StateFailed)
				return fmt.Errorf("set remote description: %w", err)
			}
			// Connection-state callback will fire StateConnected when ICE completes.
			return nil
		case <-ctx.Done():
			pc.Close() //nolint:errcheck
			p.setState(StateFailed)
			return fmt.Errorf("connection cancelled waiting for answer: %w", ctx.Err())
		}
	}
}

// trySend forwards data to the pion DataChannel when the WASM connection is
// active.  Stats have already been updated in Send() before enqueue.
func (p *Peer) trySend(data []byte) {
	p.mu.RLock()
	dc := p.dataChannel
	p.mu.RUnlock()

	if dc == nil {
		return
	}
	pionDC, ok := dc.(*pionwebrtc.DataChannel)
	if !ok {
		return
	}
	if err := pionDC.Send(data); err != nil {
		log.WithFields(log.Fields{
			"peer_id": p.ID,
			"error":   err,
		}).Warn("WebRTC data channel send failed")
	}
}
