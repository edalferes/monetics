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
		// Fallback to old behavior if fails
		fmt.Printf("Error loading configuration: %v\n", err)
		fmt.Println("Using fallback to environment variables...")
		return loadFromEnvOnly()
	}
	return config
}

// loadFromEnvOnly loads only from environment variables (fallback)
func loadFromEnvOnly() *Config {
	return &Config{
		App: AppConfig{
			Name:        getEnvOrDefault("APP_NAME", "monetics"),
			Version:     getEnvOrDefault("APP_VERSION", "1.0.0"),
			Environment: getEnvOrDefault("APP_ENVIRONMENT", "development"),
			Port:        getIntEnvOrDefault("PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     getIntEnvOrDefault("DB_PORT", 5432),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     os.Getenv("JWT_SECRET"),
			ExpiryHour: getIntEnvOrDefault("JWT_EXPIRY_HOUR", 24),
		},
		Logger: LoggerConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
		},
		Admin: AdminConfig{
			Username: getEnvOrDefault("ADMIN_USERNAME", "admin"),
			Password: getEnvOrDefault("ADMIN_PASSWORD", "root123"),
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue := parseInt(value); intValue > 0 {
			return intValue
		}
	}
	return defaultValue
}

func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

