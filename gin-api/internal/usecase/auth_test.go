package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/service"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeUserRepository struct {
	users map[string]*domain.User

	createdUser *domain.User

	createErr     error
	getByEmailErr error
	getByIDErr    error
	deleteErr     error

	activateFn func(context.Context, uuid.UUID) error
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (f *fakeUserRepository) Create(
	_ context.Context,
	user *domain.User,
) error {
	if f.createErr != nil {
		return f.createErr
	}
	// Database fill this fields
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	f.createdUser = user

	if f.users == nil {
		f.users = make(map[string]*domain.User)
	}

	f.users[user.Email] = user

	return nil
}

func (f *fakeUserRepository) GetByEmail(
	_ context.Context,
	email string,
) (*domain.User, error) {
	if f.getByEmailErr != nil {
		return nil, f.getByEmailErr
	}

	user, ok := f.users[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

func (f *fakeUserRepository) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domain.User, error) {
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}

	for _, user := range f.users {
		if user.ID == id {
			return user, nil
		}
	}

	if f.createdUser != nil && f.createdUser.ID == id {
		return f.createdUser, nil
	}

	return nil, domain.ErrUserNotFound
}

func (f *fakeUserRepository) Delete(
	_ context.Context,
	id uuid.UUID,
) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}

	for email, user := range f.users {
		if user.ID == id {
			delete(f.users, email)
			return nil
		}
	}

	return nil
}

func (f *fakeUserRepository) Activate(
	ctx context.Context,
	userID uuid.UUID,
) error {
	if f.activateFn != nil {
		return f.activateFn(ctx, userID)
	}

	user, err := f.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.EmailVerified = true

	return nil
}

var _ usecase.UserRepositoryInterface = (*fakeUserRepository)(nil)

type fakeEmailVerificationRepository struct {
	createdVerification *domain.EmailVerification

	createErr            error
	getByIDErr           error
	updateCodeErr        error
	incrementAttemptsErr error
	markVerifiedErr      error

	createFn func(
		context.Context,
		*domain.EmailVerification,
	) (*domain.EmailVerification, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (*domain.EmailVerification, error)

	updateCodeFn func(
		context.Context,
		uuid.UUID,
		string,
		time.Time,
		time.Time,
	) error

	incrementAttemptsFn func(
		context.Context,
		uuid.UUID,
		int,
	) error

	markVerifiedFn func(
		context.Context,
		uuid.UUID,
		time.Time,
	) error
}

func (f *fakeEmailVerificationRepository) Create(
	ctx context.Context,
	verification *domain.EmailVerification,
) (*domain.EmailVerification, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}

	if f.createFn != nil {
		var err error
		verification, err = f.createFn(ctx, verification)
		if err != nil {
			return nil, err
		}
	}

	verification.ID = uuid.New()
	verification.CreatedAt = time.Now()

	f.createdVerification = verification

	return verification, nil
}

func (f *fakeEmailVerificationRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.EmailVerification, error) {
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}

	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}

	if f.createdVerification != nil &&
		f.createdVerification.ID == id {
		return f.createdVerification, nil
	}

	return nil, domain.ErrVerificationNotFound
}

func (f *fakeEmailVerificationRepository) UpdateCode(
	ctx context.Context,
	id uuid.UUID,
	codeHash string,
	expiresAt time.Time,
	createdAt time.Time,
) error {
	if f.updateCodeErr != nil {
		return f.updateCodeErr
	}

	if f.updateCodeFn != nil {
		return f.updateCodeFn(
			ctx,
			id,
			codeHash,
			expiresAt,
			createdAt,
		)
	}

	if f.createdVerification != nil &&
		f.createdVerification.ID == id {
		f.createdVerification.CodeHash = codeHash
		f.createdVerification.ExpiresAt = expiresAt
		f.createdVerification.CreatedAt = createdAt
		f.createdVerification.Attempts = 0
		f.createdVerification.VerifiedAt = nil
	}

	return nil
}

