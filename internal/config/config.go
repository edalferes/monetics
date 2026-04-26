package config

import (
	"fmt"
	"os"
)

// Load loads configuration with validations
func Load(opts ...LoadOptions) (*Config, error) {
	loader := NewLoader()
	validator := NewValidator()

	config, err := loader.Load(opts...)
	if err != nil {
		return nil, err
	}

	if err := validator.Validate(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadConfig loads configuration with default options (for compatibility)
func LoadConfig() *Config {
	config, err := Load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		// exit app
		os.Exit(1)
	}
	return config
}