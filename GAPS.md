# Security Gaps — 2026-04-21

This file documents security gaps identified during the 2026-04-21 audit. Each gap describes the delta between what the project claims and what is currently implemented, the exploitable risk, and specific controls needed to close the gap.

---

## Gap 1 — Observability Endpoints Lack Authentication

- **Stated Goal**: Prometheus metrics and health checks are an operational tool for server operators (`pkg/observability/doc.go`).
- **Current State**: Six HTTP handlers (`/metrics`, `/health`, `/healthz`, `/ready`, `/readyz`, `/status`) are served on `:9090` (all interfaces) with no authentication, no IP allowlist, and no TLS. The `/status` and `/metrics` endpoints return runtime internals: heap bytes, goroutine count, GC cycles, player count, entity count, trade volume, FPS, memory usage.
- **Risk**: Any network-reachable host — including players connecting to the game server — can enumerate server internals without credentials. This aids reconnaissance, confirms player activity patterns, and exposes timing information.
- **Closing the Gap**:
  1. Bind metrics to `127.0.0.1:9090` by default; add a `--metrics-bind` flag for operators who need external access.
  2. If external binding is required, add a `--metrics-token` flag and enforce `Authorization: Bearer <token>` header verification on all handlers using `crypto/subtle.ConstantTimeCompare`.
  3. For TLS, either reuse the game server certificate or generate a self-signed cert at startup.
  4. Document the security posture in the metrics endpoint help text.

---

## Gap 2 — Game Server TCP Channel Is Not Authenticated at Transport Layer

- **Stated Goal**: The game provides E2E encryption for player traffic using DH key exchange + AES-256-GCM (`pkg/network/crypto.go`, `pkg/security/doc.go`).
- **Current State**: The TCP listener (`pkg/network/server.go:221`) uses plain `net.Listen("tcp", ...)`. The DH handshake is performed over this unauthenticated channel. No server certificate is presented to clients; no client verifies server identity before accepting the DH public key. The federation layer correctly uses ed25519 signatures for server-to-server authentication (`pkg/network/federation/handshake.go`), but this protection does not extend to client-to-server connections.
- **Risk**: A network-position attacker (e.g., on a shared network segment, VPN exit node, or ISP) can MITM the DH exchange, establishing two independent encrypted sessions — one with the server and one with the client — while transparently decrypting and relaying all traffic including chat, voice, player state, and session tokens.
- **Closing the Gap**:
  1. **Preferred**: Wrap the TCP listener with TLS: generate an ed25519 keypair at startup (or load from `--tls-cert`/`--tls-key` flags), create a self-signed certificate, and use `tls.NewListener`. Clients verify the server fingerprint on first connection (TOFU model) and pin it for future connections.
  2. **Alternative**: Extend the federation ed25519 handshake pattern to client connections: have the server sign its DH public key, and ship the server's ed25519 public key fingerprint in the client config or as a flag.
  3. **Minimum**: Document clearly in README that the DH exchange does not prevent MITM on untrusted networks; advise operators to use a VPN or Tor for player-to-server connections until TLS is added.

---

## Gap 3 — Nonce Replay Prevention Has a Race Condition

- **Stated Goal**: Nonces prevent replay attacks on federation handshakes and player transfers (`pkg/network/federation/auth.go`, `pkg/security/doc.go`).
- **Current State**: `ValidateNonce` acquires only a read lock; `MarkNonceUsed` acquires a separate write lock. Callers must call both functions sequentially. Between the two calls, a concurrent request with the same nonce can pass `ValidateNonce` and also succeed — defeating replay prevention.
- **Risk**: Two federation requests carrying the same nonce can both be accepted if they arrive concurrently, allowing a replay of a player transfer or authentication event during the nonce TTL (5 minutes).
- **Closing the Gap**:
  1. Replace `ValidateNonce` + `MarkNonceUsed` with a single atomic `ConsumeNonce(nonce string) error` method that acquires a write lock for the entire validate-and-delete operation:
     ```go
     func (am *AuthManager) ConsumeNonce(nonce string) error {
         am.mu.Lock()
         defer am.mu.Unlock()
         timestamp, exists := am.nonces[nonce]
         if !exists {
             return fmt.Errorf("nonce not found or already used")
         }
         if time.Now().Unix()-timestamp > int64(am.nonceTTL.Seconds()) {
             delete(am.nonces, nonce)
             return fmt.Errorf("nonce expired")
         }
         delete(am.nonces, nonce)
         return nil
     }
     ```
  2. Update all callers to use `ConsumeNonce` instead of the two-step pattern.
  3. Add a test that launches two goroutines racing on the same nonce and verifies only one succeeds.

