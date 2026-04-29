  # Appointment Service - gRPC Medical Scheduling Platform

## 1. Project Overview

The Appointment Service is part of a microservices-based Medical Scheduling Platform that has been migrated from REST to **gRPC** for inter-service and client-to-service communication. This service manages the complete lifecycle of appointments in the system while maintaining Clean Architecture principles and bounded-context boundaries.

- **In-memory storage replaced** with a PostgreSQL-backed repository (`internal/repository/repository.go`).
- **Schema managed via migrations** — no DDL in application code; `golang-migrate` runs `migrations/` automatically on startup.
- **NATS publisher added** — after successful `CreateAppointment` and `UpdateAppointmentStatus` operations the service publishes `appointments.created` and `appointments.status_updated` events to NATS Core Pub/Sub.
- The publisher is injected behind the `EventPublisher` interface, so broker failures never block the gRPC response.
Everything else — domain models, use-case logic, gRPC contracts, Doctor Service gRPC client, and Clean Architecture layering — is unchanged from Assignment 2.
 
---
## Architecture
 
```
gRPC client
     │  CreateAppointment / UpdateAppointmentStatus / GetAppointment / ListAppointments
     ▼
transport/grpc  (thin handler, maps errors to gRPC status codes)
     │
     ▼
use-case        (business rules, status transitions, validation)
     │
     ├──► repository       (PostgreSQL via database/sql + lib/pq)
     │
     ├──► client.DoctorGRPC  (validates doctor exists before creating appointment)
     │
     └──► event.Publisher    (NATS Core, fire-and-forget)
```


### Service Responsibilities

**Appointment Service:**
- Manages appointment CRUD operations via gRPC
- Validates doctor existence by calling Doctor Service via gRPC
- Enforces status transition rules (new → in_progress → done)
- Returns appropriate gRPC status codes for all error conditions

**Doctor Service:**
- Manages doctor CRUD operations via gRPC
- Returns doctor data via `GetDoctor` RPC
- Enforces email uniqueness constraint
- Located in `../aitu-ap2-asik1-doctor-service/` (separate folder)


## Migrations
 
Migration files live in `migrations/`:
 
```
migrations/
├── 000001_create_appointments.up.sql
└── 000001_create_appointments.down.sql
```
 
**Automatic (default):** Migrations run automatically on service startup before the gRPC server accepts requests.
 
**Manual apply with golang-migrate CLI:**
 
```bash
migrate -path migrations -database "$DATABASE_URL" up
```
 
**Manual rollback:**
 
```bash
migrate -path migrations -database "$DATABASE_URL" down 1
```
 
The down migration executes `DROP TABLE IF EXISTS appointments;`.
 
---
 ## Start Instruction

Pull [Notification Service](https://github.com/Aiya594/aitu-ap2-asik3-notification-service) and [Appointment Service](https://github.com/Aiya594/aitu-ap2-asik1-appointment-service) in the same folder with appointment service. Create `docker-compose.yml` according to  `docker-compose.example.yml` and then:

```bash
docker-compose up -d --build
```


## Event Publishing
 
### appointments.created
 
Published after a successful `CreateAppointment`:
 
```json
{
  "event_type":  "appointments.created",
  "occurred_at": "2026-05-01T10:24:01Z",
  "id":          "appt-1",
  "title":       "initial cardiac consultation",
  "doctor_id":   "doc-1",
  "status":      "new"
}
```
 
### appointments.status_updated
 
Published after a successful `UpdateAppointmentStatus`:
 
```json
{
  "event_type":  "appointments.status_updated",
  "occurred_at": "2026-05-01T10:25:10Z",
  "id":          "appt-1",
  "old_status":  "new",
  "new_status":  "in_progress"
}
```

## Database Schema
 
Managed exclusively through migration files:
 
```sql
CREATE TABLE appointments (
  id          TEXT        PRIMARY KEY,
  title       TEXT        NOT NULL,
  description TEXT        NOT NULL DEFAULT '',
  doctor_id   TEXT        NOT NULL,
  status      TEXT        NOT NULL DEFAULT 'new',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```
 
---
 
## Project Structure
 
```
appointment-service/
├── main.go
├── .env.example
├── Dockerfile
├── go.mod
├── internal/
│   ├── app/           # Wire-up: DB, NATS, Doctor client, repo, use-case, gRPC server
│   ├── client/        # Doctor Service gRPC client interface + implementation
│   ├── config/        # Config struct + DB connection pool
│   ├── event/         # EventPublisher interface + NATS implementation
│   ├── model/         # Appointment, Status, event models
│   ├── repository/    # PostgreSQL AppointmentRepository
│   ├── transport/
│   │   └── grpc/      # gRPC handler + error mapping
│   └── use-case/      # Business logic, status transitions
├── migrations/
│   ├── 000001_create_appointments.up.sql
│   └── 000001_create_appointments.down.sql
└── proto/
    ├── appointment.proto
    ├── appointment.pb.go
    ├── appointment_grpc.pb.go
    ├── doctor.proto
    ├── doctor.pb.go
    └── doctor_grpc.pb.go
```
 
---
 
## Graceful Shutdown
 
The service handles `SIGINT` and `SIGTERM`. On shutdown it:
1. Drains the gRPC server (waits for in-flight RPCs).
2. Closes the NATS connection.
3. Closes the Doctor Service gRPC client connection.
4. Closes the database connection pool.
5. Exits with code 0.