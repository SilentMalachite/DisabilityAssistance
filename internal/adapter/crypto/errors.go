package crypto

import "errors"

// crypto パッケージ固有のエラー定義
var (
	ErrPasswordRequired = errors.New("password is required")
	ErrWeakPassword     = errors.New("password is too weak")
	ErrInvalidPassword  = errors.New("invalid password")
)