func (f *fakeEmailVerificationRepository) IncrementAttempts(
	ctx context.Context,
	id uuid.UUID,
	maxAttempts int,
) error {
	if f.incrementAttemptsErr != nil {
		return f.incrementAttemptsErr
	}

	if f.incrementAttemptsFn != nil {
		return f.incrementAttemptsFn(
			ctx,
			id,
			maxAttempts,
		)
	}

	if f.createdVerification != nil &&
		f.createdVerification.ID == id {
		f.createdVerification.Attempts++
	}

	return nil
}

func (f *fakeEmailVerificationRepository) MarkVerified(
	ctx context.Context,
	id uuid.UUID,
	verifiedAt time.Time,
) error {
	if f.markVerifiedErr != nil {
		return f.markVerifiedErr
	}

	if f.markVerifiedFn != nil {
		return f.markVerifiedFn(
			ctx,
			id,
			verifiedAt,
		)
	}

	if f.createdVerification != nil &&
		f.createdVerification.ID == id {
		f.createdVerification.VerifiedAt = &verifiedAt
	}

	return nil
}

var _ usecase.EmailVerificationRepositoryInterface = (*fakeEmailVerificationRepository)(nil)

type fakePasswordHasher struct {
	hashErr    error
	compareErr error

	hashedPassword string
	hashCalls      int
	compareCalls   int

	comparedHash     string
	comparedPassword string
}

func (h *fakePasswordHasher) Hash(password string) (string, error) {
	h.hashCalls++

	if h.hashErr != nil {
		return "", h.hashErr
	}

	h.hashedPassword = "hash:" + password

	return h.hashedPassword, nil
}

func (h *fakePasswordHasher) Compare(
	hash string,
	password string,
) error {
	h.compareCalls++

	h.comparedHash = hash
	h.comparedPassword = password

	return h.compareErr
}

var _ service.PasswordHasher = (*fakePasswordHasher)(nil)

type fakeVerificationCodeGenerator struct {
	generateErr error

	plainCode string
	codeHash  string

	generateCalls int
}

func (g *fakeVerificationCodeGenerator) Compare(codeHash string, code string) bool {
	return false
}

func (g *fakeVerificationCodeGenerator) Generate() (
	string,
	string,
	error,
) {
	g.generateCalls++

	if g.generateErr != nil {
		return "", "", g.generateErr
	}

	return g.plainCode, g.codeHash, nil
}

var _ service.VerificationCodeGenerator = (*fakeVerificationCodeGenerator)(nil)

type fakeEmailSender struct {
	sendErr error

	email string
	code  string

	sendCalls int
}

func (s *fakeEmailSender) SendRegistrationCode(
	_ context.Context,
	email string,
	code string,
) error {
	s.sendCalls++

	if s.sendErr != nil {
		return s.sendErr
	}

	s.email = email
	s.code = code

	return nil
}

var _ service.EmailSender = (*fakeEmailSender)(nil)

