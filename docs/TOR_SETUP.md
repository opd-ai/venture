# Tor Setup Guide for Venture Multiplayer

This guide provides comprehensive instructions for running Venture multiplayer over the Tor network, including both server (hidden service) and client (SOCKS5 proxy) configurations.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Server Setup (Hidden Service)](#server-setup-hidden-service)
- [Client Setup (SOCKS5 Proxy)](#client-setup-socks5-proxy)
- [Testing and Validation](#testing-and-validation)
- [Troubleshooting](#troubleshooting)
- [Security Considerations](#security-considerations)
- [Performance Expectations](#performance-expectations)
- [Advanced Configuration](#advanced-configuration)

## Overview

Venture supports high-latency multiplayer networking designed for Tor/onion services with latencies up to 5000ms. This enables:

- **Anonymous multiplayer gaming** over Tor network
- **Censorship-resistant connectivity** in restricted regions
- **Privacy-preserving co-op gameplay** without revealing IP addresses
- **Decentralized server hosting** via onion services

**Key Features:**
- Automatic timeout adjustment for high-latency connections (60s read, 30s write)
- TCP keepalive to prevent NAT/proxy disconnections (30s period)
- Increased buffer sizes (512 messages) for latency spike tolerance
- Automatic reconnection with exponential backoff (up to 10 retries)
- Extended lag compensation (15s history) and prediction buffers (12.8s history)

## Prerequisites

### System Requirements

**Both Server and Client:**
- Venture server/client binaries (built with Go 1.24+)
- Tor installed and running (version 0.4.5+ recommended)
- Linux, macOS, or Windows (Tor Browser Bundle on Windows)

### Installing Tor

**Debian/Ubuntu:**
```bash
sudo apt update
sudo apt install tor
sudo systemctl enable tor
sudo systemctl start tor
```

**macOS (Homebrew):**
```bash
brew install tor
brew services start tor
```

**Arch Linux:**
```bash
sudo pacman -S tor
sudo systemctl enable tor
sudo systemctl start tor
```

**Windows:**
Download and install [Tor Expert Bundle](https://www.torproject.org/download/tor/) or use [Tor Browser](https://www.torproject.org/download/).

### Verifying Tor Installation

```bash
# Check Tor is running
sudo systemctl status tor  # Linux
# or
ps aux | grep tor  # macOS/generic

# Test SOCKS5 proxy (should return Tor exit node IP)
curl --socks5-hostname localhost:9050 https://check.torproject.org/api/ip
```

Expected output:
```json
{"IsTor":true,"IP":"<tor_exit_ip>"}
```

## Server Setup (Hidden Service)

Running a Venture server as a Tor hidden service (onion service) makes it accessible anonymously without revealing the server's IP address.

### Step 1: Configure Tor Hidden Service

Edit the Tor configuration file:

**Linux/macOS:**
```bash
sudo nano /etc/tor/torrc
```

**Windows:**
Edit `C:\Users\<YourUser>\Desktop\Tor Browser\Browser\TorBrowser\Data\Tor\torrc`

Add the following configuration at the end of the file:

```
# Venture Hidden Service
HiddenServiceDir /var/lib/tor/venture/
HiddenServicePort 8080 127.0.0.1:8080
```

**Configuration Explanation:**
- `HiddenServiceDir`: Directory where Tor stores hidden service keys and hostname
- `HiddenServicePort`: Maps public onion port (8080) to local server port (127.0.0.1:8080)

**Windows Path Adjustment:**
```
HiddenServiceDir C:\Users\<YourUser>\AppData\Roaming\tor\venture\
```

### Step 2: Restart Tor

**Linux:**
```bash
sudo systemctl restart tor
```

**macOS:**
```bash
brew services restart tor
```

**Windows:**
Restart Tor Browser or Tor service.

### Step 3: Retrieve Onion Address

**Linux/macOS:**
```bash
sudo cat /var/lib/tor/venture/hostname
```

**Windows:**
```cmd
type C:\Users\<YourUser>\AppData\Roaming\tor\venture\hostname
```

**Expected Output:**
```
abcdefghijklmnop.onion
```

This is your server's onion address. Share this with clients (keep the private key secret).

**Important:** The first time Tor generates the onion address, it may take 30-60 seconds. Subsequent restarts are faster.

### Step 4: Start Venture Server

Start the Venture server with high-latency configuration:

```bash
./venture-server -high-latency -port 8080 -max-players 4
```

**Server Configuration Details:**
- `-high-latency`: Enables 60s read timeout, 30s write timeout, 512-message buffers
- `-port 8080`: Must match the local port in Tor's HiddenServicePort configuration
- `-max-players 4`: Adjust based on bandwidth capacity

**Important:** The server binds to `127.0.0.1:8080` (localhost only). Tor forwards external onion connections to this local port. Do NOT use `0.0.0.0` as it would expose the server on your local network.

### Step 5: Verify Hidden Service

Test that your hidden service is accessible:

```bash
# From the same machine
curl --socks5-hostname localhost:9050 http://$(sudo cat /var/lib/tor/venture/hostname):8080
```

If the server is running, you should see a connection attempt in the server logs.

### Advanced Server Configuration

**Custom Tor SOCKS Port:**
If Tor runs on a non-default port, specify in client configuration (see Client Setup).

**Multiple Hidden Services:**
To run multiple Venture servers on different ports:

```
# Server 1 - Port 8080
HiddenServiceDir /var/lib/tor/venture1/
HiddenServicePort 8080 127.0.0.1:8080

# Server 2 - Port 8081
HiddenServiceDir /var/lib/tor/venture2/
HiddenServicePort 8081 127.0.0.1:8081
```

Each gets a unique onion address.

**Persistence:**
The onion address is deterministic based on the private key in `HiddenServiceDir`. Backup this directory to preserve your onion address across reinstalls:

```bash
sudo tar -czf venture-onion-backup.tar.gz /var/lib/tor/venture/
```

## Client Setup (SOCKS5 Proxy)

Clients connect to onion services through Tor's SOCKS5 proxy.

### Step 1: Ensure Tor is Running

```bash
# Verify Tor SOCKS5 proxy is listening
netstat -an | grep 9050
# or
lsof -i :9050
```

Expected: Tor listening on `127.0.0.1:9050`

### Step 2: Configure Venture Client

Venture uses Go's standard `net.Dial` which supports SOCKS5 proxies via environment variables or custom dialer.

**Option A: Environment Variable (Simple)**

Set the `ALL_PROXY` environment variable before running the client:

```bash
export ALL_PROXY=socks5://127.0.0.1:9050
./venture-client -multiplayer -server abcdefghijklmnop.onion:8080
```

**Option B: Custom Dialer (Recommended for Production)**

Modify `pkg/network/client.go` to use a SOCKS5 dialer:

```go
import (
    "golang.org/x/net/proxy"
    "net"
)

// In TCPClient.Connect() method, replace:
// conn, err := net.DialTimeout("tcp", c.config.ServerAddress, c.config.ConnectionTimeout)

// With:
dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
if err != nil {
    return fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
}

conn, err := dialer.Dial("tcp", c.config.ServerAddress)
if err != nil {
    return fmt.Errorf("failed to connect to server: %w", err)
}
```

**Add dependency:**
```bash
go get golang.org/x/net/proxy
```

**Option C: Tor-Specific Client Configuration**

Use the `TorClientConfig()` preset designed for high-latency Tor connections:

```go
// In client application code
config := network.TorClientConfig()
config.ServerAddress = "abcdefghijklmnop.onion:8080"

client := network.NewTCPClient(config)
err := client.Connect()
```

This automatically configures:
- 60-second connection timeout (for circuit building)
- 5000ms max latency tolerance
- 512-message buffer size
- 5-second ping interval
- Automatic reconnection with Tor-optimized backoff

### Step 3: Connect to Server

```bash
# Using environment variable
export ALL_PROXY=socks5://127.0.0.1:9050
./venture-client -multiplayer -server abcdefghijklmnop.onion:8080

# Using Tor Browser's SOCKS port (9150)
export ALL_PROXY=socks5://127.0.0.1:9150
./venture-client -multiplayer -server abcdefghijklmnop.onion:8080
```

**First Connection Notes:**
- Initial connection may take 30-60 seconds as Tor builds circuits
- Subsequent connections are faster (circuit reuse)
- Client will show "Connecting..." during circuit building
- Once connected, you'll see "Connected to server" message

### Troubleshooting Connection Issues

**Connection Timeout:**
```
Error: connection timeout
```

**Solutions:**
1. Verify Tor is running: `systemctl status tor`
2. Test SOCKS proxy: `curl --socks5-hostname localhost:9050 https://check.torproject.org/api/ip`
3. Increase connection timeout in client config (already 60s with `-high-latency`)
4. Check server is running and accessible via onion address
5. Verify onion address is correct (16-character v2 or 56-character v3)

**SOCKS Proxy Error:**
```
Error: SOCKS5 proxy error
```

**Solutions:**
1. Verify SOCKS port (default 9050, Tor Browser uses 9150)
2. Check Tor configuration: `cat /etc/tor/torrc | grep SocksPort`
3. Ensure no firewall blocking port 9050/9150
4. Try `ALL_PROXY=socks5h://127.0.0.1:9050` (forces DNS through proxy)

## Testing and Validation

### Local Testing (Same Machine)

Test server and client on the same machine before deploying remotely:

**Terminal 1 - Start Server:**
```bash
./venture-server -high-latency -port 8080
```

**Terminal 2 - Start Client:**
```bash
export ALL_PROXY=socks5://127.0.0.1:9050
./venture-client -multiplayer -server $(sudo cat /var/lib/tor/venture/hostname):8080
```

### Latency Simulation

Simulate high latency to verify configuration tolerance:

**Linux (using `tc` - traffic control):**
```bash
# Add 2500ms latency (5000ms round-trip)
sudo tc qdisc add dev lo root netem delay 2500ms

# Test connection
export ALL_PROXY=socks5://127.0.0.1:9050
./venture-client -multiplayer -server localhost:8080

# Remove latency simulation
sudo tc qdisc del dev lo root
```

**Expected Behavior:**
- Connection establishes within 60 seconds
- No timeout disconnections during gameplay
- Smooth entity interpolation despite latency
- Automatic reconnection if connection drops

### Remote Testing

Test with a friend over actual Tor network:

1. **Server operator** shares onion address (from `/var/lib/tor/venture/hostname`)
2. **Client** connects using onion address:
   ```bash
   export ALL_PROXY=socks5://127.0.0.1:9050
   ./venture-client -multiplayer -server <onion_address>:8080
   ```
3. Verify gameplay is responsive and stable for 10+ minutes

### Network Diagnostics

Monitor connection quality:

**Server-side:**
```bash
# Watch server logs for warnings
tail -f venture-server.log | grep -i "timeout\|disconnect\|error"

# Monitor buffer utilization (should stay <80%)
# Check server output for "buffer utilization" warnings
```

**Client-side:**
```bash
# Watch client logs
tail -f venture-client.log | grep -i "latency\|timeout\|reconnect"

# Measure actual latency to onion service
torsocks ping <onion_address>  # Note: many onion services don't respond to ping
```

**Tor Circuit Information:**
```bash
# Monitor Tor logs for circuit issues
sudo journalctl -u tor -f | grep -i "circuit\|failed\|timeout"
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Server Not Accessible via Onion Address

**Symptoms:**
- Client times out connecting to onion address
- `curl` test fails with "SOCKS5 request failed"

**Diagnosis:**
```bash
# Check hidden service directory permissions
sudo ls -la /var/lib/tor/venture/

# Verify hostname file exists
sudo cat /var/lib/tor/venture/hostname

# Check Tor logs for errors
sudo journalctl -u tor | grep -i "venture\|error"
```

**Solutions:**
1. Ensure Tor has write permissions to HiddenServiceDir:
   ```bash
   sudo chown -R debian-tor:debian-tor /var/lib/tor/venture/
   sudo chmod 700 /var/lib/tor/venture/
   ```
2. Verify `torrc` configuration is correct (no typos)
3. Restart Tor: `sudo systemctl restart tor`
4. Check Venture server is running on correct port: `netstat -an | grep 8080`

#### 2. Client Connection Hangs During Circuit Building

**Symptoms:**
- Client shows "Connecting..." for >90 seconds
- Eventually times out with connection error

**Diagnosis:**
```bash
# Check Tor connectivity
curl --socks5-hostname localhost:9050 https://check.torproject.org/api/ip

# Monitor Tor circuit building
sudo journalctl -u tor -f
```

**Solutions:**
1. Verify Tor is connected to network:
   ```bash
   sudo journalctl -u tor | grep "Bootstrapped 100%"
   ```
2. Check firewall isn't blocking Tor:
   ```bash
   sudo ufw status  # Linux
   # or
   sudo iptables -L  # Linux
   ```
3. Try using Tor bridges if Tor is blocked:
   ```
   # Add to /etc/tor/torrc
   UseBridges 1
   Bridge obfs4 <bridge_line>  # Get from https://bridges.torproject.org
   ```
4. Increase client connection timeout (already 60s with TorClientConfig)

#### 3. Frequent Disconnections During Gameplay

**Symptoms:**
- Client disconnects every 5-15 minutes
- Server logs show "read timeout" or "write timeout"

**Diagnosis:**
```bash
# Check server buffer utilization
# Look for "buffer utilization 80%" warnings in server logs

# Monitor network latency
# Client logs should show "latency: XXXXms" messages
```

**Solutions:**
1. Verify high-latency mode is enabled on server:
   ```bash
   ./venture-server -high-latency -port 8080
   ```
2. Check for network congestion:
   ```bash
   # Monitor Tor bandwidth
   sudo journalctl -u tor | grep "bandwidth"
   ```
3. Reduce update rate on server (trade-off: less smooth gameplay):
   ```go
   // In server config
   config.UpdateRate = 10  // 10 Hz instead of 20 Hz (50% bandwidth reduction)
   ```
4. Enable buffer monitoring and check for drops:
   ```go
   stats := server.GetBufferStats()
   // Check stats["state_updates"].Dropped count
   ```

#### 4. High Latency Warnings

**Symptoms:**
- Client logs show "high latency detected: XXXXms"
- Gameplay feels laggy despite connection stability

**Diagnosis:**
This is expected for Tor connections. Typical Tor latency: 500-3000ms.

**Solutions:**
1. Adjust client MaxLatency threshold to suppress warnings:
   ```go
   config := network.TorClientConfig()
   config.MaxLatency = 10000 * time.Millisecond  // 10s tolerance
   ```
2. Optimize Tor circuit path:
   ```
   # Add to /etc/tor/torrc (server and client)
   ExitNodes {us},{ca},{gb}  # Prefer fast exit nodes
   StrictNodes 0  # Allow fallback if preferred nodes unavailable
   ```
3. Use Tor bridges geographically closer to server
4. Consider running server and client on same continent

#### 5. "Too Many Hops" or Circuit Failure

**Symptoms:**
- Tor logs show "circuit failed" or "no route to host"
- Cannot establish any onion connections

**Diagnosis:**
```bash
# Check Tor guard nodes and circuits
sudo journalctl -u tor | grep -i "circuit\|guard"
```

**Solutions:**
1. Delete Tor state to force new circuit selection:
   ```bash
   sudo systemctl stop tor
   sudo rm /var/lib/tor/state
   sudo systemctl start tor
   ```
2. Try connecting at different times (network congestion)
3. Use newer Tor version (0.4.7+ recommended):
   ```bash
   sudo apt update
   sudo apt upgrade tor
   ```

#### 6. Onion Address Changed After Tor Restart

**Symptoms:**
- Server onion address is different after reboot
- Clients cannot connect with old address

**Diagnosis:**
```bash
# Check if hostname file exists
sudo cat /var/lib/tor/venture/hostname

# Verify private key exists
sudo ls -la /var/lib/tor/venture/
```

**Solutions:**
1. Restore from backup:
   ```bash
   sudo tar -xzf venture-onion-backup.tar.gz -C /
   sudo systemctl restart tor
   ```
2. If no backup, the old address is permanently lost
3. Share new address with clients
4. Set up regular backups:
   ```bash
   # Add to crontab
   0 0 * * * sudo tar -czf /backup/venture-onion-$(date +\%Y\%m\%d).tar.gz /var/lib/tor/venture/
   ```

## Security Considerations

### Server Security

**1. Never Expose Real IP Address:**
- Always bind server to `127.0.0.1` (localhost), never `0.0.0.0`
- Verify no port forwarding rules expose port 8080 externally
- Use firewall to block external access:
  ```bash
  sudo ufw deny 8080  # Block external access
  sudo ufw status
  ```

**2. Protect Hidden Service Private Key:**
- The `hs_ed25519_secret_key` file in HiddenServiceDir must remain secret
- Compromise allows attacker to impersonate your onion service
- Set strict permissions:
  ```bash
  sudo chmod 700 /var/lib/tor/venture/
  sudo chmod 600 /var/lib/tor/venture/hs_ed25519_secret_key
  ```
- Back up to encrypted storage only:
  ```bash
  sudo tar -czf venture-onion.tar.gz /var/lib/tor/venture/
  gpg -c venture-onion.tar.gz  # Encrypt with password
  rm venture-onion.tar.gz  # Delete unencrypted backup
  ```

**3. Monitor for Abuse:**
- Enable server logging to detect unusual activity
- Set reasonable rate limits and max players
- Consider implementing IP ban lists (though Tor makes this challenging)

**4. Use Onion Service Authentication (Advanced):**
Tor supports client authentication to restrict who can access your hidden service:

```
# In /etc/tor/torrc
HiddenServiceDir /var/lib/tor/venture/
HiddenServicePort 8080 127.0.0.1:8080
HiddenServiceAuthorizeClient stealth client1,client2
```

Tor generates auth credentials to share with authorized clients.

### Client Security

**1. Verify Onion Address:**
- Always verify onion address through trusted channel (encrypted messaging, in-person)
- Beware of typosquatting (similar-looking addresses)
- V3 onion addresses (56 characters) are more secure than V2 (16 characters)

**2. Use Tor Browser for Initial Testing:**
- Test server accessibility via Tor Browser first
- Ensures Tor configuration is correct before using game client

**3. Avoid Leaking Traffic:**
- Ensure ALL_PROXY is set or use custom SOCKS5 dialer
- Test with `curl --socks5-hostname localhost:9050` to verify proxy is working
- Do NOT use clearnet fallback if onion connection fails

**4. Keep Tor Updated:**
- Regularly update Tor to patch security vulnerabilities:
  ```bash
  sudo apt update && sudo apt upgrade tor
  ```
- Subscribe to Tor security announcements: https://blog.torproject.org

**5. Consider VPN + Tor (Advanced):**
- Using VPN before Tor hides Tor usage from ISP
- Configuration: `You -> VPN -> Tor -> Onion Service`
- Note: Adds latency, may exceed 5000ms tolerance

### General Security Best Practices

**1. Operational Security (OpSec):**
- Don't correlate onion address with real identity on social media
- Use separate Tor identity for server operation vs personal browsing
- Avoid posting onion address on clearnet forums linked to real identity

**2. Traffic Analysis Resistance:**
- High-latency mode includes some padding via larger buffers
- For maximum security, consider running server on dedicated hardware
- Avoid distinctive traffic patterns (e.g., always playing at same time)

**3. Logging:**
- Be aware Venture server logs player activity
- Store logs securely or disable detailed logging:
  ```go
  // In server configuration
  logrus.SetLevel(logrus.WarnLevel)  // Only log warnings and errors
  ```
- Rotate and delete old logs regularly:
  ```bash
  # Add to crontab
  0 0 * * 0 find /var/log/venture/ -name "*.log" -mtime +7 -delete
  ```

**4. Denial of Service (DoS) Mitigation:**
- Onion services are vulnerable to DoS attacks
- Set conservative max players limit (4-8 recommended)
- Monitor CPU and bandwidth usage
- Consider using onion service PoW defense (Tor 0.4.6+):
  ```
  # In /etc/tor/torrc
  HiddenServicePoWDefense 1
  ```

## Performance Expectations

### Latency

**Typical Tor Latency:**
| Connection Type | Latency (One-Way) | Round-Trip Time |
|-----------------|-------------------|-----------------|
| Same Country    | 500-1000ms        | 1-2 seconds     |
| Same Continent  | 1000-2000ms       | 2-4 seconds     |
| Intercontinental| 2000-4000ms       | 4-8 seconds     |
| Congested/Slow  | 3000-5000ms       | 6-10 seconds    |

**Gameplay Impact:**
- **500-1000ms:** Responsive, similar to high-ping online games
- **1000-2000ms:** Noticeable delay, still playable with prediction
- **2000-4000ms:** Significant lag, requires patience
- **4000-5000ms:** Near tolerance limit, expect delays

Venture's client-side prediction and lag compensation mitigate latency effects, but player actions will always feel delayed compared to low-latency connections.

### Bandwidth

**Per-Client Bandwidth Usage:**
| Update Rate | Default Epsilon | High-Latency Epsilon | Bandwidth     |
|-------------|----------------|----------------------|---------------|
| 20 Hz       | 0.001          | N/A                  | ~100 KB/s     |
| 20 Hz       | 0.01 (Tor)     | 0.01                 | ~50-70 KB/s   |
| 10 Hz       | 0.01 (Tor)     | 0.01                 | ~25-35 KB/s   |

**Server Bandwidth:**
```
Total Bandwidth = (Per-Client Bandwidth) * (Number of Clients)

4 clients @ 20 Hz with epsilon 0.01:
  = 60 KB/s * 4 = 240 KB/s (~1.9 Mbps)
```

**Tor Bandwidth Limitations:**
- Typical onion service bandwidth: 1-10 Mbps
- Burst capacity limited by circuit bandwidth
- Constrained by slowest relay in circuit (usually middle relay)

**Optimizations for Low Bandwidth:**
1. Reduce update rate to 10 Hz (50% bandwidth reduction)
2. Use higher delta compression epsilon (bandwidth vs accuracy tradeoff)
3. Limit max players (fewer clients = less bandwidth)
4. Run server on dedicated Tor relay (higher bandwidth allocation)

### Connection Stability

**Expected Metrics (High-Latency Mode):**
| Metric | Target | Typical Tor |
|--------|--------|-------------|
| Connection Success Rate | >95% | 90-95% |
| Disconnection Rate | <1/hour | 1-2/hour |
| Message Drop Rate | <2% | 1-3% |
| Automatic Reconnection Success | >90% | 85-90% |

**Factors Affecting Stability:**
- Circuit quality (determined by Tor relay selection)
- Network congestion (peak hours: 1800-2200 UTC)
- Geographic distance (intercontinental circuits less stable)
- Tor network health (check status.torproject.org)

## Advanced Configuration

### Customizing Timeouts for Specific Latency

If you know your typical latency is different from 5000ms, customize timeouts:

```go
// For 2000ms latency (4s round-trip)
config := network.ServerConfig{
    Address:      ":8080",
    MaxPlayers:   4,
    ReadTimeout:  20 * time.Second,  // 4s RTT + 16s safety
    WriteTimeout: 15 * time.Second,  // 2s latency + 13s safety
    UpdateRate:   20,
    BufferSize:   256,  // 4s RTT * 20 Hz = 80 messages (3x buffer)
}
```

**Formula:**
```
ReadTimeout = RTT + (RTT * 3)  # 3x safety margin
WriteTimeout = L + (L * 6)     # 6x safety margin
BufferSize = (RTT * UpdateRate) * SafetyFactor

Where:
  L = one-way latency
  RTT = 2 * L (round-trip time)
  SafetyFactor = 2.5 to 3.0 (recommended)
```

### Multiple Servers with Load Balancing

Run multiple hidden services for redundancy:

```
# /etc/tor/torrc
# Primary server
HiddenServiceDir /var/lib/tor/venture-primary/
HiddenServicePort 8080 127.0.0.1:8080

# Backup server
HiddenServiceDir /var/lib/tor/venture-backup/
HiddenServicePort 8080 127.0.0.1:8081
```

Clients can try primary, fallback to backup on timeout:

```go
servers := []string{
    "primary-onion-address.onion:8080",
    "backup-onion-address.onion:8080",
}

for _, server := range servers {
    config.ServerAddress = server
    client := network.NewTCPClient(config)
    err := client.Connect()
    if err == nil {
        break  // Connected successfully
    }
}
```

### Tor Daemon Configuration Tuning

Optimize Tor for gaming traffic:

```
# /etc/tor/torrc

# Increase circuit build timeout
CircuitBuildTimeout 60  # Default: 60 seconds (already max)

# Prefer faster relays
AvoidDiskWrites 1       # Reduce disk I/O
OptimisticData 1        # Send data before circuit fully established

# Connection pooling
ConnectionPadding 1     # Pad connections to resist traffic analysis

# Bandwidth allocation
RelayBandwidthRate 1 MB     # If running relay
RelayBandwidthBurst 2 MB    # Burst capacity
```

### Monitoring and Logging

**Enable detailed network logging:**

```go
// In server main.go
import "github.com/sirupsen/logrus"

logrus.SetLevel(logrus.DebugLevel)
logrus.SetFormatter(&logrus.JSONFormatter{})  // Machine-readable logs
```

**Monitor buffer statistics:**

```go
// In server loop
go func() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        stats := server.GetBufferStats()
        for name, stat := range stats {
            logrus.WithFields(logrus.Fields{
                "buffer":      name,
                "utilization": stat.Utilization(),
                "dropped":     stat.Dropped,
            }).Info("Buffer stats")
        }
    }
}()
```

**Prometheus/Grafana Integration (Advanced):**

Export metrics for monitoring:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    latencyHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name: "venture_client_latency_ms",
        Help: "Client latency in milliseconds",
        Buckets: []float64{100, 500, 1000, 2000, 3000, 5000, 10000},
    })
)

func init() {
    prometheus.MustRegister(latencyHistogram)
}

// In network code
latencyHistogram.Observe(float64(latency.Milliseconds()))
```

## References

**Official Documentation:**
- [Tor Project](https://www.torproject.org/)
- [Tor Hidden Services Guide](https://community.torproject.org/onion-services/)
- [Tor Protocol Specification](https://spec.torproject.org/)

**Venture Documentation:**
- [MULTIPLAYER.md](./MULTIPLAYER.md) - General multiplayer networking
- [ARCHITECTURE.md](./ARCHITECTURE.md) - System architecture overview
- [PERFORMANCE.md](./PERFORMANCE.md) - Performance optimization guide

**Network Programming:**
- [Go SOCKS5 Proxy](https://pkg.go.dev/golang.org/x/net/proxy)
- [Go net Package](https://pkg.go.dev/net)

## Conclusion

Running Venture over Tor enables anonymous, censorship-resistant multiplayer gaming. While latency is higher than clearnet connections, Venture's high-latency networking stack is designed to provide a playable experience even at 5000ms latencies.

**Quick Start Summary:**
1. Install Tor on server and client
2. Configure hidden service in `/etc/tor/torrc`
3. Start Venture server with `-high-latency` flag
4. Clients connect via SOCKS5 proxy using onion address
5. Expect 1-5 second latencies, stable connections with automatic reconnection

**For Support:**
- File issues on [GitHub](https://github.com/opd-ai/venture)
- Check [TROUBLESHOOTING.md](./MULTIPLAYER.md#troubleshooting) for common problems
- Review server/client logs for detailed error messages

**Security Reminder:** Always verify onion addresses through trusted channels, keep Tor updated, and protect hidden service private keys.

Happy anonymous gaming! 🎮🔒
