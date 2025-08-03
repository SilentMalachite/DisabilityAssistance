package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"strconv"
	"time"

	"shien-system/internal/adapter/crypto"
	"shien-system/internal/domain"
)

// RecipientRepository implements domain.RecipientRepository
type RecipientRepository struct {
	db     *Database
	cipher *crypto.FieldCipher
}

// NewRecipientRepository creates a new recipient repository
func NewRecipientRepository(db *Database) (*RecipientRepository, error) {
	cipher, err := crypto.NewFieldCipher()
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	return &RecipientRepository{
		db:     db,
		cipher: cipher,
	}, nil
}

// Create creates a new recipient
func (r *RecipientRepository) Create(ctx context.Context, recipient *domain.Recipient) error {
	query := `
		INSERT INTO recipients (
			id, name_cipher, kana_cipher, sex_cipher, birth_date_cipher,
			disability_name_cipher, has_disability_id_cipher, grade_cipher,
			address_cipher, phone_cipher, email_cipher, public_assistance_cipher,
			admission_date, discharge_date, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Encrypt fields
	nameCipher, err := r.cipher.Encrypt(recipient.Name)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt name", Err: err}
	}

	kanaCipher, err := r.cipher.Encrypt(recipient.Kana)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt kana", Err: err}
	}

	sexCipher, err := r.cipher.Encrypt(string(recipient.Sex))
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt sex", Err: err}
	}

	birthDateCipher, err := r.cipher.Encrypt(recipient.BirthDate.Format(time.RFC3339))
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt birth_date", Err: err}
	}

	disabilityNameCipher, err := r.cipher.Encrypt(recipient.DisabilityName)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt disability_name", Err: err}
	}

	hasDisabilityIDCipher, err := r.cipher.Encrypt(strconv.FormatBool(recipient.HasDisabilityID))
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt has_disability_id", Err: err}
	}

	gradeCipher, err := r.cipher.Encrypt(recipient.Grade)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt grade", Err: err}
	}

	addressCipher, err := r.cipher.Encrypt(recipient.Address)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt address", Err: err}
	}

	phoneCipher, err := r.cipher.Encrypt(recipient.Phone)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt phone", Err: err}
	}

	emailCipher, err := r.cipher.Encrypt(recipient.Email)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt email", Err: err}
	}

	publicAssistanceCipher, err := r.cipher.Encrypt(strconv.FormatBool(recipient.PublicAssistance))
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt public_assistance", Err: err}
	}

	// Handle optional dates
	var admissionDate, dischargeDate *string
	if recipient.AdmissionDate != nil {
		dateStr := recipient.AdmissionDate.Format(time.RFC3339)
		admissionDate = &dateStr
	}
	if recipient.DischargeDate != nil {
		dateStr := recipient.DischargeDate.Format(time.RFC3339)
		dischargeDate = &dateStr
	}

	// Execute the query
	executor := r.getExecutor(ctx)
	_, err = executor.ExecContext(ctx, query,
		recipient.ID, nameCipher, kanaCipher, sexCipher, birthDateCipher,
		disabilityNameCipher, hasDisabilityIDCipher, gradeCipher,
		addressCipher, phoneCipher, emailCipher, publicAssistanceCipher,
		admissionDate, dischargeDate,
		recipient.CreatedAt.Format(time.RFC3339),
		recipient.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return &domain.RepositoryError{Op: "create recipient", Err: err}
	}

	// Clear sensitive encrypted data from memory
	crypto.ClearBytes(nameCipher)
	crypto.ClearBytes(kanaCipher)
	crypto.ClearBytes(sexCipher)
	crypto.ClearBytes(birthDateCipher)
	crypto.ClearBytes(disabilityNameCipher)
	crypto.ClearBytes(hasDisabilityIDCipher)
	crypto.ClearBytes(gradeCipher)
	crypto.ClearBytes(addressCipher)
	crypto.ClearBytes(phoneCipher)
	crypto.ClearBytes(emailCipher)
	crypto.ClearBytes(publicAssistanceCipher)

	return nil
}

// GetByID retrieves a recipient by ID
func (r *RecipientRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Recipient, error) {
	query := `
		SELECT id, name_cipher, kana_cipher, sex_cipher, birth_date_cipher,
			   disability_name_cipher, has_disability_id_cipher, grade_cipher,
			   address_cipher, phone_cipher, email_cipher, public_assistance_cipher,
			   admission_date, discharge_date, created_at, updated_at
		FROM recipients 
		WHERE id = ?`

	executor := r.getExecutor(ctx)
	row := executor.QueryRowContext(ctx, query, id)

	return r.scanRecipient(row)
}

// Update updates an existing recipient
func (r *RecipientRepository) Update(ctx context.Context, recipient *domain.Recipient) error {
	query := `
		UPDATE recipients 
		SET name_cipher = ?, kana_cipher = ?, sex_cipher = ?, birth_date_cipher = ?,
			disability_name_cipher = ?, has_disability_id_cipher = ?, grade_cipher = ?,
			address_cipher = ?, phone_cipher = ?, email_cipher = ?, public_assistance_cipher = ?,
			admission_date = ?, discharge_date = ?, updated_at = ?
		WHERE id = ?`

	// Encrypt fields
	nameCipher, err := r.cipher.Encrypt(recipient.Name)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt name", Err: err}
	}

	kanaCipher, err := r.cipher.Encrypt(recipient.Kana)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt kana", Err: err}
	}

	sexCipher, err := r.cipher.Encrypt(string(recipient.Sex))
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt sex", Err: err}
	}

	birthDateCipher, err := r.cipher.Encrypt(recipient.BirthDate.Format(time.RFC3339))
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt birth_date", Err: err}
	}

	disabilityNameCipher, err := r.cipher.Encrypt(recipient.DisabilityName)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt disability_name", Err: err}
	}

	hasDisabilityIDCipher, err := r.cipher.Encrypt(strconv.FormatBool(recipient.HasDisabilityID))
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt has_disability_id", Err: err}
	}

	gradeCipher, err := r.cipher.Encrypt(recipient.Grade)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt grade", Err: err}
	}

	addressCipher, err := r.cipher.Encrypt(recipient.Address)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt address", Err: err}
	}

	phoneCipher, err := r.cipher.Encrypt(recipient.Phone)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt phone", Err: err}
	}

	emailCipher, err := r.cipher.Encrypt(recipient.Email)
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt email", Err: err}
	}

	publicAssistanceCipher, err := r.cipher.Encrypt(strconv.FormatBool(recipient.PublicAssistance))
	if err != nil {
		return &domain.RepositoryError{Op: "encrypt public_assistance", Err: err}
	}

	// Handle optional dates
	var admissionDate, dischargeDate *string
	if recipient.AdmissionDate != nil {
		dateStr := recipient.AdmissionDate.Format(time.RFC3339)
		admissionDate = &dateStr
	}
	if recipient.DischargeDate != nil {
		dateStr := recipient.DischargeDate.Format(time.RFC3339)
		dischargeDate = &dateStr
	}

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query,
		nameCipher, kanaCipher, sexCipher, birthDateCipher,
		disabilityNameCipher, hasDisabilityIDCipher, gradeCipher,
		addressCipher, phoneCipher, emailCipher, publicAssistanceCipher,
		admissionDate, dischargeDate,
		recipient.UpdatedAt.Format(time.RFC3339),
		recipient.ID,
	)

	if err != nil {
		return &domain.RepositoryError{Op: "update recipient", Err: err}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &domain.RepositoryError{Op: "check rows affected", Err: err}
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	// Clear sensitive encrypted data from memory
	crypto.ClearBytes(nameCipher)
	crypto.ClearBytes(kanaCipher)
	crypto.ClearBytes(sexCipher)
	crypto.ClearBytes(birthDateCipher)
	crypto.ClearBytes(disabilityNameCipher)
	crypto.ClearBytes(hasDisabilityIDCipher)
	crypto.ClearBytes(gradeCipher)
	crypto.ClearBytes(addressCipher)
	crypto.ClearBytes(phoneCipher)
	crypto.ClearBytes(emailCipher)
	crypto.ClearBytes(publicAssistanceCipher)

	return nil
}

// Delete deletes a recipient by ID
func (r *RecipientRepository) Delete(ctx context.Context, id domain.ID) error {
	query := `DELETE FROM recipients WHERE id = ?`

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return &domain.RepositoryError{Op: "delete recipient", Err: err}
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

// List retrieves recipients with pagination
func (r *RecipientRepository) List(ctx context.Context, limit, offset int) ([]*domain.Recipient, error) {
	query := `
		SELECT id, name_cipher, kana_cipher, sex_cipher, birth_date_cipher,
			   disability_name_cipher, has_disability_id_cipher, grade_cipher,
			   address_cipher, phone_cipher, email_cipher, public_assistance_cipher,
			   admission_date, discharge_date, created_at, updated_at
		FROM recipients 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	executor := r.getExecutor(ctx)
	
	// Use prepared statement for better performance with repeated queries
	stmt, err := executor.PrepareContext(ctx, query)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "prepare list recipients", Err: err}
	}
	defer stmt.Close()
	
	rows, err := stmt.QueryContext(ctx, limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "list recipients", Err: err}
	}
	defer rows.Close()

	// Pre-allocate slice with estimated capacity for better memory performance
	recipients := make([]*domain.Recipient, 0, limit)
	for rows.Next() {
		recipient, err := r.scanRecipient(rows)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return recipients, nil
}

