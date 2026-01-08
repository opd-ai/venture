# Runbook: Network Connectivity Issues

**Severity:** P1 (Service Disruption)  
**Symptoms:** Players cannot connect, disconnections, high latency, packet loss  
**Owner:** Infrastructure Team  
**Last Updated:** 2026-01-07

---

## Overview

This runbook helps diagnose and resolve network connectivity issues affecting Venture game servers. Network issues can range from complete inability to connect to degraded performance from packet loss or high latency.

**Expected State:** <50ms latency on LAN, <200ms on internet, <1% packet loss  
**Alert Threshold:** >500ms latency or >5% packet loss

---

## Initial Assessment (5 minutes)

### 1. Verify Server is Running

Check if the server process is active:

```bash
# Check process status
systemctl status venture-server

# Or
ps aux | grep venture-server
```

If not running:
```bash
systemctl start venture-server
# Check logs for startup errors
journalctl -u venture-server -n 50
```

### 2. Verify Network Listener

Check if server is listening on expected port (default: 8080):

```bash
# Linux
ss -tulpn | grep 8080
# Or
netstat -tulpn | grep 8080

# macOS
lsof -i :8080

# Should show: LISTEN on 0.0.0.0:8080 or :::8080
```

If not listening:
- Check if another process is using the port
- Review server configuration: `grep port /etc/venture/server.conf`
- Check server startup logs: `journalctl -u venture-server -n 100`

### 3. Test Local Connectivity

Test connection from localhost:

```bash
# Test health endpoint
curl -v http://localhost:8081/health

# Test game port (UDP and TCP)
nc -zv localhost 8080  # TCP
nc -zuv localhost 8080 # UDP

# Expected: Connection successful
```

If localhost fails, server networking is broken (check logs, file bug).

### 4. Test External Connectivity

From another machine on the same network:

```bash
# Replace SERVER_IP with actual server IP
SERVER_IP=192.168.1.100

# Test health endpoint
curl -v http://$SERVER_IP:8081/health

# Test game port
nc -zv $SERVER_IP 8080  # TCP
nc -zuv $SERVER_IP 8080 # UDP

# Test basic connectivity
ping -c 5 $SERVER_IP
```

If external fails but localhost works: **Firewall issue** (see Scenario 1).

---

## Diagnosis (10 minutes)

### Step 1: Check Server Status and Metrics

```bash
# Get network statistics
curl http://localhost:8081/status | jq '.network'

# Check for abnormal values:
# - connected_players: Should match expected count
# - bytes_sent/received: Should be non-zero if players connected
# - packets_sent/received: Should be growing if active
```

### Step 2: Check Firewall Rules

**Linux (ufw):**
```bash
# Check firewall status
sudo ufw status verbose

# Expected rules:
# 8080/tcp ALLOW IN Anywhere
# 8080/udp ALLOW IN Anywhere
# 8081/tcp ALLOW IN Anywhere  (metrics endpoint)
```

**Linux (iptables):**
```bash
# Check INPUT chain
sudo iptables -L INPUT -n -v | grep 8080

# Expected: ACCEPT rules for tcp and udp on port 8080
```

**macOS:**
```bash
# Check if firewall is enabled
/usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate

# List applications
/usr/libexec/ApplicationFirewall/socketfilterfw --listapps | grep venture
```

**Windows:**
```powershell
# Check firewall rules
Get-NetFirewallRule | Where-Object {$_.DisplayName -like "*venture*"}

# Check if port 8080 is allowed
Get-NetFirewallPortFilter | Where-Object {$_.LocalPort -eq 8080}
```

### Step 3: Check Router/NAT Configuration

If players on internet cannot connect but LAN works:

```bash
# Verify public IP
curl ifconfig.me

# Provide this IP to players for connection
```

Check router port forwarding:
1. Access router admin (usually 192.168.1.1 or 192.168.0.1)
2. Navigate to Port Forwarding / Virtual Server
3. Verify rules:
   - External Port: 8080, Internal Port: 8080, Protocol: TCP+UDP, IP: [server-local-ip]
   - External Port: 8081, Internal Port: 8081, Protocol: TCP, IP: [server-local-ip]

### Step 4: Check Network Latency and Packet Loss

From client machine:

```bash
SERVER_IP=<server-public-ip>

# Test latency
ping -c 20 $SERVER_IP

# Calculate stats
# Expected: avg <50ms LAN, <200ms internet
# Alert if: avg >500ms or >5% packet loss
```

If high latency or packet loss:
- Check if server network is congested: `iftop` or `nethogs`
- Check if ISP having issues: `mtr $SERVER_IP` (shows hop-by-hop latency)
- Consider enabling high-latency mode: `-high-latency` flag

### Step 5: Check Server Logs for Network Errors

```bash
# Check for network-related errors
grep -i "network\|connection\|socket" /var/log/venture-server.log | tail -50

# Common error patterns:
# - "bind: address already in use" → Port conflict
# - "too many open files" → File descriptor limit
# - "connection refused" → Client cannot reach server
# - "connection reset" → Network instability or firewall
# - "i/o timeout" → High latency or packet loss
```

