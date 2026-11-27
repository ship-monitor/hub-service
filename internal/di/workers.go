package di

import (
	"github.com/ship-monitor/hub-service/internal/config"
	"github.com/ship-monitor/hub-service/internal/publisher"
	"github.com/ship-monitor/hub-service/internal/repositories/sensor"
	"github.com/ship-monitor/hub-service/internal/services/info"
	"github.com/ship-monitor/hub-service/internal/workers/device"
	"github.com/ship-monitor/hub-service/internal/workers/sensors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func newDeviceWorker(
	lc fx.Lifecycle,
	logger *zap.Logger,
	pub *publisher.P,
	info *info.Service,
) *device.Worker {
	worker := device.New(logger.Named("Workers"), pub, pub, pub, info)

	lc.Append(fx.StartStopHook(worker.Start, worker.Stop))

	return worker
}

func newSensorsWorker(
	lc fx.Lifecycle,
	logger *zap.Logger,
	config *config.Config,
	pub *publisher.P,
	repo *sensor.Repo,
) (*sensors.Worker, error) {
	w := sensors.New(logger.Named("Workers"), pub, repo)

	lc.Append(fx.StartStopHook(w.Start, w.Stop))

	return w, nil
}