// Search searches recipients by query (placeholder implementation)
func (r *RecipientRepository) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Recipient, error) {
	if query == "" {
		return r.List(ctx, limit, offset)
	}

	// Use cached hash-based search for better performance
	// This assumes search hashes are stored for encrypted fields
	sqlQuery := `
		SELECT id, name_cipher, kana_cipher, sex_cipher, birth_date_cipher,
			   disability_name_cipher, has_disability_id_cipher, grade_cipher,
			   address_cipher, phone_cipher, email_cipher, public_assistance_cipher,
			   admission_date, discharge_date, created_at, updated_at
		FROM recipients 
		WHERE id IN (
			SELECT DISTINCT recipient_id FROM search_index 
			WHERE search_hash = ?
		)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	// Create a search hash from the query (simplified implementation)
	searchHash := r.createSearchHash(query)
	
	executor := r.getExecutor(ctx)
	stmt, err := executor.PrepareContext(ctx, sqlQuery)
	if err != nil {
		// Fallback to full scan if search index is not available
		return r.searchFullScan(ctx, query, limit, offset)
	}
	defer stmt.Close()
	
	rows, err := stmt.QueryContext(ctx, searchHash, limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "search recipients", Err: err}
	}
	defer rows.Close()

	recipients := make([]*domain.Recipient, 0, limit)
	for rows.Next() {
		recipient, err := r.scanRecipient(rows)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "search rows iteration", Err: err}
	}

	return recipients, nil
}

// createSearchHash creates a hash for search indexing
func (r *RecipientRepository) createSearchHash(query string) string {
	// Normalize the search query for consistent hashing
	normalized := strings.ToLower(strings.TrimSpace(query))
	// Use SHA-256 for search hash (not for security, but for consistency)
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// searchFullScan performs a full table scan when search index is not available
func (r *RecipientRepository) searchFullScan(ctx context.Context, query string, limit, offset int) ([]*domain.Recipient, error) {
	// This is a fallback method - decrypt and search in memory
	// Not optimal for large datasets, but works when search index is unavailable
	allRecipients, err := r.List(ctx, 1000, 0) // Limit full scan to 1000 records
	if err != nil {
		return nil, err
	}

	var matched []*domain.Recipient
	searchLower := strings.ToLower(query)
	
	for _, recipient := range allRecipients {
		// Simple text matching on name and kana
		if strings.Contains(strings.ToLower(recipient.Name), searchLower) ||
		   strings.Contains(strings.ToLower(recipient.Kana), searchLower) {
			matched = append(matched, recipient)
			if len(matched) >= limit {
				break
			}
		}
	}

	// Apply offset
	if offset >= len(matched) {
		return []*domain.Recipient{}, nil
	}
	
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}
	
	return matched[offset:end], nil
}

// GetByStaffID retrieves recipients assigned to a staff member
func (r *RecipientRepository) GetByStaffID(ctx context.Context, staffID domain.ID) ([]*domain.Recipient, error) {
	query := `
		SELECT r.id, r.name_cipher, r.kana_cipher, r.sex_cipher, r.birth_date_cipher,
			   r.disability_name_cipher, r.has_disability_id_cipher, r.grade_cipher,
			   r.address_cipher, r.phone_cipher, r.email_cipher, r.public_assistance_cipher,
			   r.admission_date, r.discharge_date, r.created_at, r.updated_at
		FROM recipients r
		INNER JOIN staff_assignments sa ON r.id = sa.recipient_id
		WHERE sa.staff_id = ? AND sa.unassigned_at IS NULL
		ORDER BY r.created_at DESC`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, staffID)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get recipients by staff", Err: err}
	}
	defer rows.Close()

	var recipients []*domain.Recipient
	for rows.Next() {
		recipient, err := r.scanRecipient(rows)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return recipients, nil
}

// GetActive retrieves active recipients (not discharged)
func (r *RecipientRepository) GetActive(ctx context.Context, limit, offset int) ([]*domain.Recipient, error) {
	query := `
		SELECT id, name_cipher, kana_cipher, sex_cipher, birth_date_cipher,
			   disability_name_cipher, has_disability_id_cipher, grade_cipher,
			   address_cipher, phone_cipher, email_cipher, public_assistance_cipher,
			   admission_date, discharge_date, created_at, updated_at
		FROM recipients 
		WHERE discharge_date IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	executor := r.getExecutor(ctx)
	rows, err := executor.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "get active recipients", Err: err}
	}
	defer rows.Close()

	var recipients []*domain.Recipient
	for rows.Next() {
		recipient, err := r.scanRecipient(rows)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}

	if err := rows.Err(); err != nil {
		return nil, &domain.RepositoryError{Op: "rows iteration", Err: err}
	}

	return recipients, nil
}

