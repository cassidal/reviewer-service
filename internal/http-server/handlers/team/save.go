package team

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"reviewer-service/internal/domain/team"
	logUtil "reviewer-service/internal/lib/logger/slog"
	"reviewer-service/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

func Save(log *slog.Logger, txManager team.TransactionManager, repo team.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.team.Save"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req SaveRequest
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			responseError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "request body is empty")
			return
		}
		if err != nil {
			log.Error("failed to decode request body", logUtil.Err(err))
			responseError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "failed to decode request")
			return
		}

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", logUtil.Err(err))
			responseError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", validationErrorResponse(validateErr))
			return
		}

		savedTeam, err := team.SaveTeam(r.Context(), log, txManager, repo, toDomain(&req.Team))
		if err != nil {
			log.Error("failed to save team", logUtil.Err(err))

			if storageErr, ok := storage.IsError(err); ok {
				statusCode := getStatusCodeForError(storageErr.Code)
				responseError(w, r, statusCode, storageErr.Code, storageErr.Message)
			} else {
				responseError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			}
			return
		}

		responseOK(w, r, 201, toDto(savedTeam))
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, statusCode int, team *DTO) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, SaveResponse{
		Team: team,
	})
}