func newAuthUseCase(
	userRepo *fakeUserRepository,
	verificationRepo *fakeEmailVerificationRepository,
	passwordHasher *fakePasswordHasher,
	codeGenerator *fakeVerificationCodeGenerator,
	emailSender *fakeEmailSender,
) *usecase.AuthUseCase {
	return usecase.NewAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	userRepo := newFakeUserRepository()

	verificationRepo := &fakeEmailVerificationRepository{}

	passwordHasher := &fakePasswordHasher{}

	codeGenerator := &fakeVerificationCodeGenerator{
		plainCode: "123456",
		codeHash:  "hashed-code",
	}

	emailSender := &fakeEmailSender{}

	useCase := newAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	ctx := context.Background()

	result, err := useCase.Register(
		ctx,
		"  User@Example.COM  ",
		"password123",
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	require.NotEqual(t, uuid.Nil, result.VerificationID)
	require.Equal(t, int32(600), result.ExpiresIn)

	require.NotNil(t, userRepo.createdUser)

	user := userRepo.createdUser

	require.NotEqual(t, uuid.Nil, user.ID)
	require.Equal(t, "user@example.com", user.Email)
	require.Equal(t, "hash:password123", user.PasswordHash)
	require.Equal(t, false, user.EmailVerified)
	require.False(t, user.CreatedAt.IsZero())
	require.False(t, user.UpdatedAt.IsZero())

	require.NotNil(t, verificationRepo.createdVerification)

	verification := verificationRepo.createdVerification

	require.Equal(t, user.ID, verification.UserID)
	require.Equal(t, result.VerificationID, verification.ID)
	require.Equal(t, "hashed-code", verification.CodeHash)
	require.False(t, verification.CreatedAt.IsZero())
	require.True(t, verification.ExpiresAt.After(verification.CreatedAt))

	require.Equal(t, "user@example.com", emailSender.email)
	require.Equal(t, "123456", emailSender.code)

	require.Equal(t, 1, passwordHasher.hashCalls)
	require.Equal(t, 1, codeGenerator.generateCalls)
	require.Equal(t, 1, emailSender.sendCalls)
}

func TestAuthUseCase_Register_UserAlreadyExists(t *testing.T) {
	userRepo := newFakeUserRepository()

	existingUser := &domain.User{
		ID:            uuid.New(),
		Email:         "user@example.com",
		PasswordHash:  "existing-hash",
		EmailVerified: false,
	}

	userRepo.users[existingUser.Email] = existingUser

	verificationRepo := &fakeEmailVerificationRepository{}
	passwordHasher := &fakePasswordHasher{}
	codeGenerator := &fakeVerificationCodeGenerator{}
	emailSender := &fakeEmailSender{}

	useCase := newAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	result, err := useCase.Register(
		context.Background(),
		"user@example.com",
		"password123",
	)

	require.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	require.Nil(t, result)

	require.Equal(t, 0, passwordHasher.hashCalls)
	require.Equal(t, 0, codeGenerator.generateCalls)
	require.Equal(t, 0, emailSender.sendCalls)
}

func TestAuthUseCase_Register_GetUserByEmailError(t *testing.T) {
	expectedErr := errors.New("repository error")

	userRepo := newFakeUserRepository()
	userRepo.getByEmailErr = expectedErr

	verificationRepo := &fakeEmailVerificationRepository{}
	passwordHasher := &fakePasswordHasher{}
	codeGenerator := &fakeVerificationCodeGenerator{}
	emailSender := &fakeEmailSender{}

	useCase := newAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	result, err := useCase.Register(
		context.Background(),
		"user@example.com",
		"password123",
	)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, result)

	require.Nil(t, userRepo.createdUser)
	require.Equal(t, 0, passwordHasher.hashCalls)
}

func TestAuthUseCase_Register_PasswordHashError(t *testing.T) {
	expectedErr := errors.New("hash error")

	userRepo := newFakeUserRepository()

	verificationRepo := &fakeEmailVerificationRepository{}

	passwordHasher := &fakePasswordHasher{
		hashErr: expectedErr,
	}

	codeGenerator := &fakeVerificationCodeGenerator{}
	emailSender := &fakeEmailSender{}

	useCase := newAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	result, err := useCase.Register(
		context.Background(),
		"user@example.com",
		"password123",
	)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, result)

	require.Nil(t, userRepo.createdUser)
	require.Nil(t, verificationRepo.createdVerification)
	require.Equal(t, 0, codeGenerator.generateCalls)
	require.Equal(t, 0, emailSender.sendCalls)
}

func TestAuthUseCase_Register_CreateUserError(t *testing.T) {
	expectedErr := errors.New("create user error")

	userRepo := newFakeUserRepository()
	userRepo.createErr = expectedErr

	verificationRepo := &fakeEmailVerificationRepository{}
	passwordHasher := &fakePasswordHasher{}
	codeGenerator := &fakeVerificationCodeGenerator{}
	emailSender := &fakeEmailSender{}

	useCase := newAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	result, err := useCase.Register(
		context.Background(),
		"user@example.com",
		"password123",
	)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, result)

	require.Equal(t, 1, passwordHasher.hashCalls)
	require.Nil(t, verificationRepo.createdVerification)
	require.Equal(t, 0, codeGenerator.generateCalls)
	require.Equal(t, 0, emailSender.sendCalls)
}

