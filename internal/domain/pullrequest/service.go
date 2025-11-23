package pullrequest

import (
	"context"
	"log/slog"
	"reviewer-service/internal/domain/user"
	"reviewer-service/internal/storage"
)

type Repository interface {
	CreatePullRequest(ctx context.Context, pr *Model) (int64, error)
	GetPullRequestById(ctx context.Context, pullRequestId string) (*Model, error)
	AssignReviewer(ctx context.Context, pullRequestId string, reviewerId string) error
	GetPullRequestsByReviewer(ctx context.Context, reviewerId string) ([]*Model, error)
	MergePullRequest(ctx context.Context, pullRequestId string) (*Model, error)
	RemoveReviewer(ctx context.Context, pullRequestId string, reviewerId string) error
	GetUserByUserId(ctx context.Context, userId string) (*user.Model, error)
	GetActiveReviewersByTeam(ctx context.Context, teamName string, excludeUserId string, limit int) ([]string, error)
	GetActiveReviewersByTeamExcluding(ctx context.Context, teamName string, excludeUserIds []string, limit int) ([]string, error)
}

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

func CreatePullRequest(ctx context.Context, log *slog.Logger, txManager TransactionManager, repo Repository, pr *Model) (*Model, error) {
	var createdPR *Model

	err := txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		author, err := repo.GetUserByUserId(txCtx, pr.AuthorId)
		if err != nil {
			return err
		}

		reviewers, err := repo.GetActiveReviewersByTeam(txCtx, author.TeamName, pr.AuthorId, 2)
		if err != nil {
			return err
		}

		pr.AssignedReviewers = reviewers

		_, err = repo.CreatePullRequest(txCtx, pr)
		if err != nil {
			return err
		}

		for _, reviewerId := range reviewers {
			err := repo.AssignReviewer(txCtx, pr.PullRequestId, reviewerId)
			if err != nil {
				return err
			}
		}

		createdPR, err = repo.GetPullRequestById(txCtx, pr.PullRequestId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Info("pull request created", slog.String("pull_request_id", pr.PullRequestId))

	return createdPR, nil
}

func MergePullRequest(ctx context.Context, log *slog.Logger, txManager TransactionManager, repo Repository, pullRequestId string) (*Model, error) {
	var mergedPR *Model

	err := txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		pr, err := repo.GetPullRequestById(txCtx, pullRequestId)
		if err != nil {
			return err
		}

		if pr.Status == "MERGED" {
			mergedPR = pr
			return nil
		}

		mergedPR, err = repo.MergePullRequest(txCtx, pullRequestId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Info("pull request merged", slog.String("pull_request_id", pullRequestId))

	return mergedPR, nil
}

func ReassignReviewer(ctx context.Context, log *slog.Logger, txManager TransactionManager, repo Repository, pullRequestId string, oldReviewerId string) (*Model, string, error) {
	var updatedPR *Model
	var newReviewerId string

	err := txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		pr, err := repo.GetPullRequestById(txCtx, pullRequestId)
		if err != nil {
			return err
		}

		if pr.Status == "MERGED" {
			return storage.ErrPullRequestMerged
		}

		isAssigned := false
		for _, reviewerId := range pr.AssignedReviewers {
			if reviewerId == oldReviewerId {
				isAssigned = true
				break
			}
		}

		if !isAssigned {
			return storage.ErrReviewerNotAssigned
		}

		oldReviewer, err := repo.GetUserByUserId(txCtx, oldReviewerId)
		if err != nil {
			return err
		}

		excludeList := []string{oldReviewerId, pr.AuthorId}
		for _, reviewerId := range pr.AssignedReviewers {
			if reviewerId != oldReviewerId {
				excludeList = append(excludeList, reviewerId)
			}
		}

		candidates, err := repo.GetActiveReviewersByTeamExcluding(txCtx, oldReviewer.TeamName, excludeList, 1)
		if err != nil {
			return err
		}

		err = repo.RemoveReviewer(txCtx, pullRequestId, oldReviewerId)
		if err != nil {
			return err
		}

		if len(candidates) > 0 {
			newReviewerId = candidates[0]
			err = repo.AssignReviewer(txCtx, pullRequestId, newReviewerId)
			if err != nil {
				return err
			}
		}

		updatedPR, err = repo.GetPullRequestById(txCtx, pullRequestId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	log.Info("reviewer reassigned",
		slog.String("pull_request_id", pullRequestId),
		slog.String("old_reviewer_id", oldReviewerId),
		slog.String("new_reviewer_id", newReviewerId))

	return updatedPR, newReviewerId, nil
}

func GetPullRequests(ctx context.Context, log *slog.Logger, repo Repository, userId string) ([]*Model, error) {
	_, err := repo.GetUserByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	log.Info("user retrieved", slog.String("user_id", userId))
	prs, err := repo.GetPullRequestsByReviewer(ctx, userId)
	if err != nil {
		return nil, err
	}

	return prs, nil
}
