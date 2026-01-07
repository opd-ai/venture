# Runbook: Federation Debugging

**Severity:** P2 (Feature Degradation)  
**Symptoms:** Cannot connect to federated servers, server discovery fails, cross-server travel broken  
**Owner:** Infrastructure Team  
**Last Updated:** 2026-01-07

---

## Overview

This runbook helps diagnose and resolve issues with Venture's federation system, which enables cross-server travel, guild coordination, and distributed gameplay. Federation relies on UDP multicast for discovery, WebRTC for P2P connections, and HTTP for fallback communication.

**Expected State:** Server discovers peers in <10 seconds, cross-server travel <5 seconds  
**Alert Threshold:** No peers discovered in 5 minutes, federation health check failing

---

## Federation Architecture Quick Reference

**Discovery Methods:**
1. **UDP Multicast** (LAN) - Servers broadcast presence on 239.255.42.99:9999
2. **Central Registry** (Internet) - HTTP registration with central discovery server
3. **Manual Peers** - Configured peer addresses in server.conf

**Connection Methods:**
1. **WebRTC P2P** (Preferred) - Direct peer-to-peer with NAT traversal
2. **HTTP Relay** (Fallback) - Proxied through relay server
3. **Direct TCP** (Manual) - Direct connection to known IP:port

---

## Initial Assessment (5 minutes)

### 1. Verify Federation is Enabled

```bash
# Check server configuration
grep -i "federation" /etc/venture/server.conf

# Expected:
# federation_enabled=true

# Or check via command line flags
ps aux | grep venture-server | grep -o "\-federation[^ ]*"
```

If federation disabled, enable it:
```bash
# Edit config
vim /etc/venture/server.conf
# Set: federation_enabled=true

# Or start with flag
venture-server -federation -port 8080
```

### 2. Check Federation Health

```bash
# Check federation status via health endpoint
curl http://localhost:8081/status | jq '.federation'

# Expected fields:
# - enabled: true
# - peers_discovered: >0
# - active_connections: >=0
# - discovery_method: "multicast" or "registry" or "manual"

# If federation object missing or null: Federation not initialized
# If peers_discovered == 0: Discovery failing (see Scenario 1)
# If active_connections == 0 but peers_discovered >0: Connection failing (see Scenario 2)
```

### 3. Check Server Logs

```bash
# Check for federation errors
grep -i "federation\|discovery\|webrtc\|peer" /var/log/venture-server.log | tail -50

# Common error patterns:
# - "multicast: permission denied" → Network permissions issue
# - "no peers discovered" → Discovery failing
# - "webrtc connection failed" → NAT traversal issue
# - "circuit breaker open" → Peer marked as unhealthy
# - "handshake timeout" → Connection negotiation failing
```

### 4. Test Network Connectivity

```bash
# Test multicast capability (LAN discovery)
# On server machine, send test multicast
echo "test" | nc -u 239.255.42.99 9999

# On another machine in same network, receive
nc -ul 9999

# If receive fails: Multicast blocked or network doesn't support it
```

---

## Diagnosis (10 minutes)

### Step 1: Verify Discovery Method

Determine which discovery method is active:

```bash
# Check logs for discovery initialization
grep "discovery.*started\|discovery.*method" /var/log/venture-server.log | tail -5

# Expected log messages:
# "Starting multicast discovery on 239.255.42.99:9999"
# Or: "Registering with central discovery at https://discovery.venture.example.com"
# Or: "Using manual peer list: 192.168.1.10:8080, 192.168.1.11:8080"
```

**LAN (Multicast):**
- Fastest, works within local network only
- Requires multicast support (most LANs support this)
- Default for servers not configured with registry

**Internet (Registry):**
- Works globally, requires internet access
- Servers register with central discovery server
- Enables players worldwide to find servers

**Manual:**
- No automatic discovery
- Requires manual configuration of peer addresses
- Most reliable but least flexible

### Step 2: Test Multicast Discovery (LAN)

If using multicast discovery:

```bash
# Check if multicast is being sent
sudo tcpdump -i any 'host 239.255.42.99 and udp port 9999' -v

# Expected output (every 30 seconds):
# IP server-ip > 239.255.42.99: UDP, length X
# (Broadcast message with server info)

# If no output: Server not sending broadcasts (check logs for errors)
# If output exists: Broadcasts working, check receivers
```

Test multicast reception:

```bash
# On potential peer server, check if receiving
sudo tcpdump -i any 'host 239.255.42.99 and udp port 9999' -v

# Should see broadcasts from other servers
# If not receiving: Network blocking multicast or wrong interface
```

