# ShellFromBrowser

**Version modernisée et refactorisée de ShellInBox** — Terminal shell Unix/Linux accessible via navigateur web.

---

## 📋 Spécifications Contractuelles

| Paramètre | Valeur |
|-----------|--------|
| **Volumétrie max** | 100 sessions simultanées |
| **Latence P95** | < 500ms spawn container |
| **Budget infra** | < $500/mois AWS |
| **Architecture** | Standalone (SSO optionnel Phase 3) |
| **Public cible** | Admins système niveau intermédiaire |
| **Sécurité** | OWASP Top 10:2021, pentest externe obligatoire |

---

## 🏗️ Architecture (4 Composants)

```
┌─────────────────┐
│  Web Frontend   │  React 18 + TypeScript + xterm.js
│  (SPA)          │  2 instances HAProxy round-robin
└────────┬────────┘
         │ HTTPS
┌────────▼─────────┐
│ WebSocket Gateway│  Go 1.21, Gorilla WebSocket
│                  │  Rate limiting 10 req/s, CORS strict
└────────┬─────────┘
         │ Internal
┌────────▼─────────┐
│ Session Manager  │  JWT HttpOnly cookies (15min access, 7j refresh)
│                  │  Redis Sentinel (1 master + 2 replicas, HA)
└────────┬─────────┘
         │
┌────────▼─────────┐
│ Shell Executor   │  Docker 24.x containers
│                  │  512MB RAM, 0.5 CPU, idle timeout 30min
│                  │  Seccomp, AppArmor, capabilities drop
└──────────────────┘
```

---

## 🎯 Roadmap

| Phase | Durée | Focus | Validation |
|-------|-------|-------|-----------|
| **Phase 1 MVP** | 5 semaines | Auth JWT, Redis Sentinel HA, 1 shell/user, monitoring basique | Pentest white-box interne, OWASP ZAP scan |
| **Phase 2** | 4 semaines | Multi-tabs, thèmes, historique, session recording | Load test k6 100 sessions, chaos Pumba |
| **Phase 3** | 4 semaines | OAuth2 SSO, MFA, RBAC, GDPR compliance, blue-green | Pentest externe $3-5K, OWASP ASVS L2 |

**Total :** 13 semaines, 2 devs, budget $150K (dev + pentest).

---

## 🔐 Threat Model

### Attaquant A : Utilisateur Légitime Compromis
- **Scénario :** Laptop admin volé, JWT en cache
- **Mitigation :** Device fingerprinting, blacklist JWT immédiate, "Kill All My Sessions"

### Attaquant B : Utilisateur Interne Malveillant
- **Scénario :** Privilege escalation, container escape CVE
- **Mitigation :** Seccomp allowlist 50 syscalls, AppArmor, Docker scan Trivy daily

### Attaquant C : Externe Sans Credentials
- **Scénario :** RCE via WebSocket (ANSI escape, Unicode RLO)
- **Mitigation :** Input sanitization regex whitelist, fuzzing AFL++ 1M inputs

---

## 🛠️ Stack Technique

**Backend :**
- Go 1.21+ (concurrence native, binaire standalone)
- Gorilla WebSocket (rate limiting, CORS)
- Redis Sentinel (HA master + 2 replicas)

**Frontend :**
- React 18 + TypeScript
- xterm.js (terminal emulator)
- Vite (build tool)

**Infra :**
- Docker 24.x (isolation cgroups, seccomp, AppArmor)
- HAProxy 2.8 (load balancer, TLS termination)
- PostgreSQL 15 (audit logs, encryption at rest)

**Monitoring :**
- Prometheus + Grafana
- 7 metrics clés : active_sessions, spawn_latency_p95, websocket_errors, jwt_revocations, container_escapes, ddos_blocked, redis_failover_count

---

## 📂 Structure Projet

```
ShellFromBrowser/
├── backend/               # Go backend
│   ├── cmd/
│   │   ├── gateway/      # WebSocket Gateway
│   │   ├── session/      # Session Manager
│   │   └── executor/     # Shell Executor
│   ├── pkg/
│   │   ├── auth/         # JWT, device fingerprinting
│   │   ├── redis/        # Redis Sentinel client
│   │   ├── docker/       # Container orchestration
│   │   └── security/     # Input sanitization, rate limiting
│   └── tests/
├── frontend/              # React frontend
│   ├── src/
│   │   ├── components/
│   │   ├── hooks/
│   │   └── services/
│   └── tests/
├── infra/
│   ├── docker/           # Dockerfiles
│   ├── k8s/              # Kubernetes manifests (Phase 3+)
│   ├── terraform/        # AWS provisioning
│   └── monitoring/       # Prometheus, Grafana configs
├── docs/
│   ├── ARCHITECTURE.md
│   ├── SECURITY.md
│   ├── DEPLOYMENT.md
│   └── RUNBOOK.md
└── .github/
    └── workflows/        # CI/CD GitHub Actions
```

---

## 🚀 Quick Start (Dev Local)

### Pré-requis
- Go 1.21+
- Node.js 18+
- Docker 24+
- Redis (ou Docker Compose pour stack complète)

### Backend
```bash
cd backend
go mod download
go run cmd/gateway/main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Stack complète (Docker Compose)
```bash
docker-compose up -d
# Frontend : http://localhost:3000
# Backend : http://localhost:8080
```

---

## 🧪 Tests

### Tests unitaires
```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm test
```

### Tests sécurité
```bash
# OWASP ZAP scan
docker run -t owasp/zap2docker-stable zap-baseline.py -t http://localhost:3000

# Fuzzing AFL++ (corpus 1M inputs)
cd backend/tests/fuzzing && ./run-afl.sh
```

### Load testing (k6)
```bash
k6 run tests/load/100-concurrent-sessions.js
```

---

## 📖 Documentation

- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** : Diagrammes détaillés, flux de données
- **[SECURITY.md](docs/SECURITY.md)** : Threat model, OWASP Top 10:2021 coverage
- **[DEPLOYMENT.md](docs/DEPLOYMENT.md)** : Blue-green, rollback, monitoring
- **[RUNBOOK.md](docs/RUNBOOK.md)** : Incidents courants, procédures escalade

---

## 🔒 Sécurité

**Pentesting :**
- Phase 1 : White-box interne + OWASP ZAP
- Phase 3 : Externe ($3-5K budget) + OWASP ASVS L2

**Compliance :**
- OWASP Top 10:2021 coverage complète
- GDPR : Anonymisation IP après 30j, right to erasure
- Audit logs immuables (PostgreSQL encryption at rest)

**Reporting vulnérabilités :**
- Email : security@shellfrombroswer.local (Phase 3 : bug bounty programme)

---

## 📜 License

[À définir — MIT / Apache 2.0 / Proprietary]

---

## 👥 Contributeurs

- **Maintainer principal :** valorisa
- **Équipe dev :** [À compléter]

---

## 🎓 Méthodologie

Ce projet a été conçu via **Promptor v3.1 Council Edition** avec audit multi-perspective (5 advisors + peer review) ayant identifié 11 angles morts critiques résolus avant Phase 1.

**Artefacts audit disponibles :**
- `council-report-20260514-170128.html`
- `council-transcript-20260514-170128.md`
- `shellfrombroswer-prompt-v2-council.md`

---

**Status :** 🚧 Phase 1 en cours (semaine 1/5)  
**Last update :** 2026-05-14
