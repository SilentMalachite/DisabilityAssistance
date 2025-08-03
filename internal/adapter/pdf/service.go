package pdf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-pdf/fpdf"
	"shien-system/internal/adapter/crypto"
	"shien-system/internal/domain"
)

type PDFService struct {
	fontPath string
	cipher   *crypto.FieldCipher
}

func NewPDFService(fontPath string, cipher *crypto.FieldCipher) *PDFService {
	return &PDFService{
		fontPath: fontPath,
		cipher:   cipher,
	}
}

// GenerateRecipientReport generates a comprehensive recipient report
func (p *PDFService) GenerateRecipientReport(ctx context.Context, recipient *domain.Recipient, certificates []domain.BenefitCertificate, assignments []domain.StaffAssignment) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")

	// Use Arial as default font (Japanese fonts would be added in production)
	pdf.SetFont("Arial", "", 12)

	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "利用者情報報告書")
	pdf.Ln(15)

	// Basic information section
	p.addBasicInfo(pdf, recipient)

	// Certificates section
	if len(certificates) > 0 {
		p.addCertificatesSection(pdf, certificates)
	}

	// Staff assignments section
	if len(assignments) > 0 {
		p.addAssignmentsSection(pdf, assignments)
	}

	// Footer
	p.addFooter(pdf)

	// Generate PDF bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateAuditReport generates an audit log report
func (p *PDFService) GenerateAuditReport(ctx context.Context, logs []domain.AuditLog, startDate, endDate time.Time) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")

	// Use Arial as default font (Japanese fonts would be added in production)
	pdf.SetFont("Arial", "", 12)

	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	title := fmt.Sprintf("監査ログ報告書 (%s - %s)",
		startDate.Format("2006/01/02"), endDate.Format("2006/01/02"))
	pdf.Cell(0, 10, title)
	pdf.Ln(15)

	// Audit logs table
	p.addAuditLogsTable(pdf, logs)

	// Footer
	p.addFooter(pdf)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateStaffReport generates a PDF report for staff list
func (p *PDFService) GenerateStaffReport(ctx context.Context, staff []domain.Staff) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")

	// Use Arial as default font (Japanese fonts would be added in production)
	pdf.SetFont("Arial", "", 12)

	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	title := fmt.Sprintf("職員一覧 (%s)", time.Now().Format("2006年01月02日"))
	pdf.Cell(0, 10, title)
	pdf.Ln(15)

	// Staff table
	p.addStaffTable(pdf, staff)

	// Footer
	p.addFooter(pdf)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateCertificateReport generates a PDF report for certificate list
func (p *PDFService) GenerateCertificateReport(ctx context.Context, certificates []domain.BenefitCertificate, recipientMap map[domain.ID]*domain.Recipient) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")

	// Use Arial as default font (Japanese fonts would be added in production)
	pdf.SetFont("Arial", "", 12)

	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	title := fmt.Sprintf("受給者証一覧 (%s)", time.Now().Format("2006年01月02日"))
	pdf.Cell(0, 10, title)
	pdf.Ln(15)

	// Certificate table
	p.addCertificateTable(pdf, certificates, recipientMap)

	// Footer
	p.addFooter(pdf)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// addJapaneseFont adds Japanese font support to the PDF
func (p *PDFService) addJapaneseFont(pdf *fpdf.Fpdf) error {
	// Try to use embedded fonts first
	fontManager := NewFontManager()
	
	// Check for Noto Sans CJK JP font (recommended for Japanese)
	if fontManager.HasFont("NotoSansCJK-Regular.ttf") {
		fontBytes, err := fontManager.GetFontBytes("NotoSansCJK-Regular.ttf")
		if err == nil {
			// Write font to temporary file for fpdf
			tmpFile, err := os.CreateTemp("", "font-*.ttf")
			if err == nil {
				defer os.Remove(tmpFile.Name())
				defer tmpFile.Close()
				
				if _, err := tmpFile.Write(fontBytes); err == nil {
					pdf.AddUTF8Font("NotoSansCJK", "", tmpFile.Name())
					pdf.SetFont("NotoSansCJK", "", 12)
					return nil
				}
			}
		}
	}
	
	// Check for external font files
	fontFile := filepath.Join(p.fontPath, "NotoSansCJK-Regular.ttf")
	if _, err := os.Stat(fontFile); err == nil {
		pdf.AddUTF8Font("NotoSansCJK", "", fontFile)
		boldFontFile := filepath.Join(p.fontPath, "NotoSansCJK-Bold.ttf")
		if _, err := os.Stat(boldFontFile); err == nil {
			pdf.AddUTF8Font("NotoSansCJK", "B", boldFontFile)
		}
		pdf.SetFont("NotoSansCJK", "", 12)
		return nil
	}
	
	// Fallback to Arial for basic ASCII text
	pdf.SetFont("Arial", "", 12)
	return nil
}