### Step 3: Test Registry Discovery (Internet)

If using registry discovery:

```bash
# Check if server registered successfully
grep "registry.*success\|registry.*registered" /var/log/venture-server.log

# Expected: "Successfully registered with discovery registry"

# If registration failed, check error
grep "registry.*error\|registry.*fail" /var/log/venture-server.log

# Common errors:
# - "connection refused" → Registry server down or unreachable
# - "unauthorized" → Invalid API key or authentication
# - "timeout" → Network issue or slow registry
```

Test registry manually:

```bash
# Query discovery registry for available servers
curl https://discovery.venture.example.com/api/v1/servers

# Expected: JSON list of registered servers
# Should include this server (check by server_id or address)
```

### Step 4: Test WebRTC Connection

If peers discovered but connections failing:

```bash
# Enable WebRTC debug logging
# Edit /etc/venture/server.conf:
webrtc_debug=true

# Restart server
systemctl restart venture-server

# Monitor WebRTC negotiation
tail -f /var/log/venture-server.log | grep -i "webrtc\|ice\|sdp\|stun"

# Expected log flow:
# 1. "Creating WebRTC offer for peer X"
# 2. "ICE candidate gathering started"
# 3. "ICE candidate found: type=host/srflx/relay"
# 4. "Received answer from peer X"
# 5. "WebRTC connection established"

# If stuck at step 2-3: STUN/TURN issue (see Scenario 3)
# If stuck at step 4: Signaling issue (peers not exchanging messages)
# If fails at step 5: NAT/firewall incompatibility
```

### Step 5: Check Circuit Breaker Status

Federation includes circuit breaker for failing peers:

```bash
# Check circuit breaker status
curl http://localhost:8081/federation/circuit-breakers | jq '.'

# Expected output:
# {
#   "peers": [
#     {"address": "peer1:8080", "state": "closed", "failures": 0},
#     {"address": "peer2:8080", "state": "open", "failures": 5}
#   ]
# }

# States:
# - "closed": Healthy, requests allowed
# - "open": Unhealthy, requests blocked
# - "half-open": Testing recovery

# If peer is "open": Too many failures, circuit breaker protecting
```

Reset circuit breaker for specific peer:

```bash
curl -X POST http://localhost:8081/admin/federation/reset-circuit-breaker \
  -d '{"peer": "peer2:8080"}'
```

### Step 6: Test Manual Connection

Test direct connection to known peer:

```bash
# Add manual peer via API
curl -X POST http://localhost:8081/admin/federation/add-peer \
  -d '{"address": "192.168.1.10:8080"}'

# Monitor logs for connection attempt
tail -f /var/log/venture-server.log | grep -i "192.168.1.10"

# Expected:
# "Connecting to manual peer 192.168.1.10:8080"
# "Handshake with 192.168.1.10:8080 successful"
# "Peer 192.168.1.10:8080 added to federation"
```

---

## Resolution

### Scenario 1: Multicast Discovery Failing (LAN)

**Symptom:** No peers discovered, no multicast traffic  
**Root Cause:** Network blocking multicast or wrong interface

```bash
# Check network interface
ip addr show

# Identify correct interface (e.g., eth0, wlan0)
# Server may be binding to wrong interface

# Specify interface for multicast
# Edit /etc/venture/server.conf:
multicast_interface=eth0

# Or via command line
venture-server -multicast-interface eth0

# Restart and verify
systemctl restart venture-server
sudo tcpdump -i eth0 'host 239.255.42.99 and udp port 9999'
```

If multicast still blocked:

```bash
# Check firewall rules
sudo iptables -L INPUT -v | grep 9999
sudo ufw status | grep 9999

# Allow multicast discovery port
sudo iptables -A INPUT -p udp --dport 9999 -j ACCEPT
sudo ufw allow 9999/udp

# Enable IP multicast routing (if needed)
echo 1 | sudo tee /proc/sys/net/ipv4/ip_forward
```

**Fallback:** Switch to registry or manual discovery if multicast unavailable.

### Scenario 2: Registry Discovery Failing (Internet)

**Symptom:** Registration fails, "connection refused" or "timeout"  
**Root Cause:** Registry server unreachable or invalid configuration

```bash
# Verify registry URL
grep "registry" /etc/venture/server.conf

# Test registry connectivity
curl -v https://discovery.venture.example.com/api/v1/health

# If fails:
# - Check DNS: ping discovery.venture.example.com
# - Check firewall: sudo iptables -L OUTPUT | grep HTTPS
# - Check if registry is down (status page or contact admin)
```

