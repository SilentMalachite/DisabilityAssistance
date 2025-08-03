package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"shien-system/internal/adapter/backup"
	"shien-system/internal/config"
	"shien-system/internal/domain"
)

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", message, err)
	}
	return fmt.Errorf("%s", message)
}

// generateID generates a unique ID
func generateID() string {
	return uuid.New().String()
}

// getClientIP extracts client IP from context
func getClientIP(ctx context.Context) string {
	if ip := ctx.Value("client_ip"); ip != nil {
		if ipStr, ok := ip.(string); ok {
			return ipStr
		}
	}
	return "127.0.0.1"
}

type BackupUseCase struct {
	backupService *backup.Service
	scheduler     *backup.Scheduler
	auditRepo     domain.AuditLogRepository
	logger        backup.Logger
}

// NewBackupUseCase creates a new backup use case
func NewBackupUseCase(
	backupService *backup.Service,
	scheduler *backup.Scheduler,
	auditRepo domain.AuditLogRepository,
	logger backup.Logger,
) *BackupUseCase {
	return &BackupUseCase{
		backupService: backupService,
		scheduler:     scheduler,
		auditRepo:     auditRepo,
		logger:        logger,
	}
}

// CreateBackup creates a manual backup
func (u *BackupUseCase) CreateBackup(ctx context.Context, req CreateBackupRequest) (*CreateBackupResponse, error) {
	// Validate request
	if err := u.validateCreateBackupRequest(req); err != nil {
		return nil, NewValidationError("Invalid backup request", err)
	}
	
	// Log start of backup operation
	u.logger.Info("Starting manual backup creation", 
		"actor_id", req.ActorID,
		"type", req.Type,
		"description", req.Description)
	
	// Create backup using service
	serviceReq := config.CreateBackupRequest{
		Type:        config.BackupType(req.Type),
		Description: req.Description,
		ActorID:     req.ActorID,
	}
	
	resp, err := u.backupService.CreateBackup(ctx, serviceReq)
	if err != nil {
		// Log failure
		u.logAuditEvent(ctx, req.ActorID, "backup_create_failed", "backup", err.Error())
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}
	
	// Log successful backup creation
	u.logAuditEvent(ctx, req.ActorID, "backup_created", "backup", 
		fmt.Sprintf("Backup created: ID=%s, Size=%d bytes", resp.BackupID, resp.Size))
	
	return &CreateBackupResponse{
		BackupID:    resp.BackupID,
		FilePath:    resp.FilePath,
		Size:        resp.Size,
		CreatedAt:   resp.Created,
		Description: req.Description,
		Type:        req.Type,
	}, nil
}

// RestoreBackup restores from a backup
func (u *BackupUseCase) RestoreBackup(ctx context.Context, req RestoreBackupRequest) (*RestoreBackupResponse, error) {
	// Validate request
	if err := u.validateRestoreBackupRequest(req); err != nil {
		return nil, NewValidationError("Invalid restore request", err)
	}
	
	// Log start of restore operation
	u.logger.Info("Starting backup restoration", 
		"actor_id", req.ActorID,
		"backup_id", req.BackupID,
		"overwrite", req.Overwrite)
	
	// Restore using service
	serviceReq := config.RestoreBackupRequest{
		BackupID:  req.BackupID,
		ActorID:   req.ActorID,
		Overwrite: req.Overwrite,
	}
	
	resp, err := u.backupService.RestoreBackup(ctx, serviceReq)
	if err != nil {
		// Log failure
		u.logAuditEvent(ctx, req.ActorID, "backup_restore_failed", "backup", err.Error())
		return nil, fmt.Errorf("failed to restore backup: %w", err)
	}
	
	// Log successful restoration
	u.logAuditEvent(ctx, req.ActorID, "backup_restored", "backup", 
		fmt.Sprintf("Backup restored: ID=%s, Records=%d", req.BackupID, resp.RecordsCount))
	
	return &RestoreBackupResponse{
		Success:      resp.Success,
		RestoredAt:   resp.RestoredAt,
		RecordsCount: resp.RecordsCount,
		BackupID:     req.BackupID,
	}, nil
}

