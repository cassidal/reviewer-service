package integration

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"reviewer-service/internal/config"
	"reviewer-service/internal/http-server/handlers/pullrequest"
	"reviewer-service/internal/http-server/handlers/team"
	"reviewer-service/internal/http-server/handlers/user"
	"reviewer-service/internal/http-server/middleware/logger"
	"reviewer-service/internal/storage/postgresql"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestServer struct {
	Server      *http.Server
	Storage     *postgresql.Storage
	URL         string
	postgresC   testcontainers.Container
	postgresCtx context.Context
	postgresCancel context.CancelFunc
}

func init() {
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

func SetupTestServer(t *testing.T) (*TestServer, error) {
	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("reviewer_test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	storage := &postgresql.Storage{Db: pool}

	if err := setupTestDatabase(ctx, pool); err != nil {
		pool.Close()
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to setup test database: %w", err)
	}

	postgresCtx, postgresCancel := context.WithCancel(context.Background())

	log := config.MustConfigureLogger("test")

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/team/add", team.Save(log, storage, storage))
	router.Get("/team/get", team.Get(log, storage))
	router.Post("/users/setIsActive", user.SetIsActive(log, storage, storage))
	router.Get("/users/getReview", user.GetReview(log, storage))
	router.Post("/pullRequest/create", pullrequest.Create(log, storage, storage))
	router.Post("/pullRequest/merge", pullrequest.Merge(log, storage, storage))
	router.Post("/pullRequest/reassign", pullrequest.Reassign(log, storage, storage))

	server := &http.Server{
		Addr:    ":0",
		Handler: router,
	}

	return &TestServer{
		Server:        server,
		Storage:      storage,
		postgresC:    postgresContainer,
		postgresCtx:  postgresCtx,
		postgresCancel: postgresCancel,
	}, nil
}

func (ts *TestServer) Start() error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	ts.URL = fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)
	ts.Server.Addr = listener.Addr().String()

	go func() {
		if err := ts.Server.Serve(listener); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return nil
}

func (ts *TestServer) Close() error {
	var errs []error

	if ts.Storage != nil {
		ts.Storage.Close()
	}

	if ts.postgresC != nil {
		if err := ts.postgresC.Terminate(ts.postgresCtx); err != nil {
			errs = append(errs, err)
		}
	}

	if ts.postgresCancel != nil {
		ts.postgresCancel()
	}

	if ts.Server != nil {
		if err := ts.Server.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing test server: %v", errs)
	}

	return nil
}

func setupTestDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	schema := `
		DROP TABLE IF EXISTS pr_reviewers CASCADE;
		DROP TABLE IF EXISTS pull_requests CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
		DROP TABLE IF EXISTS team CASCADE;

		CREATE TABLE team (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL
		);

		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			user_id VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(255) NOT NULL,
			team_name VARCHAR(255) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true
		);

		CREATE TABLE pull_requests (
			id BIGSERIAL PRIMARY KEY,
			pull_request_id VARCHAR(255) UNIQUE NOT NULL,
			pull_request_name VARCHAR(255) NOT NULL,
			author_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			merged_at TIMESTAMP
		);

		CREATE TABLE pr_reviewers (
			pull_request_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			PRIMARY KEY (pull_request_id, user_id),
			FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id ON pull_requests(author_id);
		CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status);
		CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user_id ON pr_reviewers(user_id);
	`

	_, err := pool.Exec(ctx, schema)
	return err
}