// Count returns the total number of recipients
func (r *RecipientRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM recipients`

	executor := r.getExecutor(ctx)
	var count int
	err := executor.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, &domain.RepositoryError{Op: "count recipients", Err: err}
	}

	return count, nil
}

// CountActive returns the number of active recipients
func (r *RecipientRepository) CountActive(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM recipients WHERE discharge_date IS NULL`

	executor := r.getExecutor(ctx)
	var count int
	err := executor.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, &domain.RepositoryError{Op: "count active recipients", Err: err}
	}

	return count, nil
}

// getExecutor returns either a transaction or the database connection
func (r *RecipientRepository) getExecutor(ctx context.Context) executor {
	if tx := ctx.Value("tx"); tx != nil {
		return tx.(*sql.Tx)
	}
	return r.db.DB()
}

// executor interface for both *sql.DB and *sql.Tx
type executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// scanner interface for both *sql.Row and *sql.Rows
type scanner interface {
	Scan(dest ...interface{}) error
}

// scanRecipient scans a recipient from a database row
func (r *RecipientRepository) scanRecipient(row scanner) (*domain.Recipient, error) {
	var recipient domain.Recipient
	var nameCipher, kanaCipher, sexCipher, birthDateCipher []byte
	var disabilityNameCipher, hasDisabilityIDCipher, gradeCipher []byte
	var addressCipher, phoneCipher, emailCipher, publicAssistanceCipher []byte
	var admissionDateStr, dischargeDateStr, createdAtStr, updatedAtStr *string

	err := row.Scan(
		&recipient.ID, &nameCipher, &kanaCipher, &sexCipher, &birthDateCipher,
		&disabilityNameCipher, &hasDisabilityIDCipher, &gradeCipher,
		&addressCipher, &phoneCipher, &emailCipher, &publicAssistanceCipher,
		&admissionDateStr, &dischargeDateStr, &createdAtStr, &updatedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, &domain.RepositoryError{Op: "scan recipient", Err: err}
	}

	// Decrypt fields
	recipient.Name, err = r.cipher.Decrypt(nameCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt name", Err: err}
	}

	recipient.Kana, err = r.cipher.Decrypt(kanaCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt kana", Err: err}
	}

	sexStr, err := r.cipher.Decrypt(sexCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt sex", Err: err}
	}
	recipient.Sex = domain.Sex(sexStr)

	birthDateStr, err := r.cipher.Decrypt(birthDateCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt birth_date", Err: err}
	}
	recipient.BirthDate, err = time.Parse(time.RFC3339, birthDateStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse birth_date", Err: err}
	}

	recipient.DisabilityName, err = r.cipher.Decrypt(disabilityNameCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt disability_name", Err: err}
	}

	hasDisabilityIDStr, err := r.cipher.Decrypt(hasDisabilityIDCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt has_disability_id", Err: err}
	}
	recipient.HasDisabilityID, err = strconv.ParseBool(hasDisabilityIDStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse has_disability_id", Err: err}
	}

	recipient.Grade, err = r.cipher.Decrypt(gradeCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt grade", Err: err}
	}

	recipient.Address, err = r.cipher.Decrypt(addressCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt address", Err: err}
	}

	recipient.Phone, err = r.cipher.Decrypt(phoneCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt phone", Err: err}
	}

	recipient.Email, err = r.cipher.Decrypt(emailCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt email", Err: err}
	}

	publicAssistanceStr, err := r.cipher.Decrypt(publicAssistanceCipher)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "decrypt public_assistance", Err: err}
	}
	recipient.PublicAssistance, err = strconv.ParseBool(publicAssistanceStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse public_assistance", Err: err}
	}

	// Handle optional dates
	if admissionDateStr != nil {
		admissionDate, err := time.Parse(time.RFC3339, *admissionDateStr)
		if err != nil {
			return nil, &domain.RepositoryError{Op: "parse admission_date", Err: err}
		}
		recipient.AdmissionDate = &admissionDate
	}

	if dischargeDateStr != nil {
		dischargeDate, err := time.Parse(time.RFC3339, *dischargeDateStr)
		if err != nil {
			return nil, &domain.RepositoryError{Op: "parse discharge_date", Err: err}
		}
		recipient.DischargeDate = &dischargeDate
	}

	// Parse timestamps
	if createdAtStr != nil {
		recipient.CreatedAt, err = time.Parse(time.RFC3339, *createdAtStr)
		if err != nil {
			return nil, &domain.RepositoryError{Op: "parse created_at", Err: err}
		}
	}

	if updatedAtStr != nil {
		recipient.UpdatedAt, err = time.Parse(time.RFC3339, *updatedAtStr)
		if err != nil {
			return nil, &domain.RepositoryError{Op: "parse updated_at", Err: err}
		}
	}

	return &recipient, nil
}
