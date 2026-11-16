package sensor

import (
	"context"

	"github.com/google/uuid"
	"github.com/ship-monitor/hub-service/internal/types"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals
var sensors = []types.Sensor{
	{ID: uuid.New(), Name: "coldstat-sensor-1"},
	{ID: uuid.New(), Name: "coldstat-sensor-2"},
}

type Repo struct {
	logger *zap.Logger
}

func New(
	logger *zap.Logger,
) *Repo {
	return &Repo{
		logger: logger.Named("Reposiories.Sensor"),
	}
}

func (r *Repo) GetSensors(ctx context.Context) ([]types.Sensor, error) {
	r.logger.Warn("Returns builtin constants sensors")

	return sensors, nil
}
