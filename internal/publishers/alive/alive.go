package alive

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/ship-monitor/cloud/pkg/connector/schemas/messages"
)

// PubTimeout represents timeout between sending heartbeats.
const PubTimeout = time.Second * 20

type DeviceInfoService interface {
	ID() uuid.UUID
}

type Publisher struct {
	nats   *nats.Conn
	logger *slog.Logger
	info   DeviceInfoService
	stop   chan struct{}

	subject string
}

func New(
	nats *nats.Conn,
	logger *slog.Logger,
	info DeviceInfoService,
) *Publisher {
	return &Publisher{
		nats:    nats,
		logger:  logger,
		info:    info,
		stop:    make(chan struct{}),
		subject: fmt.Sprintf("device.%s.health", info.ID().String()),
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

	msg, err := heartbeatToMsg(p.subject, hb)
	if err != nil {
		p.logger.Error("Convert heartbeat to message", "error", err.Error())
	}

	p.logger.Info("Sending heartbeat", "sbject", p.subject)

	err = p.nats.PublishMsg(msg)
	if err != nil {
		p.logger.Error("Publish message", "error", err.Error())
	}
}

func NewHeartbeat(deviceID uuid.UUID) *messages.Heartbeat {
	return &messages.Heartbeat{
		Payload: messages.HeartbeatPayload{
			Timestamp: time.Now().UTC().Unix(),
		},
		Headers: messages.HeartbeatHeaders{
			IdentificationHeader: messages.IdentificationHeader{
				DeviceID: deviceID,
			},
		},
	}
}

const DeviceIDHeader = "device-id"

func heartbeatToMsg(subj string, h *messages.Heartbeat) (*nats.Msg, error) {
	msg := nats.NewMsg(subj)

	body, err := json.Marshal(&h.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshall json: %w", err)
	}

	msg.Data = body

	msg.Header.Set(DeviceIDHeader, h.Headers.DeviceID.String())

	return msg, nil
}
