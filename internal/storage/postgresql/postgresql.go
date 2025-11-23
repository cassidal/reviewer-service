package postgresql

import (
	"database/sql"
	"fmt"
	"reviewer-service/internal/config"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(databaseConfig config.Datasource) (*Storage, error) {
	const op = "storage.postgresql.New"

	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		databaseConfig.User,
		databaseConfig.Pass,
		databaseConfig.Host,
		databaseConfig.Port,
		databaseConfig.Database,
	)

	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}
