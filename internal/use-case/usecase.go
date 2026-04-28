package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Aiya594/appointment-services/internal/client"
	natspub "github.com/Aiya594/appointment-services/internal/event"
	"github.com/Aiya594/appointment-services/internal/model"
	"github.com/Aiya594/appointment-services/internal/repository"
	"github.com/google/uuid"
)

// Appointment Service Rules
//  title is required.
//  doctor_id is required.
//  The referenced doctor must exist in the Doctor Service (validated over
// REST).
//  status must be one of: new, in_progress, done.
//  Transitioning a status from done back to new is not allowed.

type AppointmentUseCase interface {
	CreateAppointment(ctx context.Context, title, description, doctorID string) (*model.Appointment, error)
	UpdateStatus(id string, stat model.Status) (*model.Appointment, error)
	GetByID(id string) (*model.Appointment, error)
	GetAll() ([]*model.Appointment, error)
}

type AppointmentService struct {
	repo      repository.AppointmentRepository
	logger    *slog.Logger
	client    client.DoctorGRPC
	publisher natspub.EventPublisher
}

func NewAppointmentUseCase(repo repository.AppointmentRepository,
	logger *slog.Logger,
	client client.DoctorGRPC, pub natspub.EventPublisher) AppointmentUseCase {
	return &AppointmentService{
		repo: repo, logger: logger, client: client, publisher: pub,
	}
}

func (a *AppointmentService) CreateAppointment(ctx context.Context, title, description, doctorID string) (*model.Appointment, error) {
	title = strings.TrimSpace(strings.ToLower(title))
	description = strings.TrimSpace(strings.ToLower(description))
	doctorID = strings.TrimSpace(doctorID)

	if title == "" || description == "" || doctorID == "" {
		a.logger.Error("failed create an appointment",
			"error", ErrEmptyFields,
			"title", title,
			"description", description,
			"doctor_id", doctorID)
		return nil, fmt.Errorf("title, description and doctor_id are required:%w", ErrEmptyFields)
	}

	doc, err := a.client.GetDoctor(ctx, doctorID)
	if err != nil {
		a.logger.Error("failed check the doctor",
			"error", err,
			"doctor_id", doctorID,
		)

		return nil, err
	}

	id := uuid.New().String()
	created := time.Now()
	updated := time.Now()
	status := model.StatusNew

	ap := &model.Appointment{
		ID:          id,
		Title:       title,
		Description: description,
		DoctorID:    doc.ID,
		Status:      status,
		CreatedAt:   created,
		UpdatedAt:   updated,
	}

	err = a.repo.Create(ap)
	if err != nil {
		a.logger.Error("failed create an appointment",
			"error", err.Error(),
			"title", title,
			"description", description,
			"doctor_id", doctorID)
		return nil, fmt.Errorf("failed to create an appointment:%w", err)
	}
	a.logger.Info("appointment created", "id", id)

	eventType := model.AppointmentCreatedEventName
	event := model.AppointmentCreated{
		Event_type:  eventType,
		Occurred_at: created,
		ID:          ap.ID,
		Title:       ap.Title,
		Doctor_id:   ap.DoctorID,
		Status:      string(ap.Status),
	}

	data, err := json.Marshal(event)
	if err != nil {
		a.logger.Error("failed matshal event",
			"error", err.Error(),
			"event_type", eventType,
			"id", id,
			"title", title,
			"description", description,
			"doctor_id", doctorID)
		return nil, err
	}
	err = a.publisher.Publish(eventType, data)
	if err != nil {
		a.logger.Error("failed to publish event",
			"error", err.Error(),
			"event_type", eventType,
			"id", id,
			"title", title,
			"description", description,
			"doctor_id", doctorID)
		return nil, err
	}

	return ap, nil

}

func (a *AppointmentService) UpdateStatus(id string, stat model.Status) (*model.Appointment, error) {
	ap, err := a.repo.GetById(id)
	if err != nil {
		return nil, ErrAppointmentNotFound
	}

	if stat != model.StatusNew && stat != model.Done && stat != model.InProgress {
		a.logger.Warn("invalid status provided",
			"appointment_id", id,
			"status", stat,
		)
		return nil, ErrInvalidStatus
	}
	oldStat := ap.Status

	// validate transition
	if !ap.ValidateStatusTransition(stat) {
		a.logger.Warn("invalid status transition",
			"appointment_id", id,
			"from", ap.Status,
			"to", stat,
		)
		return nil, ErrInvalidStatusTransition
	}

	ap.Status = stat
	ap.UpdatedAt = time.Now()
	err = a.repo.Update(ap)
	if err != nil {
		a.logger.Error("failed to update appointment",
			"error", err,
			"appointment_id", id,
		)
		return nil, err
	}
	a.logger.Info("status updated successfully",
		"appointment_id", id,
		"status", stat,
	)

	eventType := model.AppointmentStatusUpdatedEventName
	event := model.AppointmentStatusUpdated{
		Event_type:  eventType,
		Occurred_at: ap.UpdatedAt,
		ID:          ap.ID,
		Old_status:  string(oldStat),
		New_status:  string(stat),
	}
	data, err := json.Marshal(event)
	if err != nil {
		a.logger.Error("failed to marshal event",
			"error", err,
			"event_type", eventType,
			"appointment_id", id,
			"old_status", oldStat,
			"new_stat", stat,
		)
		return nil, err
	}
	err = a.publisher.Publish(eventType, data)
	if err != nil {
		a.logger.Error("failed to publish event",
			"error", err,
			"event_type", eventType,
			"appointment_id", id,
			"old_status", oldStat,
			"new_stat", stat,
		)
		return nil, err
	}

	return ap, nil
}

func (a *AppointmentService) GetByID(id string) (*model.Appointment, error) {
	app, err := a.repo.GetById(id)
	if err != nil {
		return nil, ErrAppointmentNotFound
	}
	return app, nil
}

func (a *AppointmentService) GetAll() ([]*model.Appointment, error) {
	return a.repo.List()
}
