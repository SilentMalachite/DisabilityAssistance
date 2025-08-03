package session

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"shien-system/internal/adapter/crypto"
	"shien-system/internal/adapter/db"
	"shien-system/internal/config"
	"shien-system/internal/domain"
	"shien-system/internal/usecase"
)

// SecureSessionManager implements SessionManager interface with database persistence
// and enhanced security features including session fixation protection, CSRF tokens,
// and configurable security policies
type SecureSessionManager struct {
	repository      *db.SessionRepository
	memoryManager   *MemorySessionManager // fallback for memory-only mode
	randomGenerator *crypto.SecureRandomGenerator
	config          *config.SessionConfig
	mutex           sync.RWMutex
	cleanupTicker   *time.Ticker
	cleanupStop     chan bool
	auditLogger     usecase.AuditUseCase
}

// NewSecureSessionManager creates a new secure session manager
func NewSecureSessionManager(
	repository *db.SessionRepository,
	randomGenerator *crypto.SecureRandomGenerator,
	sessionConfig *config.SessionConfig,
	auditLogger usecase.AuditUseCase,
) usecase.SessionManager {

	// メモリ管理用のフォールバック
	sessionTimeout, _ := time.ParseDuration("24h") // デフォルト値
	memoryManager := NewMemorySessionManager(sessionTimeout).(*MemorySessionManager)

	manager := &SecureSessionManager{
		repository:      repository,
		memoryManager:   memoryManager,
		randomGenerator: randomGenerator,
		config:          sessionConfig,
		mutex:           sync.RWMutex{},
		cleanupStop:     make(chan bool),
		auditLogger:     auditLogger,
	}

	// 定期クリーンアップの開始
	if sessionConfig.CleanupIntervalMinutes > 0 {
		manager.startCleanupRoutine()
	}

	return manager
}

