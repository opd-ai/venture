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

Venture is a fully procedural multiplayer action-RPG designed for production deployment with high-latency tolerance (200-5000ms) and cross-platform support. This guide covers deploying the dedicated server for persistent multiplayer gameplay.

**Key Features:**
- Authoritative server architecture for multiplayer
- Client-side prediction and lag compensation
- Deterministic procedural generation (seed-based)
- Structured JSON logging for log aggregation
- Support for 2-4+ concurrent players per server
- Memory-efficient design (<1GB per 4 players)

**Target Environments:**
- Cloud VMs (AWS, GCP, Azure, DigitalOcean)
- Bare metal servers
- Container platforms (Docker, Kubernetes)
- Edge locations (Tor onion services supported)

---

## System Requirements

### Minimum Server Requirements

**Hardware:**
- **CPU:** 2 cores @ 2.0 GHz (x86_64 or ARM64)
- **RAM:** 2 GB (512 MB per player recommended)
- **Disk:** 100 MB for binaries, 10 MB per save file
- **Network:** 1 Mbps upstream (250 KB/s per player)

**Operating Systems:**
- Linux (Ubuntu 20.04+, Debian 11+, RHEL 8+)
- macOS 11+ (Big Sur or later)
- Windows Server 2019+

**Dependencies:**
- None! Static binaries with no runtime dependencies

### Recommended Production Requirements

**Hardware:**
- **CPU:** 4 cores @ 3.0 GHz
- **RAM:** 4 GB (1 GB per player for headroom)
- **Disk:** 500 MB (logs, saves, binaries)
- **Network:** 10 Mbps upstream with low latency (<50ms to players)

**Performance Targets:**
- 20 TPS (ticks per second) server update rate
- <100 KB/s bandwidth per player
- <50ms server tick time (99th percentile)
- <500 MB memory usage per 4 players

---

## Deployment Architectures

### Architecture 1: Single Server (Simple)

**Best for:** Small groups, LAN parties, testing

```
┌─────────────┐
│   Server    │ :8080
│  (single    │
│  instance)  │
└──────┬──────┘
       │
   ┌───┴───┬───────┬───────┐
   │       │       │       │
Client1 Client2 Client3 Client4
```

**Characteristics:**
- Single point of failure
- No load balancing
- Simple to manage
- Cost-effective
- Suitable for <10 concurrent players

**Setup:**
```bash
# Start dedicated server
./venture-server -port 8080 -max-players 10 -tick-rate 20
```

### Architecture 2: Multi-Server (Sharded)

**Best for:** Multiple independent games, horizontal scaling

```
┌────────────┐  ┌────────────┐  ┌────────────┐
│  Server 1  │  │  Server 2  │  │  Server 3  │
│   :8080    │  │   :8081    │  │   :8082    │
│ (Seed A)   │  │ (Seed B)   │  │ (Seed C)   │
└─────┬──────┘  └─────┬──────┘  └─────┬──────┘
      │               │               │
  Players 1-4     Players 5-8     Players 9-12
```

**Characteristics:**
- Independent worlds (different seeds)
- Linear scaling (add more servers)
- Load distribution across instances
- Isolated failures
- Suitable for 10-100+ concurrent players

**Setup:**
```bash
# Start multiple servers on different ports
./venture-server -port 8080 -seed 12345 -max-players 4 &
./venture-server -port 8081 -seed 67890 -max-players 4 &
./venture-server -port 8082 -seed 11111 -max-players 4 &
```

### Architecture 3: Cloud Deployment (Managed)

**Best for:** Production services, global distribution, high availability

```
┌─────────────────────────────────────┐
│        Load Balancer (Optional)     │
│    (Round-robin new connections)    │
└───────────────┬─────────────────────┘
                │
        ┌───────┴─────────┬──────────┐
        │                 │          │
   ┌────▼────┐      ┌────▼────┐ ┌──▼─────┐
   │ Server  │      │ Server  │ │ Server │
   │  Pod 1  │      │  Pod 2  │ │  Pod 3 │
   │ :8080   │      │ :8080   │ │ :8080  │
   └────┬────┘      └────┬────┘ └───┬────┘
        │                │          │
   ┌────▼────────────────▼──────────▼────┐
   │     Centralized Logging/Metrics     │
   │   (ELK, Prometheus, CloudWatch)     │
   └─────────────────────────────────────┘
```

**Characteristics:**
- Auto-scaling based on demand
- Health checks and automatic recovery
- Centralized logging and monitoring
- Geographic distribution
- Suitable for 100+ concurrent players

**Technologies:**
- **Orchestration:** Kubernetes, Docker Swarm, AWS ECS
- **Logging:** ELK Stack, CloudWatch, Datadog
- **Metrics:** Prometheus + Grafana
- **Load Balancing:** NGINX, HAProxy, AWS ALB

---

## Server Setup

### Quick Start (5 Minutes)

#### 1. Download Release Binary

