package db

import (
	"context"
	"testing"
	"time"

	"shien-system/internal/domain"
)

func setupStaffAssignmentTestData(t *testing.T, db *Database) (context.Context, *domain.Staff, *domain.Recipient) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Create staff member
	staffRepo := NewStaffRepository(db)
	staff := &domain.Staff{
		ID:        "assignment-staff-001",
		Name:      "担当者テスト太郎",
		Role:      domain.RoleStaff,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := staffRepo.Create(ctx, staff)
	if err != nil {
		t.Fatalf("Create staff error = %v", err)
	}

	// Create recipient
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	recipient := &domain.Recipient{
		ID:               "assignment-recipient-001",
		Name:             "利用者テスト太郎",
		Sex:              domain.SexMale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  true,
		PublicAssistance: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err = recipientRepo.Create(ctx, recipient)
	if err != nil {
		t.Fatalf("Create recipient error = %v", err)
	}

	return ctx, staff, recipient
}

func TestStaffAssignmentRepository_Create(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff, recipient := setupStaffAssignmentTestData(t, db)

	assignmentRepo := NewStaffAssignmentRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	assignment := &domain.StaffAssignment{
		ID:           "assignment-create-001",
		RecipientID:  recipient.ID,
		StaffID:      staff.ID,
		Role:         "主担当",
		AssignedAt:   now,
		UnassignedAt: nil, // Currently active
	}

	err := assignmentRepo.Create(ctx, assignment)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	// Verify the assignment was created
	retrieved, err := assignmentRepo.GetByID(ctx, assignment.ID)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetByID() returned nil")
	}

	// Compare all fields
	if retrieved.ID != assignment.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, assignment.ID)
	}
	if retrieved.RecipientID != assignment.RecipientID {
		t.Errorf("RecipientID = %v, want %v", retrieved.RecipientID, assignment.RecipientID)
	}
	if retrieved.StaffID != assignment.StaffID {
		t.Errorf("StaffID = %v, want %v", retrieved.StaffID, assignment.StaffID)
	}
	if retrieved.Role != assignment.Role {
		t.Errorf("Role = %v, want %v", retrieved.Role, assignment.Role)
	}
	if !retrieved.AssignedAt.Equal(assignment.AssignedAt) {
		t.Errorf("AssignedAt = %v, want %v", retrieved.AssignedAt, assignment.AssignedAt)
	}
	if retrieved.UnassignedAt != nil {
		t.Errorf("UnassignedAt = %v, want nil", retrieved.UnassignedAt)
	}
}

