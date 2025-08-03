package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Security SecurityConfig `yaml:"security"`
	UI       UIConfig       `yaml:"ui"`
	Logging  LoggingConfig  `yaml:"logging"`
	Backup   BackupConfig   `yaml:"backup"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Path       string `yaml:"path"`
	BackupDir  string `yaml:"backup_dir"`
	MaxBackups int    `yaml:"max_backups"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	SessionTimeout string        `yaml:"session_timeout"`
	SessionConfig  SessionConfig `yaml:"session_config"`
	PasswordPolicy struct {
		MinLength      int  `yaml:"min_length"`
		RequireSpecial bool `yaml:"require_special"`
		RequireNumbers bool `yaml:"require_numbers"`
	} `yaml:"password_policy"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled                  bool     `yaml:"enabled"`
	MaxAttemptsPerIP         int      `yaml:"max_attempts_per_ip"`
	MaxAttemptsPerUser       int      `yaml:"max_attempts_per_user"`
	WindowSizeMinutes        int      `yaml:"window_size_minutes"`
	LockoutDurationMinutes   int      `yaml:"lockout_duration_minutes"`
	BackoffMultiplier        float64  `yaml:"backoff_multiplier"`
	MaxLockoutHours          int      `yaml:"max_lockout_hours"`
	WhitelistIPs             []string `yaml:"whitelist_ips"`
	EnableProgressiveLockout bool     `yaml:"enable_progressive_lockout"`
}

// SessionConfig holds session management configuration
type SessionConfig struct {
	StorageType                string `yaml:"storage_type"`                  // "memory", "database", "file"
	MaxSessionsPerUser         int    `yaml:"max_sessions_per_user"`         // ユーザーあたり最大セッション数
	ForceSingleSession         bool   `yaml:"force_single_session"`          // 単一セッション強制
	RequireIPValidation        bool   `yaml:"require_ip_validation"`         // IP検証要求
	RequireUserAgentValidation bool   `yaml:"require_user_agent_validation"` // User-Agent検証要求
	CleanupIntervalMinutes     int    `yaml:"cleanup_interval_minutes"`      // クリーンアップ間隔（分）
	SessionIDLength            int    `yaml:"session_id_length"`             // セッションID長（バイト）
	CSRFTokenLength            int    `yaml:"csrf_token_length"`             // CSRFトークン長（バイト）
	PersistenceEnabled         bool   `yaml:"persistence_enabled"`           // セッション永続化有効
}

// UIConfig holds UI-related configuration
type UIConfig struct {
	Theme    string `yaml:"theme"`
	Language string `yaml:"language"`
	FontSize int    `yaml:"font_size"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level    string `yaml:"level"`
	FilePath string `yaml:"file_path"`
}

// BackupConfig holds backup-related configuration
type BackupConfig struct {
	// 基本設定
	Enabled             bool   `yaml:"enabled"`               // バックアップ機能有効/無効
	BackupDir          string `yaml:"backup_dir"`            // バックアップディレクトリ
	
	// スケジュール設定
	AutoBackup         bool   `yaml:"auto_backup"`           // 自動バックアップ有効/無効
	ScheduleInterval   string `yaml:"schedule_interval"`     // スケジュール間隔 ("daily", "weekly", "monthly")
	ScheduleTime       string `yaml:"schedule_time"`         // 実行時刻 (HH:MM形式)
	
	// 保存管理
	MaxBackups         int    `yaml:"max_backups"`           // 最大保存世代数
	RetentionDays      int    `yaml:"retention_days"`        // 保存期間（日数）
	MaxBackupSizeMB    int    `yaml:"max_backup_size_mb"`    // 単一バックアップの最大サイズ（MB）
	
	// 暗号化設定
	EncryptionEnabled  bool   `yaml:"encryption_enabled"`    // バックアップ暗号化有効/無効
	CompressionEnabled bool   `yaml:"compression_enabled"`   // バックアップ圧縮有効/無効
	
	// 対象設定
	IncludeDatabase    bool     `yaml:"include_database"`     // データベース含める
	IncludeConfig      bool     `yaml:"include_config"`       // 設定ファイル含める
	IncludeLogs        bool     `yaml:"include_logs"`         // ログファイル含める
	ExcludePatterns    []string `yaml:"exclude_patterns"`     // 除外パターン
	
	// 整合性チェック
	VerifyBackups      bool   `yaml:"verify_backups"`        // バックアップファイル検証有効
	ChecksumEnabled    bool   `yaml:"checksum_enabled"`      // チェックサム計算有効
	
	// 通知設定
	NotifyOnSuccess    bool   `yaml:"notify_on_success"`     // 成功時通知
	NotifyOnFailure    bool   `yaml:"notify_on_failure"`     // 失敗時通知
	
	// リトライ設定
	RetryCount         int    `yaml:"retry_count"`           // 失敗時リトライ回数
	RetryIntervalSec   int    `yaml:"retry_interval_sec"`    // リトライ間隔（秒）
}

