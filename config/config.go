package config

import "github.com/portey/batch-saver/service"

type Config struct {
	LogLevel        string
	HealthCheckPort int
	GRPCServerPort  int

	ServiceCfg service.Config
}
