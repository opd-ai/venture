# Production Deployment Guide

**Version:** 1.2 | **Status:** Production Ready | **Last Updated:** 2026-01-20

## Overview

Procedural multiplayer action-RPG with high-latency tolerance (200-5000ms), deterministic generation, and cross-platform support.

**Features:** Authoritative server, client prediction, lag compensation, structured logging, 2-4+ players, <1GB per 4 players

## Table of Contents

1. [System Requirements](#system-requirements)
2. [Deployment Architectures](#deployment-architectures)
3. [Quick Start](#quick-start)
4. [Configuration](#configuration)
5. [Systemd Service Examples](#systemd-service-examples)
6. [Reverse Proxy Setup](#reverse-proxy-setup)
7. [Monitoring Setup](#monitoring-setup)
8. [Scaling Recommendations](#scaling-recommendations)
9. [Backup & Disaster Recovery](#backup--disaster-recovery)
10. [Security](#security)
11. [Performance Tuning](#performance-tuning)
12. [Troubleshooting](#troubleshooting)
13. [CI/CD](#cicd)

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
-enable-metrics               # Prometheus metrics on :9090
-metrics-port 9090            # Metrics endpoint port
```

**Environment Variables:**
```bash
export VENTURE_PORT=8080
export LOG_LEVEL=info
export LOG_FORMAT=json
./venture-server
```

## Systemd Service Examples

### Ubuntu/Debian

1. Create service user:
```bash
sudo useradd -r -s /usr/sbin/nologin -d /var/lib/venture venture
sudo mkdir -p /var/lib/venture /var/log/venture
sudo chown venture:venture /var/lib/venture /var/log/venture
```

2. Create service file `/etc/systemd/system/venture-server.service`:
```ini
[Unit]
Description=Venture Game Server
Documentation=https://github.com/opd-ai/venture
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=venture
Group=venture
WorkingDirectory=/var/lib/venture
ExecStart=/usr/bin/venture-server \
    -port 8080 \
    -max-players 10 \
    -log-format json \
    -enable-metrics
Restart=on-failure
RestartSec=5s
StandardOutput=append:/var/log/venture/server.log
StandardError=append:/var/log/venture/server.log

# Resource limits
LimitNOFILE=65536
MemoryMax=1G
CPUQuota=200%

# Security hardening
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes
ReadWritePaths=/var/lib/venture /var/log/venture

[Install]
WantedBy=multi-user.target
```

3. Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable venture-server
sudo systemctl start venture-server
sudo systemctl status venture-server
```

### RHEL/CentOS/Fedora

1. Create service user:
```bash
sudo useradd -r -s /sbin/nologin -d /var/lib/venture venture
sudo mkdir -p /var/lib/venture /var/log/venture
sudo chown venture:venture /var/lib/venture /var/log/venture
sudo restorecon -Rv /var/lib/venture /var/log/venture  # SELinux context
```

2. Create service file `/etc/systemd/system/venture-server.service`:
```ini
[Unit]
Description=Venture Game Server
Documentation=https://github.com/opd-ai/venture
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=venture
Group=venture
WorkingDirectory=/var/lib/venture
ExecStart=/usr/bin/venture-server \
    -port 8080 \
    -max-players 10 \
    -log-format json \
    -enable-metrics
Restart=on-failure
RestartSec=5s
StandardOutput=append:/var/log/venture/server.log
StandardError=append:/var/log/venture/server.log

# Resource limits
LimitNOFILE=65536
MemoryMax=1G

# Security hardening (compatible with SELinux)
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=/var/lib/venture /var/log/venture

[Install]
WantedBy=multi-user.target
```

3. Configure firewalld:
```bash
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

4. Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable venture-server
sudo systemctl start venture-server
```

### Log Rotation

Create `/etc/logrotate.d/venture`:
```
/var/log/venture/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    copytruncate
}
```

## Reverse Proxy Setup

### NGINX

Install NGINX:
```bash
# Ubuntu/Debian
sudo apt install nginx

# RHEL/Fedora
sudo dnf install nginx
```

Create `/etc/nginx/sites-available/venture`:
```nginx
upstream venture_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

upstream venture_metrics {
    server 127.0.0.1:9090;
}

server {
    listen 80;
    server_name venture.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name venture.example.com;

    ssl_certificate /etc/letsencrypt/live/venture.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/venture.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;

    # Game server endpoint
    location / {
        proxy_pass http://venture_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts for high-latency support
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
        proxy_connect_timeout 60s;
    }

    # Health endpoints (internal only)
    location /health {
        allow 127.0.0.1;
        allow 10.0.0.0/8;
        deny all;
        proxy_pass http://venture_metrics/health;
    }

    location /metrics {
        allow 127.0.0.1;
        allow 10.0.0.0/8;
        deny all;
        proxy_pass http://venture_metrics/metrics;
    }
}
```

Enable configuration:
```bash
sudo ln -s /etc/nginx/sites-available/venture /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Caddy

Install Caddy:
```bash
# Ubuntu/Debian
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

Create `/etc/caddy/Caddyfile`:
```caddyfile
venture.example.com {
    # Automatic HTTPS with Let's Encrypt
    
    # Game server
    reverse_proxy /* localhost:8080 {
        transport http {
            keepalive 30s
            keepalive_idle_conns 32
        }
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
    }

    # Health endpoints (internal only)
    @internal {
        remote_ip 127.0.0.1 10.0.0.0/8 192.168.0.0/16
    }
    
    handle /health {
        @internal {
            reverse_proxy localhost:9090
        }
        respond "Forbidden" 403
    }

    handle /metrics {
        @internal {
            reverse_proxy localhost:9090
        }
        respond "Forbidden" 403
    }

    # Rate limiting
    rate_limit {
        zone game {
            key {remote_host}
            events 100
            window 1s
        }
    }

    # Logging
    log {
        output file /var/log/caddy/venture.log
        format json
    }
}
```

Start Caddy:
```bash
sudo systemctl enable caddy
sudo systemctl start caddy
```

## Monitoring Setup

### Health Endpoints

The server exposes health endpoints on the metrics port (default: 9090):

```bash
# Basic liveness check
curl http://localhost:9090/health
# Returns: OK

# Readiness check with component validation
curl http://localhost:9090/ready
# Returns: {"status":"ready"} or {"status":"not_ready","failed_checks":["component: error"]}

# Detailed status with metrics
curl http://localhost:9090/status | jq
# Returns comprehensive JSON with uptime, performance, network, game state, and runtime info
```

### Prometheus Setup

1. Install Prometheus:
```bash
# Ubuntu/Debian
sudo apt install prometheus

# Or download from https://prometheus.io/download/
wget https://github.com/prometheus/prometheus/releases/download/v2.48.0/prometheus-2.48.0.linux-amd64.tar.gz
tar xvfz prometheus-*.tar.gz
```

2. Configure Prometheus `/etc/prometheus/prometheus.yml`:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "venture_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'venture'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
    scrape_interval: 10s
    scrape_timeout: 5s

  # Multiple servers
  - job_name: 'venture-cluster'
    static_configs:
      - targets:
        - 'server1.example.com:9090'
        - 'server2.example.com:9090'
        - 'server3.example.com:9090'
```

3. Create alert rules `/etc/prometheus/venture_rules.yml`:
```yaml
groups:
  - name: venture_alerts
    rules:
      # Server down
      - alert: VentureServerDown
        expr: up{job="venture"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Venture server {{ $labels.instance }} is down"
          description: "Server has been unreachable for more than 1 minute"

      # High memory usage
      - alert: VentureHighMemory
        expr: venture_memory_mb > 800
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage on {{ $labels.instance }}"
          description: "Memory usage is {{ $value }}MB (threshold: 800MB)"

      # Low FPS
      - alert: VentureLowFPS
        expr: venture_fps < 30
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Low FPS on {{ $labels.instance }}"
          description: "FPS is {{ $value }} (threshold: 30)"

      # Player count spike
      - alert: VenturePlayerSpike
        expr: venture_players_connected > 8
        for: 5m
        labels:
          severity: info
        annotations:
          summary: "High player count on {{ $labels.instance }}"
          description: "{{ $value }} players connected"
```

4. Start Prometheus:
```bash
sudo systemctl enable prometheus
sudo systemctl start prometheus
```

### Grafana Setup

1. Install Grafana:
```bash
# Ubuntu/Debian
sudo apt-get install -y apt-transport-https software-properties-common
sudo mkdir -p /etc/apt/keyrings/
wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
sudo apt update
sudo apt install grafana
```

2. Start Grafana:
```bash
sudo systemctl enable grafana-server
sudo systemctl start grafana-server
```

3. Access Grafana at `http://localhost:3000` (default: admin/admin)

4. Add Prometheus data source:
   - Go to Configuration → Data Sources → Add data source
   - Select Prometheus
   - URL: `http://localhost:9090`
   - Save & Test

5. Import Venture dashboard:

Create dashboard JSON (`venture-dashboard.json`):
```json
{
  "title": "Venture Game Server",
  "panels": [
    {
      "title": "Players Connected",
      "type": "stat",
      "targets": [{"expr": "venture_players_connected"}]
    },
    {
      "title": "FPS",
      "type": "gauge",
      "targets": [{"expr": "venture_fps"}],
      "options": {"min": 0, "max": 120}
    },
    {
      "title": "Memory Usage (MB)",
      "type": "graph",
      "targets": [{"expr": "venture_memory_mb"}]
    },
    {
      "title": "Network Traffic",
      "type": "graph",
      "targets": [
        {"expr": "rate(venture_bytes_sent[5m])", "legendFormat": "Sent"},
        {"expr": "rate(venture_bytes_received[5m])", "legendFormat": "Received"}
      ]
    },
    {
      "title": "Entity Count",
      "type": "stat",
      "targets": [{"expr": "venture_entity_count"}]
    },
    {
      "title": "Active Quests",
      "type": "stat",
      "targets": [{"expr": "venture_active_quests"}]
    }
  ]
}
```

### Log Aggregation

For JSON log aggregation with ELK/Loki:

```bash
# Output JSON logs
./venture-server -log-format json | tee /var/log/venture/server.log

# Filter errors with jq
tail -f /var/log/venture/server.log | jq 'select(.level == "error")'

# Send to Loki with Promtail
# /etc/promtail/config.yml
```

Promtail configuration for Loki:
```yaml
server:
  http_listen_port: 9080

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://localhost:3100/loki/api/v1/push

scrape_configs:
  - job_name: venture
    static_configs:
      - targets:
          - localhost
        labels:
          job: venture
          __path__: /var/log/venture/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            msg: msg
            time: time
      - labels:
          level:
```

## Scaling Recommendations

### Vertical Scaling

**When to Use:** Single server with growing player count (up to ~20 players)

**CPU Scaling:**
- 2 cores → 4 players (baseline)
- 4 cores → 8 players
- 8 cores → 16-20 players
- Beyond 8 cores: diminishing returns, consider horizontal scaling

**Memory Scaling:**
- 2GB RAM → 4 players (~500MB)
- 4GB RAM → 8 players (~1GB)
- 8GB RAM → 16 players (~2GB)
- Memory usage: ~125MB per player

**Tuning for Vertical Scale:**
```bash
# Increase GOMAXPROCS to match cores
export GOMAXPROCS=8

# Increase connection limits
ulimit -n 65536

# Tune kernel network buffers
sudo sysctl -w net.core.rmem_max=16777216
sudo sysctl -w net.core.wmem_max=16777216
```

### Horizontal Scaling

**When to Use:** 20+ players or need geographic distribution

**Architecture Options:**

1. **Sharded (Easiest):** Independent servers with different world seeds
```bash
# Server 1: World A (US East)
./venture-server -port 8080 -seed 12345 -max-players 10

# Server 2: World B (US West)
./venture-server -port 8080 -seed 67890 -max-players 10

# Server 3: World C (EU)
./venture-server -port 8080 -seed 11111 -max-players 10
```

2. **Federated (Advanced):** Cross-server travel and shared features
```bash
# Enable federation on each server
./venture-server -port 8080 -federation-enabled -federation-peers "server2:8080,server3:8080"
```

3. **Load Balanced (High Availability):**
```haproxy
frontend venture_frontend
    bind *:8080
    default_backend venture_servers

backend venture_servers
    balance roundrobin
    option httpchk GET /health
    http-check expect status 200
    server srv1 10.0.1.10:8080 check port 9090 inter 5s fall 3 rise 2
    server srv2 10.0.1.11:8080 check port 9090 inter 5s fall 3 rise 2
    server srv3 10.0.1.12:8080 check port 9090 inter 5s fall 3 rise 2
```

### Kubernetes Deployment

Example Kubernetes deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: venture-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: venture
  template:
    metadata:
      labels:
        app: venture
    spec:
      containers:
        - name: venture
          image: ghcr.io/opd-ai/venture-server:1.0.0
          ports:
            - containerPort: 8080
              name: game
            - containerPort: 9090
              name: metrics
          resources:
            requests:
              cpu: "1"
              memory: "512Mi"
            limits:
              cpu: "2"
              memory: "1Gi"
          readinessProbe:
            httpGet:
              path: /ready
              port: 9090
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /health
              port: 9090
            initialDelaySeconds: 30
            periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: venture-service
spec:
  selector:
    app: venture
  ports:
    - name: game
      port: 8080
      targetPort: 8080
    - name: metrics
      port: 9090
      targetPort: 9090
  type: LoadBalancer
```

### Capacity Planning

| Players | Servers | CPU (total) | RAM (total) | Bandwidth |
|---------|---------|-------------|-------------|-----------|
| 1-4     | 1       | 2 cores     | 2 GB        | 1 Mbps    |
| 5-10    | 1       | 4 cores     | 4 GB        | 2.5 Mbps  |
| 11-20   | 1-2     | 8 cores     | 8 GB        | 5 Mbps    |
| 21-50   | 3-5     | 16 cores    | 16 GB       | 12.5 Mbps |
| 50-100  | 5-10    | 32 cores    | 32 GB       | 25 Mbps   |
| 100+    | 10+     | 64+ cores   | 64+ GB      | 50+ Mbps  |

See also: [docs/CAPACITY_PLANNING.md](CAPACITY_PLANNING.md) for detailed benchmarks.

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

## Backup & Disaster Recovery

### Backup Strategy

**What to Back Up:**
- Save files: `~/.venture/saves/` or `/var/lib/venture/saves/`
- Configuration: Service files, environment variables
- World seeds: Document seeds for reproducibility

**Automated Backup Script:**

Create `/opt/venture/backup.sh`:
```bash
#!/bin/bash
set -e

BACKUP_DIR="/backup/venture"
DATE=$(date +%Y%m%d_%H%M%S)
SAVE_DIR="/var/lib/venture/saves"
RETENTION_DAYS=30

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Create compressed backup with checksum
tar -czf "$BACKUP_DIR/venture-$DATE.tar.gz" -C /var/lib/venture saves/
sha256sum "$BACKUP_DIR/venture-$DATE.tar.gz" > "$BACKUP_DIR/venture-$DATE.tar.gz.sha256"

# Upload to offsite storage (optional)
# aws s3 cp "$BACKUP_DIR/venture-$DATE.tar.gz" s3://your-bucket/backups/
# rclone copy "$BACKUP_DIR/venture-$DATE.tar.gz" remote:backups/

# Clean up old backups
find "$BACKUP_DIR" -name "venture-*.tar.gz" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "venture-*.sha256" -mtime +$RETENTION_DAYS -delete

echo "Backup complete: venture-$DATE.tar.gz"
```

**Schedule with cron:**
```bash
# Add to crontab: sudo crontab -e
# Daily backup at 3 AM
0 3 * * * /opt/venture/backup.sh >> /var/log/venture/backup.log 2>&1

# Hourly backup for high-availability
0 * * * * /opt/venture/backup.sh >> /var/log/venture/backup.log 2>&1
```

### Disaster Recovery Procedures

#### Scenario 1: Save File Corruption

1. Stop the server:
```bash
sudo systemctl stop venture-server
```

2. Check for automatic backups:
```bash
ls -la /var/lib/venture/saves/*.bak
ls -la /var/lib/venture/saves/*.sha256
```

3. Restore from backup:
```bash
# Use automatic recovery (built-in)
cd /var/lib/venture/saves
cp world-12345.json.bak world-12345.json

# Or restore from daily backup
tar -xzf /backup/venture/venture-20260120_030000.tar.gz -C /var/lib/venture/
```

4. Verify checksum:
```bash
sha256sum -c /var/lib/venture/saves/world-12345.sha256
```

5. Start the server:
```bash
sudo systemctl start venture-server
```

See also: [docs/runbooks/save-corruption.md](runbooks/save-corruption.md)

#### Scenario 2: Server Hardware Failure

1. Provision new server with same specs

2. Install Venture:
```bash
# From package
sudo apt install venture

# Or from release
wget https://github.com/opd-ai/venture/releases/download/v1.0.0/venture-server-linux-amd64.tar.gz
tar -xzf venture-server-linux-amd64.tar.gz
sudo cp venture-server /usr/bin/
```

3. Restore configuration:
```bash
# Copy systemd service
sudo cp /backup/config/venture-server.service /etc/systemd/system/
sudo systemctl daemon-reload
```

4. Restore data from backup:
```bash
tar -xzf /backup/venture/venture-latest.tar.gz -C /var/lib/venture/
```

5. Verify and start:
```bash
# Verify checksums
sha256sum -c /var/lib/venture/saves/*.sha256

# Start server
sudo systemctl start venture-server
```

#### Scenario 3: Complete Data Loss

1. Use deterministic regeneration:
```bash
# Same seed = same world
./venture-server -seed 12345
```

2. Players reconnect with local save files (if available)

3. Document incident and improve backup procedures

### Recovery Time Objectives (RTO)

| Scenario              | RTO Target | Procedure                    |
|-----------------------|------------|------------------------------|
| Save corruption       | 5 minutes  | Automatic .bak restore       |
| Server restart        | 2 minutes  | systemd auto-restart         |
| Hardware failure      | 30 minutes | New server + backup restore  |
| Complete data loss    | 1 hour     | Fresh install + seed regen   |

### Testing Disaster Recovery

Quarterly DR drill checklist:
- [ ] Verify backups are being created
- [ ] Test restore from backup on staging server
- [ ] Verify save file checksums
- [ ] Test failover to backup server
- [ ] Document any issues and update procedures

## Security

**Firewall:**
```bash
# Ubuntu/Debian (ufw)
sudo ufw allow 8080/tcp   # Game server
sudo ufw deny 9090/tcp    # Metrics (internal only)
sudo ufw deny 6060/tcp    # pprof (internal only)
sudo ufw enable

# RHEL/CentOS (firewalld)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

**Best Practices:**
- Run as non-root user (see systemd examples above)
- Disable profiling in production (`-enable-profiling=false`)
- Use TLS via reverse proxy for public servers
- Rate limit connections with fail2ban
- Keep metrics/health endpoints on internal network only
- Enable input validation (built-in via pkg/validation)

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
EXPOSE 8080 9090
CMD ["/venture-server", "-port", "8080", "-enable-metrics"]
```

## Related Documentation

- [CAPACITY_PLANNING.md](CAPACITY_PLANNING.md) - Detailed capacity benchmarks
- [HEALTH_ENDPOINTS.md](HEALTH_ENDPOINTS.md) - Health/readiness endpoint details
- [PERFORMANCE.md](PERFORMANCE.md) - Performance tuning guide
- [TOR_SETUP.md](TOR_SETUP.md) - High-latency/Tor network setup
- [runbooks/](runbooks/) - Operational runbooks for common issues:
  - [high-cpu.md](runbooks/high-cpu.md) - High CPU troubleshooting
  - [memory-leak.md](runbooks/memory-leak.md) - Memory leak investigation
  - [network-issues.md](runbooks/network-issues.md) - Network connectivity problems
  - [save-corruption.md](runbooks/save-corruption.md) - Save file recovery
  - [federation-debug.md](runbooks/federation-debug.md) - Federation debugging

---

*This guide is maintained as part of the Venture project. Report issues at https://github.com/opd-ai/venture/issues*