### Step 6: Test Federation Connectivity (if enabled)

If federation is enabled, test connectivity to other servers:

```bash
# Check federation health
curl http://localhost:8081/status | jq '.federation'

# Or check federation-specific endpoint if available
# See federation-debug.md for detailed federation troubleshooting
```

---

## Resolution

### Scenario 1: Firewall Blocking Connections

**Symptom:** Localhost works, external connections fail  
**Root Cause:** Firewall not allowing inbound traffic

**Linux (ufw):**
```bash
# Allow game port (TCP + UDP)
sudo ufw allow 8080/tcp
sudo ufw allow 8080/udp

# Allow metrics endpoint (TCP only)
sudo ufw allow 8081/tcp

# Verify rules
sudo ufw status
```

**Linux (iptables):**
```bash
# Add rules to allow game port
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 8080 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8081 -j ACCEPT

# Save rules (persistence varies by distro)
sudo iptables-save > /etc/iptables/rules.v4  # Debian/Ubuntu
```

**macOS:**
```bash
# Allow venture-server through firewall
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --add /path/to/venture-server
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --unblock /path/to/venture-server
```

**Windows:**
```powershell
# Add firewall rule
New-NetFirewallRule -DisplayName "Venture Server" -Direction Inbound -Protocol TCP -LocalPort 8080,8081 -Action Allow
New-NetFirewallRule -DisplayName "Venture Server UDP" -Direction Inbound -Protocol UDP -LocalPort 8080 -Action Allow
```

### Scenario 2: Port Already in Use

**Symptom:** Server fails to start, "bind: address already in use" in logs  
**Root Cause:** Another process is using port 8080

```bash
# Find process using port 8080
sudo lsof -i :8080
# Or
sudo ss -tulpn | grep 8080

# Kill conflicting process
sudo kill <PID>

# Or change server port
venture-server -port 8081
```

### Scenario 3: File Descriptor Limit

**Symptom:** "too many open files" in logs, cannot accept new connections  
**Root Cause:** Server exceeded max file descriptors (each connection = 1 FD)

```bash
# Check current limit
ulimit -n

# Expected: 4096+ for production servers
# Alert if: <1024 (default on many systems)

# Increase limit temporarily
ulimit -n 65536

# Restart server
systemctl restart venture-server
```

Make permanent by editing `/etc/security/limits.conf`:
```
venture  soft  nofile  65536
venture  hard  nofile  65536
```

### Scenario 4: High Latency / Packet Loss

**Symptom:** Players experience lag, disconnections, delayed actions  
**Root Cause:** Network congestion or poor connection quality

**Server-side mitigation:**
```bash
# Enable high-latency mode (optimizes for 200-5000ms RTT)
venture-server -high-latency

# Reduce tick rate (less frequent updates = less bandwidth)
venture-server -tick-rate 20  # Default: 60

# Enable network compression (reduces bandwidth usage)
venture-server -network-compression
```

**Client-side:**
```bash
# Connect with high-latency mode
venture-client -server $SERVER_IP:8080 -high-latency

# Increase timeout
venture-client -server $SERVER_IP:8080 -timeout 30
```

**Network-side:**
- Ensure server has sufficient bandwidth (recommend 1Mbps per 10 players)
- Check for bandwidth-heavy processes: `iftop` or `nethogs`
- Consider QoS rules to prioritize game traffic

### Scenario 5: Router/NAT Issues

**Symptom:** LAN connections work, internet connections fail  
**Root Cause:** Port forwarding not configured or incorrect

1. **Verify port forwarding:**
   - Log into router admin
   - Navigate to Port Forwarding (varies by router model)
   - Add rule: External 8080 → Internal [server-local-ip]:8080, Protocol: TCP+UDP
   - Add rule: External 8081 → Internal [server-local-ip]:8081, Protocol: TCP

2. **Test from outside network:**
   ```bash
   # From phone with mobile data (not WiFi)
   curl http://[public-ip]:8081/health
   ```

3. **Alternative: UPnP (if supported by router):**
   ```bash
   # Enable UPnP on server (if available)
   venture-server -upnp
   
   # Server will automatically request port forwarding
   ```

### Scenario 6: UDP Blocked

**Symptom:** TCP works (health endpoint accessible), game doesn't (uses UDP)  
**Root Cause:** Firewall or ISP blocking UDP

```bash
# Test UDP specifically
nc -zuv $SERVER_IP 8080

# If UDP blocked by ISP (no fix), use TCP mode
venture-server -force-tcp

# Note: TCP mode has higher latency, not ideal for action game
```

---

## Emergency Response (Server Unreachable)

If server is completely unreachable and standard troubleshooting fails:

### Option 1: Restart Network Stack

