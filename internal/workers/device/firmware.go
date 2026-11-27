package device

import (
	"context"
	"fmt"
)

func (w *Worker) PublishFirmware(ctx context.Context) error {
	f, err := w.info.Firmware()
	if err != nil {
		return fmt.Errorf("get firmware: %w", err)
	}

	err = w.firmware.PublishFirmware(ctx, *f)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	w.logger.Info("Firmware published")

	return nil
}
