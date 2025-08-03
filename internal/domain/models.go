package domain

import "time"

type ID = string

type Staff struct {
	ID           ID        `json:"id"`
	Name         string    `json:"name"`
	Role         StaffRole `json:"role"`
	PasswordHash string    `json:"-"` // Never include in JSON output for security
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type StaffRole string

const (
	RoleAdmin    StaffRole = "admin"
	RoleStaff    StaffRole = "staff"
	RoleReadOnly StaffRole = "readonly"
)

type Recipient struct {
	ID               ID         `json:"id"`
	Name             string     `json:"name"`
	Kana             string     `json:"kana"`
	Sex              Sex        `json:"sex"`
	BirthDate        time.Time  `json:"birth_date"`
	DisabilityName   string     `json:"disability_name"`
	HasDisabilityID  bool       `json:"has_disability_id"`
	Grade            string     `json:"grade"`
	Address          string     `json:"address"`
	Phone            string     `json:"phone"`
	Email            string     `json:"email"`
	PublicAssistance bool       `json:"public_assistance"`
	AdmissionDate    *time.Time `json:"admission_date,omitempty"`
	DischargeDate    *time.Time `json:"discharge_date,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Sex string

const (
	SexFemale Sex = "female"
	SexMale   Sex = "male"
	SexOther  Sex = "other"
	SexNA     Sex = "na"
)

type BenefitCertificate struct {
	ID                     ID        `json:"id"`
	RecipientID            ID        `json:"recipient_id"`
	StartDate              time.Time `json:"start_date"`
	EndDate                time.Time `json:"end_date"`
	Issuer                 string    `json:"issuer"`
	ServiceType            string    `json:"service_type"`
	MaxBenefitDaysPerMonth int       `json:"max_benefit_days_per_month"`
	BenefitDetails         string    `json:"benefit_details"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type StaffAssignment struct {
	ID           ID         `json:"id"`
	RecipientID  ID         `json:"recipient_id"`
	StaffID      ID         `json:"staff_id"`
	Role         string     `json:"role"`
	AssignedAt   time.Time  `json:"assigned_at"`
	UnassignedAt *time.Time `json:"unassigned_at,omitempty"`
}

type Consent struct {
	ID          ID         `json:"id"`
	RecipientID ID         `json:"recipient_id"`
	StaffID     ID         `json:"staff_id"`
	ConsentType string     `json:"consent_type"`
	Content     string     `json:"content"`
	Method      string     `json:"method"`
	ObtainedAt  time.Time  `json:"obtained_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

type AuditLog struct {
	ID      ID        `json:"id"`
	ActorID ID        `json:"actor_id"`
	Action  string    `json:"action"`
	Target  string    `json:"target"`
	At      time.Time `json:"at"`
	IP      string    `json:"ip"`
	Details string    `json:"details"`
}

// AuditLogFilter defines filters for querying audit logs
type AuditLogFilter struct {
	ActorID   *ID
	Action    *string
	Target    *string
	StartTime *time.Time
	EndTime   *time.Time
}

// ブルートフォース攻撃対策のためのレート制限モデル

// LoginAttempt represents a login attempt record for rate limiting
type LoginAttempt struct {
	ID          ID        `json:"id"`
	IPAddress   string    `json:"ip_address"`
	Username    string    `json:"username"`
	Success     bool      `json:"success"`
	AttemptedAt time.Time `json:"attempted_at"`
	UserAgent   string    `json:"user_agent,omitempty"`
}

// AccountLockout represents an account lockout record
type AccountLockout struct {
	ID           ID          `json:"id"`
	Username     string      `json:"username"`
	IPAddress    string      `json:"ip_address"`
	LockoutType  LockoutType `json:"lockout_type"`
	LockedAt     time.Time   `json:"locked_at"`
	UnlockedAt   *time.Time  `json:"unlocked_at,omitempty"`
	Reason       string      `json:"reason"`
	FailureCount int         `json:"failure_count"`
	Duration     int         `json:"duration"` // seconds
}

// LockoutType defines the type of lockout
type LockoutType string

const (
	LockoutTypeAccount LockoutType = "account" // Username-based lockout
	LockoutTypeIP      LockoutType = "ip"      // IP address-based lockout
	LockoutTypeMixed   LockoutType = "mixed"   // Both account and IP
	LockoutTypeManual  LockoutType = "manual"  // Manual lockout by admin
)

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	ID                       ID        `json:"id"`
	MaxAttemptsPerIP         int       `json:"max_attempts_per_ip"`
	MaxAttemptsPerUser       int       `json:"max_attempts_per_user"`
	WindowSizeMinutes        int       `json:"window_size_minutes"`
	LockoutDurationMinutes   int       `json:"lockout_duration_minutes"`
	BackoffMultiplier        float64   `json:"backoff_multiplier"`
	MaxLockoutHours          int       `json:"max_lockout_hours"`
	WhitelistIPs             []string  `json:"whitelist_ips"`
	EnableProgressiveLockout bool      `json:"enable_progressive_lockout"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// AttackPattern represents detected attack patterns
type AttackPattern struct {
	ID              ID        `json:"id"`
	PatternType     string    `json:"pattern_type"`
	SourceIP        string    `json:"source_ip"`
	TargetUsernames []string  `json:"target_usernames"`
	AttemptsCount   int       `json:"attempts_count"`
	FirstDetectedAt time.Time `json:"first_detected_at"`
	LastDetectedAt  time.Time `json:"last_detected_at"`
	Severity        string    `json:"severity"`
	Status          string    `json:"status"` // active, blocked, resolved
	Details         string    `json:"details"`
}
