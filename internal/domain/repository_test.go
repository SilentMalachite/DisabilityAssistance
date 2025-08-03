package domain

import (
	"context"
	"testing"
	"time"
)

func TestRepositoryError(t *testing.T) {
	tests := []struct {
		name     string
		err      *RepositoryError
		expected string
	}{
		{
			name:     "error with operation only",
			err:      &RepositoryError{Op: "test operation"},
			expected: "test operation",
		},
		{
			name:     "error with operation and underlying error",
			err:      &RepositoryError{Op: "test operation", Err: ErrNotFound},
			expected: "test operation: not found",
		},
		{
			name:     "predefined not found error",
			err:      ErrNotFound,
			expected: "not found",
		},
		{
			name:     "predefined already exists error",
			err:      ErrAlreadyExists,
			expected: "already exists",
		},
		{
			name:     "predefined invalid input error",
			err:      ErrInvalidInput,
			expected: "invalid input",
		},
		{
			name:     "predefined constraint error",
			err:      ErrConstraint,
			expected: "constraint violation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("RepositoryError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRepositoryError_Unwrap(t *testing.T) {
	underlying := ErrNotFound
	err := &RepositoryError{Op: "test", Err: underlying}

	if unwrapped := err.Unwrap(); unwrapped != underlying {
		t.Errorf("RepositoryError.Unwrap() = %v, want %v", unwrapped, underlying)
	}

	// Test nil underlying error
	errWithoutUnderlying := &RepositoryError{Op: "test"}
	if unwrapped := errWithoutUnderlying.Unwrap(); unwrapped != nil {
		t.Errorf("RepositoryError.Unwrap() = %v, want nil", unwrapped)
	}
}

func TestAuditLogQuery_Structure(t *testing.T) {
	actorID := "staff-001"
	action := "CREATE"
	target := "recipient:recipient-001"
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()
	ip := "192.168.1.1"

	query := AuditLogQuery{
		ActorID:   &actorID,
		Action:    &action,
		Target:    &target,
		StartTime: &startTime,
		EndTime:   &endTime,
		IP:        &ip,
	}

	// Test that all fields are properly set
	if query.ActorID == nil || *query.ActorID != actorID {
		t.Errorf("ActorID = %v, want %v", query.ActorID, actorID)
	}

	if query.Action == nil || *query.Action != action {
		t.Errorf("Action = %v, want %v", query.Action, action)
	}

	if query.Target == nil || *query.Target != target {
		t.Errorf("Target = %v, want %v", query.Target, target)
	}

	if query.StartTime == nil || !query.StartTime.Equal(startTime) {
		t.Errorf("StartTime = %v, want %v", query.StartTime, startTime)
	}

	if query.EndTime == nil || !query.EndTime.Equal(endTime) {
		t.Errorf("EndTime = %v, want %v", query.EndTime, endTime)
	}

	if query.IP == nil || *query.IP != ip {
		t.Errorf("IP = %v, want %v", query.IP, ip)
	}
}

func TestMigrationStatus_Structure(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	status := MigrationStatus{
		Version:   "0001",
		Name:      "init",
		AppliedAt: now,
	}

	if status.Version != "0001" {
		t.Errorf("Version = %v, want %v", status.Version, "0001")
	}

	if status.Name != "init" {
		t.Errorf("Name = %v, want %v", status.Name, "init")
	}

	if !status.AppliedAt.Equal(now) {
		t.Errorf("AppliedAt = %v, want %v", status.AppliedAt, now)
	}
}

// Test that repository interfaces are properly defined by creating mock implementations
type mockStaffRepository struct{}

func (m *mockStaffRepository) Create(ctx context.Context, staff *Staff) error {
	return nil
}

func (m *mockStaffRepository) GetByID(ctx context.Context, id ID) (*Staff, error) {
	return nil, ErrNotFound
}

func (m *mockStaffRepository) Update(ctx context.Context, staff *Staff) error {
	return nil
}

func (m *mockStaffRepository) Delete(ctx context.Context, id ID) error {
	return nil
}

func (m *mockStaffRepository) List(ctx context.Context, limit, offset int) ([]*Staff, error) {
	return nil, nil
}

func (m *mockStaffRepository) GetByRole(ctx context.Context, role StaffRole) ([]*Staff, error) {
	return nil, nil
}

func (m *mockStaffRepository) GetByName(ctx context.Context, name string) ([]*Staff, error) {
	return nil, nil
}

func (m *mockStaffRepository) GetByExactName(ctx context.Context, name string) (*Staff, error) {
	return nil, ErrNotFound
}

func (m *mockStaffRepository) Count(ctx context.Context) (int, error) {
	return 0, nil
}

func TestStaffRepositoryInterface(t *testing.T) {
	var repo StaffRepository = &mockStaffRepository{}

	ctx := context.Background()

	// Test that all methods are callable
	staff := &Staff{ID: "test-001", Name: "Test User", Role: RoleStaff}

	err := repo.Create(ctx, staff)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	_, err = repo.GetByID(ctx, "test-001")
	if err != ErrNotFound {
		t.Errorf("GetByID() error = %v, want %v", err, ErrNotFound)
	}

	err = repo.Update(ctx, staff)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	err = repo.Delete(ctx, "test-001")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	_, err = repo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	_, err = repo.GetByRole(ctx, RoleAdmin)
	if err != nil {
		t.Errorf("GetByRole() error = %v", err)
	}

	_, err = repo.GetByName(ctx, "Test")
	if err != nil {
		t.Errorf("GetByName() error = %v", err)
	}

	_, err = repo.Count(ctx)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
}

// Test interface segregation - each repository should be usable independently
func TestRepositoryInterfaceSegregation(t *testing.T) {
	// This test ensures that each repository interface can be used independently
	// and doesn't require implementation of other interfaces

	var staffRepo StaffRepository
	var recipientRepo RecipientRepository
	var certRepo BenefitCertificateRepository
	var assignmentRepo StaffAssignmentRepository
	var consentRepo ConsentRepository
	var auditRepo AuditLogRepository

	// Test that interface variables can be declared (compile-time check)
	if staffRepo != nil || recipientRepo != nil || certRepo != nil ||
		assignmentRepo != nil || consentRepo != nil || auditRepo != nil {
		// This will never execute, but ensures interfaces are properly defined
		t.Log("All repository interfaces are properly defined")
	}
}

// Test that DatabaseRepository provides access to all interfaces correctly
func TestDatabaseRepositoryAccess(t *testing.T) {
	// This is a compile-time test to ensure DatabaseRepository
	// properly provides access to all repository interfaces
	var db DatabaseRepository

	// Test that DatabaseRepository provides methods to access component interfaces
	if db != nil {
		var _ StaffRepository = db.Staff()
		var _ RecipientRepository = db.Recipients()
		var _ BenefitCertificateRepository = db.BenefitCertificates()
		var _ StaffAssignmentRepository = db.StaffAssignments()
		var _ ConsentRepository = db.Consents()
		var _ AuditLogRepository = db.AuditLogs()
		var _ Transactional = db
		var _ Migrator = db
	}

	// This test ensures the interface design is correct at compile time
	t.Log("DatabaseRepository properly provides access to all interfaces")
}

func TestAuditLogQueryOptionalFields(t *testing.T) {
	// Test that AuditLogQuery works with all fields nil (empty query)
	query := AuditLogQuery{}

	if query.ActorID != nil {
		t.Error("ActorID should be nil in empty query")
	}

	if query.Action != nil {
		t.Error("Action should be nil in empty query")
	}

	if query.Target != nil {
		t.Error("Target should be nil in empty query")
	}

	if query.StartTime != nil {
		t.Error("StartTime should be nil in empty query")
	}

	if query.EndTime != nil {
		t.Error("EndTime should be nil in empty query")
	}

	if query.IP != nil {
		t.Error("IP should be nil in empty query")
	}
}

func BenchmarkRepositoryError_Error(b *testing.B) {
	err := &RepositoryError{Op: "benchmark test", Err: ErrNotFound}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
