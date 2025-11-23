package user

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"reviewer-service/internal/domain/user"
	"reviewer-service/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

func SetIsActive(log *slog.Logger, txManager user.TransactionManager, repo user.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.SetIsActive"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req SetIsActiveRequest
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			responseError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "request body is empty")
			return
		}
		if err != nil {
			log.Error("failed to decode request body", slog.String("error", err.Error()))
			responseError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "failed to decode request")
			return
		}

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", slog.String("error", err.Error()))
			responseError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request")
			return
		}

		updatedUser, err := user.SetUserIsActive(r.Context(), log, txManager, repo, req.UserId, req.IsActive)
		if err != nil {
			log.Error("failed to update user", slog.String("user_id", req.UserId))

			if storageErr, ok := storage.IsError(err); ok {
				statusCode := getStatusCodeForError(storageErr.Code)
				responseError(w, r, statusCode, storageErr.Code, storageErr.Message)
			} else {
				responseError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			}
			return
		}

		render.JSON(w, r, SetIsActiveResponse{
			User: toDto(updatedUser),
		})
	}
}
