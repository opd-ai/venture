# Production Deployment Guide

**Version:** 1.1 | **Status:** Production Ready

## Overview

Procedural multiplayer action-RPG with high-latency tolerance (200-5000ms), deterministic generation, and cross-platform support.

**Features:** Authoritative server, client prediction, lag compensation, structured logging, 2-4+ players, <1GB per 4 players

## System Requirements

**Minimum:** 2 cores @ 2.0GHz, 2GB RAM, 100MB disk, 1 Mbps (250 KB/s per player)  
**Recommended:** 4 cores @ 3.0GHz, 4GB RAM, 500MB disk, 10 Mbps, <50ms latency

**OS:** Linux (Ubuntu 20.04+, Debian 11+, RHEL 8+), macOS 11+, Windows Server 2019+  
**Dependencies:** None (static binaries)

**Targets:** 20 TPS, <100 KB/s per player, <50ms tick, <500MB per 4 players

## Deployment Architectures

**Single Server:** Simple, <10 players, LAN/testing
```bash
./venture-server -port 8080 -max-players 10
```

**Multi-Server (Sharded):** Independent worlds, 10-100+ players
```bash
./venture-server -port 8080 -seed 12345 -max-players 4 &
./venture-server -port 8081 -seed 67890 -max-players 4 &
```

**Cloud (Managed):** Auto-scaling, health checks, logging, 100+ players  
Tech: Kubernetes/Docker, ELK/CloudWatch, Prometheus+Grafana, NGINX/HAProxy

## Quick Start

**From GitHub Releases:**
```bash
wget https://github.com/opd-ai/venture/releases/download/v1.1.0/venture-server-linux-amd64.tar.gz
tar -xzf venture-server-linux-amd64.tar.gz
./venture-server-linux-amd64/venture-server -port 8080
```

**Build from Source:**
```bash
git clone https://github.com/opd-ai/venture.git && cd venture
make build-server
./build/server/venture-server -port 8080
```

**Docker:**
```bash
docker run -d -p 8080:8080 ghcr.io/opd-ai/venture-server:latest
```

## Configuration

**CLI Flags:**
```bash
-port 8080                    # Listen port
-max-players 4                # Max concurrent players
-tick-rate 20                 # Updates per second
-high-latency                 # Enable 60s timeouts, 512msg buffers
-seed 12345                   # World seed (deterministic)
-log-level info               # debug, info, warn, error
-log-format json              # json, text
-enable-profiling             # pprof on :6060
```

**Environment Variables:**
```bash
export VENTURE_PORT=8080
export LOG_LEVEL=info
export LOG_FORMAT=json
./venture-server
```

**systemd Service:**
```ini
[Unit]
Description=Venture Server
After=network.target

[Service]
Type=simple
User=venture
ExecStart=/opt/venture/venture-server -port 8080 -log-format json
Restart=on-failure
RestartSec=5s
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

## Monitoring

**Health Check:**
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy","players":2,"uptime":"5h30m"}
```

**Metrics (Prometheus):**
```bash
./venture-server -enable-profiling -port 8080
curl http://localhost:6060/metrics
```

**Key Metrics:** Players connected, TPS, memory usage, network bandwidth, lag compensation queue

**Log Aggregation (JSON):**
```bash
./venture-server -log-format json | jq '.level == "error"'
# Forward to ELK, Splunk, or CloudWatch
```

## Security

**Firewall:**
```bash
sudo ufw allow 8080/tcp
sudo ufw deny 6060/tcp  # pprof internal only
```

**Reverse Proxy (NGINX):**
```nginx
upstream venture {
    server 127.0.0.1:8080;
}
server {
    listen 80;
    server_name venture.example.com;
    location / {
        proxy_pass http://venture;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

**Best Practices:**
- Run as non-root user
- Disable profiling in production (`-enable-profiling=false`)
- Use TLS for public servers
- Rate limit connections (fail2ban, iptables)

## Scaling

**Vertical:** Increase cores/RAM (4 cores → 8 cores = 2x players)  
**Horizontal:** Multiple servers with load balancer  
**Sharding:** Separate worlds per server (easiest)

**Load Balancing (HAProxy):**
```haproxy
backend venture_servers
    balance roundrobin
    server srv1 10.0.1.10:8080 check
    server srv2 10.0.1.11:8080 check
```

## Performance Tuning

**Network:**
```bash
sysctl -w net.core.rmem_max=8388608
sysctl -w net.core.wmem_max=8388608
sysctl -w net.ipv4.tcp_keepalive_time=300
```

**File Descriptors:**
```bash
ulimit -n 65536
```

**Go Runtime:**
```bash
export GOMAXPROCS=4        # Match CPU cores
export GOGC=100            # GC frequency (default)
```

## Backup & Recovery

**Save Files:**
```bash
# Location: $HOME/.venture/saves/
tar -czf venture-saves-$(date +%Y%m%d).tar.gz ~/.venture/saves/
```

**Automated Backups:**
```bash
0 2 * * * tar -czf /backup/venture-$(date +\%Y\%m\%d).tar.gz ~/.venture/saves/
```

**Restore:**
```bash
tar -xzf venture-saves-20250115.tar.gz -C ~/
./venture-server -load-save ~/.venture/saves/world-12345.json
```

## Troubleshooting

**High CPU:** Reduce `-tick-rate` to 10-15, check for runaway AI  
**High Memory:** Reduce `-max-players`, profile with `-enable-profiling`  
**Disconnects:** Enable `-high-latency`, check network latency  
**Desyncs:** Verify deterministic generation (same seed = same world)

**Logs:**
```bash
tail -f venture-server.log | jq 'select(.level == "error")'
```

**Profiling:**
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## CI/CD

**GitHub Actions:**
```yaml
- name: Build Server
  run: make build-server
- name: Run Tests
  run: go test ./...
- name: Deploy
  run: scp build/server/venture-server user@server:/opt/venture/
```

**Docker Build:**
```dockerfile
FROM golang:1.24 AS builder
COPY . /app
RUN cd /app && make build-server

FROM alpine:latest
COPY --from=builder /app/build/server/venture-server /venture-server
EXPOSE 8080
CMD ["/venture-server", "-port", "8080"]
```

---

**Last Updated:** November 14, 2025
