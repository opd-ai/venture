// Package hostplay provides in-process server lifecycle management for host-and-play mode.
//
// The hostplay package enables a single-command workflow where a player can start both
// a server and client in the same process, automatically connecting the client to the
// local server. This is ideal for LAN parties and casual local co-op sessions.
//
// # Security Model
//
// By default, the server binds to localhost (127.0.0.1) for security. To allow LAN
// connections, users must explicitly enable --host-lan flag, which binds to 0.0.0.0.
// No UPnP or public internet exposure occurs without explicit configuration.
//
// # Port Management
//
// The ServerManager attempts to bind to the requested port (default 8080). If that
// port is in use, it automatically tries ports 8081-8089 as fallbacks. This prevents
// conflicts with other local services.
//
// # Lifecycle Management
//
// The server runs in a dedicated goroutine within the client process. When the client
// exits (normally or via Ctrl+C), the ServerManager ensures graceful shutdown of the
// server, properly closing network listeners and cleaning up resources.
//
// # Cross-Platform Support
//
// The implementation uses only standard library networking primitives to ensure
// compatibility across Linux, macOS, and Windows. No platform-specific syscalls
// are required.
//
// # Usage Example
//
//	config := &hostplay.Config{
//		Port:       8080,
//		MaxPlayers: 4,
//		BindLAN:    false, // localhost only
//		WorldSeed:  12345,
//	}
//
//	manager, err := hostplay.NewServerManager(config, logger)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Start server and wait for ready
//	if err := manager.Start(); err != nil {
//		log.Fatal(err)
//	}
//
//	// Get connection details
//	addr := manager.Address()
//	fmt.Printf("Server ready at %s\n", addr)
//
//	// Ensure cleanup on exit
//	defer manager.Stop()
//
// # Design Decisions
//
//  1. Goroutine-based: Server runs in goroutine rather than subprocess for simpler
//     lifecycle management and lower overhead.
//
//  2. Localhost-first: Default bind to 127.0.0.1 follows security best practices.
//     LAN access requires explicit opt-in.
//
//  3. Port fallback: Automatic fallback to alternate ports improves UX when
//     default port is occupied.
//
//  4. Synchronous start: Start() blocks until server is listening or error occurs,
//     ensuring client can immediately connect.
//
//  5. Graceful shutdown: Stop() sends shutdown signal and waits for server goroutine
//     to complete cleanup before returning.
package hostplay
