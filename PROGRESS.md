# ShellFromBrowser - Progress Tracker

> **Last Updated:** 2026-05-14 18:30 GMT+2  
> **Current Phase:** Phase 1 Week 3 (Frontend React + xterm.js)  
> **Status:** Backend + Frontend UI complets, intégration backend↔frontend à faire

---

## 📊 État Actuel du Projet

### ✅ Phases Complétées

#### **Phase 0 : Council Audit (2026-05-14 après-midi)**
- ✅ Audit LLM Council avec 5 advisors + peer review
- ✅ 11 angles morts détectés et résolus (Redis SPOF, threat model, JWT revocation, GDPR, etc.)
- ✅ Décisions contractuelles fixées (volumétrie 100, budget $500/mois, timeline 13 sem)
- ✅ Prompt système v2 généré production-ready
- **Artefacts :**
  - `council-report-20260514-170128.html`
  - `council-transcript-20260514-170128.md`
  - `shellfrombroswer-prompt-v2-council.md`

#### **Phase 1 Week 1 : Backend Core (2026-05-14 soir)**
- ✅ WebSocket Gateway (`cmd/gateway/main.go`) : HTTP/WS server, auth JWT, rate limiting, CORS, Prometheus metrics
- ✅ JWT Manager (`pkg/auth/jwt.go`) : HS256, HttpOnly cookies, refresh tokens
- ✅ Redis Sentinel Client (`pkg/redis/client.go`) : sessions HA, blacklist JWT, "Kill All My Sessions"
- ✅ Rate Limiter (`pkg/security/ratelimiter.go`) : token bucket 10 req/s per-user
- ✅ Input Sanitizer (`pkg/security/sanitizer.go`) : ANSI escape, Unicode RLO, shell metacharacters, 6-step pipeline
- ✅ Docker Executor stub (echo mode)
- **Commits :** 3 initiaux

#### **Phase 1 Week 2 : Docker Shell Executor (2026-05-14 soir)**
- ✅ **Problème résolu :** Dépendance docker/docker SDK (v27 incompatible macOS arm64 → v28 OK)
- ✅ Full implementation activée :
  - `pkg/docker/executor.go` (278 lignes) : container lifecycle (spawn, attach, stop, remove, stats, list)
  - `pkg/docker/pty.go` (166 lignes) : PTY bridge bidirectionnel (WebSocket ↔ Docker stdin/stdout/stderr)
  - `pkg/docker/executor_test.go` (125 lignes) : unit tests
- ✅ Security isolation : capabilities drop ALL + add 5 minimal (CHOWN, DAC_OVERRIDE, FOWNER, SETGID, SETUID)
- ✅ Resource limits : 512MB RAM, 0.5 CPU, idle timeout 30 min
- ✅ Build réussi : binary `backend/bin/gateway` 15MB
- ✅ Tests validés : `go test -short -v ./pkg/docker` (PASS)
- ✅ Documentation : `backend/TESTING.md` (guide complet)
- **Commits :** ba6b188, a6d71c4

#### **Phase 1 Week 3 : Frontend React + xterm.js (2026-05-14 soir)**
- ✅ Stack : Vite 5.2 + React 18.3 + TypeScript 5.4 strict
- ✅ xterm.js 5.5 + FitAddon + WebLinksAddon
- ✅ Composants :
  - `App.tsx` : orchestration auth, state management
  - `Login.tsx` : UI auth moderne (mock pour MVP)
  - `Terminal.tsx` : terminal emulator complet, WebSocket client, reconnect logic, status indicators
- ✅ Build production : 124KB gzipped (target <230KB ✅)
- ✅ WebSocket proxy Vite : `/api` → `http://localhost:8080`, `/ws` → `ws://localhost:8080`
- ✅ Documentation : `frontend/README.md`
- **Commits :** 472e9b3

### ⏳ Phase 1 Week 3 Suite (En Cours)

**Ce qui manque pour compléter Week 3 :**

