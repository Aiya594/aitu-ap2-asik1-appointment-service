# appointment-service


# Project Overview

The Appointment Service manages the lifecycle of appointments in a microservices architecture. It is responsible for creating, updating, and retrieving appointments while ensuring that referenced doctors exist.

## Purpose

This service follows
Clean Architecture principles,service-to-service communication via REST and
Separation of concerns and independent data ownership.
The system is structured this way to ensure scalability and maintainability

## Service Responsibilities

The Appointment Service manages:

- Appointment creation
- Appointment retrieval 
- Status management (new → in_progress → done)

Validation rules:
- required fields
- valid status transitions
- doctor existence validation via [Doctor Service](https://github.com/Aiya594/aitu-ap2-asik1-doctor-service)

## Folder Structure

```
internal/
├── transport/http/      -> HTTP handlers (Gin)
├── use-case/       -> Business logic
├── repository/     -> In-memory storage
├── model/          -> Domain models + rules
├── client/         -> Doctor Service HTTP client
├── config/         -> Configuration
├── app/            -> Server management
```

## Dependency Flow
```
Handler -> UseCase -> Repository
              |
           Client (Doctor Service)
```

Rules:
- Handlers only handle HTTP
- UseCase contains all business logic
- Repository handles storage only
- Client handles external HTTP calls

## Inter-Service Communication
When communication happens during appointment creation:
```
CreateAppointment -> call Doctor Service -> validate doctor exists -> if valid -> create appointment
            if not -> return error
```

# How to Run
1. Set environment variables as given in ```.env.example```
2. Run [Doctor Service](https://github.com/Aiya594/aitu-ap2-asik1-doctor-service) first
```
go run main.go
```
3. Run Appointment Service
```
go run main.go
```

Test endpoints
- ```POST /appointments``` - create a new appointment
```
{
    "title":"appointment",
    "description":"test",
    "doctor_id":"valide_doctor_id
}
```
- ```GET /appointments``` - list all appointments
- ```GET /appointments/{id}``` - get by ID
- ```PATCH /appointments/{id}/status``` - update status of appointment

## Why No Shared Database?

Each service owns its data to ensure loose coupling, independent deployment, and
schema isolation

Using a shared database would tightly couple services, break service boundaries and introduce dependencies
Instead, `*communication happens via REST APIs*.` 

## Failure Scenario
Doctor Service unavailable:
```
Doctor Service DOWN ->
    validation fails -> appointment creation fails
```
Service behavior:
- request fails immediately
- error returned to client
- no appointment created

