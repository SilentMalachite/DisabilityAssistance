package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shien-system/internal/domain"
	"shien-system/internal/usecase"
)

func TestSessionManager_CreateSession(t *testing.T) {
	manager := NewMemorySessionManager(8 * time.Hour)
	ctx := context.Background()

	userID := "user-123"
	userRole := domain.RoleStaff

	session, err := manager.CreateSession(ctx, userID, userRole)

	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, userRole, session.UserRole)
	assert.True(t, session.CreatedAt.Before(time.Now().Add(time.Second)))
	assert.True(t, session.ExpiresAt.After(time.Now().Add(7*time.Hour)))
}

func TestSessionManager_ValidateSession(t *testing.T) {
	manager := NewMemorySessionManager(8 * time.Hour)
	ctx := context.Background()

	// Create a session first
	userID := "user-123"
	userRole := domain.RoleAdmin
	session, err := manager.CreateSession(ctx, userID, userRole)
	require.NoError(t, err)

	// Validate the session
	validatedSession, err := manager.ValidateSession(ctx, session.ID)

	require.NoError(t, err)
	assert.Equal(t, session.ID, validatedSession.ID)
	assert.Equal(t, session.UserID, validatedSession.UserID)
	assert.Equal(t, session.UserRole, validatedSession.UserRole)
}

func TestSessionManager_ValidateSession_NonExistent(t *testing.T) {
	manager := NewMemorySessionManager(8 * time.Hour)
	ctx := context.Background()

	session, err := manager.ValidateSession(ctx, "non-existent-session")

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Equal(t, usecase.ErrInvalidSession, err)
}

func TestSessionManager_ValidateSession_Expired(t *testing.T) {
	manager := NewMemorySessionManager(100 * time.Millisecond) // Very short expiration
	ctx := context.Background()

	// Create a session
	session, err := manager.CreateSession(ctx, "user-123", domain.RoleStaff)
	require.NoError(t, err)

	// Wait for session to expire
	time.Sleep(200 * time.Millisecond)

	// Try to validate expired session
	validatedSession, err := manager.ValidateSession(ctx, session.ID)

	assert.Error(t, err)
	assert.Nil(t, validatedSession)
	assert.Equal(t, usecase.ErrSessionExpired, err)
}

func TestSessionManager_DeleteSession(t *testing.T) {
	manager := NewMemorySessionManager(8 * time.Hour)
	ctx := context.Background()

	// Create a session
	session, err := manager.CreateSession(ctx, "user-123", domain.RoleStaff)
	require.NoError(t, err)

	// Delete the session
	err = manager.DeleteSession(ctx, session.ID)
	require.NoError(t, err)

	// Try to validate deleted session
	validatedSession, err := manager.ValidateSession(ctx, session.ID)
	assert.Error(t, err)
	assert.Nil(t, validatedSession)
	assert.Equal(t, usecase.ErrInvalidSession, err)
}

func TestSessionManager_DeleteSession_NonExistent(t *testing.T) {
	manager := NewMemorySessionManager(8 * time.Hour)
	ctx := context.Background()

	// Try to delete non-existent session (should not error)
	err := manager.DeleteSession(ctx, "non-existent-session")
	assert.NoError(t, err)
}

func TestSessionManager_RefreshSession(t *testing.T) {
	manager := NewMemorySessionManager(8 * time.Hour)
	ctx := context.Background()

	// Create a session
	originalSession, err := manager.CreateSession(ctx, "user-123", domain.RoleStaff)
	require.NoError(t, err)

	// Store the original expiration time for comparison
	originalExpiresAt := originalSession.ExpiresAt

	// Wait a bit to ensure time difference
	time.Sleep(50 * time.Millisecond)

	// Refresh the session
	refreshedSession, err := manager.RefreshSession(ctx, originalSession.ID)

	require.NoError(t, err)
	assert.Equal(t, originalSession.ID, refreshedSession.ID)
	assert.Equal(t, originalSession.UserID, refreshedSession.UserID)
	assert.Equal(t, originalSession.UserRole, refreshedSession.UserRole)
	assert.Equal(t, originalSession.CreatedAt, refreshedSession.CreatedAt)

	// The refreshed session should have a later expiration time
	assert.True(t, refreshedSession.ExpiresAt.After(originalExpiresAt))
}

func TestSessionManager_RefreshSession_NonExistent(t *testing.T) {
	manager := NewMemorySessionManager(8 * time.Hour)
	ctx := context.Background()

	session, err := manager.RefreshSession(ctx, "non-existent-session")

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Equal(t, usecase.ErrInvalidSession, err)
}

func TestSessionManager_RefreshSession_Expired(t *testing.T) {
	manager := NewMemorySessionManager(100 * time.Millisecond)
	ctx := context.Background()

	// Create a session
	session, err := manager.CreateSession(ctx, "user-123", domain.RoleStaff)
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Try to refresh expired session
	refreshedSession, err := manager.RefreshSession(ctx, session.ID)

	assert.Error(t, err)
	assert.Nil(t, refreshedSession)
	assert.Equal(t, usecase.ErrSessionExpired, err)
}

func TestSessionManager_CleanupExpiredSessions(t *testing.T) {
	manager := NewMemorySessionManager(100 * time.Millisecond)
	ctx := context.Background()

	// Create multiple sessions
	session1, err := manager.CreateSession(ctx, "user-1", domain.RoleStaff)
	require.NoError(t, err)

	session2, err := manager.CreateSession(ctx, "user-2", domain.RoleAdmin)
	require.NoError(t, err)

	// Wait for sessions to expire
	time.Sleep(200 * time.Millisecond)

	// Create a new session (should not be expired)
	session3, err := manager.CreateSession(ctx, "user-3", domain.RoleReadOnly)
	require.NoError(t, err)

	// Run cleanup
	err = manager.CleanupExpiredSessions(ctx)
	require.NoError(t, err)

	// Verify expired sessions are gone
	_, err = manager.ValidateSession(ctx, session1.ID)
	assert.Equal(t, usecase.ErrInvalidSession, err)

	_, err = manager.ValidateSession(ctx, session2.ID)
	assert.Equal(t, usecase.ErrInvalidSession, err)

	// Verify active session is still there
	validSession, err := manager.ValidateSession(ctx, session3.ID)
	assert.NoError(t, err)
	assert.Equal(t, session3.ID, validSession.ID)
}

func TestSessionManager_ConcurrentAccess(t *testing.T) {
	manager := NewMemorySessionManager(8 * time.Hour)
	ctx := context.Background()

	// Test concurrent session creation
	const numGoroutines = 10
	sessionsChan := make(chan *usecase.Session, numGoroutines)
	errorsChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(userNum int) {
			session, err := manager.CreateSession(ctx, "user-"+string(rune(userNum)), domain.RoleStaff)
			if err != nil {
				errorsChan <- err
				return
			}
			sessionsChan <- session
		}(i)
	}

	// Collect results
	sessions := make([]*usecase.Session, 0, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		select {
		case session := <-sessionsChan:
			sessions = append(sessions, session)
		case err := <-errorsChan:
			t.Fatalf("Unexpected error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatalf("Timeout waiting for concurrent operations")
		}
	}

	// Verify all sessions are unique
	sessionIDs := make(map[string]bool)
	for _, session := range sessions {
		assert.False(t, sessionIDs[session.ID], "Duplicate session ID: %s", session.ID)
		sessionIDs[session.ID] = true
	}

	assert.Equal(t, numGoroutines, len(sessions))
}
