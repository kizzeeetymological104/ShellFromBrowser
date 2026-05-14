package docker

import (
	"context"
	"testing"
	"time"
)

// TestNewExecutor verifies executor creation
func TestNewExecutor(t *testing.T) {
	// Skip if Docker not available
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	executor, err := NewExecutor(
		"alpine:latest",
		512, // 512MB
		0.5, // 0.5 CPU
		30*time.Minute,
		"none",
	)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer executor.Close()

	// Verify configuration
	if executor.image != "alpine:latest" {
		t.Errorf("Expected image alpine:latest, got %s", executor.image)
	}
	if executor.memoryLimit != 512*1024*1024 {
		t.Errorf("Expected memory limit 536870912, got %d", executor.memoryLimit)
	}
	if executor.cpuLimit != 500000000 {
		t.Errorf("Expected CPU limit 500000000, got %d", executor.cpuLimit)
	}
	if executor.idleTimeout != 30*time.Minute {
		t.Errorf("Expected idle timeout 30m, got %v", executor.idleTimeout)
	}
}

// TestContainerConfig verifies container configuration structure
func TestContainerConfig(t *testing.T) {
	config := ContainerConfig{
		UserID:     "user123",
		SessionID:  "session456",
		WorkingDir: "/home/user",
		Environment: []string{
			"USER=testuser",
			"SHELL=/bin/bash",
		},
	}

	if config.UserID != "user123" {
		t.Errorf("Expected UserID user123, got %s", config.UserID)
	}
	if config.SessionID != "session456" {
		t.Errorf("Expected SessionID session456, got %s", config.SessionID)
	}
	if len(config.Environment) != 2 {
		t.Errorf("Expected 2 environment variables, got %d", len(config.Environment))
	}
}

// TestSpawnContainer tests container spawn (requires Docker daemon)
func TestSpawnContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	executor, err := NewExecutor(
		"alpine:latest",
		512,
		0.5,
		30*time.Minute,
		"none",
	)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Pull image first
	if err := executor.PullImage(ctx); err != nil {
		t.Fatalf("Failed to pull image: %v", err)
	}

	// Spawn container
	config := ContainerConfig{
		UserID:     "test_user",
		SessionID:  "test_session",
		WorkingDir: "/",
		Environment: []string{
			"USER=testuser",
		},
	}

	containerID, err := executor.SpawnContainer(ctx, config)
	if err != nil {
		t.Fatalf("Failed to spawn container: %v", err)
	}

	if containerID == "" {
		t.Fatal("Container ID should not be empty")
	}

	// Cleanup
	if err := executor.StopContainer(ctx, containerID); err != nil {
		t.Errorf("Failed to stop container: %v", err)
	}
}

// TestResourceLimits verifies resource limits are applied
func TestResourceLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	executor, err := NewExecutor(
		"alpine:latest",
		256, // 256MB
		0.25, // 0.25 CPU
		10*time.Minute,
		"none",
	)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer executor.Close()

	// Verify limits are set correctly
	if executor.memoryLimit != 256*1024*1024 {
		t.Errorf("Expected memory limit 268435456, got %d", executor.memoryLimit)
	}
	if executor.cpuLimit != 250000000 {
		t.Errorf("Expected CPU limit 250000000, got %d", executor.cpuLimit)
	}
}