```bash
# Restart networking (will briefly disconnect all clients)
sudo systemctl restart networking  # Debian/Ubuntu
sudo systemctl restart NetworkManager  # Fedora/RHEL

# Restart server
systemctl restart venture-server

# Verify connectivity
curl http://localhost:8081/health
```

### Option 2: Restart Server on Different Port

```bash
# Stop current server
systemctl stop venture-server

# Start on alternate port
venture-server -port 8082 &

# Update firewall
sudo ufw allow 8082/tcp
sudo ufw allow 8082/udp

# Test connectivity
nc -zv localhost 8082
```

### Option 3: Fallback to Safe Mode

```bash
# Start server with minimal features (no federation, mods, etc.)
venture-server -safe-mode -port 8080

# This disables complex networking features that might be failing
```

---

## Monitoring and Verification (15 minutes)

After resolution, verify connectivity and stability:

```bash
# Test from multiple clients
# Client 1: localhost
curl http://localhost:8081/health

# Client 2: LAN
curl http://192.168.1.100:8081/health

# Client 3: Internet (from outside network)
curl http://[public-ip]:8081/health

# Monitor connection count
watch -n 5 'curl -s http://localhost:8081/status | jq .network.connected_players'

# Monitor packet rates
watch -n 5 'curl -s http://localhost:8081/metrics | grep packets'
```

**Success Criteria:**
- All three connectivity tests pass
- Players can connect and stay connected >5 minutes
- Latency <200ms on internet, <50ms on LAN
- Packet loss <1%

---

## Prevention

### Network Monitoring

Set up alerts for network issues:

```yaml
# Prometheus alert rules
- alert: NoConnectedPlayers
  expr: venture_players_connected == 0
  for: 10m
  annotations:
    summary: "No players connected for 10 minutes, connectivity issue?"
    
- alert: HighPacketLoss
  expr: rate(venture_network_packets_sent_total[1m]) - rate(venture_network_packets_received_total[1m]) > 100
  annotations:
    summary: "High packet loss detected"
    
- alert: NetworkStorm
  expr: rate(venture_network_packets_sent_total[1m]) > 5000
  annotations:
    summary: "Sending >5000 packets/sec, network storm"
```

### Regular Connectivity Tests

Set up automated connectivity tests from external monitoring service:

```bash
# Use external monitoring service (e.g., UptimeRobot, Pingdom)
# Monitor: http://[public-ip]:8081/health
# Frequency: Every 5 minutes
# Alert on: 2 consecutive failures
```

### Bandwidth Capacity Planning

Review bandwidth usage weekly:

```bash
# Check peak bandwidth
curl http://localhost:8081/metrics | grep network_bytes

# Calculate bandwidth: (bytes_sent + bytes_received) / uptime_seconds

# Expected: <100 KB/s per player
# Alert if: Approaching 80% of available bandwidth
```

If approaching limits, consider:
- Upgrading network connection
- Implementing compression (`-network-compression`)
- Reducing tick rate (`-tick-rate 30`)
- Scaling horizontally (more instances, geo-distributed)

---

## Escalation

If network issues persist after all resolution attempts:

1. **Collect Network Diagnostics:**
   ```bash
   # Create diagnostic directory
   mkdir venture-network-diag-$(date +%Y%m%d-%H%M%S)
   cd venture-network-diag-*
   
   # Capture network stats
   curl http://localhost:8081/status > status.json
   curl http://localhost:8081/metrics > metrics.txt
   
   # Capture firewall rules
   sudo iptables -L -n -v > iptables.txt
   sudo ufw status verbose > ufw.txt
   
   # Capture listening ports
   ss -tulpn > listening-ports.txt
   
   # Capture route table
   ip route > routes.txt
   netstat -rn >> routes.txt
   
   # Capture interface stats
   ip -s link > interfaces.txt
   ifconfig -a >> interfaces.txt
   
   # Test connectivity
   ping -c 20 8.8.8.8 > ping-google.txt
   traceroute 8.8.8.8 > traceroute-google.txt
   
   # Capture logs
   cp /var/log/venture-server.log .
   journalctl -u venture-server -n 1000 > systemd.log
   
   # Create archive
   cd ..
   tar czf venture-network-diag-$(date +%Y%m%d-%H%M%S).tar.gz venture-network-diag-*
   ```

2. **File GitHub Issue:**
   - Title: "Production: Network connectivity issue - <brief description>"
   - Label: `P1`, `network`, `bug`
   - Attach diagnostic bundle
   - Include: server version, network topology, ISP, affected region

3. **Temporary Workaround:**
   - Use alternative ports if current port is problematic
   - Enable TCP-only mode if UDP is blocked
   - Use high-latency mode for poor connections
   - Consider VPN/tunneling if network path is problematic

---

## Related Runbooks

- [Federation Debugging](federation-debug.md) - If federation connectivity is the issue
- [High CPU Usage](high-cpu.md) - If network storm causing high CPU

---

## Revision History

| Date       | Author           | Changes                          |
|------------|------------------|----------------------------------|
| 2026-01-07 | Infrastructure   | Initial version for v10.0        |
