Goal: Add a single-command "host-and-play" mode that lets a player start an authoritative server and immediately join it as a client (ideal for LAN parties and casual local co-op). This is an implementation request — do NOT execute these instructions; only implement them in code.

Requirements and constraints:
- Single CLI flag (e.g., --host-and-play or --host-local) or subcommand that starts a server and then launches the client connected to the local server.
- Keep existing dedicated-server and client behaviors unchanged by default. This is an additive feature.
- Cross-platform: must work on Linux, macOS, and Windows. Avoid platform-specific syscalls where possible.
- Determinism: generation, seed derivation, and multiplayer synchronization must remain deterministic (no use of time.Now() for game state or global RNG).
- Security: by default bind the server to localhost; provide an explicit option to listen on LAN (0.0.0.0) with a clear warning. Do not open public internet ports without an explicit flag.
- Networking defaults: sensible default port (documented), auto-port-fallback if in use, and clear error messages if binding fails. Support optional UPnP/announcement only if explicitly enabled.
- Resource and lifecycle management: server must run in same process or a well-managed goroutine/subprocess with graceful shutdown when client exits or on user request. Ensure proper cleanup of sockets and threads.
- Compatibility: preserve existing save/load, network snapshot, and authoritative server mechanics. Client-side prediction and reconciliation must not be altered.

Behavior and UX:
- Command should start server, wait until listening, then start client and auto-connect to local server.
- Provide verbose logging and a minimal UI/CLI acknowledgement that server is ready and client is connected.
- Provide CLI flags to control bind address, port, max players, and whether to allow LAN connections.
- Document how to discover/join from other LAN machines (IP:port or LAN discovery if implemented).

Testing and validation:
- Add unit tests for the CLI parsing and server start/stop logic.
- Integration test(s) that start both server and client in a controlled environment (CI-friendly) and assert successful client connection and clean shutdown.
- Regression tests to confirm dedicated server start and normal client-only start unchanged.
- Manual test steps for a developer: start host-and-play on a host, join from a LAN client, verify deterministic generation with same seed.

Acceptance criteria:
1) Single CLI command successfully starts server and auto-connects local client.
2) Default behavior is secure (localhost bind) and no change to existing modes.
3) Tests added and passing locally; clear docs/CLI help updated ([README.md](http://_vscodecontentref_/0) and `cmd/*` help).
4) Graceful shutdown and resource cleanup verified.

Deliverables:
- Code implementing the flag/subcommand and lifecycle management.
- Unit + integration tests.
- Documentation update and CLI help text.
- Brief developer notes describing design decisions (bind defaults, threading model, security).