If registry unreachable, use manual peers temporarily:

```bash
# Edit /etc/venture/server.conf:
federation_discovery_method=manual
federation_manual_peers=192.168.1.10:8080,192.168.1.11:8080

systemctl restart venture-server
```

### Scenario 3: WebRTC Connection Failing

**Symptom:** Peers discovered, connection attempts fail at ICE negotiation  
**Root Cause:** NAT traversal issue, STUN/TURN server unavailable

**Test STUN server connectivity:**

```bash
# Check STUN server configuration
grep "stun" /etc/venture/server.conf

# Expected: stun_servers=stun:stun.l.google.com:19302

# Test STUN server manually (requires stun-client)
stun stun.l.google.com 19302

# Or use curl test
curl -v stun://stun.l.google.com:19302
```

If STUN failing:

```bash
# Try alternative STUN servers
# Edit /etc/venture/server.conf:
stun_servers=stun:stun1.l.google.com:19302,stun:stun2.l.google.com:19302

# Or use TURN server for relay (requires TURN server setup)
turn_servers=turn:turn.example.com:3478?transport=udp
turn_username=user
turn_credential=pass

systemctl restart venture-server
```

**Fallback to HTTP relay:**

```bash
# Force HTTP relay mode (bypasses WebRTC)
# Edit /etc/venture/server.conf:
federation_force_relay=true

systemctl restart venture-server
```

Note: HTTP relay has higher latency but works through restrictive firewalls.

### Scenario 4: Circuit Breaker Open (Peer Marked Unhealthy)

**Symptom:** Peer discovered but circuit breaker prevents connection  
**Root Cause:** Too many failed requests to peer

```bash
# Check circuit breaker status
curl http://localhost:8081/federation/circuit-breakers | jq '.'

# If peer is "open", check peer's health directly
curl http://peer-ip:8081/health

# If peer is healthy, reset circuit breaker
curl -X POST http://localhost:8081/admin/federation/reset-circuit-breaker \
  -d '{"peer": "peer-ip:8080"}'

# Monitor connection attempt
tail -f /var/log/venture-server.log | grep "peer-ip"
```

If peer is genuinely unhealthy:
- Contact peer admin
- Check peer's logs for issues
- Remove peer temporarily: `curl -X DELETE http://localhost:8081/admin/federation/remove-peer?address=peer-ip:8080`

### Scenario 5: Handshake Timeout

**Symptom:** Connection initiated but handshake never completes  
**Root Cause:** Version mismatch, authentication issue, or network latency

```bash
# Check handshake timeout setting
grep "handshake.*timeout" /etc/venture/server.conf

# Default: 10 seconds
# Increase for high-latency networks (e.g., Tor)

# Edit /etc/venture/server.conf:
federation_handshake_timeout_seconds=30

systemctl restart venture-server
```

Check version compatibility:

```bash
# Get this server's version
curl http://localhost:8081/status | jq '.version'

# Get peer's version
curl http://peer-ip:8081/status | jq '.version'

# Federation requires compatible versions (same major version)
# If major version differs: Upgrade required
```

### Scenario 6: Firewall Blocking Federation

**Symptom:** All discovery methods fail, connections timeout  
**Root Cause:** Firewall blocking federation ports

```bash
# Federation uses multiple ports:
# - 9999/UDP: Multicast discovery
# - 8080/TCP+UDP: Game traffic (or configured port)
# - 8081/TCP: Metrics/admin (if exposed)
# - Random UDP: WebRTC data channels

# Allow federation ports
sudo ufw allow 9999/udp comment "Federation discovery"
sudo ufw allow 8080/tcp comment "Federation game traffic TCP"
sudo ufw allow 8080/udp comment "Federation game traffic UDP"

# For WebRTC, allow UDP range (if possible)
sudo ufw allow 10000:20000/udp comment "WebRTC data channels"

# Reload and verify
sudo ufw reload
sudo ufw status numbered
```

---

## Monitoring and Verification (15 minutes)

After applying fixes, verify federation is working:

```bash
# Monitor peer discovery
watch -n 10 'curl -s http://localhost:8081/status | jq .federation'

# Expected to see:
# - peers_discovered increasing (discovery working)
# - active_connections increasing (connections working)

# Test cross-server travel
# 1. Connect client to this server
venture-client -server localhost:8080

# 2. Request travel to peer
# (In-game: Use federation portal or /travel command)

# 3. Monitor logs for travel sequence
tail -f /var/log/venture-server.log | grep -i "travel\|transfer\|federation"

# Expected log flow:
# "Player X requested travel to peer Y"
# "Initiating player transfer to peer Y"
# "Player X transferred successfully"
```