// addBasicInfo adds basic recipient information to PDF
func (p *PDFService) addBasicInfo(pdf *fpdf.Fpdf, recipient *domain.Recipient) {
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "基本情報")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)

	// Name
	pdf.Cell(40, 6, "氏名:")
	pdf.Cell(0, 6, recipient.Name)
	pdf.Ln(8)

	// Kana
	if recipient.Kana != "" {
		pdf.Cell(40, 6, "フリガナ:")
		pdf.Cell(0, 6, recipient.Kana)
		pdf.Ln(8)
	}

	// Birth date
	pdf.Cell(40, 6, "生年月日:")
	pdf.Cell(0, 6, recipient.BirthDate.Format("2006年01月02日"))
	pdf.Ln(8)

	// Sex
	pdf.Cell(40, 6, "性別:")
	sexStr := p.formatSex(recipient.Sex)
	pdf.Cell(0, 6, sexStr)
	pdf.Ln(8)

	// Disability information
	if recipient.DisabilityName != "" {
		pdf.Cell(40, 6, "障害名:")
		pdf.Cell(0, 6, recipient.DisabilityName)
		pdf.Ln(8)
	}

	// Address
	if recipient.Address != "" {
		pdf.Cell(40, 6, "住所:")
		pdf.Cell(0, 6, recipient.Address)
		pdf.Ln(8)
	}

	pdf.Ln(8)
}

// addCertificatesSection adds benefit certificates section
func (p *PDFService) addCertificatesSection(pdf *fpdf.Fpdf, certificates []domain.BenefitCertificate) {
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "受給者証情報")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 9)

	// Table header
	pdf.Cell(40, 6, "サービス種別")
	pdf.Cell(30, 6, "開始日")
	pdf.Cell(30, 6, "終了日")
	pdf.Cell(30, 6, "発行者")
	pdf.Cell(30, 6, "月間日数")
	pdf.Ln(8)

	// Table content
	for _, cert := range certificates {
		pdf.Cell(40, 6, cert.ServiceType)
		pdf.Cell(30, 6, cert.StartDate.Format("2006/01/02"))
		pdf.Cell(30, 6, cert.EndDate.Format("2006/01/02"))
		pdf.Cell(30, 6, cert.Issuer)
		pdf.Cell(30, 6, fmt.Sprintf("%d日", cert.MaxBenefitDaysPerMonth))
		pdf.Ln(6)
	}

	pdf.Ln(8)
}

// addAssignmentsSection adds staff assignments section
func (p *PDFService) addAssignmentsSection(pdf *fpdf.Fpdf, assignments []domain.StaffAssignment) {
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "担当者割り当て")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 9)

	// Table header
	pdf.Cell(50, 6, "担当者ID")
	pdf.Cell(40, 6, "役割")
	pdf.Cell(40, 6, "開始日")
	pdf.Cell(40, 6, "終了日")
	pdf.Ln(8)

	// Table content
	for _, assignment := range assignments {
		pdf.Cell(50, 6, assignment.StaffID)
		pdf.Cell(40, 6, assignment.Role)
		pdf.Cell(40, 6, assignment.AssignedAt.Format("2006/01/02"))

		endDate := ""
		if assignment.UnassignedAt != nil {
			endDate = assignment.UnassignedAt.Format("2006/01/02")
		}
		pdf.Cell(40, 6, endDate)
		pdf.Ln(6)
	}

	pdf.Ln(8)
}

// addAuditLogsTable adds audit logs table to PDF
func (p *PDFService) addAuditLogsTable(pdf *fpdf.Fpdf, logs []domain.AuditLog) {
	pdf.SetFont("Arial", "", 8)

	// Table header
	pdf.Cell(30, 6, "日時")
	pdf.Cell(30, 6, "実行者")
	pdf.Cell(30, 6, "アクション")
	pdf.Cell(30, 6, "対象")
	pdf.Cell(70, 6, "詳細")
	pdf.Ln(8)

	// Table content
	for _, log := range logs {
		pdf.Cell(30, 6, log.At.Format("01/02 15:04"))
		pdf.Cell(30, 6, log.ActorID)
		pdf.Cell(30, 6, log.Action)
		pdf.Cell(30, 6, log.Target)

		// Truncate details if too long
		details := log.Details
		if len(details) > 30 {
			details = details[:27] + "..."
		}
		pdf.Cell(70, 6, details)
		pdf.Ln(6)
	}
}

