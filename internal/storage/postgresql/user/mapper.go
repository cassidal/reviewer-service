package user

import (
	"errors"
	"reviewer-service/internal/domain/user"
	"reviewer-service/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func ToEntity(dto *user.Model) *user.Model {
	return &user.Model{
		ID:       dto.ID,
		UserId:   dto.UserId,
		Username: dto.Username,
		TeamName: dto.TeamName,
		IsActive: dto.IsActive,
	}
}

func ToDomain(dto *Entity) *user.Model {
	return &user.Model{
		ID:       dto.ID,
		UserId:   dto.UserId,
		Username: dto.Username,
		TeamName: dto.TeamName,
		IsActive: dto.IsActive,
	}
}

func MapPGError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return storage.ErrUserNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return storage.ErrUserIdAlreadyExists
		case "23503":
			return storage.ErrUserNotFound
		}
	}
	return err
}