func TestAuthUseCase_Register_CodeGenerationError(t *testing.T) {
	expectedErr := errors.New("code generation error")

	userRepo := newFakeUserRepository()

	verificationRepo := &fakeEmailVerificationRepository{}

	passwordHasher := &fakePasswordHasher{}

	codeGenerator := &fakeVerificationCodeGenerator{
		generateErr: expectedErr,
	}

	emailSender := &fakeEmailSender{}

	useCase := newAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	result, err := useCase.Register(
		context.Background(),
		"user@example.com",
		"password123",
	)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, result)

	require.NotNil(t, userRepo.createdUser)
	require.Nil(t, verificationRepo.createdVerification)
	require.Equal(t, 1, codeGenerator.generateCalls)
	require.Equal(t, 0, emailSender.sendCalls)
}

func TestAuthUseCase_Register_CreateVerificationError(t *testing.T) {
	expectedErr := errors.New("create verification error")

	userRepo := newFakeUserRepository()

	verificationRepo := &fakeEmailVerificationRepository{
		createErr: expectedErr,
	}

	passwordHasher := &fakePasswordHasher{}

	codeGenerator := &fakeVerificationCodeGenerator{
		plainCode: "123456",
		codeHash:  "hashed-code",
	}

	emailSender := &fakeEmailSender{}

	useCase := newAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	result, err := useCase.Register(
		context.Background(),
		"user@example.com",
		"password123",
	)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, result)

	require.NotNil(t, userRepo.createdUser)
	require.Nil(t, verificationRepo.createdVerification)
	require.Equal(t, 1, codeGenerator.generateCalls)
	require.Equal(t, 0, emailSender.sendCalls)
}

func TestAuthUseCase_Register_EmailSendError(t *testing.T) {
	expectedErr := errors.New("email send error")

	userRepo := newFakeUserRepository()

	verificationRepo := &fakeEmailVerificationRepository{}

	passwordHasher := &fakePasswordHasher{}

	codeGenerator := &fakeVerificationCodeGenerator{
		plainCode: "123456",
		codeHash:  "hashed-code",
	}

	emailSender := &fakeEmailSender{
		sendErr: expectedErr,
	}

	useCase := newAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	result, err := useCase.Register(
		context.Background(),
		"user@example.com",
		"password123",
	)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, result)

	require.NotNil(t, userRepo.createdUser)
	require.NotNil(t, verificationRepo.createdVerification)
	require.Equal(t, 1, emailSender.sendCalls)
}

func TestAuthUseCase_Register_InvalidEmail(t *testing.T) {
	userRepo := newFakeUserRepository()

	verificationRepo := &fakeEmailVerificationRepository{}
	passwordHasher := &fakePasswordHasher{}
	codeGenerator := &fakeVerificationCodeGenerator{}
	emailSender := &fakeEmailSender{}

	useCase := newAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	result, err := useCase.Register(
		context.Background(),
		"not-an-email",
		"password123",
	)

	require.ErrorIs(t, err, domain.ErrInvalidEmail)
	require.Nil(t, result)

	require.Nil(t, userRepo.createdUser)
	require.Equal(t, 0, passwordHasher.hashCalls)
	require.Equal(t, 0, codeGenerator.generateCalls)
	require.Equal(t, 0, emailSender.sendCalls)
}

func TestAuthUseCase_Register_InvalidPassword(t *testing.T) {
	userRepo := newFakeUserRepository()

	verificationRepo := &fakeEmailVerificationRepository{}
	passwordHasher := &fakePasswordHasher{}
	codeGenerator := &fakeVerificationCodeGenerator{}
	emailSender := &fakeEmailSender{}

	useCase := newAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	result, err := useCase.Register(
		context.Background(),
		"user@example.com",
		"short",
	)

	require.ErrorIs(t, err, domain.ErrInvalidPassword)
	require.Nil(t, result)

	require.Nil(t, userRepo.createdUser)
	require.Equal(t, 0, passwordHasher.hashCalls)
	require.Equal(t, 0, codeGenerator.generateCalls)
	require.Equal(t, 0, emailSender.sendCalls)
}