// BackupService represents the backup service interface
type BackupService interface {
	CreateBackup(ctx context.Context, req CreateBackupRequest) (*CreateBackupResponse, error)
	RestoreBackup(ctx context.Context, req RestoreBackupRequest) (*RestoreBackupResponse, error)
	ListBackups(ctx context.Context, req ListBackupsRequest) (*ListBackupsResponse, error)
	DeleteBackup(ctx context.Context, req DeleteBackupRequest) error
	ValidateBackup(ctx context.Context, req ValidateBackupRequest) (*ValidateBackupResponse, error)
	StartScheduledBackup(ctx context.Context) error
	StopScheduledBackup(ctx context.Context) error
}

// CreateBackupRequest represents backup creation request
type CreateBackupRequest struct {
	Type        BackupType `json:"type"`
	Description string     `json:"description"`
	ActorID     string     `json:"actor_id"`
}

// CreateBackupResponse represents backup creation response
type CreateBackupResponse struct {
	BackupID string    `json:"backup_id"`
	FilePath string    `json:"file_path"`
	Size     int64     `json:"size"`
	Created  time.Time `json:"created"`
}

// RestoreBackupRequest represents backup restoration request
type RestoreBackupRequest struct {
	BackupID  string `json:"backup_id"`
	ActorID   string `json:"actor_id"`
	Overwrite bool   `json:"overwrite"`
}

// RestoreBackupResponse represents backup restoration response
type RestoreBackupResponse struct {
	Success      bool      `json:"success"`
	RestoredAt   time.Time `json:"restored_at"`
	RecordsCount int       `json:"records_count"`
}

// ListBackupsRequest represents backup listing request
type ListBackupsRequest struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Type   string `json:"type,omitempty"`
}

// ListBackupsResponse represents backup listing response
type ListBackupsResponse struct {
	Backups []BackupInfo `json:"backups"`
	Total   int          `json:"total"`
}

// DeleteBackupRequest represents backup deletion request
type DeleteBackupRequest struct {
	BackupID string `json:"backup_id"`
	ActorID  string `json:"actor_id"`
}

// ValidateBackupRequest represents backup validation request
type ValidateBackupRequest struct {
	BackupID string `json:"backup_id"`
}

// ValidateBackupResponse represents backup validation response
type ValidateBackupResponse struct {
	Valid       bool   `json:"valid"`
	Error       string `json:"error,omitempty"`
	Checksum    string `json:"checksum"`
	FileSize    int64  `json:"file_size"`
	RecordCount int    `json:"record_count"`
}

// BackupInfo represents backup metadata
type BackupInfo struct {
	ID          string     `json:"id"`
	Type        BackupType `json:"type"`
	Description string     `json:"description"`
	FilePath    string     `json:"file_path"`
	Size        int64      `json:"size"`
	Checksum    string     `json:"checksum"`
	CreatedAt   time.Time  `json:"created_at"`
	CreatedBy   string     `json:"created_by"`
}

