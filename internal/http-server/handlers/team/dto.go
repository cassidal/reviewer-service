package team

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type DTO struct {
	Name    string    `json:"team_name" validate:"required"`
	Members []*Member `json:"members,omitempty" validate:"dive" required:"true"`
}

type Member struct {
	UserId   string `json:"user_id" required:"true"`
	Username string `json:"username" required:"true"`
	IsActive bool   `json:"is_active" required:"true"`
}

type SaveRequest struct {
	Team DTO `json:"team"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SaveResponse struct {
	Team  *DTO           `json:"team,omitempty"`
	Error *ErrorResponse `json:"error,omitempty"`
}

func validationErrorResponse(errs validator.ValidationErrors) string {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return strings.Join(errMsgs, ", ")
}

func responseError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, SaveResponse{
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
	})
}

func getStatusCodeForError(errorCode string) int {
	switch errorCode {
	case "TEAM_EXISTS", "USER_EXISTS":
		return http.StatusBadRequest
	case "NOT_FOUND":
		return http.StatusNotFound
	case "VALIDATION_ERROR", "INVALID_REQUEST":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
