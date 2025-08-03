package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"shien-system/internal/domain"
)

// LoginAttemptRepository implements domain.LoginAttemptRepository
type LoginAttemptRepository struct {
	db *Database
}

// NewLoginAttemptRepository creates a new LoginAttemptRepository instance
func NewLoginAttemptRepository(db *Database) *LoginAttemptRepository {
	return &LoginAttemptRepository{
		db: db,
	}
}

// Create creates a new login attempt record
func (r *LoginAttemptRepository) Create(ctx context.Context, attempt *LoginAttempt) error {
	if attempt.ID == "" {
		attempt.ID = uuid.New().String()
	}

	query := `
		INSERT INTO login_attempts (id, ip_address, username, success, attempted_at, user_agent)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.getExecutor(ctx).ExecContext(ctx, query,
		attempt.ID,
		attempt.IPAddress,
		attempt.Username,
		attempt.Success,
		attempt.AttemptedAt.Format(time.RFC3339),
		attempt.UserAgent,
	)

	if err != nil {
		return fmt.Errorf("failed to create login attempt: %w", err)
	}

	return nil
}

// GetByIPAddress retrieves login attempts for a specific IP address since a given time
func (r *LoginAttemptRepository) GetByIPAddress(ctx context.Context, ipAddress string, since time.Time) ([]*LoginAttempt, error) {
	query := `
		SELECT id, ip_address, username, success, attempted_at, user_agent
		FROM login_attempts
		WHERE ip_address = ? AND attempted_at >= ?
		ORDER BY attempted_at DESC
	`

	rows, err := r.getExecutor(ctx).QueryContext(ctx, query, ipAddress, since.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to get login attempts by IP: %w", err)
	}
	defer rows.Close()

	var attempts []*LoginAttempt
	for rows.Next() {
		attempt, err := r.scanLoginAttempt(rows)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}

	return attempts, nil
}

// GetByUsername retrieves login attempts for a specific username since a given time
func (r *LoginAttemptRepository) GetByUsername(ctx context.Context, username string, since time.Time) ([]*LoginAttempt, error) {
	query := `
		SELECT id, ip_address, username, success, attempted_at, user_agent
		FROM login_attempts
		WHERE username = ? AND attempted_at >= ?
		ORDER BY attempted_at DESC
	`

	rows, err := r.getExecutor(ctx).QueryContext(ctx, query, username, since.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to get login attempts by username: %w", err)
	}
	defer rows.Close()

	var attempts []*LoginAttempt
	for rows.Next() {
		attempt, err := r.scanLoginAttempt(rows)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}

	return attempts, nil
}

// GetFailedAttempts retrieves failed login attempts for a specific IP and/or username
func (r *LoginAttemptRepository) GetFailedAttempts(ctx context.Context, ipAddress, username string, since time.Time) ([]*LoginAttempt, error) {
	var query string
	var args []interface{}

	if ipAddress != "" && username != "" {
		query = `
			SELECT id, ip_address, username, success, attempted_at, user_agent
			FROM login_attempts
			WHERE (ip_address = ? OR username = ?) AND success = FALSE AND attempted_at >= ?
			ORDER BY attempted_at DESC
		`
		args = []interface{}{ipAddress, username, since.Format(time.RFC3339)}
	} else if ipAddress != "" {
		query = `
			SELECT id, ip_address, username, success, attempted_at, user_agent
			FROM login_attempts
			WHERE ip_address = ? AND success = FALSE AND attempted_at >= ?
			ORDER BY attempted_at DESC
		`
		args = []interface{}{ipAddress, since.Format(time.RFC3339)}
	} else if username != "" {
		query = `
			SELECT id, ip_address, username, success, attempted_at, user_agent
			FROM login_attempts
			WHERE username = ? AND success = FALSE AND attempted_at >= ?
			ORDER BY attempted_at DESC
		`
		args = []interface{}{username, since.Format(time.RFC3339)}
	} else {
		return nil, fmt.Errorf("either ipAddress or username must be provided")
	}

	rows, err := r.getExecutor(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed login attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*LoginAttempt
	for rows.Next() {
		attempt, err := r.scanLoginAttempt(rows)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}

	return attempts, nil
}

// CountRecentFailures counts recent failed login attempts
func (r *LoginAttemptRepository) CountRecentFailures(ctx context.Context, ipAddress, username string, since time.Time) (int, error) {
	var query string
	var args []interface{}

	if ipAddress != "" && username != "" {
		query = `
			SELECT COUNT(*)
			FROM login_attempts
			WHERE (ip_address = ? OR username = ?) AND success = FALSE AND attempted_at >= ?
		`
		args = []interface{}{ipAddress, username, since.Format(time.RFC3339)}
	} else if ipAddress != "" {
		query = `
			SELECT COUNT(*)
			FROM login_attempts
			WHERE ip_address = ? AND success = FALSE AND attempted_at >= ?
		`
		args = []interface{}{ipAddress, since.Format(time.RFC3339)}
	} else if username != "" {
		query = `
			SELECT COUNT(*)
			FROM login_attempts
			WHERE username = ? AND success = FALSE AND attempted_at >= ?
		`
		args = []interface{}{username, since.Format(time.RFC3339)}
	} else {
		return 0, fmt.Errorf("either ipAddress or username must be provided")
	}

	var count int
	err := r.getExecutor(ctx).QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count recent failures: %w", err)
	}

	return count, nil
}

// DeleteOldAttempts deletes login attempts older than the specified time
func (r *LoginAttemptRepository) DeleteOldAttempts(ctx context.Context, before time.Time) error {
	query := `DELETE FROM login_attempts WHERE attempted_at < ?`

	result, err := r.getExecutor(ctx).ExecContext(ctx, query, before.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to delete old login attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Log how many records were cleaned up (for monitoring)
	_ = rowsAffected

	return nil
}

// Helper methods

func (r *LoginAttemptRepository) getExecutor(ctx context.Context) executor {
	if tx, ok := ctx.Value("tx").(*sql.Tx); ok {
		return tx
	}
	return r.db.DB()
}

func (r *LoginAttemptRepository) scanLoginAttempt(scanner scanner) (*LoginAttempt, error) {
	var attempt LoginAttempt
	var attemptedAtStr string
	var userAgent sql.NullString

	err := scanner.Scan(
		&attempt.ID,
		&attempt.IPAddress,
		&attempt.Username,
		&attempt.Success,
		&attemptedAtStr,
		&userAgent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan login attempt: %w", err)
	}

	// Parse timestamp
	attemptedAt, err := r.parseTimestamp(attemptedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse attempted_at: %w", err)
	}
	attempt.AttemptedAt = attemptedAt

	if userAgent.Valid {
		attempt.UserAgent = userAgent.String
	}

	return &attempt, nil
}

// parseTimestamp parses timestamps in either RFC3339 or SQLite datetime format
func (r *LoginAttemptRepository) parseTimestamp(timestampStr string) (time.Time, error) {
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
type LoginAttempt = domain.LoginAttempt
