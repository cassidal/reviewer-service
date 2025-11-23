package user

import (
	"context"
	"log/slog"
)

type Repository interface {
	UpdateUserIsActive(ctx context.Context, userId string, isActive bool) (int64, error)
	GetUser(ctx context.Context, id int64) (*Model, error)
	GetUserByUserId(ctx context.Context, userId string) (*Model, error)
}

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

func SetUserIsActive(ctx context.Context, log *slog.Logger, txManager TransactionManager, repo Repository, userId string, isActive bool) (*Model, error) {
	var updatedUser *Model

	err := txManager.WithTransaction(ctx, func(ctx context.Context) error {
		updatedUserId, err := repo.UpdateUserIsActive(ctx, userId, isActive)

		if err != nil {
			return err
		}

		updatedUser, err = repo.GetUser(ctx, updatedUserId)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Info("user is_active updated", slog.String("user_id", userId), slog.Bool("is_active", isActive))

	return updatedUser, nil
}
