package session

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"shien-system/internal/adapter/crypto"
	"shien-system/internal/adapter/db"
	"shien-system/internal/config"
	"shien-system/internal/domain"
	"shien-system/internal/usecase"

	_ "github.com/mattn/go-sqlite3"
)

func TestSecureSessionManager_CreateSession(t *testing.T) {
	manager, cleanup := setupSecureSessionManager(t)
	defer cleanup()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	userID := "test-user-001"
	userRole := domain.RoleStaff

	session, err := manager.CreateSession(ctx, userID, userRole)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// セッションの基本検証
	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}
	if session.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
	}
	if session.UserRole != userRole {
		t.Errorf("Expected user role %s, got %s", userRole, session.UserRole)
	}
	if session.ClientIP != "192.168.1.100" {
		t.Errorf("Expected client IP 192.168.1.100, got %s", session.ClientIP)
	}
	if session.UserAgent != "TestAgent/1.0" {
		t.Errorf("Expected user agent TestAgent/1.0, got %s", session.UserAgent)
	}
	if session.CSRFToken == "" {
		t.Error("CSRF token should not be empty")
	}
	if !session.IsActive {
		t.Error("Session should be active")
	}
}

func TestSecureSessionManager_ValidateSession(t *testing.T) {
	manager, cleanup := setupSecureSessionManager(t)
	defer cleanup()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	// セッションを作成
	session, err := manager.CreateSession(ctx, "test-user-001", domain.RoleStaff)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// セッションを検証
	validatedSession, err := manager.ValidateSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to validate session: %v", err)
	}

	if validatedSession.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, validatedSession.ID)
	}
	if validatedSession.UserID != session.UserID {
		t.Errorf("Expected user ID %s, got %s", session.UserID, validatedSession.UserID)
	}
}

func TestSecureSessionManager_ValidateSession_InvalidSessionID(t *testing.T) {
	manager, cleanup := setupSecureSessionManager(t)
	defer cleanup()

	ctx := context.Background()

	// 無効なセッションIDで検証
	_, err := manager.ValidateSession(ctx, "invalid-session-id")
	if err != usecase.ErrInvalidSession {
		t.Errorf("Expected ErrInvalidSession, got %v", err)
	}
}

func TestSecureSessionManager_ValidateSession_ExpiredSession(t *testing.T) {
	manager, cleanup := setupSecureSessionManager(t)
	defer cleanup()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	// 短い有効期限でセッションを作成
	config := &config.SessionConfig{
		StorageType:                "memory",
		MaxSessionsPerUser:         3,
		ForceSingleSession:         false,
		RequireIPValidation:        false,
		RequireUserAgentValidation: false,
		CleanupIntervalMinutes:     1,
		SessionIDLength:            32,
		CSRFTokenLength:            32,
		PersistenceEnabled:         false,
	}

	randomGen := crypto.NewSecureRandomGenerator()

	manager = NewSecureSessionManager(nil, randomGen, config, nil).(*SecureSessionManager)

	session, err := manager.CreateSession(ctx, "test-user-001", domain.RoleStaff)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// セッションの有効期限を過去に設定
	session.ExpiresAt = time.Now().Add(-1 * time.Hour)
	manager.memoryManager.sessions[session.ID] = session

	// 期限切れセッションの検証
	_, err = manager.ValidateSession(ctx, session.ID)
	if err != usecase.ErrSessionExpired {
		t.Errorf("Expected ErrSessionExpired, got %v", err)
	}
}

