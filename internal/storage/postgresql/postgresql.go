package postgresql

import (
	"context"
	"fmt"
	"net/url"
	"reviewer-service/internal/config"
	"reviewer-service/internal/domain/team"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Db *pgxpool.Pool
}

// Ключ для хранения транзакции в контексте
type txKey struct{}

func NewStorage(cfg *config.Datasource, ctx context.Context) (*Storage, error) {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Pass),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.Database,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	pool, err := pgxpool.New(ctx, u.String())
	if err != nil {
		return nil, err
	}

	return &Storage{Db: pool}, nil
}

func (s *Storage) Close() {
	if s.Db != nil {
		s.Db.Close()
	}
}

// WithTransaction реализует интерфейс TransactionManager из domain слоя
func (s *Storage) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx, err := s.Db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Сохраняем транзакцию в контексте
	txCtx := context.WithValue(ctx, txKey{}, tx)

	// Выполняем функцию в контексте транзакции
	err = fn(txCtx)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rollbackErr)
		}
		return err
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// getTx извлекает транзакцию из контекста, если она есть
// Если транзакции нет, возвращает обычное соединение из пула
func (s *Storage) getTx(ctx context.Context) (pgx.Tx, *pgxpool.Pool, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if ok {
		return tx, nil, true
	}
	return nil, s.Db, false
}

// EnsureStorageImplementsInterfaces проверяет, что Storage реализует необходимые интерфейсы
var (
	_ team.TransactionManager = (*Storage)(nil)
)
