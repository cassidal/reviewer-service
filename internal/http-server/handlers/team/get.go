package team

import (
	"log/slog"
	"net/http"
	"reviewer-service/internal/domain/team"
	"reviewer-service/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func Get(log *slog.Logger, teamRepo team.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.team.Get"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		teamName := r.URL.Query().Get("team_name")
		if teamName == "" {
			responseError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "team_name parameter is required")
			return
		}

		teamModel, err := team.GetTeamByName(r.Context(), log, teamRepo, teamName)
		if err != nil {
			log.Error("failed to get team", slog.String("team_name", teamName))

			if storageErr, ok := storage.IsError(err); ok {
				statusCode := getStatusCodeForError(storageErr.Code)
				responseError(w, r, statusCode, storageErr.Code, storageErr.Message)
			} else {
				responseError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			}
			return
		}

		render.JSON(w, r, toDto(teamModel))
	}
}
