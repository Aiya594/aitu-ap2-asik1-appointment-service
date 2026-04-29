package cfg

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port         string
	DoctorClient string
	ConnStrDB    string
	NatsURL      string
}

func LoadCfg() *Config {
	return &Config{Port: os.Getenv("APP_PORT"),
		DoctorClient: os.Getenv("DOCTOR_SCV_URL"),
		ConnStrDB:    os.Getenv("DATABASE_URL"),
		NatsURL:      os.Getenv("NATS_URL"),
	}
}

func (c *Config) Connect() (*sql.DB, error) {
	db, err := sql.Open("postgres", c.ConnStrDB)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return db, nil
}
