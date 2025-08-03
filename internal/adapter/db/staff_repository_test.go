package db

import (
	"context"
	"testing"
	"time"

	"shien-system/internal/domain"
)

func TestStaffRepository_Create(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	staff := &domain.Staff{
		ID:        "staff-create-001",
		Name:      "作成テスト太郎",
		Role:      domain.RoleStaff,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := staffRepo.Create(ctx, staff)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	// Verify the staff was created by retrieving it
	retrieved, err := staffRepo.GetByID(ctx, staff.ID)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetByID() returned nil")
	}

	// Compare all fields
	if retrieved.ID != staff.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, staff.ID)
	}
	if retrieved.Name != staff.Name {
		t.Errorf("Name = %v, want %v", retrieved.Name, staff.Name)
	}
	if retrieved.Role != staff.Role {
		t.Errorf("Role = %v, want %v", retrieved.Role, staff.Role)
	}
	if !retrieved.CreatedAt.Equal(staff.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", retrieved.CreatedAt, staff.CreatedAt)
	}
	if !retrieved.UpdatedAt.Equal(staff.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", retrieved.UpdatedAt, staff.UpdatedAt)
	}
}

func TestStaffRepository_CreateDuplicate(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	staff := &domain.Staff{
		ID:        "staff-duplicate-001",
		Name:      "重複テスト太郎",
		Role:      domain.RoleStaff,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// First creation should succeed
	err := staffRepo.Create(ctx, staff)
	if err != nil {
		t.Errorf("First Create() error = %v", err)
	}

	// Second creation with same ID should fail
	err = staffRepo.Create(ctx, staff)
	if err == nil {
		t.Error("Second Create() should return error for duplicate ID")
	}
}

func TestStaffRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()

	staff, err := staffRepo.GetByID(ctx, "nonexistent-staff-id")
	if err != domain.ErrNotFound {
		t.Errorf("GetByID() error = %v, want %v", err, domain.ErrNotFound)
	}
	if staff != nil {
		t.Error("GetByID() should return nil for nonexistent ID")
	}
}

