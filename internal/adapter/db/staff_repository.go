package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"shien-system/internal/domain"
)

// StaffRepository implements domain.StaffRepository
type StaffRepository struct {
	db *Database
}

// NewStaffRepository creates a new staff repository
func NewStaffRepository(db *Database) *StaffRepository {
	return &StaffRepository{
		db: db,
	}
}

// Create creates a new staff member
func (r *StaffRepository) Create(ctx context.Context, staff *domain.Staff) error {
	query := `
		INSERT INTO staff (id, name, role, password_hash, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`

	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query,
		staff.ID,
		staff.Name,
		string(staff.Role),
		staff.PasswordHash,
		staff.CreatedAt.Format(time.RFC3339),
		staff.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return &domain.RepositoryError{Op: "create staff", Err: err}
	}

	return nil
}

// GetByID retrieves a staff member by ID
func (r *StaffRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Staff, error) {
	query := `
		SELECT id, name, role, password_hash, created_at, updated_at
		FROM staff 
		WHERE id = ?`

	executor := r.getExecutor(ctx)
	row := executor.QueryRowContext(ctx, query, id)

	return r.scanStaff(row)
}

// Update updates an existing staff member
func (r *StaffRepository) Update(ctx context.Context, staff *domain.Staff) error {
	query := `
		UPDATE staff 
		SET name = ?, role = ?, password_hash = ?, updated_at = ?
		WHERE id = ?`

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query,
		staff.Name,
		string(staff.Role),
		staff.PasswordHash,
		staff.UpdatedAt.Format(time.RFC3339),
		staff.ID,
	)

	if err != nil {
		return &domain.RepositoryError{Op: "update staff", Err: err}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &domain.RepositoryError{Op: "check rows affected", Err: err}
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete deletes a staff member by ID
func (r *StaffRepository) Delete(ctx context.Context, id domain.ID) error {
	query := `DELETE FROM staff WHERE id = ?`

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return &domain.RepositoryError{Op: "delete staff", Err: err}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &domain.RepositoryError{Op: "check rows affected", Err: err}
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// List retrieves staff members with pagination
func (r *StaffRepository) List(ctx context.Context, limit, offset int) ([]*domain.Staff, error) {
	query := `
		SELECT id, name, role, password_hash, created_at, updated_at
		FROM staff 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "list staff", Err: err}
	}
	defer rows.Close()

	var staffMembers []*domain.Staff
	for rows.Next() {
		staff, err := r.scanStaff(rows)
		if err != nil {
			return nil, err
		}
		staffMembers = append(staffMembers, staff)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return staffMembers, nil
}

// GetByRole retrieves staff members by role
func (r *StaffRepository) GetByRole(ctx context.Context, role domain.StaffRole) ([]*domain.Staff, error) {
	query := `
		SELECT id, name, role, password_hash, created_at, updated_at
		FROM staff 
		WHERE role = ?
		ORDER BY created_at DESC`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, string(role))
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get staff by role", Err: err}
	}
	defer rows.Close()

	var staffMembers []*domain.Staff
	for rows.Next() {
		staff, err := r.scanStaff(rows)
		if err != nil {
			return nil, err
		}
		staffMembers = append(staffMembers, staff)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return staffMembers, nil
}

// GetByExactName retrieves a single staff member by exact name match
func (r *StaffRepository) GetByExactName(ctx context.Context, name string) (*domain.Staff, error) {
	query := `
		SELECT id, name, role, password_hash, created_at, updated_at
		FROM staff 
		WHERE name = ?`

	executor := r.getExecutor(ctx)
	row := executor.QueryRowContext(ctx, query, name)

	return r.scanStaff(row)
}

// GetByName retrieves staff members by name (partial match)
func (r *StaffRepository) GetByName(ctx context.Context, name string) ([]*domain.Staff, error) {
	query := `
		SELECT id, name, role, password_hash, created_at, updated_at
		FROM staff 
		WHERE name LIKE ?
		ORDER BY name`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, "%"+name+"%")
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get staff by name", Err: err}
	}
	defer rows.Close()

	var staffMembers []*domain.Staff
	for rows.Next() {
		staff, err := r.scanStaff(rows)
		if err != nil {
			return nil, err
		}
		staffMembers = append(staffMembers, staff)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return staffMembers, nil
}

// Count returns the total number of staff members
func (r *StaffRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM staff`

	executor := r.getExecutor(ctx)
	var count int
	err := executor.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, &domain.RepositoryError{Op: "count staff", Err: err}
	}

	return count, nil
}

// getExecutor returns either a transaction or the database connection
func (r *StaffRepository) getExecutor(ctx context.Context) executor {
	if tx := ctx.Value("tx"); tx != nil {
		return tx.(*sql.Tx)
	}
	return r.db.DB()
}

// scanStaff scans a staff member from a database row
func (r *StaffRepository) scanStaff(row scanner) (*domain.Staff, error) {
	var staff domain.Staff
	var roleStr, createdAtStr, updatedAtStr string

	err := row.Scan(
		&staff.ID,
		&staff.Name,
		&roleStr,
		&staff.PasswordHash,
		&createdAtStr,
		&updatedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, &domain.RepositoryError{Op: "scan staff", Err: err}
	}

	// Parse role
	staff.Role = domain.StaffRole(roleStr)

	// Parse timestamps - support both RFC3339 and SQLite datetime format
	staff.CreatedAt, err = r.parseTimestamp(createdAtStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse created_at", Err: err}
	}

	staff.UpdatedAt, err = r.parseTimestamp(updatedAtStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse updated_at", Err: err}
	}

	return &staff, nil
}

// parseTimestamp parses timestamps in either RFC3339 or SQLite datetime format
func (r *StaffRepository) parseTimestamp(timestampStr string) (time.Time, error) {
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

	return time.Time{}, &domain.RepositoryError{
		Op:  "parse timestamp",
		Err: fmt.Errorf("unsupported timestamp format: %s", timestampStr),
	}
}
