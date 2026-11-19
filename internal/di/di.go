package di

import (
	"fmt"

	"github.com/lerenn/asyncapi-codegen/pkg/extensions/brokers/nats"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/cloud/pkg/connector/common"
	"github.com/ship-monitor/hub-service/internal/config"
	"github.com/ship-monitor/hub-service/internal/publishers"
	"github.com/ship-monitor/hub-service/internal/repositories/sensor"
	"github.com/ship-monitor/hub-service/internal/services/auth"
	"github.com/ship-monitor/hub-service/internal/services/info"
	"github.com/ship-monitor/hub-service/internal/subscriber"
	"github.com/ship-monitor/hub-service/internal/workers/device"
	"github.com/ship-monitor/hub-service/internal/workers/sensors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewApp(config *config.Config, logger *zap.Logger) *fx.App {
	return fx.New(
		fx.Supply(config),
		fx.Supply(logger),

		fx.Provide(newInfoService),
		fx.Provide(newAuthService),

		fx.Provide(newPub),
		fx.Provide(newSensorsWorker),

		fx.Provide(newSub),

		fx.Provide(newConnectorController),

		fx.Provide(newDeviceWorker),

		fx.Provide(sensor.New),

		fx.Invoke(func(*subscriber.S) {}),
		fx.Invoke(func(*device.Worker) {}),
		fx.Invoke(func(*sensors.Worker) {}),
	)
}

func newInfoService(
	lc fx.Lifecycle,
	config *config.Config,
	logger *zap.Logger,
) (*info.Service, error) {
	s, err := info.New(config, logger)
	if err != nil {
		return nil, fmt.Errorf("new info service: %w", err)
	}

	lc.Append(fx.StartHook(func() {
		logger.Info("Device info", zap.Any("id", s.ID()))
	}))

	return s, nil
}
func newPub(
	logger *zap.Logger,
	conn *asyncapi.UserController,
	info *info.Service,
	auth *auth.Service,
) (*publishers.P, error) {
	return publishers.New(conn, logger, info, auth), nil
}
func newSub(
	lc fx.Lifecycle,
	controller *asyncapi.UserController,
	info *info.Service,
	logger *zap.Logger,
	pub *publishers.P,
	firmware *device.Worker,
) (*subscriber.S, error) {
	s := subscriber.New(controller, info, logger, pub, firmware)

	lc.Append(fx.StartStopHook(s.Start, s.Stop))

	return s, nil
}

func newConnectorController(
	conf *config.Config,
	logger *zap.Logger,
) (*asyncapi.UserController, error) {
	natsController, err := nats.NewController(conf.NATS.URL)
	if err != nil {
		return nil, fmt.Errorf("new nats controller: %w", err)
	}

	c, err := asyncapi.NewUserController(natsController, asyncapi.WithLogger(common.LoggerFromZap(logger)))
	if err != nil {
		return nil, fmt.Errorf("new app controller: %w", err)
	}

	return c, nil
}
func newAuthService(info *info.Service) *auth.Service {
	return auth.New(info)
}