---

## Gap 4 — Session Token Validation Is Not Constant-Time

- **Stated Goal**: The security package provides `ConstantTimeCompare` to prevent timing attacks (`pkg/security/audit.go:1044–1053`, `pkg/security/doc.go:129`).
- **Current State**: `ValidateToken` performs a Go map lookup: `am.tokens[token]`. Go map lookups involve a hash computation followed by string equality via `==`. The equality comparison is not constant-time. The existing `security.ConstantTimeCompare` wrapping `crypto/subtle.ConstantTimeCompare` is defined but never called in the token validation path.
- **Risk**: A remote attacker who can make many requests and measure response latency may be able to distinguish token prefix matches from full misses, accelerating brute-force attacks. Practical exploitation is difficult given 128-bit token entropy and network jitter, but the existing infrastructure makes the fix trivial.
- **Closing the Gap**:
  1. Store tokens in the map indexed by `hex(HMAC-SHA256(serverSecret, token))` rather than the raw token string. Compute the HMAC index for every lookup, making map key comparison constant-time.
  2. Or: iterate the token map and use `security.ConstantTimeCompare` for the final match (acceptable if token count is small).
  3. The `serverSecret` should be generated with `crypto/rand` at startup and held only in memory.

---

## Gap 5 — No Per-IP Connection Rate Limiting on the TCP Acceptor

- **Stated Goal**: The server protects against DoS via rate limiting on application-layer operations (chat: `pkg/network/chat/system.go:61`, trade: `pkg/network/trade/system.go:137`).
- **Current State**: `acceptLoop` in `pkg/network/server.go:524` accepts all TCP connections without tracking or limiting connections per remote IP. The only gate is `MaxPlayers` (default 32). No TCP handshake timeout is set between `Accept()` and the point where the player slot is registered.
- **Risk**: A single attacker IP can open 32 connections and hold all player slots indefinitely (or until idle timeout), preventing legitimate players from joining. This is a trivial single-source DoS requiring no authentication bypass.
- **Closing the Gap**:
  1. Track active connection count per remote IP in a `sync.Map[string]int` keyed on `conn.RemoteAddr().(*net.TCPAddr).IP.String()`.
  2. Reject new connections from IPs already holding more than a configurable threshold (e.g., `--max-connections-per-ip`, default 3).
  3. Set a short authentication/handshake deadline (e.g., 10 seconds) after which a connection that hasn't sent a valid initial packet is dropped.
  4. Add a `--connection-rate-limit` flag (e.g., 10 new connections per IP per 60 seconds).

---

## Gap 6 — Federation Session Tokens Travel Over an Unencrypted Channel

