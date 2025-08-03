package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"shien-system/internal/domain"
)

// AccountLockoutRepository implements domain.AccountLockoutRepository
type AccountLockoutRepository struct {
	db *Database
}

// NewAccountLockoutRepository creates a new AccountLockoutRepository instance
func NewAccountLockoutRepository(db *Database) *AccountLockoutRepository {
	return &AccountLockoutRepository{
		db: db,
	}
}

// Create creates a new account lockout record
func (r *AccountLockoutRepository) Create(ctx context.Context, lockout *AccountLockout) error {
	if lockout.ID == "" {
		lockout.ID = uuid.New().String()
	}

	query := `
		INSERT INTO account_lockouts (
			id, username, ip_address, lockout_type, locked_at, 
			unlocked_at, reason, failure_count, duration
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var unlockedAt *string
	if lockout.UnlockedAt != nil {
		unlockedAtStr := lockout.UnlockedAt.Format(time.RFC3339)
		unlockedAt = &unlockedAtStr
	}

	_, err := r.getExecutor(ctx).ExecContext(ctx, query,
		lockout.ID,
		lockout.Username,
		lockout.IPAddress,
		string(lockout.LockoutType),
		lockout.LockedAt.Format(time.RFC3339),
		unlockedAt,
		lockout.Reason,
		lockout.FailureCount,
		lockout.Duration,
	)

	if err != nil {
		return fmt.Errorf("failed to create account lockout: %w", err)
	}

	return nil
}

// GetByID retrieves an account lockout by ID
func (r *AccountLockoutRepository) GetByID(ctx context.Context, id domain.ID) (*AccountLockout, error) {
	query := `
		SELECT id, username, ip_address, lockout_type, locked_at, 
		       unlocked_at, reason, failure_count, duration
		FROM account_lockouts
		WHERE id = ?
	`

	row := r.getExecutor(ctx).QueryRowContext(ctx, query, id)
	lockout, err := r.scanAccountLockout(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get account lockout by ID: %w", err)
	}

	return lockout, nil
}

// GetActiveByUsername retrieves active lockout for a username
func (r *AccountLockoutRepository) GetActiveByUsername(ctx context.Context, username string) (*AccountLockout, error) {
	query := `
		SELECT id, username, ip_address, lockout_type, locked_at, 
		       unlocked_at, reason, failure_count, duration
		FROM account_lockouts
		WHERE username = ? AND unlocked_at IS NULL
		ORDER BY locked_at DESC
		LIMIT 1
	`

	row := r.getExecutor(ctx).QueryRowContext(ctx, query, username)
	lockout, err := r.scanAccountLockout(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active lockout found
		}
		return nil, fmt.Errorf("failed to get active lockout by username: %w", err)
	}

	// Check if lockout has expired
	if r.isLockoutExpired(lockout) {
		// Auto-unlock expired lockout
		if err := r.Unlock(ctx, lockout.ID, time.Now()); err != nil {
			return nil, fmt.Errorf("failed to auto-unlock expired lockout: %w", err)
		}
		return nil, nil
	}

	return lockout, nil
}

// GetActiveByIPAddress retrieves active lockout for an IP address
func (r *AccountLockoutRepository) GetActiveByIPAddress(ctx context.Context, ipAddress string) (*AccountLockout, error) {
	query := `
		SELECT id, username, ip_address, lockout_type, locked_at, 
		       unlocked_at, reason, failure_count, duration
		FROM account_lockouts
		WHERE ip_address = ? AND unlocked_at IS NULL
		ORDER BY locked_at DESC
		LIMIT 1
	`

	row := r.getExecutor(ctx).QueryRowContext(ctx, query, ipAddress)
	lockout, err := r.scanAccountLockout(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active lockout found
		}
		return nil, fmt.Errorf("failed to get active lockout by IP: %w", err)
	}

	// Check if lockout has expired
	if r.isLockoutExpired(lockout) {
		// Auto-unlock expired lockout
		if err := r.Unlock(ctx, lockout.ID, time.Now()); err != nil {
			return nil, fmt.Errorf("failed to auto-unlock expired lockout: %w", err)
		}
		return nil, nil
	}

	return lockout, nil
}

// GetActiveLockouts retrieves all active lockouts
func (r *AccountLockoutRepository) GetActiveLockouts(ctx context.Context) ([]*AccountLockout, error) {
	query := `
		SELECT id, username, ip_address, lockout_type, locked_at, 
		       unlocked_at, reason, failure_count, duration
		FROM account_lockouts
		WHERE unlocked_at IS NULL
		ORDER BY locked_at DESC
	`

	rows, err := r.getExecutor(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active lockouts: %w", err)
	}
	defer rows.Close()

	var lockouts []*AccountLockout
	for rows.Next() {
		lockout, err := r.scanAccountLockout(rows)
		if err != nil {
			return nil, err
		}

		// Filter out expired lockouts
		if !r.isLockoutExpired(lockout) {
			lockouts = append(lockouts, lockout)
		}
	}

	return lockouts, nil
}

// Unlock unlocks an account by setting the unlocked_at timestamp
func (r *AccountLockoutRepository) Unlock(ctx context.Context, id domain.ID, unlockedAt time.Time) error {
	query := `UPDATE account_lockouts SET unlocked_at = ? WHERE id = ?`

	result, err := r.getExecutor(ctx).ExecContext(ctx, query, unlockedAt.Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("failed to unlock account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UnlockByUsername unlocks all active lockouts for a username
func (r *AccountLockoutRepository) UnlockByUsername(ctx context.Context, username string, unlockedAt time.Time) error {
	query := `UPDATE account_lockouts SET unlocked_at = ? WHERE username = ? AND unlocked_at IS NULL`

	_, err := r.getExecutor(ctx).ExecContext(ctx, query, unlockedAt.Format(time.RFC3339), username)
	if err != nil {
		return fmt.Errorf("failed to unlock account by username: %w", err)
	}

	return nil
}

// UnlockByIPAddress unlocks all active lockouts for an IP address
func (r *AccountLockoutRepository) UnlockByIPAddress(ctx context.Context, ipAddress string, unlockedAt time.Time) error {
	query := `UPDATE account_lockouts SET unlocked_at = ? WHERE ip_address = ? AND unlocked_at IS NULL`

	_, err := r.getExecutor(ctx).ExecContext(ctx, query, unlockedAt.Format(time.RFC3339), ipAddress)
	if err != nil {
		return fmt.Errorf("failed to unlock account by IP: %w", err)
	}

	return nil
}

// CleanupExpiredLockouts removes expired lockouts
func (r *AccountLockoutRepository) CleanupExpiredLockouts(ctx context.Context) error {
	// Auto-unlock expired lockouts first
	query := `
		UPDATE account_lockouts 
		SET unlocked_at = datetime('now') 
		WHERE unlocked_at IS NULL 
		AND datetime(locked_at, '+' || duration || ' seconds') < datetime('now')
	`

	_, err := r.getExecutor(ctx).ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired lockouts: %w", err)
	}

	return nil
}

// List retrieves account lockouts with pagination
func (r *AccountLockoutRepository) List(ctx context.Context, limit, offset int) ([]*AccountLockout, error) {
	query := `
		SELECT id, username, ip_address, lockout_type, locked_at, 
		       unlocked_at, reason, failure_count, duration
		FROM account_lockouts
		ORDER BY locked_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.getExecutor(ctx).QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list account lockouts: %w", err)
	}
	defer rows.Close()

	var lockouts []*AccountLockout
	for rows.Next() {
		lockout, err := r.scanAccountLockout(rows)
		if err != nil {
			return nil, err
		}
		lockouts = append(lockouts, lockout)
	}

	return lockouts, nil
}

