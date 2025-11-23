package pullrequest

import (
	"reviewer-service/internal/domain/pullrequest"
	"time"
)

func toDomain(dto *CreateRequest) *pullrequest.Model {
	now := time.Now()
	return &pullrequest.Model{
		PullRequestId:     dto.PullRequestId,
		PullRequestName:   dto.PullRequestName,
		AuthorId:          dto.AuthorId,
		Status:            "OPEN",
		AssignedReviewers: []string{},
		CreatedAt:         &now,
		MergedAt:          nil,
	}
}

func toDto(pr *pullrequest.Model) *PullRequestResponse {
	var assignedReviewers []string
	if pr.AssignedReviewers != nil {
		assignedReviewers = pr.AssignedReviewers
	} else {
		assignedReviewers = []string{}
	}
	return &PullRequestResponse{
		PullRequestId:     pr.PullRequestId,
		PullRequestName:   pr.PullRequestName,
		AuthorId:          pr.AuthorId,
		Status:            pr.Status,
		AssignedReviewers: assignedReviewers,
		MergedAt:          pr.MergedAt,
	}
}
