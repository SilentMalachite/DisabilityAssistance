package crypto

import (
	"crypto/rand"
	"runtime"
	"unsafe"
)

// SecureString wraps a string and provides secure cleanup
type SecureString struct {
	data []byte
}

// NewSecureString creates a new SecureString from plaintext
func NewSecureString(plaintext string) *SecureString {
	// Copy the string data to avoid referencing the original
	data := make([]byte, len(plaintext))
	copy(data, plaintext)

	s := &SecureString{data: data}

	// Set up finalizer to ensure cleanup even if Clear is not called
	runtime.SetFinalizer(s, (*SecureString).Clear)

	return s
}

// String returns the string value (use with caution)
func (s *SecureString) String() string {
	if s.data == nil {
		return ""
	}
	return string(s.data)
}

// Clear securely overwrites the memory and marks it for GC
func (s *SecureString) Clear() {
	if s.data == nil {
		return
	}

	// Overwrite with random data
	rand.Read(s.data)

	// Overwrite with zeros
	for i := range s.data {
		s.data[i] = 0
	}

	// Clear the slice
	s.data = nil

	// Remove finalizer
	runtime.SetFinalizer(s, nil)

	// Force garbage collection (optional, for critical data)
	runtime.GC()
}

// SecureBytes wraps a byte slice and provides secure cleanup
type SecureBytes struct {
	data []byte
}

// NewSecureBytes creates a new SecureBytes
func NewSecureBytes(data []byte) *SecureBytes {
	// Copy the data to avoid referencing the original
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	s := &SecureBytes{data: dataCopy}

	// Set up finalizer to ensure cleanup even if Clear is not called
	runtime.SetFinalizer(s, (*SecureBytes).Clear)

	return s
}

// Bytes returns the byte slice (use with caution)
func (s *SecureBytes) Bytes() []byte {
	if s.data == nil {
		return nil
	}
	// Return a copy to prevent external modification
	result := make([]byte, len(s.data))
	copy(result, s.data)
	return result
}

// Clear securely overwrites the memory and marks it for GC
func (s *SecureBytes) Clear() {
	if s.data == nil {
		return
	}

	// Overwrite with random data
	rand.Read(s.data)

	// Overwrite with zeros
	for i := range s.data {
		s.data[i] = 0
	}

	// Clear the slice
	s.data = nil

	// Remove finalizer
	runtime.SetFinalizer(s, nil)
}

// ClearString securely clears a string from memory
// Note: This is best-effort as Go strings are immutable
func ClearString(s *string) {
	if s == nil {
		return
	}

	// Get the underlying string header
	sh := (*struct {
		str unsafe.Pointer
		len int
	})(unsafe.Pointer(s))

	if sh.str == nil || sh.len == 0 {
		return
	}

	// Create a byte slice from the string data
	b := (*[1 << 30]byte)(sh.str)[:sh.len:sh.len]

	// Overwrite with zeros
	for i := range b {
		b[i] = 0
	}

	// Clear the string
	*s = ""
}

// ClearBytes securely clears a byte slice
func ClearBytes(b []byte) {
	if len(b) == 0 {
		return
	}

	// Overwrite with random data
	rand.Read(b)

	// Overwrite with zeros
	for i := range b {
		b[i] = 0
	}
}
