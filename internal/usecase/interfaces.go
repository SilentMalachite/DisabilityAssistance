package usecase

import (
	"context"
	"time"

	"shien-system/internal/domain"
)

// RecipientUseCase defines business operations for recipient management
type RecipientUseCase interface {
	// CreateRecipient creates a new recipient with audit logging
	CreateRecipient(ctx context.Context, req CreateRecipientRequest) (*domain.Recipient, error)

	// GetRecipient retrieves a recipient by ID with access control
	GetRecipient(ctx context.Context, id domain.ID) (*domain.Recipient, error)

	// UpdateRecipient updates recipient information with audit logging
	UpdateRecipient(ctx context.Context, req UpdateRecipientRequest) (*domain.Recipient, error)

	// DeleteRecipient soft deletes a recipient with cascade handling
	DeleteRecipient(ctx context.Context, id domain.ID) error

	// ListRecipients retrieves paginated list of recipients
	ListRecipients(ctx context.Context, req ListRecipientsRequest) (*PaginatedRecipients, error)

	// GetActiveRecipients retrieves all currently active recipients
	GetActiveRecipients(ctx context.Context) ([]*domain.Recipient, error)

	// AssignStaff assigns staff members to a recipient
	AssignStaff(ctx context.Context, req AssignStaffRequest) error

	// UnassignStaff removes staff assignment from a recipient
	UnassignStaff(ctx context.Context, req UnassignStaffRequest) error
}

// StaffUseCase defines business operations for staff management
type StaffUseCase interface {
	// CreateStaff creates a new staff member with validation
	CreateStaff(ctx context.Context, req CreateStaffRequest) (*domain.Staff, error)

	// GetStaff retrieves a staff member by ID
	GetStaff(ctx context.Context, id domain.ID) (*domain.Staff, error)

	// UpdateStaff updates staff information
	UpdateStaff(ctx context.Context, req UpdateStaffRequest) (*domain.Staff, error)

	// DeleteStaff deletes a staff member with assignment validation
	DeleteStaff(ctx context.Context, id domain.ID) error

	// ListStaff retrieves paginated list of staff
	ListStaff(ctx context.Context, req ListStaffRequest) (*PaginatedStaff, error)

	// GetStaffByRole retrieves staff members by role
	GetStaffByRole(ctx context.Context, role domain.StaffRole) ([]*domain.Staff, error)

	// GetAssignments retrieves assignments for a staff member
	GetAssignments(ctx context.Context, staffID domain.ID) ([]*domain.StaffAssignment, error)
}

// CertificateUseCase defines business operations for benefit certificate management
type CertificateUseCase interface {
	// CreateCertificate creates a new benefit certificate
	CreateCertificate(ctx context.Context, req CreateCertificateRequest) (*domain.BenefitCertificate, error)

	// GetCertificate retrieves a certificate by ID
	GetCertificate(ctx context.Context, id domain.ID) (*domain.BenefitCertificate, error)

	// UpdateCertificate updates certificate information
	UpdateCertificate(ctx context.Context, req UpdateCertificateRequest) (*domain.BenefitCertificate, error)

	// DeleteCertificate deletes a certificate
	DeleteCertificate(ctx context.Context, id domain.ID) error

	// GetCertificatesByRecipient retrieves all certificates for a recipient
	GetCertificatesByRecipient(ctx context.Context, recipientID domain.ID) ([]*domain.BenefitCertificate, error)

	// GetExpiringSoon retrieves certificates expiring soon
	GetExpiringSoon(ctx context.Context, days int) ([]*domain.BenefitCertificate, error)

	// ValidateCertificate checks if a certificate is valid for a given date
	ValidateCertificate(ctx context.Context, certificateID domain.ID, date time.Time) (*ValidationResult, error)
}

