package driver

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
)

type Metric struct {
	Value     float64
	Timestamp time.Time
}

type Driver struct{}

func New(deviceID uuid.UUID) (*Driver, error) {
	return &Driver{}, nil
}

var ErrUnknownMetric = errors.New("unknown metric")

func (d *Driver) Read(ctx context.Context, metric string) (*Metric, error) {
	switch metric {
	case "temperature":
		return &Metric{Value: temperature(), Timestamp: time.Now()}, nil
	case "battery-level":
		return &Metric{Value: 100, Timestamp: time.Now()}, nil //nolint:mnd
	case "power-connected":
		return &Metric{Value: 1, Timestamp: time.Now()}, nil
	default:
		return nil, ErrUnknownMetric
	}
}

func (d *Driver) ReadTemperature(ctx context.Context) (*Metric, error) {
	return d.Read(ctx, "temperature")
}

// temperature generate mock temperature value.
//
//nolint:mnd, gosec
func temperature() float64 {
	base := -16.5 + math.Sin(float64(time.Now().UnixNano())/1e9/60)*0.6

	return math.Round((base+rand.NormFloat64()*0.2)*10) / 10
}