**Success Criteria:**
- Peers discovered: >0 (ideally all expected peers)
- Active connections: Matches peer count
- Cross-server travel: <5 seconds
- No circuit breakers in "open" state

---

## Prevention

### Federation Health Monitoring

Set up monitoring for federation issues:

```yaml
# Prometheus alert rules
- alert: NoPeersDiscovered
  expr: venture_federation_peers_discovered == 0
  for: 10m
  annotations:
    summary: "No federation peers discovered in 10 minutes"
    
- alert: FederationConnectionsLow
  expr: venture_federation_active_connections < venture_federation_peers_discovered * 0.5
  for: 5m
  annotations:
    summary: "Less than 50% of peers connected"
    
- alert: CircuitBreakerOpen
  expr: venture_federation_circuit_breakers_open > 0
  annotations:
    summary: "Federation circuit breaker opened for peer"
```

### Regular Federation Tests

Schedule automated federation tests:

```bash
# Add daily federation test
crontab -e

# Add line:
0 5 * * * /usr/local/bin/test-federation.sh

# Create test script
cat > /usr/local/bin/test-federation.sh << 'EOF'
#!/bin/bash
PEERS=$(curl -s http://localhost:8081/status | jq -r '.federation.peers_discovered')

if [ "$PEERS" -eq 0 ]; then
  echo "$(date): FAIL - No peers discovered" >> /var/log/federation-test.log
  # Send alert
  echo "Federation test failed: No peers discovered" | mail -s "Federation Alert" admin@example.com
else
  echo "$(date): OK - $PEERS peers discovered" >> /var/log/federation-test.log
fi
EOF

chmod +x /usr/local/bin/test-federation.sh
```

### Capacity Planning for Federation

Monitor federation resource usage:

```bash
# Check bandwidth usage for federation
# Federation adds ~10KB/s per peer for heartbeats
# Player transfers add 100-500KB per transfer

# Monitor federation bandwidth separately
curl http://localhost:8081/metrics | grep federation_bytes

# If approaching bandwidth limits, consider:
# - Reducing peer count
# - Optimizing transfer size
# - Rate limiting cross-server travel
```

---

## Escalation

If federation issues persist:

1. **Collect Federation Diagnostics:**
   ```bash
   mkdir venture-federation-diag-$(date +%Y%m%d-%H%M%S)
   cd venture-federation-diag-*
   
   # Capture federation status
   curl http://localhost:8081/status | jq '.federation' > status.json
   curl http://localhost:8081/federation/circuit-breakers > circuit-breakers.json
   
   # Capture network info
   ip addr > network-interfaces.txt
   ip route > routes.txt
   ss -tulpn | grep 9999 > listening-ports.txt
   
   # Capture multicast traffic sample
   timeout 60 sudo tcpdump -i any 'host 239.255.42.99' -w multicast.pcap
   
   # Capture WebRTC logs (if debug enabled)
   grep -i "webrtc\|ice\|stun" /var/log/venture-server.log > webrtc.log
   
   # Capture general federation logs
   grep -i "federation\|peer\|discovery" /var/log/venture-server.log > federation.log
   
   # Test connectivity to known peers
   for peer in $(cat /etc/venture/server.conf | grep manual_peers | cut -d= -f2 | tr ',' ' '); do
     echo "Testing $peer" >> connectivity.txt
     nc -zv $(echo $peer | cut -d: -f1) $(echo $peer | cut -d: -f2) >> connectivity.txt 2>&1
   done
   
   cd ..
   tar czf venture-federation-diag-$(date +%Y%m%d-%H%M%S).tar.gz venture-federation-diag-*
   ```

2. **File GitHub Issue:**
   - Title: "Production: Federation connectivity issue - <brief description>"
   - Label: `P2`, `federation`, `bug`
   - Attach diagnostic bundle
   - Include: network topology, discovery method, NAT configuration

3. **Temporary Workaround:**
   - Disable federation if not critical: `-federation=false`
   - Use manual peer configuration for known reliable peers
   - Fall back to HTTP relay if WebRTC failing
   - Reduce peer count to most stable servers

---

## Related Runbooks

- [Network Issues](network-issues.md) - If basic network connectivity is the issue
- [High CPU Usage](high-cpu.md) - If federation causing performance issues

---

## Revision History

| Date       | Author           | Changes                          |
|------------|------------------|----------------------------------|
| 2026-01-07 | Infrastructure   | Initial version for v10.0        |