// AuditUseCase defines business operations for audit log management
type AuditUseCase interface {
	// LogAction records an audit log entry
	LogAction(ctx context.Context, req LogActionRequest) error

	// GetAuditLogs retrieves paginated audit logs
	GetAuditLogs(ctx context.Context, req GetAuditLogsRequest) (*PaginatedAuditLogs, error)

	// SearchAuditLogs searches audit logs with criteria
	SearchAuditLogs(ctx context.Context, req SearchAuditLogsRequest) (*PaginatedAuditLogs, error)

	// GetAuditLogsByActor retrieves audit logs for a specific actor
	GetAuditLogsByActor(ctx context.Context, actorID domain.ID, limit, offset int) ([]*domain.AuditLog, error)
}

// BackupUseCase defines business operations for backup management
// BackupUseCase interface moved to backup_usecase.go to avoid duplication

// AuthUseCase defines business operations for authentication and session management
type AuthUseCase interface {
	// Login authenticates a user and creates a session
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)

	// Logout ends a user session
	Logout(ctx context.Context, req LogoutRequest) error

	// ValidateSession validates a session and returns user information
	ValidateSession(ctx context.Context, sessionID string) (*SessionInfo, error)

	// ChangePassword changes user password with validation
	ChangePassword(ctx context.Context, req ChangePasswordRequest) error

	// RefreshSession extends session expiration time
	RefreshSession(ctx context.Context, sessionID string) (*SessionInfo, error)
}

// PasswordHasher defines interface for password hashing operations
type PasswordHasher interface {
	// HashPassword creates a secure hash from a password
	HashPassword(password string) (string, error)

	// CheckPassword verifies a password against its hash
	CheckPassword(hashedPassword, password string) error
}

// SessionManager defines interface for session management
type SessionManager interface {
	// CreateSession creates a new session for a user
	CreateSession(ctx context.Context, userID domain.ID, userRole domain.StaffRole) (*Session, error)

	// ValidateSession validates and retrieves session information
	ValidateSession(ctx context.Context, sessionID string) (*Session, error)

	// DeleteSession removes a session
	DeleteSession(ctx context.Context, sessionID string) error

	// RefreshSession extends session expiration
	RefreshSession(ctx context.Context, sessionID string) (*Session, error)

	// CleanupExpiredSessions removes expired sessions
	CleanupExpiredSessions(ctx context.Context) error
}

// CSRFProtectedSessionManager extends SessionManager with CSRF protection
type CSRFProtectedSessionManager interface {
	SessionManager

	// ValidateCSRFToken validates CSRF token for a session
	ValidateCSRFToken(ctx context.Context, sessionID, csrfToken string) error
}

// Request/Response types for usecase operations

type CreateRecipientRequest struct {
	Name             string
	Kana             string
	Sex              domain.Sex
	BirthDate        time.Time
	DisabilityName   string
	HasDisabilityID  bool
	Grade            string
	Address          string
	Phone            string
	Email            string
	PublicAssistance bool
	AdmissionDate    *time.Time
	ActorID          domain.ID // For audit logging
}

type UpdateRecipientRequest struct {
	ID               domain.ID
	Name             string
	Kana             string
	Sex              domain.Sex
	BirthDate        time.Time
	DisabilityName   string
	HasDisabilityID  bool
	Grade            string
	Address          string
	Phone            string
	Email            string
	PublicAssistance bool
	AdmissionDate    *time.Time
	DischargeDate    *time.Time
	ActorID          domain.ID // For audit logging
}

type ListRecipientsRequest struct {
	Limit    int
	Offset   int
	FilterBy FilterRecipients
}

type FilterRecipients struct {
	AssignedToStaff     *domain.ID
	HasActiveAssignment *bool
	PublicAssistance    *bool
}

type PaginatedRecipients struct {
	Recipients []*domain.Recipient
	Total      int
	Limit      int
	Offset     int
}

type AssignStaffRequest struct {
	RecipientID domain.ID
	StaffID     domain.ID
	Role        string
	ActorID     domain.ID // For audit logging
}

type UnassignStaffRequest struct {
	AssignmentID domain.ID
	ActorID      domain.ID // For audit logging
}

type CreateStaffRequest struct {
	Name    string
	Role    domain.StaffRole
	ActorID domain.ID // For audit logging
}

type UpdateStaffRequest struct {
	ID      domain.ID
	Name    string
	Role    domain.StaffRole
	ActorID domain.ID // For audit logging
}

