package db

import (
	"context"
	"database/sql"
	"time"

	"shien-system/internal/domain"
)

// StaffAssignmentRepository implements domain.StaffAssignmentRepository
type StaffAssignmentRepository struct {
	db *Database
}

// NewStaffAssignmentRepository creates a new staff assignment repository
func NewStaffAssignmentRepository(db *Database) *StaffAssignmentRepository {
	return &StaffAssignmentRepository{
		db: db,
	}
}

// Create creates a new staff assignment
func (r *StaffAssignmentRepository) Create(ctx context.Context, assignment *domain.StaffAssignment) error {
	query := `
		INSERT INTO staff_assignments (id, recipient_id, staff_id, role, assigned_at, unassigned_at) 
		VALUES (?, ?, ?, ?, ?, ?)`

	var unassignedAtStr *string
	if assignment.UnassignedAt != nil {
		str := assignment.UnassignedAt.Format(time.RFC3339)
		unassignedAtStr = &str
	}

	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query,
		assignment.ID,
		assignment.RecipientID,
		assignment.StaffID,
		assignment.Role,
		assignment.AssignedAt.Format(time.RFC3339),
		unassignedAtStr,
	)

	if err != nil {
		return &domain.RepositoryError{Op: "create staff assignment", Err: err}
	}

	return nil
}

// GetByID retrieves a staff assignment by ID
func (r *StaffAssignmentRepository) GetByID(ctx context.Context, id domain.ID) (*domain.StaffAssignment, error) {
	query := `
		SELECT id, recipient_id, staff_id, role, assigned_at, unassigned_at
		FROM staff_assignments 
		WHERE id = ?`

	executor := r.getExecutor(ctx)
	row := executor.QueryRowContext(ctx, query, id)

	return r.scanStaffAssignment(row)
}

// Update updates an existing staff assignment
func (r *StaffAssignmentRepository) Update(ctx context.Context, assignment *domain.StaffAssignment) error {
	query := `
		UPDATE staff_assignments 
		SET recipient_id = ?, staff_id = ?, role = ?, assigned_at = ?, unassigned_at = ?
		WHERE id = ?`

	var unassignedAtStr *string
	if assignment.UnassignedAt != nil {
		str := assignment.UnassignedAt.Format(time.RFC3339)
		unassignedAtStr = &str
	}

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query,
		assignment.RecipientID,
		assignment.StaffID,
		assignment.Role,
		assignment.AssignedAt.Format(time.RFC3339),
		unassignedAtStr,
		assignment.ID,
	)

	if err != nil {
		return &domain.RepositoryError{Op: "update staff assignment", Err: err}
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

// Delete deletes a staff assignment by ID
func (r *StaffAssignmentRepository) Delete(ctx context.Context, id domain.ID) error {
	query := `DELETE FROM staff_assignments WHERE id = ?`

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return &domain.RepositoryError{Op: "delete staff assignment", Err: err}
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

// GetByRecipientID retrieves all staff assignments for a recipient
func (r *StaffAssignmentRepository) GetByRecipientID(ctx context.Context, recipientID domain.ID) ([]*domain.StaffAssignment, error) {
	query := `
		SELECT id, recipient_id, staff_id, role, assigned_at, unassigned_at
		FROM staff_assignments 
		WHERE recipient_id = ?
		ORDER BY assigned_at DESC`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, recipientID)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get assignments by recipient", Err: err}
	}
	defer rows.Close()

	var assignments []*domain.StaffAssignment
	for rows.Next() {
		assignment, err := r.scanStaffAssignment(rows)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return assignments, nil
}

// GetByStaffID retrieves all staff assignments for a staff member
func (r *StaffAssignmentRepository) GetByStaffID(ctx context.Context, staffID domain.ID) ([]*domain.StaffAssignment, error) {
	query := `
		SELECT id, recipient_id, staff_id, role, assigned_at, unassigned_at
		FROM staff_assignments 
		WHERE staff_id = ?
		ORDER BY assigned_at DESC`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, staffID)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get assignments by staff", Err: err}
	}
	defer rows.Close()

	var assignments []*domain.StaffAssignment
	for rows.Next() {
		assignment, err := r.scanStaffAssignment(rows)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return assignments, nil
}

