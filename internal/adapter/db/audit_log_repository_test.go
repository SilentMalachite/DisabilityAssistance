package db

import (
	"context"
	"testing"
	"time"

	"shien-system/internal/domain"
)

func setupAuditLogTestData(t *testing.T, db *Database) (context.Context, *domain.Staff) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Create staff member for audit logs
	staffRepo := NewStaffRepository(db)
	staff := &domain.Staff{
		ID:        "audit-staff-001",
		Name:      "監査ログテスト太郎",
		Role:      domain.RoleAdmin,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := staffRepo.Create(ctx, staff)
	if err != nil {
		t.Fatalf("Create staff error = %v", err)
	}

	return ctx, staff
}

func TestAuditLogRepository_Create(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff := setupAuditLogTestData(t, db)

	auditRepo := NewAuditLogRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	auditLog := &domain.AuditLog{
		ID:      "audit-create-001",
		ActorID: staff.ID,
		Action:  "CREATE",
		Target:  "recipient:recipient-001",
		At:      now,
		IP:      "192.168.1.100",
		Details: "新規利用者を作成しました",
	}

	err := auditRepo.Create(ctx, auditLog)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	// Verify the audit log was created
	retrieved, err := auditRepo.GetByID(ctx, auditLog.ID)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetByID() returned nil")
	}

	// Compare all fields
	if retrieved.ID != auditLog.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, auditLog.ID)
	}
	if retrieved.ActorID != auditLog.ActorID {
		t.Errorf("ActorID = %v, want %v", retrieved.ActorID, auditLog.ActorID)
	}
	if retrieved.Action != auditLog.Action {
		t.Errorf("Action = %v, want %v", retrieved.Action, auditLog.Action)
	}
	if retrieved.Target != auditLog.Target {
		t.Errorf("Target = %v, want %v", retrieved.Target, auditLog.Target)
	}
	if !retrieved.At.Equal(auditLog.At) {
		t.Errorf("At = %v, want %v", retrieved.At, auditLog.At)
	}
	if retrieved.IP != auditLog.IP {
		t.Errorf("IP = %v, want %v", retrieved.IP, auditLog.IP)
	}
	if retrieved.Details != auditLog.Details {
		t.Errorf("Details = %v, want %v", retrieved.Details, auditLog.Details)
	}
}

