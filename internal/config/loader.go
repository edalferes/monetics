package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// LoadOptions options for loading configuration
type LoadOptions struct {
	ConfigPath string
	ConfigName string
	ConfigType string
}

// DefaultLoadOptions returns default options
func DefaultLoadOptions() LoadOptions {
	return LoadOptions{
		ConfigPath: ".",
		ConfigName: "config",
		ConfigType: "yaml",
	}
}

// Loader is responsible for loading configurations
type Loader struct {
	viper *viper.Viper
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		viper: viper.New(),
	}
}

// Load loads configuration from YAML files and environment variables
func (l *Loader) Load(opts ...LoadOptions) (*Config, error) {
	options := DefaultLoadOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	// Configure search for configuration files
	l.viper.SetConfigName(options.ConfigName)
	l.viper.SetConfigType(options.ConfigType)
	l.viper.AddConfigPath(options.ConfigPath)
	l.viper.AddConfigPath("./config")
	l.viper.AddConfigPath(".")

	// Set default values
	l.setDefaults()

	// Configure reading of environment variables
	// IMPORTANT: AutomaticEnv MUST be called BEFORE ReadInConfig
	// This ensures environment variables have higher priority than config file
	l.viper.AutomaticEnv()
	l.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind legacy environment variable names for backward compatibility
	l.bindLegacyEnvVars()

	// Try to read configuration file (optional)
	if err := l.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading configuration file: %w", err)
		}
		// If file not found, use only env vars and defaults
		fmt.Println("No configuration file found, using environment variables and defaults")
	} else {
		fmt.Printf("Configuration file loaded: %s\n", l.viper.ConfigFileUsed())
	}

	var config Config
	if err := l.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling configuration: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values
func (l *Loader) setDefaults() {
	// App defaults
	l.viper.SetDefault("app.name", "monetics")
	l.viper.SetDefault("app.version", "1.0.0")
	l.viper.SetDefault("app.environment", "development")
	l.viper.SetDefault("app.port", 8080)

	// Database defaults
	l.viper.SetDefault("database.host", "localhost")
	l.viper.SetDefault("database.port", 5432)
	l.viper.SetDefault("database.ssl_mode", "disable")

	// JWT defaults
	l.viper.SetDefault("jwt.expiry_hour", 24)

	// Logger defaults
	l.viper.SetDefault("logger.level", "info")
	l.viper.SetDefault("logger.format", "json")

	// Admin user defaults
	l.viper.SetDefault("admin.username", "admin")
	l.viper.SetDefault("admin.password", "root123")
}

// bindLegacyEnvVars binds legacy environment variable names for backward compatibility
func (l *Loader) bindLegacyEnvVars() {
	// Map DB_* to database.*
	l.viper.BindEnv("database.host", "DB_HOST")
	l.viper.BindEnv("database.port", "DB_PORT")
	l.viper.BindEnv("database.user", "DB_USER")
	l.viper.BindEnv("database.password", "DB_PASSWORD")
	l.viper.BindEnv("database.name", "DB_NAME")
	l.viper.BindEnv("database.ssl_mode", "DB_SSL_MODE")

	// Map APP_* and direct vars
	l.viper.BindEnv("app.name", "APP_NAME")
	l.viper.BindEnv("app.port", "PORT")
	l.viper.BindEnv("app.environment", "APP_ENVIRONMENT", "ENVIRONMENT")
	l.viper.BindEnv("app.version", "APP_VERSION")

	// Map JWT_*
	l.viper.BindEnv("jwt.secret", "JWT_SECRET")
	l.viper.BindEnv("jwt.expiry_hour", "JWT_EXPIRY_HOUR")

	// Map LOG_*
	l.viper.BindEnv("logger.level", "LOG_LEVEL")
	l.viper.BindEnv("logger.format", "LOG_FORMAT")

	// Map ADMIN_*
	l.viper.BindEnv("admin.username", "ADMIN_USERNAME")
	l.viper.BindEnv("admin.password", "ADMIN_PASSWORD")
}
