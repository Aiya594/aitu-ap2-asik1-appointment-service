  # Appointment Service - gRPC Medical Scheduling Platform

## 1. Project Overview

The Appointment Service is part of a microservices-based Medical Scheduling Platform that has been migrated from REST to **gRPC** for inter-service and client-to-service communication. This service manages the complete lifecycle of appointments in the system while maintaining Clean Architecture principles and bounded-context boundaries.

The Appointment Service:
- Owns appointment data and retrieval logic
- Validates doctor existence by calling the Doctor Service via gRPC
- Manages appointment status transitions (new → in_progress → done)
- Returns descriptive gRPC status codes on errors
- Remains decoupled from the Doctor Service through gRPC client injection

**Key Assignment Goal:** Replace all REST communication with gRPC while preserving domain logic, use-case implementations, and Clean Architecture layering.


---

## 2. Scope and Constraints

### What Changes

- **All HTTP/REST endpoints are replaced with gRPC endpoints**
  - Gin HTTP server replaced with gRPC server
  - `transport/http/` replaced with `transport/grpc/`

- **Inter-service communication now uses gRPC**
  - Appointment Service calls Doctor Service over gRPC (not REST)
  - Both services expose a gRPC server listening on designated ports

- **Protocol Buffers define all service contracts**
  - .proto files committed to the repository
  - Generated Go stubs (`*pb.go`, `*_grpc.pb.go`) committed alongside

---

## 3. Architecture Overview

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

### Project Structure

```
appointment-service/
├── main.go                         # Entry point
├── go.mod / go.sum
├── proto/                          # Protocol Buffer definitions
│   ├── appointment.proto           # Appointment service contract
│   ├── appointment.pb.go           # Generated stubs (committed)
│   ├── appointment_grpc.pb.go      # Generated gRPC stubs (committed)
│   ├── doctor.proto                # Doctor service contract (imported)
│   ├── doctor.pb.go                # Doctor stubs (imported from doctor-service)
│   └── doctor_grpc.pb.go           # Doctor gRPC stubs (imported from doctor-service)
├── internal/
│   ├── model/
│   │   ├── appointment.go          # Domain model (unchanged from Assignment 1)
│   │   └── status.go               # Status enum
│   ├── repository/
│   │   ├── repository.go           # Storage interface (unchanged)
│   │   └── errors.go               # Repository errors
│   ├── use-case/
│   │   ├── usecase.go              # Business logic (unchanged from A1)
│   │   └── error.go                # Use-case errors
│   ├── client/
│   │   ├── doctor.go               # DoctorClient interface (NEW - injected into use case)
│   │   └── doctorgrpc.go           # gRPC client implementation (NEW)
│   ├── transport/
│   │   ├── grpc/
│   │   │   └── server.go           # gRPC server handlers (REPLACES http/)
│   │   └── http/                   # REMOVED (replaced by grpc/)
│   ├── config/
│   │   └── config.go               # Configuration (ports, etc.)
│   └── app/
│       └── app.go                  # Application setup (gRPC server instead of Gin)
└── README.md                       # This file
```

### Dependency Flow (Clean Architecture)

```
gRPC Handler
    ↓
UseCase (Business Logic)
    ↓ (injected via interface)
DoctorClient (interface)
    ↓
gRPC Client Implementation (Doctor Service)
```

**Rules:**
- gRPC handler unmarshal proto messages, calls use case, returns proto responses
- Use case contains all business logic; imports NO protobuf types
- Repository handles storage only
- DoctorClient is an interface injected into the use case (dependency injection)
- Mapping between proto messages and domain models happens ONLY in the gRPC layer

---


## 4. Installing and Regenerating Proto Stubs

### Prerequisites

