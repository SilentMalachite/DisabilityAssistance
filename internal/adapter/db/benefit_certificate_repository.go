package db

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"shien-system/internal/adapter/crypto"
	"shien-system/internal/domain"
)

// BenefitCertificateRepository implements domain.BenefitCertificateRepository
type BenefitCertificateRepository struct {
	db     *Database
	cipher *crypto.FieldCipher
}

// NewBenefitCertificateRepository creates a new benefit certificate repository
func NewBenefitCertificateRepository(db *Database) (*BenefitCertificateRepository, error) {
	cipher, err := crypto.NewFieldCipher()
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	return &BenefitCertificateRepository{
		db:     db,
		cipher: cipher,
	}, nil
}

// Create creates a new benefit certificate
func (r *BenefitCertificateRepository) Create(ctx context.Context, certificate *domain.BenefitCertificate) error {
	query := `
		INSERT INTO benefit_certificates (
			id, recipient_id, start_date, end_date, issuer_cipher, 
			service_type_cipher, max_benefit_days_per_month_cipher, 
			benefit_details_cipher, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Encrypt fields
	issuerCipher, err := r.cipher.Encrypt(certificate.Issuer)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt issuer", Err: err}
	}

	serviceTypeCipher, err := r.cipher.Encrypt(certificate.ServiceType)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt service_type", Err: err}
	}

	maxBenefitDaysCipher, err := r.cipher.Encrypt(strconv.Itoa(certificate.MaxBenefitDaysPerMonth))
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt max_benefit_days_per_month", Err: err}
	}

	benefitDetailsCipher, err := r.cipher.Encrypt(certificate.BenefitDetails)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt benefit_details", Err: err}
	}

	// Execute the query
	executor := r.getExecutor(ctx)
	_, err = executor.ExecContext(ctx, query,
		certificate.ID,
		certificate.RecipientID,
		certificate.StartDate.Format(time.RFC3339),
		certificate.EndDate.Format(time.RFC3339),
		issuerCipher,
		serviceTypeCipher,
		maxBenefitDaysCipher,
		benefitDetailsCipher,
		certificate.CreatedAt.Format(time.RFC3339),
		certificate.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return &domain.RepositoryError{Op: "create benefit certificate", Err: err}
	}

	return nil
}

// GetByID retrieves a benefit certificate by ID
func (r *BenefitCertificateRepository) GetByID(ctx context.Context, id domain.ID) (*domain.BenefitCertificate, error) {
	query := `
		SELECT id, recipient_id, start_date, end_date, issuer_cipher, 
			   service_type_cipher, max_benefit_days_per_month_cipher, 
			   benefit_details_cipher, created_at, updated_at
		FROM benefit_certificates 
		WHERE id = ?`

	executor := r.getExecutor(ctx)
	row := executor.QueryRowContext(ctx, query, id)

	return r.scanBenefitCertificate(row)
}

// Update updates an existing benefit certificate
func (r *BenefitCertificateRepository) Update(ctx context.Context, certificate *domain.BenefitCertificate) error {
	query := `
		UPDATE benefit_certificates 
		SET recipient_id = ?, start_date = ?, end_date = ?, issuer_cipher = ?, 
			service_type_cipher = ?, max_benefit_days_per_month_cipher = ?, 
			benefit_details_cipher = ?, updated_at = ?
		WHERE id = ?`

	// Encrypt fields
	issuerCipher, err := r.cipher.Encrypt(certificate.Issuer)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt issuer", Err: err}
	}

	serviceTypeCipher, err := r.cipher.Encrypt(certificate.ServiceType)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt service_type", Err: err}
	}

	maxBenefitDaysCipher, err := r.cipher.Encrypt(strconv.Itoa(certificate.MaxBenefitDaysPerMonth))
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt max_benefit_days_per_month", Err: err}
	}

	benefitDetailsCipher, err := r.cipher.Encrypt(certificate.BenefitDetails)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt benefit_details", Err: err}
	}

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query,
		certificate.RecipientID,
		certificate.StartDate.Format(time.RFC3339),
		certificate.EndDate.Format(time.RFC3339),
		issuerCipher,
		serviceTypeCipher,
		maxBenefitDaysCipher,
		benefitDetailsCipher,
		certificate.UpdatedAt.Format(time.RFC3339),
		certificate.ID,
	)

	if err != nil {
		return &domain.RepositoryError{Op: "update benefit certificate", Err: err}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &domain.RepositoryError{Op: "check rows affected", Err: err}
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete deletes a benefit certificate by ID
func (r *BenefitCertificateRepository) Delete(ctx context.Context, id domain.ID) error {
	query := `DELETE FROM benefit_certificates WHERE id = ?`

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return &domain.RepositoryError{Op: "delete benefit certificate", Err: err}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &domain.RepositoryError{Op: "check rows affected", Err: err}
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetByRecipientID retrieves all benefit certificates for a recipient
func (r *BenefitCertificateRepository) GetByRecipientID(ctx context.Context, recipientID domain.ID) ([]*domain.BenefitCertificate, error) {
	query := `
		SELECT id, recipient_id, start_date, end_date, issuer_cipher, 
			   service_type_cipher, max_benefit_days_per_month_cipher, 
			   benefit_details_cipher, created_at, updated_at
		FROM benefit_certificates 
		WHERE recipient_id = ?
		ORDER BY start_date DESC`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, recipientID)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get certificates by recipient", Err: err}
	}
	defer rows.Close()

	var certificates []*domain.BenefitCertificate
	for rows.Next() {
		certificate, err := r.scanBenefitCertificate(rows)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, certificate)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return certificates, nil
}

// GetExpiringSoon retrieves certificates expiring within the specified duration
func (r *BenefitCertificateRepository) GetExpiringSoon(ctx context.Context, within time.Duration) ([]*domain.BenefitCertificate, error) {
	cutoffTime := time.Now().Add(within)

	query := `
		SELECT id, recipient_id, start_date, end_date, issuer_cipher, 
			   service_type_cipher, max_benefit_days_per_month_cipher, 
			   benefit_details_cipher, created_at, updated_at
		FROM benefit_certificates 
		WHERE end_date <= ?
		ORDER BY end_date ASC`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, cutoffTime.Format(time.RFC3339))
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get expiring certificates", Err: err}
	}
	defer rows.Close()

	var certificates []*domain.BenefitCertificate
	for rows.Next() {
		certificate, err := r.scanBenefitCertificate(rows)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, certificate)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return certificates, nil
}

// GetActiveByRecipientID retrieves the active certificate for a recipient at a specific date
func (r *BenefitCertificateRepository) GetActiveByRecipientID(ctx context.Context, recipientID domain.ID, asOf time.Time) (*domain.BenefitCertificate, error) {
	query := `
		SELECT id, recipient_id, start_date, end_date, issuer_cipher, 
			   service_type_cipher, max_benefit_days_per_month_cipher, 
			   benefit_details_cipher, created_at, updated_at
		FROM benefit_certificates 
		WHERE recipient_id = ? AND start_date <= ? AND end_date >= ?
		ORDER BY start_date DESC
		LIMIT 1`

	executor := r.getExecutor(ctx)
	row := executor.QueryRowContext(ctx, query, recipientID, asOf.Format(time.RFC3339), asOf.Format(time.RFC3339))

	return r.scanBenefitCertificate(row)
}

// List retrieves benefit certificates with pagination
func (r *BenefitCertificateRepository) List(ctx context.Context, limit, offset int) ([]*domain.BenefitCertificate, error) {
	query := `
		SELECT id, recipient_id, start_date, end_date, issuer_cipher, 
			   service_type_cipher, max_benefit_days_per_month_cipher, 
			   benefit_details_cipher, created_at, updated_at
		FROM benefit_certificates 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "list certificates", Err: err}
	}
	defer rows.Close()

	var certificates []*domain.BenefitCertificate
	for rows.Next() {
		certificate, err := r.scanBenefitCertificate(rows)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, certificate)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return certificates, nil
}

