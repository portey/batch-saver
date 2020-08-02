package config

import "github.com/spf13/viper"

func Read() Config {
	viper.AutomaticEnv()

	viper.SetEnvPrefix("APP")

	viper.SetDefault("LOG_LEVEL", "DEBUG")
	viper.SetDefault("GRPC_PORT", 8080)
	viper.SetDefault("HEALTH_CHECK_PORT", 8888)

	return Config{
		LogLevel:        viper.GetString("LOG_LEVEL"),
		GRPCServerPort:  viper.GetInt("GRPC_PORT"),
		HealthCheckPort: viper.GetInt("HEALTH_CHECK_PORT"),
	}
}
