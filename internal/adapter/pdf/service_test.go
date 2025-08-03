package pdf

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"shien-system/internal/adapter/crypto"
	"shien-system/internal/domain"
)

func TestNewPDFService(t *testing.T) {
	cipher, err := crypto.NewFieldCipherWithKey(make([]byte, 32))
	require.NoError(t, err)

	service := NewPDFService("./fonts", cipher)
	assert.NotNil(t, service)
	assert.Equal(t, "./fonts", service.fontPath)
}

func TestPDFService_GenerateRecipientReport(t *testing.T) {
	cipher, err := crypto.NewFieldCipherWithKey(make([]byte, 32))
	require.NoError(t, err)

	service := NewPDFService("./fonts", cipher)

	// Create test recipient
	recipient := &domain.Recipient{
		ID:             "recipient-001",
		Name:           "テスト太郎",
		Kana:           "テストタロウ",
		Sex:            domain.SexMale,
		BirthDate:      time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		DisabilityName: "テスト障害",
		Address:        "東京都テスト区テスト町1-2-3",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Create test certificates
	certificates := []domain.BenefitCertificate{
		{
			ID:                     "cert-001",
			RecipientID:            "recipient-001",
			StartDate:              time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:                time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			Issuer:                 "テスト市役所",
			ServiceType:            "生活介護",
			MaxBenefitDaysPerMonth: 22,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		},
	}

	// Create test assignments
	assignments := []domain.StaffAssignment{
		{
			ID:          "assignment-001",
			RecipientID: "recipient-001",
			StaffID:     "staff-001",
			Role:        "担当者",
			AssignedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	ctx := context.Background()
	pdfBytes, err := service.GenerateRecipientReport(ctx, recipient, certificates, assignments)

	assert.NoError(t, err)
	assert.NotEmpty(t, pdfBytes)
	assert.True(t, len(pdfBytes) > 1000, "PDF should be reasonably sized")

	// Check PDF header
	assert.Equal(t, "%PDF", string(pdfBytes[:4]), "Should start with PDF header")
}

func TestPDFService_GenerateAuditReport(t *testing.T) {
	cipher, err := crypto.NewFieldCipherWithKey(make([]byte, 32))
	require.NoError(t, err)

	service := NewPDFService("./fonts", cipher)

	// Create test audit logs
	logs := []domain.AuditLog{
		{
			ID:      "log-001",
			ActorID: "staff-001",
			Action:  "CREATE",
			Target:  "recipient",
			At:      time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			IP:      "192.168.1.100",
			Details: "新規利用者を作成しました",
		},
		{
			ID:      "log-002",
			ActorID: "staff-001",
			Action:  "UPDATE",
			Target:  "recipient",
			At:      time.Date(2024, 1, 2, 14, 30, 0, 0, time.UTC),
			IP:      "192.168.1.100",
			Details: "利用者情報を更新しました",
		},
	}

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	ctx := context.Background()
	pdfBytes, err := service.GenerateAuditReport(ctx, logs, startDate, endDate)

	assert.NoError(t, err)
	assert.NotEmpty(t, pdfBytes)
	assert.True(t, len(pdfBytes) > 1000, "PDF should be reasonably sized")

	// Check PDF header
	assert.Equal(t, "%PDF", string(pdfBytes[:4]), "Should start with PDF header")
}

func TestPDFService_FormatSex(t *testing.T) {
	cipher, err := crypto.NewFieldCipherWithKey(make([]byte, 32))
	require.NoError(t, err)

	service := NewPDFService("./fonts", cipher)

	testCases := []struct {
		sex      domain.Sex
		expected string
	}{
		{domain.SexMale, "男性"},
		{domain.SexFemale, "女性"},
		{domain.SexOther, "その他"},
		{domain.SexNA, "未設定"},
		{domain.Sex("unknown"), "未設定"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.sex), func(t *testing.T) {
			result := service.formatSex(tc.sex)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFontManager(t *testing.T) {
	fm := NewFontManager()
	assert.NotNil(t, fm)

	// Test non-existent font
	hasFont := fm.HasFont("non-existent.ttf")
	assert.False(t, hasFont)

	_, err := fm.GetFontBytes("non-existent.ttf")
	assert.Error(t, err)
}

func BenchmarkPDFService_GenerateRecipientReport(b *testing.B) {
	cipher, err := crypto.NewFieldCipherWithKey(make([]byte, 32))
	require.NoError(b, err)

	service := NewPDFService("./fonts", cipher)

	recipient := &domain.Recipient{
		ID:             "recipient-001",
		Name:           "ベンチマーク太郎",
		Kana:           "ベンチマークタロウ",
		Sex:            domain.SexMale,
		BirthDate:      time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		DisabilityName: "テスト障害",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateRecipientReport(ctx, recipient, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
