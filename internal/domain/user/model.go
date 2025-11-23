package user

type Model struct {
	ID                int64
	UserId            string
	Username          string
	TeamName          string
	IsActive          bool
	PullRequestShorts []*PullRequestShort
}

type PullRequestShort struct {
	PullRequestId   string
	PullRequestName string
	AuthorId        string
	Status          string
}
