package client

import (
	"context"

	"github.com/Aiya594/appointment-services/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Doctor struct {
	ID             string
	FullName       string
	Specialization string
	Email          string
}

type DoctorGRPC interface {
	GetDoctor(ctx context.Context, id string) (*Doctor, error)
}

type DoctorGrpcClient struct {
	client proto.DoctorServiceClient
}

func NewDoctorGrpcClient(conn *grpc.ClientConn) *DoctorGrpcClient {
	return &DoctorGrpcClient{
		client: proto.NewDoctorServiceClient(conn),
	}
}

func (d *DoctorGrpcClient) GetDoctor(ctx context.Context, id string) (*Doctor, error) {
	req := &proto.GetDoctorRequest{Id: id}
	responce, err := d.client.GetDoctor(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, status.Error(codes.Internal, "unknown error from doctor service")
		}
		return nil, handleError(st)
	}

	doc := &Doctor{
		ID:             responce.Id,
		FullName:       responce.FullName,
		Specialization: responce.Specialization,
		Email:          responce.Email,
	}

	return doc, nil
}

func handleError(st *status.Status) error {
	switch st.Code() {

	case codes.NotFound:
		return status.Error(codes.NotFound, st.Message())

	case codes.InvalidArgument:
		return status.Error(codes.InvalidArgument, st.Message())

	case codes.Unavailable:
		return status.Error(codes.Unavailable, "doctor service unavailable")

	default:
		return status.Error(codes.Internal, "doctor service error")
	}
}
