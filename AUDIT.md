# SECURITY AUDIT — 2026-04-21

## Project Security Profile

- **Deployment model**: Authoritative Go TCP server (`cmd/server`) + Go/Ebiten game client (`cmd/client`). Single-binary, zero-asset, cross-platform.
- **Trust boundaries**: Server is authoritative; clients are untrusted. Federation connects multiple servers peer-to-peer. Metrics/observability served on a separate HTTP port.
- **Auth model**: Connection-bound player ID (server-stamped, not client-supplied). Federation uses ed25519-signed handshakes + session tokens. Chat/trade use per-client rate limiting.
- **Data sensitivity**: Player entity state, voice audio (ADPCM), chat messages, save files, guild/economy data. No PII collection beyond player IDs.
- **Stated security goals** (from `pkg/security/doc.go` and `pkg/security/audit.go`): federation certificate validation, E2E chat encryption, mod sandbox, input validation, anti-cheat, privacy/data minimization.
- **Crypto in use**: DH-2048 (RFC 3526 Group 14) + AES-256-GCM for game traffic; ed25519 for federation server identity; SHA-256 for checksums and key derivation; `crypto/rand` for all security-sensitive randomness.
- **No SQL / no templating / no exec calls** found in production code.

---

## Security Surface Inventory

| Package | HTTP Handlers | DB Queries | Exec Calls | File I/O | Crypto | Auth |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|
| `pkg/observability` | 6 (unauthenticated) | — | — | — | — | **None** |
| `pkg/network` | — | — | — | — | DH+AES-GCM, SHA-256 | Player ID |
| `pkg/network/federation` | — | — | — | — | ed25519, SHA-256, crypto/rand | Token + nonce |
| `pkg/network/chat` | — | — | — | — | crypto/rand (msg IDs) | Rate limit |
| `pkg/network/trade` | — | — | — | — | — | Rate limit |
| `pkg/modding` | — | — | — | R/W `.json` files | — | Sandbox |
| `pkg/validation` | — | — | — | — | — | Rate limit |
| `pkg/security` | — | — | — | — | crypto/subtle | Audit only |
| `pkg/saveload` | — | — | — | R/W save files | SHA-256 checksum | — |
| `pkg/world/persistence` | — | — | — | R/W world files | — | — |
| `pkg/engine` | — | — | — | R/W settings/mods | — | — |
| `cmd/server` | — | — | — | — | — | Player ID |
| `cmd/client` | — | — | — | R/W settings | — | — |

---

## Dependency Vulnerability Check

Direct dependencies from `go.mod`:

| Dependency | Version | Known CVEs |
|-----------|---------|-----------|
| `github.com/hajimehoshi/ebiten/v2` | v2.9.3 | None known |
| `github.com/sirupsen/logrus` | v1.9.3 | None known |
| `github.com/google/uuid` | v1.6.0 | None known |
| `github.com/ncruces/zenity` | v0.10.14 | None known |
| `golang.org/x/image` | v0.32.0 | None known |
| `golang.org/x/text` | v0.30.0 | None known |
| `golang.org/x/sys` | v0.37.0 | None known |
| `github.com/stretchr/testify` | v1.11.1 | None known (test-only) |

No known CVEs found in any direct or indirect dependency as of the audit date. The dependency surface is small and focused. Note: `golang.org/x/text` versions below v0.3.8 had a vulnerability (CVE-2021-38561) — v0.30.0 is not affected.

---

## Findings

### HIGH

- [ ] **H1 — Observability HTTP Endpoints Have No Authentication** — `pkg/observability/metrics.go:170–175` — The metrics exporter registers six HTTP handlers (`/metrics`, `/health`, `/healthz`, `/ready`, `/readyz`, `/status`) on a port bound to all interfaces (default `:9090`) with zero authentication middleware. The `/status` handler at `metrics.go:263` returns a `statusResponse` JSON body containing `goroutines`, `heap_alloc_bytes`, `gc_runs`, FPS, frame time, memory usage MB, connected player count, active quest count, trade volume, entity count, and server uptime. Any network-reachable host can retrieve this data without credentials. The `/metrics` handler (`metrics.go:246`) emits Prometheus text format exposing the same fields. **Data flow**: `cmd/server/main.go:1270` → `observability.NewMetricsExporter(":" + *metricsPort)` → `m.server.ListenAndServe()` → unauthenticated mux. **Impact**: External reconnaissance of server internals; timing information useful for attack planning; confirmation of player activity. **Remediation**: Add a `Bearer`-token middleware or bind to `localhost` only by default; add a `--metrics-bind` flag that defaults to `127.0.0.1:9090`.

