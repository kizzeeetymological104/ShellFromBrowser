package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/valorisa/ShellFromBrowser/internal/config"
)

func TestLoadConfig(t *testing.T) {
	yaml := `
server:
  addr: ":9090"
  tls:
    enabled: true
    cert: "/etc/ssl/cert.pem"
    key: "/etc/ssl/key.pem"
auth:
  enabled: true
  users:
    - username: admin
      password_hash: "$2a$10$abcdefghijklmnopqrstuuxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
shell:
  command: "/bin/bash"
  env:
    - "TERM=xterm-256color"
sessions:
  max_per_user: 5
  idle_timeout: "30m"
ssh:
  enabled: true
  known_hosts: "~/.ssh/known_hosts"
recording:
  enabled: true
  dir: "./recordings"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(yaml), 0644)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Server.Addr != ":9090" {
		t.Errorf("addr = %q, want :9090", cfg.Server.Addr)
	}
	if !cfg.Server.TLS.Enabled {
		t.Error("TLS should be enabled")
	}
	if !cfg.Auth.Enabled {
		t.Error("Auth should be enabled")
	}
	if len(cfg.Auth.Users) != 1 {
		t.Errorf("users = %d, want 1", len(cfg.Auth.Users))
	}
	if cfg.Sessions.MaxPerUser != 5 {
		t.Errorf("max_per_user = %d, want 5", cfg.Sessions.MaxPerUser)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg := config.Default()
	if cfg.Server.Addr != ":8080" {
		t.Errorf("default addr = %q, want :8080", cfg.Server.Addr)
	}
	if cfg.Auth.Enabled {
		t.Error("auth should be disabled by default")
	}
}
