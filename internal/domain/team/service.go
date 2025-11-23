package team

import (
	"context"
	"log/slog"
	"reviewer-service/internal/domain/user"
)

type Repository interface {
	CreateTeam(ctx context.Context, t *Model) (int64, error)
	GetTeam(ctx context.Context, id int64) (*Model, error)
	GetTeamByName(ctx context.Context, name string) (*Model, error)
	CreateUser(ctx context.Context, u *user.Model) (int64, error)
	GetUserByUserId(ctx context.Context, userId string) (*user.Model, error)
	GetActiveReviewersByTeam(ctx context.Context, teamName string, excludeUserId string, limit int) ([]string, error)
	GetActiveReviewersByTeamExcluding(ctx context.Context, teamName string, excludeUserIds []string, limit int) ([]string, error)
}

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

func SaveTeam(ctx context.Context, log *slog.Logger, txManager TransactionManager, repo Repository, t *Model) (*Model, error) {
	var savedTeam *Model
	var teamID int64

	err := txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		teamID, err = repo.CreateTeam(txCtx, t)
		if err != nil {
			return err
		}

		for _, member := range t.Members {
			_, err := repo.CreateUser(txCtx, member)
			if err != nil {
				return err
			}
		}

		savedTeam, err = repo.GetTeam(txCtx, teamID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Info("team saved", slog.Int64("id", teamID))

	return savedTeam, nil
}

func GetTeamByName(ctx context.Context, log *slog.Logger, repo Repository, name string) (*Model, error) {
	teamModel, err := repo.GetTeamByName(ctx, name)
	if err != nil {
		return nil, err
	}

	log.Info("team retrieved", slog.String("name", name))
	return teamModel, nil
}
