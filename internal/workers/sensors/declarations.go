package sensors

import (
	"context"
	"fmt"
	"time"

	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"go.uber.org/zap"
)

func (w *Worker) startPublishingDeclarations() {
	go w.publishDeclaration()

	ticker := time.NewTicker(w.config.SendDeclarationTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.publishDeclaration()
		case <-w.stop:
			w.logger.Info("Receive stop")

			return
		}
	}
}

func (w *Worker) publishDeclaration() {
	ctx, cancel := context.WithTimeout(context.Background(), w.config.SendDeclarationTimeout)
	defer cancel()

	err := w.declare(ctx)
	if err != nil {
		w.logger.Error("Publish declaration", zap.Error(err))
	} else {
		w.logger.Info("Published declaration")
	}
}

func (w *Worker) declare(ctx context.Context) error {
	s, err := w.repo.GetSensors(ctx)
	if err != nil {
		return fmt.Errorf("get sensors: %w", err)
	}

	err = w.pub.PublishSensorDeclaration(
		ctx,
		asyncapi.SensorDeclarationSchema{Sensors: mapSensors(s)},
	)
	if err != nil {
		return fmt.Errorf("send declaration: %w", err)
	}

	return nil
}
