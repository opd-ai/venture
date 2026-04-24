//go:build js && wasm

// Package webrtc — WASM real-WebSocket signaling transport.
// This file implements the three platform hooks declared in signaling.go using
// the browser's native WebSocket API via syscall/js:
//
//   - connectTransport: opens the WebSocket, registers onmessage/onclose
//     callbacks that push incoming JSON messages into s.recvChan.
//   - sendViaTransport: marshals the SignalingMessage to JSON and sends it via
//     ws.send(data).
//   - closeTransport: releases the js.Func callback and closes the WebSocket.
//
// The signaling server must accept JSON-framed SignalingMessage payloads on
// the WebSocket connection.  The URL used is Config.SignalingURL; the peer ID
// is appended as a query parameter (?id=<peerID>) so the server can route
// incoming messages to the correct peer.
package webrtc

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	log "github.com/sirupsen/logrus"
)

// connectTransport opens a browser WebSocket to s.url and wires up the
// onmessage callback so that incoming JSON messages are pushed to s.recvChan.
// The connection URL includes the peer ID as a query parameter for server-side
// routing: <url>?id=<peerID>.
func (s *SignalingClient) connectTransport() error {
	wsCtor := js.Global().Get("WebSocket")
	if wsCtor.IsUndefined() || wsCtor.IsNull() {
		return fmt.Errorf("WebSocket not available in this browser environment")
	}

	wsURL := s.url
	if wsURL == "" {
		return fmt.Errorf("SignalingURL is empty; set Config.SignalingURL before connecting")
	}
	// Append peer ID as query param so the server can address messages to this peer.
	if len(wsURL) > 0 && wsURL[len(wsURL)-1] != '?' {
		wsURL += "?id=" + s.peerID
	}

	ws := wsCtor.New(wsURL)

	// onmessage: deserialize incoming JSON and push to recvChan.
	onMsg := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return nil
		}
		data := args[0].Get("data").String()
		var msg SignalingMessage
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			log.WithFields(log.Fields{
				"peer_id": s.peerID,
				"error":   err,
			}).Warn("signaling: failed to parse incoming WebSocket message")
			return nil
		}
		select {
		case s.recvChan <- &msg:
		default:
			log.WithField("peer_id", s.peerID).Warn("signaling: recv channel full, dropping incoming message")
		}
		return nil
	})

	// onclose: mark the client as disconnected if the server drops the connection.
	onClose := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		s.mu.Lock()
		s.connected = false
		s.mu.Unlock()
		log.WithField("peer_id", s.peerID).Debug("signaling: WebSocket connection closed by server")
		return nil
	})

	ws.Set("onmessage", onMsg)
	ws.Set("onclose", onClose)

	s.mu.Lock()
	s.wsConn = ws
	s.wsOnMsgFunc = onMsg
	// onClose is not stored; it only needs to exist while the ws is alive.
	// We hold onMsg for Release() in closeTransport.
	s.mu.Unlock()

	_ = onClose // kept alive until ws fires onclose
	return nil
}

// sendViaTransport marshals msg to JSON and sends it over the browser WebSocket.
func (s *SignalingClient) sendViaTransport(msg *SignalingMessage) {
	s.mu.RLock()
	ws := s.wsConn
	s.mu.RUnlock()

	wsVal, ok := ws.(js.Value)
	if !ok || wsVal.IsNull() || wsVal.IsUndefined() {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.WithFields(log.Fields{
			"peer_id": s.peerID,
			"error":   err,
		}).Warn("signaling: failed to marshal outgoing message")
		return
	}

	// WebSocket.readyState == 1 means OPEN.
	if wsVal.Get("readyState").Int() != 1 {
		log.WithField("peer_id", s.peerID).Warn("signaling: WebSocket not open, dropping outgoing message")
		return
	}

	wsVal.Call("send", string(data))
}

// closeTransport releases the js.Func callback and closes the WebSocket.
func (s *SignalingClient) closeTransport() {
	s.mu.Lock()
	ws := s.wsConn
	fn := s.wsOnMsgFunc
	s.wsConn = nil
	s.wsOnMsgFunc = nil
	s.mu.Unlock()

	if f, ok := fn.(js.Func); ok {
		f.Release()
	}
	if wsVal, ok := ws.(js.Value); ok && !wsVal.IsNull() && !wsVal.IsUndefined() {
		wsVal.Call("close")
	}
}
