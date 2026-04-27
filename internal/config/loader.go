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
	l.bindEnvs()

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

func (l *Loader) bindEnvs() {
	for _, key := range l.viper.AllKeys() {
		_ = l.viper.BindEnv(key)
	}
}

// setDefaults sets default values
func (l *Loader) setDefaults() {
	// App defaults
	l.viper.SetDefault("app.name", "monetics")
	l.viper.SetDefault("app.version", "1.0.0")
	l.viper.SetDefault("app.environment", "development")
	l.viper.SetDefault("app.port", 8080)

	// Logger defaults
	l.viper.SetDefault("logger.level", "info")
	l.viper.SetDefault("logger.format", "json")

	// Database defaults
	l.viper.SetDefault("database.host", "localhost")
	l.viper.SetDefault("database.port", 5432)
	l.viper.SetDefault("database.user", "postgres")
	l.viper.SetDefault("database.password", "")
	l.viper.SetDefault("database.name", "monetics")
	l.viper.SetDefault("database.ssl_mode", "disable")

	// JWT defaults
	l.viper.SetDefault("jwt.secret", "")
	l.viper.SetDefault("jwt.expiry_hour", 24)

	// Admin user defaults
	l.viper.SetDefault("admin.username", "admin")
	l.viper.SetDefault("admin.password", "root123")

	// AI defaults (disabled by default)
	l.viper.SetDefault("ai.enabled", false)
	l.viper.SetDefault("ai.provider", "openai")
	l.viper.SetDefault("ai.api_key", "")
	l.viper.SetDefault("ai.model", "gpt-4o-mini")
	l.viper.SetDefault("ai.base_url", "https://api.openai.com/v1")
	l.viper.SetDefault("ai.timeout_seconds", 60)
	l.viper.SetDefault("ai.max_items_per_request", 500)
	l.viper.SetDefault("ai.min_confidence", 0.4)
	l.viper.SetDefault("ai.history_lookback_days", 90)
	l.viper.SetDefault("ai.history_max_examples", 20)
	l.viper.SetDefault("ai.rate_limit_per_minute", 10)
}