func TestStaffAssignmentRepository_GetByRecipientID(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff, recipient := setupStaffAssignmentTestData(t, db)

	// Create additional staff
	staffRepo := NewStaffRepository(db)
	staff2 := &domain.Staff{
		ID:        "assignment-staff-002",
		Name:      "担当者テスト花子",
		Role:      domain.RoleStaff,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := staffRepo.Create(ctx, staff2)
	if err != nil {
		t.Fatalf("Create staff2 error = %v", err)
	}

	assignmentRepo := NewStaffAssignmentRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	// Create multiple assignments for the same recipient
	assignments := []*domain.StaffAssignment{
		{
			ID:           "assignment-recipient-001",
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "主担当",
			AssignedAt:   now,
			UnassignedAt: nil,
		},
		{
			ID:           "assignment-recipient-002",
			RecipientID:  recipient.ID,
			StaffID:      staff2.ID,
			Role:         "副担当",
			AssignedAt:   now.Add(time.Hour),
			UnassignedAt: nil,
		},
	}

	for _, assignment := range assignments {
		err := assignmentRepo.Create(ctx, assignment)
		if err != nil {
			t.Fatalf("Create assignment error = %v", err)
		}
	}

	// Test GetByRecipientID
	retrievedAssignments, err := assignmentRepo.GetByRecipientID(ctx, recipient.ID)
	if err != nil {
		t.Errorf("GetByRecipientID() error = %v", err)
	}

	if len(retrievedAssignments) != 2 {
		t.Errorf("GetByRecipientID() returned %d assignments, want 2", len(retrievedAssignments))
	}

	// Verify all assignments are for the correct recipient
	for _, assignment := range retrievedAssignments {
		if assignment.RecipientID != recipient.ID {
			t.Errorf("Assignment %s has RecipientID %s, want %s", assignment.ID, assignment.RecipientID, recipient.ID)
		}
	}
}

func TestStaffAssignmentRepository_GetActiveByRecipientID(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff, recipient := setupStaffAssignmentTestData(t, db)

	assignmentRepo := NewStaffAssignmentRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	pastTime := now.Add(-24 * time.Hour)

	// Create both active and inactive assignments
	assignments := []*domain.StaffAssignment{
		{
			ID:           "assignment-active-001",
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "主担当",
			AssignedAt:   now,
			UnassignedAt: nil, // Active
		},
		{
			ID:           "assignment-inactive-001",
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "旧主担当",
			AssignedAt:   pastTime,
			UnassignedAt: &now, // Inactive
		},
	}

	for _, assignment := range assignments {
		err := assignmentRepo.Create(ctx, assignment)
		if err != nil {
			t.Fatalf("Create assignment error = %v", err)
		}
	}

	// Test GetActiveByRecipientID
	activeAssignments, err := assignmentRepo.GetActiveByRecipientID(ctx, recipient.ID)
	if err != nil {
		t.Errorf("GetActiveByRecipientID() error = %v", err)
	}

	if len(activeAssignments) != 1 {
		t.Errorf("GetActiveByRecipientID() returned %d assignments, want 1", len(activeAssignments))
	}

	if len(activeAssignments) > 0 {
		if activeAssignments[0].ID != "assignment-active-001" {
			t.Errorf("GetActiveByRecipientID() returned assignment %s, want assignment-active-001", activeAssignments[0].ID)
		}
		if activeAssignments[0].UnassignedAt != nil {
			t.Error("GetActiveByRecipientID() should only return assignments with UnassignedAt = nil")
		}
	}
}

func TestStaffAssignmentRepository_GetByStaffID(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff, recipient := setupStaffAssignmentTestData(t, db)

	// Create additional recipient
	recipientRepo, err := NewRecipientRepository(db)
	require.NoError(t, err)
	recipient2 := &domain.Recipient{
		ID:               "assignment-recipient-002",
		Name:             "利用者テスト花子",
		Sex:              domain.SexFemale,
		BirthDate:        time.Date(1985, 5, 15, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  true,
		PublicAssistance: false,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	err := recipientRepo.Create(ctx, recipient2)
	if err != nil {
		t.Fatalf("Create recipient2 error = %v", err)
	}

	assignmentRepo := NewStaffAssignmentRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	// Create multiple assignments for the same staff
	assignments := []*domain.StaffAssignment{
		{
			ID:           "assignment-staff-001",
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "主担当",
			AssignedAt:   now,
			UnassignedAt: nil,
		},
		{
			ID:           "assignment-staff-002",
			RecipientID:  recipient2.ID,
			StaffID:      staff.ID,
			Role:         "主担当",
			AssignedAt:   now.Add(time.Hour),
			UnassignedAt: nil,
		},
	}

	for _, assignment := range assignments {
		err := assignmentRepo.Create(ctx, assignment)
		if err != nil {
			t.Fatalf("Create assignment error = %v", err)
		}
	}

	// Test GetByStaffID
	staffAssignments, err := assignmentRepo.GetByStaffID(ctx, staff.ID)
	if err != nil {
		t.Errorf("GetByStaffID() error = %v", err)
	}

	if len(staffAssignments) != 2 {
		t.Errorf("GetByStaffID() returned %d assignments, want 2", len(staffAssignments))
	}

	// Verify all assignments are for the correct staff
	for _, assignment := range staffAssignments {
		if assignment.StaffID != staff.ID {
			t.Errorf("Assignment %s has StaffID %s, want %s", assignment.ID, assignment.StaffID, staff.ID)
		}
	}
}

func TestStaffAssignmentRepository_GetActiveByStaffID(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff, recipient := setupStaffAssignmentTestData(t, db)

	assignmentRepo := NewStaffAssignmentRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	pastTime := now.Add(-24 * time.Hour)

	// Create both active and inactive assignments for the staff
	assignments := []*domain.StaffAssignment{
		{
			ID:           "assignment-staff-active-001",
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "主担当",
			AssignedAt:   now,
			UnassignedAt: nil, // Active
		},
		{
			ID:           "assignment-staff-inactive-001",
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "旧担当",
			AssignedAt:   pastTime,
			UnassignedAt: &now, // Inactive
		},
	}

	for _, assignment := range assignments {
		err := assignmentRepo.Create(ctx, assignment)
		if err != nil {
			t.Fatalf("Create assignment error = %v", err)
		}
	}

	// Test GetActiveByStaffID
	activeAssignments, err := assignmentRepo.GetActiveByStaffID(ctx, staff.ID)
	if err != nil {
		t.Errorf("GetActiveByStaffID() error = %v", err)
	}

	if len(activeAssignments) != 1 {
		t.Errorf("GetActiveByStaffID() returned %d assignments, want 1", len(activeAssignments))
	}

	if len(activeAssignments) > 0 {
		if activeAssignments[0].ID != "assignment-staff-active-001" {
			t.Errorf("GetActiveByStaffID() returned assignment %s, want assignment-staff-active-001", activeAssignments[0].ID)
		}
	}
}

func TestStaffAssignmentRepository_UnassignAll(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff, recipient := setupStaffAssignmentTestData(t, db)

	// Create additional staff
	staffRepo := NewStaffRepository(db)
	staff2 := &domain.Staff{
		ID:        "assignment-staff-003",
		Name:      "担当者テスト次郎",
		Role:      domain.RoleStaff,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := staffRepo.Create(ctx, staff2)
	if err != nil {
		t.Fatalf("Create staff2 error = %v", err)
	}

	assignmentRepo := NewStaffAssignmentRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	// Create multiple active assignments for the recipient
	assignments := []*domain.StaffAssignment{
		{
			ID:           "assignment-unassign-001",
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "主担当",
			AssignedAt:   now,
			UnassignedAt: nil,
		},
		{
			ID:           "assignment-unassign-002",
			RecipientID:  recipient.ID,
			StaffID:      staff2.ID,
			Role:         "副担当",
			AssignedAt:   now,
			UnassignedAt: nil,
		},
	}

	for _, assignment := range assignments {
		err := assignmentRepo.Create(ctx, assignment)
		if err != nil {
			t.Fatalf("Create assignment error = %v", err)
		}
	}

	// Verify assignments are active before unassign
	activeAssignments, err := assignmentRepo.GetActiveByRecipientID(ctx, recipient.ID)
	if err != nil {
		t.Fatalf("GetActiveByRecipientID() before unassign error = %v", err)
	}
	if len(activeAssignments) != 2 {
		t.Fatalf("Expected 2 active assignments before unassign, got %d", len(activeAssignments))
	}

	// Test UnassignAll
	unassignTime := now.Add(time.Hour)
	err = assignmentRepo.UnassignAll(ctx, recipient.ID, unassignTime)
	if err != nil {
		t.Errorf("UnassignAll() error = %v", err)
	}

	// Verify all assignments are now inactive
	activeAssignmentsAfter, err := assignmentRepo.GetActiveByRecipientID(ctx, recipient.ID)
	if err != nil {
		t.Errorf("GetActiveByRecipientID() after unassign error = %v", err)
	}
	if len(activeAssignmentsAfter) != 0 {
		t.Errorf("Expected 0 active assignments after unassign, got %d", len(activeAssignmentsAfter))
	}

	// Verify the unassigned_at field is set correctly
	allAssignments, err := assignmentRepo.GetByRecipientID(ctx, recipient.ID)
	if err != nil {
		t.Errorf("GetByRecipientID() after unassign error = %v", err)
	}

	for _, assignment := range allAssignments {
		if assignment.UnassignedAt == nil {
			t.Errorf("Assignment %s should have UnassignedAt set", assignment.ID)
		} else if !assignment.UnassignedAt.Equal(unassignTime) {
			t.Errorf("Assignment %s UnassignedAt = %v, want %v", assignment.ID, assignment.UnassignedAt, unassignTime)
		}
	}
}

func TestStaffAssignmentRepository_Update(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff, recipient := setupStaffAssignmentTestData(t, db)

	assignmentRepo := NewStaffAssignmentRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	// Create initial assignment
	assignment := &domain.StaffAssignment{
		ID:           "assignment-update-001",
		RecipientID:  recipient.ID,
		StaffID:      staff.ID,
		Role:         "主担当",
		AssignedAt:   now,
		UnassignedAt: nil,
	}

	err := assignmentRepo.Create(ctx, assignment)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the assignment
	unassignTime := now.Add(time.Hour)
	assignment.Role = "旧主担当"
	assignment.UnassignedAt = &unassignTime

	err = assignmentRepo.Update(ctx, assignment)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify the updates
	retrieved, err := assignmentRepo.GetByID(ctx, assignment.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.Role != "旧主担当" {
		t.Errorf("Role = %v, want %v", retrieved.Role, "旧主担当")
	}
	if retrieved.UnassignedAt == nil {
		t.Error("UnassignedAt should not be nil after update")
	} else if !retrieved.UnassignedAt.Equal(unassignTime) {
		t.Errorf("UnassignedAt = %v, want %v", *retrieved.UnassignedAt, unassignTime)
	}
}

func TestStaffAssignmentRepository_Delete(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff, recipient := setupStaffAssignmentTestData(t, db)

	assignmentRepo := NewStaffAssignmentRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	assignment := &domain.StaffAssignment{
		ID:           "assignment-delete-001",
		RecipientID:  recipient.ID,
		StaffID:      staff.ID,
		Role:         "削除テスト担当",
		AssignedAt:   now,
		UnassignedAt: nil,
	}

	err := assignmentRepo.Create(ctx, assignment)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Delete the assignment
	err = assignmentRepo.Delete(ctx, assignment.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify the assignment was deleted
	retrieved, err := assignmentRepo.GetByID(ctx, assignment.ID)
	if err != domain.ErrNotFound {
		t.Errorf("GetByID() after delete error = %v, want %v", err, domain.ErrNotFound)
	}
	if retrieved != nil {
		t.Error("GetByID() after delete should return nil")
	}
}

func TestStaffAssignmentRepository_List(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	ctx, staff, recipient := setupStaffAssignmentTestData(t, db)

	assignmentRepo := NewStaffAssignmentRepository(db)
	now := time.Now().UTC().Truncate(time.Second)

	// Create multiple assignments
	assignments := []*domain.StaffAssignment{
		{
			ID:           "assignment-list-001",
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "一覧テスト担当1",
			AssignedAt:   now,
			UnassignedAt: nil,
		},
		{
			ID:           "assignment-list-002",
			RecipientID:  recipient.ID,
			StaffID:      staff.ID,
			Role:         "一覧テスト担当2",
			AssignedAt:   now.Add(time.Minute),
			UnassignedAt: nil,
		},
	}

	for _, assignment := range assignments {
		err := assignmentRepo.Create(ctx, assignment)
		if err != nil {
			t.Fatalf("Create assignment error = %v", err)
		}
	}

	// Test listing all
	allAssignments, err := assignmentRepo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	if len(allAssignments) < 2 {
		t.Errorf("List() returned %d assignments, want at least 2", len(allAssignments))
	}

	// Test pagination
	firstPage, err := assignmentRepo.List(ctx, 1, 0)
	if err != nil {
		t.Errorf("List() first page error = %v", err)
	}

	secondPage, err := assignmentRepo.List(ctx, 1, 1)
	if err != nil {
		t.Errorf("List() second page error = %v", err)
	}

	if len(firstPage) > 1 {
		t.Errorf("First page returned %d assignments, want at most 1", len(firstPage))
	}

	if len(secondPage) > 1 {
		t.Errorf("Second page returned %d assignments, want at most 1", len(secondPage))
	}

	// Verify no duplicates between pages (if both pages have results)
	if len(firstPage) > 0 && len(secondPage) > 0 && firstPage[0].ID == secondPage[0].ID {
		t.Error("Duplicate assignment found between pages")
	}
}

func TestStaffAssignmentRepository_Count(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	assignmentRepo := NewStaffAssignmentRepository(db)
	ctx := context.Background()

	// Initial count should be 0
	count, err := assignmentRepo.Count(ctx)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	// Note: Full count test would require creating staff and recipient
	// This is a basic test to verify the Count method works
}

func TestStaffAssignmentRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	assignmentRepo := NewStaffAssignmentRepository(db)
	ctx := context.Background()

	assignment, err := assignmentRepo.GetByID(ctx, "nonexistent-assignment")
	if err != domain.ErrNotFound {
		t.Errorf("GetByID() error = %v, want %v", err, domain.ErrNotFound)
	}
	if assignment != nil {
		t.Error("GetByID() should return nil for nonexistent ID")
	}
}
