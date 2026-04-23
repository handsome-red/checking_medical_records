package config

import (
	"fmt"
	"time"
	"github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct  {
	Port string						`mapstructure:"port"`
	ReadTimeout	time.Duration		`mapstructure:"read_timeout"`
	WriteTimeout	time.Duration	`mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
    Driver       string `mapstructure:"driver"`
    DSN          string `mapstructure:"dsn"`           
    MaxOpenConns int    `mapstructure:"max_open_conns"`
}

func Load() (*Config, error){
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshal config: %w", err)
	}

	if cfg.Server.Port != "" && cfg.Server.Port[0] != ':' {
		cfg.Server.Port = ":" + cfg.Server.Port
	}

	return cfg, nil
}