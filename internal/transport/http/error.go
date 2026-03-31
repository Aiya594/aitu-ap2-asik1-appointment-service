package httpappoi

import (
	"errors"
	"net/http"

	"github.com/Aiya594/appointment-services/internal/client"
	"github.com/Aiya594/appointment-services/internal/repository"
	usecase "github.com/Aiya594/appointment-services/internal/use-case"
)

func parseError(err error) int {
	switch {
	case errors.Is(err, repository.ErrAppointmentNotFound):
		return http.StatusNotFound

	case errors.Is(err, client.ErrDocNotFound):
		return http.StatusNotFound

	case errors.Is(err, usecase.ErrEmptyFields):
		return http.StatusBadRequest

	case errors.Is(err, usecase.ErrInvalidStatus):
		return http.StatusBadRequest

	case errors.Is(err, usecase.ErrInvalidStatusTransition):
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}
