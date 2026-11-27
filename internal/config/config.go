package config

import (
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/ship-monitor/cloud/pkg/configer"
)

type NATS struct {
	URL string
}

type Config struct {
	NATS NATS
	Mode string

	SendDeclarationTimeout time.Duration
	// SendMetricsTimeout represents timeout between sending metrics.
	SendMetricsTimeout time.Duration
	// SendHeartbeatTimeout represents timeout between sending heartbeats.
	SendHeartbeatTimeout time.Duration

	CheckTemperatureTimeout time.Duration
}

func (c *Config) IsDevelopment() bool {
	return strings.Contains(strings.ToUpper(c.Mode), "DEV")
}

//nolint:mnd
func New() Config {
	_ = godotenv.Load()

	return Config{
		Mode: configer.GetOrDefault("MODE", "DEVELOPMENT"),
		NATS: NATS{
			URL: configer.GetOrDefault("NATS_URL", "nats://localhost:4222"),
		},
		SendDeclarationTimeout:  configer.GetDurationOrDefault("SEND_DECLARATION_TIMEOUT", time.Minute*2),
		SendMetricsTimeout:      configer.GetDurationOrDefault("SEND_METRICS_TIMEOUT", time.Second*30),
		CheckTemperatureTimeout: configer.GetDurationOrDefault("CHECK_TEMP_TIMEOUT", time.Second*20),
		SendHeartbeatTimeout:    configer.GetDurationOrDefault("SEND_HEARTBEAT_TIMEOUT", time.Second*20),
	}
}
