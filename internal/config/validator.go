package config

import (
	"fmt"
	"strings"
)

// Validator is responsible for validating configurations
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks if the configuration has the required fields
func (v *Validator) Validate(config *Config) error {
	if config == nil {
		return fmt.Errorf("config must not be nil")
	}

	if config.App.Port <= 0 || config.App.Port > 65535 {
		return fmt.Errorf("app.port must be between 1 and 65535")
	}

	// Validate environment
	validEnvs := []string{"development", "staging", "production"}
	found := false
	for _, env := range validEnvs {
		if config.App.Environment == env {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("app.environment must be one of: %s", strings.Join(validEnvs, ", "))
	}

	// Validate database
	if strings.TrimSpace(config.Database.Host) == "" {
		return fmt.Errorf("database.host is required")
	}
	if strings.TrimSpace(config.Database.User) == "" {
		return fmt.Errorf("database.user is required")
	}
	if strings.TrimSpace(config.Database.Name) == "" {
		return fmt.Errorf("database.name is required")
	}
	if config.Database.Port < 0 || config.Database.Port > 65535 {
		return fmt.Errorf("database.port must be between 0 and 65535")
	}

	// Validate JWT
	if strings.TrimSpace(config.JWT.Secret) == "" {
		return fmt.Errorf("jwt.secret is required")
	}

	return nil
}