1. **Intégration Backend ↔ Frontend (Priorité 1)**
   - [ ] Connecter Login.tsx → API backend `/api/auth/login`
   - [ ] Remplacer mock auth par JWT HttpOnly cookie réel
   - [ ] Implémenter refresh token logic
   - [ ] Gérer erreurs auth (401, 403)

2. **Test End-to-End WebSocket + Docker PTY (Priorité 2)**
   - [ ] Lancer backend gateway : `cd backend && go run ./cmd/gateway`
   - [ ] Lancer frontend dev : `cd frontend && npm run dev`
   - [ ] Tester login → spawn container → commandes shell réelles
   - [ ] Vérifier PTY bidirectionnel (input/output)
   - [ ] Tester terminal resize
   - [ ] Tester idle timeout (30 min)

3. **Session Management UI Preview (Priorité 3)**
   - [ ] Endpoint backend `/api/sessions` (liste containers actifs user)
   - [ ] UI frontend : afficher sessions actives
   - [ ] Bouton "Kill All My Sessions" → API `/api/sessions/kill-all`
   - [ ] Multi-tabs preview (Phase 2 complet)

### 🔜 Phases Suivantes

#### **Phase 1 Week 4 : Monitoring + Security Testing (5 jours)**
- [ ] **Prometheus metrics implementation** (7 metrics définies dans ARCHITECTURE.md)
  - `shellfrombroswer_active_sessions` (gauge)
  - `shellfrombroswer_container_spawn_duration_seconds` (histogram)
  - `shellfrombroswer_websocket_errors_total` (counter)
  - `shellfrombroswer_jwt_revocations_total` (counter)
  - `shellfrombroswer_container_escapes_total` (counter)
  - `shellfrombroswer_ddos_blocked_total` (counter)
  - `shellfrombroswer_redis_failover_count` (counter)
- [ ] **Grafana dashboards deployment**
  - Dashboard session overview
  - Dashboard security events
  - Dashboard performance (latency P95, throughput)
- [ ] **OWASP ZAP scanning**
  - Baseline scan (automated)
  - Full scan (manual review)
  - Rapport SARIF upload GitHub Security
- [ ] **AFL++ fuzzing setup**
  - Corpus 1M inputs
  - Target : input sanitizer
  - Coverage report

#### **Phase 1 Week 5 : Testing + Documentation (5 jours)**
- [ ] **Pentest white-box interne**
  - Threat model 3 attackers (A/B/C)
  - OWASP Top 10:2021 coverage validation
  - Container escape attempts
  - JWT blacklist bypass attempts
- [ ] **Load test k6**
  - 100 sessions concurrentes (contractuel)
  - Latency P95 < 500ms spawn container (contractuel)
  - Resource usage monitoring
- [ ] **Documentation finale**
  - `DEPLOYMENT.md` : blue-green deployment strategy
  - `RUNBOOK.md` : incident response procedures
  - `CHANGELOG.md` : version tracking

---

## 🛠️ Instructions pour Reprendre le Travail

### Prérequis

**Backend :**
```bash
cd /Users/valorisa/Projets/ShellFromBrowser/backend
go version  # 1.25.0
go mod tidy
go build ./...
```

**Frontend :**
```bash
cd /Users/valorisa/Projets/ShellFromBrowser/frontend
node --version  # v20.18.1
npm install
npm run build
```

**Docker :**
```bash
docker version  # Daemon must be running
# macOS: Start Docker Desktop
```

### Lancer l'Application End-to-End

**Terminal 1 - Backend :**
```bash
cd backend

# Start Redis Sentinel + PostgreSQL (Docker Compose)
docker-compose up -d redis-sentinel-1 redis-sentinel-2 redis-sentinel-3 \
  redis-master redis-replica-1 redis-replica-2 postgres

# Vérifier Redis Sentinel HA
docker exec redis-sentinel-1 redis-cli -p 26379 SENTINEL MASTER mymaster

# Lancer le gateway
export JWT_SECRET="dev-secret-change-in-prod-32bytes-min"
export REDIS_SENTINELS="localhost:26379,localhost:26380,localhost:26381"
export REDIS_MASTER_NAME="mymaster"
export DOCKER_IMAGE="alpine:latest"
export DOCKER_MEMORY_LIMIT="512"
export DOCKER_CPU_LIMIT="0.5"
export WS_RATE_LIMIT_PER_SECOND="10"

go run ./cmd/gateway
# Serveur lancé sur http://localhost:8080
```

