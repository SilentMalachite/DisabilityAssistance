package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"shien-system/internal/domain"
)

// RateLimitConfigRepository implements domain.RateLimitConfigRepository
type RateLimitConfigRepository struct {
	db *Database
}

// NewRateLimitConfigRepository creates a new RateLimitConfigRepository instance
func NewRateLimitConfigRepository(db *Database) *RateLimitConfigRepository {
	return &RateLimitConfigRepository{
		db: db,
	}
}

// Get retrieves the current rate limit configuration
func (r *RateLimitConfigRepository) Get(ctx context.Context) (*RateLimitConfig, error) {
	query := `
		SELECT id, max_attempts_per_ip, max_attempts_per_user, window_size_minutes,
		       lockout_duration_minutes, backoff_multiplier, max_lockout_hours,
		       whitelist_ips, enable_progressive_lockout, created_at, updated_at
		FROM rate_limit_config
		ORDER BY updated_at DESC
		LIMIT 1
	`

	row := r.getExecutor(ctx).QueryRowContext(ctx, query)
	config, err := r.scanRateLimitConfig(row)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default configuration if no config exists
			return r.GetDefault(), nil
		}
		return nil, fmt.Errorf("failed to get rate limit config: %w", err)
	}

	return config, nil
}

// Update updates the rate limit configuration
func (r *RateLimitConfigRepository) Update(ctx context.Context, config *RateLimitConfig) error {
	// Set ID if not provided
	if config.ID == "" {
		config.ID = "default-config"
	}

	config.UpdatedAt = time.Now()

	// Serialize whitelist IPs to JSON
	whitelistJSON, err := json.Marshal(config.WhitelistIPs)
	if err != nil {
		return fmt.Errorf("failed to marshal whitelist IPs: %w", err)
	}

	// Try to update first
	updateQuery := `
		UPDATE rate_limit_config SET
			max_attempts_per_ip = ?,
			max_attempts_per_user = ?,
			window_size_minutes = ?,
			lockout_duration_minutes = ?,
			backoff_multiplier = ?,
			max_lockout_hours = ?,
			whitelist_ips = ?,
			enable_progressive_lockout = ?,
			updated_at = ?
		WHERE id = ?
	`

	result, err := r.getExecutor(ctx).ExecContext(ctx, updateQuery,
		config.MaxAttemptsPerIP,
		config.MaxAttemptsPerUser,
		config.WindowSizeMinutes,
		config.LockoutDurationMinutes,
		config.BackoffMultiplier,
		config.MaxLockoutHours,
		string(whitelistJSON),
		config.EnableProgressiveLockout,
		config.UpdatedAt.Format(time.RFC3339),
		config.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update rate limit config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// If no rows were updated, insert new record
	if rowsAffected == 0 {
		config.CreatedAt = config.UpdatedAt

		insertQuery := `
			INSERT INTO rate_limit_config (
				id, max_attempts_per_ip, max_attempts_per_user, window_size_minutes,
				lockout_duration_minutes, backoff_multiplier, max_lockout_hours,
				whitelist_ips, enable_progressive_lockout, created_at, updated_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err = r.getExecutor(ctx).ExecContext(ctx, insertQuery,
			config.ID,
			config.MaxAttemptsPerIP,
			config.MaxAttemptsPerUser,
			config.WindowSizeMinutes,
			config.LockoutDurationMinutes,
			config.BackoffMultiplier,
			config.MaxLockoutHours,
			string(whitelistJSON),
			config.EnableProgressiveLockout,
			config.CreatedAt.Format(time.RFC3339),
			config.UpdatedAt.Format(time.RFC3339),
		)

		if err != nil {
			return fmt.Errorf("failed to insert rate limit config: %w", err)
		}
	}

	return nil
}

// GetDefault returns the default rate limit configuration
func (r *RateLimitConfigRepository) GetDefault() *RateLimitConfig {
	return &RateLimitConfig{
		ID:                       "default-config",
		MaxAttemptsPerIP:         5,
		MaxAttemptsPerUser:       3,
		WindowSizeMinutes:        15,
		LockoutDurationMinutes:   30,
		BackoffMultiplier:        2.0,
		MaxLockoutHours:          24,
		WhitelistIPs:             []string{},
		EnableProgressiveLockout: true,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}
}

// Helper methods

func (r *RateLimitConfigRepository) getExecutor(ctx context.Context) executor {
	if tx, ok := ctx.Value("tx").(*sql.Tx); ok {
		return tx
	}
	return r.db.DB()
}

func (r *RateLimitConfigRepository) scanRateLimitConfig(scanner scanner) (*RateLimitConfig, error) {
	var config RateLimitConfig
	var createdAtStr, updatedAtStr string
	var whitelistJSON string

	err := scanner.Scan(
		&config.ID,
		&config.MaxAttemptsPerIP,
		&config.MaxAttemptsPerUser,
		&config.WindowSizeMinutes,
		&config.LockoutDurationMinutes,
		&config.BackoffMultiplier,
		&config.MaxLockoutHours,
		&whitelistJSON,
		&config.EnableProgressiveLockout,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan rate limit config: %w", err)
	}

	// Parse whitelist IPs from JSON
	err = json.Unmarshal([]byte(whitelistJSON), &config.WhitelistIPs)
	if err != nil {
		// If JSON parsing fails, use empty slice
		config.WhitelistIPs = []string{}
	}

	// Parse timestamps
	createdAt, err := r.parseTimestamp(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	config.CreatedAt = createdAt

	updatedAt, err := r.parseTimestamp(updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}
	config.UpdatedAt = updatedAt

	return &config, nil
}

// parseTimestamp parses timestamps in either RFC3339 or SQLite datetime format
func (r *RateLimitConfigRepository) parseTimestamp(timestampStr string) (time.Time, error) {
	// Try RFC3339 format first (what we store for new records)
	if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
		return t, nil
	}

	// Try SQLite datetime format (used in migrations with datetime('now'))
	if t, err := time.Parse("2006-01-02 15:04:05", timestampStr); err == nil {
		return t.UTC(), nil
	}

	// Try additional SQLite formats just in case
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02T15:04:05.000",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestampStr); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestampStr)
}

// Type alias for domain model to avoid import cycles
type RateLimitConfig = domain.RateLimitConfig
