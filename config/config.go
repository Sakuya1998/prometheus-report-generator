package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config represents the entire configuration structure.
type Config struct {
	Prometheus map[string]PrometheusConfig `yaml:"prometheus"`
	Logging    LoggingConfig               `yaml:"logging"`
}

// PrometheusConfig represents Prometheus specific configuration.
type PrometheusConfig struct {
	URL string `yaml:"url"`
}

// LoggingConfig represents logging configuration.
type LoggingConfig struct {
	Level    string `yaml:"level"`
	Path     string `yaml:"path"`
	MaxAge   int    `yaml:"max_age"`
	Compress bool   `yaml:"compress"`
}

// LoadConfig loads the configuration from the specified YAML file path.
func LoadConfig(path string) (*Config, error) {
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	// Optional: Add validation for the loaded config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// validateConfig validates the loaded configuration.
func validateConfig(config *Config) error {
	// Example validation logic
	for key, promConfig := range config.Prometheus {
		if promConfig.URL == "" {
			return fmt.Errorf("prometheus URL for '%s' is empty", key)
		}
	}
	if config.Logging.Path == "" {
		return fmt.Errorf("logging path is empty")
	}
	return nil
}