// ListBackups lists available backups
func (u *BackupUseCase) ListBackups(ctx context.Context, req ListBackupsRequest) (*ListBackupsResponse, error) {
	// Validate request
	if err := u.validateListBackupsRequest(req); err != nil {
		return nil, NewValidationError("Invalid list request", err)
	}
	
	// List using service
	serviceReq := config.ListBackupsRequest{
		Limit:  req.Limit,
		Offset: req.Offset,
		Type:   req.Type,
	}
	
	resp, err := u.backupService.ListBackups(ctx, serviceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}
	
	// Convert response
	backups := make([]BackupInfo, 0, len(resp.Backups))
	for _, backup := range resp.Backups {
		backups = append(backups, BackupInfo{
			ID:          backup.ID,
			Type:        string(backup.Type),
			Description: backup.Description,
			Size:        backup.Size,
			Checksum:    backup.Checksum,
			CreatedAt:   backup.CreatedAt,
			CreatedBy:   backup.CreatedBy,
		})
	}
	
	return &ListBackupsResponse{
		Backups: backups,
		Total:   resp.Total,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}, nil
}

// DeleteBackup deletes a backup
func (u *BackupUseCase) DeleteBackup(ctx context.Context, req DeleteBackupRequest) error {
	// Validate request
	if err := u.validateDeleteBackupRequest(req); err != nil {
		return NewValidationError("Invalid delete request", err)
	}
	
	// Log start of delete operation
	u.logger.Info("Starting backup deletion", 
		"actor_id", req.ActorID,
		"backup_id", req.BackupID)
	
	// Delete using service
	serviceReq := config.DeleteBackupRequest{
		BackupID: req.BackupID,
		ActorID:  req.ActorID,
	}
	
	if err := u.backupService.DeleteBackup(ctx, serviceReq); err != nil {
		// Log failure
		u.logAuditEvent(ctx, req.ActorID, "backup_delete_failed", "backup", err.Error())
		return fmt.Errorf("failed to delete backup: %w", err)
	}
	
	// Log successful deletion
	u.logAuditEvent(ctx, req.ActorID, "backup_deleted", "backup", 
		fmt.Sprintf("Backup deleted: ID=%s", req.BackupID))
	
	return nil
}

// ValidateBackup validates a backup file
func (u *BackupUseCase) ValidateBackup(ctx context.Context, req ValidateBackupRequest) (*ValidateBackupResponse, error) {
	// Validate request
	if req.BackupID == "" {
		return nil, NewValidationError("Invalid validate request", fmt.Errorf("backup ID is required"))
	}
	
	// Validate using service
	serviceReq := config.ValidateBackupRequest{
		BackupID: req.BackupID,
	}
	
	resp, err := u.backupService.ValidateBackup(ctx, serviceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to validate backup: %w", err)
	}
	
	return &ValidateBackupResponse{
		BackupID:    req.BackupID,
		Valid:       resp.Valid,
		Error:       resp.Error,
		Checksum:    resp.Checksum,
		FileSize:    resp.FileSize,
		RecordCount: resp.RecordCount,
	}, nil
}

// StartScheduledBackup starts the backup scheduler
func (u *BackupUseCase) StartScheduledBackup(ctx context.Context, actorID string) error {
	if actorID == "" {
		return NewValidationError("Invalid request", fmt.Errorf("actor ID is required"))
	}
	
	u.logger.Info("Starting backup scheduler", "actor_id", actorID)
	
	if err := u.scheduler.Start(ctx); err != nil {
		u.logAuditEvent(ctx, actorID, "backup_scheduler_start_failed", "system", err.Error())
		return fmt.Errorf("failed to start backup scheduler: %w", err)
	}
	
	u.logAuditEvent(ctx, actorID, "backup_scheduler_started", "system", "Automatic backup scheduler started")
	return nil
}

// StopScheduledBackup stops the backup scheduler
func (u *BackupUseCase) StopScheduledBackup(ctx context.Context, actorID string) error {
	if actorID == "" {
		return NewValidationError("Invalid request", fmt.Errorf("actor ID is required"))
	}
	
	u.logger.Info("Stopping backup scheduler", "actor_id", actorID)
	
	if err := u.scheduler.Stop(); err != nil {
		u.logAuditEvent(ctx, actorID, "backup_scheduler_stop_failed", "system", err.Error())
		return fmt.Errorf("failed to stop backup scheduler: %w", err)
	}
	
	u.logAuditEvent(ctx, actorID, "backup_scheduler_stopped", "system", "Automatic backup scheduler stopped")
	return nil
}

// GetSchedulerStatus returns the current status of the backup scheduler
func (u *BackupUseCase) GetSchedulerStatus() SchedulerStatusResponse {
	status := u.scheduler.GetStatus()
	
	return SchedulerStatusResponse{
		Running:       status.Running,
		LastBackup:    status.LastBackup,
		NextScheduled: status.NextScheduled,
		Interval:      status.Interval,
	}
}

