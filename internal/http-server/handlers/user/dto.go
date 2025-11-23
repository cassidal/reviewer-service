package user

import (
	"net/http"

	"github.com/go-chi/render"
)

type SetIsActiveRequest struct {
	UserId   string `json:"user_id" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type SetIsActiveResponse struct {
	User  *UserResponse  `json:"user,omitempty"`
	Error *ErrorResponse `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type GetReviewResponse struct {
	UserId       string                      `json:"user_id"`
	PullRequests []*PullRequestShortResponse `json:"pull_requests"`
	Error        *ErrorResponse              `json:"error,omitempty"`
}

type PullRequestShortResponse struct {
	PullRequestId   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorId        string `json:"author_id"`
	Status          string `json:"status"`
}

func responseError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, SetIsActiveResponse{
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
	})
}

func getStatusCodeForError(errorCode string) int {
	switch errorCode {
	case "USER_EXISTS":
		return http.StatusBadRequest
	case "NOT_FOUND":
		return http.StatusNotFound
	case "VALIDATION_ERROR", "INVALID_REQUEST":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