- [ ] **H2 — DH Key Exchange Is Susceptible to MITM (No Server Certificate)** — `pkg/network/server.go:221`, `pkg/network/crypto.go:40–60` — The game server creates a plain `net.Listen("tcp", address)` listener with no TLS wrapper. Application-layer encryption uses Diffie-Hellman (RFC 3526 Group 14, 2048-bit) + AES-256-GCM implemented in `pkg/network/crypto.go`. While message confidentiality is achieved after key agreement, the DH handshake itself occurs over an unauthenticated TCP channel: neither the server nor the client presents a certificate or performs signature verification of the DH public key. A network-position attacker can intercept the initial `PublicKey` exchange, substitute their own DH public key, and establish two independent encrypted sessions — one with the server and one with the client — transparently relaying and decrypting all traffic. The federation handshake (`pkg/network/federation/handshake.go`) does use ed25519 signatures for server-to-server authentication, but no equivalent protection exists for client-to-server connections. **Impact**: Full plaintext access to all game traffic, including voice (ADPCM), chat, player positions, and session tokens. **Remediation**: Wrap the TCP listener with TLS (`tls.NewListener`) using a self-signed or operator-provided certificate; or perform a signed DH exchange where the server's DH public key is signed with an ed25519 key whose public key is pinned on the client.

- [ ] **H3 — Nonce TOCTOU Race Enables Federation Replay Attacks** — `pkg/network/federation/auth.go:119–149` — `ValidateNonce` acquires a read lock (`am.mu.RLock`) to check nonce existence. `MarkNonceUsed` acquires a separate write lock (`am.mu.Lock`) to delete the nonce. These are two separate operations with no atomicity guarantee. If two goroutines concurrently call `ValidateNonce` with the same nonce before either calls `MarkNonceUsed`, both return success — the nonce is replayed. **Data flow**: Any caller that does `ValidateNonce(n)` → [context switch] → `MarkNonceUsed(n)` is vulnerable. **Impact**: Replayed federation handshakes; potential for duplicate player transfers; undermines the replay-prevention mechanism that nonces are intended to provide. **Remediation**: Combine validate-and-delete into a single method under a write lock:
  ```go
  func (am *AuthManager) ConsumeNonce(nonce string) error {
      am.mu.Lock()
      defer am.mu.Unlock()
      timestamp, exists := am.nonces[nonce]
      if !exists { return fmt.Errorf("nonce not found") }
      if time.Now().Unix()-timestamp > int64(am.nonceTTL.Seconds()) {
          delete(am.nonces, nonce)
          return fmt.Errorf("nonce expired")
      }
      delete(am.nonces, nonce)
      return nil
  }
  ```

---

### MEDIUM

- [ ] **M1 — Session Token Comparison Is Not Constant-Time** — `pkg/network/federation/auth.go:80` — `ValidateToken` looks up tokens via Go map access: `sessionToken, exists := am.tokens[token]`. Go map string key lookup performs hashing followed by equality comparison (`==`), which is not guaranteed to be constant-time — particularly the equality step after hash collision. A remote attacker making many requests can exploit timing variance to enumerate valid token prefixes. The project already implements `security.ConstantTimeCompare` (`pkg/security/audit.go:1048`) using `crypto/subtle.ConstantTimeCompare`, but it is never called in the token validation path. **Impact**: Token oracle that could accelerate brute-force of short-lived tokens. Low practical risk given 128-bit token entropy, but the existing `ConstantTimeCompare` infrastructure means the fix is trivial. **Remediation**: Store tokens as `HMAC-SHA256(secret, token)` as the map key, so lookup itself is constant-time; or iterate the token map with `subtle.ConstantTimeCompare` for final equality.

