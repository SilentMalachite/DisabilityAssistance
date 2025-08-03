package session

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"shien-system/internal/domain"
	"shien-system/internal/usecase"
)

// MemorySessionManager implements SessionManager interface using in-memory storage
type MemorySessionManager struct {
	sessions      map[string]*usecase.Session
	sessionExpiry time.Duration
	mutex         sync.RWMutex
}

// NewMemorySessionManager creates a new memory-based session manager
func NewMemorySessionManager(sessionExpiry time.Duration) usecase.SessionManager {
	return &MemorySessionManager{
		sessions:      make(map[string]*usecase.Session),
		sessionExpiry: sessionExpiry,
		mutex:         sync.RWMutex{},
	}
}

// CreateSession creates a new session for a user
func (m *MemorySessionManager) CreateSession(ctx context.Context, userID domain.ID, userRole domain.StaffRole) (*usecase.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	session := &usecase.Session{
		ID:             uuid.New().String(),
		UserID:         userID,
		UserRole:       userRole,
		CreatedAt:      now,
		ExpiresAt:      now.Add(m.sessionExpiry),
		LastAccessedAt: now,
		ClientIP:       getClientIPFromContext(ctx),
		UserAgent:      getUserAgentFromContext(ctx),
		CSRFToken:      uuid.New().String(), // 簡単なCSRFトークン生成
		IsActive:       true,
	}

	m.sessions[session.ID] = session
	return session, nil
}

// ValidateSession validates and retrieves session information
func (m *MemorySessionManager) ValidateSession(ctx context.Context, sessionID string) (*usecase.Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, usecase.ErrInvalidSession
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Remove expired session
		m.mutex.RUnlock()
		m.mutex.Lock()
		delete(m.sessions, sessionID)
		m.mutex.Unlock()
		m.mutex.RLock()
		return nil, usecase.ErrSessionExpired
	}

	// 最終アクセス時刻を更新
	session.LastAccessedAt = time.Now()

	// Return a copy to avoid external modifications
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

// DeleteSession removes a session
func (m *MemorySessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.sessions, sessionID)
	return nil
}

// RefreshSession extends session expiration
func (m *MemorySessionManager) RefreshSession(ctx context.Context, sessionID string) (*usecase.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, usecase.ErrInvalidSession
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		delete(m.sessions, sessionID)
		return nil, usecase.ErrSessionExpired
	}

	// Extend expiration
	newExpiresAt := time.Now().Add(m.sessionExpiry)
	session.ExpiresAt = newExpiresAt

	// Return a copy
	return &usecase.Session{
		ID:             session.ID,
		UserID:         session.UserID,
		UserRole:       session.UserRole,
		CreatedAt:      session.CreatedAt,
		ExpiresAt:      newExpiresAt,
		LastAccessedAt: session.LastAccessedAt,
		ClientIP:       session.ClientIP,
		UserAgent:      session.UserAgent,
		CSRFToken:      session.CSRFToken,
		IsActive:       session.IsActive,
	}, nil
}

// CleanupExpiredSessions removes expired sessions
func (m *MemorySessionManager) CleanupExpiredSessions(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for sessionID, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			delete(m.sessions, sessionID)
		}
	}

	return nil
}

// getClientIPFromContext extracts client IP from context
func getClientIPFromContext(ctx context.Context) string {
	if clientIP := ctx.Value(usecase.ContextKeyClientIP); clientIP != nil {
		if ip, ok := clientIP.(string); ok {
			return ip
		}
	}
	return "unknown"
}

// getUserAgentFromContext extracts User-Agent from context
func getUserAgentFromContext(ctx context.Context) string {
	if userAgent := ctx.Value(usecase.ContextKeyUserAgent); userAgent != nil {
		if ua, ok := userAgent.(string); ok {
			return ua
		}
	}
	return "unknown"
}
