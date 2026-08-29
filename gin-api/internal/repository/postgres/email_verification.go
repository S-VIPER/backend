package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/S-VIPER/backend/gin-api/gen/db"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailVerification struct {
	queries *db.Queries
}

func NewEmailVerificationRepository(pool *pgxpool.Pool) *EmailVerification {
	return &EmailVerification{
		queries: db.New(pool),
	}
}

var _ usecase.EmailVerificationRepositoryInterface = (*EmailVerification)(nil)

func (r *EmailVerification) Create(
	ctx context.Context,
	verification *domain.EmailVerification,
) (*domain.EmailVerification, error) {
	created, err := r.queries.CreateEmailVerification(
		ctx,
		db.CreateEmailVerificationParams{
			UserID:    uuidToPGUUID(verification.UserID),
			CodeHash:  verification.CodeHash,
			ExpiresAt: timeToPGTimestamptz(verification.ExpiresAt),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create email verification: %w", err)
	}

	*verification = toDomainEmailVerification(created)

	return verification, nil
}

func (r *EmailVerification) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.EmailVerification, error) {
	verification, err := r.queries.GetEmailVerificationByID(
		ctx,
		uuidToPGUUID(id),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVerificationNotFound
		}

		return nil, fmt.Errorf(
			"get email verification by id: %w",
			err,
		)
	}

	result := toDomainEmailVerification(verification)

	return &result, nil
}

func (r *EmailVerification) UpdateCode(
	ctx context.Context,
	id uuid.UUID,
	codeHash string,
	expiresAt time.Time,
	createdAt time.Time,
) error {
	err := r.queries.UpdateEmailVerificationCode(
		ctx,
		db.UpdateEmailVerificationCodeParams{
			ID:        uuidToPGUUID(id),
			CodeHash:  codeHash,
			ExpiresAt: timeToPGTimestamptz(expiresAt),
			CreatedAt: timeToPGTimestamptz(createdAt),
		},
	)
	if err != nil {
		return fmt.Errorf(
			"update email verification code: %w",
			err,
		)
	}

	return nil
}

func (r *EmailVerification) IncrementAttempts(
	ctx context.Context,
	id uuid.UUID,
	maxAttempts int,
) error {
	err := r.queries.IncrementEmailVerificationAttempts(
		ctx,
		db.IncrementEmailVerificationAttemptsParams{
			ID:       uuidToPGUUID(id),
			Attempts: int32(maxAttempts),
		},
	)
	if err != nil {
		return fmt.Errorf(
			"increment email verification attempts: %w",
			err,
		)
	}

	return nil
}

func (r *EmailVerification) MarkVerified(
	ctx context.Context,
	id uuid.UUID,
	verifiedAt time.Time,
) error {
	err := r.queries.MarkEmailVerificationVerified(
		ctx,
		db.MarkEmailVerificationVerifiedParams{
			ID:         uuidToPGUUID(id),
			VerifiedAt: timeToPGTimestamptz(verifiedAt),
		},
	)
	if err != nil {
		return fmt.Errorf(
			"mark email verification verified: %w",
			err,
		)
	}

	return nil
}

func toDomainEmailVerification(
	verification db.EmailVerification,
) domain.EmailVerification {
	result := domain.EmailVerification{
		ID:        uuid.UUID(verification.ID.Bytes),
		UserID:    uuid.UUID(verification.UserID.Bytes),
		CodeHash:  verification.CodeHash,
		ExpiresAt: verification.ExpiresAt.Time,
		Attempts:  int(verification.Attempts),
		CreatedAt: verification.CreatedAt.Time,
	}

	if verification.VerifiedAt.Valid {
		verifiedAt := verification.VerifiedAt.Time
		result.VerifiedAt = &verifiedAt
	}

	return result
}

func uuidToPGUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

func timeToPGTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}
