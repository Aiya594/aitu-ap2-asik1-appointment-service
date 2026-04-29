package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Aiya594/appointment-services/internal/app"
	cfg "github.com/Aiya594/appointment-services/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	config := cfg.LoadCfg()

	app, err := app.NewApp(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Run(config.Port); err != nil {
			log.Fatal("server error:", err)
		}
	}()

	log.Println("server started on port", config.Port)

	// Wait for signal
	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop gRPC server
	done := make(chan struct{})
	go func() {
		app.Stop()
		close(done)
	}()
	select {
	case <-done:
		log.Println("gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("graceful shutdown timeout exceeded")
	}

	// Close resources
	app.Close()

	log.Println("application shutdown complete")
}
