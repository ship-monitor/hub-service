package connect

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/ship-monitor/cloud/pkg/connector/channels"
	"github.com/ship-monitor/cloud/pkg/connector/schemas/messages"
)

type P struct {
	logger *slog.Logger
	conn   *nats.Conn
}

func New(
	logger *slog.Logger,
	conn *nats.Conn,
) *P {
	return &P{
		logger: logger,
		conn:   conn,
	}
}

const DeviceIDHeader = "device-id"

func (p *P) PublishConnectResponse(response *messages.ConnectResponse) error {
	msg := nats.NewMsg(channels.ConnectResponseChannel(response.Header.DeviceID))

	data, err := json.Marshal(response.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	msg.Data = data

	msg.Header.Add(DeviceIDHeader, response.Header.DeviceID.String())

	err = p.conn.PublishMsg(msg)
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}
