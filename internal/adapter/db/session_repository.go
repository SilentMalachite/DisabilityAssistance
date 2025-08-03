package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"shien-system/internal/adapter/crypto"
	"shien-system/internal/domain"
	"shien-system/internal/usecase"
)

// SessionRepository handles session data persistence with encryption
type SessionRepository struct {
	db     *sql.DB
	cipher *crypto.FieldCipher
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB, cipher *crypto.FieldCipher) *SessionRepository {
	return &SessionRepository{
		db:     db,
		cipher: cipher,
	}
}

// CreateSession stores a new session in the database
func (r *SessionRepository) CreateSession(ctx context.Context, session *usecase.Session) error {
	// 暗号化するフィールド
	userRoleCipher, err := r.cipher.Encrypt(string(session.UserRole))
	if err != nil {
		return fmt.Errorf("failed to encrypt user role: %w", err)
	}

	clientIPCipher, err := r.cipher.Encrypt(session.ClientIP)
	if err != nil {
		return fmt.Errorf("failed to encrypt client IP: %w", err)
	}

	userAgentCipher, err := r.cipher.Encrypt(session.UserAgent)
	if err != nil {
		return fmt.Errorf("failed to encrypt user agent: %w", err)
	}

	query := `
		INSERT INTO sessions (
			id, user_id, user_role_cipher, client_ip_cipher, user_agent_cipher,
			csrf_token, created_at, expires_at, last_accessed_at, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		userRoleCipher,
		clientIPCipher,
		userAgentCipher,
		session.CSRFToken,
		session.CreatedAt.Format(time.RFC3339),
		session.ExpiresAt.Format(time.RFC3339),
		session.LastAccessedAt.Format(time.RFC3339),
		1, // is_active = true
	)

	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	// セッション履歴に記録
	if err := r.logSessionHistory(ctx, session.ID, session.UserID, "CREATE", session.ClientIP, session.UserAgent, ""); err != nil {
		// ログ記録の失敗は致命的エラーではない
		// TODO: ログに記録
	}

	return nil
}

// GetSession retrieves a session by ID
func (r *SessionRepository) GetSession(ctx context.Context, sessionID string) (*usecase.Session, error) {
	query := `
		SELECT id, user_id, user_role_cipher, client_ip_cipher, user_agent_cipher,
		       csrf_token, created_at, expires_at, last_accessed_at, is_active,
		       invalidation_reason, invalidated_at
		FROM sessions 
		WHERE id = ? AND is_active = 1`

	row := r.db.QueryRowContext(ctx, query, sessionID)

	var session usecase.Session
	var userRoleCipher, clientIPCipher, userAgentCipher []byte
	var createdAtStr, expiresAtStr, lastAccessedAtStr string
	var isActive int
	var invalidationReason sql.NullString
	var invalidatedAtStr sql.NullString

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&userRoleCipher,
		&clientIPCipher,
		&userAgentCipher,
		&session.CSRFToken,
		&createdAtStr,
		&expiresAtStr,
		&lastAccessedAtStr,
		&isActive,
		&invalidationReason,
		&invalidatedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, usecase.ErrInvalidSession
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 復号化
	userRoleStr, err := r.cipher.Decrypt(userRoleCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt user role: %w", err)
	}
	session.UserRole = domain.StaffRole(userRoleStr)

	session.ClientIP, err = r.cipher.Decrypt(clientIPCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt client IP: %w", err)
	}

	session.UserAgent, err = r.cipher.Decrypt(userAgentCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt user agent: %w", err)
	}

	// 時刻のパース
	session.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	session.ExpiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_at: %w", err)
	}

	session.LastAccessedAt, err = time.Parse(time.RFC3339, lastAccessedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_accessed_at: %w", err)
	}

	session.IsActive = isActive == 1

	if invalidationReason.Valid {
		session.InvalidationReason = invalidationReason.String
	}

	if invalidatedAtStr.Valid {
		invalidatedAt, err := time.Parse(time.RFC3339, invalidatedAtStr.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse invalidated_at: %w", err)
		}
		session.InvalidatedAt = &invalidatedAt
	}

	return &session, nil
}

// UpdateSessionLastAccessed updates the last accessed time of a session
func (r *SessionRepository) UpdateSessionLastAccessed(ctx context.Context, sessionID string, lastAccessedAt time.Time) error {
	query := `UPDATE sessions SET last_accessed_at = ? WHERE id = ? AND is_active = 1`

	result, err := r.db.ExecContext(ctx, query, lastAccessedAt.Format(time.RFC3339), sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session last accessed time: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return usecase.ErrInvalidSession
	}

	return nil
}

// InvalidateSession marks a session as invalid
func (r *SessionRepository) InvalidateSession(ctx context.Context, sessionID string, reason string) error {
	now := time.Now()
	query := `
		UPDATE sessions 
		SET is_active = 0, invalidation_reason = ?, invalidated_at = ?
		WHERE id = ? AND is_active = 1`

	result, err := r.db.ExecContext(ctx, query, reason, now.Format(time.RFC3339), sessionID)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return usecase.ErrInvalidSession
	}

	// セッション履歴に記録
	if err := r.logSessionHistory(ctx, sessionID, "", "INVALIDATE", "", "", reason); err != nil {
		// ログ記録の失敗は致命的エラーではない
	}

	return nil
}

// GetSessionsByUserID retrieves all active sessions for a user
func (r *SessionRepository) GetSessionsByUserID(ctx context.Context, userID domain.ID) ([]*usecase.Session, error) {
	query := `
		SELECT id, user_id, user_role_cipher, client_ip_cipher, user_agent_cipher,
		       csrf_token, created_at, expires_at, last_accessed_at, is_active
		FROM sessions 
		WHERE user_id = ? AND is_active = 1 
		ORDER BY last_accessed_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions by user ID: %w", err)
	}
	defer rows.Close()

	var sessions []*usecase.Session

	for rows.Next() {
		var session usecase.Session
		var userRoleCipher, clientIPCipher, userAgentCipher []byte
		var createdAtStr, expiresAtStr, lastAccessedAtStr string
		var isActive int

		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&userRoleCipher,
			&clientIPCipher,
			&userAgentCipher,
			&session.CSRFToken,
			&createdAtStr,
			&expiresAtStr,
			&lastAccessedAtStr,
			&isActive,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}

		// 復号化
		userRoleStr, err := r.cipher.Decrypt(userRoleCipher)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt user role: %w", err)
		}
		session.UserRole = domain.StaffRole(userRoleStr)

		session.ClientIP, err = r.cipher.Decrypt(clientIPCipher)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt client IP: %w", err)
		}

		session.UserAgent, err = r.cipher.Decrypt(userAgentCipher)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt user agent: %w", err)
		}

		// 時刻のパース
		session.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		session.ExpiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expires_at: %w", err)
		}

		session.LastAccessedAt, err = time.Parse(time.RFC3339, lastAccessedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last_accessed_at: %w", err)
		}

		session.IsActive = isActive == 1

		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (r *SessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	now := time.Now()

	// 期限切れセッション数を取得
	countQuery := `SELECT COUNT(*) FROM sessions WHERE expires_at < ? AND is_active = 1`
	var expiredCount int
	err := r.db.QueryRowContext(ctx, countQuery, now.Format(time.RFC3339)).Scan(&expiredCount)
	if err != nil {
		return fmt.Errorf("failed to count expired sessions: %w", err)
	}

	if expiredCount == 0 {
		return nil // 期限切れセッションなし
	}

	// 期限切れセッションを無効化
	query := `
		UPDATE sessions 
		SET is_active = 0, invalidation_reason = 'expired', invalidated_at = ?
		WHERE expires_at < ? AND is_active = 1`

	result, err := r.db.ExecContext(ctx, query, now.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// TODO: ログに記録
	_ = rowsAffected

	return nil
}

// GetSessionConfig retrieves session configuration from database
func (r *SessionRepository) GetSessionConfig(ctx context.Context) (*SessionConfig, error) {
	query := `
		SELECT max_sessions_per_user, session_timeout_hours, cleanup_interval_hours,
		       force_single_session, require_ip_validation, require_user_agent_validation
		FROM session_config 
		WHERE id = 1`

	row := r.db.QueryRowContext(ctx, query)

	var config SessionConfig
	var forceSingleSession, requireIPValidation, requireUserAgentValidation int

	err := row.Scan(
		&config.MaxSessionsPerUser,
		&config.SessionTimeoutHours,
		&config.CleanupIntervalHours,
		&forceSingleSession,
		&requireIPValidation,
		&requireUserAgentValidation,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// デフォルト設定を返す
			return &SessionConfig{
				MaxSessionsPerUser:         3,
				SessionTimeoutHours:        24,
				CleanupIntervalHours:       1,
				ForceSingleSession:         false,
				RequireIPValidation:        true,
				RequireUserAgentValidation: true,
			}, nil
		}
		return nil, fmt.Errorf("failed to get session config: %w", err)
	}

	config.ForceSingleSession = forceSingleSession == 1
	config.RequireIPValidation = requireIPValidation == 1
	config.RequireUserAgentValidation = requireUserAgentValidation == 1

	return &config, nil
}

