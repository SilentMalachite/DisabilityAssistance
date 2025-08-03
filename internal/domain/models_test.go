package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStaffRole_Constants(t *testing.T) {
	tests := []struct {
		name     string
		role     StaffRole
		expected string
	}{
		{"admin role", RoleAdmin, "admin"},
		{"staff role", RoleStaff, "staff"},
		{"readonly role", RoleReadOnly, "readonly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.expected {
				t.Errorf("StaffRole %v = %v, want %v", tt.name, string(tt.role), tt.expected)
			}
		})
	}
}

func TestSex_Constants(t *testing.T) {
	tests := []struct {
		name     string
		sex      Sex
		expected string
	}{
		{"female", SexFemale, "female"},
		{"male", SexMale, "male"},
		{"other", SexOther, "other"},
		{"not applicable", SexNA, "na"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.sex) != tt.expected {
				t.Errorf("Sex %v = %v, want %v", tt.name, string(tt.sex), tt.expected)
			}
		})
	}
}

func TestStaff_JSONSerialization(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	staff := Staff{
		ID:        "staff-001",
		Name:      "田中太郎",
		Role:      RoleAdmin,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test marshaling
	data, err := json.Marshal(staff)
	if err != nil {
		t.Fatalf("failed to marshal staff: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Staff
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal staff: %v", err)
	}

	// Compare fields
	if unmarshaled.ID != staff.ID {
		t.Errorf("ID mismatch: got %v, want %v", unmarshaled.ID, staff.ID)
	}
	if unmarshaled.Name != staff.Name {
		t.Errorf("Name mismatch: got %v, want %v", unmarshaled.Name, staff.Name)
	}
	if unmarshaled.Role != staff.Role {
		t.Errorf("Role mismatch: got %v, want %v", unmarshaled.Role, staff.Role)
	}
	if !unmarshaled.CreatedAt.Equal(staff.CreatedAt) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", unmarshaled.CreatedAt, staff.CreatedAt)
	}
	if !unmarshaled.UpdatedAt.Equal(staff.UpdatedAt) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", unmarshaled.UpdatedAt, staff.UpdatedAt)
	}
}

