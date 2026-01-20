package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv        string `mapstructure:"APP_ENV"`
	Port          string `mapstructure:"PORT"`
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	JWTSecret     string `mapstructure:"JWT_SECRET"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("APP_ENV", "dev")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("JWT_SECRET", "lingbao-secret-key-change-me")

	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Clean up port if it has a colon
	if strings.HasPrefix(cfg.Port, ":") {
		cfg.Port = cfg.Port[1:]
	}

	return &cfg, nil
}
