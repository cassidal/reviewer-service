package postgresql

import (
	"context"
	"reviewer-service/internal/domain/pullrequest"
	storagePR "reviewer-service/internal/storage/postgresql/pullrequest"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreatePullRequest(ctx context.Context, pr *pullrequest.Model) (int64, error) {
	entity := storagePR.ToEntity(pr)

	var id int64
	var err error

	tx, pool, hasTx := s.getTx(ctx)

	sql := `
		INSERT INTO pull_requests 
			(pull_request_id, pull_request_name, author_id, status, created_at) 
		VALUES 
			($1, $2, $3, $4, NOW())
		RETURNING id
	`

	if hasTx {
		err = tx.QueryRow(
			ctx,
			sql,
			entity.PullRequestId,
			entity.PullRequestName,
			entity.AuthorId,
			entity.Status,
		).Scan(&id)
	} else {
		err = pool.QueryRow(
			ctx,
			sql,
			entity.PullRequestId,
			entity.PullRequestName,
			entity.AuthorId,
			entity.Status,
		).Scan(&id)
	}

	if err != nil {
		return 0, storagePR.MapPGError(err)
	}

	return id, nil
}

func (s *Storage) GetPullRequestById(ctx context.Context, pullRequestId string) (*pullrequest.Model, error) {
	tx, pool, hasTx := s.getTx(ctx)

	query := `
		SELECT 
			pr.id,
			pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			pr.status,
			pr.created_at,
			pr.merged_at
		FROM pull_requests pr
		WHERE pr.pull_request_id = $1
	`

	var entity storagePR.Entity
	var err error

	if hasTx {
		err = tx.QueryRow(ctx, query, pullRequestId).Scan(
			&entity.ID,
			&entity.PullRequestId,
			&entity.PullRequestName,
			&entity.AuthorId,
			&entity.Status,
			&entity.CreatedAt,
			&entity.MergedAt,
		)
	} else {
		err = pool.QueryRow(ctx, query, pullRequestId).Scan(
			&entity.ID,
			&entity.PullRequestId,
			&entity.PullRequestName,
			&entity.AuthorId,
			&entity.Status,
			&entity.CreatedAt,
			&entity.MergedAt,
		)
	}

	if err != nil {
		return nil, storagePR.MapPGError(err)
	}

	reviewersQuery := `
		SELECT user_id 
		FROM pr_reviewers 
		WHERE pull_request_id = $1
		ORDER BY user_id
	`

	var rows pgx.Rows
	if hasTx {
		rows, err = tx.Query(ctx, reviewersQuery, pullRequestId)
	} else {
		rows, err = pool.Query(ctx, reviewersQuery, pullRequestId)
	}

	if err != nil {
		return nil, storagePR.MapPGError(err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerId string
		if err := rows.Scan(&reviewerId); err != nil {
			return nil, storagePR.MapPGError(err)
		}
		reviewers = append(reviewers, reviewerId)
	}

	return storagePR.ToDomain(&entity, reviewers), nil
}

func (s *Storage) AssignReviewer(ctx context.Context, pullRequestId string, reviewerId string) error {
	tx, pool, hasTx := s.getTx(ctx)

	sql := `
		INSERT INTO pr_reviewers (pull_request_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (pull_request_id, user_id) DO NOTHING
	`

	var err error
	if hasTx {
		_, err = tx.Exec(ctx, sql, pullRequestId, reviewerId)
	} else {
		_, err = pool.Exec(ctx, sql, pullRequestId, reviewerId)
	}

	if err != nil {
		return storagePR.MapPGError(err)
	}

	return nil
}

func (s *Storage) GetPullRequestsByReviewer(ctx context.Context, reviewerId string) ([]*pullrequest.Model, error) {
	tx, pool, hasTx := s.getTx(ctx)

	query := `
		SELECT 
			pr.id,
			pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			pr.status,
			pr.created_at,
			pr.merged_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`

	var rows pgx.Rows
	var err error

	if hasTx {
		rows, err = tx.Query(ctx, query, reviewerId)
	} else {
		rows, err = pool.Query(ctx, query, reviewerId)
	}

	if err != nil {
		return nil, storagePR.MapPGError(err)
	}
	defer rows.Close()

	var prs []*pullrequest.Model
	for rows.Next() {
		var entity storagePR.Entity
		err := rows.Scan(
			&entity.ID,
			&entity.PullRequestId,
			&entity.PullRequestName,
			&entity.AuthorId,
			&entity.Status,
			&entity.CreatedAt,
			&entity.MergedAt,
		)
		if err != nil {
			return nil, storagePR.MapPGError(err)
		}

		reviewersQuery := `
			SELECT user_id 
			FROM pr_reviewers 
			WHERE pull_request_id = $1
			ORDER BY user_id
		`

		var reviewerRows pgx.Rows
		if hasTx {
			reviewerRows, err = tx.Query(ctx, reviewersQuery, entity.PullRequestId)
		} else {
			reviewerRows, err = pool.Query(ctx, reviewersQuery, entity.PullRequestId)
		}

		if err != nil {
			return nil, storagePR.MapPGError(err)
		}

		var reviewers []string
		for reviewerRows.Next() {
			var reviewerId string
			if err := reviewerRows.Scan(&reviewerId); err != nil {
				reviewerRows.Close()
				return nil, storagePR.MapPGError(err)
			}
			reviewers = append(reviewers, reviewerId)
		}
		reviewerRows.Close()

		prs = append(prs, storagePR.ToDomain(&entity, reviewers))
	}

	if err = rows.Err(); err != nil {
		return nil, storagePR.MapPGError(err)
	}

	return prs, nil
}

func (s *Storage) MergePullRequest(ctx context.Context, pullRequestId string) (*pullrequest.Model, error) {
	tx, pool, hasTx := s.getTx(ctx)

	query := `
		UPDATE pull_requests 
		SET status = 'MERGED', merged_at = NOW()
		WHERE pull_request_id = $1 AND status != 'MERGED'
		RETURNING id, pull_request_id, pull_request_name, author_id, status, created_at, merged_at
	`

	var entity storagePR.Entity
	var err error

	if hasTx {
		err = tx.QueryRow(ctx, query, pullRequestId).Scan(
			&entity.ID,
			&entity.PullRequestId,
			&entity.PullRequestName,
			&entity.AuthorId,
			&entity.Status,
			&entity.CreatedAt,
			&entity.MergedAt,
		)
	} else {
		err = pool.QueryRow(ctx, query, pullRequestId).Scan(
			&entity.ID,
			&entity.PullRequestId,
			&entity.PullRequestName,
			&entity.AuthorId,
			&entity.Status,
			&entity.CreatedAt,
			&entity.MergedAt,
		)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return s.GetPullRequestById(ctx, pullRequestId)
		}
		return nil, storagePR.MapPGError(err)
	}

	reviewersQuery := `
		SELECT user_id 
		FROM pr_reviewers 
		WHERE pull_request_id = $1
		ORDER BY user_id
	`

	var rows pgx.Rows
	if hasTx {
		rows, err = tx.Query(ctx, reviewersQuery, pullRequestId)
	} else {
		rows, err = pool.Query(ctx, reviewersQuery, pullRequestId)
	}

	if err != nil {
		return nil, storagePR.MapPGError(err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerId string
		if err := rows.Scan(&reviewerId); err != nil {
			return nil, storagePR.MapPGError(err)
		}
		reviewers = append(reviewers, reviewerId)
	}

	return storagePR.ToDomain(&entity, reviewers), nil
}

func (s *Storage) RemoveReviewer(ctx context.Context, pullRequestId string, reviewerId string) error {
	tx, pool, hasTx := s.getTx(ctx)

	sql := `
		DELETE FROM pr_reviewers 
		WHERE pull_request_id = $1 AND user_id = $2
	`

	var err error
	if hasTx {
		_, err = tx.Exec(ctx, sql, pullRequestId, reviewerId)
	} else {
		_, err = pool.Exec(ctx, sql, pullRequestId, reviewerId)
	}

	if err != nil {
		return storagePR.MapPGError(err)
	}

	return nil
}

