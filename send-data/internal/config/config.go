package config

import (
	"fmt"
	"path/filepath"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Storage     StorageConfig     `yaml:"storage"`
	Auth        AuthConfig        `yaml:"auth"`
	RateLimit   RateLimitConfig   `yaml:"rate_limit"`
	Limits      LimitsConfig      `yaml:"limits"`
	Proxy       ProxyConfig       `yaml:"proxy"`
	Origin      OriginConfig      `yaml:"origin"`
	PostProcess PostProcessConfig `yaml:"post_process"`
}

type ServerConfig struct {
	Listen string    `yaml:"listen"`
	Port   int       `yaml:"port"`
	TLS    TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type StorageConfig struct {
	SpoolDir string `yaml:"spool_dir"`
}

type AuthConfig struct {
	TokenSecret string `yaml:"token_secret"`
}

type RateLimitConfig struct {
	GetIDMaxPerIP     int `yaml:"getid_max_per_ip"`
	GetIDIntervalSec  int `yaml:"getid_interval_sec"`
	ReportMaxPerIP    int `yaml:"report_max_per_ip"`
	ReportIntervalSec int `yaml:"report_interval_sec"`
}

type LimitsConfig struct {
	MaxAuthGetSize  int `yaml:"max_auth_get_size"`
	MinPostBodySize int `yaml:"min_post_body_size"`
	MaxPostBodySize int `yaml:"max_post_body_size"`
}

type ProxyConfig struct {
	TrustedAddrs []string `yaml:"trusted_addrs"`
}

type OriginConfig struct {
	Site string `yaml:"site"`
}

type PostProcessConfig struct {
	Enabled bool   `yaml:"enabled"`
	Script  string `yaml:"script"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Listen: "0.0.0.0",
			Port:   8080,
		},
		Storage: StorageConfig{
			SpoolDir: "/usr/local/www/lbperf/spool",
		},
		RateLimit: RateLimitConfig{
			GetIDMaxPerIP:     20,
			GetIDIntervalSec:  60,
			ReportMaxPerIP:      10,
			ReportIntervalSec:   60,
		},
		Limits: LimitsConfig{
			MaxAuthGetSize:  128,
			MinPostBodySize: 64,
			MaxPostBodySize: 10 * 1024 * 1024,
		},
		Proxy: ProxyConfig{
			TrustedAddrs: []string{"127.0.0.1", "::1"},
		},
	}
}

func (c *Config) validate() error {
	if c.Server.Listen == "" {
		c.Server.Listen = "0.0.0.0"
	}
	if c.Server.Port <= 0 {
		c.Server.Port = 8080
	}
	if c.Storage.SpoolDir == "" {
		return fmt.Errorf("storage.spool_dir is required")
	}
	if c.Auth.TokenSecret == "" {
		return fmt.Errorf("auth.token_secret is required")
	}
	if c.Limits.MaxAuthGetSize <= 0 {
		c.Limits.MaxAuthGetSize = 128
	}
	if c.Limits.MinPostBodySize < 0 {
		c.Limits.MinPostBodySize = 64
	}
	if c.Limits.MaxPostBodySize <= 0 {
		c.Limits.MaxPostBodySize = 10 * 1024 * 1024
	}
	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" || c.Server.TLS.KeyFile == "" {
			return fmt.Errorf("tls.enabled requires cert_file and key_file")
		}
	}
	return nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Listen, c.Server.Port)
}

func (c *Config) DataDir(token string) string {
	return filepath.Join(c.Storage.SpoolDir, "data", token)
}

func (c *Config) LockDir(subsystem string) string {
	return filepath.Join(c.Storage.SpoolDir, subsystem)
}