**Terminal 2 - Frontend :**
```bash
cd frontend
npm run dev
# Dev server lancé sur http://localhost:3000
```

**Terminal 3 - Tests :**
```bash
# Test backend
cd backend
go test -v ./...

# Test Docker executor (require Docker daemon)
go test -v ./pkg/docker

# Check gateway health
curl http://localhost:8080/health

# Check Prometheus metrics
curl http://localhost:8080/metrics
```

**Browser :**
1. Ouvrir `http://localhost:3000`
2. Login avec n'importe quel username/password (mock auth)
3. Terminal devrait afficher "Connected!"
4. Taper des commandes → elles seront envoyées au backend WebSocket

### Debugging

**Logs backend structurés (Zerolog) :**
```bash
export LOG_LEVEL=debug
go run ./cmd/gateway
```

**Logs Docker containers :**
```bash
# Liste containers actifs
docker ps | grep shellfrombroswer

# Logs d'un container shell
docker logs -f <container_id>

# Inspect container (limits, security)
docker inspect <container_id> | jq '.[0].HostConfig'
```

**Vérifier Redis Sentinel :**
```bash
# Master actuel
docker exec redis-sentinel-1 redis-cli -p 26379 SENTINEL GET-MASTER-ADDR-BY-NAME mymaster

# Replicas
docker exec redis-sentinel-1 redis-cli -p 26379 SENTINEL REPLICAS mymaster

# Forcer failover (test)
docker exec redis-sentinel-1 redis-cli -p 26379 SENTINEL FAILOVER mymaster
```

---

## 📋 Checklist Phase 1 Week 3 Suite

### Intégration Backend Auth

- [ ] **Créer endpoint `/api/auth/login` (backend)**
  - Fichier : `backend/pkg/auth/handlers.go` (nouveau)
  - Route : `POST /api/auth/login` → `{"username": "...", "password": "..."}`
  - Validation : bcrypt password hash (mock pour MVP : hardcoded user/pass)
  - Response : JWT HttpOnly cookie + `{"user_id": "...", "username": "...", "role": "..."}`
  - Erreurs : 401 si credentials invalides

- [ ] **Créer endpoint `/api/auth/logout` (backend)**
  - Route : `POST /api/auth/logout`
  - Action : blacklist JWT + clear cookie
  - Response : 200 OK

- [ ] **Créer endpoint `/api/auth/refresh` (backend)**
  - Route : `POST /api/auth/refresh`
  - Validation : JWT refresh token
  - Response : nouveau access token (cookie)

- [ ] **Modifier `Login.tsx` (frontend)**
  - Remplacer mock auth par `fetch('/api/auth/login', {method: 'POST', body: JSON.stringify({username, password})})`
  - Gérer erreurs 401 (afficher message)
  - Cookie HttpOnly sera set automatiquement par backend

- [ ] **Modifier `App.tsx` (frontend)**
  - Logout : `fetch('/api/auth/logout', {method: 'POST'})`
  - Refresh token : intercepter 401 WebSocket → call `/api/auth/refresh`

### Test End-to-End

- [ ] **Scénario 1 : Login + Shell**
  1. Login avec user/pass valid → JWT cookie set
  2. WebSocket connect → backend valide JWT
  3. Container spawn → logs backend "Spawning Docker container"
  4. Terminal display → "$ " prompt
  5. Commande `ls /` → output liste fichiers Alpine
  6. Commande `echo "test"` → output "test"

- [ ] **Scénario 2 : Rate Limiting**
  1. Spammer 20 commandes en < 1s
  2. Après 10 commandes → WebSocket error 429 Too Many Requests
  3. Attendre 1s → rate limit reset

