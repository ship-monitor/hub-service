package info

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/hub-service/internal/config"
	"go.uber.org/zap"
)

func getPrefix(config *config.Config) string {
	prefix := "/etc/coldstat/"

	if config.IsDevelopment() {
		prefix = "."
	}

	return prefix
}

type Service struct {
	id          uuid.UUID
	securityKey string
	config      *config.Config
}

func New(
	config *config.Config,
	logger *zap.Logger,
) (*Service, error) {
	prefix := getPrefix(config)

	data, err := os.ReadFile(
		path.Clean(
			path.Join(prefix, "device-id")))
	if err != nil {
		return nil, fmt.Errorf("read device-id file: %w", err)
	}

	id, err := uuid.Parse(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, fmt.Errorf("invalid device id: parse uuid: %w", err)
	}

	securityKey, err := os.ReadFile(
		path.Clean(
			path.Join(prefix, "security-key")))
	if err != nil {
		return nil, fmt.Errorf("read security-key file: %w", err)
	}

	return &Service{
		id:          id,
		securityKey: strings.TrimSpace(string(securityKey)),
		config:      config,
	}, nil
}

func (s *Service) ID() uuid.UUID {
	return s.id
}

func (s *Service) CheckSecurityKey(key string) bool {
	return key == s.securityKey
}

type Firmware struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Revision     string `json:"revision"`
}

// SOFTWARE represents name and version of runnig service.
//
//nolint:gochecknoglobals
var SOFTWARE = "hub-service@dev"

func (s *Service) Firmware() (*asyncapi.FirmwareSchema, error) {
	data, err := os.ReadFile(path.Join(getPrefix(s.config), "firmware.json"))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var firmware Firmware

	err = json.Unmarshal(data, &firmware)
	if err != nil {
		return nil, fmt.Errorf("unmarshal data: %w", err)
	}

	return &asyncapi.FirmwareSchema{
		Manufacturer: firmware.Manufacturer,
		Model:        firmware.Model,
		Revision:     firmware.Revision,
		Software:     SOFTWARE,
	}, nil
}