// BackupType represents the type of backup
type BackupType string

const (
	BackupTypeFull        BackupType = "full"
	BackupTypeIncremental BackupType = "incremental"
	BackupTypeManual      BackupType = "manual"
	BackupTypeScheduled   BackupType = "scheduled"
)

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *Config {
	appDataDir := getAppDataDir()

	return &Config{
		Database: DatabaseConfig{
			Path:       filepath.Join(appDataDir, "data", "shien-system.db"),
			BackupDir:  filepath.Join(appDataDir, "backups"),
			MaxBackups: 10,
		},
		Security: SecurityConfig{
			SessionTimeout: "24h",
			SessionConfig: SessionConfig{
				StorageType:                "database",
				MaxSessionsPerUser:         3,
				ForceSingleSession:         false,
				RequireIPValidation:        true,
				RequireUserAgentValidation: true,
				CleanupIntervalMinutes:     60,
				SessionIDLength:            32,
				CSRFTokenLength:            32,
				PersistenceEnabled:         true,
			},
			PasswordPolicy: struct {
				MinLength      int  `yaml:"min_length"`
				RequireSpecial bool `yaml:"require_special"`
				RequireNumbers bool `yaml:"require_numbers"`
			}{
				MinLength:      8,
				RequireSpecial: true,
				RequireNumbers: true,
			},
			RateLimit: RateLimitConfig{
				Enabled:                  true,
				MaxAttemptsPerIP:         5,
				MaxAttemptsPerUser:       3,
				WindowSizeMinutes:        15,
				LockoutDurationMinutes:   30,
				BackoffMultiplier:        2.0,
				MaxLockoutHours:          24,
				WhitelistIPs:             []string{"127.0.0.1", "::1"},
				EnableProgressiveLockout: true,
			},
		},
		UI: UIConfig{
			Theme:    "japanese",
			Language: "ja",
			FontSize: 12,
		},
		Logging: LoggingConfig{
			Level:    "info",
			FilePath: filepath.Join(appDataDir, "logs", "shien-system.log"),
		},
		Backup: BackupConfig{
			// 基本設定
			Enabled:   true,
			BackupDir: filepath.Join(appDataDir, "backups"),
			
			// スケジュール設定
			AutoBackup:       true,
			ScheduleInterval: "daily",
			ScheduleTime:     "02:00", // 深夜2時
			
			// 保存管理
			MaxBackups:      30,  // 30世代保存
			RetentionDays:   90,  // 90日間保存
			MaxBackupSizeMB: 500, // 500MB制限
			
			// 暗号化設定
			EncryptionEnabled:  true,
			CompressionEnabled: true,
			
			// 対象設定
			IncludeDatabase: true,
			IncludeConfig:   true,
			IncludeLogs:     false, // ログは除外（サイズ制約）
			ExcludePatterns: []string{
				"*.tmp",
				"*.lock",
				"*.log.old",
				".DS_Store",
			},
			
			// 整合性チェック
			VerifyBackups:   true,
			ChecksumEnabled: true,
			
			// 通知設定
			NotifyOnSuccess: false, // 成功時は通知しない
			NotifyOnFailure: true,  // 失敗時は通知
			
			// リトライ設定
			RetryCount:       3,  // 3回リトライ
			RetryIntervalSec: 60, // 60秒間隔
		},
	}
}

