package domain

import (
	"context"
	"time"
)

// PasswordHasher defines the interface for password hashing
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(hashedPassword, password string) error
}

// StaffRepository defines the interface for staff data access
type StaffRepository interface {
	Create(ctx context.Context, staff *Staff) error
	GetByID(ctx context.Context, id ID) (*Staff, error)
	GetByExactName(ctx context.Context, name string) (*Staff, error)
	Update(ctx context.Context, staff *Staff) error
	Delete(ctx context.Context, id ID) error
	List(ctx context.Context, limit, offset int) ([]*Staff, error)
	GetByRole(ctx context.Context, role StaffRole) ([]*Staff, error)
	GetByName(ctx context.Context, name string) ([]*Staff, error)
	Count(ctx context.Context) (int, error)
}

// RecipientRepository defines the interface for recipient data access
type RecipientRepository interface {
	Create(ctx context.Context, recipient *Recipient) error
	GetByID(ctx context.Context, id ID) (*Recipient, error)
	Update(ctx context.Context, recipient *Recipient) error
	Delete(ctx context.Context, id ID) error
	List(ctx context.Context, limit, offset int) ([]*Recipient, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*Recipient, error)
	GetByStaffID(ctx context.Context, staffID ID) ([]*Recipient, error)
	GetActive(ctx context.Context, limit, offset int) ([]*Recipient, error) // Not discharged
	Count(ctx context.Context) (int, error)
	CountActive(ctx context.Context) (int, error)
}

// BenefitCertificateRepository defines the interface for benefit certificate data access
type BenefitCertificateRepository interface {
	Create(ctx context.Context, cert *BenefitCertificate) error
	GetByID(ctx context.Context, id ID) (*BenefitCertificate, error)
	Update(ctx context.Context, cert *BenefitCertificate) error
	Delete(ctx context.Context, id ID) error
	GetByRecipientID(ctx context.Context, recipientID ID) ([]*BenefitCertificate, error)
	GetExpiringSoon(ctx context.Context, within time.Duration) ([]*BenefitCertificate, error)
	GetActiveByRecipientID(ctx context.Context, recipientID ID, asOf time.Time) (*BenefitCertificate, error)
	List(ctx context.Context, limit, offset int) ([]*BenefitCertificate, error)
	Count(ctx context.Context) (int, error)
}

// StaffAssignmentRepository defines the interface for staff assignment data access
type StaffAssignmentRepository interface {
	Create(ctx context.Context, assignment *StaffAssignment) error
	GetByID(ctx context.Context, id ID) (*StaffAssignment, error)
	Update(ctx context.Context, assignment *StaffAssignment) error
	Delete(ctx context.Context, id ID) error
	GetByRecipientID(ctx context.Context, recipientID ID) ([]*StaffAssignment, error)
	GetByStaffID(ctx context.Context, staffID ID) ([]*StaffAssignment, error)
	GetActiveByRecipientID(ctx context.Context, recipientID ID) ([]*StaffAssignment, error)
	GetActiveByStaffID(ctx context.Context, staffID ID) ([]*StaffAssignment, error)
	UnassignAll(ctx context.Context, recipientID ID, unassignedAt time.Time) error
	List(ctx context.Context, limit, offset int) ([]*StaffAssignment, error)
	Count(ctx context.Context) (int, error)
}

// ConsentRepository defines the interface for consent data access
type ConsentRepository interface {
	Create(ctx context.Context, consent *Consent) error
	GetByID(ctx context.Context, id ID) (*Consent, error)
	Update(ctx context.Context, consent *Consent) error
	Delete(ctx context.Context, id ID) error
	GetByRecipientID(ctx context.Context, recipientID ID) ([]*Consent, error)
	GetByType(ctx context.Context, consentType string) ([]*Consent, error)
	GetActiveByRecipientID(ctx context.Context, recipientID ID) ([]*Consent, error)
	RevokeAllByRecipientID(ctx context.Context, recipientID ID, revokedAt time.Time) error
	List(ctx context.Context, limit, offset int) ([]*Consent, error)
	Count(ctx context.Context) (int, error)
}

// AuditLogRepository defines the interface for audit log data access
type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetByID(ctx context.Context, id ID) (*AuditLog, error)
	GetByActorID(ctx context.Context, actorID ID, limit, offset int) ([]*AuditLog, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]*AuditLog, error)
	GetByTarget(ctx context.Context, target string, limit, offset int) ([]*AuditLog, error)
	GetByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*AuditLog, error)
	Search(ctx context.Context, query AuditLogQuery, limit, offset int) ([]*AuditLog, error)
	List(ctx context.Context, limit, offset int) ([]*AuditLog, error)
	Count(ctx context.Context) (int, error)
	// Note: Audit logs should never be updated or deleted for integrity
}

