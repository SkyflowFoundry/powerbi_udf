// internal/config/config.go
package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	Skyflow SkyflowConfig `mapstructure:"skyflow"`
	Server  ServerConfig  `mapstructure:"server"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// SkyflowConfig holds all Skyflow-specific configuration
type SkyflowConfig struct {
	VaultURL           string `mapstructure:"vault_url"`
	BearerToken        string `mapstructure:"bearer_token"`
	BatchSize          int    `mapstructure:"batch_size"`
	ParallelCalls      int    `mapstructure:"parallel_calls"`
	RequestTimeoutSecs int    `mapstructure:"request_timeout_seconds"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// LoggingConfig holds logging-specific configuration
type LoggingConfig struct {
	Level     string `mapstructure:"level"`      // "info" or "debug"
	LogToFile bool   `mapstructure:"log_to_file"` // true to log to file, false for console
	FilePath  string `mapstructure:"file_path"`  // path to log file if LogToFile is true
}

// LoadConfig reads configuration from file or environment variables
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	// Set default values
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("skyflow.batch_size", 25)
	viper.SetDefault("skyflow.parallel_calls", 10)
	viper.SetDefault("skyflow.request_timeout_seconds", 30)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.log_to_file", false)
	viper.SetDefault("logging.file_path", "./logs/skyflow_detokenizer.log")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Unable to read config file: %v", err)
		log.Println("Will attempt to use environment variables...")
	}

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Validate required configuration
	if config.Skyflow.VaultURL == "" {
		return nil, fmt.Errorf("skyflow vault_url is required")
	}
	if config.Skyflow.BearerToken == "" {
		return nil, fmt.Errorf("skyflow bearer_token is required")
	}

	return &config, nil
}
