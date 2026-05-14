package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// ============================================
// Redis Sentinel Client — HA Session Store
// ============================================
// Provides high-availability session storage with:
// - Automatic failover (master + 2 replicas)
// - JWT blacklist (revoked tokens)
// - Session metadata storage

// Client wraps Redis Sentinel client
type Client struct {
	client *redis.Client
}

// NewSentinelClient creates a new Redis Sentinel client
func NewSentinelClient(sentinels []string, masterName, password string) (*Client, error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinels,
		Password:      password,
		DB:            0,
		DialTimeout:   5 * time.Second,
		ReadTimeout:   3 * time.Second,
		WriteTimeout:  3 * time.Second,
		PoolSize:      10,
		MinIdleConns:  5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis Sentinel: %w", err)
	}

	return &Client{client: client}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// ============================================
// JWT Blacklist Operations
// ============================================

// BlacklistToken adds a JWT token to the blacklist (revocation)
func (c *Client) BlacklistToken(ctx context.Context, token string, expiry time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", token)
	return c.client.Set(ctx, key, "revoked", expiry).Err()
}

// IsTokenBlacklisted checks if a JWT token is blacklisted
func (c *Client) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// ============================================
// Session Metadata Operations
// ============================================

// SessionMetadata holds session information
type SessionMetadata struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	ContainerID  string    `json:"container_id"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
	DeviceInfo   string    `json:"device_info"`
}

// StoreSession stores session metadata
func (c *Client) StoreSession(ctx context.Context, sessionID string, metadata SessionMetadata, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)

	data := map[string]interface{}{
		"user_id":       metadata.UserID,
		"username":      metadata.Username,
		"container_id":  metadata.ContainerID,
		"created_at":    metadata.CreatedAt.Unix(),
		"last_activity": metadata.LastActivity.Unix(),
		"device_info":   metadata.DeviceInfo,
	}

	pipe := c.client.Pipeline()
	pipe.HSet(ctx, key, data)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

// GetSession retrieves session metadata
func (c *Client) GetSession(ctx context.Context, sessionID string) (*SessionMetadata, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	data, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("session not found")
	}

	// Parse timestamps (simplified for Phase 1)
	createdAt := time.Now() // TODO: parse data["created_at"]
	lastActivity := time.Now()

	return &SessionMetadata{
		UserID:       data["user_id"],
		Username:     data["username"],
		ContainerID:  data["container_id"],
		CreatedAt:    createdAt,
		LastActivity: lastActivity,
		DeviceInfo:   data["device_info"],
	}, nil
}

// DeleteSession removes session metadata
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return c.client.Del(ctx, key).Err()
}

// UpdateSessionActivity updates last activity timestamp
func (c *Client) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return c.client.HSet(ctx, key, "last_activity", time.Now().Unix()).Err()
}

// ============================================
// User Sessions Management (Kill All)
// ============================================

// GetUserSessions retrieves all active sessions for a user
func (c *Client) GetUserSessions(ctx context.Context, userID string) ([]string, error) {
	pattern := "session:*"
	var sessions []string

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// Check if session belongs to this user
		sessionUserID, err := c.client.HGet(ctx, key, "user_id").Result()
		if err != nil {
			continue
		}

		if sessionUserID == userID {
			// Extract session ID from key
			sessionID := key[8:] // Remove "session:" prefix
			sessions = append(sessions, sessionID)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// RevokeAllUserSessions revokes all active sessions for a user (Kill All My Sessions)
func (c *Client) RevokeAllUserSessions(ctx context.Context, userID string) (int, error) {
	sessions, err := c.GetUserSessions(ctx, userID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, sessionID := range sessions {
		if err := c.DeleteSession(ctx, sessionID); err == nil {
			count++
		}
	}

	return count, nil
}
