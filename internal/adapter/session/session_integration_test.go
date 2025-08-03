package session

import (
	"context"
	"testing"
	"time"

	"shien-system/internal/adapter/crypto"
	"shien-system/internal/config"
	"shien-system/internal/domain"
	"shien-system/internal/usecase"
)

func TestSecureSessionManager_BasicFunctionality(t *testing.T) {
	// テスト用の設定
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

	randomGen := crypto.NewSecureRandomGenerator()
	manager := NewSecureSessionManager(nil, randomGen, config, nil).(*SecureSessionManager)
	defer manager.Stop()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	userID := "test-user-001"
	userRole := domain.RoleStaff

	// セッション作成のテスト
	session, err := manager.CreateSession(ctx, userID, userRole)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}
	if session.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
	}
	if session.CSRFToken == "" {
		t.Error("CSRF token should not be empty")
	}

	// セッション検証のテスト
	validatedSession, err := manager.ValidateSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to validate session: %v", err)
	}

	if validatedSession.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, validatedSession.ID)
	}

	// CSRFトークン検証のテスト
	err = manager.ValidateCSRFToken(ctx, session.ID, session.CSRFToken)
	if err != nil {
		t.Errorf("Failed to validate CSRF token: %v", err)
	}

	// 無効なCSRFトークンのテスト
	err = manager.ValidateCSRFToken(ctx, session.ID, "invalid-token")
	if err != usecase.ErrInvalidCSRFToken {
		t.Errorf("Expected ErrInvalidCSRFToken, got %v", err)
	}

	// セッションリフレッシュのテスト
	refreshedSession, err := manager.RefreshSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to refresh session: %v", err)
	}

	if refreshedSession.ID == session.ID {
		t.Error("Refreshed session should have a new session ID")
	}
	if refreshedSession.CSRFToken == session.CSRFToken {
		t.Error("Refreshed session should have a new CSRF token")
	}

	// 古いセッションが無効化されていることを確認
	_, err = manager.ValidateSession(ctx, session.ID)
	if err != usecase.ErrInvalidSession {
		t.Errorf("Expected ErrInvalidSession for old session, got %v", err)
	}

	// セッション削除のテスト
	err = manager.DeleteSession(ctx, refreshedSession.ID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// 削除されたセッションが無効であることを確認
	_, err = manager.ValidateSession(ctx, refreshedSession.ID)
	if err != usecase.ErrInvalidSession {
		t.Errorf("Expected ErrInvalidSession for deleted session, got %v", err)
	}
}

func TestSecureSessionManager_SecurityValidation(t *testing.T) {
	// IP検証ありの設定
	config := &config.SessionConfig{
		StorageType:                "memory",
		MaxSessionsPerUser:         3,
		ForceSingleSession:         false,
		RequireIPValidation:        true,
		RequireUserAgentValidation: true,
		CleanupIntervalMinutes:     60,
		SessionIDLength:            32,
		CSRFTokenLength:            32,
		PersistenceEnabled:         false,
	}

	randomGen := crypto.NewSecureRandomGenerator()
	manager := NewSecureSessionManager(nil, randomGen, config, nil).(*SecureSessionManager)
	defer manager.Stop()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	// セッション作成
	session, err := manager.CreateSession(ctx, "test-user-001", domain.RoleStaff)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 異なるIPアドレスでの検証
	ctxDifferentIP := context.Background()
	ctxDifferentIP = context.WithValue(ctxDifferentIP, usecase.ContextKeyClientIP, "192.168.1.200")
	ctxDifferentIP = context.WithValue(ctxDifferentIP, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	_, err = manager.ValidateSession(ctxDifferentIP, session.ID)
	if err != usecase.ErrIPAddressMismatch {
		t.Errorf("Expected ErrIPAddressMismatch, got %v", err)
	}

	// 異なるUser-Agentでの検証
	ctxDifferentUA := context.Background()
	ctxDifferentUA = context.WithValue(ctxDifferentUA, usecase.ContextKeyClientIP, "192.168.1.100")
	ctxDifferentUA = context.WithValue(ctxDifferentUA, usecase.ContextKeyUserAgent, "DifferentAgent/2.0")

	_, err = manager.ValidateSession(ctxDifferentUA, session.ID)
	if err != usecase.ErrUserAgentMismatch {
		t.Errorf("Expected ErrUserAgentMismatch, got %v", err)
	}
}

func TestSecureSessionManager_ExpirationHandling(t *testing.T) {
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

	randomGen := crypto.NewSecureRandomGenerator()
	manager := NewSecureSessionManager(nil, randomGen, config, nil).(*SecureSessionManager)
	defer manager.Stop()

	ctx := context.Background()
	ctx = context.WithValue(ctx, usecase.ContextKeyClientIP, "192.168.1.100")
	ctx = context.WithValue(ctx, usecase.ContextKeyUserAgent, "TestAgent/1.0")

	// セッション作成
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

func TestCryptoRandomGenerator(t *testing.T) {
	gen := crypto.NewSecureRandomGenerator()

	// セッションID生成のテスト
	sessionID, err := gen.GenerateSessionID(32)
	if err != nil {
		t.Fatalf("Failed to generate session ID: %v", err)
	}
	if sessionID == "" {
		t.Error("Session ID should not be empty")
	}

	// CSRFトークン生成のテスト
	csrfToken, err := gen.GenerateCSRFToken(32)
	if err != nil {
		t.Fatalf("Failed to generate CSRF token: %v", err)
	}
	if csrfToken == "" {
		t.Error("CSRF token should not be empty")
	}

	// セッションID検証のテスト
	err = gen.ValidateSessionID(sessionID)
	if err != nil {
		t.Errorf("Failed to validate session ID: %v", err)
	}

	// CSRFトークン検証のテスト
	err = gen.ValidateCSRFToken(csrfToken)
	if err != nil {
		t.Errorf("Failed to validate CSRF token: %v", err)
	}

	// 定数時間比較のテスト
	if !gen.ConstantTimeCompare("test", "test") {
		t.Error("Constant time compare should return true for equal strings")
	}
	if gen.ConstantTimeCompare("test", "different") {
		t.Error("Constant time compare should return false for different strings")
	}
}
