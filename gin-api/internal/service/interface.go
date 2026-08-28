package service

import "context"

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type VerificationCodeGenerator interface {
	Generate() (plainCode string, codeHash string, err error)
	Compare(codeHash string, code string) bool
}

type VerificationCodeHasher interface {
	Hash(code string) (string, error)
	Compare(hash, code string) error
}

type EmailSender interface {
	SendRegistrationCode(
		ctx context.Context,
		email string,
		code string,
	) error
}
