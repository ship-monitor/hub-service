package main

import (
	"fmt"
	"os"

	"github.com/ship-monitor/hub-service/internal/config"
	"github.com/ship-monitor/hub-service/internal/di"
	"github.com/ship-monitor/hub-service/internal/logger"
)

func main() {
	conf := config.New()

	logger, err := logger.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: New logger: %s", err.Error())
		os.Exit(1)
	}

	di.NewApp(&conf, logger).Run()
}
