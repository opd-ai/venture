# Production Deployment Guide

**Version:** 1.1  
**Last Updated:** October 29, 2025  
**Status:** Production Ready

---

## Table of Contents

1. [Overview](#overview)
2. [System Requirements](#system-requirements)
3. [Deployment Architectures](#deployment-architectures)
4. [Server Setup](#server-setup)
5. [Configuration](#configuration)
6. [Monitoring & Logging](#monitoring--logging)
7. [Security Best Practices](#security-best-practices)
8. [Scaling Strategies](#scaling-strategies)
9. [Performance Tuning](#performance-tuning)
10. [Backup & Recovery](#backup--recovery)
11. [Troubleshooting](#troubleshooting)
12. [CI/CD Integration](#cicd-integration)

---

## Overview

Venture is a fully procedural multiplayer action-RPG designed for production deployment with high-latency tolerance (200-5000ms) and cross-platform support.

**Key Features:**
- Authoritative server architecture with client-side prediction and lag compensation
- Deterministic procedural generation (seed-based)
- Structured JSON logging
- Support for 2-4+ concurrent players per server
- Memory-efficient design (<1GB per 4 players)

**Target Environments:** Cloud VMs (AWS, GCP, Azure), bare metal servers, Docker/Kubernetes, edge locations (Tor supported)

---

## System Requirements

### Minimum Server Requirements

**Hardware:**
- CPU: 2 cores @ 2.0 GHz (x86_64/ARM64)
- RAM: 2 GB (512 MB per player)
- Disk: 100 MB binaries, 10 MB per save
- Network: 1 Mbps upstream (250 KB/s per player)

**OS:** Linux (Ubuntu 20.04+, Debian 11+, RHEL 8+), macOS 11+, Windows Server 2019+
**Dependencies:** None (static binaries)

### Recommended Production

- CPU: 4 cores @ 3.0 GHz
- RAM: 4 GB (1 GB per player)
- Disk: 500 MB
- Network: 10 Mbps upstream, <50ms latency

**Performance Targets:** 20 TPS, <100 KB/s per player, <50ms tick time, <500 MB memory per 4 players

---

## Deployment Architectures

### Single Server (Simple)
Best for small groups, LAN parties, testing. Single point of failure, no load balancing, <10 concurrent players.

```bash
./venture-server -port 8080 -max-players 10 -tick-rate 20
```

### Multi-Server (Sharded)
Independent worlds with different seeds. Linear scaling, isolated failures, 10-100+ players.

```bash
./venture-server -port 8080 -seed 12345 -max-players 4 &
./venture-server -port 8081 -seed 67890 -max-players 4 &
```

### Cloud Deployment (Managed)
Auto-scaling, health checks, centralized logging, geographic distribution, 100+ players.

**Technologies:** Kubernetes/Docker Swarm/AWS ECS, ELK Stack/CloudWatch, Prometheus+Grafana, NGINX/HAProxy

---

## Server Setup

### Quick Start

**From GitHub Releases:**
```bash
VERSION=v1.1.0
wget https://github.com/opd-ai/venture/releases/download/${VERSION}/venture-server-linux-amd64.tar.gz
tar -xzf venture-server-linux-amd64.tar.gz && cd venture-server-linux-amd64
./venture-server -port 8080
```

**From Source:**
```bash
git clone https://github.com/opd-ai/venture.git && cd venture
go build -ldflags="-s -w" -o venture-server ./cmd/server
./venture-server -port 8080
```

### Systemd Service

**File:** `/etc/systemd/system/venture-server.service`

```ini
[Unit]
Description=Venture Game Server
After=network.target

[Service]
Type=simple
User=venture
Group=venture
WorkingDirectory=/opt/venture
ExecStart=/opt/venture/venture-server -port 8080 -max-players 10 -tick-rate 20
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
Environment="LOG_LEVEL=info" "LOG_FORMAT=json"
LimitNOFILE=65536
MemoryLimit=2G
CPUQuota=400%

[Install]
WantedBy=multi-user.target
```

**Setup:**
```bash
sudo useradd -r -s /bin/false venture
sudo mkdir -p /opt/venture && sudo cp venture-server /opt/venture/
sudo chown -R venture:venture /opt/venture && sudo chmod +x /opt/venture/venture-server
sudo cp venture-server.service /etc/systemd/system/ && sudo systemctl daemon-reload
sudo systemctl enable venture-server && sudo systemctl start venture-server
```

### Docker Deployment

**Dockerfile:**
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o venture-server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates && addgroup -S venture && adduser -S venture -G venture
WORKDIR /app
COPY --from=builder /build/venture-server .
USER venture
EXPOSE 8080
ENTRYPOINT ["./venture-server"]
CMD ["-port", "8080", "-max-players", "10"]
```

**Run:**
```bash
docker build -t venture-server:latest .
docker run -d --name venture-server -p 8080:8080 -e LOG_LEVEL=info --restart unless-stopped venture-server:latest
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: venture-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: venture-server
  template:
    metadata:
      labels:
        app: venture-server
    spec:
      containers:
      - name: venture-server
        image: venture-server:latest
        ports:
        - containerPort: 8080
        env:
        - name: LOG_LEVEL
          value: "info"
        args: ["-port", "8080", "-max-players", "4"]
        resources:
          requests: {memory: "512Mi", cpu: "500m"}
          limits: {memory: "1Gi", cpu: "2000m"}
        livenessProbe:
          tcpSocket: {port: 8080}
          initialDelaySeconds: 30
---
apiVersion: v1
kind: Service
metadata:
  name: venture-server
spec:
  type: LoadBalancer
  selector:
    app: venture-server
  ports:
  - protocol: TCP
    port: 8080
```

---

## Configuration

### Command-Line Flags

```bash
./venture-server \
  -port 8080 \              # Server port (default: 8080)
  -max-players 10 \         # Max concurrent players (default: 4)
  -tick-rate 20 \           # Update rate Hz (default: 20)
  -seed 12345 \             # World seed (default: random)
  -genre fantasy \          # fantasy/scifi/horror/cyberpunk/postapoc
  -verbose                  # Debug logging
```

### Environment Variables

```bash
export LOG_LEVEL=info        # debug/info/warn/error/fatal
export LOG_FORMAT=json       # json/text
```

### Genre Configuration

| Genre | Theme | Characteristics |
|-------|-------|-----------------|
| `fantasy` | Medieval Fantasy | Magic, dungeons, dragons |
| `scifi` | Science Fiction | Tech, lasers, robots |
| `horror` | Horror/Gothic | Dark atmosphere, monsters |
| `cyberpunk` | Cyberpunk | Neon cities, tech noir |
| `postapoc` | Post-Apocalyptic | Wasteland, survivors |

### Tick Rate Selection

| Tick Rate | Use Case | CPU | Network | Lag |
|-----------|----------|-----|---------|-----|
| 10 Hz | High latency | Low | Minimal | 100ms |
| 20 Hz (default) | Balanced | Medium | Moderate | 50ms |
| 30 Hz | Competitive | High | High | 33ms |
| 60 Hz | LAN only | Very High | Very High | 16ms |

---

## Monitoring & Logging

### Structured Logging

Venture uses structured JSON logging via logrus. Logs include timestamp, level, component, and structured fields.

**Levels:** `debug` (diagnostics), `info` (lifecycle), `warn` (degraded), `error` (failures), `fatal` (shutdown)

**Example:**
```json
{
  "time": "2025-10-29T12:34:56Z",
  "level": "info",
  "msg": "player connected",
  "component": "server",
  "playerID": "player_abc123",
  "seed": 12345
}
```

### Log Aggregation

**ELK Stack:**
```ruby
# /etc/logstash/conf.d/venture.conf
input {
  file {
    path => "/var/log/venture/server.log"
    codec => "json"
    type => "venture-server"
  }
}
output {
  elasticsearch {
    hosts => ["localhost:9200"]
    index => "venture-server-%{+YYYY.MM.dd}"
  }
}
```

**CloudWatch:** Install agent, configure log stream `/venture/server`, format JSON
**Datadog:** Install agent, configure `/etc/datadog-agent/conf.d/venture.d/conf.yaml`

### Key Metrics

| Metric | Type | Alert Threshold |
|--------|------|-----------------|
| `server.tick_time` | Gauge | >50ms |
| `server.player_count` | Gauge | >max_players |
| `network.bandwidth_out` | Counter | >1 Mbps/player |
| `memory.usage` | Gauge | >80% |
| `cpu.usage` | Gauge | >80% |

### Health Checks

```bash
# Basic TCP check
nc -zv <server-ip> 8080

# NGINX upstream health
upstream venture_servers {
  server 10.0.1.10:8080 max_fails=3 fail_timeout=30s;
}
```

---

## Security Best Practices

### Network Security

**Firewall:**
```bash
# UFW
sudo ufw allow 8080/tcp && sudo ufw enable

# iptables
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
```

**Rate Limiting:**
```bash
# Limit connections to 10/min per IP
sudo iptables -A INPUT -p tcp --dport 8080 -m state --state NEW -m recent --set
sudo iptables -A INPUT -p tcp --dport 8080 -m state --state NEW -m recent --update --seconds 60 --hitcount 10 -j DROP
```

**DDoS Protection:** Use Cloudflare Spectrum, AWS Shield, or fail2ban

### Application Security

**Run as Non-Root:**
```bash
sudo useradd -r -s /bin/false venture
sudo -u venture ./venture-server -port 8080
```

**Resource Limits:**
```bash
ulimit -n 65536      # Max open files
ulimit -v 2097152    # Max memory (2GB)
```

**Production Mode:**
```bash
export LOG_LEVEL=info  # Never use -verbose in production
```

### Data Security

**Random Seeds:**
```bash
SEED=$(openssl rand -hex 8 | xargs printf "%d")
./venture-server -seed $SEED
```

**Save File Protection:**
```bash
chmod 600 /opt/venture/saves/*.json
chown venture:venture /opt/venture/saves/*.json
```

**Encryption:** Use VPN (WireGuard/OpenVPN) for encrypted tunnels or NGINX SSL termination for WebSocket support

---

## Scaling Strategies

### Vertical Scaling

| Players | CPU Cores | RAM | Network |
|---------|-----------|-----|---------|
| 1-4 | 2 | 1 GB | 1 Mbps |
| 5-10 | 4 | 2 GB | 2 Mbps |
| 11-20 | 8 | 4 GB | 5 Mbps |
| 21-50 | 16 | 8 GB | 10 Mbps |

**Pros:** Simple, low latency  
**Cons:** Hardware limits (~50 players), single point of failure

### Horizontal Scaling

```bash
for port in {8080..8089}; do
  ./venture-server -port $port -seed $((12345 + port)) -max-players 4 &
done
```

**Pros:** Linear scaling, cost-effective, fault isolation  
**Cons:** Players can't interact across servers, manual load distribution

### Geographic Distribution

| Region | Location | Purpose |
|--------|----------|---------|
| `us-east-1` | Virginia | North America |
| `eu-west-1` | Ireland | Europe |
| `ap-southeast-1` | Singapore | Asia |
| `sa-east-1` | São Paulo | South America |

---

## Performance Tuning

### Server Optimization

**Tick Rate:**
```bash
./venture-server -tick-rate 60  # LAN low-latency
./venture-server -tick-rate 20  # Balanced (default)
./venture-server -tick-rate 10  # High-latency/bandwidth-constrained
```

**GC Tuning:**
```bash
export GOGC=50   # Aggressive GC (low memory)
export GOGC=200  # Relaxed GC (high memory)
```

### Network Optimization

**TCP Tuning (Linux):**
```bash
sudo sysctl -w net.core.rmem_max=16777216 net.core.wmem_max=16777216
sudo sysctl -w net.ipv4.tcp_rmem="4096 87380 16777216"
sudo sysctl -w net.ipv4.tcp_fastopen=3
```

**File Descriptors:**
```bash
ulimit -n 65536
```

### Profiling

**CPU:**
```bash
./venture-server -cpuprofile=cpu.prof -port 8080 &
# Generate load, then: kill $SERVER_PID
go tool pprof cpu.prof
```

**Memory:**
```bash
./venture-server -memprofile=mem.prof -port 8080 &
go tool pprof mem.prof
```

---

## Backup & Recovery

### Save File Backup

```bash
#!/bin/bash
BACKUP_DIR="/backup/venture"
SAVE_DIR="/opt/venture/saves"
DATE=$(date +%Y%m%d)
mkdir -p $BACKUP_DIR
tar -czf $BACKUP_DIR/saves-$DATE.tar.gz -C $SAVE_DIR .
find $BACKUP_DIR -name "saves-*.tar.gz" -mtime +30 -delete
```

**Cron:** `0 3 * * * /opt/venture/scripts/backup-saves.sh`

### Disaster Recovery

**RTO:** <5 minutes | **RPO:** Last save (<1 minute)

```bash
sudo systemctl stop venture-server
tar -xzf /backup/venture/saves-20251029.tar.gz -C /opt/venture/saves/
sudo systemctl start venture-server
```

### High Availability

Use Keepalived for virtual IP failover between primary/secondary servers with shared storage.

---

## Troubleshooting

### Common Issues

**Port Already in Use:**
```bash
sudo lsof -i :8080
sudo kill $(sudo lsof -t -i:8080)
```

**High Latency:**
```bash
# Check server tick time
grep "tick_time" /var/log/venture/server.log

# Reduce tick rate
./venture-server -tick-rate 10

# Check network path
traceroute <server-ip>
```

**Memory Leak:**
```bash
# Monitor memory
watch -n 5 'ps -p $(pgrep venture-server) -o %mem,rss'

# Tune GC
export GOGC=50

# Restart periodically
crontab -e: 0 4 * * * systemctl restart venture-server
```

**Connection Issues:**
```bash
# Test connectivity
nc -zv <server-ip> 8080

# Check firewall
sudo ufw allow 8080/tcp
```

### Debug Mode

```bash
# Verbose logging
./venture-server -verbose -port 8080
# Or
export LOG_LEVEL=debug
./venture-server -port 8080
```

---

## CI/CD Integration

See [CI_CD.md](CI_CD.md) for complete workflow documentation.

**Key Workflows:**
- `build.yml`: CI builds (push to main)
- `release.yml`: GitHub releases (tag push)
- `pages.yml`: WASM deployment (push to main)

**Blue-Green Deployment:**
```bash
scp venture-server user@blue-server:/opt/venture/
ssh user@blue-server 'sudo systemctl restart venture-server'
if nc -zv blue-server 8080; then
  # Switch traffic to blue, deregister green
  echo "Deployment successful"
fi
```

---

## Additional Resources

**Documentation:** [ARCHITECTURE.md](ARCHITECTURE.md), [DEVELOPMENT.md](DEVELOPMENT.md), [USER_MANUAL.md](USER_MANUAL.md), [API_REFERENCE.md](API_REFERENCE.md), [PERFORMANCE.md](PERFORMANCE.md)

**Community:** [GitHub Issues](https://github.com/opd-ai/venture/issues), [Discussions](https://github.com/opd-ai/venture/discussions)

---

## Quick Reference

**Start Server:**
```bash
./venture-server -port 8080 -max-players 10 -tick-rate 20 -seed 12345
sudo systemctl start venture-server
```

**Monitoring:**
```bash
top -p $(pgrep venture-server)
netstat -an | grep :8080
tail -f /var/log/venture/server.log | jq .
```

**Emergency Procedures:**
```bash
# Server not responding
pgrep venture-server && sudo journalctl -u venture-server -n 100
sudo systemctl restart venture-server

# Maintenance mode
sudo iptables -A INPUT -p tcp --dport 8080 -j REJECT
sudo systemctl stop venture-server
# Perform maintenance
sudo systemctl start venture-server
sudo iptables -D INPUT -p tcp --dport 8080 -j REJECT
```

---

**Document Version:** 1.0  
**Last Reviewed:** October 29, 2025  
**Maintained By:** Venture Development Team
