package verify

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/ship-monitor/cloud/pkg/connector/channels"
	"github.com/ship-monitor/cloud/pkg/connector/schemas/messages"
)

type DeviceInfoService interface {
	ID() uuid.UUID
	SecurityKey() string
}

type Subscriber struct {
	nats   *nats.Conn
	logger *slog.Logger
	info   DeviceInfoService
	sub    *nats.Subscription
}

func New(
	nats *nats.Conn,
	logger *slog.Logger,
	info DeviceInfoService,
) *Subscriber {
	return &Subscriber{
		nats:   nats,
		logger: logger,
		info:   info,
	}
}

func (s *Subscriber) Start() error {
	sub, err := s.nats.Subscribe(
		channels.ConnectRequestChannel(s.info.ID()),
		s.handleRequest,
	)
	if err != nil {
		return fmt.Errorf("subscribe connect requests: %w", err)
	}

	s.sub = sub

	return nil
}

func (s *Subscriber) Stop() error {
	if s.sub == nil {
		return nil
	}

	err := s.sub.Unsubscribe()
	if err != nil {
		return fmt.Errorf("unsubscribe: %w", err)
	}

	return nil
}

type responsePayload struct {
	Allow     bool      `json:"allow"`
	RequestID uuid.UUID `json:"requestId"`
}

func (s *Subscriber) handleRequest(msg *nats.Msg) {
	request, err := messages.UnmarshalConnectRequest(msg)
	if err != nil {
		s.logger.Error("Unmarshall connect request", "error", err.Error())
		return
	}

	allow := subtle.ConstantTimeCompare(
		[]byte(request.SecurityKey),
		[]byte(s.info.SecurityKey()),
	) == 1

	s.logger.Info(
		"Connect request",
		"requestId", request.RequestID,
		"deviceId", request.DeviceID,
		"allow", allow,
	)

	response := nats.NewMsg(
		channels.ConnectResponseChannel(request.DeviceID),
	)
	response.Header.Set("device-id", s.info.ID().String())

	response.Data, err = json.Marshal(responsePayload{
		Allow:     allow,
		RequestID: request.RequestID,
	})
	if err != nil {
		s.logger.Error("Marshall response", "error", err.Error())
		return
	}

	err = s.nats.PublishMsg(response)
	if err != nil {
		s.logger.Error("Publish response", "error", err.Error())
	}
}
