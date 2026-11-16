package device

import (
	"context"
	"time"

	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"go.uber.org/zap"
)

const ExecutionTimeout = time.Second * 2

type HeartbeatPub interface {
	PublishHeartbeat(ctx context.Context, h *asyncapi.HeartbeatSchema) error
}
type MetricsPub interface {
	PublishDeviceMetrics(ctx context.Context, m []asyncapi.MetricSchema) error
}

type Worker struct {
	logger    *zap.Logger
	heartbeat HeartbeatPub
	metric    MetricsPub
	stop      chan struct{}
}

func New(
	logger *zap.Logger,
	heartbeat HeartbeatPub,
	metric MetricsPub,
) *Worker {
	return &Worker{
		logger:    logger.Named("Device"),
		heartbeat: heartbeat,
		metric:    metric,
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
