package pullrequest

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type CreateRequest struct {
	PullRequestId   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorId        string `json:"author_id" validate:"required"`
}

type PullRequestResponse struct {
	PullRequestId     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorId          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type CreateResponse struct {
	PR    *PullRequestResponse `json:"pr,omitempty"`
	Error *ErrorResponse       `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func responseError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, CreateResponse{
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
	})
}

func getStatusCodeForError(errorCode string) int {
	switch errorCode {
	case "PR_EXISTS":
		return http.StatusConflict
	case "NOT_FOUND":
		return http.StatusNotFound
	case "VALIDATION_ERROR", "INVALID_REQUEST":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