**Option A: From GitHub Releases**
```bash
# Download latest release (replace VERSION with actual version)
VERSION=v1.1.0
wget https://github.com/opd-ai/venture/releases/download/${VERSION}/venture-server-linux-amd64.tar.gz

# Extract
tar -xzf venture-server-linux-amd64.tar.gz
cd venture-server-linux-amd64

# Run server
./venture-server -port 8080
```

**Option B: Build from Source**
```bash
# Clone repository
git clone https://github.com/opd-ai/venture.git
cd venture

# Build server
go build -ldflags="-s -w" -o venture-server ./cmd/server

# Run server
./venture-server -port 8080
```

#### 2. Verify Server Started

```bash
# Check server logs (should see "Starting Venture Game Server")
# Listen for incoming connections on :8080

# Test connectivity
nc -zv localhost 8080
# Should output: Connection to localhost 8080 port [tcp/*] succeeded!
```

#### 3. Connect Client

```bash
# From another terminal or machine
./venture-client -multiplayer -server localhost:8080
```

### Systemd Service (Linux Production)

Create a systemd service for automatic startup and management:

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
ExecStart=/opt/venture/venture-server -port 8080 -max-players 10 -tick-rate 20 -seed 12345 -genre fantasy
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=venture-server

# Environment variables
Environment="LOG_LEVEL=info"
Environment="LOG_FORMAT=json"

# Resource limits
LimitNOFILE=65536
MemoryLimit=2G
CPUQuota=400%

[Install]
WantedBy=multi-user.target
```

**Setup Steps:**

```bash
# Create service user
sudo useradd -r -s /bin/false venture

# Install server binary
sudo mkdir -p /opt/venture
sudo cp venture-server /opt/venture/
sudo chown -R venture:venture /opt/venture
sudo chmod +x /opt/venture/venture-server

# Install systemd service
sudo cp venture-server.service /etc/systemd/system/
sudo systemctl daemon-reload

# Enable and start service
sudo systemctl enable venture-server
sudo systemctl start venture-server

# Check status
sudo systemctl status venture-server

# View logs
sudo journalctl -u venture-server -f
```

**Service Management:**

```bash
# Start server
sudo systemctl start venture-server

# Stop server
sudo systemctl stop venture-server

# Restart server
sudo systemctl restart venture-server

# View logs (last 100 lines)
sudo journalctl -u venture-server -n 100

# Follow logs in real-time
sudo journalctl -u venture-server -f

# Check service status
sudo systemctl status venture-server
```

### Docker Deployment

**Dockerfile:**

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o venture-server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN addgroup -S venture && adduser -S venture -G venture

WORKDIR /app
COPY --from=builder /build/venture-server .

USER venture
EXPOSE 8080

ENTRYPOINT ["./venture-server"]
CMD ["-port", "8080", "-max-players", "10", "-tick-rate", "20"]
```

**Build and Run:**

```bash
# Build image
docker build -t venture-server:latest .

# Run container
docker run -d \
  --name venture-server \
  -p 8080:8080 \
  -e LOG_LEVEL=info \
  -e LOG_FORMAT=json \
  --restart unless-stopped \
  venture-server:latest

# View logs
docker logs -f venture-server

# Stop container
docker stop venture-server

# Remove container
docker rm venture-server
```

**Docker Compose:**

**File:** `docker-compose.yml`

```yaml
version: '3.8'

services:
  venture-server-1:
    image: venture-server:latest
    build: .
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=info
      - LOG_FORMAT=json
    command: ["-port", "8080", "-seed", "12345", "-max-players", "4"]
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "3"
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
        reservations:
          cpus: '1.0'
          memory: 512M

  venture-server-2:
    image: venture-server:latest
    ports:
      - "8081:8080"
    environment:
      - LOG_LEVEL=info
      - LOG_FORMAT=json
    command: ["-port", "8080", "-seed", "67890", "-max-players", "4"]
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "3"
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
```

**Usage:**

```bash
# Start all servers
docker-compose up -d

# View logs from all services
docker-compose logs -f

# Scale servers (add more instances)
docker-compose up -d --scale venture-server-1=3

# Stop all servers
docker-compose down
```

### Kubernetes Deployment

**File:** `k8s-deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: venture-server
  labels:
    app: venture-server
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
          protocol: TCP
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: LOG_FORMAT
          value: "json"
        args:
        - "-port"
        - "8080"
        - "-max-players"
        - "4"
        - "-tick-rate"
        - "20"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "2000m"
        livenessProbe:
          tcpSocket:
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          tcpSocket:
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
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
    targetPort: 8080
```

**Deploy:**

```bash
# Apply configuration
kubectl apply -f k8s-deployment.yaml

# Check deployment status
kubectl get deployments
kubectl get pods

# View logs
kubectl logs -f deployment/venture-server

# Scale deployment
kubectl scale deployment venture-server --replicas=5

# Get service IP
kubectl get service venture-server
```

---

## Configuration

### Command-Line Flags

