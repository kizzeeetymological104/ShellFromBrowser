package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// ============================================
// Docker Executor Stub (Phase 1 Week 1)
// ============================================
// Temporary stub implementation while resolving Docker SDK dependency issues.
// TODO Phase 1 Week 2: Replace with full executor.go implementation once
// docker/docker@v27 compatibility issues are resolved.

type Executor struct {
	image       string
	memoryLimit int64
	cpuLimit    int64
	idleTimeout time.Duration
	networkMode string
}

type ContainerConfig struct {
	UserID      string
	SessionID   string
	WorkingDir  string
	Environment []string
}

func NewExecutor(image string, memoryMB int, cpuLimit float64, idleTimeout time.Duration, networkMode string) (*Executor, error) {
	log.Warn().Msg("Using Docker stub implementation (Phase 1 Week 1)")
	return &Executor{
		image:       image,
		memoryLimit: int64(memoryMB) * 1024 * 1024,
		cpuLimit:    int64(cpuLimit * 1e9),
		idleTimeout: idleTimeout,
		networkMode: networkMode,
	}, nil
}

func (e *Executor) Close() error {
	return nil
}

func (e *Executor) GetClient() interface{} {
	return nil
}

func (e *Executor) SpawnContainer(ctx context.Context, config ContainerConfig) (string, error) {
	// Stub: return fake container ID
	containerID := fmt.Sprintf("stub-%s-%d", config.UserID, time.Now().Unix())
	log.Info().
		Str("container_id", containerID).
		Str("user_id", config.UserID).
		Msg("Stub: Container spawn simulated")
	return containerID, nil
}

func (e *Executor) StopContainer(ctx context.Context, containerID string) error {
	log.Info().Str("container_id", containerID).Msg("Stub: Container stop simulated")
	return nil
}

func (e *Executor) PullImage(ctx context.Context) error {
	log.Info().Str("image", e.image).Msg("Stub: Image pull simulated")
	return nil
}

// PTY Bridge Stub
type PTYBridge struct {
	ws          *websocket.Conn
	containerID string
	userID      string
	done        chan struct{}
}

func NewPTYBridge(ws *websocket.Conn, dockerClient interface{}, containerID, userID string) *PTYBridge {
	return &PTYBridge{
		ws:          ws,
		containerID: containerID,
		userID:      userID,
		done:        make(chan struct{}),
	}
}

func (b *PTYBridge) Start(ctx context.Context) error {
	log.Info().
		Str("container_id", b.containerID).
		Str("user_id", b.userID).
		Msg("Stub: PTY bridge started (echo mode)")

	// Echo mode for Phase 1 Week 1
	for {
		select {
		case <-b.done:
			return nil
		default:
		}

		_, message, err := b.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("WebSocket read error")
			}
			return err
		}

		// Echo back with prefix
		response := fmt.Sprintf("[STUB ECHO] %s", string(message))
		if err := b.ws.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
			log.Error().Err(err).Msg("WebSocket write error")
			return err
		}
	}
}

func (b *PTYBridge) Stop() {
	close(b.done)
}

func (b *PTYBridge) ResizePTY(ctx context.Context, height, width uint) error {
	log.Debug().Uint("height", height).Uint("width", width).Msg("Stub: PTY resize simulated")
	return nil
}