- [ ] **M2 — No Per-IP Connection Rate Limiting on TCP Acceptor** — `pkg/network/server.go:524–570` — `acceptLoop()` calls `s.listener.Accept()` in a tight loop with no per-IP throttling, SYN backlog tuning, or connection cost. The only gate is `MaxPlayers` (default 32). An attacker from a single IP can open 32 TCP connections, exhausting all player slots and preventing legitimate players from joining. The connections need not be authenticated — simply connecting and holding the TCP session is sufficient. The `validation.RateLimiter` (`pkg/validation/ratelimit.go`) exists but is wired only to chat (`pkg/network/chat/system.go:61`) and trade (`pkg/network/trade/system.go:137`), not to connection establishment. **Impact**: Single-source DoS that denies service to all legitimate players until the attacker's connections are evicted by the idle timeout (default indeterminate). **Remediation**: Track connections per remote IP; reject new connections from IPs that already hold more than N slots (e.g., 2–3 for home users). Add a TCP handshake timeout to evict non-talking connections quickly.

- [ ] **M3 — Federation Session Tokens Transmitted Over Potentially Unencrypted Channel** — `pkg/network/federation/protocol.go:266–289` — `executeTransferRequest` serializes the full transfer request (including `"token"` field) as JSON onto a raw `net.Conn` parameter: `json.NewEncoder(conn).Encode(transferReq)`. The federation handshake (`pkg/network/federation/handshake.go`) uses ed25519 for server identity verification but does not establish an encrypted session. The `net.Conn` passed to `executeTransferRequest` is the raw TCP connection from `net.Dial`. No TLS or AES-GCM wrapper is applied to federation connections after the handshake. **Data flow**: `cmd/server/v4_systems.go:266` → `protocol.TransferPlayer(...)` → `buildTransferRequest(playerID, token, ...)` → `json.NewEncoder(conn).Encode(...)` over plain TCP. **Impact**: A network observer on a peering link between federated servers can capture session tokens and impersonate players on the target server during the token's 1-hour TTL. **Remediation**: Apply TLS or the existing AES-256-GCM layer to federation connections after the ed25519 handshake completes.

- [ ] **M4 — Stub Profanity List Active in Production** — `pkg/validation/chat.go:196–205` — `NewChatValidator()` (the default constructor called at `pkg/network/chat/system.go:45`) loads a 3-word stub list: `["badword1", "badword2", "offensive"]`. This list provides effectively no content moderation. The code and its documentation acknowledge this explicitly, but no startup check, warning log at `Warn` level or above, nor configuration enforcement distinguishes a production deployment from a development one. **Impact**: Players can send any harmful or abusive language through the chat system without server-side filtering. **Remediation**: Log a `logrus.Warn` at startup if `len(profanityList) < threshold` (e.g., < 50 words); provide a `--profanity-list` server flag; consider refusing to start in production without an explicit override acknowledging the stub list.

- [ ] **M5 — `LoadWordListFromFile` Has No Path Boundary Check** — `pkg/network/profanity.go:203–204` — `LoadWordListFromFile(filepath string)` calls `os.Open(filepath)` directly with the caller-supplied path. No validation that the path resides within an expected configuration directory is performed. If this function is ever wired to a network-accessible config option (e.g., a server admin API or mod parameter), a directory traversal (`../../etc/passwd`) would allow arbitrary file read. Currently called only from trusted operator context, making this a LOW-MEDIUM latent risk. **Remediation**: Apply `filepath.Abs` + prefix check against an expected config directory before calling `os.Open`, mirroring the pattern in `pkg/modding/sandbox.go:100–117`.

---

### LOW

- [ ] **L1 — Internal Error Details Pushed to Error Channel** — `pkg/network/server.go:732–734` — `readMessageLength` pushes `fmt.Errorf("player %d message too large: %d bytes", client.playerID, msgLen)` onto the `s.errors` channel. If the consumer of this channel logs errors verbosely and those logs are accessible to players (e.g., via the unauthenticated `/status` endpoint or exposed log aggregation), internal message sizes and player IDs are disclosed. **Remediation**: Use structured error types that separate internal diagnostic fields from user-visible messages; avoid embedding raw byte counts in error strings pushed to observable channels.

- [ ] **L2 — Default World Seed is Hardcoded and Predictable** — `cmd/server/main.go:60` — The default `--seed` flag is `12345`. Every server launched without explicitly setting a seed generates an identical world. While seed sharing is intentional for reproducibility, operators may not realize their "default" server shares a layout with every other default deployment, potentially enabling map-based exploits if terrain is strategically valuable (resource locations, choke points). **Remediation**: Document this in the server startup help text; consider defaulting to `0` which triggers a `crypto/rand`-seeded world, with an explicit flag required to use a fixed seed.

