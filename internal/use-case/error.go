package usecase

import "errors"

var (
	ErrEmptyFields             = errors.New("empty fields")
	ErrInvalidStatus           = errors.New("invalid status")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrAppointmentNotFound     = errors.New("appointment not found")
)
