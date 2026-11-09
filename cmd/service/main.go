package main

import (
	"github.com/ship-monitor/hub-service/internal/config"
	"github.com/ship-monitor/hub-service/internal/di"
)

func main() {
	conf := config.New()

	di.NewApp(&conf).Run()
}
