package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Weather WeatherConfig `mapstructure:"weather"`
	Logging LoggingConfig `mapstructure:"logging"`
	Metrics MetricsConfig `mapstructure:"metrics"`
}

type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type WeatherConfig struct {
	BaseURL          string        `mapstructure:"base_url"`
	RequestTimeout   time.Duration `mapstructure:"request_timeout"`
	MaxRetries       int           `mapstructure:"max_retries"`
	RetryDelay       time.Duration `mapstructure:"retry_delay"`
	CacheTTL         time.Duration `mapstructure:"cache_ttl"`
	MaxConcurrentReq int           `mapstructure:"max_concurrent_req"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Environment variables override
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.idle_timeout", "60s")

	viper.SetDefault("weather.base_url", "https://api.weather.gov")
	viper.SetDefault("weather.request_timeout", "10s")
	viper.SetDefault("weather.max_retries", 3)
	viper.SetDefault("weather.retry_delay", "1s")
	viper.SetDefault("weather.cache_ttl", "15m")
	viper.SetDefault("weather.max_concurrent_req", 100)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")
}