- [ ] **Scénario 3 : Idle Timeout**
  1. Spawn container
  2. Attendre 30 min (ou modifier `idleTimeout` pour test)
  3. Container auto-killed → logs backend "Idle timeout reached"
  4. WebSocket close → frontend affiche "Connection closed"

- [ ] **Scénario 4 : Terminal Resize**
  1. Spawn container
  2. Resize browser window
  3. FitAddon recalcule rows/cols
  4. WebSocket envoie resize event → backend appelle `ContainerResize()`
  5. Terminal ajusté

### Session Management Preview

- [ ] **Endpoint `/api/sessions` (backend)**
  - GET `/api/sessions` → liste containers actifs pour user JWT
  - Response : `[{"container_id": "...", "session_id": "...", "created_at": "..."}]`

- [ ] **Endpoint `/api/sessions/kill-all` (backend)**
  - POST `/api/sessions/kill-all` → stop tous containers user
  - Action : `redis.RevokeAllUserSessions()` + Docker stop containers
  - Response : `{"killed": 3}`

- [ ] **UI Sessions (frontend)**
  - Composant `SessionList.tsx` (nouveau)
  - Afficher liste sessions actives
  - Bouton "Kill" par session
  - Bouton "Kill All"

---

## 📚 Documentation Importante

### System Prompt (LIRE EN PREMIER)
- **shellfrombroswer-prompt-v2-council.md** : `/Users/valorisa/Projets/ShellFromBrowser/shellfrombroswer-prompt-v2-council.md`
  - **Prompt système production-ready post-Council audit**
  - Décisions contractuelles non-négociables (volumétrie, budget, timeline)
  - Threat model 3 attackers explicites (A/B/C) avec mitigations
  - OWASP Top 10:2021 coverage 10/10
  - GDPR compliance (endpoints, retention, anonymization)
  - 11 angles morts résolus (Redis SPOF, JWT revocation, monitoring, etc.)
  - Anti-patterns à éviter (7 erreurs communes)
  - Checklist pre-production 40+ items
  - **👉 LIRE CE FICHIER AVANT TOUTE IMPLÉMENTATION**

### Architecture
- **ARCHITECTURE.md** : `/Users/valorisa/Projets/ShellFromBrowser/docs/ARCHITECTURE.md`
  - 15 sections : system overview, 4 composants, data flow, stack technique, infra dev/prod
  - ADRs (Architecture Decision Records) : Go vs Node.js, Redis Sentinel vs Cluster, Docker vs namespaces, JWT cookies vs Bearer

### Sécurité
- **SECURITY.md** : `/Users/valorisa/Projets/ShellFromBrowser/docs/SECURITY.md`
  - 12 sections : threat model 3 attackers (A/B/C), OWASP Top 10:2021 coverage, container security 5 layers
  - Input sanitization 6-step pipeline
  - GDPR compliance (export/erase endpoints)
  - Incident response runbook container escape

### Testing
- **TESTING.md** : `/Users/valorisa/Projets/ShellFromBrowser/backend/TESTING.md`
  - Guide complet : unit tests, integration tests, load tests, security tests
  - CI/CD pipeline (GitHub Actions 6 jobs)
  - Debugging (logs, Docker, Prometheus metrics)

### Frontend
- **frontend/README.md** : `/Users/valorisa/Projets/ShellFromBrowser/frontend/README.md`
  - Setup, dev server, structure, roadmap
  - Performance targets (bundle <230KB gzipped ✅)
  - Security notes (CSP, XSS, CORS)

### Memory
- **project_shellfrombroswer.md** : `/Users/valorisa/.claude/projects/-Users-valorisa-Projets-Super-Prompt-LLM-Council-Generator/memory/project_shellfrombroswer.md`
  - Contexte projet complet
  - Council audit 11 angles morts
  - Décisions contractuelles
  - Timeline phases
  - Étapes réalisées

---

