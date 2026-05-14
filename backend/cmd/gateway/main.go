package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valorisa/shellfrombroswer/pkg/auth"
	"github.com/valorisa/shellfrombroswer/pkg/docker"
	"github.com/valorisa/shellfrombroswer/pkg/redis"
	"github.com/valorisa/shellfrombroswer/pkg/security"
)

// ============================================
// WebSocket Gateway — ShellFromBrowser
// ============================================
// Handles WebSocket connections with:
// - JWT authentication middleware
// - Rate limiting (10 req/s per connection)
// - Input sanitization
// - Prometheus metrics

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
)

func main() {
	// Setup structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// Load configuration from environment
	config := loadConfig()
	log.Info().Msgf("Starting ShellFromBrowser Gateway on %s:%s", config.Host, config.Port)

	// Initialize Redis Sentinel client
	redisClient, err := redis.NewSentinelClient(config.RedisSentinels, config.RedisMasterName, config.RedisPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis Sentinel")
	}
	defer redisClient.Close()

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(config.JWTSecret, config.JWTIssuer, config.JWTAudience)

	// Initialize rate limiter
	rateLimiter := security.NewRateLimiter(config.RateLimitPerSecond)

	// Initialize Docker executor
	dockerExec, err := docker.NewExecutor(
		config.DockerImage,
		config.DockerMemoryMB,
		config.DockerCPULimit,
		config.DockerIdleTimeout,
		config.DockerNetworkMode,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Docker executor")
	}
	defer dockerExec.Close()

	// Pull Docker image if needed
	if err := dockerExec.PullImage(context.Background()); err != nil {
		log.Warn().Err(err).Msg("Failed to pull Docker image (will use local)")
	}

	// HTTP routes
	mux := http.NewServeMux()

	// WebSocket endpoint (authenticated)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, jwtManager, redisClient, rateLimiter, dockerExec)
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"gateway"}`))
	})

	// Prometheus metrics
	mux.Handle("/metrics", promhttp.Handler())

	// HTTP server
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	log.Info().Msgf("Gateway listening on %s", addr)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Gateway forced to shutdown")
	}

	log.Info().Msg("Gateway stopped")
}

// handleWebSocket upgrades HTTP to WebSocket and handles terminal session
func handleWebSocket(w http.ResponseWriter, r *http.Request, jwtManager *auth.JWTManager, redisClient *redis.Client, rateLimiter *security.RateLimiter, dockerExec *docker.Executor) {
	// 1. Authenticate JWT from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Warn().Err(err).Msg("Missing session token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := jwtManager.ValidateToken(cookie.Value)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid JWT")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Check JWT blacklist (revoked tokens)
	isBlacklisted, err := redisClient.IsTokenBlacklisted(r.Context(), cookie.Value)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check token blacklist")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if isBlacklisted {
		log.Warn().Str("user_id", claims.UserID).Msg("Token is blacklisted")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 3. Rate limiting check
	if !rateLimiter.Allow(claims.UserID) {
		log.Warn().Str("user_id", claims.UserID).Msg("Rate limit exceeded")
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// 4. Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket")
		return
	}
	defer conn.Close()

	log.Info().Str("user_id", claims.UserID).Msg("WebSocket connection established")

	// 5. Spawn Docker container for shell session
	sessionID := fmt.Sprintf("%s-%d", claims.UserID, time.Now().Unix())
	containerConfig := docker.ContainerConfig{
		UserID:      claims.UserID,
		SessionID:   sessionID,
		WorkingDir:  "/workspace",
		Environment: []string{fmt.Sprintf("USER=%s", claims.Username)},
	}

	containerID, err := dockerExec.SpawnContainer(r.Context(), containerConfig)
	if err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID).Msg("Failed to spawn container")
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Failed to create shell session"))
		conn.Close()
		return
	}

	// Store session metadata in Redis
	metadata := redis.SessionMetadata{
		UserID:       claims.UserID,
		Username:     claims.Username,
		ContainerID:  containerID,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		DeviceInfo:   r.UserAgent(),
	}
	if err := redisClient.StoreSession(r.Context(), sessionID, metadata, 30*time.Minute); err != nil {
		log.Error().Err(err).Msg("Failed to store session metadata")
	}

	// 6. Bridge WebSocket <-> Docker PTY
	handleTerminalSession(conn, containerID, claims.UserID, dockerExec, redisClient, sessionID)
}

// handleTerminalSession manages terminal I/O via WebSocket <-> Docker PTY bridge
func handleTerminalSession(conn *websocket.Conn, containerID, userID string, dockerExec *docker.Executor, redisClient *redis.Client, sessionID string) {
	defer func() {
		// Cleanup on disconnect
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Info().Str("user_id", userID).Str("container_id", containerID[:12]).Msg("Cleaning up session")

		// Stop container
		if err := dockerExec.StopContainer(ctx, containerID); err != nil {
			log.Error().Err(err).Str("container_id", containerID[:12]).Msg("Failed to stop container")
		}

		// Delete session metadata
		if err := redisClient.DeleteSession(ctx, sessionID); err != nil {
			log.Error().Err(err).Str("session_id", sessionID).Msg("Failed to delete session")
		}

		conn.Close()
	}()

	// Create PTY bridge
	bridge := docker.NewPTYBridge(conn, dockerExec.GetClient(), containerID, userID)

	// Start bridge (blocks until done)
	ctx := context.Background()
	if err := bridge.Start(ctx); err != nil {
		log.Error().Err(err).Str("container_id", containerID[:12]).Msg("PTY bridge error")
		return
	}

	log.Info().Str("user_id", userID).Str("container_id", containerID[:12]).Msg("Session ended")
}

// checkOrigin validates CORS origin
func checkOrigin(r *http.Request) bool {
	// TODO: Load allowed origins from config
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://shellfrombroswer.local",
	}

	origin := r.Header.Get("Origin")
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	log.Warn().Str("origin", origin).Msg("CORS origin rejected")
	return false
}

// Config holds application configuration
type Config struct {
	Host               string
	Port               string
	RedisSentinels     []string
	RedisMasterName    string
	RedisPassword      string
	JWTSecret          string
	JWTIssuer          string
	JWTAudience        string
	RateLimitPerSecond int
	DockerImage        string
	DockerMemoryMB     int
	DockerCPULimit     float64
	DockerIdleTimeout  time.Duration
	DockerNetworkMode  string
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	return &Config{
		Host:               getEnv("SERVER_HOST", "0.0.0.0"),
		Port:               getEnv("SERVER_PORT", "8080"),
		RedisSentinels:     parseRedisSentinels(getEnv("REDIS_SENTINELS", "localhost:26379")),
		RedisMasterName:    getEnv("REDIS_MASTER_NAME", "mymaster"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		JWTSecret:          getEnv("JWT_SECRET", "CHANGE_ME"),
		JWTIssuer:          getEnv("JWT_ISSUER", "shellfrombroswer"),
		JWTAudience:        getEnv("JWT_AUDIENCE", "shellfrombroswer-users"),
		RateLimitPerSecond: 10,
		DockerImage:        getEnv("DOCKER_IMAGE", "ubuntu:22.04"),
		DockerMemoryMB:     512,
		DockerCPULimit:     0.5,
		DockerIdleTimeout:  30 * time.Minute,
		DockerNetworkMode:  getEnv("DOCKER_NETWORK_MODE", "bridge"),
	}
}

// getEnv retrieves environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// parseRedisSentinels parses comma-separated sentinel addresses
func parseRedisSentinels(input string) []string {
	// Simple split for Phase 1 (improve error handling later)
	return []string{input} // TODO: strings.Split(input, ",")
}
