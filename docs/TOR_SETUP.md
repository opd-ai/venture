# Tor Setup Guide for Venture Multiplayer

Guide for running Venture multiplayer over Tor network (hidden services and SOCKS5 proxy).

## Overview

**Features:** Anonymous multiplayer, censorship-resistant, privacy-preserving co-op  
**Network Support:** Up to 5000ms latency, 60s read timeout, 30s write timeout, 512-message buffers, TCP keepalive, auto-reconnect, 15s lag compensation

## Prerequisites

**Requirements:** Venture binaries, Tor 0.4.5+, Linux/macOS/Windows

**Install Tor:**
```bash
# Debian/Ubuntu
sudo apt install tor && sudo systemctl enable --now tor

# macOS
brew install tor && brew services start tor

# Arch
sudo pacman -S tor && sudo systemctl enable --now tor

# Windows: Download Tor Expert Bundle or Tor Browser
```

**Verify:**
```bash
curl --socks5-hostname localhost:9050 https://check.torproject.org/api/ip
# Expected: {"IsTor":true,"IP":"<tor_exit_ip>"}
```

## Server Setup (Hidden Service)

**1. Configure Tor (`/etc/tor/torrc`):**
```
HiddenServiceDir /var/lib/tor/venture/
HiddenServicePort 8080 127.0.0.1:8080
```

**2. Restart Tor:** `sudo systemctl restart tor`

**3. Get Onion Address:** `sudo cat /var/lib/tor/venture/hostname`  
(Example: `abcdefghijklmnop.onion`)

**4. Start Server:**
```bash
./venture-server -high-latency -port 8080 -max-players 4
```

**Important:** Server binds to `127.0.0.1:8080` (localhost). Tor forwards onion connections to this port. Never use `0.0.0.0`.

**Backup Onion Address:**
```bash
sudo tar -czf venture-onion-backup.tar.gz /var/lib/tor/venture/
```

## Client Setup (SOCKS5 Proxy)

**Option A: Environment Variable**
```bash
export ALL_PROXY=socks5://127.0.0.1:9050
./venture-client -multiplayer -server abcdefghijklmnop.onion:8080
```

**Option B: Custom Dialer (Add to `pkg/network/client.go`)**
```go
import "golang.org/x/net/proxy"

dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
if err != nil {
    return fmt.Errorf("SOCKS5 dialer failed: %w", err)
}
conn, err := dialer.Dial("tcp", c.config.ServerAddress)
```

**Option C: TorClientConfig Preset**
```go
config := network.TorClientConfig()
config.ServerAddress = "abcdefghijklmnop.onion:8080"
client := network.NewTCPClient(config)
```

**First Connection:** May take 30-60s for circuit building. Subsequent connects are faster.

## Testing

**Local Test:**
```bash
# Terminal 1
./venture-server -high-latency -port 8080

# Terminal 2
export ALL_PROXY=socks5://127.0.0.1:9050
./venture-client -multiplayer -server $(sudo cat /var/lib/tor/venture/hostname):8080
```

**Latency Simulation (Linux):**
```bash
sudo tc qdisc add dev lo root netem delay 2500ms  # Add 5s RTT
# Test connection
sudo tc qdisc del dev lo root  # Remove delay
```

## Troubleshooting

**Connection Timeout:**
- Verify Tor running: `systemctl status tor`
- Test SOCKS: `curl --socks5-hostname localhost:9050 https://check.torproject.org/api/ip`
- Check server accessible: `netstat -an | grep 8080`
- Verify onion address correct (16 or 56 chars)

**SOCKS Error:**
- Check SOCKS port (9050 default, 9150 Tor Browser)
- Verify config: `cat /etc/tor/torrc | grep SocksPort`
- Try `ALL_PROXY=socks5h://127.0.0.1:9050` (forces DNS through proxy)

**Frequent Disconnects:**
- Ensure `-high-latency` flag on server
- Check buffer utilization in logs
- Reduce update rate to 10 Hz: `config.UpdateRate = 10`

**Circuit Failures:**
- Delete Tor state: `sudo systemctl stop tor && sudo rm /var/lib/tor/state && sudo systemctl start tor`
- Update Tor: `sudo apt upgrade tor` (0.4.7+ recommended)

**Onion Address Changed:**
- Restore backup: `sudo tar -xzf venture-onion-backup.tar.gz -C /`
- No backup = old address permanently lost

## Security

**Server:**
- Bind to `127.0.0.1` only (never `0.0.0.0`)
- Block external access: `sudo ufw deny 8080`
- Keep `/var/lib/tor/venture/` secure (contains private key)
- Permissions: `sudo chown debian-tor:debian-tor /var/lib/tor/venture/ && sudo chmod 700 /var/lib/tor/venture/`

**Client:**
- Use Tor Browser if untrusted network
- Verify onion address authenticity
- Don't reuse Tor circuits across servers

**Both:**
- Update Tor regularly
- Monitor logs: `sudo journalctl -u tor -f`
- Use v3 onion addresses (56 chars, stronger crypto than v2)

## Performance

**Expected Latency:**
- Clearnet: 10-100ms
- Tor: 500-3000ms typical, up to 5000ms max
- First connection: 30-60s (circuit building)
- Subsequent: 5-15s

**Optimization:**
- Prefer fast exit nodes: Add `ExitNodes {us},{ca},{gb}` to `torrc`
- Use geographically close bridges
- Reduce update rate: 10-20 Hz sufficient for Tor
- Enable compression (future enhancement)

**Bandwidth:**
- Per player: ~100 KB/s at 20 Hz
- 4 players: ~400 KB/s total
- Tor handles this easily (relays typically >1 MB/s)

## Advanced

**Multiple Servers:**
```
HiddenServiceDir /var/lib/tor/venture1/
HiddenServicePort 8080 127.0.0.1:8080

HiddenServiceDir /var/lib/tor/venture2/
HiddenServicePort 8081 127.0.0.1:8081
```

**Tor Bridges (if Tor blocked):**
```
UseBridges 1
Bridge obfs4 <bridge_line>  # Get from https://bridges.torproject.org
```

**Monitor Circuits:**
```bash
sudo journalctl -u tor -f | grep -i "circuit\|bootstrap"
```

**High Latency Config Values:**
- Read timeout: 60s
- Write timeout: 30s
- Connection timeout: 60s
- Buffer size: 512 messages
- Ping interval: 5s
- Lag compensation: 15s history
- Prediction buffer: 12.8s

---

**Last Updated:** November 14, 2025