// validateCreateBackupRequest validates create backup request
func (u *BackupUseCase) validateCreateBackupRequest(req CreateBackupRequest) error {
	if req.ActorID == "" {
		return fmt.Errorf("actor ID is required")
	}
	
	if req.Type == "" {
		return fmt.Errorf("backup type is required")
	}
	
	// Validate backup type
	switch req.Type {
	case "full", "incremental", "manual":
		// Valid types
	default:
		return fmt.Errorf("invalid backup type: %s", req.Type)
	}
	
	return nil
}

// validateRestoreBackupRequest validates restore backup request
func (u *BackupUseCase) validateRestoreBackupRequest(req RestoreBackupRequest) error {
	if req.ActorID == "" {
		return fmt.Errorf("actor ID is required")
	}
	
	if req.BackupID == "" {
		return fmt.Errorf("backup ID is required")
	}
	
	return nil
}

// validateListBackupsRequest validates list backups request
func (u *BackupUseCase) validateListBackupsRequest(req ListBackupsRequest) error {
	if req.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	
	if req.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	
	return nil
}

// validateDeleteBackupRequest validates delete backup request
func (u *BackupUseCase) validateDeleteBackupRequest(req DeleteBackupRequest) error {
	if req.ActorID == "" {
		return fmt.Errorf("actor ID is required")
	}
	
	if req.BackupID == "" {
		return fmt.Errorf("backup ID is required")
	}
	
	return nil
}

// logAuditEvent logs an audit event
func (u *BackupUseCase) logAuditEvent(ctx context.Context, actorID, action, target, details string) {
	auditLog := &domain.AuditLog{
		ID:      generateID(),
		ActorID: actorID,
		Action:  action,
		Target:  target,
		At:      time.Now(),
		IP:      getClientIP(ctx),
		Details: details,
	}
	
	if err := u.auditRepo.Create(ctx, auditLog); err != nil {
		u.logger.Error("Failed to create audit log", "error", err)
	}
}

// Backup use case request/response types

// CreateBackupRequest represents a request to create a backup
type CreateBackupRequest struct {
	Type        string `json:"type"`        // "full", "incremental", "manual"
	Description string `json:"description"`
	ActorID     string `json:"actor_id"`
}

// CreateBackupResponse represents the response from creating a backup
type CreateBackupResponse struct {
	BackupID    string    `json:"backup_id"`
	FilePath    string    `json:"file_path"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
}

// RestoreBackupRequest represents a request to restore from a backup
type RestoreBackupRequest struct {
	BackupID  string `json:"backup_id"`
	ActorID   string `json:"actor_id"`
	Overwrite bool   `json:"overwrite"`
}

// RestoreBackupResponse represents the response from restoring a backup
type RestoreBackupResponse struct {
	Success      bool      `json:"success"`
	RestoredAt   time.Time `json:"restored_at"`
	RecordsCount int       `json:"records_count"`
	BackupID     string    `json:"backup_id"`
}

// ListBackupsRequest represents a request to list backups
type ListBackupsRequest struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Type   string `json:"type,omitempty"`
}

// ListBackupsResponse represents the response from listing backups
type ListBackupsResponse struct {
	Backups []BackupInfo `json:"backups"`
	Total   int          `json:"total"`
	Limit   int          `json:"limit"`
	Offset  int          `json:"offset"`
}

// DeleteBackupRequest represents a request to delete a backup
type DeleteBackupRequest struct {
	BackupID string `json:"backup_id"`
	ActorID  string `json:"actor_id"`
}

// ValidateBackupRequest represents a request to validate a backup
type ValidateBackupRequest struct {
	BackupID string `json:"backup_id"`
}

// ValidateBackupResponse represents the response from validating a backup
type ValidateBackupResponse struct {
	BackupID    string `json:"backup_id"`
	Valid       bool   `json:"valid"`
	Error       string `json:"error,omitempty"`
	Checksum    string `json:"checksum"`
	FileSize    int64  `json:"file_size"`
	RecordCount int    `json:"record_count"`
}

// BackupInfo represents backup metadata
type BackupInfo struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

// SchedulerStatusResponse represents the status of the backup scheduler
type SchedulerStatusResponse struct {
	Running       bool      `json:"running"`
	LastBackup    time.Time `json:"last_backup"`
	NextScheduled time.Time `json:"next_scheduled"`
	Interval      string    `json:"interval"`
}