package config

import (
	"strings"

	"github.com/ship-monitor/cloud/pkg/configer"
)

type NATS struct {
	URL string
}

type Config struct {
	NATS NATS
	Mode string
}

func (c *Config) IsDevelopment() bool {
	return strings.Contains(strings.ToUpper(c.Mode), "DEV")
}

func New() Config {
	return Config{
		Mode: configer.GetOrDefault("MODE", "DEVELOPMENT"),
		NATS: NATS{
			URL: configer.GetOrDefault("NATS_URL", "nats://localhost:4222"),
		},
	}
}
