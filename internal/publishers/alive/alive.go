package alive

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type DeviceInfoService interface {
	ID() uuid.UUID
}

type Publisher struct {
	nats   *nats.Conn
	logger *slog.Logger
	info   DeviceInfoService
	stop   chan struct{}
}

func New(
	nats *nats.Conn,
	logger *slog.Logger,
	info DeviceInfoService,
) *Publisher {
	return &Publisher{
		nats:   nats,
		logger: logger,
		info:   info,
		stop:   make(chan struct{}),
	}
}

func (p *Publisher) Start() {
	go p.startPublishing()
}

func (p *Publisher) Stop() {
	p.stop <- struct{}{}
}

func (p *Publisher) startPublishing() {
	ticker := time.NewTicker(2 * time.Second)
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
	hb := NewHeartbeat()

	data, err := json.Marshal(&hb)
	if err != nil {
		p.logger.Error("Marshall message", "error", err)
		return
	}

	subject := fmt.Sprintf("device.%s.health", p.info.ID().String())

	msg := nats.NewMsg(subject)
	msg.Header.Set("x-device-id", p.info.ID().String())
	msg.Data = data

	p.logger.Info("Sending heartbeat", "sbject", subject, "heartbeat", string(data))
	err = p.nats.PublishMsg(msg)
	if err != nil {
		p.logger.Error("Publish message", "error", err.Error())
	}
}

type Heartbeat struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

func NewHeartbeat() Heartbeat {
	return Heartbeat{
		Type:      "heartbeat",
		Timestamp: time.Now().UTC(),
	}
}
