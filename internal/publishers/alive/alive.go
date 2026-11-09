package alive

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"go.uber.org/zap"
)

// PubTimeout represents timeout between sending heartbeats.
const PubTimeout = time.Second * 20

type Info interface {
	ID() uuid.UUID
}

type Publisher struct {
	controller *asyncapi.UserController
	logger     *zap.Logger
	info       Info
	stop       chan struct{}
}

func New(
	controller *asyncapi.UserController,
	logger *zap.Logger,
	info Info,
) *Publisher {
	return &Publisher{
		controller: controller,
		logger:     logger,
		info:       info,
		stop:       make(chan struct{}),
	}
}

func (p *Publisher) Start() {
	go p.startPublishing()
}

func (p *Publisher) Stop() {
	p.stop <- struct{}{}
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
	hb := NewHeartbeat(p.info.ID())

	err := p.controller.SendToOnDeviceHeartbeatOperation(
		context.TODO(),
		asyncapi.DevicesAliveChannelParameters{DeviceId: p.info.ID().String()},
		*hb,
	)
	if err != nil {
		p.logger.Error("Publish heartbeat", zap.Error(err))

		return
	}

	p.logger.Info("Published heartbeat")
}

func NewHeartbeat(deviceID uuid.UUID) *asyncapi.HeartbeatMessage {
	timestamp := time.Now().Unix()

	return &asyncapi.HeartbeatMessage{
		Payload: asyncapi.HeartbeatSchema{
			Timestamp: timestamp,
		},
		Headers: asyncapi.IdentificationHeadersSchema{
			Authorization: deviceID.String(),
		},
	}
}
