package device

import (
	"context"
	"time"

	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/hub-service/internal/config"
	"go.uber.org/zap"
)

const ExecutionTimeout = time.Second * 2

type HeartbeatPub interface {
	PublishHeartbeat(ctx context.Context, h *asyncapi.HeartbeatSchema) error
}
type MetricsPub interface {
	PublishDeviceMetrics(ctx context.Context, m []asyncapi.MetricSchema) error
}
type FirmwarePub interface {
	PublishFirmware(ctx context.Context, firmware asyncapi.FirmwareSchema) error
}
type Info interface {
	Firmware() (*asyncapi.FirmwareSchema, error)
}

type Worker struct {
	logger    *zap.Logger
	heartbeat HeartbeatPub
	metric    MetricsPub
	firmware  FirmwarePub
	info      Info
	config    *config.Config
	stop      chan struct{}
}

func New(
	logger *zap.Logger,
	heartbeat HeartbeatPub,
	metric MetricsPub,
	firmware FirmwarePub,
	info Info,
	config *config.Config,
) *Worker {
	return &Worker{
		logger:    logger.Named("Device"),
		heartbeat: heartbeat,
		metric:    metric,
		firmware:  firmware,
		info:      info,
		config:    config,
		stop:      make(chan struct{}),
	}
}

func (w *Worker) Start() {
	go w.startPublishingHeartbeat()
	go w.startPublishingMetrics()
}

func (w *Worker) Stop() {
	w.stop <- struct{}{}

	w.stop <- struct{}{}
}
