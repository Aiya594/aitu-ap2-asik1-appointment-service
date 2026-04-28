package app

import (
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/Aiya594/appointment-services/internal/client"
	cfg "github.com/Aiya594/appointment-services/internal/config"
	"github.com/Aiya594/appointment-services/internal/repository"
	grpcAppoi "github.com/Aiya594/appointment-services/internal/transport/grpc"
	usecase "github.com/Aiya594/appointment-services/internal/use-case"
	"github.com/Aiya594/appointment-services/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	grpcSrev *grpc.Server
	logger   *slog.Logger
}

func NewApp(cfg *cfg.Config) *App {

	db, err := cfg.Connect()
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewAppointmentRepo(db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	conn, err := grpc.Dial(
		cfg.DoctorClient,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := client.NewDoctorGrpcClient(conn)
	uc := usecase.NewAppointmentUseCase(repo, logger, client)
	handler := grpcAppoi.NewAppointmentServer(logger, uc)

	grpcServer := grpc.NewServer()

	proto.RegisterAppointmentServiceServer(grpcServer, handler)

	return &App{
		grpcSrev: grpcServer,
		logger:   logger,
	}

}

func (a *App) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	err = a.grpcSrev.Serve(lis)
	if err != nil {
		return err
	}

	a.logger.Info("gRPC server started", "port", port)

	return nil

}
