package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderLoad_UsesDefaultsWhenConfigFileIsMissing(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()

	cfg, err := loader.Load(LoadOptions{
		ConfigPath: tempDir,
		ConfigName: "missing",
		ConfigType: "yaml",
	})
	if err != nil {
		t.Fatalf("expected no error when config file is missing, got: %v", err)
	}

	if cfg.Database.Host != "localhost" {
		t.Fatalf("expected default database host 'localhost', got: %q", cfg.Database.Host)
	}

	if cfg.Database.Name != "monetics" {
		t.Fatalf("expected default database name 'monetics', got: %q", cfg.Database.Name)
	}

	if cfg.App.Port != 8080 {
		t.Fatalf("expected default app port 8080, got: %d", cfg.App.Port)
	}

	if cfg.JWT.ExpiryHour != 24 {
		t.Fatalf("expected default jwt.expiry_hour 24, got: %d", cfg.JWT.ExpiryHour)
	}
}

func TestLoaderLoad_LoadsValuesFromConfigFile(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()

	content := []byte(`
app:
  port: 8181
database:
  host: "db.example.com"
  port: 5433
  name: "testdb"
jwt:
  secret: "file-secret"
`)

	if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), content, 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := loader.Load(LoadOptions{
		ConfigPath: tempDir,
		ConfigName: "config",
		ConfigType: "yaml",
	})
	if err != nil {
		t.Fatalf("expected config file to load successfully, got: %v", err)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Fatalf("expected database host from file, got: %q", cfg.Database.Host)
	}

	if cfg.Database.Port != 5433 {
		t.Fatalf("expected database port from file, got: %d", cfg.Database.Port)
	}

	if cfg.App.Port != 8181 {
		t.Fatalf("expected app.port from file, got: %d", cfg.App.Port)
	}

	if cfg.JWT.Secret != "file-secret" {
		t.Fatalf("expected jwt.secret from file, got: %q", cfg.JWT.Secret)
	}
}

func TestLoaderLoad_EnvironmentVariablesOverrideConfigFile(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()

	content := []byte(`
app:
  port: 8181
database:
  host: "db.example.com"
  port: 5433
  name: "testdb"
jwt:
  secret: "file-secret"
`)

	if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), content, 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	t.Setenv("DATABASE_HOST", "override.host.com")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("JWT_SECRET", "env-secret")

	cfg, err := loader.Load(LoadOptions{
		ConfigPath: tempDir,
		ConfigName: "config",
		ConfigType: "yaml",
	})
	if err != nil {
		t.Fatalf("expected successful load with env override, got: %v", err)
	}

	if cfg.Database.Host != "override.host.com" {
		t.Fatalf("expected DATABASE_HOST to override file, got: %q", cfg.Database.Host)
	}

	if cfg.App.Port != 9090 {
		t.Fatalf("expected APP_PORT to override file, got: %d", cfg.App.Port)
	}

	if cfg.JWT.Secret != "env-secret" {
		t.Fatalf("expected JWT_SECRET to override file, got: %q", cfg.JWT.Secret)
	}
}

func validConfigForValidation() *Config {
	return &Config{
		App: AppConfig{
			Name:        "monetics",
			Version:     "1.0.0",
			Environment: "development",
			Port:        8080,
		},
		Database: DatabaseConfig{
			Host:    "localhost",
			Port:    5432,
			User:    "postgres",
			Name:    "monetics",
			SSLMode: "disable",
		},
		JWT: JWTConfig{
			Secret:     "secret-123",
			ExpiryHour: 24,
		},
	}
}

