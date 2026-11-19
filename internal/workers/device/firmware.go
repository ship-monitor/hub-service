package device

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func (w *Worker) startPublishingFirmware() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	err := w.PublishFirmware(ctx)
	if err != nil {
		w.logger.Error("Publish firmare", zap.Error(err))
	} else {
		w.logger.Info("Published firmware")
	}
}

func (w *Worker) PublishFirmware(ctx context.Context) error {
	f, err := w.info.Firmware()
	if err != nil {
		return fmt.Errorf("get firmware: %w", err)
	}

	err = w.firmware.PublishFirmware(ctx, *f)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}
