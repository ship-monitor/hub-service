package subscriber

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/ship-monitor/cloud/pkg/connector/channels"
	"github.com/ship-monitor/cloud/pkg/connector/schemas/messages"
	"github.com/ship-monitor/hub-service/internal/publishers/connect"
)

type Info interface {
	ID() uuid.UUID
}

type S struct {
	conn   *nats.Conn
	info   Info
	logger *slog.Logger
	pub    *connect.P
}

func New(
	conn *nats.Conn,
	info Info,
	logger *slog.Logger,
	pub *connect.P,
) *S {
	return &S{
		conn:   conn,
		info:   info,
		logger: logger,
		pub:    pub,
	}
}

func (s *S) Start() error {
	_, err := s.conn.Subscribe(channels.ConnectRequestChannel(s.info.ID()), s.HandleConnectRequest)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	return nil
}

func (s *S) HandleConnectRequest(msg *nats.Msg) {
	request, err := messages.UnmarshalConnectRequest(msg)
	if err != nil {
		s.logger.Error("Unmarshal request", slog.String("error", err.Error()))

		return
	}

	s.logger.Info("Connect request received", "requestId", request.RequestID)

	err = s.pub.PublishConnectResponse(&messages.ConnectResponse{
		Payload: messages.ConnectResponsePayload{
			Approve:   true,
			RequestID: request.RequestID,
		},
		Header: messages.ConnectResponseHeader{
			IdentificationHeader: messages.IdentificationHeader{
				DeviceID: s.info.ID(),
			},
		},
	})
	if err != nil {
		s.logger.Error("Publish connect response", "error", err.Error())
	}
}