type ListStaffRequest struct {
	Limit    int
	Offset   int
	FilterBy FilterStaff
}

type FilterStaff struct {
	Role           *domain.StaffRole
	HasAssignments *bool
}

type PaginatedStaff struct {
	Staff  []*domain.Staff
	Total  int
	Limit  int
	Offset int
}

type CreateCertificateRequest struct {
	RecipientID            domain.ID
	StartDate              time.Time
	EndDate                time.Time
	Issuer                 string
	ServiceType            string
	MaxBenefitDaysPerMonth int
	BenefitDetails         string
	ActorID                domain.ID // For audit logging
}

type UpdateCertificateRequest struct {
	ID                     domain.ID
	StartDate              time.Time
	EndDate                time.Time
	Issuer                 string
	ServiceType            string
	MaxBenefitDaysPerMonth int
	BenefitDetails         string
	ActorID                domain.ID // For audit logging
}

type ValidationResult struct {
	IsValid   bool
	Reason    string
	ExpiresAt *time.Time
}

type LogActionRequest struct {
	ActorID domain.ID
	Action  string
	Target  string
	IP      string
	Details string
}

type GetAuditLogsRequest struct {
	Limit  int
	Offset int
}

type SearchAuditLogsRequest struct {
	Query  domain.AuditLogQuery
	Limit  int
	Offset int
}

type PaginatedAuditLogs struct {
	Logs   []*domain.AuditLog
	Total  int
	Limit  int
	Offset int
}

// Backup related types

// CreateBackupRequest moved to backup_usecase.go to avoid duplication

type BackupResult struct {
	BackupInfo *BackupInfo
	Success    bool
	Message    string
	Error      string
}

// BackupInfo moved to backup_usecase.go to avoid duplication

// RestoreBackupRequest moved to backup_usecase.go to avoid duplication

type RestoreResult struct {
	Success        bool                `json:"success"`
	StartedAt      time.Time           `json:"started_at"`
	CompletedAt    time.Time           `json:"completed_at"`
	Duration       time.Duration       `json:"duration"`
	RestoredFiles  []RestoredFileInfo  `json:"restored_files"`
	Errors         []string            `json:"errors,omitempty"`
	PreBackupID    string              `json:"pre_backup_id,omitempty"`
}

type RestoredFileInfo struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Size       int64  `json:"size"`
	Restored   bool   `json:"restored"`
	Error      string `json:"error,omitempty"`
}

type BackupStats struct {
	TotalBackups      int          `json:"total_backups"`
	TotalSizeBytes    int64        `json:"total_size_bytes"`
	OldestBackup      *BackupInfo  `json:"oldest_backup"`
	NewestBackup      *BackupInfo  `json:"newest_backup"`
	VerifiedBackups   int          `json:"verified_backups"`
	EncryptedBackups  int          `json:"encrypted_backups"`
	CompressedBackups int          `json:"compressed_backups"`
}

type ScheduleInfo struct {
	Enabled          bool      `json:"enabled"`
	Running          bool      `json:"running"`
	Interval         string    `json:"interval"`
	ScheduleTime     string    `json:"schedule_time"`
	LastRun          time.Time `json:"last_run"`
	NextRun          time.Time `json:"next_run"`
	RetryCount       int       `json:"retry_count"`
	RetryIntervalSec int       `json:"retry_interval_sec"`
}

// Authentication related types

type LoginRequest struct {
	Username  string
	Password  string
	ClientIP  string
	UserAgent string
}

type LoginResponse struct {
	SessionID string
	User      *domain.Staff
	ExpiresAt time.Time
	CSRFToken string
}

type LogoutRequest struct {
	SessionID string
	ClientIP  string
}

type ChangePasswordRequest struct {
	UserID      domain.ID
	OldPassword string
	NewPassword string
	ClientIP    string
}

