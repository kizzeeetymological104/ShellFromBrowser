package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Auth      AuthConfig      `yaml:"auth"`
	Shell     ShellConfig     `yaml:"shell"`
	Sessions  SessionsConfig  `yaml:"sessions"`
	SSH       SSHConfig       `yaml:"ssh"`
	Recording RecordingConfig `yaml:"recording"`
}

type ServerConfig struct {
	Addr string    `yaml:"addr"`
	TLS  TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	Enabled bool   `yaml:"enabled"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
}

type AuthConfig struct {
	Enabled   bool      `yaml:"enabled"`
	Users     []UserDef `yaml:"users"`
	JWTSecret string    `yaml:"jwt_secret"`
}

type UserDef struct {
	Username     string `yaml:"username"`
	PasswordHash string `yaml:"password_hash"`
}

type ShellConfig struct {
	Command string   `yaml:"command"`
	Env     []string `yaml:"env"`
}

type SessionsConfig struct {
	MaxPerUser       int    `yaml:"max_per_user"`
	IdleTimeoutStr   string `yaml:"idle_timeout"`
	idleTimeoutCache time.Duration
}

// GetIdleTimeout returns the parsed idle timeout duration
func (s *SessionsConfig) GetIdleTimeout() time.Duration {
	if s.idleTimeoutCache != 0 {
		return s.idleTimeoutCache
	}
	if s.IdleTimeoutStr == "" {
		return 30 * time.Minute
	}
	d, err := time.ParseDuration(s.IdleTimeoutStr)
	if err != nil {
		return 30 * time.Minute
	}
	s.idleTimeoutCache = d
	return d
}

type SSHConfig struct {
	Enabled    bool   `yaml:"enabled"`
	KnownHosts string `yaml:"known_hosts"`
}

type RecordingConfig struct {
	Enabled bool   `yaml:"enabled"`
	Dir     string `yaml:"dir"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{Addr: ":8080"},
		Shell: ShellConfig{
			Env: []string{"TERM=xterm-256color"},
		},
		Sessions: SessionsConfig{
			MaxPerUser:     10,
			IdleTimeoutStr: "30m",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
