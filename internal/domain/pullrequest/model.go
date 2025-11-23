package pullrequest

import "time"

type Model struct {
	ID                int64
	PullRequestId     string
	PullRequestName   string
	AuthorId          string
	Status            string
	AssignedReviewers []string
	CreatedAt         *time.Time
	MergedAt          *time.Time
}
