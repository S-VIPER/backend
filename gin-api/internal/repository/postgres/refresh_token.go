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
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshToken struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshToken {
	return &RefreshToken{
		pool:    pool,
		queries: db.New(pool),
	}
}

var _ usecase.RefreshTokenRepositoryInterface = (*RefreshToken)(nil)

func (r *RefreshToken) Create(ctx context.Context, session *domain.RefreshTokenSession) error {
	created, err := r.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		ID:        uuidToPGUUID(session.ID),
		UserID:    uuidToPGUUID(session.UserID),
		TokenHash: session.TokenHash,
		ExpiresAt: timeToPGTimestamptz(session.ExpiresAt),
		CreatedAt: timeToPGTimestamptz(session.CreatedAt),
	})
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	*session = toDomainRefreshToken(created)
	return nil
}

func (r *RefreshToken) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshTokenSession, error) {
	session, err := r.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	result := toDomainRefreshToken(session)
	return &result, nil
}

func (r *RefreshToken) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	if err := r.queries.RevokeRefreshToken(ctx, db.RevokeRefreshTokenParams{
		ID:        uuidToPGUUID(id),
		RevokedAt: timeToPGTimestamptz(revokedAt),
	}); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *RefreshToken) Rotate(
	ctx context.Context,
	oldID uuid.UUID,
	newSession *domain.RefreshTokenSession,
	revokedAt time.Time,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin refresh token rotation: %w", err)
	}

	rollback := func() {
		_ = tx.Rollback(ctx)
	}

	q := r.queries.WithTx(tx)

	if err := q.RevokeRefreshToken(ctx, db.RevokeRefreshTokenParams{
		ID:        uuidToPGUUID(oldID),
		RevokedAt: timeToPGTimestamptz(revokedAt),
	}); err != nil {
		rollback()
		return fmt.Errorf("revoke old refresh token: %w", err)
	}

	created, err := q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		ID:        uuidToPGUUID(newSession.ID),
		UserID:    uuidToPGUUID(newSession.UserID),
		TokenHash: newSession.TokenHash,
		ExpiresAt: timeToPGTimestamptz(newSession.ExpiresAt),
		CreatedAt: timeToPGTimestamptz(newSession.CreatedAt),
	})
	if err != nil {
		rollback()
		return fmt.Errorf("create rotated refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit refresh token rotation: %w", err)
	}

	*newSession = toDomainRefreshToken(created)
	return nil
}

func toDomainRefreshToken(row db.RefreshToken) domain.RefreshTokenSession {
	result := domain.RefreshTokenSession{
		ID:        uuid.UUID(row.ID.Bytes),
		UserID:    uuid.UUID(row.UserID.Bytes),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt.Time,
		CreatedAt: row.CreatedAt.Time,
	}

	if row.RevokedAt.Valid {
		revokedAt := row.RevokedAt.Time
		result.RevokedAt = &revokedAt
	}

	return result
}
