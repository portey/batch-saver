package config

import (
	"github.com/portey/batch-saver/service"
	"github.com/portey/batch-saver/storage/postgres"
	"github.com/spf13/viper"
)

func Read() Config {
	viper.AutomaticEnv()

	viper.SetEnvPrefix("APP")

	viper.SetDefault("LOG_LEVEL", "DEBUG")
	viper.SetDefault("GRPC_PORT", 8080)
	viper.SetDefault("HEALTH_CHECK_PORT", 8888)

	viper.SetDefault("NUMBER_OF_WORKERS", 10)
	viper.SetDefault("BATCH_MAX_SIZE", 3)
	viper.SetDefault("BATCH_FLUSH_TIMEOUT", "1s")

	return Config{
		LogLevel:        viper.GetString("LOG_LEVEL"),
		GRPCServerPort:  viper.GetInt("GRPC_PORT"),
		HealthCheckPort: viper.GetInt("HEALTH_CHECK_PORT"),
		ServiceCfg: service.Config{
			NumberOfWorkers:   viper.GetInt("NUMBER_OF_WORKERS"),
			BatchMaxSize:      viper.GetInt("BATCH_MAX_SIZE"),
			BatchFlushTimeout: viper.GetDuration("BATCH_FLUSH_TIMEOUT"),
		},
		PostgresCfg: postgres.Config{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetInt("DB_PORT"),
			Db:       viper.GetString("DB_NAME"),
			User:     viper.GetString("DB_USERNAME"),
			Password: viper.GetString("DB_PWD"),
			Ssl:      viper.GetBool("DB_SSL"),
		},
	}
}