func TestValidatorValidate(t *testing.T) {
	validator := NewValidator()

	t.Run("valid config", func(t *testing.T) {
		cfg := validConfigForValidation()
		if err := validator.Validate(cfg); err != nil {
			t.Fatalf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("nil config", func(t *testing.T) {
		err := validator.Validate(nil)
		if err == nil || err.Error() != "config must not be nil" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid port", func(t *testing.T) {
		cfg := validConfigForValidation()
		cfg.App.Port = 0
		err := validator.Validate(cfg)
		if err == nil || err.Error() != "app.port must be between 1 and 65535" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid environment", func(t *testing.T) {
		cfg := validConfigForValidation()
		cfg.App.Environment = "qa"
		err := validator.Validate(cfg)
		if err == nil || err.Error() != "app.environment must be one of: development, staging, production" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing database host", func(t *testing.T) {
		cfg := validConfigForValidation()
		cfg.Database.Host = "  "
		err := validator.Validate(cfg)
		if err == nil || err.Error() != "database.host is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing database user", func(t *testing.T) {
		cfg := validConfigForValidation()
		cfg.Database.User = ""
		err := validator.Validate(cfg)
		if err == nil || err.Error() != "database.user is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing database name", func(t *testing.T) {
		cfg := validConfigForValidation()
		cfg.Database.Name = ""
		err := validator.Validate(cfg)
		if err == nil || err.Error() != "database.name is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing jwt secret", func(t *testing.T) {
		cfg := validConfigForValidation()
		cfg.JWT.Secret = "  "
		err := validator.Validate(cfg)
		if err == nil || err.Error() != "jwt.secret is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ai disabled skips ai validation", func(t *testing.T) {
		cfg := validConfigForValidation()
		cfg.AI = AIConfig{Enabled: false}
		if err := validator.Validate(cfg); err != nil {
			t.Fatalf("Validate() unexpected error when AI disabled: %v", err)
		}
	})

	t.Run("ai enabled requires api key", func(t *testing.T) {
		cfg := validConfigForValidation()
		cfg.AI = AIConfig{
			Enabled:            true,
			Provider:           "openai",
			Model:              "gpt-4o-mini",
			MaxItemsPerRequest: 500,
			MinConfidence:      0.4,
		}
		err := validator.Validate(cfg)
		if err == nil || err.Error() != "ai.api_key is required when ai.enabled is true" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ai enabled requires min_confidence in [0,1]", func(t *testing.T) {
		cfg := validConfigForValidation()
		cfg.AI = AIConfig{
			Enabled:            true,
			Provider:           "openai",
			APIKey:             "sk-test",
			Model:              "gpt-4o-mini",
			MaxItemsPerRequest: 500,
			MinConfidence:      1.5,
		}
		err := validator.Validate(cfg)
		if err == nil || err.Error() != "ai.min_confidence must be between 0 and 1" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ai enabled valid config", func(t *testing.T) {
		cfg := validConfigForValidation()
		cfg.AI = AIConfig{
			Enabled:            true,
			Provider:           "openai",
			APIKey:             "sk-test",
			Model:              "gpt-4o-mini",
			MaxItemsPerRequest: 500,
			MinConfidence:      0.4,
		}
		if err := validator.Validate(cfg); err != nil {
			t.Fatalf("Validate() unexpected error: %v", err)
		}
	})
}

func TestLoaderLoad_AIDefaults(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()

	cfg, err := loader.Load(LoadOptions{
		ConfigPath: tempDir,
		ConfigName: "missing",
		ConfigType: "yaml",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.AI.Enabled {
		t.Fatalf("expected ai.enabled default false")
	}
	if cfg.AI.Provider != "openai" {
		t.Fatalf("expected ai.provider default 'openai', got: %q", cfg.AI.Provider)
	}
	if cfg.AI.Model != "gpt-4o-mini" {
		t.Fatalf("expected ai.model default 'gpt-4o-mini', got: %q", cfg.AI.Model)
	}
	if cfg.AI.MaxItemsPerRequest != 500 {
		t.Fatalf("expected ai.max_items_per_request default 500, got: %d", cfg.AI.MaxItemsPerRequest)
	}
	if cfg.AI.MinConfidence != 0.4 {
		t.Fatalf("expected ai.min_confidence default 0.4, got: %v", cfg.AI.MinConfidence)
	}
}

func TestLoaderLoad_AIEnvOverrides(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()

	t.Setenv("AI_ENABLED", "true")
	t.Setenv("AI_API_KEY", "sk-env-test")
	t.Setenv("AI_MODEL", "gpt-4o")

	cfg, err := loader.Load(LoadOptions{
		ConfigPath: tempDir,
		ConfigName: "missing",
		ConfigType: "yaml",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !cfg.AI.Enabled {
		t.Fatalf("expected AI_ENABLED env to enable AI")
	}
	if cfg.AI.APIKey != "sk-env-test" {
		t.Fatalf("expected AI_API_KEY override, got: %q", cfg.AI.APIKey)
	}
	if cfg.AI.Model != "gpt-4o" {
		t.Fatalf("expected AI_MODEL override, got: %q", cfg.AI.Model)
	}
}
