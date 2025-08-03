package widgets

import (
	"context"
	"errors"

	"shien-system/internal/usecase"
)

// LogoutHandler handles logout operations with proper session management
type LogoutHandler struct {
	appState *AppState
}

// NewLogoutHandler creates a new logout handler
func NewLogoutHandler(appState *AppState) *LogoutHandler {
	return &LogoutHandler{
		appState: appState,
	}
}

// PerformLogout performs logout with session invalidation
func (lh *LogoutHandler) PerformLogout() error {
	return lh.PerformLogoutWithIP("")
}

// PerformLogoutWithIP performs logout with session invalidation and client IP
func (lh *LogoutHandler) PerformLogoutWithIP(clientIP string) error {
	if !lh.appState.IsAuthenticated() {
		return errors.New("not authenticated")
	}

	// Get current session ID before clearing state
	sessionID := lh.appState.GetSessionID()

	// Invalidate session on server
	ctx := context.Background()
	req := usecase.LogoutRequest{
		SessionID: sessionID,
		ClientIP:  clientIP,
	}

	// Call logout use case to invalidate session
	err := lh.appState.authUseCase.Logout(ctx, req)

	// Clear local app state regardless of server response
	// This ensures the UI is updated even if session invalidation fails
	lh.appState.Logout()

	// Return any error from session invalidation
	return err
}

// GetConfirmationMessage returns a personalized logout confirmation message
func (lh *LogoutHandler) GetConfirmationMessage() string {
	if user := lh.appState.GetCurrentUser(); user != nil {
		return user.Name + "さん、ログアウトしますか？"
	}
	return "ログアウトしますか？"
}
