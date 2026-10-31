package main

import (
	"github.com/google/uuid"
	"github.com/ship-monitor/hub-service/internal/config"
	"github.com/ship-monitor/hub-service/internal/di"
)

var DeviceID = uuid.New()

func main() {
	conf := config.New()

	di.NewApp(&conf).Run()
}