// Count returns the total number of account lockouts
func (r *AccountLockoutRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM account_lockouts`

	var count int
	err := r.getExecutor(ctx).QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count account lockouts: %w", err)
	}

	return count, nil
}

// Helper methods

func (r *AccountLockoutRepository) getExecutor(ctx context.Context) executor {
	if tx, ok := ctx.Value("tx").(*sql.Tx); ok {
		return tx
	}
	return r.db.DB()
}

func (r *AccountLockoutRepository) scanAccountLockout(scanner scanner) (*AccountLockout, error) {
	var lockout AccountLockout
	var lockedAtStr string
	var unlockedAtStr sql.NullString
	var lockoutTypeStr string

	err := scanner.Scan(
		&lockout.ID,
		&lockout.Username,
		&lockout.IPAddress,
		&lockoutTypeStr,
		&lockedAtStr,
		&unlockedAtStr,
		&lockout.Reason,
		&lockout.FailureCount,
		&lockout.Duration,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err // Return sql.ErrNoRows as-is
		}
		return nil, fmt.Errorf("failed to scan account lockout: %w", err)
	}

	// Parse lockout type
	lockout.LockoutType = domain.LockoutType(lockoutTypeStr)

	// Parse timestamps
	lockedAt, err := r.parseTimestamp(lockedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse locked_at: %w", err)
	}
	lockout.LockedAt = lockedAt

	if unlockedAtStr.Valid {
		unlockedAt, err := r.parseTimestamp(unlockedAtStr.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse unlocked_at: %w", err)
		}
		lockout.UnlockedAt = &unlockedAt
	}

	return &lockout, nil
}

// parseTimestamp parses timestamps in either RFC3339 or SQLite datetime format
func (r *AccountLockoutRepository) parseTimestamp(timestampStr string) (time.Time, error) {
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

// isLockoutExpired checks if a lockout has expired based on its duration
func (r *AccountLockoutRepository) isLockoutExpired(lockout *AccountLockout) bool {
	if lockout.UnlockedAt != nil {
		return true // Already unlocked
	}

	if lockout.Duration <= 0 {
		return false // Permanent lockout
	}

	expirationTime := lockout.LockedAt.Add(time.Duration(lockout.Duration) * time.Second)
	return time.Now().After(expirationTime)
}

// Type alias for domain model to avoid import cycles
type AccountLockout = domain.AccountLockout
