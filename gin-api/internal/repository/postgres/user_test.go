package postgres_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	myPostgres "github.com/S-VIPER/backend/gin-api/internal/repository/postgres"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/gofrs/uuid"
	"github.com/golang-migrate/migrate/v4"
	migrateDriver "github.com/golang-migrate/migrate/v4/database/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type UserPostgresIntegrationTestSuite struct {
	suite.Suite

	ctx       context.Context
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool

	repository usecase.UserRepositoryInterface
}

func (s *UserPostgresIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := postgres.Run(
		s.ctx,
		"postgres:18.6-alpine",
		postgres.WithDatabase("user-repository-test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(40*time.Second),
		),
	)
	require.NoError(s.T(), err)
	s.container = container

	connectionStr, err := container.ConnectionString(
		s.ctx,
		"sslmode=disable",
	)
	require.NoError(s.T(), err)

	pool, err := pgxpool.New(s.ctx, connectionStr)
	require.NoError(s.T(), err)

	require.NoError(s.T(), pool.Ping(s.ctx))

	s.pool = pool

	s.runMigrations()
	s.repository = myPostgres.NewUserRepository(s.pool)

}

func (s *UserPostgresIntegrationTestSuite) runMigrations() {
	db := stdlib.OpenDB(*s.pool.Config().ConnConfig)
	defer db.Close()

	driver, err := migrateDriver.WithInstance(db, &migrateDriver.Config{})
	require.NoError(s.T(), err)

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"pgx",
		driver,
	)
	require.NoError(s.T(), err)

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(s.T(), err, "failed to run migrations")
	}

}

func (s *UserPostgresIntegrationTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.container != nil {
		err := s.container.Terminate(s.ctx)
		require.NoError(s.T(), err)
	}
}

func (s *UserPostgresIntegrationTestSuite) TearDownTest() {
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

func TestUserPostgresIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserPostgresIntegrationTestSuite))
}

func (s *UserPostgresIntegrationTestSuite) TestCreate_Success() {
	user := &domain.User{
		Email:        "migrate-test@example.com",
		PasswordHash: "secret_hash",
	}
	err := s.repository.Create(s.ctx, user)
	require.NoError(s.T(), err)
	s.NotEqual(uuid.Nil, user.ID)
}
