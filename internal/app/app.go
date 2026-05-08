package app

import (
	"database/sql"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/Aiya594/appointment-services/internal/cache"
	"github.com/Aiya594/appointment-services/internal/client"
	cfg "github.com/Aiya594/appointment-services/internal/config"
	natspub "github.com/Aiya594/appointment-services/internal/event"
	"github.com/Aiya594/appointment-services/internal/middleware"
	"github.com/Aiya594/appointment-services/internal/repository"
	grpcAppoi "github.com/Aiya594/appointment-services/internal/transport/grpc"
	usecase "github.com/Aiya594/appointment-services/internal/use-case"
	"github.com/Aiya594/appointment-services/proto"
	"github.com/golang-migrate/migrate/v4"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type App struct {
	grpcServ    *grpc.Server
	logger      *slog.Logger
	pub         *natspub.Publisher
	conn        *grpc.ClientConn
	db          *sql.DB
	redisClient *redis.Client
}

func NewApp(cfg *cfg.Config) (*App, error) {
	runMigrations(cfg.ConnStrDB)

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
	redisClient := cache.NewRedisClient(logger)
	var cacheRepo cache.CacheRepository
	if redisClient != nil {
		cacheRepo = cache.NewRedisCacheRepository(redisClient, logger)
	} else {
		cacheRepo = cache.NewNoop()
	}

	uc := usecase.NewAppointmentUseCase(repo, logger, client, publisher, cacheRepo)
	handler := grpcAppoi.NewAppointmentServer(logger, uc)

	rateLimiter := middleware.RateLimiterInterceptor(redisClient, logger)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(rateLimiter),
	)
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
	err = a.grpcServ.Serve(lis)
	if err != nil {
		a.logger.Info("gRPC server stopped", "reason", err)
		return nil
	}
	return nil
}

func (a *App) Close() {
	a.pub.Close()
	a.conn.Close()
	a.db.Close()
}

func (a *App) Stop() {
	a.grpcServ.GracefulStop()
}

func runMigrations(dbURL string) {
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
