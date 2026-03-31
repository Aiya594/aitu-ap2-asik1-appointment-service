package app

import (
	"log/slog"
	"os"

	"github.com/Aiya594/appointment-services/internal/client"
	cfg "github.com/Aiya594/appointment-services/internal/config"
	"github.com/Aiya594/appointment-services/internal/repository"
	httpappoi "github.com/Aiya594/appointment-services/internal/transport/http"
	usecase "github.com/Aiya594/appointment-services/internal/use-case"
	"github.com/gin-gonic/gin"
)

type App struct {
	router *gin.Engine
}

func NewApp(cfg *cfg.Config) *App {
	repo := repository.NewAppointmentRepo()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	docClient := client.NewDoctorClient(cfg.DoctorClient)

	usecase := usecase.NewAppointmentUseCase(repo, logger, docClient)

	h := httpappoi.NewAppointmentHandler(usecase)

	r := gin.Default()

	httpappoi.RegisterRoutes(r, h)

	return &App{router: r}
}

func (a *App) Run(port string) {
	a.router.Run(":" + port)
}
