# ShellFromBrowser — Architecture Documentation

**Version:** Phase 1 MVP  
**Last Updated:** 2026-05-14  
**Status:** In Development (Week 1/5)

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Component Architecture](#component-architecture)
3. [Data Flow](#data-flow)
4. [Technology Stack](#technology-stack)
5. [Infrastructure](#infrastructure)
6. [Security Architecture](#security-architecture)
7. [Scalability & Performance](#scalability--performance)
8. [Deployment Architecture](#deployment-architecture)
9. [Monitoring & Observability](#monitoring--observability)
10. [Decision Records](#decision-records)

---

## System Overview

ShellFromBrowser is a web-based terminal solution providing secure shell access through a browser interface. It modernizes the legacy ShellInBox tool with:

- **Modern stack**: Go 1.21 + React 18
- **Security-first design**: OWASP Top 10:2021 compliance, Docker isolation, JWT auth
- **High availability**: Redis Sentinel, HAProxy load balancing
- **Production-ready**: Monitoring, logging, GDPR compliance

### Design Principles

1. **Security by default**: All components default to restrictive security posture
2. **Fail-safe**: Services degrade gracefully, no cascading failures
3. **Observable**: Comprehensive metrics + logs for troubleshooting
4. **Auditable**: All user actions logged immutably
5. **Maintainable**: Clear separation of concerns, documented interfaces

---

## Component Architecture

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                         Users (Browsers)                      │
└───────────────────────────┬──────────────────────────────────┘
                            │ HTTPS
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    HAProxy Load Balancer                     │
│  • TLS termination                                          │
│  • Health checks                                            │
│  • Round-robin routing                                      │
└────────────┬──────────────────────────┬─────────────────────┘
             │                          │
    ┌────────▼────────┐        ┌────────▼────────┐
    │  Frontend (1)   │        │  Frontend (2)   │
    │  React SPA      │        │  React SPA      │
    │  Port 3000      │        │  Port 3001      │
    └────────┬────────┘        └────────┬────────┘
             │                          │
             └────────────┬─────────────┘
                          │ WebSocket
                          ▼
             ┌────────────────────────┐
             │   WebSocket Gateway    │
             │   (Go Backend)         │
             │   • JWT auth           │
             │   • Rate limiting      │
             │   • Input sanitization │
             └────────────┬───────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Redis        │  │ PostgreSQL   │  │ Docker API   │
│ Sentinel HA  │  │ Audit Logs   │  │ Shell        │
│ • Sessions   │  │ • Immutable  │  │ Containers   │
│ • Blacklist  │  │ • Encrypted  │  │ • Isolated   │
└──────────────┘  └──────────────┘  └──────────────┘
```

### Component 1: Web Frontend

**Technology**: React 18 + TypeScript + xterm.js  
**Responsibility**: User interface, terminal emulation, WebSocket client

**Key Features**:
- xterm.js terminal emulator (full VT100/ANSI support)
- WebSocket connection management (reconnect, heartbeat)
- JWT token management (HttpOnly cookies)
- Session management UI (list, kill, export history)

**Interfaces**:
- **IN**: User interactions (keyboard, mouse, resize)
- **OUT**: WebSocket messages to Gateway (terminal input, resize events)
- **IN**: WebSocket messages from Gateway (terminal output)

**Files**:
```
frontend/
├── src/
│   ├── components/
│   │   ├── Terminal.tsx      # xterm.js wrapper
│   │   ├── SessionList.tsx   # Active sessions
│   │   └── Auth.tsx          # Login/logout
│   ├── services/
│   │   ├── websocket.ts      # WebSocket client
│   │   └── auth.ts           # JWT management
│   └── hooks/
│       └── useTerminal.ts    # Terminal state hook
└── tests/
```

---

### Component 2: WebSocket Gateway

**Technology**: Go 1.21 + Gorilla WebSocket  
**Responsibility**: WebSocket server, authentication, rate limiting, session orchestration

**Key Features**:
- JWT authentication middleware (cookie-based)
- Rate limiting (10 req/s per user, token bucket algorithm)
- Input sanitization (ANSI escape removal, shell metachar blocking)
- Session lifecycle management (spawn container, attach PTY, cleanup)
- Prometheus metrics export

**Interfaces**:
- **IN**: HTTP `/ws` (WebSocket upgrade)
- **IN**: HTTP `/health` (health check)
- **IN**: HTTP `/metrics` (Prometheus metrics)
- **OUT**: Redis Sentinel (session CRUD, blacklist check)
- **OUT**: Docker API (container spawn, attach, stop)
- **OUT**: PostgreSQL (audit logs)

**Files**:
```
backend/
├── cmd/gateway/
│   └── main.go                    # HTTP server, WebSocket handler
├── pkg/
│   ├── auth/
│   │   └── jwt.go                 # JWT generation/validation
│   ├── redis/
│   │   └── client.go              # Redis Sentinel client
│   ├── docker/
│   │   ├── executor.go            # Container lifecycle
│   │   └── pty.go                 # PTY bridge
│   └── security/
│       ├── ratelimiter.go         # Per-user rate limiter
│       └── sanitizer.go           # Input sanitization
└── tests/
```

**Configuration** (`.env`):
```bash
SERVER_PORT=8080
WS_MAX_CONNECTIONS=100
WS_RATE_LIMIT_PER_SECOND=10
JWT_SECRET=<256-bit-key>
REDIS_SENTINELS=sentinel1:26379,sentinel2:26379,sentinel3:26379
DOCKER_IMAGE=ubuntu:22.04
DOCKER_MEMORY_LIMIT=512m
DOCKER_CPU_LIMIT=0.5
```

---

### Component 3: Session Manager (Redis Sentinel)

**Technology**: Redis 7 + Sentinel (3 instances)  
**Responsibility**: Session state storage, JWT blacklist, high availability

**Architecture**:
```
┌───────────────┐
│ Redis Master  │  ← Write operations
│ (Port 6379)   │
└───────┬───────┘
        │ Replication
    ┌───┴───┐
    ▼       ▼
┌────────┐ ┌────────┐
│Replica1│ │Replica2│  ← Read operations (optional)
└────────┘ └────────┘

Sentinel1  Sentinel2  Sentinel3  ← Quorum = 2
(26379)    (26380)    (26381)
```

**Data Structures**:
```
# Session metadata
session:{session_id} → Hash
  user_id: string
  username: string
  container_id: string
  created_at: unix_timestamp
  last_activity: unix_timestamp
  device_info: string

# JWT blacklist (revoked tokens)
blacklist:{jwt_token} → "revoked"  (TTL = token expiry)

# Active sessions per user (for "Kill All")
user:{user_id}:sessions → Set [session_id1, session_id2, ...]
```

**Failover**:
- Sentinel detects master failure (heartbeat miss > 5s)
- Quorum (2/3 sentinels) votes for failover
- Replica promoted to master (< 30s)
- Clients reconnect automatically (failover-aware driver)

**Backup**: Redis RDB snapshots every 15 minutes, AOF append-only file for durability.

---

### Component 4: Shell Executor (Docker Containers)

**Technology**: Docker 24.x  
**Responsibility**: Isolated shell environments, resource limits, security isolation

**Container Spec**:
```yaml
Image: ubuntu:22.04
Resources:
  Memory: 512MB (hard limit)
  CPU: 0.5 cores (50% of 1 core)
  Swap: 0 (no swap)
Security:
  Capabilities: [CHOWN, DAC_OVERRIDE, FOWNER, SETGID, SETUID]  # Minimal
  SecurityOpt: [no-new-privileges:true]
  Seccomp: Custom profile (50 syscalls allowlist) — Phase 2
  AppArmor: Custom profile — Phase 2
Network:
  Mode: bridge (isolated, no external access by default)
Timeout:
  Idle: 30 minutes (auto-kill)
```

**Lifecycle**:
1. **Spawn**: User connects via WebSocket → Gateway spawns container
2. **Attach**: Gateway attaches to container stdin/stdout/stderr via PTY
3. **Bridge**: Bidirectional I/O between WebSocket ↔ Docker PTY
4. **Monitor**: Idle timeout tracker (30min inactivity → auto-kill)
5. **Cleanup**: User disconnects → Gateway stops + removes container

**Security Isolation**:
- **Process isolation**: cgroups, namespaces (PID, NET, IPC, UTS, MOUNT)
- **Filesystem**: Read-only root filesystem + writable /tmp (Phase 2)
- **Network**: No internet access by default (bridge mode, no gateway)
- **Capabilities**: Drop ALL, add minimal required (no CAP_SYS_ADMIN)

---

## Data Flow

### Session Creation Flow

```
1. User navigates to https://shellfrombroswer.local
   ↓
2. Frontend loads, checks JWT cookie
   ↓ (if missing)
3. Redirect to /login
   ↓
4. User enters credentials → POST /api/auth/login
   ↓
5. Backend validates credentials (LDAP/OAuth2 in Phase 3)
   ↓
6. Backend generates JWT (access 15min, refresh 7d)
   ↓
7. Set HttpOnly cookie, return success
   ↓
8. Frontend connects WebSocket: ws://backend:8080/ws
   ↓
9. Gateway reads JWT from cookie, validates
   ↓
10. Gateway checks Redis blacklist (revoked?)
   ↓ (if valid)
11. Gateway spawns Docker container
   ↓
12. Gateway stores session metadata in Redis
   ↓
13. Gateway attaches to container PTY
   ↓
14. Gateway starts bidirectional bridge (WebSocket ↔ PTY)
   ↓
15. User sees shell prompt in browser terminal
```

### Terminal I/O Flow

```
User types command in browser
   ↓ WebSocket message
Gateway receives message
   ↓ Input sanitization (ANSI escape removal, shell metachar check)
Gateway writes to Docker PTY stdin
   ↓
Container shell processes command
   ↓ Command output
Gateway reads from Docker PTY stdout
   ↓ WebSocket message
Frontend xterm.js displays output
```

### Session Termination Flow

```
User closes browser tab OR idle timeout (30min)
   ↓
WebSocket connection closes
   ↓
Gateway detects close event
   ↓
Gateway logs session end to PostgreSQL
   ↓
Gateway stops Docker container (grace 30s, then force kill)
   ↓
Gateway removes container
   ↓
Gateway deletes session metadata from Redis
```

---

## Technology Stack

| Layer | Technology | Version | Justification |
|-------|-----------|---------|---------------|
| **Frontend** | React | 18.x | Modern UI, hooks, concurrent rendering |
| | TypeScript | 5.x | Type safety, better IDE support |
| | xterm.js | 5.x | Industry-standard terminal emulator |
| | Vite | 5.x | Fast build, HMR, optimized bundles |
| **Backend** | Go | 1.21+ | Concurrency (goroutines), performance, single binary |
| | Gorilla WebSocket | 1.5.x | Battle-tested WebSocket library |
| | Zerolog | 1.31.x | Structured logging, high performance |
| **Session Store** | Redis | 7.x | In-memory performance, Sentinel HA |
| **Audit DB** | PostgreSQL | 15.x | ACID, encryption at rest, GDPR compliant |
| **Container Runtime** | Docker | 24.x | Isolation, resource limits, industry standard |
| **Load Balancer** | HAProxy | 2.8.x | High performance, TLS termination, health checks |
| **Monitoring** | Prometheus | 2.x | Metrics collection, time-series DB |
| | Grafana | 10.x | Dashboards, alerting |
| **Testing** | k6 | Latest | Load testing (100 concurrent sessions) |
| | OWASP ZAP | 2.x | Security scanning (baseline + full) |
| | AFL++ | Latest | Fuzzing (1M inputs corpus) |

---

## Infrastructure

### Development Environment

**Docker Compose** stack (local dev):
```yaml
services:
  - redis-master, redis-replica-1, redis-replica-2
  - redis-sentinel-1, redis-sentinel-2, redis-sentinel-3
  - postgres
  - backend (Go)
  - frontend (React)
  - haproxy
  - prometheus
  - grafana
```

**Startup**: `docker-compose up -d`

### Production Environment (AWS)

**Architecture**:
```
[Route 53 DNS]
      ↓
[CloudFront CDN] → S3 (static assets)
      ↓
[Application Load Balancer]
      ↓
┌──────────────────────────────────┐
│  EC2 Auto Scaling Group          │
│  • 2-4 instances (m5.large)      │
│  • Backend + Frontend            │
└──────────────────────────────────┘
      ↓
┌──────────────────────────────────┐
│  ElastiCache Redis Sentinel      │
│  • 1 master + 2 replicas         │
│  • cache.t3.medium               │
└──────────────────────────────────┘
      ↓
┌──────────────────────────────────┐
│  RDS PostgreSQL                  │
│  • db.t3.medium                  │
│  • Encryption at rest            │
│  • Automated backups (7 days)    │
└──────────────────────────────────┘
```

**Estimated Monthly Cost**: $450-500 (100 concurrent sessions)

**Breakdown**:
- EC2 (2× m5.large): $140
- ElastiCache (3× cache.t3.medium): $120
- RDS (1× db.t3.medium): $80
- ALB: $25
- Data transfer: $50
- CloudWatch: $20
- Backups (S3): $15

---

## Security Architecture

See [SECURITY.md](SECURITY.md) for comprehensive security documentation.

**Key Security Layers**:

1. **Transport**: TLS 1.3, HSTS, certificate pinning (Phase 2)
2. **Authentication**: JWT (HS256), HttpOnly cookies, 15min access tokens
3. **Authorization**: RBAC (Phase 3), per-session permissions
4. **Input Validation**: Sanitization, regex whitelist, max length 4096
5. **Container Isolation**: Docker cgroups, namespaces, seccomp, AppArmor
6. **Audit Logging**: Immutable PostgreSQL logs, GDPR compliant
7. **Rate Limiting**: Token bucket (10 req/s per user)
8. **Secrets Management**: AWS Secrets Manager (production)

---

## Scalability & Performance

### Performance Targets

| Metric | Target | Measured |
|--------|--------|----------|
| **Latency** | | |
| WebSocket connect | < 200ms | TBD (Phase 1 Week 5) |
| Container spawn | < 500ms (P95) | TBD |
| Terminal input → output | < 50ms (P95) | TBD |
| **Throughput** | | |
| Concurrent sessions | 100 (Phase 1) | TBD |
| Commands per minute (per session) | 100 | TBD |
| **Availability** | | |
| Uptime (SLA) | 99.5% (Phase 1 MVP) | TBD |
| MTTR (Mean Time To Recovery) | < 15 minutes | TBD |

### Scaling Strategy

**Vertical Scaling** (Phase 1: 10-100 sessions):
- Single backend instance (m5.large: 2 vCPU, 8GB RAM)
- Docker container density: ~50 containers per host (512MB each)
- Redis Sentinel: 256MB cache sufficient for 1000 sessions

**Horizontal Scaling** (Phase 2: 100-500 sessions):
- Add backend instances (2-4× m5.large)
- HAProxy round-robin load balancing
- Session affinity not required (stateless WebSocket, Redis stores state)
- Redis Cluster (sharding) if > 1000 sessions

**Kubernetes Migration** (Phase 3+: 500-5000 sessions):
- EKS cluster, auto-scaling (HPA based on CPU/memory)
- Redis Cluster (6-node setup: 3 masters + 3 replicas)
- Distributed Docker via Kubernetes DaemonSet

### Bottlenecks & Mitigations

| Bottleneck | Symptom | Mitigation |
|-----------|---------|------------|
| **Docker spawn latency** | P95 > 500ms | Pre-warm container pool (Phase 2) |
| **Redis SPOF** | Downtime on master failure | Sentinel HA (< 30s failover) |
| **PostgreSQL write throughput** | Audit log lag | Async writes, bulk insert (batches of 100) |
| **Network bandwidth** | Terminal output lag | Compression (gzip), output throttling |

---

## Deployment Architecture

### Blue-Green Deployment (Phase 3)

```
┌─────────────────────────────────────────┐
│         ALB (Traffic Routing)           │
└────────┬────────────────────┬───────────┘
         │                    │
    100% │ (Blue)        0%   │ (Green)
         ▼                    ▼
  ┌─────────────┐      ┌─────────────┐
  │  Blue Env   │      │  Green Env  │
  │  (v1.2.3)   │      │  (v1.3.0)   │
  │  2 instances│      │  2 instances│
  └─────────────┘      └─────────────┘
         │                    │
         └────────┬───────────┘
                  ▼
         ┌─────────────────┐
         │ Shared Redis    │
         │ (Session Store) │
         └─────────────────┘
```

**Deployment Steps**:
1. Deploy Green (v1.3.0), verify health checks
2. Run smoke tests (synthetic sessions)
3. Shift 10% traffic to Green (canary)
4. Monitor metrics (error rate, latency)
5. Shift 50% → 100% traffic to Green
6. Checkpoint sessions to Redis (resume on Green)
7. Drain Blue connections (graceful shutdown 5 min)
8. Terminate Blue

**Rollback**: Shift traffic back to Blue (< 1 minute)

---

## Monitoring & Observability

### Prometheus Metrics

```
# Active sessions
shellfrombroswer_active_sessions{user_id="alice"} 2

# Container spawn latency (histogram)
shellfrombroswer_spawn_latency_seconds_bucket{le="0.5"} 450
shellfrombroswer_spawn_latency_seconds_bucket{le="1.0"} 480
shellfrombroswer_spawn_latency_seconds_sum 237.4
shellfrombroswer_spawn_latency_seconds_count 500

# WebSocket errors (counter)
shellfrombroswer_websocket_errors_total{reason="rate_limit"} 23

# JWT revocations (counter)
shellfrombroswer_jwt_revocations_total 5

# Container escapes detected (counter, should be 0)
shellfrombroswer_container_escapes_total 0

# DDoS attempts blocked (counter)
shellfrombroswer_ddos_blocked_total{source_ip="1.2.3.4"} 150

# Redis Sentinel failovers (counter)
shellfrombroswer_redis_failover_total 1
```

### Grafana Dashboards

1. **Overview**: Active sessions, error rate, latency P50/P95/P99
2. **Performance**: Container spawn latency, WebSocket throughput
3. **Security**: Auth failures, rate limit hits, blacklisted tokens
4. **Infra**: CPU/memory/disk (backend), Redis ops/sec, PostgreSQL connections

### Alerting Rules

```yaml
- alert: HighErrorRate
  expr: rate(shellfrombroswer_websocket_errors_total[5m]) > 0.1
  for: 5m
  annotations:
    summary: "Error rate > 10% for 5 minutes"

- alert: ContainerEscapeDetected
  expr: shellfrombroswer_container_escapes_total > 0
  for: 1m
  annotations:
    summary: "CRITICAL: Container escape detected!"

- alert: RedisMasterDown
  expr: redis_up{role="master"} == 0
  for: 1m
  annotations:
    summary: "Redis master is down (Sentinel failover in progress)"
```

---

## Decision Records

### ADR-001: Go vs Node.js for Backend

**Date**: 2026-05-14  
**Status**: Accepted

**Context**: Need high-performance WebSocket server with Docker API integration.

**Decision**: Use Go 1.21

**Rationale**:
- Native concurrency (goroutines) scales to 10K+ WebSocket connections per instance
- Docker SDK official support (docker/docker client)
- Single binary deployment (no runtime dependencies)
- Lower memory footprint vs Node.js (50MB vs 200MB idle)

**Consequences**:
- +: Performance, deployment simplicity, type safety
- -: Smaller ecosystem vs Node.js, steeper learning curve for frontend devs

---

### ADR-002: Redis Sentinel vs Redis Cluster

**Date**: 2026-05-14  
**Status**: Accepted

**Context**: Need HA session store for 100 concurrent sessions.

**Decision**: Redis Sentinel (1 master + 2 replicas)

**Rationale**:
- Sentinel sufficient for < 500 sessions (no sharding needed)
- Simpler ops vs Cluster (no hash slots, no resharding)
- Automatic failover (< 30s)
- Cost: 3× instances vs 6× for Cluster

**Consequences**:
- +: Simplicity, lower cost, adequate for Phase 1-2
- -: No horizontal scaling (migrate to Cluster if > 500 sessions)

---

### ADR-003: Docker vs Namespaces for Isolation

**Date**: 2026-05-14  
**Status**: Accepted

**Context**: Need isolated shell environments with resource limits.

**Decision**: Docker containers (not raw namespaces)

**Rationale**:
- Resource limits (cgroups) easier with Docker
- Image-based deployment (reproducible environments)
- Security profiles (seccomp, AppArmor) via Docker
- Team familiarity with Docker

**Consequences**:
- +: Ease of use, reproducibility, security profiles
- -: Overhead ~15% (Docker daemon, image layers) vs raw namespaces

---

### ADR-004: JWT HttpOnly Cookies vs Authorization Header

**Date**: 2026-05-14  
**Status**: Accepted

**Context**: Need XSS-resistant auth mechanism.

**Decision**: JWT in HttpOnly cookies (not `Authorization: Bearer` header)

**Rationale**:
- HttpOnly cookies immune to XSS (JavaScript can't read)
- CSRF protection via SameSite=Strict
- Automatic inclusion in WebSocket upgrade (no manual header injection)

**Consequences**:
- +: XSS protection, CSRF protection
- -: CORS complexity (credentials: true), mobile app integration harder (Phase 3)

---

**Next**: [SECURITY.md](SECURITY.md) | [DEPLOYMENT.md](DEPLOYMENT.md) | [RUNBOOK.md](RUNBOOK.md)
