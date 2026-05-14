# ShellFromBrowser — Security Documentation

**Version:** Phase 1 MVP  
**Last Updated:** 2026-05-14  
**Compliance:** OWASP Top 10:2021, GDPR

---

## Table of Contents

1. [Threat Model](#threat-model)
2. [OWASP Top 10:2021 Coverage](#owasp-top-102021-coverage)
3. [Authentication & Authorization](#authentication--authorization)
4. [Input Validation & Sanitization](#input-validation--sanitization)
5. [Container Security](#container-security)
6. [Network Security](#network-security)
7. [Data Protection](#data-protection)
8. [Audit Logging](#audit-logging)
9. [GDPR Compliance](#gdpr-compliance)
10. [Security Testing](#security-testing)
11. [Incident Response](#incident-response)
12. [Security Checklist](#security-checklist)

---

## Threat Model

### Threat Actors

#### Attacker A: Legitimate User Compromised
**Profile**: External attacker who has stolen credentials of a legitimate user.

**Scenario**: Laptop of admin "alice" is stolen. JWT token cached in browser, no device fingerprinting yet implemented.

**Attack Vectors**:
- Reuse stolen JWT to spawn malicious containers
- Exfiltrate data from existing sessions
- Lateral movement to other users' containers (if isolation fails)

**Mitigations**:
| Mitigation | Implementation | Status |
|-----------|----------------|--------|
| JWT blacklist (revocation) | Redis sorted set, endpoint `/auth/revoke` | ✅ Phase 1 |
| Device fingerprinting | User-Agent + IP + Canvas hash | 🔄 Phase 2 |
| "Kill All My Sessions" button | User-initiated revocation of all active sessions | ✅ Phase 1 |
| Short-lived access tokens | 15 minutes expiry, refresh token rotation | ✅ Phase 1 |
| IP allowlisting (optional) | Admin config per user | ❌ Phase 3+ |

---

#### Attacker B: Internal User Malicious (Privilege Escalation)
**Profile**: Legitimate user (e.g., junior admin) attempting privilege escalation.

**Scenario**: User "bob" tries `sudo su` to escalate to root, then attempts container escape via kernel CVE.

**Attack Vectors**:
- Privilege escalation inside container (sudo, setuid binaries)
- Container escape via CVE (dirty cow, runc vulnerability)
- Abuse of overly permissive capabilities (CAP_SYS_ADMIN)
- Access Docker socket from inside container

**Mitigations**:
| Mitigation | Implementation | Status |
|-----------|----------------|--------|
| Seccomp profile (allowlist) | 50 syscalls allowlist (blocks ptrace, mount, etc.) | 🔄 Phase 2 |
| AppArmor profile | Block /proc, /sys writes, Docker socket access | 🔄 Phase 2 |
| Capabilities drop | Drop ALL, add minimal (CHOWN, DAC_OVERRIDE, FOWNER, SETGID, SETUID) | ✅ Phase 1 |
| Read-only root filesystem | Writable /tmp only | 🔄 Phase 2 |
| Docker image scanning | Trivy daily scan for CVEs | ✅ Phase 1 (CI/CD) |
| No Docker socket mount | Explicitly blocked in container config | ✅ Phase 1 |
| Resource limits | 512MB RAM, 0.5 CPU (prevents DoS via fork bomb) | ✅ Phase 1 |

---

#### Attacker C: External Without Credentials (RCE)
**Profile**: External attacker attempting remote code execution without valid credentials.

**Scenario**: Attacker crafts malicious input to exploit WebSocket handler (ANSI escape sequence injection, Unicode RLO spoofing, command injection).

**Attack Vectors**:
- Command injection via shell metacharacters (`;`, `|`, `$()`)
- ANSI escape sequence hijacking (terminal control sequences)
- Unicode right-to-left override (U+202E) for visual spoofing
- Fuzzing WebSocket binary frames to trigger crashes
- DDoS via connection exhaustion

**Mitigations**:
| Mitigation | Implementation | Status |
|-----------|----------------|--------|
| Input sanitization | Regex whitelist, ANSI escape removal, shell metachar blocking | ✅ Phase 1 |
| Fuzzing (AFL++) | 1M+ inputs corpus, crash detection | 🔄 Phase 1 Week 4 |
| Rate limiting | 10 req/s per user (token bucket), 100 cmd/min per session | ✅ Phase 1 |
| Max input length | 4096 bytes hard limit | ✅ Phase 1 |
| WebSocket frame size limit | 1MB max | ✅ Phase 1 (Gorilla default) |
| CORS strict origin check | Allowlist configured origins | ✅ Phase 1 |
| TLS 1.3 | Encryption in transit | 🔄 Phase 1 Week 3 |
| OWASP ZAP scan | Baseline + full scan before production | 🔄 Phase 1 Week 4 |

---

## OWASP Top 10:2021 Coverage

| # | Category | Risk | Mitigations | Status |
|---|----------|------|-------------|--------|
| **A01** | **Broken Access Control** | 🔴 Critical | JWT auth, RBAC (Phase 3), session isolation (Docker) | ✅ Phase 1 |
| **A02** | **Cryptographic Failures** | 🟠 High | TLS 1.3, bcrypt passwords (Phase 2), JWT HS256, PostgreSQL encryption at rest | 🔄 Phase 1-2 |
| **A03** | **Injection** | 🔴 Critical | Input sanitization (regex whitelist), shell metachar blocking, parameterized SQL | ✅ Phase 1 |
| **A04** | **Insecure Design** | 🟠 High | Threat model documented, security review (Council audit), pentest | ✅ Phase 1 |
| **A05** | **Security Misconfiguration** | 🟠 High | Docker seccomp/AppArmor, no-new-privileges, HSTS, security headers | 🔄 Phase 1-2 |
| **A06** | **Vulnerable Components** | 🟡 Medium | Trivy daily scan, Dependabot, Go mod security updates | ✅ Phase 1 (CI/CD) |
| **A07** | **Authentication Failures** | 🔴 Critical | JWT (HttpOnly cookies), 15min expiry, blacklist revocation, rate limiting auth endpoints | ✅ Phase 1 |
| **A08** | **Software & Data Integrity** | 🟡 Medium | Docker image signing (Phase 2), GitHub Actions OIDC, checksums | 🔄 Phase 2 |
| **A09** | **Logging & Monitoring Failures** | 🟠 High | Structured logs (Zerolog), immutable PostgreSQL audit, Prometheus metrics, alerts | ✅ Phase 1 |
| **A10** | **SSRF (Server-Side Request Forgery)** | 🟡 Medium | No outbound HTTP from containers (bridge network, no gateway), URL parsing validation | ✅ Phase 1 |

**Legend**: 🔴 Critical, 🟠 High, 🟡 Medium, 🟢 Low

---

## Authentication & Authorization

### JWT Token Strategy

**Token Types**:
1. **Access Token**: Short-lived (15 minutes), carried in HttpOnly cookie
2. **Refresh Token**: Long-lived (7 days), used to obtain new access tokens

**Token Structure** (JWT claims):
```json
{
  "iss": "shellfrombroswer",
  "aud": "shellfrombroswer-users",
  "sub": "user:alice",
  "user_id": "alice",
  "username": "Alice Admin",
  "role": "admin",
  "iat": 1715702400,
  "exp": 1715703300,
  "nbf": 1715702400
}
```

**Signing Algorithm**: HS256 (HMAC-SHA256, symmetric key)
- **Rationale**: Simpler key management than RS256 (no public/private keypair), sufficient for Phase 1 (backend-only validation)
- **Phase 3**: Migrate to RS256 for microservices (public key distribution)

**Storage**:
- **HttpOnly cookie**: JavaScript cannot read (XSS protection)
- **SameSite=Strict**: CSRF protection
- **Secure flag**: HTTPS-only transmission

**Validation Flow**:
```
1. User sends request with cookie: session_token=<jwt>
   ↓
2. Backend reads cookie, extracts JWT
   ↓
3. Validate signature (HS256, secret key)
   ↓
4. Validate claims (issuer, audience, expiry, not-before)
   ↓
5. Check Redis blacklist (revoked token?)
   ↓ (if valid)
6. Extract user_id, role from claims
   ↓
7. Proceed with request
```

### Token Revocation (Blacklist)

**Problem**: JWT is stateless, cannot be invalidated server-side (until expiry).

**Solution**: Redis blacklist with TTL.

**Implementation**:
```redis
# Blacklist token (key = JWT hash, TTL = token expiry - now)
SET blacklist:sha256(jwt_token) "revoked" EX 900

# Check blacklist
EXISTS blacklist:sha256(jwt_token)
```

**Endpoints**:
- `POST /auth/revoke` — Revoke current token (logout)
- `POST /auth/revoke-all` — Revoke all tokens for user (Kill All My Sessions)

---

### Authorization (RBAC — Phase 3)

**Roles**:
- `user`: Basic access (own sessions only)
- `admin`: Full access (all sessions, user management)
- `viewer`: Read-only access (view sessions, no shell access)

**Permissions Matrix**:
| Action | User | Admin | Viewer |
|--------|------|-------|--------|
| Create session | ✅ | ✅ | ❌ |
| View own sessions | ✅ | ✅ | ✅ |
| View all sessions | ❌ | ✅ | ✅ |
| Kill own sessions | ✅ | ✅ | ❌ |
| Kill any session | ❌ | ✅ | ❌ |
| Export session history | ✅ | ✅ | ✅ |

---

## Input Validation & Sanitization

### WebSocket Input Sanitization

**Pipeline**:
```
User input (WebSocket message)
  ↓
1. Length check (max 4096 bytes)
  ↓
2. ANSI escape sequence removal (regex: \x1b\[[0-9;]*[a-zA-Z])
  ↓
3. Unicode RLO detection (U+202E right-to-left override)
  ↓
4. Shell metacharacter blocking (;, &, |, `, $, <, >, \n, \r)
  ↓
5. Null byte detection (\x00)
  ↓
6. Whitespace trimming
  ↓
Sanitized input → Docker PTY stdin
```

**Code** (`pkg/security/sanitizer.go`):
```go
func SanitizeInput(input string, maxLength int) (string, bool) {
    // 1. Length check
    if len(input) > maxLength {
        return "", false
    }

    // 2. Remove ANSI escape sequences
    sanitized := ansiEscapeRegex.ReplaceAllString(input, "")

    // 3. Check Unicode RLO
    if strings.ContainsRune(sanitized, '‮') {
        return "", false
    }

    // 4. Check shell metacharacters
    for _, meta := range shellMetachars {
        if strings.Contains(sanitized, meta) {
            return "", false
        }
    }

    // 5. Check null bytes
    if strings.ContainsRune(sanitized, '\x00') {
        return "", false
    }

    // 6. Trim whitespace
    return strings.TrimSpace(sanitized), true
}
```

**Dangerous Patterns Detected**:
- `rm -rf` (destructive command)
- `:(){ :|:& };:` (fork bomb)
- `/dev/tcp` (network redirection)
- `chmod +s` (SUID escalation)
- `curl | sh` / `wget | sh` (remote code execution)

**Action**: Log to audit DB, reject input, alert admin.

---

## Container Security

### Docker Isolation Layers

#### 1. Linux Namespaces
**Purpose**: Process, network, filesystem isolation.

| Namespace | Isolation | Impact |
|-----------|-----------|--------|
| PID | Process IDs | Container sees only its own processes (PID 1 = shell) |
| NET | Network stack | Isolated network interfaces, no access to host network |
| IPC | Inter-process communication | Isolated shared memory, message queues |
| UTS | Hostname | Isolated hostname (prevents host identification) |
| MOUNT | Filesystem mounts | Isolated /proc, /sys, tmpfs |
| USER | User IDs (Phase 2) | Map root inside container to unprivileged user on host |

#### 2. Cgroups (Resource Limits)
**Purpose**: Prevent resource exhaustion (DoS).

```yaml
Resources:
  Memory: 512MB (hard limit, OOM killer if exceeded)
  CPU: 0.5 cores (50% of 1 core, throttled if exceeded)
  Swap: 0 (no swap, prevents memory leak persistence)
  PIDs: 100 (max processes, prevents fork bomb)
```

**Enforcement**: Docker daemon applies limits via cgroups v2.

#### 3. Capabilities (Least Privilege)
**Purpose**: Remove dangerous kernel capabilities.

**Default Docker capabilities** (33 total):
```
CHOWN, DAC_OVERRIDE, FOWNER, FSETID, KILL, SETGID, SETUID, SETPCAP,
NET_BIND_SERVICE, NET_RAW, SYS_CHROOT, MKNOD, AUDIT_WRITE, SETFCAP
```

**ShellFromBrowser allowlist** (5 minimal):
```yaml
CapDrop: [ALL]
CapAdd:
  - CHOWN       # Change file ownership
  - DAC_OVERRIDE  # Bypass file read/write/execute checks
  - FOWNER      # Bypass file owner checks
  - SETGID      # Set group ID
  - SETUID      # Set user ID
```

**Blocked capabilities** (dangerous):
- `CAP_SYS_ADMIN` — Mount filesystems, load kernel modules
- `CAP_NET_ADMIN` — Configure network (iptables, routing)
- `CAP_SYS_PTRACE` — Trace processes (debug other containers)
- `CAP_SYS_MODULE` — Insert kernel modules

#### 4. Seccomp (Syscall Filtering) — Phase 2
**Purpose**: Whitelist allowed syscalls, block dangerous ones.

**Default Docker seccomp** profile: ~300 syscalls allowed.

**ShellFromBrowser custom profile** (50 syscalls):
```json
{
  "defaultAction": "SCMP_ACT_ERRNO",
  "syscalls": [
    {"names": ["read", "write", "open", "close", "stat", "fstat"], "action": "SCMP_ACT_ALLOW"},
    {"names": ["execve", "fork", "clone"], "action": "SCMP_ACT_ALLOW"},
    {"names": ["mmap", "mprotect", "munmap"], "action": "SCMP_ACT_ALLOW"},
    ...
  ]
}
```

**Blocked syscalls** (dangerous):
- `ptrace` — Debug processes (container escape vector)
- `mount`, `umount2` — Modify filesystems
- `reboot`, `sethostname` — Affect host system
- `create_module`, `init_module` — Load kernel modules

#### 5. AppArmor Profile — Phase 2
**Purpose**: Mandatory Access Control (MAC), restrict file/network access.

**Profile** (`/etc/apparmor.d/shellfrombroswer-container`):
```apparmor
profile shellfrombroswer-container flags=(attach_disconnected,mediate_deleted) {
  # Deny Docker socket access (container escape)
  deny /var/run/docker.sock rw,

  # Deny /proc writes (information leakage)
  deny /proc/* w,

  # Deny /sys writes (kernel parameter modification)
  deny /sys/* w,

  # Allow /tmp (writable workspace)
  /tmp/** rw,

  # Allow user home directory (Phase 2: per-user volume)
  /home/*/** rw,
}
```

---

### Docker Image Security

**Base Image**: `ubuntu:22.04` (LTS, 5 years security updates)

**Image Hardening**:
1. **Minimal packages**: Only essential tools (bash, coreutils, no curl/wget by default)
2. **No SUID binaries**: Remove setuid bit from all binaries (prevents privilege escalation)
3. **Read-only filesystem** (Phase 2): Only /tmp writable
4. **Non-root user** (Phase 2): Run shell as unprivileged user (UID 1000)

**Image Scanning**:
- **Tool**: Trivy (Aqua Security)
- **Frequency**: Daily scan in CI/CD
- **Action**: Block deployment if HIGH/CRITICAL CVEs detected

**Example scan**:
```bash
trivy image shellfrombroswer/shell-executor:latest
```

---

## Network Security

### TLS Configuration

**Protocol**: TLS 1.3 (no TLS 1.2, SSLv3)

**Cipher Suites** (HAProxy config):
```
ssl-default-bind-ciphers TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256
ssl-default-bind-options ssl-min-ver TLSv1.3
```

**HSTS Header**:
```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```

**Certificate Management**:
- **Let's Encrypt**: Automatic renewal via Certbot
- **Expiry**: 90 days, auto-renew at 30 days
- **Backup**: Manual wildcard cert in AWS Secrets Manager

---

### CORS Policy

**Allowed Origins** (`.env` config):
```bash
CORS_ALLOWED_ORIGINS=https://shellfrombroswer.local,https://app.shellfrombroswer.local
```

**Headers**:
```go
w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
w.Header().Set("Access-Control-Allow-Credentials", "true")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
```

**Preflight Cache**: 1 hour (`Access-Control-Max-Age: 3600`)

---

### Firewall Rules (AWS Security Groups)

**Ingress**:
| Port | Protocol | Source | Purpose |
|------|----------|--------|---------|
| 443 | TCP | 0.0.0.0/0 | HTTPS (ALB) |
| 80 | TCP | 0.0.0.0/0 | HTTP → HTTPS redirect |
| 22 | TCP | Admin IP | SSH (bastion only) |

**Egress**:
| Port | Protocol | Destination | Purpose |
|------|----------|-------------|---------|
| 443 | TCP | 0.0.0.0/0 | HTTPS (Docker Hub, AWS APIs) |
| 6379 | TCP | Redis SG | Redis Sentinel |
| 5432 | TCP | RDS SG | PostgreSQL |

**Deny all other traffic** (default deny).

---

## Data Protection

### Data at Rest

**PostgreSQL (Audit Logs)**:
- **Encryption**: AES-256 (AWS RDS encryption at rest)
- **Key Management**: AWS KMS (customer-managed key)
- **Backups**: Encrypted, 7-day retention
- **Access**: IAM role (no password in config)

**Redis (Session State)**:
- **Encryption**: TLS in transit, no sensitive data (session IDs only)
- **Backups**: RDB snapshots every 15 min, encrypted S3

**Secrets (Phase 2)**:
- **AWS Secrets Manager**: JWT secret, database passwords, TLS certs
- **Rotation**: Automatic every 90 days

### Data in Transit

**TLS 1.3**: All external communication (browser ↔ ALB ↔ backend)

**Internal Communication**:
- Backend ↔ Redis: TLS (Phase 2)
- Backend ↔ PostgreSQL: TLS (forced via `sslmode=require`)

---

### GDPR Compliance

#### Data Minimization

**Data Collected**:
| Data Type | Purpose | Retention | Legal Basis |
|-----------|---------|-----------|-------------|
| Username | Authentication | Until account deleted | Contractual |
| Email | Password reset (Phase 2) | Until account deleted | Contractual |
| IP Address | Audit logging, abuse prevention | 30 days (anonymized after) | Legitimate interest |
| Session logs | Security auditing | 365 days | Legal obligation |
| Command history | User-requested export | On-demand only | Consent |

**No PII collected**: No real names, phone numbers, addresses (unless user includes in terminal).

---

#### Data Subject Rights

**Right to Access** (GDPR Article 15):
- **Endpoint**: `GET /gdpr/export/:user_id`
- **Format**: JSON or CSV
- **Content**: All audit logs, session metadata, IP addresses (last 30 days)

**Right to Erasure** (GDPR Article 17):
- **Endpoint**: `DELETE /gdpr/erase/:user_id`
- **Action**: Delete user account, sessions, audit logs (except legal hold)
- **Retention Exception**: Security incident logs (6 years legal retention)

**Right to Rectification** (GDPR Article 16):
- **Endpoint**: `PATCH /api/users/:user_id`
- **Action**: Update username, email

---

#### Data Breach Notification

**Timeline**: Report to supervisory authority within 72 hours (GDPR Article 33).

**Notification Process**:
1. Detect breach (automated alert: container escape, PostgreSQL unauthorized access)
2. Assess impact (data exfiltrated? PII exposed?)
3. Notify DPA (Data Protection Authority) within 72h
4. Notify affected users within 72h (if high risk)
5. Document in breach register

---

## Security Testing

### Penetration Testing

**Phase 1 (Week 5)**: Internal white-box pentest
- **Scope**: WebSocket Gateway, container isolation, JWT auth
- **Tools**: Burp Suite, custom scripts
- **Deliverable**: Findings report (severity: Critical/High/Medium/Low)

**Phase 3**: External black-box pentest
- **Vendor**: TBD (budget $3-5K)
- **Scope**: Full application (unauthenticated + authenticated)
- **Standard**: OWASP ASVS Level 2

---

### Fuzzing (AFL++)

**Target**: WebSocket message handler

**Corpus**: 1M+ inputs (valid commands, shell metacharacters, ANSI escapes, Unicode edge cases)

**Run**:
```bash
cd backend/tests/fuzzing
./run-afl.sh
```

**Monitoring**: Check `crashes/` and `hangs/` directories daily.

**Action**: Fix crashes within 24h (high priority).

---

### OWASP ZAP Scan

**Baseline Scan** (CI/CD, every commit):
```bash
docker run -t owasp/zap2docker-stable zap-baseline.py -t http://localhost:3000
```

**Full Scan** (Phase 1 Week 4, before production):
```bash
docker run -t owasp/zap2docker-stable zap-full-scan.py -t http://localhost:3000
```

**Results**: Upload to GitHub Security tab (SARIF format).

---

## Incident Response

### Runbook: Container Escape Detected

**Trigger**: Prometheus alert `shellfrombroswer_container_escapes_total > 0`

**Actions**:
1. **Immediate** (5 min):
   - Kill all running containers: `docker ps -q | xargs docker kill`
   - Block new sessions: Feature flag `FEATURE_SESSIONS_ENABLED=false`
   - Alert security team (PagerDuty)

2. **Investigate** (30 min):
   - Check audit logs: `SELECT * FROM audit_logs WHERE event='container_escape'`
   - Identify exploited CVE: `docker inspect <container_id>`
   - Collect forensics: Container logs, kernel logs

3. **Remediate** (2 hours):
   - Patch Docker / kernel CVE
   - Update seccomp/AppArmor profiles
   - Rescan all images with Trivy

4. **Restore** (1 hour):
   - Deploy patched version (blue-green)
   - Re-enable sessions
   - Notify users of downtime

5. **Post-Mortem** (48 hours):
   - Root cause analysis (RCA document)
   - Update threat model
   - GDPR breach notification (if PII exposed)

---

## Security Checklist

### Pre-Production Checklist

- [ ] **Authentication**
  - [ ] JWT secret is 256-bit random (not default)
  - [ ] Access tokens expire in 15 minutes
  - [ ] HttpOnly + Secure + SameSite=Strict cookies
  - [ ] Blacklist implemented (Redis)

- [ ] **Container Security**
  - [ ] Capabilities dropped (only 5 allowed)
  - [ ] Seccomp profile enabled (Phase 2)
  - [ ] AppArmor profile enabled (Phase 2)
  - [ ] Resource limits (512MB RAM, 0.5 CPU)
  - [ ] Read-only root filesystem (Phase 2)
  - [ ] Docker images scanned (Trivy, no HIGH/CRITICAL)

- [ ] **Network**
  - [ ] TLS 1.3 enabled (no TLS 1.2)
  - [ ] HSTS header enabled
  - [ ] CORS origins whitelisted (not *)
  - [ ] Security groups restrictive (deny all except required)

- [ ] **Input Validation**
  - [ ] Max input length (4096 bytes)
  - [ ] ANSI escape removal
  - [ ] Shell metacharacter blocking
  - [ ] Dangerous pattern detection

- [ ] **Monitoring**
  - [ ] Prometheus alerts configured (error rate, container escapes)
  - [ ] Audit logs enabled (PostgreSQL)
  - [ ] Grafana dashboards deployed

- [ ] **Testing**
  - [ ] OWASP ZAP full scan passed (no HIGH/CRITICAL)
  - [ ] Fuzzing 1M inputs (no crashes)
  - [ ] Internal pentest completed (findings remediated)
  - [ ] Load test 100 concurrent sessions (success)

- [ ] **Compliance**
  - [ ] GDPR export endpoint tested
  - [ ] GDPR erasure endpoint tested
  - [ ] Data retention policies documented
  - [ ] Privacy policy published

---

**Next**: [DEPLOYMENT.md](DEPLOYMENT.md) | [RUNBOOK.md](RUNBOOK.md) | [ARCHITECTURE.md](ARCHITECTURE.md)
