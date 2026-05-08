package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/Aiya594/appointment-services/internal/model"
	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	GetAppointment(ctx context.Context, id string) (*model.Appointment, error)
	SetAppointment(ctx context.Context, ap *model.Appointment) error
	GetAppointmentList(ctx context.Context) ([]*model.Appointment, error)
	SetAppointmentList(ctx context.Context, aps []*model.Appointment) error
	InvalidateAppointment(ctx context.Context, id string) error
	InvalidateAppointmentList(ctx context.Context) error
}

type RedisCacheRepository struct {
	client *redis.Client
	ttl    time.Duration
	logger *slog.Logger
}

func NewRedisCacheRepository(client *redis.Client, logger *slog.Logger) CacheRepository {
	ttlSec := 60
	if v := os.Getenv("CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			ttlSec = n
		}
	}
	return &RedisCacheRepository{client: client, ttl: time.Duration(ttlSec) * time.Second, logger: logger}
}

func apptKey(id string) string          { return fmt.Sprintf("appointment:%s", id) }
const apptListKey = "appointments:list"

func (r *RedisCacheRepository) GetAppointment(ctx context.Context, id string) (*model.Appointment, error) {
	val, err := r.client.Get(ctx, apptKey(id)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var ap model.Appointment
	if err := json.Unmarshal([]byte(val), &ap); err != nil {
		return nil, err
	}
	return &ap, nil
}

func (r *RedisCacheRepository) SetAppointment(ctx context.Context, ap *model.Appointment) error {
	data, err := json.Marshal(ap)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, apptKey(ap.ID), data, r.ttl).Err()
}

func (r *RedisCacheRepository) GetAppointmentList(ctx context.Context) ([]*model.Appointment, error) {
	val, err := r.client.Get(ctx, apptListKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var aps []*model.Appointment
	if err := json.Unmarshal([]byte(val), &aps); err != nil {
		return nil, err
	}
	return aps, nil
}

func (r *RedisCacheRepository) SetAppointmentList(ctx context.Context, aps []*model.Appointment) error {
	data, err := json.Marshal(aps)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, apptListKey, data, r.ttl).Err()
}

func (r *RedisCacheRepository) InvalidateAppointment(ctx context.Context, id string) error {
	return r.client.Del(ctx, apptKey(id)).Err()
}

func (r *RedisCacheRepository) InvalidateAppointmentList(ctx context.Context) error {
	return r.client.Del(ctx, apptListKey).Err()
}
