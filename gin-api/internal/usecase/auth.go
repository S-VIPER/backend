package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/service"
	"github.com/google/uuid"
)

type AuthUseCase struct {
	userRepository         UserRepositoryInterface
	verificationRepository EmailVerificationRepositoryInterface

	passwordHasher service.PasswordHasher
	codeGenerator  service.VerificationCodeGenerator
	emailSender    service.EmailSender
}

func NewAuthUseCase(
	userRepository UserRepositoryInterface,
	verificationRepository EmailVerificationRepositoryInterface,
	passwordHasher service.PasswordHasher,
	codeGenerator service.VerificationCodeGenerator,
	emailSender service.EmailSender,
) *AuthUseCase {
	return &AuthUseCase{
		userRepository:         userRepository,
		verificationRepository: verificationRepository,
		passwordHasher:         passwordHasher,
		codeGenerator:          codeGenerator,
		emailSender:            emailSender,
	}
}

type AuthUseCaseInterface interface {
	Register(
		ctx context.Context,
		email string,
		password string,
	) (*RegistrationResult, error)

	VerifyRegistration(
		ctx context.Context,
		verificationID uuid.UUID,
		code string,
	) (*domain.User, error)

	ResendRegistrationVerification(
		ctx context.Context,
		verificationID uuid.UUID,
	) (*ResendVerificationResult, error)
}

type RegistrationResult struct {
	VerificationID uuid.UUID
	ExpiresIn      int32
}

type ResendVerificationResult struct {
	VerificationID uuid.UUID
	ExpiresIn      int32
	RetryAfter     int32
}

var _ AuthUseCaseInterface = (*AuthUseCase)(nil)

const (
	registrationVerificationTTL = 10 * time.Minute
	verificationResendCooldown  = 60 * time.Second
	maxVerificationAttempts     = 5
)

func (u *AuthUseCase) Register(
	ctx context.Context,
	email string,
	password string,
) (*RegistrationResult, error) {
	email = normalizeEmail(email)

	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	existingUser, err := u.userRepository.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	passwordHash, err := u.passwordHasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:         email,
		PasswordHash:  passwordHash,
		EmailVerified: false,
	}
	if err := u.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	user, err = u.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	plainCode, codeHash, err := u.codeGenerator.Generate()
	if err != nil {
		return nil, err
	}

	verification := &domain.EmailVerification{
		UserID:    user.ID,
		CodeHash:  codeHash,
		ExpiresAt: time.Now().Add(registrationVerificationTTL),
	}

	if verification, err = u.verificationRepository.Create(ctx, verification); err != nil {
		return nil, err
	}

	if err := u.emailSender.SendRegistrationCode(
		ctx,
		user.Email,
		plainCode,
	); err != nil {
		return nil, err
	}

	return &RegistrationResult{
		VerificationID: verification.ID,
		ExpiresIn:      int32(registrationVerificationTTL.Seconds()),
	}, nil
}

func (u *AuthUseCase) ResendRegistrationVerification(
	ctx context.Context,
	verificationID uuid.UUID,
) (*ResendVerificationResult, error) {
	verification, err := u.verificationRepository.GetByID(
		ctx,
		verificationID,
	)
	if err != nil {
		return nil, err
	}

	if verification.VerifiedAt != nil {
		return nil, domain.ErrVerificationInvalid
	}

	now := time.Now()

	if now.Before(verification.CreatedAt.Add(verificationResendCooldown)) {
		return nil, domain.ErrVerificationRateLimited
	}

	plainCode, codeHash, err := u.codeGenerator.Generate()
	if err != nil {
		return nil, err
	}

	expiresAt := now.Add(registrationVerificationTTL)

	if err := u.verificationRepository.UpdateCode(
		ctx,
		verification.ID,
		codeHash,
		expiresAt,
		now,
	); err != nil {
		return nil, err
	}

	user, err := u.userRepository.GetByID(ctx, verification.UserID)
	if err != nil {
		return nil, err
	}

	if err := u.emailSender.SendRegistrationCode(
		ctx,
		user.Email,
		plainCode,
	); err != nil {
		return nil, err
	}

	return &ResendVerificationResult{
		VerificationID: verification.ID,
		ExpiresIn:      int32(registrationVerificationTTL.Seconds()),
		RetryAfter:     int32(verificationResendCooldown.Seconds()),
	}, nil
}

func (u *AuthUseCase) VerifyRegistration(
	ctx context.Context,
	verificationID uuid.UUID,
	code string,
) (*domain.User, error) {
	verification, err := u.verificationRepository.GetByID(
		ctx,
		verificationID,
	)
	if err != nil {
		return nil, err
	}

	if verification.VerifiedAt != nil {
		return nil, domain.ErrVerificationInvalid
	}

	now := time.Now()

	if !now.Before(verification.ExpiresAt) {
		return nil, domain.ErrVerificationExpired
	}

	if verification.Attempts >= maxVerificationAttempts {
		return nil, domain.ErrVerificationTooManyAttempts
	}

	if ok := u.codeGenerator.Compare(
		verification.CodeHash,
		code,
	); !ok {
		if err := u.verificationRepository.IncrementAttempts(
			ctx,
			verification.ID,
			maxVerificationAttempts,
		); err != nil {
			return nil, fmt.Errorf("increment verification attempts: %w", err)
		}

		return nil, domain.ErrVerificationInvalid
	}

	if err := u.verificationRepository.MarkVerified(
		ctx,
		verification.ID,
		now,
	); err != nil {
		return nil, err
	}

	if err := u.userRepository.Activate(
		ctx,
		verification.UserID,
	); err != nil {
		return nil, err
	}

	user, err := u.userRepository.GetByID(
		ctx,
		verification.UserID,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}
