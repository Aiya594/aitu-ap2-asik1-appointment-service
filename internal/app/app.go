package app

import (
	"database/sql"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/Aiya594/appointment-services/internal/client"
	cfg "github.com/Aiya594/appointment-services/internal/config"
	natspub "github.com/Aiya594/appointment-services/internal/event"
	"github.com/Aiya594/appointment-services/internal/repository"
	grpcAppoi "github.com/Aiya594/appointment-services/internal/transport/grpc"
	usecase "github.com/Aiya594/appointment-services/internal/use-case"
	"github.com/Aiya594/appointment-services/proto"
	"github.com/golang-migrate/migrate/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	grpcServ *grpc.Server
	logger   *slog.Logger
	pub      *natspub.Publisher
	conn     *grpc.ClientConn
	db       *sql.DB
}

func NewApp(cfg *cfg.Config) (*App, error) {
	runMigrations()

	db, err := cfg.Connect()
	if err != nil {
		return nil, err
	}

	repo := repository.NewAppointmentRepo(db)

	publisher, err := natspub.NewPublisher(cfg.NatsURL)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	conn, err := grpc.Dial(
		cfg.DoctorClient,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := client.NewDoctorGrpcClient(conn)

	uc := usecase.NewAppointmentUseCase(repo, logger, client, publisher)
	handler := grpcAppoi.NewAppointmentServer(logger, uc)

	grpcServer := grpc.NewServer()
	proto.RegisterAppointmentServiceServer(grpcServer, handler)

	return &App{
		grpcServ: grpcServer,
		logger:   logger,
		pub:      publisher,
		conn:     conn,
		db:       db,
	}, nil
}

func (a *App) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	a.logger.Info("gRPC server starting", "port", port)
	return a.grpcServ.Serve(lis)
}

func (a *App) Close() {
	a.pub.Close()
	a.conn.Close()
	a.db.Close()
}

func (a *App) Stop() {
	a.grpcServ.GracefulStop()
}

func runMigrations() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		log.Fatal("migration init error:", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("migration failed:", err)
	}

	log.Println("migrations applied successfully")
}
