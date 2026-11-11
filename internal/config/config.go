package config

import (
	"os"

	"github.com/google/uuid"
)

const defaultDeviceID = "3dea52fe-4fa8-43cf-9ea1-d54c7b74d467"

type NATS struct {
	URL string
}

type Config struct {
	NATS        NATS
	DeviceID    uuid.UUID
	SecurityKey string
	SensorID    string
	SensorName  string
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func New() Config {
	return Config{
		NATS: NATS{
			URL: getenv("NATS_URL", "nats://localhost:4222"),
		},
		DeviceID: uuid.MustParse(
			getenv("DEVICE_ID", defaultDeviceID),
		),
		SecurityKey: os.Getenv("SECURITY_KEY"),
		SensorID:    getenv("SENSOR_ID", "temp-1"),
		SensorName:  getenv("SENSOR_NAME", "Температура камеры"),
	}
}