// GetActiveByRecipientID retrieves active staff assignments for a recipient
func (r *StaffAssignmentRepository) GetActiveByRecipientID(ctx context.Context, recipientID domain.ID) ([]*domain.StaffAssignment, error) {
	query := `
		SELECT id, recipient_id, staff_id, role, assigned_at, unassigned_at
		FROM staff_assignments 
		WHERE recipient_id = ? AND unassigned_at IS NULL
		ORDER BY assigned_at DESC`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, recipientID)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get active assignments by recipient", Err: err}
	}
	defer rows.Close()

	var assignments []*domain.StaffAssignment
	for rows.Next() {
		assignment, err := r.scanStaffAssignment(rows)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return assignments, nil
}

// GetActiveByStaffID retrieves active staff assignments for a staff member
func (r *StaffAssignmentRepository) GetActiveByStaffID(ctx context.Context, staffID domain.ID) ([]*domain.StaffAssignment, error) {
	query := `
		SELECT id, recipient_id, staff_id, role, assigned_at, unassigned_at
		FROM staff_assignments 
		WHERE staff_id = ? AND unassigned_at IS NULL
		ORDER BY assigned_at DESC`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, staffID)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get active assignments by staff", Err: err}
	}
	defer rows.Close()

	var assignments []*domain.StaffAssignment
	for rows.Next() {
		assignment, err := r.scanStaffAssignment(rows)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return assignments, nil
}

// UnassignAll sets unassigned_at for all active assignments for a recipient
func (r *StaffAssignmentRepository) UnassignAll(ctx context.Context, recipientID domain.ID, unassignedAt time.Time) error {
	query := `
		UPDATE staff_assignments 
		SET unassigned_at = ?
		WHERE recipient_id = ? AND unassigned_at IS NULL`

	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query, unassignedAt.Format(time.RFC3339), recipientID)
	if err != nil {
		return &domain.RepositoryError{Op: "unassign all assignments", Err: err}
	}

	return nil
}

// List retrieves staff assignments with pagination
func (r *StaffAssignmentRepository) List(ctx context.Context, limit, offset int) ([]*domain.StaffAssignment, error) {
	query := `
		SELECT id, recipient_id, staff_id, role, assigned_at, unassigned_at
		FROM staff_assignments 
		ORDER BY assigned_at DESC
		LIMIT ? OFFSET ?`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "list assignments", Err: err}
	}
	defer rows.Close()

	var assignments []*domain.StaffAssignment
	for rows.Next() {
		assignment, err := r.scanStaffAssignment(rows)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return assignments, nil
}

// Count returns the total number of staff assignments
func (r *StaffAssignmentRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM staff_assignments`

	executor := r.getExecutor(ctx)
	var count int
	err := executor.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, &domain.RepositoryError{Op: "count assignments", Err: err}
	}

	return count, nil
}

// getExecutor returns either a transaction or the database connection
func (r *StaffAssignmentRepository) getExecutor(ctx context.Context) executor {
	if tx := ctx.Value("tx"); tx != nil {
		return tx.(*sql.Tx)
	}
	return r.db.DB()
}

// scanStaffAssignment scans a staff assignment from a database row
func (r *StaffAssignmentRepository) scanStaffAssignment(row scanner) (*domain.StaffAssignment, error) {
	var assignment domain.StaffAssignment
	var assignedAtStr string
	var unassignedAtStr *string

	err := row.Scan(
		&assignment.ID,
		&assignment.RecipientID,
		&assignment.StaffID,
		&assignment.Role,
		&assignedAtStr,
		&unassignedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, &domain.RepositoryError{Op: "scan staff assignment", Err: err}
	}

	// Parse assigned_at timestamp
	assignment.AssignedAt, err = time.Parse(time.RFC3339, assignedAtStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse assigned_at", Err: err}
	}

	// Parse optional unassigned_at timestamp
	if unassignedAtStr != nil {
		unassignedAt, err := time.Parse(time.RFC3339, *unassignedAtStr)
		if err != nil {
			return nil, &domain.RepositoryError{Op: "parse unassigned_at", Err: err}
		}
		assignment.UnassignedAt = &unassignedAt
	}

	return &assignment, nil
}