- **Stated Goal**: Federation uses authenticated server identities (ed25519) and session tokens to authorize player transfers (`pkg/network/federation/auth.go`, `pkg/network/federation/handshake.go`).
- **Current State**: After the ed25519 handshake verifies server identity, subsequent messages — including the `PlayerTransfer` request that contains the `"token"` field — are written as plain JSON to the raw `net.Conn` (`pkg/network/federation/protocol.go:289`). No encryption is applied to the federation channel after the handshake.
- **Risk**: A network observer on a link between federated servers can capture session tokens (128-bit, 1-hour TTL) and use them to impersonate players on the target server within the token lifetime.
- **Closing the Gap**:
  1. After the ed25519 handshake completes, derive a shared AES-256-GCM key using the existing `DeriveAESKey` function from `pkg/network/crypto.go`, seeding it from ECDH or from a hash of both servers' public keys + a freshly exchanged nonce.
  2. Wrap the `net.Conn` in an encrypted stream before passing it to `executeTransferRequest`.
  3. Alternatively, wrap federation connections with TLS (mutual TLS using the ed25519 keys as client/server certificates in a TLS 1.3 handshake, which Go's `crypto/tls` supports via `tls.Certificate`).

---

## Gap 7 — Stub Profanity List Has No Production Enforcement

- **Stated Goal**: Chat messages are validated and sanitized before delivery; inappropriate content is blocked (`pkg/validation/chat.go`, `pkg/network/chat/system.go`).
- **Current State**: `NewChatValidator()` loads `buildProfanityList()` which returns a 3-word stub (`badword1`, `badword2`, `offensive`). This is the default constructor called by `pkg/network/chat/system.go:45`. The code is extensively documented as development-only, but no runtime mechanism enforces use of a real list in production.
- **Risk**: Server operators who do not read the documentation will deploy with effectively no profanity filtering, exposing players to abusive content. Players can also easily bypass the stub list since none of the stub words are actual slurs.
- **Closing the Gap**:
  1. Add a `--profanity-list` server flag that accepts a path to a word list file; pass it as `ChatValidatorConfig.CustomProfanityList`.
  2. At startup, log `logrus.Warn("using stub profanity list — configure --profanity-list for production")` when the default list is active.
  3. Consider adding a `--require-profanity-list` flag that causes the server to refuse to start without a real list, for operators who need content-moderated servers.

---

## Gap 8 — `ProfanityFilter.LoadWordListFromFile` Lacks Path Validation

- **Stated Goal**: The mod and file sandboxing philosophy (see `pkg/modding/sandbox.go`) restricts file access to designated directories.
- **Current State**: `LoadWordListFromFile(filepath string)` in `pkg/network/profanity.go:203` calls `os.Open(filepath)` with no path boundary check. If this API is ever exposed through a network-accessible configuration endpoint or mod parameter, a traversal path (`../../etc/passwd`) would allow arbitrary file read.
- **Risk**: Currently low (called from trusted operator configuration only), but a latent arbitrary-file-read vulnerability if the call site ever accepts user-supplied input.
- **Closing the Gap**:
  1. Apply the same pattern used in `pkg/modding/sandbox.go:100–117`:
     ```go
     absPath, _ := filepath.Abs(filepath)
     allowedDir, _ := filepath.Abs(expectedConfigDir)
     if !strings.HasPrefix(absPath, allowedDir+string(filepath.Sep)) {
         return fmt.Errorf("path outside allowed config directory")
     }
     ```
  2. Document the expected config directory in `LoadWordListFromFile`'s godoc.

---

## Gap 9 — Save and Settings Files Written World-Readable

- **Stated Goal**: The save system protects player data; settings are user-private (`pkg/saveload/`, `pkg/engine/settings.go`).
- **Current State**: Save files (`pkg/world/persistence.go:437`), housing files (`pkg/world/housing/persistence.go:43`), guild data (`pkg/world/economy/guild_bank.go:539`), settings (`pkg/engine/settings.go:187`, `pkg/rendering/ui/settings.go:637`), and keybind files (`pkg/rendering/ui/keybinds.go:420`) are all written with `os.WriteFile(path, data, 0o644)` or `os.Create` (which also defaults to `0o666` before umask). On multi-user Linux systems with a permissive umask, these files are readable by all local users.
- **Risk**: On shared systems, other local users can read server addresses, player configurations, world state, guild bank balances, and keybinding layouts.
- **Closing the Gap**:
  1. Use `0o600` for all user-specific files (settings, keybinds, save files, housing data).
  2. Use `os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)` instead of `os.WriteFile(path, data, 0o644)`.
  3. For newly created directories, use `0o700` instead of `0o755`.

---

## Gap 10 — No Authenticated Anti-Cheat Enforcement Path

- **Stated Goal**: The security package includes anti-cheat validation: stat bounds, inventory integrity, movement speed (`pkg/security/audit.go`).
- **Current State**: `RunFullAudit()` is called at server startup (when `--security-audit=true`) and logs results, but the audit is advisory: `pkg/security/audit.go` produces a `SecurityResults` struct with pass/fail counts, and the server continues regardless of failures (`cmd/server/main.go:runStartupValidations`). The anti-cheat checks (movement speed, stat limits) are implemented as audit checks but are not wired into the per-frame player state validation loop.
- **Risk**: Players can send crafted `InputCommand` packets with arbitrarily large velocity or stat values. While the server stamps the `EntityID` authoritatively, it does not validate that the numeric values in commands fall within game-legal bounds before applying them.
- **Closing the Gap**:
  1. Extract the bounds-checking logic from `pkg/security/audit.go` anti-cheat checks into a per-packet validation function called in `cmd/server/main.go:handleInputCommands` before commands are dispatched to the game world.
  2. Reject and log commands whose values exceed game-legal bounds (e.g., velocity > `maxSpeedForClass`), and increment a violation counter per player.
  3. Disconnect players who accumulate more than a configurable number of violations within a sliding time window.