// CreateSession creates a new session with enhanced security features
func (m *SecureSessionManager) CreateSession(ctx context.Context, userID domain.ID, userRole domain.StaffRole) (*usecase.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// ユーザーの既存セッション数をチェック
	if err := m.checkSessionLimits(ctx, userID); err != nil {
		return nil, err
	}

	// セッションIDとCSRFトークンの生成
	sessionID, err := m.randomGenerator.GenerateSessionID(m.config.SessionIDLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	csrfToken, err := m.randomGenerator.GenerateCSRFToken(m.config.CSRFTokenLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	// コンテキストから情報を取得
	clientIP := getClientIPFromContext(ctx)
	userAgent := getUserAgentFromContextSecure(ctx)

	now := time.Now()
	sessionTimeout, _ := time.ParseDuration("24h") // TODO: 設定から取得

	session := &usecase.Session{
		ID:             sessionID,
		UserID:         userID,
		UserRole:       userRole,
		CreatedAt:      now,
		ExpiresAt:      now.Add(sessionTimeout),
		LastAccessedAt: now,
		ClientIP:       clientIP,
		UserAgent:      userAgent,
		CSRFToken:      csrfToken,
		IsActive:       true,
	}

	// データベースに保存（設定による）
	if m.config.PersistenceEnabled && m.config.StorageType == "database" {
		if err := m.repository.CreateSession(ctx, session); err != nil {
			return nil, fmt.Errorf("failed to persist session: %w", err)
		}
	} else {
		// メモリに保存
		m.memoryManager.sessions[sessionID] = session
	}

	// 監査ログ記録
	if m.auditLogger != nil {
		logReq := usecase.LogActionRequest{
			ActorID: userID,
			Action:  "SESSION_CREATE",
			Target:  "session:" + sessionID,
			IP:      clientIP,
			Details: fmt.Sprintf("User-Agent: %s", userAgent),
		}
		// ログ記録エラーは致命的ではない
		_ = m.auditLogger.LogAction(ctx, logReq)
	}

	return session, nil
}

// ValidateSession validates and retrieves session information with security checks
func (m *SecureSessionManager) ValidateSession(ctx context.Context, sessionID string) (*usecase.Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// セッションIDフォーマットの検証
	if err := m.randomGenerator.ValidateSessionID(sessionID); err != nil {
		return nil, usecase.ErrInvalidSession
	}

	var session *usecase.Session
	var err error

	// セッションの取得
	if m.config.PersistenceEnabled && m.config.StorageType == "database" {
		session, err = m.repository.GetSession(ctx, sessionID)
	} else {
		// メモリから取得
		if storedSession, exists := m.memoryManager.sessions[sessionID]; exists {
			session = storedSession
		} else {
			err = usecase.ErrInvalidSession
		}
	}

	if err != nil {
		return nil, err
	}

	// 有効期限チェック
	if time.Now().After(session.ExpiresAt) {
		// 期限切れセッションを無効化
		_ = m.InvalidateSession(ctx, sessionID, "expired")
		return nil, usecase.ErrSessionExpired
	}

	// セキュリティチェック
	if err := m.performSecurityChecks(ctx, session); err != nil {
		// セキュリティ違反でセッションを無効化
		_ = m.InvalidateSession(ctx, sessionID, fmt.Sprintf("security_violation: %s", err.Error()))
		return nil, err
	}

	// 最終アクセス時刻の更新
	now := time.Now()
	session.LastAccessedAt = now

	if m.config.PersistenceEnabled && m.config.StorageType == "database" {
		if err := m.repository.UpdateSessionLastAccessed(ctx, sessionID, now); err != nil {
			// 更新失敗は致命的ではないが、ログに記録
		}
	}

	// セッションのコピーを返す（元データの変更を防ぐ）
	return &usecase.Session{
		ID:             session.ID,
		UserID:         session.UserID,
		UserRole:       session.UserRole,
		CreatedAt:      session.CreatedAt,
		ExpiresAt:      session.ExpiresAt,
		LastAccessedAt: session.LastAccessedAt,
		ClientIP:       session.ClientIP,
		UserAgent:      session.UserAgent,
		CSRFToken:      session.CSRFToken,
		IsActive:       session.IsActive,
	}, nil
}

// DeleteSession removes a session (logout)
func (m *SecureSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.InvalidateSession(ctx, sessionID, "logout")
}

// RefreshSession extends session expiration with security re-validation
func (m *SecureSessionManager) RefreshSession(ctx context.Context, sessionID string) (*usecase.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// まず現在のセッションを検証
	session, err := m.ValidateSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// セッション固定攻撃の対策：新しいセッションIDを生成
	newSessionID, err := m.randomGenerator.GenerateSessionID(m.config.SessionIDLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new session ID: %w", err)
	}

	// 新しいCSRFトークンも生成
	newCSRFToken, err := m.randomGenerator.GenerateCSRFToken(m.config.CSRFTokenLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new CSRF token: %w", err)
	}

	// 古いセッションを無効化
	_ = m.InvalidateSession(ctx, sessionID, "refreshed")

	// 新しいセッションを作成
	now := time.Now()
	sessionTimeout, _ := time.ParseDuration("24h") // TODO: 設定から取得

	newSession := &usecase.Session{
		ID:             newSessionID,
		UserID:         session.UserID,
		UserRole:       session.UserRole,
		CreatedAt:      now,
		ExpiresAt:      now.Add(sessionTimeout),
		LastAccessedAt: now,
		ClientIP:       session.ClientIP,
		UserAgent:      session.UserAgent,
		CSRFToken:      newCSRFToken,
		IsActive:       true,
	}

	// 新しいセッションを保存
	if m.config.PersistenceEnabled && m.config.StorageType == "database" {
		if err := m.repository.CreateSession(ctx, newSession); err != nil {
			return nil, fmt.Errorf("failed to persist refreshed session: %w", err)
		}
	} else {
		m.memoryManager.sessions[newSessionID] = newSession
	}

	// 監査ログ記録
	if m.auditLogger != nil {
		logReq := usecase.LogActionRequest{
			ActorID: session.UserID,
			Action:  "SESSION_REFRESH",
			Target:  "session:" + newSessionID,
			IP:      session.ClientIP,
			Details: fmt.Sprintf("Old session: %s", sessionID),
		}
		_ = m.auditLogger.LogAction(ctx, logReq)
	}

	return newSession, nil
}

// CleanupExpiredSessions removes expired sessions
func (m *SecureSessionManager) CleanupExpiredSessions(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.config.PersistenceEnabled && m.config.StorageType == "database" {
		return m.repository.CleanupExpiredSessions(ctx)
	} else {
		// メモリからクリーンアップ
		return m.memoryManager.CleanupExpiredSessions(ctx)
	}
}

// InvalidateSession marks a session as invalid
func (m *SecureSessionManager) InvalidateSession(ctx context.Context, sessionID string, reason string) error {
	if m.config.PersistenceEnabled && m.config.StorageType == "database" {
		return m.repository.InvalidateSession(ctx, sessionID, reason)
	} else {
		// メモリから削除
		delete(m.memoryManager.sessions, sessionID)
		return nil
	}
}

// ValidateCSRFToken validates CSRF token for a session
func (m *SecureSessionManager) ValidateCSRFToken(ctx context.Context, sessionID, csrfToken string) error {
	session, err := m.ValidateSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if !m.randomGenerator.ConstantTimeCompare(session.CSRFToken, csrfToken) {
		return usecase.ErrInvalidCSRFToken
	}

	return nil
}

// checkSessionLimits verifies session limits for a user
func (m *SecureSessionManager) checkSessionLimits(ctx context.Context, userID domain.ID) error {
	if m.config.ForceSingleSession {
		// 単一セッション強制の場合、既存セッションを無効化
		if m.config.PersistenceEnabled && m.config.StorageType == "database" {
			sessions, err := m.repository.GetSessionsByUserID(ctx, userID)
			if err != nil {
				return err
			}
			for _, session := range sessions {
				_ = m.InvalidateSession(ctx, session.ID, "new_session_force_single")
			}
		}
		return nil
	}

	// 最大セッション数のチェック
	if m.config.PersistenceEnabled && m.config.StorageType == "database" {
		sessions, err := m.repository.GetSessionsByUserID(ctx, userID)
		if err != nil {
			return err
		}

		if len(sessions) >= m.config.MaxSessionsPerUser {
			// 最も古いセッションを無効化
			oldestSession := sessions[len(sessions)-1]
			_ = m.InvalidateSession(ctx, oldestSession.ID, "session_limit_exceeded")
		}
	}

	return nil
}

// performSecurityChecks performs various security validations
func (m *SecureSessionManager) performSecurityChecks(ctx context.Context, session *usecase.Session) error {
	// IP アドレスの検証
	if m.config.RequireIPValidation {
		currentIP := getClientIPFromContext(ctx)
		if currentIP != session.ClientIP {
			return usecase.ErrIPAddressMismatch
		}
	}

	// User-Agent の検証
	if m.config.RequireUserAgentValidation {
		currentUserAgent := getUserAgentFromContextSecure(ctx)
		if currentUserAgent != session.UserAgent {
			return usecase.ErrUserAgentMismatch
		}
	}

	return nil
}

// startCleanupRoutine starts the background cleanup routine
func (m *SecureSessionManager) startCleanupRoutine() {
	interval := time.Duration(m.config.CleanupIntervalMinutes) * time.Minute
	m.cleanupTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-m.cleanupTicker.C:
				ctx := context.Background()
				_ = m.CleanupExpiredSessions(ctx)
			case <-m.cleanupStop:
				m.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// Stop stops the session manager and cleanup routines
func (m *SecureSessionManager) Stop() {
	if m.cleanupTicker != nil {
		close(m.cleanupStop)
	}
}

// getUserAgentFromContextSecure extracts User-Agent from context for secure manager
func getUserAgentFromContextSecure(ctx context.Context) string {
	if userAgent := ctx.Value(usecase.ContextKeyUserAgent); userAgent != nil {
		if ua, ok := userAgent.(string); ok {
			return ua
		}
	}
	return "unknown"
}

// SanitizeUserAgent sanitizes User-Agent string for security
func SanitizeUserAgent(userAgent string) string {
	// 基本的なサニタイゼーション
	sanitized := strings.TrimSpace(userAgent)
	if len(sanitized) > 500 {
		sanitized = sanitized[:500]
	}
	return sanitized
}

// SanitizeIPAddress sanitizes IP address for security
func SanitizeIPAddress(ip string) string {
	// 基本的なサニタイゼーション
	sanitized := strings.TrimSpace(ip)
	if len(sanitized) > 45 { // IPv6の最大長
		sanitized = sanitized[:45]
	}
	return sanitized
}
