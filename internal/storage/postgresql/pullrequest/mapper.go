package pullrequest

import (
	"errors"
	"reviewer-service/internal/domain/pullrequest"
	"reviewer-service/internal/storage"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func ToEntity(pr *pullrequest.Model) *Entity {
	var createdAt time.Time
	if pr.CreatedAt != nil {
		createdAt = *pr.CreatedAt
	}

	return &Entity{
		ID:              pr.ID,
		PullRequestId:   pr.PullRequestId,
		PullRequestName: pr.PullRequestName,
		AuthorId:        pr.AuthorId,
		Status:          pr.Status,
		CreatedAt:       createdAt,
		MergedAt:        pr.MergedAt,
	}
}

func ToDomain(entity *Entity, reviewers []string) *pullrequest.Model {
	return &pullrequest.Model{
		ID:                entity.ID,
		PullRequestId:     entity.PullRequestId,
		PullRequestName:   entity.PullRequestName,
		AuthorId:          entity.AuthorId,
		Status:            entity.Status,
		AssignedReviewers: reviewers,
		CreatedAt:         &entity.CreatedAt,
		MergedAt:          entity.MergedAt,
	}
}

func MapPGError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return storage.ErrPullRequestNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return storage.ErrPullRequestAlreadyExists
		case "23503":
			return storage.ErrPullRequestNotFound
		}
	}
	return err
}