// SessionConfig holds database session configuration
type SessionConfig struct {
	MaxSessionsPerUser         int
	SessionTimeoutHours        int
	CleanupIntervalHours       int
	ForceSingleSession         bool
	RequireIPValidation        bool
	RequireUserAgentValidation bool
}

// logSessionHistory records session activity for audit purposes
func (r *SessionRepository) logSessionHistory(ctx context.Context, sessionID, userID, action, clientIP, userAgent, details string) error {
	// 暗号化するフィールド
	clientIPCipher, err := r.cipher.Encrypt(clientIP)
	if err != nil {
		return fmt.Errorf("failed to encrypt client IP for history: %w", err)
	}

	userAgentCipher, err := r.cipher.Encrypt(userAgent)
	if err != nil {
		return fmt.Errorf("failed to encrypt user agent for history: %w", err)
	}

	// 詳細情報をJSONとして保存
	detailsJSON := ""
	if details != "" {
		detailsMap := map[string]string{"reason": details}
		detailsBytes, err := json.Marshal(detailsMap)
		if err == nil {
			detailsJSON = string(detailsBytes)
		}
	}

	query := `
		INSERT INTO session_history (
			id, session_id, user_id, action, client_ip_cipher, user_agent_cipher, details
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

	historyID := fmt.Sprintf("history_%d", time.Now().UnixNano())

	_, err = r.db.ExecContext(ctx, query,
		historyID,
		sessionID,
		userID,
		action,
		clientIPCipher,
		userAgentCipher,
		detailsJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to insert session history: %w", err)
	}

	return nil
}
