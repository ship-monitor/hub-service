package info

import (
	"github.com/google/uuid"
	"github.com/ship-monitor/hub-service/internal/config"
)

type Service struct {
	id          uuid.UUID
	securityKey string
}

func New(config *config.Config) *Service {
	return &Service{
		id:          config.DeviceID,
		securityKey: config.SecurityKey,
	}
}

func (s *Service) ID() uuid.UUID {
	return s.id
}

func (s *Service) SecurityKey() string {
	return s.securityKey
}