// ブルートフォース攻撃対策のためのリポジトリインターフェース

// LoginAttemptRepository defines the interface for login attempt data access
type LoginAttemptRepository interface {
	Create(ctx context.Context, attempt *LoginAttempt) error
	GetByIPAddress(ctx context.Context, ipAddress string, since time.Time) ([]*LoginAttempt, error)
	GetByUsername(ctx context.Context, username string, since time.Time) ([]*LoginAttempt, error)
	GetFailedAttempts(ctx context.Context, ipAddress, username string, since time.Time) ([]*LoginAttempt, error)
	DeleteOldAttempts(ctx context.Context, before time.Time) error
	CountRecentFailures(ctx context.Context, ipAddress, username string, since time.Time) (int, error)
}

// AccountLockoutRepository defines the interface for account lockout data access
type AccountLockoutRepository interface {
	Create(ctx context.Context, lockout *AccountLockout) error
	GetByID(ctx context.Context, id ID) (*AccountLockout, error)
	GetActiveByUsername(ctx context.Context, username string) (*AccountLockout, error)
	GetActiveByIPAddress(ctx context.Context, ipAddress string) (*AccountLockout, error)
	GetActiveLockouts(ctx context.Context) ([]*AccountLockout, error)
	Unlock(ctx context.Context, id ID, unlockedAt time.Time) error
	UnlockByUsername(ctx context.Context, username string, unlockedAt time.Time) error
	UnlockByIPAddress(ctx context.Context, ipAddress string, unlockedAt time.Time) error
	CleanupExpiredLockouts(ctx context.Context) error
	List(ctx context.Context, limit, offset int) ([]*AccountLockout, error)
	Count(ctx context.Context) (int, error)
}

// RateLimitConfigRepository defines the interface for rate limit configuration
type RateLimitConfigRepository interface {
	Get(ctx context.Context) (*RateLimitConfig, error)
	Update(ctx context.Context, config *RateLimitConfig) error
	GetDefault() *RateLimitConfig
}

// AttackPatternRepository defines the interface for attack pattern detection
type AttackPatternRepository interface {
	Create(ctx context.Context, pattern *AttackPattern) error
	GetByID(ctx context.Context, id ID) (*AttackPattern, error)
	GetActiveByIP(ctx context.Context, sourceIP string) ([]*AttackPattern, error)
	GetByTimeRange(ctx context.Context, start, end time.Time) ([]*AttackPattern, error)
	UpdateStatus(ctx context.Context, id ID, status string) error
	List(ctx context.Context, limit, offset int) ([]*AttackPattern, error)
	Count(ctx context.Context) (int, error)
}

// AuditLogQuery represents search criteria for audit logs
type AuditLogQuery struct {
	ActorID   *ID        `json:"actor_id,omitempty"`
	Action    *string    `json:"action,omitempty"`
	Target    *string    `json:"target,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	IP        *string    `json:"ip,omitempty"`
}

// Repository interfaces for transactions and migrations
type Transactional interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Migrator interface {
	RunMigrations(ctx context.Context) error
	GetMigrationStatus(ctx context.Context) ([]MigrationStatus, error)
}

type MigrationStatus struct {
	Version   string    `json:"version"`
	Name      string    `json:"name"`
	AppliedAt time.Time `json:"applied_at"`
}

// DatabaseRepository provides access to all repository interfaces
type DatabaseRepository interface {
	Staff() StaffRepository
	Recipients() RecipientRepository
	BenefitCertificates() BenefitCertificateRepository
	StaffAssignments() StaffAssignmentRepository
	Consents() ConsentRepository
	AuditLogs() AuditLogRepository
	Transactional
	Migrator
	Close() error
}

// Repository errors
type RepositoryError struct {
	Op  string // operation that failed
	Err error  // underlying error
}

func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return e.Op + ": " + e.Err.Error()
	}
	return e.Op
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// Common repository errors
var (
	ErrNotFound      = &RepositoryError{Op: "not found"}
	ErrAlreadyExists = &RepositoryError{Op: "already exists"}
	ErrInvalidInput  = &RepositoryError{Op: "invalid input"}
	ErrConstraint    = &RepositoryError{Op: "constraint violation"}
)
