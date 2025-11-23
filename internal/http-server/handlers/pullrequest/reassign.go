package pullrequest

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"reviewer-service/internal/domain/pullrequest"
	"reviewer-service/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type ReassignRequest struct {
	PullRequestId string `json:"pull_request_id" validate:"required"`
	OldUserId     string `json:"old_reviewer_id" validate:"required"`
}

type ReassignResponse struct {
	PR         *PullRequestResponse `json:"pr,omitempty"`
	ReplacedBy string               `json:"replaced_by,omitempty"`
	Error      *ErrorResponse       `json:"error,omitempty"`
}

func Reassign(log *slog.Logger, txManager pullrequest.TransactionManager, repo pullrequest.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.pullrequest.Reassign"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req ReassignRequest
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			responseErrorReassign(w, r, http.StatusBadRequest, "INVALID_REQUEST", "request body is empty")
			return
		}
		if err != nil {
			log.Error("failed to decode request body", slog.String("error", err.Error()))
			responseErrorReassign(w, r, http.StatusBadRequest, "INVALID_REQUEST", "failed to decode request")
			return
		}

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", slog.String("error", err.Error()))
			responseErrorReassign(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request")
			return
		}

		updatedPR, newReviewerId, err := pullrequest.ReassignReviewer(r.Context(), log, txManager, repo, req.PullRequestId, req.OldUserId)
		if err != nil {
			log.Error("failed to reassign reviewer", slog.String("error", err.Error()))

			if storageErr, ok := storage.IsError(err); ok {
				statusCode := getStatusCodeForErrorReassign(storageErr.Code)
				responseErrorReassign(w, r, statusCode, storageErr.Code, storageErr.Message)
			} else {
				responseErrorReassign(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			}
			return
		}

		response := ReassignResponse{
			PR: toDto(updatedPR),
		}
		if newReviewerId != "" {
			response.ReplacedBy = newReviewerId
		}

		render.JSON(w, r, response)
	}
}

func responseErrorReassign(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, ReassignResponse{
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
	})
}

func getStatusCodeForErrorReassign(errorCode string) int {
	switch errorCode {
	case "PR_EXISTS":
		return http.StatusConflict
	case "PR_MERGED", "NOT_ASSIGNED", "NO_CANDIDATE":
		return http.StatusConflict
	case "NOT_FOUND":
		return http.StatusNotFound
	case "VALIDATION_ERROR", "INVALID_REQUEST":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