## 🎯 Objectifs Immédiats (Prochaine Session)

### Priorité 1 : Backend Auth Endpoints (1-2h)
1. Créer `backend/pkg/auth/handlers.go`
2. Implémenter `/api/auth/login` (bcrypt password, JWT generation)
3. Implémenter `/api/auth/logout` (JWT blacklist)
4. Implémenter `/api/auth/refresh`
5. Router dans `cmd/gateway/main.go`

### Priorité 2 : Frontend Auth Integration (1h)
1. Modifier `Login.tsx` → fetch `/api/auth/login`
2. Modifier `App.tsx` → fetch `/api/auth/logout`
3. Gérer erreurs 401/403

### Priorité 3 : Test End-to-End (1h)
1. Lancer backend + frontend
2. Test login → shell → commandes
3. Test rate limiting
4. Test terminal resize

### Priorité 4 : Session Management Preview (1h)
1. Endpoint `/api/sessions` + `/api/sessions/kill-all`
2. UI `SessionList.tsx` basique

**Temps estimé total : 4-5h** → Phase 1 Week 3 complète ✅

---

## 🔗 Ressources

- **Repository GitHub :** https://github.com/valorisa/ShellFromBrowser
- **Commits récents :**
  - `ba6b188` — Phase 1 Week 2 Docker Shell Executor
  - `a6d71c4` — Testing guide
  - `472e9b3` — Phase 1 Week 3 Frontend React
- **Stack technique :**
  - Backend : Go 1.25, docker/docker v28, Redis Sentinel, JWT HS256
  - Frontend : React 18.3, TypeScript 5.4, Vite 5.2, xterm.js 5.5
- **Threat Model :** 3 attackers (A: compromised user, B: malicious insider, C: external RCE)
- **Security :** OWASP Top 10:2021 compliance, container 5-layer isolation, input sanitization 6-step
- **Monitoring :** Prometheus 7 metrics (à implémenter Week 4)

---

## 🚨 Points d'Attention

### Connus
- **Docker daemon required** pour tests integration (backend)
- **Redis Sentinel HA** : 3 sentinels + 1 master + 2 replicas (Docker Compose)
- **Mock auth active** : n'importe quel username/password accepté (frontend)
- **WebSocket mock mode** : backend stub echo mode si Docker pas disponible

### Risques
- **JWT secret** : utiliser env var `JWT_SECRET` (32 bytes min production)
- **CORS** : strict en production (HAProxy config), permissif en dev (Vite proxy)
- **Rate limiting** : 10 req/s contractuel, tester avec spam
- **Idle timeout** : 30 min contractuel, peut causer tests longs

### Prochains Bloquants Potentiels
- **Bcrypt password hashing** : ajouter package `golang.org/x/crypto/bcrypt`
- **User DB** : pour l'instant hardcoded (Phase 2 : PostgreSQL users table)
- **MFA** : Phase 3 uniquement
- **OAuth2 SSO** : Phase 3 uniquement

---

## 📞 Contact / Handoff

**Session 2026-05-14 :**
- Agent : Claude Sonnet 4.5
- Durée : ~2h30
- Phases complétées : Phase 1 Week 1, 2, 3 (partiel)
- Commits : 6 totaux

**Prochaine session :**
- Reprendre avec `PROGRESS.md` (ce fichier)
- Vérifier `memory/project_shellfrombroswer.md` pour contexte complet
- Lire `ARCHITECTURE.md` et `SECURITY.md` si besoin clarification
- Commencer par **Priorité 1 : Backend Auth Endpoints**

**Questions ? Consulter :**
- `backend/TESTING.md` pour debugging
- `frontend/README.md` pour dev frontend
- `docs/ARCHITECTURE.md` pour architecture
- `docs/SECURITY.md` pour threat model

---

**Bonne continuation ! 🚀**

_Phase 1 MVP : 13 semaines restantes (Week 3 en cours)_  
_Budget : $500/mois AWS, 100 sessions simultanées_  
_Pentest externe obligatoire $3-5K avant production_