func TestSecureSessionManager_IPValidation(t *testing.T) {
	manager, cleanup := setupSecureSessionManagerWithIPValidation(t)
	defer cleanup()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	// セッションを作成
	session, err := manager.CreateSession(ctx, "test-user-001", domain.RoleStaff)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 異なるIPアドレスで検証を試行
	ctxDifferentIP := context.Background()
	ctxDifferentIP = context.WithValue(ctxDifferentIP, usecase.ContextKeyClientIP, "192.168.1.200")
	ctxDifferentIP = context.WithValue(ctxDifferentIP, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	_, err = manager.ValidateSession(ctxDifferentIP, session.ID)
	if err != usecase.ErrIPAddressMismatch {
		t.Errorf("Expected ErrIPAddressMismatch, got %v", err)
	}
}

func TestSecureSessionManager_UserAgentValidation(t *testing.T) {
	manager, cleanup := setupSecureSessionManagerWithUserAgentValidation(t)
	defer cleanup()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	// セッションを作成
	session, err := manager.CreateSession(ctx, "test-user-001", domain.RoleStaff)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 異なるUser-Agentで検証を試行
	ctxDifferentUA := context.Background()
	ctxDifferentUA = context.WithValue(ctxDifferentUA, usecase.ContextKeyClientIP, "192.168.1.100")
	ctxDifferentUA = context.WithValue(ctxDifferentUA, usecase.ContextKeyUserAgent, "DifferentAgent/2.0")

	_, err = manager.ValidateSession(ctxDifferentUA, session.ID)
	if err != usecase.ErrUserAgentMismatch {
		t.Errorf("Expected ErrUserAgentMismatch, got %v", err)
	}
}

func TestSecureSessionManager_RefreshSession(t *testing.T) {
	manager, cleanup := setupSecureSessionManager(t)
	defer cleanup()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	// セッションを作成
	originalSession, err := manager.CreateSession(ctx, "test-user-001", domain.RoleStaff)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// セッションをリフレッシュ
	refreshedSession, err := manager.RefreshSession(ctx, originalSession.ID)
	if err != nil {
		t.Fatalf("Failed to refresh session: %v", err)
	}

	// 新しいセッションIDが生成されていることを確認
	if refreshedSession.ID == originalSession.ID {
		t.Error("Refreshed session should have a new session ID")
	}

	// 新しいCSRFトークンが生成されていることを確認
	if refreshedSession.CSRFToken == originalSession.CSRFToken {
		t.Error("Refreshed session should have a new CSRF token")
	}

	// ユーザー情報は保持されていることを確認
	if refreshedSession.UserID != originalSession.UserID {
		t.Errorf("Expected user ID %s, got %s", originalSession.UserID, refreshedSession.UserID)
	}

	// 古いセッションは無効化されていることを確認
	_, err = manager.ValidateSession(ctx, originalSession.ID)
	if err != usecase.ErrInvalidSession {
		t.Errorf("Expected ErrInvalidSession for old session, got %v", err)
	}
}

func TestSecureSessionManager_CSRFTokenValidation(t *testing.T) {
	manager, cleanup := setupSecureSessionManager(t)
	defer cleanup()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	// セッションを作成
	session, err := manager.CreateSession(ctx, "test-user-001", domain.RoleStaff)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 正しいCSRFトークンの検証
	err = manager.ValidateCSRFToken(ctx, session.ID, session.CSRFToken)
	if err != nil {
		t.Errorf("Failed to validate correct CSRF token: %v", err)
	}

	// 間違ったCSRFトークンの検証
	err = manager.ValidateCSRFToken(ctx, session.ID, "invalid-csrf-token")
	if err != usecase.ErrInvalidCSRFToken {
		t.Errorf("Expected ErrInvalidCSRFToken, got %v", err)
	}
}

func TestSecureSessionManager_SessionLimit(t *testing.T) {
	manager, cleanup := setupSecureSessionManagerWithLimit(t, 2)
	defer cleanup()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	userID := "test-user-001"

	// 最大数までセッションを作成
	session1, err := manager.CreateSession(ctx, userID, domain.RoleStaff)
	if err != nil {
		t.Fatalf("Failed to create first session: %v", err)
	}

	session2, err := manager.CreateSession(ctx, userID, domain.RoleStaff)
	if err != nil {
		t.Fatalf("Failed to create second session: %v", err)
	}

	// 3つ目のセッションを作成（最初のセッションが無効化されるはず）
	session3, err := manager.CreateSession(ctx, userID, domain.RoleStaff)
	if err != nil {
		t.Fatalf("Failed to create third session: %v", err)
	}

	// 最初のセッションが無効化されていることを確認
	_, err = manager.ValidateSession(ctx, session1.ID)
	if err != usecase.ErrInvalidSession {
		t.Errorf("Expected first session to be invalidated, got %v", err)
	}

	// 2番目と3番目のセッションは有効であることを確認
	_, err = manager.ValidateSession(ctx, session2.ID)
	if err != nil {
		t.Errorf("Second session should be valid: %v", err)
	}

	_, err = manager.ValidateSession(ctx, session3.ID)
	if err != nil {
		t.Errorf("Third session should be valid: %v", err)
	}
}

// setupSecureSessionManager creates a test session manager
func setupSecureSessionManager(t *testing.T) (*SecureSessionManager, func()) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// テーブル作成
	createSessionTables(t, sqlDB)

	cipher, err := crypto.NewFieldCipher()
	if err != nil {
		t.Fatalf("Failed to create cipher: %v", err)
	}

	repository := db.NewSessionRepository(sqlDB, cipher)
	randomGen := crypto.NewSecureRandomGenerator()

	config := &config.SessionConfig{
		StorageType:                "memory",
		MaxSessionsPerUser:         3,
		ForceSingleSession:         false,
		RequireIPValidation:        false,
		RequireUserAgentValidation: false,
		CleanupIntervalMinutes:     60,
		SessionIDLength:            32,
		CSRFTokenLength:            32,
		PersistenceEnabled:         false,
	}

	manager := NewSecureSessionManager(repository, randomGen, config, nil).(*SecureSessionManager)

	cleanup := func() {
		manager.Stop()
		sqlDB.Close()
	}

	return manager, cleanup
}

// setupSecureSessionManagerWithIPValidation creates a test session manager with IP validation
func setupSecureSessionManagerWithIPValidation(t *testing.T) (*SecureSessionManager, func()) {
	manager, cleanup := setupSecureSessionManager(t)
	manager.config.RequireIPValidation = true
	return manager, cleanup
}

// setupSecureSessionManagerWithUserAgentValidation creates a test session manager with User-Agent validation
func setupSecureSessionManagerWithUserAgentValidation(t *testing.T) (*SecureSessionManager, func()) {
	manager, cleanup := setupSecureSessionManager(t)
	manager.config.RequireUserAgentValidation = true
	return manager, cleanup
}

// setupSecureSessionManagerWithLimit creates a test session manager with session limit
func setupSecureSessionManagerWithLimit(t *testing.T, limit int) (*SecureSessionManager, func()) {
	manager, cleanup := setupSecureSessionManager(t)
	manager.config.MaxSessionsPerUser = limit
	return manager, cleanup
}

// createSessionTables creates session tables for testing
func createSessionTables(t *testing.T, db *sql.DB) {
	// ここでは簡単なテーブル作成を行う
	// 実際のマイグレーションは別途実装される
	_, err := db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			user_role_cipher BLOB NOT NULL,
			client_ip_cipher BLOB,
			user_agent_cipher BLOB,
			csrf_token TEXT NOT NULL,
			created_at TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			last_accessed_at TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create sessions table: %v", err)
	}
}
