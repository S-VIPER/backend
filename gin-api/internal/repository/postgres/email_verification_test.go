package postgres_test

import (
	"context"
	"errors"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	myPostgres "github.com/S-VIPER/backend/gin-api/internal/repository/postgres"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/gofrs/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/golang-migrate/migrate/v4"
	migratePgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type EmailVerificationIntegrationTest struct {
	suite.Suite

	ctx       context.Context
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool

	repositoryUser              usecase.UserRepositoryInterface
	repositoryEmailVerification usecase.EmailVerificationRepositoryInterface
}

func (s *EmailVerificationIntegrationTest) SetupTest() {
	ctx := context.Background()
	s.ctx = ctx

	container, err := postgres.Run(s.ctx,
		"postgres:18.6-alpine",
		postgres.WithDatabase("email-verification-test-db"),
		postgres.WithPassword("postgres"),
		postgres.WithUsername("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(40*time.Second),
		),
	)
	require.NoError(s.T(), err)
	s.container = container

	connectionStr, err := s.container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(s.T(), err)

	pool, err := pgxpool.New(s.ctx, connectionStr)
	require.NoError(s.T(), err)

	require.NoError(s.T(), pool.Ping(ctx))

	s.pool = pool

	db := stdlib.OpenDB(*s.pool.Config().ConnConfig)
	defer db.Close()

	driver, err := migratePgx.WithInstance(db, &migratePgx.Config{})
	require.NoError(s.T(), err)

	migrationPath := filepath.Join("..", "..", "..", "migrations")

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationPath, "pgx", driver)
	require.NoError(s.T(), err)

	err = m.Up()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(s.T(), err, "failed to run migrations")
	}

	s.repositoryEmailVerification = myPostgres.NewEmailVerificationRepository(s.pool)
	s.repositoryUser = myPostgres.NewUserRepository(s.pool)
}

func (s *EmailVerificationIntegrationTest) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.container != nil {
		err := s.container.Terminate(s.ctx)
		require.NoError(s.T(), err)
	}
}

func (s *EmailVerificationIntegrationTest) TearDownTest() {
	clearSQL := `
	DO $$
	DECLARE r RECORD;
	BEGIN
		FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename <> 'schema_migrations') LOOP
			EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE;';
		END LOOP;
	END $$;`

	_, err := s.pool.Exec(s.ctx, clearSQL)
	require.NoError(s.T(), err, "failed to truncate tables between tests")
}

func TestEmailVerificationTestSuite(t *testing.T) {
	suite.Run(t, new(EmailVerificationIntegrationTest))
}

func (s *EmailVerificationIntegrationTest) TestCreate_Sucess() {
	user := &domain.User{
		Email:        "verificationEmailTest@mail.com",
		PasswordHash: "secretPswdHash",
	}
	err := s.repositoryUser.Create(s.ctx, user)
	require.NoError(s.T(), err, "failed to create user for test creation email verification")

	userFromPostgres, err := s.repositoryUser.GetByEmail(s.ctx, user.Email)
	require.NoError(s.T(), err, "failed to get full user after creation for test creation email verification")
	verification := &domain.EmailVerification{
		UserID:    userFromPostgres.ID,
		CodeHash:  "secretCodeHash",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	log.Printf("%+v", verification)

	verificationFromRepo, err := s.repositoryEmailVerification.Create(s.ctx, verification)
	require.NoError(s.T(), err, "failed to create verification")
	require.NotEqual(s.T(), verificationFromRepo.ID, uuid.Nil, "failed generate new UUID to verification record")

	verificationGet, err := s.repositoryEmailVerification.GetByID(s.ctx, verificationFromRepo.ID)
	require.NoError(s.T(), err, "failed to get created verification record")
	require.NotEqual(s.T(), new(time.Time), verificationGet.CreatedAt, "postgres didn't change field createdAt")

}
