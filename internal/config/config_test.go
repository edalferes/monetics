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
}
