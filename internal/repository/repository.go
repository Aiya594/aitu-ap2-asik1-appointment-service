package repository

import (
	"sync"

	"github.com/Aiya594/appointment-services/internal/model"
)

type AppointmentRepository interface {
	Create(ap *model.Appointment) error
	GetById(id string) (*model.Appointment, error)
	List() ([]*model.Appointment, error)
	Update(id string, status model.Status) error
}

type InMemoryAppointmentStorage struct {
	mu       sync.RWMutex
	appoints map[string]*model.Appointment
}

func NewAppointmentRepo() AppointmentRepository {
	return &InMemoryAppointmentStorage{
		appoints: make(map[string]*model.Appointment),
	}
}

func (a *InMemoryAppointmentStorage) Create(ap *model.Appointment) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.appoints[ap.ID] = ap
	return nil
}

func (a *InMemoryAppointmentStorage) GetById(id string) (*model.Appointment, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ap, ok := a.appoints[id]
	if !ok {
		return nil, ErrAppointmentNotFound
	}
	return ap, nil
}

func (a *InMemoryAppointmentStorage) List() ([]*model.Appointment, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	list := make([]*model.Appointment, 0, len(a.appoints))
	for _, ap := range a.appoints {
		list = append(list, ap)
	}
	return list, nil
}

func (a *InMemoryAppointmentStorage) Update(id string, status model.Status) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	ap, ok := a.appoints[id]
	if !ok {
		return ErrAppointmentNotFound
	}

	ap.Status = status
	return nil
}