The server supports comprehensive configuration via flags:

```bash
./venture-server \
  -port 8080 \              # Server port (default: 8080)
  -max-players 10 \         # Maximum concurrent players (default: 4)
  -tick-rate 20 \           # Server update rate in Hz (default: 20)
  -seed 12345 \             # World generation seed (default: random)
  -genre fantasy \          # Genre: fantasy, scifi, horror, cyberpunk, postapoc (default: fantasy)
  -verbose                  # Enable verbose debug logging (default: false)
```

**Flag Reference:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-port` | string | `"8080"` | Server listen port |
| `-max-players` | int | `4` | Maximum concurrent players |
| `-tick-rate` | int | `20` | Server update rate (ticks per second) |
| `-seed` | int64 | random | World generation seed (deterministic) |
| `-genre` | string | `"fantasy"` | World genre theme |
| `-verbose` | bool | `false` | Enable debug logging |
| `-aerial-sprites` | bool | `true` | Use top-down sprite perspective |

### Environment Variables

Override configuration via environment variables:

```bash
# Logging configuration
export LOG_LEVEL=info        # debug, info, warn, error, fatal
export LOG_FORMAT=json       # json or text

# Run server
./venture-server -port 8080
```

**Environment Variable Reference:**

| Variable | Values | Description |
|----------|--------|-------------|
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error`, `fatal` | Logging verbosity |
| `LOG_FORMAT` | `json`, `text` | Log output format |

### Genre Configuration

Each genre provides a distinct theme and aesthetic:

| Genre ID | Theme | Characteristics |
|----------|-------|-----------------|
| `fantasy` | Medieval Fantasy | Magic, dungeons, dragons, medieval weapons |
| `scifi` | Science Fiction | Technology, lasers, robots, space stations |
| `horror` | Horror/Gothic | Dark atmosphere, monsters, limited visibility |
| `cyberpunk` | Cyberpunk | Neon cities, hackers, corporations, tech noir |
| `postapoc` | Post-Apocalyptic | Wasteland, survivors, scavenged equipment |

**Usage:**
```bash
# Fantasy theme (default)
./venture-server -genre fantasy

# Sci-fi theme
./venture-server -genre scifi

# Horror theme with low lighting
./venture-server -genre horror
```

### Performance Tuning Configuration

**Tick Rate Selection:**

| Tick Rate | Use Case | CPU Impact | Network Impact |
|-----------|----------|------------|----------------|
| 10 Hz | Low bandwidth, high latency | Low | Minimal |
| 20 Hz (default) | Balanced gameplay | Medium | Moderate |
| 30 Hz | Competitive, low latency | High | High |
| 60 Hz | LAN only, <20ms latency | Very High | Very High |

**Recommendations:**
- **LAN/Low Latency:** 30-60 Hz for responsive gameplay
- **Internet/WAN:** 20 Hz for balanced performance
- **High Latency (>200ms):** 10-20 Hz to reduce bandwidth

```bash
# Low bandwidth mode (10 Hz)
./venture-server -tick-rate 10 -max-players 8

# Competitive mode (30 Hz)
./venture-server -tick-rate 30 -max-players 4
```

---

## Monitoring & Logging

### Structured Logging

Venture uses **structured JSON logging** via logrus for production deployments. All logs include:
- Timestamp (ISO 8601)
- Log level
- Component/system context
- Structured fields (key-value pairs)

**Log Levels:**

| Level | Usage | Example |
|-------|-------|---------|
| `debug` | Development, detailed diagnostics | `"seed": 12345, "roomCount": 25` |
| `info` | Normal operations, lifecycle events | `"server started", "player connected"` |
| `warn` | Non-fatal issues, degraded performance | `"high latency detected: 500ms"` |
| `error` | Operation failures, retry scenarios | `"terrain generation failed"` |
| `fatal` | Critical failures, server shutdown | `"failed to bind port 8080"` |

**Example Log Output (JSON):**

```json
{
  "time": "2025-10-29T12:34:56Z",
  "level": "info",
  "msg": "player connected",
  "component": "server",
  "playerID": "player_abc123",
  "seed": 12345,
  "genre": "fantasy"
}
```

### Log Aggregation

**ELK Stack (Elasticsearch, Logstash, Kibana):**

1. **Configure Logstash Input:**

**File:** `/etc/logstash/conf.d/venture.conf`

```ruby
input {
  file {
    path => "/var/log/venture/server.log"
    codec => "json"
    type => "venture-server"
  }
}

filter {
  # Add hostname
  mutate {
    add_field => { "hostname" => "%{host}" }
  }
}

output {
  elasticsearch {
    hosts => ["localhost:9200"]
    index => "venture-server-%{+YYYY.MM.dd}"
  }
}
```

2. **Query Logs in Kibana:**

```
# Search for errors
level: "error"

# Search by player
playerID: "player_abc123"

# Search by component
component: "server" AND msg: "connected"
```

**CloudWatch (AWS):**

