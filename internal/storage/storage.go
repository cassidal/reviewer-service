package storage

import "errors"

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

var (
	ErrTeamNotFound          = &Error{Code: "NOT_FOUND", Message: "team not found"}
	ErrTeamNameAlreadyExists = &Error{Code: "TEAM_EXISTS", Message: "team_name already exists"}

	ErrUserNotFound        = &Error{Code: "NOT_FOUND", Message: "user not found"}
	ErrUserIdAlreadyExists = &Error{Code: "USER_EXISTS", Message: "user_id already exists"}

	ErrPullRequestNotFound      = &Error{Code: "NOT_FOUND", Message: "pull request not found"}
	ErrPullRequestAlreadyExists = &Error{Code: "PR_EXISTS", Message: "PR id already exists"}
	ErrPullRequestMerged        = &Error{Code: "PR_MERGED", Message: "cannot reassign on merged PR"}
	ErrReviewerNotAssigned      = &Error{Code: "NOT_ASSIGNED", Message: "reviewer is not assigned to this PR"}
	ErrNoReplacementCandidate   = &Error{Code: "NO_CANDIDATE", Message: "no active replacement candidate in team"}
)

func IsError(err error) (*Error, bool) {
	var storageErr *Error
	if errors.As(err, &storageErr) {
		return storageErr, true
	}
	return nil, false
}
