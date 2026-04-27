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

	// Validate AI (only when enabled)
	if config.AI.Enabled {
		if strings.TrimSpace(config.AI.APIKey) == "" {
			return fmt.Errorf("ai.api_key is required when ai.enabled is true")
		}
		if strings.TrimSpace(config.AI.Provider) == "" {
			return fmt.Errorf("ai.provider is required when ai.enabled is true")
		}
		if strings.TrimSpace(config.AI.Model) == "" {
			return fmt.Errorf("ai.model is required when ai.enabled is true")
		}
		if config.AI.MaxItemsPerRequest <= 0 {
			return fmt.Errorf("ai.max_items_per_request must be > 0")
		}
		if config.AI.MinConfidence < 0 || config.AI.MinConfidence > 1 {
			return fmt.Errorf("ai.min_confidence must be between 0 and 1")
		}
	}

	return nil
}