```bash
# Install CloudWatch agent
sudo yum install amazon-cloudwatch-agent

# Configure log stream
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-config-wizard

# Select:
# - Log file: /var/log/venture/server.log
# - Log group: /venture/server
# - Format: JSON
```

**Datadog:**

```bash
# Install Datadog agent
DD_API_KEY=<your-key> DD_SITE="datadoghq.com" bash -c "$(curl -L https://s3.amazonaws.com/dd-agent/scripts/install_script.sh)"

# Configure log collection
sudo vim /etc/datadog-agent/conf.d/venture.d/conf.yaml
```

**File:** `/etc/datadog-agent/conf.d/venture.d/conf.yaml`

```yaml
logs:
  - type: file
    path: /var/log/venture/server.log
    service: venture-server
    source: go
    sourcecategory: gamelog
```

### Metrics Collection

**Key Metrics to Monitor:**

| Metric | Type | Alert Threshold | Description |
|--------|------|-----------------|-------------|
| `server.tick_time` | Gauge | >50ms | Server update duration |
| `server.player_count` | Gauge | >max_players | Active player count |
| `network.bandwidth_out` | Counter | >1 Mbps/player | Outbound network traffic |
| `memory.usage` | Gauge | >80% limit | Server memory consumption |
| `cpu.usage` | Gauge | >80% | CPU utilization |
| `errors.rate` | Counter | >10/min | Error occurrence rate |

**Prometheus Integration (Future Enhancement):**

The server does not currently expose Prometheus metrics. To add instrumentation:

```go
// Example: Add to cmd/server/main.go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// Expose metrics endpoint
http.Handle("/metrics", promhttp.Handler())
go http.ListenAndServe(":9090", nil)
```

**Manual Monitoring:**

```bash
# Monitor server logs for errors
tail -f /var/log/venture/server.log | grep '"level":"error"'

# Monitor network connections
watch -n 5 'netstat -an | grep :8080 | wc -l'

# Monitor memory usage
watch -n 5 'ps aux | grep venture-server'

# Monitor CPU usage
top -p $(pgrep venture-server)
```

### Health Checks

**Basic TCP Check:**

```bash
# Check if server port is listening
nc -zv <server-ip> 8080

# Continuous monitoring
while true; do
  nc -zv <server-ip> 8080 || echo "Server down!"
  sleep 30
done
```

**Load Balancer Health Check:**

For load balancers (HAProxy, NGINX, AWS ALB), configure TCP health checks:

```nginx
# NGINX upstream health check
upstream venture_servers {
  server 10.0.1.10:8080 max_fails=3 fail_timeout=30s;
  server 10.0.1.11:8080 max_fails=3 fail_timeout=30s;
  server 10.0.1.12:8080 max_fails=3 fail_timeout=30s;
}

server {
  listen 80;
  location / {
    proxy_pass http://venture_servers;
  }
}
```

---

## Security Best Practices

### Network Security

**1. Firewall Configuration:**

```bash
# UFW (Ubuntu/Debian)
sudo ufw allow 8080/tcp comment 'Venture game server'
sudo ufw enable

# iptables
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
sudo iptables-save > /etc/iptables/rules.v4

# firewalld (RHEL/CentOS)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

**2. Rate Limiting:**

Protect against connection floods using iptables:

```bash
# Limit new connections to 10 per minute per IP
sudo iptables -A INPUT -p tcp --dport 8080 -m state --state NEW -m recent --set
sudo iptables -A INPUT -p tcp --dport 8080 -m state --state NEW -m recent --update --seconds 60 --hitcount 10 -j DROP
```

**3. DDoS Protection:**

For production deployments, use:
- **Cloudflare:** Spectrum for TCP load balancing and DDoS mitigation
- **AWS Shield:** Integrated DDoS protection for EC2/ECS
- **fail2ban:** Automatic IP banning for repeated connection failures

```bash
# Install fail2ban
sudo apt-get install fail2ban

# Configure Venture jail
sudo vim /etc/fail2ban/jail.d/venture.conf
```

**File:** `/etc/fail2ban/jail.d/venture.conf`

```ini
[venture]
enabled = true
port = 8080
filter = venture
logpath = /var/log/venture/server.log
maxretry = 5
bantime = 3600
findtime = 600
```

### Application Security

**1. Run as Non-Root User:**

Always run the server as a dedicated user with minimal privileges:

```bash
# Create service user
sudo useradd -r -s /bin/false venture

# Run server as venture user
sudo -u venture ./venture-server -port 8080
```

**2. Resource Limits:**

Prevent resource exhaustion:

```bash
# Set ulimits
ulimit -n 65536      # Max open files
ulimit -u 4096       # Max processes
ulimit -v 2097152    # Max virtual memory (2GB)

# Permanent limits in /etc/security/limits.conf
venture soft nofile 65536
venture hard nofile 65536
venture soft nproc 4096
venture hard nproc 4096
```

**3. Disable Debug Logging in Production:**

```bash
# Production mode (info logging only)
export LOG_LEVEL=info
./venture-server -port 8080

