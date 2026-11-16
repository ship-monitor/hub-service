package di

import (
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/hub-service/internal/config"
	"github.com/ship-monitor/hub-service/internal/publishers"
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
	pub *publishers.P,
) *device.Worker {
	worker := device.New(logger.Named("Workers"), pub, pub)

	lc.Append(fx.StartStopHook(worker.Start, worker.Stop))

	return worker
}

func newSensorsWorker(
	lc fx.Lifecycle,
	conn *asyncapi.UserController,
	logger *zap.Logger,
	info *info.Service,
	config *config.Config,
	pub *publishers.P,
	repo *sensor.Repo,
) (*sensors.Worker, error) {
	w := sensors.New(conn, logger.Named("Workers"), info, pub, repo)

	lc.Append(fx.StartStopHook(w.Start, w.Stop))

	return w, nil
}
