package config

import "fmt"

// Config represents the complete application configuration
type Config struct {
	App      AppConfig      `mapstructure:"app" yaml:"app"`
	Database DatabaseConfig `mapstructure:"database" yaml:"database"`
	JWT      JWTConfig      `mapstructure:"jwt" yaml:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger" yaml:"logger"`
	Admin    AdminConfig    `mapstructure:"admin" yaml:"admin"`
}

// AppConfig application configuration
type AppConfig struct {
	Name        string `mapstructure:"name" yaml:"name"`
	Version     string `mapstructure:"version" yaml:"version"`
	Environment string `mapstructure:"environment" yaml:"environment"`
	Port        int    `mapstructure:"port" yaml:"port"`
}

// DatabaseConfig database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	Name     string `mapstructure:"name" yaml:"name"`
	SSLMode  string `mapstructure:"ssl_mode" yaml:"ssl_mode"`
}

// JWTConfig JWT configuration
type JWTConfig struct {
	Secret     string `mapstructure:"secret" yaml:"secret"`
	ExpiryHour int    `mapstructure:"expiry_hour" yaml:"expiry_hour"`
}

// LoggerConfig logger configuration
type LoggerConfig struct {
	Level  string `mapstructure:"level" yaml:"level"`
	Format string `mapstructure:"format" yaml:"format"`
}

// AdminConfig admin user credentials for seeding
type AdminConfig struct {
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
}

// Utility methods for Config

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// IsDevelopment checks if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction checks if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsStaging checks if running in staging environment
func (c *Config) IsStaging() bool {
	return c.App.Environment == "staging"
}
