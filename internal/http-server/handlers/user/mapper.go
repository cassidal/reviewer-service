package user

import (
	"reviewer-service/internal/domain/pullrequest"
	"reviewer-service/internal/domain/user"
)

func toDto(userModel *user.Model) *UserResponse {
	return &UserResponse{
		UserId:   userModel.UserId,
		Username: userModel.Username,
		TeamName: userModel.TeamName,
		IsActive: userModel.IsActive,
	}
}

func toPullRequestShortDtos(prShorts []*pullrequest.Model) []*PullRequestShortResponse {
	result := make([]*PullRequestShortResponse, 0, len(prShorts))
	for _, pr := range prShorts {
		result = append(result, &PullRequestShortResponse{
			PullRequestId:   pr.PullRequestId,
			PullRequestName: pr.PullRequestName,
			AuthorId:        pr.AuthorId,
			Status:          pr.Status,
		})
	}
	return result
}