func TestRecipient_JSONSerialization(t *testing.T) {
	birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	admissionDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now().UTC().Truncate(time.Second)

	recipient := Recipient{
		ID:               "recipient-001",
		Name:             "山田花子",
		Kana:             "ヤマダハナコ",
		Sex:              SexFemale,
		BirthDate:        birthDate,
		DisabilityName:   "知的障害",
		HasDisabilityID:  true,
		Grade:            "B1",
		Address:          "東京都渋谷区1-1-1",
		Phone:            "03-1234-5678",
		Email:            "hanako@example.com",
		PublicAssistance: false,
		AdmissionDate:    &admissionDate,
		DischargeDate:    nil,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Test marshaling
	data, err := json.Marshal(recipient)
	if err != nil {
		t.Fatalf("failed to marshal recipient: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Recipient
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal recipient: %v", err)
	}

	// Compare required fields
	if unmarshaled.ID != recipient.ID {
		t.Errorf("ID mismatch: got %v, want %v", unmarshaled.ID, recipient.ID)
	}
	if unmarshaled.Name != recipient.Name {
		t.Errorf("Name mismatch: got %v, want %v", unmarshaled.Name, recipient.Name)
	}
	if unmarshaled.Sex != recipient.Sex {
		t.Errorf("Sex mismatch: got %v, want %v", unmarshaled.Sex, recipient.Sex)
	}
	if !unmarshaled.BirthDate.Equal(recipient.BirthDate) {
		t.Errorf("BirthDate mismatch: got %v, want %v", unmarshaled.BirthDate, recipient.BirthDate)
	}
	if unmarshaled.HasDisabilityID != recipient.HasDisabilityID {
		t.Errorf("HasDisabilityID mismatch: got %v, want %v", unmarshaled.HasDisabilityID, recipient.HasDisabilityID)
	}
	if unmarshaled.PublicAssistance != recipient.PublicAssistance {
		t.Errorf("PublicAssistance mismatch: got %v, want %v", unmarshaled.PublicAssistance, recipient.PublicAssistance)
	}

	// Test optional fields
	if (unmarshaled.AdmissionDate == nil) != (recipient.AdmissionDate == nil) {
		t.Errorf("AdmissionDate pointer mismatch")
	} else if unmarshaled.AdmissionDate != nil && !unmarshaled.AdmissionDate.Equal(*recipient.AdmissionDate) {
		t.Errorf("AdmissionDate value mismatch: got %v, want %v", *unmarshaled.AdmissionDate, *recipient.AdmissionDate)
	}

	if unmarshaled.DischargeDate != recipient.DischargeDate {
		t.Errorf("DischargeDate mismatch: got %v, want %v", unmarshaled.DischargeDate, recipient.DischargeDate)
	}
}

func TestBenefitCertificate_DateValidation(t *testing.T) {
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	cert := BenefitCertificate{
		ID:                     "cert-001",
		RecipientID:            "recipient-001",
		StartDate:              startDate,
		EndDate:                endDate,
		Issuer:                 "○○市",
		ServiceType:            "生活介護",
		MaxBenefitDaysPerMonth: 22,
		BenefitDetails:         "月額上限37,200円",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	// Test that end date is after start date
	if !cert.EndDate.After(cert.StartDate) {
		t.Error("EndDate should be after StartDate")
	}

	// Test JSON serialization
	data, err := json.Marshal(cert)
	if err != nil {
		t.Fatalf("failed to marshal certificate: %v", err)
	}

	var unmarshaled BenefitCertificate
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal certificate: %v", err)
	}

	if unmarshaled.MaxBenefitDaysPerMonth != cert.MaxBenefitDaysPerMonth {
		t.Errorf("MaxBenefitDaysPerMonth mismatch: got %v, want %v", unmarshaled.MaxBenefitDaysPerMonth, cert.MaxBenefitDaysPerMonth)
	}
}

func TestStaffAssignment_TimeTracking(t *testing.T) {
	assignedAt := time.Now().UTC().Truncate(time.Second)
	unassignedAt := assignedAt.Add(24 * time.Hour)

	assignment := StaffAssignment{
		ID:           "assignment-001",
		RecipientID:  "recipient-001",
		StaffID:      "staff-001",
		Role:         "主担当",
		AssignedAt:   assignedAt,
		UnassignedAt: &unassignedAt,
	}

	// Test JSON serialization
	data, err := json.Marshal(assignment)
	if err != nil {
		t.Fatalf("failed to marshal assignment: %v", err)
	}

	var unmarshaled StaffAssignment
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal assignment: %v", err)
	}

	if !unmarshaled.AssignedAt.Equal(assignment.AssignedAt) {
		t.Errorf("AssignedAt mismatch: got %v, want %v", unmarshaled.AssignedAt, assignment.AssignedAt)
	}

	if (unmarshaled.UnassignedAt == nil) != (assignment.UnassignedAt == nil) {
		t.Errorf("UnassignedAt pointer mismatch")
	} else if unmarshaled.UnassignedAt != nil && !unmarshaled.UnassignedAt.Equal(*assignment.UnassignedAt) {
		t.Errorf("UnassignedAt value mismatch: got %v, want %v", *unmarshaled.UnassignedAt, *assignment.UnassignedAt)
	}
}

func TestConsent_StatusTracking(t *testing.T) {
	obtainedAt := time.Now().UTC().Truncate(time.Second)
	revokedAt := obtainedAt.Add(30 * 24 * time.Hour) // 30 days later

	consent := Consent{
		ID:          "consent-001",
		RecipientID: "recipient-001",
		StaffID:     "staff-001",
		ConsentType: "個人情報利用",
		Content:     "サービス提供に必要な個人情報の利用について同意いたします",
		Method:      "書面",
		ObtainedAt:  obtainedAt,
		RevokedAt:   &revokedAt,
	}

	// Test JSON serialization
	data, err := json.Marshal(consent)
	if err != nil {
		t.Fatalf("failed to marshal consent: %v", err)
	}

	var unmarshaled Consent
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal consent: %v", err)
	}

	if unmarshaled.ConsentType != consent.ConsentType {
		t.Errorf("ConsentType mismatch: got %v, want %v", unmarshaled.ConsentType, consent.ConsentType)
	}

	if !unmarshaled.ObtainedAt.Equal(consent.ObtainedAt) {
		t.Errorf("ObtainedAt mismatch: got %v, want %v", unmarshaled.ObtainedAt, consent.ObtainedAt)
	}

	if (unmarshaled.RevokedAt == nil) != (consent.RevokedAt == nil) {
		t.Errorf("RevokedAt pointer mismatch")
	} else if unmarshaled.RevokedAt != nil && !unmarshaled.RevokedAt.Equal(*consent.RevokedAt) {
		t.Errorf("RevokedAt value mismatch: got %v, want %v", *unmarshaled.RevokedAt, *consent.RevokedAt)
	}
}

