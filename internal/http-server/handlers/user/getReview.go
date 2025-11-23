package user

import (
	"log/slog"
	"net/http"
	"reviewer-service/internal/domain/pullrequest"
	"reviewer-service/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func GetReview(log *slog.Logger, repo pullrequest.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.GetReview"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userId := r.URL.Query().Get("user_id")
		if userId == "" {
			responseErrorGetReview(w, r, http.StatusBadRequest, "INVALID_REQUEST", "user_id parameter is required")
			return
		}

		prs, err := pullrequest.GetPullRequests(r.Context(), log, repo, userId)
		if err != nil {
			log.Error("failed to get pull requests", slog.String("user_id", userId), slog.String("error", err.Error()))

			if storageErr, ok := storage.IsError(err); ok {
				statusCode := getStatusCodeForError(storageErr.Code)
				responseErrorGetReview(w, r, statusCode, storageErr.Code, storageErr.Message)
			} else {
				responseErrorGetReview(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			}
			return
		}

		render.JSON(w, r, GetReviewResponse{
			UserId:       userId,
			PullRequests: toPullRequestShortDtos(prs),
		})
	}
}

func responseErrorGetReview(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, GetReviewResponse{
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
	})
}
