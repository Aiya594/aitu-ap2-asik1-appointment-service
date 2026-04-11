package grpcAppoi

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Aiya594/appointment-services/internal/model"
	usecase "github.com/Aiya594/appointment-services/internal/use-case"
	"github.com/Aiya594/appointment-services/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppointmentGRPCServer struct {
	proto.UnimplementedAppointmentServiceServer
	logger *slog.Logger
	svc    usecase.AppointmentUseCase
}

func NewAppointmentServer(logger *slog.Logger, svc usecase.AppointmentUseCase) *AppointmentGRPCServer {
	return &AppointmentGRPCServer{
		logger: logger,
		svc:    svc,
	}
}

func (h *AppointmentGRPCServer) CreateAppointment(ctx context.Context,
	in *proto.CreateAppointmentRequest) (*proto.AppointmentResponse, error) {
	h.logger.Info("CreateAppointment called",
		slog.String("doctor_id", in.GetDoctorId()),
	)

	app, err := h.svc.CreateAppointment(ctx, in.GetTitle(), in.GetDescription(), in.GetDoctorId())
	if err != nil {
		h.logger.Error("failed to create appointment", slog.String("err", err.Error()))

		return nil, mapGRPCError(err)
	}

	resp := &proto.AppointmentResponse{
		Id:          app.ID,
		Title:       app.Title,
		Description: app.Description,
		DoctorId:    app.DoctorID,
		Status:      string(app.Status),
		CreatedAt:   app.CreatedAt.String(),
		UpdatedAt:   app.UpdatedAt.String(),
	}
	return resp, nil

}

func (h *AppointmentGRPCServer) GetAppointment(ctx context.Context,
	in *proto.GetAppointmentRequest) (*proto.AppointmentResponse, error) {
	h.logger.Info("GetAppointment", slog.String("id", in.GetId()))

	app, err := h.svc.GetByID(in.GetId())
	if err != nil {
		h.logger.Error("failed to get appointment", slog.String("err", err.Error()))
		return nil, mapGRPCError(err)
	}

	resp := &proto.AppointmentResponse{
		Id:          app.ID,
		Title:       app.Title,
		Description: app.Description,
		DoctorId:    app.DoctorID,
		Status:      string(app.Status),
		CreatedAt:   app.CreatedAt.String(),
		UpdatedAt:   app.UpdatedAt.String(),
	}
	return resp, nil

}

func (h *AppointmentGRPCServer) ListAppointments(ctx context.Context,
	in *proto.ListAppointmentsRequest) (*proto.ListAppointmentsResponse, error) {
	h.logger.Info("ListAppointments")

	apps, err := h.svc.GetAll()
	if err != nil {
		h.logger.Error("failed to list appointments", slog.String("err", err.Error()))
		return nil, mapGRPCError(err)
	}

	res := &proto.ListAppointmentsResponse{
		Appointments: make([]*proto.AppointmentResponse, 0, len(apps)),
	}

	for _, app := range apps {
		resp := &proto.AppointmentResponse{
			Id:          app.ID,
			Title:       app.Title,
			Description: app.Description,
			DoctorId:    app.DoctorID,
			Status:      string(app.Status),
			CreatedAt:   app.CreatedAt.String(),
			UpdatedAt:   app.UpdatedAt.String(),
		}
		res.Appointments = append(res.Appointments, resp)
	}

	return res, nil
}

func (h *AppointmentGRPCServer) UpdateAppointmentStatus(ctx context.Context,
	in *proto.UpdateStatusRequest) (*proto.AppointmentResponse, error) {
	h.logger.Info("UpdateAppointmentStatus",
		slog.String("id", in.GetId()),
		slog.String("status", in.GetStatus()),
	)

	app, err := h.svc.UpdateStatus(in.GetId(), model.Status(in.GetStatus()))
	if err != nil {
		h.logger.Error("failed to update status", slog.String("err", err.Error()))
		return nil, mapGRPCError(err)
	}

	resp := &proto.AppointmentResponse{
		Id:          app.ID,
		Title:       app.Title,
		Description: app.Description,
		DoctorId:    app.DoctorID,
		Status:      string(app.Status),
		CreatedAt:   app.CreatedAt.String(),
		UpdatedAt:   app.UpdatedAt.String(),
	}
	return resp, nil
}

func mapGRPCError(err error) error {
	// from doctor-service
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.NotFound:
			return status.Error(codes.NotFound, st.Message())

		case codes.InvalidArgument:
			return status.Error(codes.InvalidArgument, st.Message())

		case codes.Unavailable:
			return status.Error(codes.Unavailable, "doctor service unavailable")

		default:
			return status.Error(codes.Internal, "external service error")
		}
	}

	switch {
	case errors.Is(err, usecase.ErrAppointmentNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, usecase.ErrEmptyFields):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, usecase.ErrInvalidStatus):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, usecase.ErrInvalidStatusTransition):
		return status.Error(codes.FailedPrecondition, err.Error())

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
