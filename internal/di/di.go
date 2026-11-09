package di

import (
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/ship-monitor/hub-service/internal/config"
	"github.com/ship-monitor/hub-service/internal/publishers/alive"
	"github.com/ship-monitor/hub-service/internal/publishers/connect"
	"github.com/ship-monitor/hub-service/internal/services/info"
	"github.com/ship-monitor/hub-service/internal/subscriber"
	"go.uber.org/fx"
)

func NewApp(config *config.Config) *fx.App {
	return fx.New(
		fx.Supply(config),
		fx.Provide(newLogger),

		fx.Provide(info.New),
		fx.Provide(newNats),
		fx.Provide(newAlivePub),
		fx.Provide(newConnectPub),
		fx.Provide(newSub),

		fx.Invoke(func(*alive.Publisher) {}),
		fx.Invoke(func(*subscriber.S) {}),
	)
}
func newNats(lc fx.Lifecycle, config *config.Config) (*nats.Conn, error) {
	conn, err := nats.Connect(config.NATS.URL)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	lc.Append(fx.StopHook(conn.Close))

	return conn, nil
}
func newAlivePub(
	lc fx.Lifecycle,
	conn *nats.Conn,
	logger *slog.Logger,
	info *info.Service,
) *alive.Publisher {
	pub := alive.New(conn, logger, info)

	lc.Append(fx.StartStopHook(pub.Start, pub.Stop))

	return pub
}
func newConnectPub(
	logger *slog.Logger,
	conn *nats.Conn,
	info *info.Service,
) *connect.P {
	return connect.New(logger, conn)
}
func newLogger() *slog.Logger {
	return slog.Default()
}

func newSub(
	lc fx.Lifecycle,
	conn *nats.Conn,
	info *info.Service,
	logger *slog.Logger,
	pub *connect.P,
) *subscriber.S {
	s := subscriber.New(conn, info, logger, pub)

	lc.Append(fx.StartHook(s.Start))

	return s
}
