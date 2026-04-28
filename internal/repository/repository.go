package repository

import (
	"database/sql"
	"errors"

	"github.com/Aiya594/appointment-services/internal/model"
)

type AppointmentRepository interface {
	Create(ap *model.Appointment) error
	GetById(id string) (*model.Appointment, error)
	List() ([]*model.Appointment, error)
	Update(ap *model.Appointment) error
}

type PostgresAppointmentRepository struct {
	db *sql.DB
}

func NewAppointmentRepo(db *sql.DB) AppointmentRepository {
	return &PostgresAppointmentRepository{db: db}
}

func (r *PostgresAppointmentRepository) Create(ap *model.Appointment) error {
	query := `
		INSERT INTO appointments (id, title, description, doctor_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(
		query,
		ap.ID,
		ap.Title,
		ap.Description,
		ap.DoctorID,
		ap.Status,
		ap.CreatedAt,
		ap.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresAppointmentRepository) GetById(id string) (*model.Appointment, error) {
	query := `
		SELECT id, title, description, doctor_id, status, created_at, updated_at
		FROM appointments
		WHERE id = $1
	`

	var ap model.Appointment

	err := r.db.QueryRow(query, id).Scan(
		&ap.ID,
		&ap.Title,
		&ap.Description,
		&ap.DoctorID,
		&ap.Status,
		&ap.CreatedAt,
		&ap.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}

	return &ap, nil
}

func (r *PostgresAppointmentRepository) List() ([]*model.Appointment, error) {
	query := `
		SELECT id, title, description, doctor_id, status, created_at, updated_at
		FROM appointments
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.Appointment

	for rows.Next() {
		var ap model.Appointment

		err := rows.Scan(
			&ap.ID,
			&ap.Title,
			&ap.Description,
			&ap.DoctorID,
			&ap.Status,
			&ap.CreatedAt,
			&ap.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, &ap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresAppointmentRepository) Update(ap *model.Appointment) error {
	query := `
		UPDATE appointments
		SET status = $5,
		    updated_at = $6
		WHERE id = $1
	`

	res, err := r.db.Exec(
		query,
		ap.Status,
		ap.UpdatedAt,
		ap.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAppointmentNotFound
	}

	return nil
}