func TestAuditLogRepository_GetByActorID(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff := setupAuditLogTestData(t, db)

	// Create additional staff
	staffRepo := NewStaffRepository(db)
	staff2 := &domain.Staff{
		ID:        "audit-staff-002",
		Name:      "監査ログテスト花子",
		Role:      domain.RoleStaff,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := staffRepo.Create(ctx, staff2)
	if err != nil {
		t.Fatalf("Create staff2 error = %v", err)
	}

	auditRepo := NewAuditLogRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	// Create multiple audit logs for different actors
	auditLogs := []*domain.AuditLog{
		{
			ID:      "audit-actor-001",
			ActorID: staff.ID,
			Action:  "LOGIN",
			Target:  "system",
			At:      now,
			IP:      "192.168.1.100",
			Details: "ログインしました",
		},
		{
			ID:      "audit-actor-002",
			ActorID: staff.ID,
			Action:  "CREATE",
			Target:  "recipient:recipient-001",
			At:      now.Add(time.Minute),
			IP:      "192.168.1.100",
			Details: "利用者を作成しました",
		},
		{
			ID:      "audit-actor-003",
			ActorID: staff2.ID,
			Action:  "VIEW",
			Target:  "recipient:recipient-001",
			At:      now.Add(2 * time.Minute),
			IP:      "192.168.1.101",
			Details: "利用者情報を閲覧しました",
		},
	}

	for _, log := range auditLogs {
		err := auditRepo.Create(ctx, log)
		if err != nil {
			t.Fatalf("Create audit log error = %v", err)
		}
	}

	// Test GetByActorID for staff
	staffLogs, err := auditRepo.GetByActorID(ctx, staff.ID, 10, 0)
	if err != nil {
		t.Errorf("GetByActorID() error = %v", err)
	}

	if len(staffLogs) != 2 {
		t.Errorf("GetByActorID() returned %d logs, want 2", len(staffLogs))
	}

	// Verify all logs are for the correct actor
	for _, log := range staffLogs {
		if log.ActorID != staff.ID {
			t.Errorf("Log %s has ActorID %s, want %s", log.ID, log.ActorID, staff.ID)
		}
	}

	// Test GetByActorID for staff2
	staff2Logs, err := auditRepo.GetByActorID(ctx, staff2.ID, 10, 0)
	if err != nil {
		t.Errorf("GetByActorID() for staff2 error = %v", err)
	}

	if len(staff2Logs) != 1 {
		t.Errorf("GetByActorID() for staff2 returned %d logs, want 1", len(staff2Logs))
	}
}

func TestAuditLogRepository_GetByAction(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff := setupAuditLogTestData(t, db)

	auditRepo := NewAuditLogRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	// Create audit logs with different actions
	auditLogs := []*domain.AuditLog{
		{
			ID:      "audit-action-001",
			ActorID: staff.ID,
			Action:  "CREATE",
			Target:  "recipient:recipient-001",
			At:      now,
			IP:      "192.168.1.100",
			Details: "利用者を作成",
		},
		{
			ID:      "audit-action-002",
			ActorID: staff.ID,
			Action:  "CREATE",
			Target:  "recipient:recipient-002",
			At:      now.Add(time.Minute),
			IP:      "192.168.1.100",
			Details: "利用者を作成",
		},
		{
			ID:      "audit-action-003",
			ActorID: staff.ID,
			Action:  "UPDATE",
			Target:  "recipient:recipient-001",
			At:      now.Add(2 * time.Minute),
			IP:      "192.168.1.100",
			Details: "利用者を更新",
		},
		{
			ID:      "audit-action-004",
			ActorID: staff.ID,
			Action:  "DELETE",
			Target:  "recipient:recipient-002",
			At:      now.Add(3 * time.Minute),
			IP:      "192.168.1.100",
			Details: "利用者を削除",
		},
	}

	for _, log := range auditLogs {
		err := auditRepo.Create(ctx, log)
		if err != nil {
			t.Fatalf("Create audit log error = %v", err)
		}
	}

	// Test GetByAction for CREATE
	createLogs, err := auditRepo.GetByAction(ctx, "CREATE", 10, 0)
	if err != nil {
		t.Errorf("GetByAction() error = %v", err)
	}

	if len(createLogs) != 2 {
		t.Errorf("GetByAction('CREATE') returned %d logs, want 2", len(createLogs))
	}

	// Verify all logs have the correct action
	for _, log := range createLogs {
		if log.Action != "CREATE" {
			t.Errorf("Log %s has Action %s, want CREATE", log.ID, log.Action)
		}
	}

	// Test GetByAction for UPDATE
	updateLogs, err := auditRepo.GetByAction(ctx, "UPDATE", 10, 0)
	if err != nil {
		t.Errorf("GetByAction('UPDATE') error = %v", err)
	}

	if len(updateLogs) != 1 {
		t.Errorf("GetByAction('UPDATE') returned %d logs, want 1", len(updateLogs))
	}
}

func TestAuditLogRepository_GetByTarget(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff := setupAuditLogTestData(t, db)

	auditRepo := NewAuditLogRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	// Create audit logs with different targets
	auditLogs := []*domain.AuditLog{
		{
			ID:      "audit-target-001",
			ActorID: staff.ID,
			Action:  "CREATE",
			Target:  "recipient:recipient-001",
			At:      now,
			IP:      "192.168.1.100",
			Details: "利用者作成",
		},
		{
			ID:      "audit-target-002",
			ActorID: staff.ID,
			Action:  "UPDATE",
			Target:  "recipient:recipient-001",
			At:      now.Add(time.Minute),
			IP:      "192.168.1.100",
			Details: "利用者更新",
		},
		{
			ID:      "audit-target-003",
			ActorID: staff.ID,
			Action:  "CREATE",
			Target:  "certificate:cert-001",
			At:      now.Add(2 * time.Minute),
			IP:      "192.168.1.100",
			Details: "受給者証作成",
		},
	}

	for _, log := range auditLogs {
		err := auditRepo.Create(ctx, log)
		if err != nil {
			t.Fatalf("Create audit log error = %v", err)
		}
	}

	// Test GetByTarget for recipient:recipient-001
	recipientLogs, err := auditRepo.GetByTarget(ctx, "recipient:recipient-001", 10, 0)
	if err != nil {
		t.Errorf("GetByTarget() error = %v", err)
	}

	if len(recipientLogs) != 2 {
		t.Errorf("GetByTarget('recipient:recipient-001') returned %d logs, want 2", len(recipientLogs))
	}

	// Verify all logs have the correct target
	for _, log := range recipientLogs {
		if log.Target != "recipient:recipient-001" {
			t.Errorf("Log %s has Target %s, want recipient:recipient-001", log.ID, log.Target)
		}
	}
}

func TestAuditLogRepository_GetByTimeRange(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff := setupAuditLogTestData(t, db)

	auditRepo := NewAuditLogRepository(db)
	baseTime := time.Date(2024, 10, 1, 10, 0, 0, 0, time.UTC)

	// Create audit logs at different times
	auditLogs := []*domain.AuditLog{
		{
			ID:      "audit-time-001",
			ActorID: staff.ID,
			Action:  "LOGIN",
			Target:  "system",
			At:      baseTime,
			IP:      "192.168.1.100",
			Details: "ログイン",
		},
		{
			ID:      "audit-time-002",
			ActorID: staff.ID,
			Action:  "CREATE",
			Target:  "recipient:recipient-001",
			At:      baseTime.Add(30 * time.Minute),
			IP:      "192.168.1.100",
			Details: "利用者作成",
		},
		{
			ID:      "audit-time-003",
			ActorID: staff.ID,
			Action:  "UPDATE",
			Target:  "recipient:recipient-001",
			At:      baseTime.Add(2 * time.Hour),
			IP:      "192.168.1.100",
			Details: "利用者更新",
		},
		{
			ID:      "audit-time-004",
			ActorID: staff.ID,
			Action:  "LOGOUT",
			Target:  "system",
			At:      baseTime.Add(4 * time.Hour),
			IP:      "192.168.1.100",
			Details: "ログアウト",
		},
	}

	for _, log := range auditLogs {
		err := auditRepo.Create(ctx, log)
		if err != nil {
			t.Fatalf("Create audit log error = %v", err)
		}
	}

	// Test GetByTimeRange for first hour
	startTime := baseTime
	endTime := baseTime.Add(1 * time.Hour)

	timeLogs, err := auditRepo.GetByTimeRange(ctx, startTime, endTime, 10, 0)
	if err != nil {
		t.Errorf("GetByTimeRange() error = %v", err)
	}

	if len(timeLogs) != 2 {
		t.Errorf("GetByTimeRange() returned %d logs, want 2", len(timeLogs))
	}

	// Verify all logs are within the time range
	for _, log := range timeLogs {
		if log.At.Before(startTime) || log.At.After(endTime) {
			t.Errorf("Log %s at %v is outside time range %v - %v", log.ID, log.At, startTime, endTime)
		}
	}
}

func TestAuditLogRepository_Search(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff := setupAuditLogTestData(t, db)

	auditRepo := NewAuditLogRepository(db)
	baseTime := time.Date(2024, 10, 1, 10, 0, 0, 0, time.UTC)

	// Create audit logs for search testing
	auditLogs := []*domain.AuditLog{
		{
			ID:      "audit-search-001",
			ActorID: staff.ID,
			Action:  "CREATE",
			Target:  "recipient:recipient-001",
			At:      baseTime,
			IP:      "192.168.1.100",
			Details: "利用者作成",
		},
		{
			ID:      "audit-search-002",
			ActorID: staff.ID,
			Action:  "UPDATE",
			Target:  "recipient:recipient-001",
			At:      baseTime.Add(30 * time.Minute),
			IP:      "192.168.1.100",
			Details: "利用者更新",
		},
		{
			ID:      "audit-search-003",
			ActorID: staff.ID,
			Action:  "CREATE",
			Target:  "certificate:cert-001",
			At:      baseTime.Add(1 * time.Hour),
			IP:      "192.168.1.101",
			Details: "受給者証作成",
		},
	}

	for _, log := range auditLogs {
		err := auditRepo.Create(ctx, log)
		if err != nil {
			t.Fatalf("Create audit log error = %v", err)
		}
	}

	// Test complex search
	query := domain.AuditLogQuery{
		ActorID:   &staff.ID,
		Action:    stringPtr("CREATE"),
		StartTime: &baseTime,
		EndTime:   timePtr(baseTime.Add(2 * time.Hour)),
	}

	searchResults, err := auditRepo.Search(ctx, query, 10, 0)
	if err != nil {
		t.Errorf("Search() error = %v", err)
	}

	if len(searchResults) != 2 {
		t.Errorf("Search() returned %d logs, want 2", len(searchResults))
	}

	// Verify search results match criteria
	for _, log := range searchResults {
		if log.ActorID != staff.ID {
			t.Errorf("Search result %s has ActorID %s, want %s", log.ID, log.ActorID, staff.ID)
		}
		if log.Action != "CREATE" {
			t.Errorf("Search result %s has Action %s, want CREATE", log.ID, log.Action)
		}
		if log.At.Before(baseTime) || log.At.After(baseTime.Add(2*time.Hour)) {
			t.Errorf("Search result %s At %v is outside time range", log.ID, log.At)
		}
	}
}

func TestAuditLogRepository_List(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff := setupAuditLogTestData(t, db)

	auditRepo := NewAuditLogRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	// Create multiple audit logs
	auditLogs := []*domain.AuditLog{
		{
			ID:      "audit-list-001",
			ActorID: staff.ID,
			Action:  "LOGIN",
			Target:  "system",
			At:      now,
			IP:      "192.168.1.100",
			Details: "ログイン",
		},
		{
			ID:      "audit-list-002",
			ActorID: staff.ID,
			Action:  "CREATE",
			Target:  "recipient:recipient-001",
			At:      now.Add(time.Minute),
			IP:      "192.168.1.100",
			Details: "利用者作成",
		},
		{
			ID:      "audit-list-003",
			ActorID: staff.ID,
			Action:  "LOGOUT",
			Target:  "system",
			At:      now.Add(2 * time.Minute),
			IP:      "192.168.1.100",
			Details: "ログアウト",
		},
	}

	for _, log := range auditLogs {
		err := auditRepo.Create(ctx, log)
		if err != nil {
			t.Fatalf("Create audit log error = %v", err)
		}
	}

	// Test listing all
	allLogs, err := auditRepo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	if len(allLogs) < 3 {
		t.Errorf("List() returned %d logs, want at least 3", len(allLogs))
	}

	// Test pagination
	firstPage, err := auditRepo.List(ctx, 2, 0)
	if err != nil {
		t.Errorf("List() first page error = %v", err)
	}

	secondPage, err := auditRepo.List(ctx, 2, 2)
	if err != nil {
		t.Errorf("List() second page error = %v", err)
	}

	if len(firstPage) > 2 {
		t.Errorf("First page returned %d logs, want at most 2", len(firstPage))
	}

	// Verify no duplicates between pages
	for _, first := range firstPage {
		for _, second := range secondPage {
			if first.ID == second.ID {
				t.Errorf("Duplicate log %s found between pages", first.ID)
			}
		}
	}
}

func TestAuditLogRepository_Count(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	auditRepo := NewAuditLogRepository(db)
	ctx := context.Background()

	// Initial count should be 0
	count, err := auditRepo.Count(ctx)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	// Note: Full count test would require creating staff and audit logs
	// This is a basic test to verify the Count method works
}

func TestAuditLogRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	auditRepo := NewAuditLogRepository(db)
	ctx := context.Background()

	log, err := auditRepo.GetByID(ctx, "nonexistent-audit-log")
	if err != domain.ErrNotFound {
		t.Errorf("GetByID() error = %v, want %v", err, domain.ErrNotFound)
	}
	if log != nil {
		t.Error("GetByID() should return nil for nonexistent ID")
	}
}

// Helper functions for test data
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
