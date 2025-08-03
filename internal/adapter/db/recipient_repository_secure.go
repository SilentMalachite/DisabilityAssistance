package db

import (
	"context"
	"strconv"
	"time"

	"shien-system/internal/adapter/crypto"
	"shien-system/internal/domain"
)

// scanRecipientSecure scans a row into a Recipient with secure memory handling
func (r *RecipientRepository) scanRecipientSecure(row scanner) (*domain.Recipient, error) {
	var recipient domain.Recipient
	var nameCipher, kanaCipher, sexCipher, birthDateCipher []byte
	var disabilityNameCipher, hasDisabilityIDCipher, gradeCipher []byte
	var addressCipher, phoneCipher, emailCipher, publicAssistanceCipher []byte
	var admissionDateStr, dischargeDateStr, createdAtStr, updatedAtStr *string

	err := row.Scan(
		&recipient.ID,
		&nameCipher,
		&kanaCipher,
		&sexCipher,
		&birthDateCipher,
		&disabilityNameCipher,
		&hasDisabilityIDCipher,
		&gradeCipher,
		&addressCipher,
		&phoneCipher,
		&emailCipher,
		&publicAssistanceCipher,
		&admissionDateStr,
		&dischargeDateStr,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		return nil, err
	}

	// Use secure memory for decryption
	decryptSecureField := func(ciphertext []byte, fieldName string) (string, error) {
		secureStr, err := r.cipher.DecryptSecure(ciphertext)
		if err != nil {
			return "", &domain.RepositoryError{Op: "decrypt " + fieldName, Err: err}
		}
		defer secureStr.Clear() // Ensure cleanup
		return secureStr.String(), nil
	}

	// Decrypt all fields with secure memory handling
	var err2 error

	recipient.Name, err2 = decryptSecureField(nameCipher, "name")
	if err2 != nil {
		return nil, err2
	}

	recipient.Kana, err2 = decryptSecureField(kanaCipher, "kana")
	if err2 != nil {
		return nil, err2
	}

	sexStr, err2 := decryptSecureField(sexCipher, "sex")
	if err2 != nil {
		return nil, err2
	}
	recipient.Sex = domain.Sex(sexStr)

	birthDateStr, err2 := decryptSecureField(birthDateCipher, "birth_date")
	if err2 != nil {
		return nil, err2
	}
	recipient.BirthDate, err = time.Parse(time.RFC3339, birthDateStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse birth_date", Err: err}
	}

	recipient.DisabilityName, err2 = decryptSecureField(disabilityNameCipher, "disability_name")
	if err2 != nil {
		return nil, err2
	}

	hasDisabilityIDStr, err2 := decryptSecureField(hasDisabilityIDCipher, "has_disability_id")
	if err2 != nil {
		return nil, err2
	}
	recipient.HasDisabilityID, err = strconv.ParseBool(hasDisabilityIDStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse has_disability_id", Err: err}
	}

	recipient.Grade, err2 = decryptSecureField(gradeCipher, "grade")
	if err2 != nil {
		return nil, err2
	}

	recipient.Address, err2 = decryptSecureField(addressCipher, "address")
	if err2 != nil {
		return nil, err2
	}

	recipient.Phone, err2 = decryptSecureField(phoneCipher, "phone")
	if err2 != nil {
		return nil, err2
	}

	recipient.Email, err2 = decryptSecureField(emailCipher, "email")
	if err2 != nil {
		return nil, err2
	}

	publicAssistanceStr, err2 := decryptSecureField(publicAssistanceCipher, "public_assistance")
	if err2 != nil {
		return nil, err2
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
	recipient.CreatedAt, err = time.Parse(time.RFC3339, *createdAtStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse created_at", Err: err}
	}

	recipient.UpdatedAt, err = time.Parse(time.RFC3339, *updatedAtStr)
	if err != nil {
		return nil, &domain.RepositoryError{Op: "parse updated_at", Err: err}
	}

	return &recipient, nil
}

// GetByIDSecure retrieves a recipient by ID with secure memory handling
func (r *RecipientRepository) GetByIDSecure(ctx context.Context, id domain.ID) (*domain.Recipient, error) {
	query := `
        SELECT 
            id, name_cipher, kana_cipher, sex_cipher, birth_date_cipher,
            disability_name_cipher, has_disability_id_cipher, grade_cipher,
            address_cipher, phone_cipher, email_cipher, public_assistance_cipher,
            admission_date, discharge_date, created_at, updated_at
        FROM recipients
        WHERE id = ?`

	executor := r.getExecutor(ctx)
	row := executor.QueryRowContext(ctx, query, id)
	return r.scanRecipientSecure(row)
}

// SecureRecipientData wraps recipient data with automatic cleanup
type SecureRecipientData struct {
	recipient    *domain.Recipient
	secureFields []*crypto.SecureString
}

// NewSecureRecipientData creates a new SecureRecipientData
func NewSecureRecipientData(recipient *domain.Recipient) *SecureRecipientData {
	return &SecureRecipientData{
		recipient:    recipient,
		secureFields: make([]*crypto.SecureString, 0),
	}
}

// Clear cleans up all secure fields
func (s *SecureRecipientData) Clear() {
	for _, field := range s.secureFields {
		if field != nil {
			field.Clear()
		}
	}
	s.secureFields = nil

	// Clear string fields in recipient
	if s.recipient != nil {
		crypto.ClearString(&s.recipient.Name)
		crypto.ClearString(&s.recipient.Kana)
		crypto.ClearString(&s.recipient.DisabilityName)
		crypto.ClearString(&s.recipient.Grade)
		crypto.ClearString(&s.recipient.Address)
		crypto.ClearString(&s.recipient.Phone)
		crypto.ClearString(&s.recipient.Email)
	}
}