// LoadConfig loads configuration from file, falling back to defaults
func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	// If config file doesn't exist, create default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := GetDefaultConfig()
		if err := SaveConfig(config); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return config, nil
	}

	// Read existing config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := GetDefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for empty values
	applyDefaults(config)

	// Apply environment variable overrides
	applyEnvironmentOverrides(config)

	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure required directories exist
	if err := ensureDirectories(config); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config) error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file with restricted permissions
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig validates the configuration values
func ValidateConfig(config *Config) error {
	// Validate database path
	if config.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	// Validate backup directory
	if config.Database.BackupDir == "" {
		return fmt.Errorf("backup directory cannot be empty")
	}

	// Validate max backups
	if config.Database.MaxBackups < 1 {
		return fmt.Errorf("max backups must be at least 1")
	}

	// Validate password policy
	if config.Security.PasswordPolicy.MinLength < 4 {
		return fmt.Errorf("minimum password length must be at least 4")
	}

	// Validate session configuration
	if config.Security.SessionConfig.StorageType != "" {
		validStorageTypes := []string{"memory", "database", "file"}
		valid := false
		for _, validType := range validStorageTypes {
			if config.Security.SessionConfig.StorageType == validType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid session storage type: %s (must be one of: memory, database, file)", config.Security.SessionConfig.StorageType)
		}
	}

	if config.Security.SessionConfig.MaxSessionsPerUser < 1 {
		return fmt.Errorf("max sessions per user must be at least 1")
	}

	if config.Security.SessionConfig.CleanupIntervalMinutes < 1 {
		return fmt.Errorf("cleanup interval must be at least 1 minute")
	}

	if config.Security.SessionConfig.SessionIDLength < 16 {
		return fmt.Errorf("session ID length must be at least 16 bytes")
	}

	if config.Security.SessionConfig.CSRFTokenLength < 16 {
		return fmt.Errorf("CSRF token length must be at least 16 bytes")
	}

	// Validate UI settings
	if config.UI.FontSize < 8 || config.UI.FontSize > 24 {
		return fmt.Errorf("font size must be between 8 and 24")
	}

	// Validate backup configuration
	if config.Backup.Enabled {
		if config.Backup.BackupDir == "" {
			return fmt.Errorf("backup directory cannot be empty when backup is enabled")
		}

		// スケジュール間隔の検証
		validIntervals := []string{"daily", "weekly", "monthly"}
		valid := false
		for _, validInterval := range validIntervals {
			if config.Backup.ScheduleInterval == validInterval {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid schedule interval: %s (must be one of: daily, weekly, monthly)", config.Backup.ScheduleInterval)
		}

		// 時刻フォーマットの検証（HH:MM）
		if config.Backup.ScheduleTime != "" {
			if len(config.Backup.ScheduleTime) != 5 || config.Backup.ScheduleTime[2] != ':' {
				return fmt.Errorf("invalid schedule time format: %s (must be HH:MM)", config.Backup.ScheduleTime)
			}
		}

		// 数値範囲の検証
		if config.Backup.MaxBackups < 1 {
			return fmt.Errorf("max backups must be at least 1")
		}

		if config.Backup.RetentionDays < 1 {
			return fmt.Errorf("retention days must be at least 1")
		}

		if config.Backup.MaxBackupSizeMB < 1 {
			return fmt.Errorf("max backup size must be at least 1 MB")
		}

		if config.Backup.RetryCount < 0 {
			return fmt.Errorf("retry count cannot be negative")
		}

		if config.Backup.RetryIntervalSec < 1 {
			return fmt.Errorf("retry interval must be at least 1 second")
		}
	}

	return nil
}

// applyDefaults applies default values for empty configuration fields
func applyDefaults(config *Config) {
	defaults := GetDefaultConfig()

	if config.Database.Path == "" {
		config.Database.Path = defaults.Database.Path
	}

	if config.Database.BackupDir == "" {
		config.Database.BackupDir = defaults.Database.BackupDir
	}

	if config.Database.MaxBackups == 0 {
		config.Database.MaxBackups = defaults.Database.MaxBackups
	}

	if config.Security.SessionTimeout == "" {
		config.Security.SessionTimeout = defaults.Security.SessionTimeout
	}

	// セッション設定のデフォルト値適用
	if config.Security.SessionConfig.StorageType == "" {
		config.Security.SessionConfig.StorageType = defaults.Security.SessionConfig.StorageType
	}
	if config.Security.SessionConfig.MaxSessionsPerUser == 0 {
		config.Security.SessionConfig.MaxSessionsPerUser = defaults.Security.SessionConfig.MaxSessionsPerUser
	}
	if config.Security.SessionConfig.CleanupIntervalMinutes == 0 {
		config.Security.SessionConfig.CleanupIntervalMinutes = defaults.Security.SessionConfig.CleanupIntervalMinutes
	}
	if config.Security.SessionConfig.SessionIDLength == 0 {
		config.Security.SessionConfig.SessionIDLength = defaults.Security.SessionConfig.SessionIDLength
	}
	if config.Security.SessionConfig.CSRFTokenLength == 0 {
		config.Security.SessionConfig.CSRFTokenLength = defaults.Security.SessionConfig.CSRFTokenLength
	}

	if config.Security.PasswordPolicy.MinLength == 0 {
		config.Security.PasswordPolicy.MinLength = defaults.Security.PasswordPolicy.MinLength
	}

	if config.UI.Theme == "" {
		config.UI.Theme = defaults.UI.Theme
	}

	if config.UI.Language == "" {
		config.UI.Language = defaults.UI.Language
	}

	if config.UI.FontSize == 0 {
		config.UI.FontSize = defaults.UI.FontSize
	}

	if config.Logging.Level == "" {
		config.Logging.Level = defaults.Logging.Level
	}

	if config.Logging.FilePath == "" {
		config.Logging.FilePath = defaults.Logging.FilePath
	}

	// バックアップ設定のデフォルト値適用
	if config.Backup.BackupDir == "" {
		config.Backup.BackupDir = defaults.Backup.BackupDir
	}

	if config.Backup.ScheduleInterval == "" {
		config.Backup.ScheduleInterval = defaults.Backup.ScheduleInterval
	}

	if config.Backup.ScheduleTime == "" {
		config.Backup.ScheduleTime = defaults.Backup.ScheduleTime
	}

	if config.Backup.MaxBackups == 0 {
		config.Backup.MaxBackups = defaults.Backup.MaxBackups
	}

	if config.Backup.RetentionDays == 0 {
		config.Backup.RetentionDays = defaults.Backup.RetentionDays
	}

	if config.Backup.MaxBackupSizeMB == 0 {
		config.Backup.MaxBackupSizeMB = defaults.Backup.MaxBackupSizeMB
	}

	if config.Backup.ExcludePatterns == nil {
		config.Backup.ExcludePatterns = defaults.Backup.ExcludePatterns
	}

	if config.Backup.RetryCount == 0 {
		config.Backup.RetryCount = defaults.Backup.RetryCount
	}

	if config.Backup.RetryIntervalSec == 0 {
		config.Backup.RetryIntervalSec = defaults.Backup.RetryIntervalSec
	}
}

// applyEnvironmentOverrides applies environment variable overrides
func applyEnvironmentOverrides(config *Config) {
	if dbPath := os.Getenv("SHIEN_DB_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}

	if backupDir := os.Getenv("SHIEN_BACKUP_DIR"); backupDir != "" {
		config.Database.BackupDir = backupDir
		config.Backup.BackupDir = backupDir // 新しいバックアップ設定にも適用
	}

	if sessionTimeout := os.Getenv("SHIEN_SESSION_TIMEOUT"); sessionTimeout != "" {
		config.Security.SessionTimeout = sessionTimeout
	}

	if logLevel := os.Getenv("SHIEN_LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}

	if logFile := os.Getenv("SHIEN_LOG_FILE"); logFile != "" {
		config.Logging.FilePath = logFile
	}

	// バックアップ関連の環境変数オーバーライド
	if backupEnabled := os.Getenv("SHIEN_BACKUP_ENABLED"); backupEnabled != "" {
		config.Backup.Enabled = backupEnabled == "true"
	}

	if autoBackup := os.Getenv("SHIEN_AUTO_BACKUP"); autoBackup != "" {
		config.Backup.AutoBackup = autoBackup == "true"
	}

	if scheduleInterval := os.Getenv("SHIEN_BACKUP_INTERVAL"); scheduleInterval != "" {
		config.Backup.ScheduleInterval = scheduleInterval
	}

	if scheduleTime := os.Getenv("SHIEN_BACKUP_TIME"); scheduleTime != "" {
		config.Backup.ScheduleTime = scheduleTime
	}

	if encryption := os.Getenv("SHIEN_BACKUP_ENCRYPTION"); encryption != "" {
		config.Backup.EncryptionEnabled = encryption == "true"
	}
}

// ensureDirectories creates necessary directories
func ensureDirectories(config *Config) error {
	dirs := []string{
		filepath.Dir(config.Database.Path),
		config.Database.BackupDir,
		filepath.Dir(config.Logging.FilePath),
		config.Backup.BackupDir, // バックアップディレクトリも作成
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	configDir := getConfigDir()
	return filepath.Join(configDir, "config.yaml")
}

// getConfigDir returns the configuration directory based on OS
func getConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "ShienSystem")
		}
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "ShienSystem")
	case "darwin":
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "Library", "Application Support", "ShienSystem")
	default: // Linux and others
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			return filepath.Join(xdgConfig, "ShienSystem")
		}
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".config", "ShienSystem")
	}
}

