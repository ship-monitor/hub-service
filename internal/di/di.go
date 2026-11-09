package di

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/ship-monitor/cloud/pkg/connector/servers"
	"github.com/ship-monitor/hub-service/internal/config"
	"github.com/ship-monitor/hub-service/internal/publishers/alive"
	"github.com/ship-monitor/hub-service/internal/services/info"
	"github.com/ship-monitor/hub-service/internal/subscriber"
	"go.uber.org/fx"
)

func NewApp(config *config.Config) *fx.App {
	return fx.New(
		fx.Supply(config),
		fx.Provide(newLogger),

		fx.Provide(info.New),
		fx.Provide(newServer),
		fx.Provide(newAlivePub),
		fx.Provide(newSub),

		fx.Invoke(func(*alive.Publisher) {}),
		fx.Invoke(func(*subscriber.S) {}),
	)
}

func newLogger() *slog.Logger {
	return slog.Default()
}

func newServer(lc fx.Lifecycle, config *config.Config) (servers.Development, error) {
	serverURL, err := url.Parse(config.NATS.URL)
	if err != nil {
		return servers.Development{}, err
	}

	conn, err := servers.ConnectDevelopmentBidi(nil, serverURL)
	if err != nil {
		return servers.Development{}, err
	}

	lc.Append(fx.StopHook(conn.Close))

	return conn.Development, nil
}

func newAlivePub(
	lc fx.Lifecycle,
	server servers.Development,
	logger *slog.Logger,
	info *info.Service,
) (*alive.Publisher, error) {
	pub, err := alive.New(server, logger, info)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.StartStopHook(pub.Start, pub.Stop))

	return pub, nil
}

func newSub(
	lc fx.Lifecycle,
	server servers.Development,
	logger *slog.Logger,
	info *info.Service,
) (*subscriber.S, error) {
	sub, err := subscriber.New(server, logger, info)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return sub.Start(ctx)
		},
		OnStop: func(_ context.Context) error {
			return sub.Stop()
		},
	})

	return sub, nil
}
