package model

import "time"

type Appointment struct {
	ID          string
	Title       string
	Description string
	DoctorID    string
	Status      Status // define a custom Status type
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (a *Appointment) ValidateStatusTransition(to Status) bool {

	allowed := validTransitions[a.Status]
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