func TestStaffRepository_Update(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Create initial staff
	staff := &domain.Staff{
		ID:        "staff-update-001",
		Name:      "更新前太郎",
		Role:      domain.RoleStaff,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := staffRepo.Create(ctx, staff)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the staff
	updatedTime := now.Add(time.Hour)
	staff.Name = "更新後太郎"
	staff.Role = domain.RoleAdmin
	staff.UpdatedAt = updatedTime

	err = staffRepo.Update(ctx, staff)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify the updates
	retrieved, err := staffRepo.GetByID(ctx, staff.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.Name != "更新後太郎" {
		t.Errorf("Name = %v, want %v", retrieved.Name, "更新後太郎")
	}
	if retrieved.Role != domain.RoleAdmin {
		t.Errorf("Role = %v, want %v", retrieved.Role, domain.RoleAdmin)
	}
	if !retrieved.UpdatedAt.Equal(updatedTime) {
		t.Errorf("UpdatedAt = %v, want %v", retrieved.UpdatedAt, updatedTime)
	}
	// CreatedAt should remain unchanged
	if !retrieved.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", retrieved.CreatedAt, now)
	}
}

func TestStaffRepository_Delete(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	staff := &domain.Staff{
		ID:        "staff-delete-001",
		Name:      "削除テスト太郎",
		Role:      domain.RoleStaff,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := staffRepo.Create(ctx, staff)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Delete the staff
	err = staffRepo.Delete(ctx, staff.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify the staff was deleted
	retrieved, err := staffRepo.GetByID(ctx, staff.ID)
	if err != domain.ErrNotFound {
		t.Errorf("GetByID() after delete error = %v, want %v", err, domain.ErrNotFound)
	}
	if retrieved != nil {
		t.Error("GetByID() after delete should return nil")
	}
}

func TestStaffRepository_List(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Create multiple staff members
	staffMembers := []*domain.Staff{
		{
			ID:        "staff-list-001",
			Name:      "一覧テスト太郎",
			Role:      domain.RoleStaff,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "staff-list-002",
			Name:      "一覧テスト花子",
			Role:      domain.RoleAdmin,
			CreatedAt: now.Add(time.Minute),
			UpdatedAt: now.Add(time.Minute),
		},
		{
			ID:        "staff-list-003",
			Name:      "一覧テスト次郎",
			Role:      domain.RoleReadOnly,
			CreatedAt: now.Add(2 * time.Minute),
			UpdatedAt: now.Add(2 * time.Minute),
		},
	}

	for _, staff := range staffMembers {
		err := staffRepo.Create(ctx, staff)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test listing all (should include the initial admin user from migration)
	allStaff, err := staffRepo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	// Should have at least our 3 staff members + the initial admin
	if len(allStaff) < 4 {
		t.Errorf("List() returned %d staff, want at least 4", len(allStaff))
	}

	// Test pagination
	firstPage, err := staffRepo.List(ctx, 2, 0)
	if err != nil {
		t.Errorf("List() first page error = %v", err)
	}

	secondPage, err := staffRepo.List(ctx, 2, 2)
	if err != nil {
		t.Errorf("List() second page error = %v", err)
	}

	if len(firstPage) > 2 {
		t.Errorf("First page returned %d staff, want at most 2", len(firstPage))
	}

	// Verify no duplicates between pages
	for _, first := range firstPage {
		for _, second := range secondPage {
			if first.ID == second.ID {
				t.Errorf("Duplicate staff %s found between pages", first.ID)
			}
		}
	}
}

func TestStaffRepository_GetByRole(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Create staff with different roles
	staffMembers := []*domain.Staff{
		{
			ID:        "staff-role-admin-001",
			Name:      "管理者太郎",
			Role:      domain.RoleAdmin,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "staff-role-admin-002",
			Name:      "管理者花子",
			Role:      domain.RoleAdmin,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "staff-role-staff-001",
			Name:      "職員太郎",
			Role:      domain.RoleStaff,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "staff-role-readonly-001",
			Name:      "閲覧者太郎",
			Role:      domain.RoleReadOnly,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, staff := range staffMembers {
		err := staffRepo.Create(ctx, staff)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test getting admins (should include initial admin + our 2 test admins)
	admins, err := staffRepo.GetByRole(ctx, domain.RoleAdmin)
	if err != nil {
		t.Errorf("GetByRole(admin) error = %v", err)
	}

	if len(admins) < 3 {
		t.Errorf("GetByRole(admin) returned %d staff, want at least 3", len(admins))
	}

	// Verify all returned staff are admins
	for _, admin := range admins {
		if admin.Role != domain.RoleAdmin {
			t.Errorf("GetByRole(admin) returned staff with role %v", admin.Role)
		}
	}

	// Test getting regular staff
	staffList, err := staffRepo.GetByRole(ctx, domain.RoleStaff)
	if err != nil {
		t.Errorf("GetByRole(staff) error = %v", err)
	}

	if len(staffList) != 1 {
		t.Errorf("GetByRole(staff) returned %d staff, want 1", len(staffList))
	}

	// Test getting readonly staff
	readOnlyList, err := staffRepo.GetByRole(ctx, domain.RoleReadOnly)
	if err != nil {
		t.Errorf("GetByRole(readonly) error = %v", err)
	}

	if len(readOnlyList) != 1 {
		t.Errorf("GetByRole(readonly) returned %d staff, want 1", len(readOnlyList))
	}
}

func TestStaffRepository_GetByName(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Create staff with searchable names
	staffMembers := []*domain.Staff{
		{
			ID:        "staff-search-001",
			Name:      "検索テスト太郎",
			Role:      domain.RoleStaff,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "staff-search-002",
			Name:      "検索テスト花子",
			Role:      domain.RoleStaff,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "staff-search-003",
			Name:      "別の名前太郎",
			Role:      domain.RoleStaff,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, staff := range staffMembers {
		err := staffRepo.Create(ctx, staff)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test partial name search
	searchResults, err := staffRepo.GetByName(ctx, "検索テスト")
	if err != nil {
		t.Errorf("GetByName() error = %v", err)
	}

	if len(searchResults) != 2 {
		t.Errorf("GetByName('検索テスト') returned %d staff, want 2", len(searchResults))
	}

	// Verify search results contain expected names
	found := make(map[string]bool)
	for _, staff := range searchResults {
		found[staff.Name] = true
	}

	if !found["検索テスト太郎"] {
		t.Error("GetByName() did not return '検索テスト太郎'")
	}
	if !found["検索テスト花子"] {
		t.Error("GetByName() did not return '検索テスト花子'")
	}
	if found["別の名前太郎"] {
		t.Error("GetByName() should not return '別の名前太郎'")
	}

	// Test exact name search
	exactResults, err := staffRepo.GetByName(ctx, "別の名前太郎")
	if err != nil {
		t.Errorf("GetByName() exact search error = %v", err)
	}

	if len(exactResults) != 1 {
		t.Errorf("GetByName('別の名前太郎') returned %d staff, want 1", len(exactResults))
	}

	if len(exactResults) > 0 && exactResults[0].Name != "別の名前太郎" {
		t.Errorf("GetByName() exact search returned %v, want '別の名前太郎'", exactResults[0].Name)
	}
}

func TestStaffRepository_Count(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()

	// Initial count should be 1 (the admin user from migration)
	initialCount, err := staffRepo.Count(ctx)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if initialCount != 1 {
		t.Errorf("Initial count = %d, want 1", initialCount)
	}

	// Create staff and verify count increases
	now := time.Now().UTC().Truncate(time.Second)

	for i := 1; i <= 3; i++ {
		staff := &domain.Staff{
			ID:        domain.ID("count-staff-" + string(rune('0'+i))),
			Name:      "カウントテスト太郎" + string(rune('0'+i)),
			Role:      domain.RoleStaff,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := staffRepo.Create(ctx, staff)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		count, err := staffRepo.Count(ctx)
		if err != nil {
			t.Errorf("Count() error = %v", err)
		}
		expectedCount := initialCount + i
		if count != expectedCount {
			t.Errorf("Count after %d creates = %d, want %d", i, count, expectedCount)
		}
	}
}

func TestStaffRepository_UpdateNonExistent(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	staff := &domain.Staff{
		ID:        "nonexistent-staff",
		Name:      "存在しない職員",
		Role:      domain.RoleStaff,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := staffRepo.Update(ctx, staff)
	if err != domain.ErrNotFound {
		t.Errorf("Update() nonexistent staff error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestStaffRepository_DeleteNonExistent(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	staffRepo := NewStaffRepository(db)

	ctx := context.Background()

	err := staffRepo.Delete(ctx, "nonexistent-staff")
	if err != domain.ErrNotFound {
		t.Errorf("Delete() nonexistent staff error = %v, want %v", err, domain.ErrNotFound)
	}
}
