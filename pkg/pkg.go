package pkg

import (
	"context"
	"errors"
)

func ExecuteWitrRetry(ctx context.Context, retryCount int, exec func(ctx context.Context) error) error {
	var finalErr error

	for range retryCount {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		err := exec(ctx)
		if err != nil {
			finalErr = errors.Join(finalErr, err)

			continue
		} else {
			return nil
		}
	}

	return finalErr
}
