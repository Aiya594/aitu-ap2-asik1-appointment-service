package model

import "time"

type AppointmentCreated struct {
	Event_type  string
	Occurred_at time.Time
	ID          string
	Title       string
	Doctor_id   string
	Status      string
}

type AppointmentStatusUpdated struct {
	Event_type  string
	Occurred_at time.Time
	ID          string
	Old_status  string
	New_status  string
}
