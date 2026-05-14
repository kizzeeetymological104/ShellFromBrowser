# ShellFromBrowser Backend Testing Guide

## Quick Start

### Prerequisites

```bash
# Check Go version (1.25+)
go version

# Check Docker daemon
docker version

# Start Docker daemon if not running
# macOS: Start Docker Desktop
# Linux: sudo systemctl start docker
```

### Run Tests

```bash
# All tests (requires Docker daemon)
cd backend
go test -v ./...

# Unit tests only (no Docker required)
go test -short -v ./...

# Specific package
go test -v ./pkg/auth
go test -v ./pkg/redis
go test -v ./pkg/security
go test -v ./pkg/docker

# With coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Integration Testing

### Phase 1 Week 2 - Docker Shell Executor

**Manual test avec Docker daemon actif :**

```bash
# 1. Start gateway
cd backend
go run ./cmd/gateway

# 2. Test WebSocket connection (nouveau terminal)
# TODO Phase 1 Week 3 : Frontend React pour test manuel
# Pour l'instant, test unitaire avec Docker daemon :

# Start Docker Desktop, puis :
go test -v ./pkg/docker -run TestSpawnContainer

# Expected output :
# === RUN   TestSpawnContainer
# INF Pulling Docker image image=alpine:latest
# INF Image pulled successfully image=alpine:latest
# INF Spawning Docker container user_id=test_user session_id=test_session
# INF Container spawned successfully container_id=abc123... user_id=test_user
# INF Stopping container container_id=abc123...
# INF Container removed container_id=abc123...
# --- PASS: TestSpawnContainer (5.32s)
```

**Validation des limites de ressources :**

```bash
# Spawn container et vérifier limites
docker run -d --name test_shell \
  --memory=512m \
  --cpus=0.5 \
  --cap-drop=ALL \
  --cap-add=CHOWN \
  --cap-add=DAC_OVERRIDE \
  --cap-add=FOWNER \
  --cap-add=SETGID \
  --cap-add=SETUID \
  --security-opt=no-new-privileges:true \
  alpine:latest /bin/sh -c "sleep 3600"

# Vérifier les limites appliquées
docker inspect test_shell | jq '.[0].HostConfig.Memory'       # 536870912
docker inspect test_shell | jq '.[0].HostConfig.NanoCpus'     # 500000000
docker inspect test_shell | jq '.[0].HostConfig.CapDrop'      # ["ALL"]
docker inspect test_shell | jq '.[0].HostConfig.CapAdd'       # ["CHOWN","DAC_OVERRIDE",...]
docker inspect test_shell | jq '.[0].HostConfig.SecurityOpt'  # ["no-new-privileges:true"]

# Cleanup
docker stop test_shell && docker rm test_shell
```

### Phase 1 Week 3 - Load Test (10 sessions)

**Test 10 containers concurrents :**

```bash
# Script de charge basique
for i in {1..10}; do
  docker run -d --name shell_$i \
    --memory=512m \
    --cpus=0.5 \
    alpine:latest /bin/sh -c "sleep 600"
done

# Vérifier tous actifs
docker ps | grep shell_

# Mesurer ressources
docker stats --no-stream

# Cleanup
for i in {1..10}; do docker stop shell_$i && docker rm shell_$i; done
```

**Test idle timeout (30 minutes) :**

```bash
# Spawn container via gateway, attendre 30 min, vérifier auto-kill
# TODO : Implémenter script automatique Phase 1 Week 4
```

## Security Testing

### Input Sanitization

```bash
# Test sanitizer
go test -v ./pkg/security -run TestSanitizeInput

# Test dangerous patterns
go test -v ./pkg/security -run TestContainsDangerousPatterns

# Expected : tous les tests passent
```

### Rate Limiting

```bash
# Test rate limiter (10 req/s)
go test -v ./pkg/security -run TestRateLimiter

# Manual test : spawn 20 WebSocket messages en < 1s
# Expected : 10 acceptés, 10 rejetés (429 Too Many Requests)
```

### JWT Authentication

```bash
# Test JWT generation + validation
go test -v ./pkg/auth -run TestGenerateToken
go test -v ./pkg/auth -run TestValidateToken

# Test JWT blacklist
go test -v ./pkg/redis -run TestBlacklistToken
```

## Performance Testing

### Latency P95 (< 500ms spawn container)

```bash
# Benchmark spawn latency
go test -bench=BenchmarkSpawnContainer -benchmem ./pkg/docker

# Expected : P95 < 500ms (Phase 1 contractuel)
```

### Memory Usage

```bash
# Check gateway memory footprint
go build -o bin/gateway ./cmd/gateway
./bin/gateway &
GATEWAY_PID=$!

# Monitor avec pprof
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Cleanup
kill $GATEWAY_PID
```

## CI/CD Pipeline Tests

### GitHub Actions Jobs

**Backend Tests :**
```bash
go test -v ./...
go test -race ./...
```

**Backend Security :**
```bash
# gosec (static analysis)
gosec -fmt sarif -out gosec-results.sarif ./...

# staticcheck (linter)
staticcheck ./...
```

**Docker Build :**
```bash
docker build -t shellfrombroswer-backend:test .
```

**OWASP ZAP Scan :**
```bash
# Baseline scan (Phase 1)
docker run -v $(pwd):/zap/wrk:rw \
  owasp/zap2docker-stable zap-baseline.py \
  -t http://localhost:8080 \
  -r zap-baseline-report.html

# Full scan (Phase 3)
docker run -v $(pwd):/zap/wrk:rw \
  owasp/zap2docker-stable zap-full-scan.py \
  -t http://localhost:8080 \
  -r zap-full-report.html
```

**Dependency Scan (Trivy) :**
```bash
trivy image shellfrombroswer-backend:latest
trivy fs --severity HIGH,CRITICAL .
```

## Debugging

### Enable Debug Logging

```bash
# Set log level
export LOG_LEVEL=debug
go run ./cmd/gateway

# Logs format :
# {"level":"debug","time":"2026-05-14T18:00:00Z","message":"..."}
```

### Docker Logs

```bash
# Gateway logs
docker logs -f shellfrombroswer-backend

# Container shell logs
docker logs -f <container_id>
```

### Prometheus Metrics

```bash
# Check metrics endpoint
curl http://localhost:8080/metrics

# Key metrics :
# - shellfrombroswer_active_sessions (gauge)
# - shellfrombroswer_container_spawn_duration_seconds (histogram)
# - shellfrombroswer_websocket_errors_total (counter)
# - shellfrombroswer_jwt_revocations_total (counter)
# - shellfrombroswer_container_escapes_total (counter)
# - shellfrombroswer_ddos_blocked_total (counter)
# - shellfrombroswer_redis_failover_count (counter)
```

## Next Steps (Phase 1 Week 3-5)

- [ ] **Week 3 :** Frontend React avec xterm.js pour test manuel WebSocket
- [ ] **Week 4 :** OWASP ZAP full scan, AFL++ fuzzing 1M inputs, Grafana dashboards
- [ ] **Week 5 :** Pentest white-box interne, load test k6 100 sessions, DEPLOYMENT.md + RUNBOOK.md

---

**Conseil :** Toujours tester avec Docker daemon actif en local avant de pusher. Les tests CI/CD GitHub Actions s'exécutent dans des runners Linux avec Docker disponible.