// addStaffTable adds a staff table to the PDF
func (p *PDFService) addStaffTable(pdf *fpdf.Fpdf, staff []domain.Staff) {
	pdf.SetFont("Arial", "", 9)

	// Header
	pdf.Cell(60, 6, "職員名")
	pdf.Cell(40, 6, "ロール")
	pdf.Cell(50, 6, "作成日")
	pdf.Cell(40, 6, "更新日")
	pdf.Ln(8)

	// Data rows
	for _, s := range staff {
		pdf.Cell(60, 6, s.Name)
		pdf.Cell(40, 6, p.formatStaffRole(s.Role))
		pdf.Cell(50, 6, s.CreatedAt.Format("2006/01/02 15:04"))
		pdf.Cell(40, 6, s.UpdatedAt.Format("2006/01/02 15:04"))
		pdf.Ln(6)
	}

	pdf.Ln(8)
}

// formatStaffRole formats staff role for Japanese display
func (p *PDFService) formatStaffRole(role domain.StaffRole) string {
	switch role {
	case domain.RoleAdmin:
		return "管理者"
	case domain.RoleStaff:
		return "職員"
	case domain.RoleReadOnly:
		return "閲覧のみ"
	default:
		return string(role)
	}
}

// addCertificateTable adds a certificate table to the PDF
func (p *PDFService) addCertificateTable(pdf *fpdf.Fpdf, certificates []domain.BenefitCertificate, recipientMap map[domain.ID]*domain.Recipient) {
	pdf.SetFont("Arial", "", 8)

	// Header
	pdf.Cell(50, 6, "利用者名")
	pdf.Cell(30, 6, "サービス種別")
	pdf.Cell(25, 6, "開始日")
	pdf.Cell(25, 6, "終了日")
	pdf.Cell(30, 6, "発行者")
	pdf.Cell(20, 6, "月間日数")
	pdf.Cell(20, 6, "状態")
	pdf.Ln(8)

	// Data rows
	for _, cert := range certificates {
		// Get recipient name
		recipientName := "不明"
		if recipient, exists := recipientMap[cert.RecipientID]; exists && recipient != nil {
			recipientName = recipient.Name
		}

		// Calculate status
		status := p.calculateCertificateStatus(cert)

		pdf.Cell(50, 6, recipientName)
		pdf.Cell(30, 6, cert.ServiceType)
		pdf.Cell(25, 6, cert.StartDate.Format("2006/01/02"))
		pdf.Cell(25, 6, cert.EndDate.Format("2006/01/02"))
		pdf.Cell(30, 6, cert.Issuer)
		pdf.Cell(20, 6, fmt.Sprintf("%d日", cert.MaxBenefitDaysPerMonth))
		pdf.Cell(20, 6, status)
		pdf.Ln(6)
	}

	pdf.Ln(8)
}

// calculateCertificateStatus calculates the status of a certificate
func (p *PDFService) calculateCertificateStatus(cert domain.BenefitCertificate) string {
	now := time.Now()
	
	if cert.EndDate.Before(now) {
		return "期限切れ"
	}
	
	// Check if expiring within 30 days
	thirtyDaysFromNow := now.AddDate(0, 0, 30)
	if cert.EndDate.Before(thirtyDaysFromNow) {
		return "期限間近"
	}
	
	return "有効"
}

// addFooter adds footer to PDF
func (p *PDFService) addFooter(pdf *fpdf.Fpdf) {
	pdf.SetY(-15)
	pdf.SetFont("Arial", "", 8)
	pdf.Cell(0, 10, fmt.Sprintf("生成日時: %s", time.Now().Format("2006年01月02日 15:04:05")))
}

// formatSex formats sex enum to Japanese string
func (p *PDFService) formatSex(sex domain.Sex) string {
	switch sex {
	case domain.SexMale:
		return "男性"
	case domain.SexFemale:
		return "女性"
	case domain.SexOther:
		return "その他"
	default:
		return "未設定"
	}
}
