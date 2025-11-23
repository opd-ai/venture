# Security Policy

**Venture - Fully Procedural Multiplayer Action-RPG**  
**Version:** 10.0  
**Last Updated:** December 2025

This document outlines the security model, threat mitigation strategies, and vulnerability reporting process for Venture.

---

## Table of Contents

1. [Security Model](#security-model)
2. [Threat Analysis](#threat-analysis)
3. [Mitigation Strategies](#mitigation-strategies)
4. [Reporting Vulnerabilities](#reporting-vulnerabilities)
5. [Security Best Practices](#security-best-practices)

---

## Security Model

### Trust Boundaries

**Client-Server Architecture:**
- **Server:** Authoritative for all game state (trusted)
- **Client:** Untrusted input source (validated by server)
- **Network:** Potentially hostile (encryption required)

**Federation Trust Levels:**
1. **Trusted Servers:** Known servers with reputation >0.8 (full feature access)
2. **Verified Servers:** Certificate exchange complete (limited features)
3. **Unknown Servers:** No prior interaction (read-only federation)

### Authentication & Authorization

**Player Authentication:**
- **Session Tokens:** Ed25519-signed tokens, 24-hour expiration
- **Nonce-Based Replay Prevention:** Each request includes unique nonce
- **Password Hashing:** Argon2id (memory-hard, GPU-resistant)

**Server Authentication:**
- **Ed25519 Certificates:** Self-signed, exchanged during federation handshake
- **Signature Verification:** All server-to-server messages cryptographically signed
- **Trust Score:** 0.0-1.0 based on player reports and detected anomalies

---

## Threat Analysis

### Client-Side Threats

**T1: Input Injection**
- **Attack:** Malicious input via keyboard/mouse/chat (SQL injection, XSS equivalent)
- **Impact:** Code execution, data corruption, denial of service
- **Likelihood:** Medium (requires exploitable parsing bug)
- **Mitigation:** Input sanitization, bounds checking, no eval() or similar constructs

**T2: Memory Manipulation (Cheating)**
- **Attack:** External tools modify game memory (health, gold, XP)
- **Impact:** Unfair advantage in multiplayer, economy disruption
- **Likelihood:** High (common in games)
- **Mitigation:** Server-authoritative state, client predictions reconciled, stat sanity checks

**T3: Packet Sniffing / MITM**
- **Attack:** Network traffic interception, credential theft
- **Impact:** Account compromise, message eavesdropping
- **Likelihood:** Low (encrypted traffic)
- **Mitigation:** E2E encryption (AES-256-GCM), TLS 1.3 for server-to-server

### Server-Side Threats

**T4: Denial of Service (DoS)**
- **Attack:** Flood server with requests, exhaust resources
- **Impact:** Server crash, unavailability for legitimate players
- **Likelihood:** Medium (multiplayer servers are targets)
- **Mitigation:** Rate limiting (10 msg/s per player), connection throttling, IP bans

**T5: Unauthorized Access**
- **Attack:** Exploit authentication bypass, privilege escalation
- **Impact:** Admin access, world state corruption
- **Likelihood:** Low (no default admin accounts, strong auth)
- **Mitigation:** Least privilege principle, audit logs, no debug endpoints in production

**T6: State Desync / Duplication**
- **Attack:** Exploit race conditions to duplicate items, XP
- **Impact:** Economy inflation, unfair advantages
- **Likelihood:** Medium (distributed systems are complex)
- **Mitigation:** Two-phase commits, optimistic locking, rollback on conflict

### Federation Threats

**T7: Rogue Server**
- **Attack:** Malicious federated server sends corrupted data
- **Impact:** Player state corruption, item theft
- **Likelihood:** Medium (federation requires trust)
- **Mitigation:** Trust score system, player state validation on transfer, rollback on anomaly

**T8: Server Impersonation**
- **Attack:** Fake server claims to be trusted server
- **Impact:** Player credentials stolen, state corruption
- **Likelihood:** Low (cryptographic verification)
- **Mitigation:** Ed25519 certificate pinning, signature verification

**T9: Eclipse Attack (P2P)**
- **Attack:** Isolate server from network, feed false peer information
- **Impact:** Denial of service, split-brain scenarios
- **Likelihood:** Low (requires network-level control)
- **Mitigation:** Multiple discovery mechanisms (LAN broadcast + manual peers + gossip)

### Modding Threats

**T10: Malicious Mods**
- **Attack:** Mod exploits server vulnerabilities, executes arbitrary code
- **Impact:** Server compromise, data theft
- **Likelihood:** Medium (mods are user-generated)
- **Mitigation:** Lua sandbox (no file I/O, no network access, CPU limits), mod signing

---

## Mitigation Strategies

### Input Validation

**All Inputs Validated:**
```go
// Example: Chat message validation
func validateChatMessage(msg string) error {
    if len(msg) > 500 { return errors.New("message too long") }
    if containsNullBytes(msg) { return errors.New("invalid characters") }
    if len(strings.TrimSpace(msg)) == 0 { return errors.New("empty message") }
    return nil
}
```

**Bounds Checking:**
- All array/slice access: `if idx < 0 || idx >= len(arr) { return error }`
- All numeric inputs: `if value < MIN || value > MAX { return error }`

**Sanitization:**
- Chat messages: Strip control characters, limit length
- Player names: Alphanumeric + spaces only, 3-20 characters
- File paths: Reject `../` and absolute paths

### Rate Limiting

**Per-Player Limits:**
- Chat: 1 msg/3s (global), 1 msg/1s (local), 1 msg/0.5s (party)
- Images: 1 upload/60s
- Trades: 5 proposals/minute
- Server transfers: 10/minute (per server)

**Per-IP Limits:**
- Connection attempts: 10/minute
- Failed logins: 5/10 minutes (temporary ban)

**Enforcement:**
- Token bucket algorithm (configurable burst)
- Exponential backoff for violations (30s → 60s → 120s → 600s max)

### Encryption

**End-to-End (Player-to-Player):**
- **Key Exchange:** Diffie-Hellman 2048-bit (RFC 3526 Group 14)
- **Symmetric:** AES-256-GCM (random IV per message)
- **Integrity:** HMAC-SHA256

**Transport (Server-to-Server):**
- **TLS 1.3:** Modern ciphers only (ECDHE-RSA-AES256-GCM-SHA384)
- **Certificate Validation:** Mutual TLS (both server and client authenticate)

**At-Rest (Save Files):**
- **Optional:** User-provided password → Argon2id → AES-256-CBC
- **Default:** Plaintext JSON (local files, user controls physical security)

### Access Control

**Principle of Least Privilege:**
- No default admin accounts
- Server operators have read-only access to logs, cannot modify world state without player consent
- Mods run in sandboxed environment (no file system, no network access except allowed APIs)

**Capability-Based Security:**
- Each player session has explicit capabilities (can_trade, can_chat, can_use_portal)
- Capabilities granted by server based on level, reputation, server policy

### Audit Logging

**All Security-Relevant Events Logged:**
- Authentication: Login success/failure, session creation/expiration
- Authorization: Permission denials, capability checks
- Federation: Server connection/disconnection, trust score changes
- Moderation: Player reports, admin actions, bans

**Log Format:**
```
[2025-12-25 14:30:22] [SECURITY] [player:alice] [action:login_success] [ip:192.168.1.100]
[2025-12-25 14:35:10] [SECURITY] [player:bob] [action:trade_rejected] [reason:trust_too_low]
```

**Log Retention:**
- 30 days for all events (rotating log files)
- Critical events (bans, admin actions) archived indefinitely

---

## Reporting Vulnerabilities

### Responsible Disclosure

**DO:**
- Email security@venture-rpg.com with details
- Allow 90 days for fix before public disclosure
- Provide proof-of-concept (code or steps to reproduce)

**DON'T:**
- Exploit vulnerabilities on public servers
- Share vulnerability details publicly before disclosure deadline
- Demand payment (no bug bounty program currently)

### What to Include

1. **Description:** Clear explanation of vulnerability
2. **Impact:** What attacker can achieve (data theft, DoS, etc.)
3. **Reproduction Steps:** Exact commands, inputs, or code to trigger
4. **Affected Versions:** Specify versions tested (e.g., "v10.0, v9.0")
5. **Suggested Fix:** Optional (appreciated, not required)

### Response Timeline

- **Acknowledgment:** Within 48 hours
- **Triage:** Within 7 days (severity classification)
- **Fix:** 30-90 days depending on severity
- **Disclosure:** After fix deployed and tested

### Severity Classification

| Severity | Impact | Response Time |
|----------|--------|---------------|
| **Critical** | Remote code execution, data breach | 7 days |
| **High** | Authentication bypass, DoS | 30 days |
| **Medium** | Information disclosure, XSS | 60 days |
| **Low** | Minor leaks, cosmetic issues | 90 days |

---

## Security Best Practices

### For Server Operators

**Essential:**
1. **Update Regularly:** Apply security patches within 7 days
2. **Firewall:** Only expose ports 8080 (game) and 80/443 (WebRTC signaling)
3. **Strong Passwords:** 16+ characters, random, unique per server
4. **Backups:** Daily backups to separate storage (offline or cloud)
5. **Monitor Logs:** Check `logs/venture-server.log` daily for anomalies

**Recommended:**
6. **Fail2Ban:** Auto-ban IPs after 5 failed login attempts
7. **HTTPS:** Use Let's Encrypt for server web UI (if exposing admin panel)
8. **Non-Root:** Run server as dedicated user (not `root` or `Administrator`)
9. **SELinux/AppArmor:** Enable on Linux for additional sandboxing
10. **Rate Limit:** Reduce default limits for public servers (`-rate-limit-global 0.5`)

**Advanced:**
11. **DDoS Protection:** Cloudflare or similar service for public IPs
12. **Honeypot:** Dummy server to detect scanning/exploitation attempts
13. **Security Audits:** Third-party penetration testing (contact security@venture-rpg.com for referrals)

### For Players

**Account Security:**
1. **Unique Password:** Never reuse passwords from other services
2. **2FA (Planned v10.1):** Enable two-factor authentication when available
3. **Logout:** Always logout on shared computers

**Network Security:**
4. **VPN (Optional):** Use VPN to hide IP address from server operators
5. **Tor (Optional):** Venture supports onion services for anonymity
6. **Verify Server Certificates:** Check server fingerprint before first connection

**Data Privacy:**
7. **Read Privacy Policy:** Understand what data servers collect
8. **Disable Telemetry:** Settings → Privacy → Telemetry: Off (default: opt-in only)
9. **Encrypt Saves:** Settings → Saves → Encrypt Save Files: Enable

### For Mod Developers

**Sandbox Limitations (Enforced):**
- No file system access (read or write)
- No network access (except whitelisted APIs)
- CPU time limit: 100ms per tick (prevents infinite loops)
- Memory limit: 50MB per mod

**Safe Practices:**
1. **Validate All Inputs:** Don't trust data from server or players
2. **No Secrets:** Mods are plaintext, never embed API keys or passwords
3. **Test in Sandbox:** Use `-mod-debug` flag to test mods in isolated environment
4. **Code Review:** Request community review before publishing

---

## Security Updates

### Supported Versions

| Version | Security Updates | End of Life |
|---------|------------------|-------------|
| **10.x** | Yes | TBD |
| **9.x** | Yes | June 2026 |
| **8.x** | Yes | December 2026 |
| **7.x** | Critical Only | December 2025 |
| **≤6.x** | No | Already EOL |

**Recommendation:** Always use latest major version (10.x) for best security.

### Notification Channels

- **GitHub Security Advisories:** [Security Tab](https://github.com/opd-ai/venture/security/advisories)
- **Discord:** #security-announcements (auto-posted for all advisories)
- **Email:** Opt-in mailing list (security-alerts@venture-rpg.com)

---

## Contact

**Security Team Email:** security@venture-rpg.com  
**PGP Key:** [Download Public Key](https://venture-rpg.com/security/pgp-key.asc)  
**Response Time:** Within 48 hours for all inquiries

**General Support:** support@venture-rpg.com (non-security issues)

---

**Acknowledgments:**

We thank the following researchers for responsible disclosure:

*(No public CVEs to date - if you find one, you'll be listed here!)*

---

**Version:** 10.0 (December 2025)  
**Maintained By:** Venture Security Team  
**Next Review:** June 2026
