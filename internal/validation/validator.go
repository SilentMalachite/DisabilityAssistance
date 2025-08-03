package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	
	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// Validator provides input validation functions
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequired checks if a field is not empty
func (v *Validator) ValidateRequired(field, value string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   field,
			Message: "必須項目です",
		}
	}
	return nil
}

// ValidateLength checks string length constraints
func (v *Validator) ValidateLength(field, value string, min, max int) *ValidationError {
	length := utf8.RuneCountInString(value)
	if length < min {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("最低%d文字以上入力してください", min),
		}
	}
	if max > 0 && length > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("最大%d文字以内で入力してください", max),
		}
	}
	return nil
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(field, value string) *ValidationError {
	if value == "" {
		return nil // Empty email is allowed unless required
	}
	
	if _, err := mail.ParseAddress(value); err != nil {
		return &ValidationError{
			Field:   field,
			Message: "有効なメールアドレスを入力してください",
		}
	}
	return nil
}

// ValidatePhoneNumber validates Japanese phone number format
func (v *Validator) ValidatePhoneNumber(field, value string) *ValidationError {
	if value == "" {
		return nil // Empty phone is allowed unless required
	}
	
	// Remove common separators
	cleaned := strings.ReplaceAll(value, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	
	// Check if all remaining characters are digits
	if !regexp.MustCompile(`^\d+$`).MatchString(cleaned) {
		return &ValidationError{
			Field:   field,
			Message: "電話番号は数字とハイフンのみ入力してください",
		}
	}
	
	// Validate Japanese phone number patterns
	phonePattern := regexp.MustCompile(`^(0\d{1,4}-?\d{1,4}-?\d{3,4}|0\d{9,10})$`)
	if !phonePattern.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "正しい電話番号の形式で入力してください（例：03-1234-5678）",
		}
	}
	
	return nil
}

// ValidateDate validates date format (YYYY-MM-DD)
func (v *Validator) ValidateDate(field, value string) *ValidationError {
	if value == "" {
		return nil // Empty date is allowed unless required
	}
	
	// Try to parse as YYYY-MM-DD
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return &ValidationError{
			Field:   field,
			Message: "日付はYYYY-MM-DD形式で入力してください",
		}
	}
	
	return nil
}

// ValidateInteger validates integer input
func (v *Validator) ValidateInteger(field, value string) *ValidationError {
	if value == "" {
		return nil // Empty integer is allowed unless required
	}
	
	if _, err := strconv.Atoi(value); err != nil {
		return &ValidationError{
			Field:   field,
			Message: "数字を入力してください",
		}
	}
	
	return nil
}

// ValidateIntegerRange validates integer within range
func (v *Validator) ValidateIntegerRange(field, value string, min, max int) *ValidationError {
	if value == "" {
		return nil // Empty integer is allowed unless required
	}
	
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return &ValidationError{
			Field:   field,
			Message: "数字を入力してください",
		}
	}
	
	if intVal < min || intVal > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%d～%dの範囲で入力してください", min, max),
		}
	}
	
	return nil
}

// ValidateJapaneseName validates Japanese name format
func (v *Validator) ValidateJapaneseName(field, value string) *ValidationError {
	if value == "" {
		return nil // Empty name allowed unless required
	}
	
	// Check for valid Japanese name characters (hiragana, katakana, kanji, spaces)
	for _, r := range value {
		if !unicode.Is(unicode.Hiragana, r) && 
		   !unicode.Is(unicode.Katakana, r) && 
		   !unicode.Is(unicode.Han, r) && 
		   r != ' ' && r != '　' { // Include both space types
			return &ValidationError{
				Field:   field,
				Message: "日本語（ひらがな、カタカナ、漢字）で入力してください",
			}
		}
	}
	
	return nil
}

// ValidateKana validates katakana format
func (v *Validator) ValidateKana(field, value string) *ValidationError {
	if value == "" {
		return nil // Empty kana allowed unless required
	}
	
	// Check for valid katakana characters and spaces
	for _, r := range value {
		if !unicode.Is(unicode.Katakana, r) && r != ' ' && r != '　' {
			return &ValidationError{
				Field:   field,
				Message: "カタカナで入力してください",
			}
		}
	}
	
	return nil
}

// ValidateFilePath validates file path format
func (v *Validator) ValidateFilePath(field, value string) *ValidationError {
	if value == "" {
		return nil // Empty path allowed unless required
	}
	
	// Check for invalid characters in file paths
	invalidChars := `<>:"|?*`
	for _, char := range invalidChars {
		if strings.ContainsRune(value, char) {
			return &ValidationError{
				Field:   field,
				Message: "ファイルパスに無効な文字が含まれています",
			}
		}
	}
	
	return nil
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(field, value string) *ValidationError {
	if value == "" {
		return &ValidationError{
			Field:   field,
			Message: "パスワードは必須です",
		}
	}
	
	if len(value) < 8 {
		return &ValidationError{
			Field:   field,
			Message: "パスワードは8文字以上である必要があります",
		}
	}
	
	// Check for at least one uppercase, one lowercase, and one digit
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(value)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(value)
	hasDigit := regexp.MustCompile(`\d`).MatchString(value)
	
	if !hasUpper || !hasLower || !hasDigit {
		return &ValidationError{
			Field:   field,
			Message: "パスワードは大文字、小文字、数字をそれぞれ1文字以上含む必要があります",
		}
	}
	
	return nil
}

// ValidateNotContainSQLKeywords checks for potential SQL injection attempts
func (v *Validator) ValidateNotContainSQLKeywords(field, value string) *ValidationError {
	if value == "" {
		return nil
	}
	
	// List of dangerous SQL keywords
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
		"UNION", "OR", "AND", "WHERE", "FROM", "INTO", "VALUES", "SET",
		"--", "/*", "*/", ";", "xp_", "sp_",
	}
	
	upperValue := strings.ToUpper(value)
	for _, keyword := range sqlKeywords {
		if strings.Contains(upperValue, keyword) {
			return &ValidationError{
				Field:   field,
				Message: "不正な文字列が含まれています",
			}
		}
	}
	
	return nil
}

// ValidateNotContainXSS checks for potential XSS attempts
func (v *Validator) ValidateNotContainXSS(field, value string) *ValidationError {
	if value == "" {
		return nil
	}
	
	// List of dangerous XSS patterns
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "onload=", "onerror=", 
		"onclick=", "onmouseover=", "alert(", "document.cookie",
		"eval(", "expression(", "vbscript:",
	}
	
	lowerValue := strings.ToLower(value)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerValue, pattern) {
			return &ValidationError{
				Field:   field,
				Message: "不正な文字列が含まれています",
			}
		}
	}
	
	return nil
}

// ValidateMultiple validates multiple constraints and returns all errors
func (v *Validator) ValidateMultiple(validations ...func() *ValidationError) ValidationErrors {
	var errors ValidationErrors
	
	for _, validation := range validations {
		if err := validation(); err != nil {
			errors = append(errors, *err)
		}
	}
	
	return errors
}