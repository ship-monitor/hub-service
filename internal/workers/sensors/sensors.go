package sensors

import (
	"context"

	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/hub-service/internal/config"
	"github.com/ship-monitor/hub-service/internal/types"
	"go.uber.org/zap"
)

type Repo interface {
	GetSensors(ctx context.Context) ([]types.Sensor, error)
}

type Pub interface {
	PublishSensorMetrics(ctx context.Context, metrics asyncapi.SensorMetricsMessagePayload) error
	PublishSensorDeclaration(ctx context.Context, declaration asyncapi.SensorDeclarationSchema) error
	PublishNotification(ctx context.Context, n asyncapi.NotificationSchema) error
}

type Worker struct {
	logger *zap.Logger
	pub    Pub
	repo   Repo
	config *config.Config
	stop   chan struct{}
}

func New(
	logger *zap.Logger,
	pub Pub,
	repo Repo,
	config *config.Config,
) *Worker {
	return &Worker{
		logger: logger.Named("Sensors"),
		pub:    pub,
		repo:   repo,
		config: config,
		stop:   make(chan struct{}),
	}
}

func (w *Worker) Start() {
	go w.startPublishingMetrics()
	go w.startPublishingDeclarations()
	go w.startCheckingTemp()
}

func (w *Worker) Stop() {
	w.logger.Info("Stopping")

	w.stop <- struct{}{}

	w.stop <- struct{}{}

	w.stop <- struct{}{}
}
