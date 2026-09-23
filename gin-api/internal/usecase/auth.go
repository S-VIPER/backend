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
	refreshTokenRepository RefreshTokenRepositoryInterface

	passwordHasher      service.PasswordHasher
	codeGenerator       service.VerificationCodeGenerator
	emailSender         service.EmailSender
	accessTokenService  service.AccessTokenService
	refreshTokenService service.RefreshTokenService
}

func NewAuthUseCase(
	userRepository UserRepositoryInterface,
	verificationRepository EmailVerificationRepositoryInterface,
	passwordHasher service.PasswordHasher,
	codeGenerator service.VerificationCodeGenerator,
	emailSender service.EmailSender,
	refreshTokenRepository RefreshTokenRepositoryInterface,
	accessTokenService service.AccessTokenService,
	refreshTokenService service.RefreshTokenService,
) *AuthUseCase {
	return &AuthUseCase{
		userRepository:         userRepository,
		verificationRepository: verificationRepository,
		refreshTokenRepository: refreshTokenRepository,
		passwordHasher:         passwordHasher,
		codeGenerator:          codeGenerator,
		emailSender:            emailSender,
		accessTokenService:     accessTokenService,
		refreshTokenService:    refreshTokenService,
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

	Login(
		ctx context.Context,
		email string,
		password string,
	) (*AuthResult, error)

	Refresh(
		ctx context.Context,
		refreshToken string,
	) (*AuthResult, error)

	Logout(
		ctx context.Context,
		refreshToken string,
	) error
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

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int32
}

var _ AuthUseCaseInterface = (*AuthUseCase)(nil)

const (
	registrationVerificationTTL = 10 * time.Minute
	verificationResendCooldown  = 60 * time.Second
	maxVerificationAttempts     = 5
	refreshTokenTTL             = 30 * 24 * time.Hour
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

func (u *AuthUseCase) Login(
	ctx context.Context,
	email string,
	password string,
) (*AuthResult, error) {
	email = normalizeEmail(email)

	user, err := u.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.EmailVerified {
		return nil, domain.ErrInvalidCredentials
	}

	if err := u.passwordHasher.Compare(user.PasswordHash, password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	now := time.Now()
	sessionID := uuid.New()
	plainRefreshToken, refreshTokenHash, err := u.refreshTokenService.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	session := &domain.RefreshTokenSession{
		ID:        sessionID,
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: now.Add(refreshTokenTTL),
		CreatedAt: now,
	}

	if err := u.refreshTokenRepository.Create(ctx, session); err != nil {
		return nil, err
	}

	accessToken, err := u.accessTokenService.Generate(user.ID, session.ID, now)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: plainRefreshToken,
		ExpiresIn:    int32(u.accessTokenService.ExpiresIn().Seconds()),
	}, nil
}

func (u *AuthUseCase) Refresh(
	ctx context.Context,
	refreshToken string,
) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, domain.ErrInvalidCredentials
	}

	oldSession, err := u.refreshTokenRepository.GetByTokenHash(
		ctx,
		u.refreshTokenService.Hash(refreshToken),
	)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	now := time.Now()
	if oldSession.RevokedAt != nil || !now.Before(oldSession.ExpiresAt) {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := u.userRepository.GetByID(ctx, oldSession.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.EmailVerified {
		return nil, domain.ErrInvalidCredentials
	}

	newSessionID := uuid.New()
	plainRefreshToken, refreshTokenHash, err := u.refreshTokenService.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	newSession := &domain.RefreshTokenSession{
		ID:        newSessionID,
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: now.Add(refreshTokenTTL),
		CreatedAt: now,
	}

	if err := u.refreshTokenRepository.Rotate(
		ctx,
		oldSession.ID,
		newSession,
		now,
	); err != nil {
		return nil, err
	}

	accessToken, err := u.accessTokenService.Generate(user.ID, newSession.ID, now)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: plainRefreshToken,
		ExpiresIn:    int32(u.accessTokenService.ExpiresIn().Seconds()),
	}, nil
}

func (u *AuthUseCase) Logout(
	ctx context.Context,
	refreshToken string,
) error {
	if refreshToken == "" {
		return nil
	}

	session, err := u.refreshTokenRepository.GetByTokenHash(
		ctx,
		u.refreshTokenService.Hash(refreshToken),
	)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil
		}
		return err
	}

	if session.RevokedAt != nil {
		return nil
	}

	return u.refreshTokenRepository.Revoke(ctx, session.ID, time.Now())
}
