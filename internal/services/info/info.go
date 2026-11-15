package info

import (
	"github.com/google/uuid"
	"github.com/ship-monitor/hub-service/internal/config"
)

// Ключ безопасности относится к устройству, а не к сервису.
const securityKey = "ship-hub-test-key"

type Service struct {
	id uuid.UUID
}

func New(config *config.Config) *Service {
	return &Service{
		id: config.DeviceID,
	}
}

func (s *Service) ID() uuid.UUID {
	return s.id
}

func (s *Service) CheckSecurityKey(key string) bool {
	return key == securityKey
}
