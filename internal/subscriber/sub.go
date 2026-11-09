package subscriber

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/channels"
	"github.com/ship-monitor/cloud/pkg/connector/messages"
	"github.com/ship-monitor/cloud/pkg/connector/parameters"
	"github.com/ship-monitor/cloud/pkg/connector/schemas"
	"github.com/ship-monitor/cloud/pkg/connector/servers"
)

// resubscribeTimeout — пауза между попытками восстановить подписку.
const resubscribeTimeout = 3 * time.Second

type DeviceInfoService interface {
	ID() uuid.UUID
	SecurityKey() string
}

type S struct {
	requests  *channels.ConnectRequestsNats
	responses *channels.ConnectResponsesNats
	logger    *slog.Logger
	info      DeviceInfoService
	cancel    context.CancelFunc
}

func New(
	server servers.Development,
	logger *slog.Logger,
	info DeviceInfoService,
) (*S, error) {
	requests, err := server.OpenConnectRequestsNats(
		context.Background(),
		channels.ConnectRequestsParameters{
			DeviceID: parameters.DeviceID(info.ID().String()),
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("open connect requests channel: %w", err)
	}

	responses, err := server.OpenConnectResponsesNats(
		context.Background(),
		channels.ConnectResponsesParameters{
			DeviceID: parameters.DeviceID(info.ID().String()),
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("open connect responses channel: %w", err)
	}

	return &S{
		requests:  requests,
		responses: responses,
		logger:    logger,
		info:      info,
	}, nil
}

func (s *S) Start(_ context.Context) error {
	// SubscribeConnectRequest блокирует горутину до отмены контекста,
	// а при ошибке распаковки завершает подписку — поэтому слушаем
	// в фоне и восстанавливаем подписку после сбоев.
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		for {
			err := s.requests.SubscribeConnectRequest(ctx, s.HandleConnectRequest)
			if ctx.Err() != nil {
				return
			}

			s.logger.Error("Subscribe connect requests, resubscribing", "error", err)

			select {
			case <-ctx.Done():
				return
			case <-time.After(resubscribeTimeout):
			}
		}
	}()

	return nil
}

func (s *S) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}

	return s.requests.Close()
}

// Approve решает, подтверждать ли заявку на подключение. Пока ключ
// устройства не задан конфигурацией, хаб одобряет любую заявку.
func (s *S) Approve(secretKey *string) bool {
	key := s.info.SecurityKey()
	if key == "" {
		return true
	}

	if secretKey == nil {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(*secretKey), []byte(key)) == 1
}

func (s *S) HandleConnectRequest(message messages.ConnectRequestReceiver) {
	payload := message.Payload()

	approve := s.Approve(payload.SecretKey)

	requestID := "unknown"
	if payload.RequestID != nil {
		requestID = payload.RequestID.String()
	}

	s.logger.Info(
		"Connect request",
		"requestId", requestID,
		"approve", approve,
	)

	response := (&messages.ConnectResponseOut{}).
		SetPayload(schemas.ConnectResponse{
			Approve:   &approve,
			RequestID: payload.RequestID,
		}).
		SetHeaders(map[string]any{
			"DeviceID": s.info.ID().String(),
		})

	err := s.responses.PublishConnectResponse(context.Background(), response)
	if err != nil {
		s.logger.Error("Publish connect response", "error", err.Error())
	}
}
