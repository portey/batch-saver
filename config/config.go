package config

import (
	"github.com/portey/batch-saver/service"
	"github.com/portey/batch-saver/storage/postgres"
)

type Config struct {
	LogLevel        string
	HealthCheckPort int
	GRPCServerPort  int

	ServiceCfg  service.Config
	PostgresCfg postgres.Config
}
