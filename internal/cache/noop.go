package cache

import (
	"context"

	"github.com/Aiya594/appointment-services/internal/model"
)

type NoopCacheRepository struct{}

func NewNoop() CacheRepository { return &NoopCacheRepository{} }

func (n *NoopCacheRepository) GetAppointment(_ context.Context, _ string) (*model.Appointment, error) {
	return nil, nil
}
func (n *NoopCacheRepository) SetAppointment(_ context.Context, _ *model.Appointment) error {
	return nil
}
func (n *NoopCacheRepository) GetAppointmentList(_ context.Context) ([]*model.Appointment, error) {
	return nil, nil
}
func (n *NoopCacheRepository) SetAppointmentList(_ context.Context, _ []*model.Appointment) error {
	return nil
}
func (n *NoopCacheRepository) InvalidateAppointment(_ context.Context, _ string) error { return nil }
func (n *NoopCacheRepository) InvalidateAppointmentList(_ context.Context) error       { return nil }
