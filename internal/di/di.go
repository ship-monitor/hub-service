package di

import (
	"fmt"
	"log/slog"

	"github.com/lerenn/asyncapi-codegen/pkg/extensions/brokers/nats"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/cloud/pkg/connector/common"
	"github.com/ship-monitor/hub-service/internal/config"
	"github.com/ship-monitor/hub-service/internal/publishers/alive"
	"github.com/ship-monitor/hub-service/internal/publishers/connect"
	"github.com/ship-monitor/hub-service/internal/publishers/sensors"
	"github.com/ship-monitor/hub-service/internal/services/info"
	"github.com/ship-monitor/hub-service/internal/subscriber"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewApp(config *config.Config) *fx.App {
	return fx.New(
		fx.Supply(config),
		fx.Provide(newLogger),
		fx.Provide(newLoggerZap),

		fx.Provide(info.New),
		fx.Provide(newAlivePub),
		fx.Provide(newConnectPub),
		fx.Provide(newSensorsPub),
		fx.Provide(newSub),

		fx.Provide(newConnectorController),

		fx.Invoke(func(*subscriber.S) {}),
		fx.Invoke(func(*alive.Publisher) {}),
		fx.Invoke(func(*sensors.Publisher) {}),
	)
}
func newAlivePub(
	lc fx.Lifecycle,
	conn *asyncapi.UserController,
	logger *zap.Logger,
	info *info.Service,
) (*alive.Publisher, error) {
	pub := alive.New(conn, logger, info)

	lc.Append(fx.StartStopHook(pub.Start, pub.Stop))

	return pub, nil
}
func newConnectPub(
	logger *slog.Logger,
	conn *asyncapi.UserController,
	info *info.Service,
) (*connect.P, error) {
	return connect.New(logger, conn, info), nil
}
func newSensorsPub(
	lc fx.Lifecycle,
	conn *asyncapi.UserController,
	logger *slog.Logger,
	info *info.Service,
	config *config.Config,
) (*sensors.Publisher, error) {
	pub := sensors.New(conn, logger, info, config.SensorID, config.SensorName)

	lc.Append(fx.StartStopHook(pub.Start, pub.Stop))

	return pub, nil
}
func newLogger() *slog.Logger {
	return slog.Default()
}
func newLoggerZap() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("new logger: %w", err)
	}

	return logger, nil
}

func newSub(
	lc fx.Lifecycle,
	controller *asyncapi.UserController,
	info *info.Service,
	logger *slog.Logger,
	pub *connect.P,
) (*subscriber.S, error) {
	s := subscriber.New(controller, info, logger, pub)

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