- [ ] **L3 — `math/rand` Seeded from `time.Now().UnixNano()` in Client** — `cmd/client/util.go:185` — When no explicit seed is provided, the client generates one via `rand.New(rand.NewSource(nowNano)).Int63()` where `nowNano = time.Now().UnixNano()`. This is used only for world generation (non-security), but the pattern of `time`-seeded `math/rand` adjacent to security-sensitive code (tokens, nonces) warrants documentation to prevent accidental reuse. **Remediation**: Add a comment at the call site confirming this RNG is exclusively for world generation, not for security tokens.

- [ ] **L4 — Settings and Save Files Written with World-Readable Permissions** — `pkg/engine/settings.go:187`, `pkg/world/persistence.go:437`, `pkg/rendering/ui/settings.go:637` — Files are created with `0o644`, making them readable by all local users. On multi-user systems, settings files may contain server addresses, keybindings, and world state. **Remediation**: Use `0o600` for user-specific configuration and save files; use `0o640` if group access is required.

- [ ] **L5 — `security.ConstantTimeCompare` Logs at Debug Level** — `pkg/security/doc.go:129–140`, `pkg/security/audit.go:1048–1053` — The documentation acknowledges that `ConstantTimeCompare` adds a `logrus.Debug` log call. The comment correctly notes that `crypto/subtle.ConstantTimeCompare` handles the comparison before the log is reached, so the constant-time property is preserved. However, the debug log emitting comparison metadata could aid an attacker who has log access in confirming which comparison branch executed. **Remediation**: Remove the log statement from `ConstantTimeCompare` entirely, or restrict it to trace-level output that is never enabled in production builds.

---

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|-------------------|-----------------|
| `math/rand` in `pkg/engine/*`, `pkg/audio/*`, `pkg/rendering/*`, `pkg/procgen/*` | All uses are for deterministic procedural generation seeded from the world seed. Not security-relevant. |
| `math/rand` in `pkg/network/federation/retry.go` | Seeded from `cryptoRandSeed()` (uses `crypto/rand`, falls back to time only on `crypto/rand` failure). Used for retry jitter, not security tokens. |
| `math/rand` in `cmd/client/util.go` | Used for world seed generation only, not for cryptographic purposes. |
| `os.Getenv("LOG_LEVEL")` etc. | Reads only logging configuration; no secret values expected or used. |
| `filepath.Join` in `pkg/world/persistence.go`, `pkg/saveload/`, `pkg/engine/settings.go` | Paths are composed from server-controlled base directories + sanitized subcomponents. No user-supplied path components reach these calls directly. |
| `json:"-"` tagged fields (e.g., `pkg/world/persistence.go:27`) | Correctly excludes non-serializable types (mutexes, function values, runtime images) from JSON. Not a finding. |
| DH private key in `pkg/network/crypto.go` | Key is ephemeral (`crypto/rand`-generated per session), held only in memory, never written to disk. |
| `pkg/security/audit.go` running a self-audit at startup | The audit is informational and does not gate execution on results unless the operator enables the `--security-audit` flag. This is by design. |
| Token lookup via map in `ValidateToken` | Listed as M1 — not rejected, but severity assessed as MEDIUM given 128-bit entropy tokens make brute-force impractical; timing attack would require many requests measurable against network jitter. |
| Federation `generateToken()` (UUID v4) | Uses `crypto/rand` correctly; UUID format is applied after random byte generation. Entropy is 128 bits before version/variant nibble masking (122 bits effective). Adequate for session tokens. |
| Compression `LimitReader` in `pkg/network/compression.go:86` | Decompression bomb protection is correctly implemented with `io.LimitReader(reader, MaxDecompressedSize+1)`. Not a finding. |
| Modding sandbox `validateStringValue` pattern list | The mod system is JSON-only (no executable code). The pattern list (`eval(`, `<script`, etc.) is defense-in-depth; the real isolation is the absence of any script interpreter. |
| `os.MkdirAll(l.config.ModsDirectory, 0o755)` | Creates a server-controlled directory path, not a user-supplied path. |