// Count returns the total number of benefit certificates
func (r *BenefitCertificateRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM benefit_certificates`

	executor := r.getExecutor(ctx)
	var count int
	err := executor.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, &domain.RepositoryError{Op: "count certificates", Err: err}
	}

	return count, nil
}

// getExecutor returns either a transaction or the database connection
func (r *BenefitCertificateRepository) getExecutor(ctx context.Context) executor {
	if tx := ctx.Value("tx"); tx != nil {
		return tx.(*sql.Tx)
	}
	return r.db.DB()
}

// scanBenefitCertificate scans a benefit certificate from a database row
func (r *BenefitCertificateRepository) scanBenefitCertificate(row scanner) (*domain.BenefitCertificate, error) {
	var certificate domain.BenefitCertificate
	var issuerCipher, serviceTypeCipher, maxBenefitDaysCipher, benefitDetailsCipher []byte
	var startDateStr, endDateStr, createdAtStr, updatedAtStr string

	err := row.Scan(
		&certificate.ID,
		&certificate.RecipientID,
		&startDateStr,
		&endDateStr,
		&issuerCipher,
		&serviceTypeCipher,
		&maxBenefitDaysCipher,
		&benefitDetailsCipher,
		&createdAtStr,
		&updatedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, &domain.RepositoryError{Op: "scan certificate", Err: err}
	}

	// Parse dates
	certificate.StartDate, err = time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse start_date", Err: err}
	}

	certificate.EndDate, err = time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse end_date", Err: err}
	}

	certificate.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse created_at", Err: err}
	}

	certificate.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse updated_at", Err: err}
	}

	// Decrypt fields
	certificate.Issuer, err = r.cipher.Decrypt(issuerCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt issuer", Err: err}
	}

	certificate.ServiceType, err = r.cipher.Decrypt(serviceTypeCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt service_type", Err: err}
	}

	maxBenefitDaysStr, err := r.cipher.Decrypt(maxBenefitDaysCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt max_benefit_days_per_month", Err: err}
	}
	certificate.MaxBenefitDaysPerMonth, err = strconv.Atoi(maxBenefitDaysStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse max_benefit_days_per_month", Err: err}
	}

	certificate.BenefitDetails, err = r.cipher.Decrypt(benefitDetailsCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt benefit_details", Err: err}
	}

	return &certificate, nil
}
