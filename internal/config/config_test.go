package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	assert.NotEmpty(t, config.Database.Path)
	assert.NotEmpty(t, config.Database.BackupDir)
	assert.Equal(t, 10, config.Database.MaxBackups)
	assert.Equal(t, "24h", config.Security.SessionTimeout)
	assert.Equal(t, 8, config.Security.PasswordPolicy.MinLength)
	assert.True(t, config.Security.PasswordPolicy.RequireSpecial)
	assert.True(t, config.Security.PasswordPolicy.RequireNumbers)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      GetDefaultConfig(),
			expectError: false,
		},
		{
			name: "empty database path",
			config: &Config{
				Database: DatabaseConfig{
					Path:       "",
					BackupDir:  "/tmp/backup",
					MaxBackups: 10,
				},
			},
			expectError: true,
		},
		{
			name: "empty backup dir",
			config: &Config{
				Database: DatabaseConfig{
					Path:       "/tmp/db.sqlite",
					BackupDir:  "",
					MaxBackups: 10,
				},
			},
			expectError: true,
		},
		{
			name: "invalid max backups",
			config: &Config{
				Database: DatabaseConfig{
					Path:       "/tmp/db.sqlite",
					BackupDir:  "/tmp/backup",
					MaxBackups: 0,
				},
			},
			expectError: true,
		},
		{
			name: "invalid password length",
			config: &Config{
				Database: DatabaseConfig{
					Path:       "/tmp/db.sqlite",
					BackupDir:  "/tmp/backup",
					MaxBackups: 10,
				},
				Security: SecurityConfig{
					PasswordPolicy: struct {
						MinLength      int  `yaml:"min_length"`
						RequireSpecial bool `yaml:"require_special"`
						RequireNumbers bool `yaml:"require_numbers"`
					}{
						MinLength: 3,
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid font size",
			config: &Config{
				Database: DatabaseConfig{
					Path:       "/tmp/db.sqlite",
					BackupDir:  "/tmp/backup",
					MaxBackups: 10,
				},
				Security: SecurityConfig{
					PasswordPolicy: struct {
						MinLength      int  `yaml:"min_length"`
						RequireSpecial bool `yaml:"require_special"`
						RequireNumbers bool `yaml:"require_numbers"`
					}{
						MinLength: 8,
					},
				},
				UI: UIConfig{
					FontSize: 50,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	// Save original environment
	originalDBPath := os.Getenv("SHIEN_DB_PATH")
	originalBackupDir := os.Getenv("SHIEN_BACKUP_DIR")
	originalSessionTimeout := os.Getenv("SHIEN_SESSION_TIMEOUT")
	originalLogLevel := os.Getenv("SHIEN_LOG_LEVEL")
	originalLogFile := os.Getenv("SHIEN_LOG_FILE")

	// Clean up
	defer func() {
		os.Setenv("SHIEN_DB_PATH", originalDBPath)
		os.Setenv("SHIEN_BACKUP_DIR", originalBackupDir)
		os.Setenv("SHIEN_SESSION_TIMEOUT", originalSessionTimeout)
		os.Setenv("SHIEN_LOG_LEVEL", originalLogLevel)
		os.Setenv("SHIEN_LOG_FILE", originalLogFile)
	}()

	// Set test environment variables
	testDBPath := "/test/custom/db.sqlite"
	testBackupDir := "/test/custom/backup"
	testSessionTimeout := "12h"
	testLogLevel := "debug"
	testLogFile := "/test/custom/log.txt"

	os.Setenv("SHIEN_DB_PATH", testDBPath)
	os.Setenv("SHIEN_BACKUP_DIR", testBackupDir)
	os.Setenv("SHIEN_SESSION_TIMEOUT", testSessionTimeout)
	os.Setenv("SHIEN_LOG_LEVEL", testLogLevel)
	os.Setenv("SHIEN_LOG_FILE", testLogFile)

	config := GetDefaultConfig()
	applyEnvironmentOverrides(config)

	assert.Equal(t, testDBPath, config.Database.Path)
	assert.Equal(t, testBackupDir, config.Database.BackupDir)
	assert.Equal(t, testSessionTimeout, config.Security.SessionTimeout)
	assert.Equal(t, testLogLevel, config.Logging.Level)
	assert.Equal(t, testLogFile, config.Logging.FilePath)
}

func TestCreateDefaultConfigFile(t *testing.T) {
	// Test creating default config file
	err := CreateDefaultConfigFile()
	assert.NoError(t, err)

	// Verify that calling it again doesn't fail (should not overwrite existing file)
	err = CreateDefaultConfigFile()
	assert.NoError(t, err)
}

func TestGetMigrationDir(t *testing.T) {
	// Test when migrations directory exists in current directory
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	os.Chdir(tempDir)

	// Create migrations directory
	migrationDir := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationDir, 0755)
	require.NoError(t, err)

	result := GetMigrationDir()
	assert.Equal(t, "migrations", result)

	// Test when migrations directory doesn't exist
	os.RemoveAll(migrationDir)
	result = GetMigrationDir()
	assert.Equal(t, "migrations", result) // Should still return fallback
}

func TestGetAppDataDir(t *testing.T) {
	appDataDir := getAppDataDir()
	assert.NotEmpty(t, appDataDir)
	assert.Contains(t, appDataDir, "ShienSystem")
}

func TestGetConfigDir(t *testing.T) {
	configDir := getConfigDir()
	assert.NotEmpty(t, configDir)
	assert.Contains(t, configDir, "ShienSystem")
}

func TestLoadConfig(t *testing.T) {
	// Test loading configuration (should create default if not exists)
	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Verify default values are applied
	assert.NotEmpty(t, config.Database.Path)
	assert.NotEmpty(t, config.Database.BackupDir)
	assert.Equal(t, 10, config.Database.MaxBackups)
	assert.Equal(t, "24h", config.Security.SessionTimeout)
	assert.Equal(t, 8, config.Security.PasswordPolicy.MinLength)
}

func TestApplyDefaults(t *testing.T) {
	// Create config with empty values
	config := &Config{}

	// Apply defaults
	applyDefaults(config)

	// Verify defaults were applied
	assert.NotEmpty(t, config.Database.Path)
	assert.NotEmpty(t, config.Database.BackupDir)
	assert.Equal(t, 10, config.Database.MaxBackups)
	assert.Equal(t, "24h", config.Security.SessionTimeout)
	assert.Equal(t, 8, config.Security.PasswordPolicy.MinLength)
	assert.Equal(t, "japanese", config.UI.Theme)
	assert.Equal(t, "ja", config.UI.Language)
	assert.Equal(t, 12, config.UI.FontSize)
	assert.Equal(t, "info", config.Logging.Level)
	assert.NotEmpty(t, config.Logging.FilePath)
}
