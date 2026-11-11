package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() (*zap.Logger, error) {
	conf := zap.NewDevelopmentConfig()
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := conf.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger from config: %w", err)
	}

	return logger, nil
}
