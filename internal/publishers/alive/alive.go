package alive

import (
	"context"
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

// PubTimeout represents timeout between sending heartbeats.
const PubTimeout = time.Second * 20

type DeviceInfoService interface {
	ID() uuid.UUID
}

type Publisher struct {
	channel *channels.DevicesAliveNats
	logger  *slog.Logger
	info    DeviceInfoService
	stop    chan struct{}
}

func New(
	server servers.Development,
	logger *slog.Logger,
	info DeviceInfoService,
) (*Publisher, error) {
	channel, err := server.OpenDevicesAliveNats(
		context.Background(),
		channels.DevicesAliveParameters{
			DeviceID: parameters.DeviceID(info.ID().String()),
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("open devices alive channel: %w", err)
	}

	return &Publisher{
		channel: channel,
		logger:  logger,
		info:    info,
		stop:    make(chan struct{}),
	}, nil
}

func (p *Publisher) Start() error {
	go p.startPublishing()

	return nil
}

func (p *Publisher) Stop() error {
	close(p.stop)

	return p.channel.Close()
}

func (p *Publisher) startPublishing() {
	ticker := time.NewTicker(PubTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.publishHeartbeat()
		case <-p.stop:
			p.logger.Info("Receive stop")

			return
		}
	}
}

func (p *Publisher) publishHeartbeat() {
	timestamp := time.Now().UTC().Unix()
	deviceID := p.info.ID()

	message := (&messages.HeartbeatOut{}).
		SetPayload(schemas.Heartbeat{Timestamp: &timestamp}).
		SetHeaders(schemas.IdentificationHeaders{DeviceID: &deviceID})

	p.logger.Info("Sending heartbeat", "subject", p.channel.Subject())

	err := p.channel.PublishHeartbeat(context.Background(), message)
	if err != nil {
		p.logger.Error("Publish heartbeat", "error", err.Error())
	}
}