// getAppDataDir returns the application data directory based on OS
func getAppDataDir() string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "ShienSystem")
		}
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "ShienSystem")
	case "darwin":
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "Library", "Application Support", "ShienSystem")
	default: // Linux and others
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return filepath.Join(xdgData, "ShienSystem")
		}
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".local", "share", "ShienSystem")
	}
}

// GetMigrationDir returns the path to migration files
func GetMigrationDir() string {
	// In development, use relative path
	if _, err := os.Stat("migrations"); err == nil {
		return "migrations"
	}

	// In production, look for migrations relative to executable
	execPath, err := os.Executable()
	if err == nil {
		migrationPath := filepath.Join(filepath.Dir(execPath), "migrations")
		if _, err := os.Stat(migrationPath); err == nil {
			return migrationPath
		}
	}

	// Fallback to current directory
	return "migrations"
}

// CreateDefaultConfigFile creates a default configuration file with comments
func CreateDefaultConfigFile() error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // File already exists
	}

	configContent := `# 障害者サービス管理システム設定ファイル
# このファイルを編集して設定をカスタマイズできます

# データベース設定
database:
  # データベースファイルのパス（環境変数 SHIEN_DB_PATH で上書き可能）
  path: ""  # 空の場合はデフォルトパスを使用
  # バックアップディレクトリ
  backup_dir: ""  # 空の場合はデフォルトパスを使用
  # 保持するバックアップ数
  max_backups: 10

# セキュリティ設定
security:
  # セッションタイムアウト（例: 24h, 30m, 1h30m）
  session_timeout: "24h"
  # セッション管理設定
  session_config:
    # セッション保存方式: "memory", "database", "file"
    storage_type: "database"
    # ユーザーあたり最大セッション数
    max_sessions_per_user: 3
    # 単一セッション強制（trueの場合、新しいログインで既存セッション無効化）
    force_single_session: false
    # IPアドレス検証を要求
    require_ip_validation: true
    # User-Agent検証を要求
    require_user_agent_validation: true
    # セッションクリーンアップ間隔（分）
    cleanup_interval_minutes: 60
    # セッションID長（バイト）
    session_id_length: 32
    # CSRFトークン長（バイト）
    csrf_token_length: 32
    # セッション永続化有効
    persistence_enabled: true
  # パスワードポリシー
  password_policy:
    min_length: 8
    require_special: true
    require_numbers: true

# UI設定
ui:
  theme: "japanese"
  language: "ja"
  font_size: 12

# ログ設定
logging:
  # ログレベル: debug, info, warn, error
  level: "info"
  # ログファイルのパス
  file_path: ""  # 空の場合はデフォルトパスを使用
`

	// Write config file with restricted permissions
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
