package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/S-VIPER/backend/gin-api/gen/db"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	queries *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *User {
	return &User{
		queries: db.New(pool),
	}
}

var _ usecase.UserRepositoryInterface = (*User)(nil)

func (r *User) Create(
	ctx context.Context,
	user *domain.User,
) error {
	created, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}

		return fmt.Errorf("create user: %w", err)
	}

	*user = *toDomainCreateUser(created)

	return nil
}

func (r *User) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.User, error) {
	dbID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	user, err := r.queries.GetUserByID(ctx, dbID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return toDomainGetUserByID(user), nil
}

func (r *User) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return toDomainGetUserByEmail(user), nil
}

func (r *User) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	dbID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	if err := r.queries.DeleteUserByID(ctx, dbID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

func (r *User) Activate(
	ctx context.Context,
	userID uuid.UUID,
) error {
	err := r.queries.ActivateUser(
		ctx,
		pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},
	)
	if err != nil {
		return fmt.Errorf("activate user: %w", err)
	}

	return nil
}

func toDomainCreateUser(row db.User) *domain.User {
	return &domain.User{
		ID:            uuid.UUID(row.ID.Bytes),
		Email:         row.Email,
		PasswordHash:  row.PasswordHash,
		EmailVerified: row.EmailVerified,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func toDomainGetUserByID(row db.User) *domain.User {
	return &domain.User{
		ID:            uuid.UUID(row.ID.Bytes),
		Email:         row.Email,
		PasswordHash:  row.PasswordHash,
		EmailVerified: row.EmailVerified,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func toDomainGetUserByEmail(row db.User) *domain.User {
	return &domain.User{
		ID:            uuid.UUID(row.ID.Bytes),
		Email:         row.Email,
		PasswordHash:  row.PasswordHash,
		EmailVerified: row.EmailVerified,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}
