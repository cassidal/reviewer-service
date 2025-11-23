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

type MergeRequest struct {
	PullRequestId string `json:"pull_request_id" validate:"required"`
}

type MergeResponse struct {
	PR    *PullRequestResponse `json:"pr,omitempty"`
	Error *ErrorResponse       `json:"error,omitempty"`
}

func Merge(log *slog.Logger, txManager pullrequest.TransactionManager, repo pullrequest.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.pullrequest.Merge"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req MergeRequest
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			responseErrorMerge(w, r, http.StatusBadRequest, "INVALID_REQUEST", "request body is empty")
			return
		}
		if err != nil {
			log.Error("failed to decode request body", slog.String("error", err.Error()))
			responseErrorMerge(w, r, http.StatusBadRequest, "INVALID_REQUEST", "failed to decode request")
			return
		}

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", slog.String("error", err.Error()))
			responseErrorMerge(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request")
			return
		}

		mergedPR, err := pullrequest.MergePullRequest(r.Context(), log, txManager, repo, req.PullRequestId)
		if err != nil {
			log.Error("failed to merge pull request", slog.String("error", err.Error()))

			if storageErr, ok := storage.IsError(err); ok {
				statusCode := getStatusCodeForError(storageErr.Code)
				responseErrorMerge(w, r, statusCode, storageErr.Code, storageErr.Message)
			} else {
				responseErrorMerge(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			}
			return
		}

		render.JSON(w, r, MergeResponse{
			PR: toDto(mergedPR),
		})
	}
}

func responseErrorMerge(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, MergeResponse{
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
	})
}
