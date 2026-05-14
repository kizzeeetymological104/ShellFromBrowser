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

	// HTTP routes
	mux := http.NewServeMux()

	// WebSocket endpoint (authenticated)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, jwtManager, redisClient, rateLimiter)
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
func handleWebSocket(w http.ResponseWriter, r *http.Request, jwtManager *auth.JWTManager, redisClient *redis.Client, rateLimiter *security.RateLimiter) {
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

	// 5. Handle terminal session (to implement in Phase 1 Week 2)
	handleTerminalSession(conn, claims.UserID, redisClient)
}

// handleTerminalSession manages terminal I/O via WebSocket
func handleTerminalSession(conn *websocket.Conn, userID string, redisClient *redis.Client) {
	// TODO Phase 1 Week 2: Implement Docker container spawn + PTY attach
	// For now, echo back messages (skeleton)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Str("user_id", userID).Msg("WebSocket read error")
			}
			break
		}

		log.Debug().Str("user_id", userID).Str("message", string(p)).Msg("Received message")

		// Echo back (temporary skeleton)
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("WebSocket write error")
			break
		}
	}

	log.Info().Str("user_id", userID).Msg("WebSocket connection closed")
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
		RateLimitPerSecond: 10, // Fixed for Phase 1
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