func TestAuditLog_DataIntegrity(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	log := AuditLog{
		ID:      "audit-001",
		ActorID: "staff-001",
		Action:  "CREATE",
		Target:  "recipient:recipient-001",
		At:      now,
		IP:      "127.0.0.1",
		Details: "利用者情報を新規作成",
	}

	// Test JSON serialization
	data, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("failed to marshal audit log: %v", err)
	}

	var unmarshaled AuditLog
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal audit log: %v", err)
	}

	if unmarshaled.Action != log.Action {
		t.Errorf("Action mismatch: got %v, want %v", unmarshaled.Action, log.Action)
	}

	if unmarshaled.Target != log.Target {
		t.Errorf("Target mismatch: got %v, want %v", unmarshaled.Target, log.Target)
	}

	if !unmarshaled.At.Equal(log.At) {
		t.Errorf("At mismatch: got %v, want %v", unmarshaled.At, log.At)
	}

	if unmarshaled.IP != log.IP {
		t.Errorf("IP mismatch: got %v, want %v", unmarshaled.IP, log.IP)
	}
}

func TestRecipient_OptionalFields(t *testing.T) {
	// Test recipient with minimal required fields
	recipient := Recipient{
		ID:               "minimal-001",
		Name:             "最小太郎",
		Sex:              SexMale,
		BirthDate:        time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC),
		HasDisabilityID:  false,
		PublicAssistance: true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Test JSON serialization with minimal fields
	data, err := json.Marshal(recipient)
	if err != nil {
		t.Fatalf("failed to marshal minimal recipient: %v", err)
	}

	var unmarshaled Recipient
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal minimal recipient: %v", err)
	}

	// Check that optional fields are handled correctly
	if unmarshaled.AdmissionDate != nil {
		t.Error("AdmissionDate should be nil for minimal recipient")
	}

	if unmarshaled.DischargeDate != nil {
		t.Error("DischargeDate should be nil for minimal recipient")
	}

	if unmarshaled.Kana != "" {
		t.Error("Kana should be empty for minimal recipient")
	}
}

func BenchmarkStaff_JSONMarshal(b *testing.B) {
	staff := Staff{
		ID:        "bench-staff-001",
		Name:      "ベンチマーク太郎",
		Role:      RoleStaff,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(staff)
		if err != nil {
			b.Fatalf("Marshal error: %v", err)
		}
	}
}

func BenchmarkRecipient_JSONMarshal(b *testing.B) {
	recipient := Recipient{
		ID:               "bench-recipient-001",
		Name:             "ベンチマーク花子",
		Kana:             "ベンチマークハナコ",
		Sex:              SexFemale,
		BirthDate:        time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		DisabilityName:   "身体障害",
		HasDisabilityID:  true,
		PublicAssistance: false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(recipient)
		if err != nil {
			b.Fatalf("Marshal error: %v", err)
		}
	}
}