type SessionInfo struct {
	SessionID string
	User      *domain.Staff
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Session struct {
	ID                 string
	UserID             domain.ID
	UserRole           domain.StaffRole
	CreatedAt          time.Time
	ExpiresAt          time.Time
	LastAccessedAt     time.Time
	ClientIP           string
	UserAgent          string
	CSRFToken          string
	IsActive           bool
	InvalidationReason string
	InvalidatedAt      *time.Time
}

// Context keys for passing user information
type ContextKey string

const (
	ContextKeyUserID    ContextKey = "user_id"
	ContextKeyUserRole  ContextKey = "user_role"
	ContextKeyClientIP  ContextKey = "client_ip"
	ContextKeyUserAgent ContextKey = "user_agent"
	ContextKeyCSRFToken ContextKey = "csrf_token"
)

// Common errors for usecase layer
var (
	ErrUnauthorized        = &UseCaseError{Code: "UNAUTHORIZED", Message: "操作する権限がありません"}
	ErrValidationFailed    = &UseCaseError{Code: "VALIDATION_FAILED", Message: "入力値が不正です"}
	ErrRecipientNotFound   = &UseCaseError{Code: "RECIPIENT_NOT_FOUND", Message: "利用者が見つかりません"}
	ErrStaffNotFound       = &UseCaseError{Code: "STAFF_NOT_FOUND", Message: "職員が見つかりません"}
	ErrCertificateNotFound = &UseCaseError{Code: "CERTIFICATE_NOT_FOUND", Message: "受給者証が見つかりません"}
	ErrAssignmentExists    = &UseCaseError{Code: "ASSIGNMENT_EXISTS", Message: "既に担当者が割り当てられています"}
	ErrCannotDeleteStaff   = &UseCaseError{Code: "CANNOT_DELETE_STAFF", Message: "担当中のため職員を削除できません"}

	// Authentication related errors
	ErrInvalidCredentials = &UseCaseError{Code: "INVALID_CREDENTIALS", Message: "ユーザー名またはパスワードが正しくありません"}
	ErrInvalidSession     = &UseCaseError{Code: "INVALID_SESSION", Message: "セッションが無効です"}
	ErrSessionExpired     = &UseCaseError{Code: "SESSION_EXPIRED", Message: "セッションの有効期限が切れています"}
	ErrInvalidPassword    = &UseCaseError{Code: "INVALID_PASSWORD", Message: "パスワードが正しくありません"}
	ErrWeakPassword       = &UseCaseError{Code: "WEAK_PASSWORD", Message: "パスワードが安全でありません"}
	ErrPasswordRequired   = &UseCaseError{Code: "PASSWORD_REQUIRED", Message: "パスワードは必須です"}

	// Session security related errors
	ErrSessionLimitExceeded   = &UseCaseError{Code: "SESSION_LIMIT_EXCEEDED", Message: "同時セッション数の上限に達しています"}
	ErrInvalidCSRFToken       = &UseCaseError{Code: "INVALID_CSRF_TOKEN", Message: "CSRF トークンが無効です"}
	ErrIPAddressMismatch      = &UseCaseError{Code: "IP_ADDRESS_MISMATCH", Message: "IPアドレスが一致しません"}
	ErrUserAgentMismatch      = &UseCaseError{Code: "USER_AGENT_MISMATCH", Message: "User-Agentが一致しません"}
	ErrSessionFixationAttempt = &UseCaseError{Code: "SESSION_FIXATION_ATTEMPT", Message: "セッション固定攻撃が検出されました"}

	// ブルートフォース攻撃対策のためのエラー
	ErrAccountLocked   = &UseCaseError{Code: "ACCOUNT_LOCKED", Message: "アカウントがロックされています"}
	ErrIPBlocked       = &UseCaseError{Code: "IP_BLOCKED", Message: "IPアドレスがブロックされています"}
	ErrTooManyAttempts = &UseCaseError{Code: "TOO_MANY_ATTEMPTS", Message: "ログイン試行回数が上限に達しました"}
	ErrLoginRestricted = &UseCaseError{Code: "LOGIN_RESTRICTED", Message: "ログインが制限されています"}
)

type UseCaseError struct {
	Code    string
	Message string
	Cause   error
}

func (e *UseCaseError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *UseCaseError) Unwrap() error {
	return e.Cause
}
