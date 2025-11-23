package pullrequest

import "time"

type Entity struct {
	ID              int64      `db:"id"`
	PullRequestId   string     `db:"pull_request_id"`
	PullRequestName string     `db:"pull_request_name"`
	AuthorId        string     `db:"author_id"`
	Status          string     `db:"status"`
	CreatedAt       time.Time  `db:"created_at"`
	MergedAt        *time.Time `db:"merged_at"`
}

