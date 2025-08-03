package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"shien-system/internal/domain"
)

// AuditLogRepository implements domain.AuditLogRepository
type AuditLogRepository struct {
	db *Database
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *Database) *AuditLogRepository {
	return &AuditLogRepository{
		db: db,
	}
}

// Create creates a new audit log entry
// Note: Audit logs are immutable - no Update or Delete methods
func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, actor_id, action, target, at, ip, details) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query,
		log.ID,
		log.ActorID,
		log.Action,
		log.Target,
		log.At.Format(time.RFC3339),
		log.IP,
		log.Details,
	)

	if err != nil {
		return &domain.RepositoryError{Op: "create audit log", Err: err}
	}

	return nil
}

// GetByID retrieves an audit log by ID
func (r *AuditLogRepository) GetByID(ctx context.Context, id domain.ID) (*domain.AuditLog, error) {
	query := `
		SELECT id, actor_id, action, target, at, ip, details
		FROM audit_logs 
		WHERE id = ?`

	executor := r.getExecutor(ctx)
	row := executor.QueryRowContext(ctx, query, id)

	return r.scanAuditLog(row)
}

// GetByActorID retrieves audit logs by actor ID with pagination
func (r *AuditLogRepository) GetByActorID(ctx context.Context, actorID domain.ID, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, actor_id, action, target, at, ip, details
		FROM audit_logs 
		WHERE actor_id = ?
		ORDER BY at DESC
		LIMIT ? OFFSET ?`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, actorID, limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get logs by actor", Err: err}
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return logs, nil
}

// GetByAction retrieves audit logs by action with pagination
func (r *AuditLogRepository) GetByAction(ctx context.Context, action string, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, actor_id, action, target, at, ip, details
		FROM audit_logs 
		WHERE action = ?
		ORDER BY at DESC
		LIMIT ? OFFSET ?`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, action, limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get logs by action", Err: err}
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return logs, nil
}

// GetByTarget retrieves audit logs by target with pagination
func (r *AuditLogRepository) GetByTarget(ctx context.Context, target string, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, actor_id, action, target, at, ip, details
		FROM audit_logs 
		WHERE target = ?
		ORDER BY at DESC
		LIMIT ? OFFSET ?`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, target, limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get logs by target", Err: err}
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return logs, nil
}

// GetByTimeRange retrieves audit logs within a time range with pagination
func (r *AuditLogRepository) GetByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, actor_id, action, target, at, ip, details
		FROM audit_logs 
		WHERE at >= ? AND at <= ?
		ORDER BY at DESC
		LIMIT ? OFFSET ?`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query,
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
		limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get logs by time range", Err: err}
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return logs, nil
}

// Search retrieves audit logs matching the query criteria with pagination
func (r *AuditLogRepository) Search(ctx context.Context, query domain.AuditLogQuery, limit, offset int) ([]*domain.AuditLog, error) {
	// Build dynamic query based on provided criteria
	whereClause, args := r.buildSearchQuery(query)

	sqlQuery := fmt.Sprintf(`
		SELECT id, actor_id, action, target, at, ip, details
		FROM audit_logs 
		%s
		ORDER BY at DESC
		LIMIT ? OFFSET ?`, whereClause)

	// Add limit and offset to args
	args = append(args, limit, offset)

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "search audit logs", Err: err}
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return logs, nil
}

// List retrieves audit logs with pagination
func (r *AuditLogRepository) List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, actor_id, action, target, at, ip, details
		FROM audit_logs 
		ORDER BY at DESC
		LIMIT ? OFFSET ?`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "list audit logs", Err: err}
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return logs, nil
}

// Count returns the total number of audit logs
func (r *AuditLogRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM audit_logs`

	executor := r.getExecutor(ctx)
	var count int
	err := executor.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, &domain.RepositoryError{Op: "count audit logs", Err: err}
	}

	return count, nil
}

// buildSearchQuery builds WHERE clause and arguments for search
func (r *AuditLogRepository) buildSearchQuery(query domain.AuditLogQuery) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if query.ActorID != nil {
		conditions = append(conditions, "actor_id = ?")
		args = append(args, *query.ActorID)
	}

	if query.Action != nil {
		conditions = append(conditions, "action = ?")
		args = append(args, *query.Action)
	}

	if query.Target != nil {
		conditions = append(conditions, "target = ?")
		args = append(args, *query.Target)
	}

	if query.StartTime != nil {
		conditions = append(conditions, "at >= ?")
		args = append(args, query.StartTime.Format(time.RFC3339))
	}

	if query.EndTime != nil {
		conditions = append(conditions, "at <= ?")
		args = append(args, query.EndTime.Format(time.RFC3339))
	}

	if query.IP != nil {
		conditions = append(conditions, "ip = ?")
		args = append(args, *query.IP)
	}

	if len(conditions) == 0 {
		return "", args
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")
	return whereClause, args
}

// getExecutor returns either a transaction or the database connection
func (r *AuditLogRepository) getExecutor(ctx context.Context) executor {
	if tx := ctx.Value("tx"); tx != nil {
		return tx.(*sql.Tx)
	}
	return r.db.DB()
}

// scanAuditLog scans an audit log from a database row
func (r *AuditLogRepository) scanAuditLog(row scanner) (*domain.AuditLog, error) {
	var log domain.AuditLog
	var atStr string

	err := row.Scan(
		&log.ID,
		&log.ActorID,
		&log.Action,
		&log.Target,
		&atStr,
		&log.IP,
		&log.Details,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, &domain.RepositoryError{Op: "scan audit log", Err: err}
	}

	// Parse timestamp
	log.At, err = time.Parse(time.RFC3339, atStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse at timestamp", Err: err}
	}

	return &log, nil
}
