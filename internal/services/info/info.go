package info

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/ship-monitor/hub-service/internal/config"
	"go.uber.org/zap"
)

type Service struct {
	id          uuid.UUID
	securityKey string
}

func New(
	config *config.Config,
	logger *zap.Logger,
) (*Service, error) {
	prefix := "/etc/coldstat/"

	if config.IsDevelopment() {
		logger.Warn(
			"Development mode enabled, using development prefix to find info files",
			zap.String("prefix", "."),
		)

		prefix = "."
	}

	data, err := os.ReadFile(path.Join(prefix, "device-id"))
	if err != nil {
		return nil, fmt.Errorf("read device-id file: %w", err)
	}

	id, err := uuid.Parse(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, fmt.Errorf("invalid device id: parse uuid: %w", err)
	}

	securityKey, err := os.ReadFile(path.Join(prefix, "security-key"))
	if err != nil {
		return nil, fmt.Errorf("read security-key file: %w", err)
	}

	return &Service{
		id:          id,
		securityKey: strings.TrimSpace(string(securityKey)),
	}, nil
}

func (s *Service) ID() uuid.UUID {
	return s.id
}

func (s *Service) CheckSecurityKey(key string) bool {
	return key == s.securityKey
}