# Never use -verbose flag in production
# (exposes sensitive internal state)
```

### Data Security

**1. World Seed Management:**

World seeds are **not** cryptographic keys, but predictable seeds can allow world replay:

```bash
# Use cryptographically random seeds for production
SEED=$(openssl rand -hex 8 | awk '{print "0x" $1}' | xargs printf "%d")
./venture-server -seed $SEED
```

**2. Save File Protection:**

Save files contain player data and should be protected:

```bash
# Restrict save file permissions
chmod 600 /opt/venture/saves/*.json
chown venture:venture /opt/venture/saves/*.json

# Encrypt saves at rest (optional)
# Use filesystem encryption (LUKS, eCryptfs) or application-level encryption
```

### SSL/TLS Considerations

Venture currently uses **unencrypted TCP**. For secure communication:

**Option 1: VPN Tunnel (Recommended)**

Use WireGuard or OpenVPN to create encrypted tunnels:

```bash
# WireGuard setup (clients connect via VPN)
# Server and clients communicate over encrypted tunnel
# Venture traffic is automatically encrypted
```

**Option 2: Reverse Proxy with TLS (Future)**

If Venture adds WebSocket support:

```nginx
# NGINX with SSL termination
server {
  listen 443 ssl;
  ssl_certificate /etc/ssl/certs/venture.crt;
  ssl_certificate_key /etc/ssl/private/venture.key;
  
  location / {
    proxy_pass http://localhost:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
  }
}
```

---

## Scaling Strategies

### Vertical Scaling (Scale Up)

**Increase resources on a single server:**

| Players | CPU Cores | RAM | Network |
|---------|-----------|-----|---------|
| 1-4 | 2 | 1 GB | 1 Mbps |
| 5-10 | 4 | 2 GB | 2 Mbps |
| 11-20 | 8 | 4 GB | 5 Mbps |
| 21-50 | 16 | 8 GB | 10 Mbps |

**Advantages:**
- Simple configuration
- No distributed state management
- Lower latency (single network hop)

**Limitations:**
- Hardware limits (~50 concurrent players)
- Single point of failure
- Expensive high-end hardware

### Horizontal Scaling (Scale Out)

**Run multiple independent servers:**

```bash
# Start multiple servers with different seeds
for port in {8080..8089}; do
  seed=$((12345 + port))
  ./venture-server -port $port -seed $seed -max-players 4 &
done
```

**Advantages:**
- Linear scaling (add more servers)
- Cost-effective (commodity hardware)
- Fault isolation (one server failure doesn't affect others)

**Limitations:**
- Players can't interact across servers (different worlds)
- Manual load distribution
- More complex management

### Auto-Scaling (Cloud)

**AWS Auto Scaling Group Example:**

```json
{
  "AutoScalingGroupName": "venture-servers",
  "LaunchTemplate": {
    "LaunchTemplateId": "lt-1234567890abcdef",
    "Version": "$Latest"
  },
  "MinSize": 2,
  "MaxSize": 10,
  "DesiredCapacity": 3,
  "TargetGroupARNs": ["arn:aws:elasticloadbalancing:..."],
  "HealthCheckType": "ELB",
  "HealthCheckGracePeriod": 300
}
```

**Scaling Policies:**

```json
{
  "PolicyName": "scale-up-on-cpu",
  "PolicyType": "TargetTrackingScaling",
  "TargetTrackingConfiguration": {
    "PredefinedMetricSpecification": {
      "PredefinedMetricType": "ASGAverageCPUUtilization"
    },
    "TargetValue": 70.0
  }
}
```

### Geographic Distribution

Deploy servers in multiple regions for lower latency:

| Region | Location | Purpose |
|--------|----------|---------|
| `us-east-1` | Virginia, USA | North American players |
| `eu-west-1` | Ireland | European players |
| `ap-southeast-1` | Singapore | Asian players |
| `sa-east-1` | São Paulo | South American players |

**Client Connection Logic:**

```bash
# Client selects nearest server based on latency
# (implement simple ping test or use GeoDNS)

# Example: ping-based server selection
for server in us.venture.example.com eu.venture.example.com ap.venture.example.com; do
  latency=$(ping -c 3 $server | tail -1 | awk '{print $4}' | cut -d '/' -f 2)
  echo "$server: ${latency}ms"
done
```

---

## Performance Tuning

### Server Optimization

**1. Tick Rate Tuning:**

```bash
# Low-latency LAN (minimize input lag)
./venture-server -tick-rate 60

# Balanced (default, recommended)
./venture-server -tick-rate 20

# High-latency or bandwidth-constrained
./venture-server -tick-rate 10
```

**Performance Impact:**

| Tick Rate | CPU Usage | Network Usage | Input Lag |
|-----------|-----------|---------------|-----------|
| 10 Hz | 50% | 50 KB/s/player | 100ms |
| 20 Hz | 100% | 100 KB/s/player | 50ms |
| 60 Hz | 300% | 300 KB/s/player | 16ms |

**2. Player Limit Optimization:**

```bash
# Measure performance under load
./venture-server -max-players 20 &
SERVER_PID=$!

# Monitor CPU and memory
watch -n 1 'ps -p $SERVER_PID -o %cpu,%mem'

# If CPU >80%, reduce player limit
# If memory >80%, reduce player limit
```

**3. Memory Management:**

Go's garbage collector is automatic, but can be tuned:

```bash
# Set GC target percentage (default 100)
# Higher = less frequent GC, more memory
# Lower = more frequent GC, less memory
export GOGC=50  # Aggressive GC (low memory systems)
./venture-server -port 8080

export GOGC=200  # Relaxed GC (high memory systems)
./venture-server -port 8080
```

### Network Optimization

**1. TCP Tuning (Linux):**

```bash
# Increase TCP buffer sizes for high-throughput
sudo sysctl -w net.core.rmem_max=16777216
sudo sysctl -w net.core.wmem_max=16777216
sudo sysctl -w net.ipv4.tcp_rmem="4096 87380 16777216"
sudo sysctl -w net.ipv4.tcp_wmem="4096 65536 16777216"

# Enable TCP fast open
sudo sysctl -w net.ipv4.tcp_fastopen=3

# Persist changes
sudo vim /etc/sysctl.conf
# Add the above settings, then:
sudo sysctl -p
```

**2. Connection Pooling:**

The server maintains a connection pool. For high player counts:

```bash
# Increase file descriptor limits
ulimit -n 65536

# Verify limits
cat /proc/$(pgrep venture-server)/limits | grep "open files"
```

**3. Bandwidth Optimization:**

Current bandwidth usage: ~100 KB/s per player at 20 Hz

To reduce bandwidth:
- Lower tick rate: `-tick-rate 10` (50% reduction)
- Reduce update precision (future: delta compression)
- Spatial culling (only sync nearby entities - already implemented)

### Profiling

**CPU Profiling:**

```bash
# Build with profiling support
go build -o venture-server ./cmd/server

# Run with CPU profiling
./venture-server -cpuprofile=cpu.prof -port 8080 &
SERVER_PID=$!

# Generate load (run for 5 minutes)
# ...

# Stop server and analyze
kill $SERVER_PID
go tool pprof cpu.prof
# (pprof) top20
# (pprof) list <function-name>
# (pprof) web  # Requires graphviz
```

**Memory Profiling:**

```bash
# Run with memory profiling
./venture-server -memprofile=mem.prof -port 8080 &
SERVER_PID=$!

# Generate load
# ...

# Stop and analyze
kill $SERVER_PID
go tool pprof mem.prof
# (pprof) top20 -alloc_space
# (pprof) list <function-name>
```

**Trace Analysis:**

```bash
# Capture execution trace
GODEBUG=gctrace=1 ./venture-server -port 8080 2>&1 | tee trace.log

# Analyze GC pauses
grep "gc " trace.log | awk '{print $6}' | sort -n | tail -20
```

---

## Backup & Recovery

### Save File Management

**Save File Location:**

Default save location: `./saves/` (relative to working directory)

**Backup Strategy:**

```bash
# Daily backup with rotation
#!/bin/bash
BACKUP_DIR="/backup/venture"
SAVE_DIR="/opt/venture/saves"
DATE=$(date +%Y%m%d)

# Create backup
mkdir -p $BACKUP_DIR
tar -czf $BACKUP_DIR/saves-$DATE.tar.gz -C $SAVE_DIR .

# Retain last 30 days
find $BACKUP_DIR -name "saves-*.tar.gz" -mtime +30 -delete

# Upload to S3 (optional)
aws s3 cp $BACKUP_DIR/saves-$DATE.tar.gz s3://my-bucket/venture/saves/
```

**Automated Backup with Cron:**

```bash
# Add to crontab
crontab -e

# Daily backup at 3 AM
0 3 * * * /opt/venture/scripts/backup-saves.sh
```

### Disaster Recovery

**Recovery Time Objective (RTO):** < 5 minutes  
**Recovery Point Objective (RPO):** Last save (typically <1 minute)

**Recovery Procedure:**

1. **Stop Failed Server:**
```bash
sudo systemctl stop venture-server
```

2. **Restore Save Files:**
```bash
# Extract latest backup
tar -xzf /backup/venture/saves-20251029.tar.gz -C /opt/venture/saves/
```

3. **Restart Server:**
```bash
sudo systemctl start venture-server
```

4. **Verify Recovery:**
```bash
# Check logs for successful startup
sudo journalctl -u venture-server -n 50

# Test client connection
./venture-client -multiplayer -server localhost:8080
```

### High Availability Setup

**Active-Passive Failover:**

```
Primary Server (Active) ──┐
                          ├──> Shared Storage (NFS/EBS)
Secondary Server (Standby)─┘
```

**Implementation with Keepalived:**

```bash
# Install keepalived
sudo apt-get install keepalived

# Configure virtual IP failover
sudo vim /etc/keepalived/keepalived.conf
```

**File:** `/etc/keepalived/keepalived.conf` (Primary)

```
vrrp_instance VI_1 {
    state MASTER
    interface eth0
    virtual_router_id 51
    priority 100
    advert_int 1
    
    authentication {
        auth_type PASS
        auth_pass secret123
    }
    
    virtual_ipaddress {
        192.168.1.100
    }
    
    track_script {
        check_venture
    }
}

vrrp_script check_venture {
    script "/usr/local/bin/check-venture-server.sh"
    interval 5
    weight -20
}
```

**Health Check Script:**

**File:** `/usr/local/bin/check-venture-server.sh`

```bash
#!/bin/bash
nc -zv localhost 8080 > /dev/null 2>&1
exit $?
```

---

## Troubleshooting

### Common Issues

#### Issue 1: Server Won't Start - Port Already in Use

**Symptom:**
```
FATAL: failed to start server: listen tcp :8080: bind: address already in use
```

**Diagnosis:**
```bash
# Check what's using the port
sudo lsof -i :8080
sudo netstat -tulpn | grep 8080
```

**Solution:**
```bash
# Kill existing process
sudo kill $(sudo lsof -t -i:8080)

# Or use a different port
./venture-server -port 8081
```

#### Issue 2: High Latency / Lag

**Symptom:** Players report delayed inputs, stuttering

**Diagnosis:**
```bash
# Check server tick time in logs
grep "tick_time" /var/log/venture/server.log | tail -20

# Check network latency
ping <server-ip>

# Check server CPU usage
top -p $(pgrep venture-server)
```

**Solutions:**

1. **Reduce tick rate:**
```bash
./venture-server -tick-rate 10  # Instead of 20
```

2. **Increase server resources:**
```bash
# Add more CPU/RAM to VM
# Or migrate to higher-tier instance
```

3. **Check network path:**
```bash
# Traceroute to identify bottlenecks
traceroute <server-ip>

# Use MTR for continuous monitoring
mtr <server-ip>
```

#### Issue 3: Memory Leak / High Memory Usage

**Symptom:** Server memory usage grows over time

**Diagnosis:**
```bash
# Monitor memory over time
watch -n 5 'ps -p $(pgrep venture-server) -o %mem,rss'

# Check for goroutine leaks
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

**Solutions:**

1. **Restart server periodically:**
```bash
# Add to crontab: restart daily at 4 AM
0 4 * * * systemctl restart venture-server
```

2. **Tune garbage collector:**
```bash
export GOGC=50  # More aggressive GC
./venture-server -port 8080
```

3. **Report issue:**
```bash
# Capture memory profile
./venture-server -memprofile=mem.prof -port 8080 &
# ... run for several hours ...
kill $(pgrep venture-server)

# Attach mem.prof to bug report
```

#### Issue 4: Players Can't Connect

**Symptom:** Clients timeout when connecting

**Diagnosis:**
```bash
# Test from server
nc -zv localhost 8080

# Test from client
nc -zv <server-ip> 8080

# Check firewall
sudo iptables -L -n | grep 8080
sudo ufw status | grep 8080
```

**Solutions:**

1. **Open firewall port:**
```bash
sudo ufw allow 8080/tcp
sudo systemctl restart ufw
```

2. **Check NAT/router forwarding:**
```
Router settings:
External Port: 8080
Internal Port: 8080
Internal IP: <server-local-ip>
Protocol: TCP
```

3. **Verify server is listening:**
```bash
sudo netstat -tulpn | grep venture-server
# Should show: tcp ... 0.0.0.0:8080 ... LISTEN
```

#### Issue 5: Crashes / Segmentation Faults

**Symptom:** Server exits unexpectedly

**Diagnosis:**
```bash
# Check system logs
sudo journalctl -u venture-server --since "1 hour ago"

# Check for core dumps
coredumpctl list
coredumpctl info <dump-id>

# Run with race detector
go run -race ./cmd/server -port 8080
```

**Solutions:**

1. **Update to latest version:**
```bash
# Check current version
./venture-server -version  # (if implemented)

# Download latest release
wget https://github.com/opd-ai/venture/releases/latest/...
```

2. **Report bug with logs:**
```bash
# Capture crash information
sudo journalctl -u venture-server --since "1 day ago" > crash-log.txt
# Attach to GitHub issue
```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Method 1: Command-line flag
./venture-server -verbose -port 8080

# Method 2: Environment variable
export LOG_LEVEL=debug
./venture-server -port 8080

# View detailed logs
tail -f /var/log/venture/server.log | jq .
```

### Performance Debug

```bash
# Enable GC tracing
GODEBUG=gctrace=1 ./venture-server -port 8080 2>&1 | tee gc.log

# Analyze GC pauses
grep "gc " gc.log | awk '{print "GC pause:", $6, "scavenge:", $11}'

# Monitor goroutines
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

---

## CI/CD Integration

### GitHub Actions

See [CI/CD.md](CI_CD.md) for complete workflow documentation.

**Quick Reference:**

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `build.yml` | Push to main | CI builds for all platforms |
| `release.yml` | Tag push (`v*`) | Create GitHub releases |
| `pages.yml` | Push to main | Deploy WASM to GitHub Pages |

**Deployment Workflow Example:**

**File:** `.github/workflows/deploy-production.yml`

```yaml
name: Deploy to Production

on:
  release:
    types: [published]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Download release binary
        run: |
          wget https://github.com/opd-ai/venture/releases/download/${{ github.event.release.tag_name }}/venture-server-linux-amd64.tar.gz
          tar -xzf venture-server-linux-amd64.tar.gz
      
      - name: Deploy to server
        env:
          SSH_KEY: ${{ secrets.DEPLOY_SSH_KEY }}
          SERVER_HOST: ${{ secrets.SERVER_HOST }}
        run: |
          echo "$SSH_KEY" > deploy_key
          chmod 600 deploy_key
          
          scp -i deploy_key venture-server deploy@$SERVER_HOST:/opt/venture/
          ssh -i deploy_key deploy@$SERVER_HOST 'sudo systemctl restart venture-server'
          
          rm deploy_key
```

### Continuous Deployment

**Blue-Green Deployment:**

1. **Deploy new version to standby server**
2. **Health check standby server**
3. **Switch traffic to standby (now active)**
4. **Keep old version as new standby**

```bash
#!/bin/bash
# blue-green-deploy.sh

# Deploy to blue (inactive) environment
scp venture-server user@blue-server:/opt/venture/
ssh user@blue-server 'sudo systemctl restart venture-server'

# Health check
sleep 10
if nc -zv blue-server 8080; then
  echo "Blue server healthy"
  
  # Switch traffic (update load balancer)
  aws elb register-instances-with-load-balancer \
    --load-balancer-name venture-lb \
    --instances i-blue-instance
  
  aws elb deregister-instances-from-load-balancer \
    --load-balancer-name venture-lb \
    --instances i-green-instance
  
  echo "Traffic switched to blue"
else
  echo "Blue server failed health check - rollback"
  exit 1
fi
```

---

## Additional Resources

**Official Documentation:**
- [Architecture Overview](ARCHITECTURE.md)
- [Development Guide](DEVELOPMENT.md)
- [User Manual](USER_MANUAL.md)
- [Getting Started](GETTING_STARTED.md)

**Technical References:**
- [API Reference](API_REFERENCE.md)
- [Performance Guide](PERFORMANCE.md)
- [CI/CD Guide](CI_CD.md)
- [Structured Logging Guide](STRUCTURED_LOGGING_GUIDE.md)

**Community:**
- GitHub Issues: https://github.com/opd-ai/venture/issues
- GitHub Discussions: https://github.com/opd-ai/venture/discussions

---

## Appendix: Quick Reference

### Server Command Cheat Sheet

```bash
# Start server (basic)
./venture-server -port 8080

# Start with custom configuration
./venture-server -port 8080 -max-players 10 -tick-rate 20 -seed 12345 -genre fantasy

# Start with debug logging
./venture-server -verbose -port 8080

# Start as systemd service
sudo systemctl start venture-server

# Stop server
sudo systemctl stop venture-server

# View logs
sudo journalctl -u venture-server -f

# Check server status
systemctl status venture-server
```

### Monitoring Commands

```bash
# Check server is running
ps aux | grep venture-server

# Monitor CPU/memory
top -p $(pgrep venture-server)

# Check network connections
netstat -an | grep :8080

# Monitor bandwidth
iftop -f "port 8080"

# View real-time logs
tail -f /var/log/venture/server.log | jq .
```

### Emergency Procedures

**Server Not Responding:**
```bash
# 1. Check if process is running
pgrep venture-server

# 2. Check logs for errors
sudo journalctl -u venture-server -n 100

# 3. Restart server
sudo systemctl restart venture-server

# 4. If still failing, check system resources
df -h  # Disk space
free -h  # Memory
top  # CPU
```

**Immediate Maintenance Mode:**
```bash
# 1. Stop accepting new connections
sudo iptables -A INPUT -p tcp --dport 8080 -j REJECT

# 2. Wait for existing players to disconnect
watch -n 5 'netstat -an | grep :8080 | grep ESTABLISHED'

# 3. Perform maintenance
sudo systemctl stop venture-server
# ... maintenance tasks ...
sudo systemctl start venture-server

# 4. Re-enable connections
sudo iptables -D INPUT -p tcp --dport 8080 -j REJECT
```

---

**Document Version:** 1.0  
**Last Reviewed:** October 29, 2025  
**Next Review:** January 2026  
**Maintained By:** Venture Development Team