1. **Install protoc** (Protocol Buffer Compiler)

   - Windows: Download from [protobuf releases](https://github.com/protocolbuffers/protobuf/releases) and add to PATH

2. **Install Go gRPC plugins**
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

### Regenerate Stubs

From the `appointment-service/` directory:

```bash
# Generate stubs for appointment.proto 
protoc --go_out=. --go-grpc_out=. proto/doctor.proto
protoc --go_out=. --go-grpc_out=. proto/appointment.proto
```

---

## 5. Running Both Services Locally

### Prerequisites

- Both services in their respective directories:
  - `appointment-service/` (this directory)
  - `../aitu-ap2-asik1-doctor-service/` (Doctor Service)

### Step 1: Start the Doctor Service

```bash
cd ../aitu-ap2-asik1-doctor-service
go run main.go
```

Expected output:
```
Doctor Service listening on port 50051 (gRPC)
```

**Doctor Service gRPC Port:** `localhost:50051`

### Step 2: Start the Appointment Service

In a new terminal:

```bash
cd appointment-service
go run main.go
```

Expected output:
```
Appointment Service listening on port 50052 (gRPC)
```

**Appointment Service gRPC Port:** `localhost:50052`

---

---

## 6. gRPC Status Codes and Error Handling

All errors must use standard gRPC status codes from `google.golang.org/grpc/codes` package.

| Situation | gRPC Status Code | Example Message |
|-----------|------------------|-----------------|
| Required field missing (title, doctor_id) | `codes.InvalidArgument` | `"title is required"` |
| Doctor ID not found (local check) | `codes.NotFound` | `"appointment not found: apt-123"` |
| Doctor does not exist (remote Doctor Service) | `codes.FailedPrecondition` | `"doctor not found: doc-456"` |
| Doctor Service unreachable | `codes.Unavailable` | `"doctor service unavailable: connection refused"` |
| Invalid status transition (done → new) | `codes.InvalidArgument` | `"invalid status transition from done to new"` |


---

## 7. Inter-Service Communication (gRPC Flow)

### Appointment Creation Flow (with Doctor Validation)

```
Client
  |
Appointment gRPC Handler
  | (unmarshal proto)
UseCase.CreateAppointment()
  | (call injected DoctorClient interface)
DoctorClient.GetDoctor() [gRPC call]
  |
Doctor Service gRPC Handler
  |
Doctor Repository
  | (return DoctorResponse or error)
Doctor gRPC Handler → DoctorClient
  | (return error if doctor not found)
UseCase → validate doctor exists
  | (if valid, create appointment)
Repository.Create(appointment)
  | (return AppointmentResponse)/error)
Appointment Handler → gRPC Response
  |
Client
```


---

## 8. Failure Scenario

### Doctor Service is Unreachable

**When:** During `CreateAppointment`, the gRPC call to `DoctorService.GetDoctor()` fails.

**What Happens:**
1. gRPC client attempts connection to Doctor Service (localhost:50051)
2. Connection times out or is refused
3. gRPC client returns error with status code `codes.Unavailable`
4. UseCase receives the error
5. UseCase **does not** create the appointment
6. gRPC handler returns status `codes.Unavailable` to the client

**Error Message Example:**
```
rpc error: code = Unavailable
desc = doctor service unavailable: context deadline exceeded
```

**Client sees:**
```
Error: Appointment could not be created. Doctor Service is currently unavailable.
```

### Doctor Not Found

**When:** Doctor Service is reachable but the requested doctor_id doesn't exist.

**What Happens:**
1. gRPC call succeeds but Doctor Service returns NOT_FOUND
2. gRPC client receives error with status `codes.NotFound`
3. UseCase receives error
4. UseCase does NOT create appointment
5. gRPC handler converts to `codes.FailedPrecondition` and returns to client

**Error Message Example:**
```
rpc error: code = FailedPrecondition
desc = doctor not found: doc-999
```

## 9. REST vs. gRPC Trade-Offs

### Three Key Differences

| Aspect | REST | gRPC |
|--------|------|------|
| **Serialization** | JSON (text, larger payload) | Protocol Buffers (binary, smaller) |
| **Performance** | HTTP/1.1 (request-response) | HTTP/2 (multiplexed streams) |
| **Type Safety** | Runtime validation or OpenAPI | Compile-time validation via .proto |
| **Client Generation** | Manual or OpenAPI tooling | `protoc` generates code from .proto |
| **Latency** | Higher (text parsing) | Lower (binary, HTTP/2) |
| **Human Readability** | Easy to debug with curl/Postman | Requires grpcurl or Postman gRPC extension |


