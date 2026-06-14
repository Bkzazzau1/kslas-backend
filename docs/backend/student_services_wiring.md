# Student Services Backend Wiring

The student-services model, repository, service and handler layers are already present.

This document records the remaining local wiring steps to expose the new `/api/v1/student/*` endpoints.

## 1. Migration wiring

Find the file that defines `database.Bootstrap(ctx, db, cfg)`.

Inside the AutoMigrate section, add the student services models:

```go
for _, model := range StudentServicesMigrationModels() {
    if err := db.AutoMigrate(model); err != nil {
        return err
    }
}
```

The model list is defined in:

```text
internal/database/student_services_models.go
```

It includes:

- AuditLog
- SupportTicket
- SupportTicketMessage
- SupportTicketFile
- InternshipProfile
- InternshipLetterRequest
- InternshipLogbookEntry
- TranscriptRequest
- GraduationMap
- GraduationMapItem

## 2. Main dependency wiring

In `cmd/kslasbackend/main.go`, add:

```go
studentServicesRepository := repository.NewStudentServicesRepository(db)
studentServicesService := services.NewStudentServicesService(studentServicesRepository)
studentServicesHandler := handlers.NewStudentServicesHandler(studentServicesService)
```

Then pass `studentServicesHandler` into the router dependencies after adding the dependency field.

## 3. Router dependency field

In `internal/server/server.go`, add this field to `Dependencies`:

```go
StudentServicesHandler *handlers.StudentServicesHandler
```

## 4. Student v1 routes

Register these routes after `/api/auth/me`:

```go
mux.Handle(
    "/api/v1/student/graduation-map",
    chain(
        method(http.MethodGet, http.HandlerFunc(dep.StudentServicesHandler.GraduationMap)),
        middleware.AuthMiddleware(dep.JWTService),
    ),
)

mux.Handle(
    "/api/v1/student/support/tickets",
    chain(
        http.HandlerFunc(dep.StudentServicesHandler.SupportTickets),
        middleware.AuthMiddleware(dep.JWTService),
    ),
)

mux.Handle(
    "/api/v1/student/support/tickets/{ticketID}/replies",
    chain(
        http.HandlerFunc(dep.StudentServicesHandler.SupportTicketReplies),
        middleware.AuthMiddleware(dep.JWTService),
    ),
)

mux.Handle(
    "/api/v1/student/internship/profile",
    chain(
        http.HandlerFunc(dep.StudentServicesHandler.InternshipProfile),
        middleware.AuthMiddleware(dep.JWTService),
    ),
)

mux.Handle(
    "/api/v1/student/transcripts/official-requests",
    chain(
        http.HandlerFunc(dep.StudentServicesHandler.TranscriptRequests),
        middleware.AuthMiddleware(dep.JWTService),
    ),
)
```

## 5. Local verification

Run:

```powershell
cd C:\PROJECTS\kslas-backend
go test ./...
go run ./cmd/kslasbackend
```

## 6. Endpoint behavior

These routes must always use the authenticated user ID from JWT context.

The frontend must not send another student's ID for student-owned data.

## 7. Next backend work after wiring

After the routes compile and run:

1. Add seed/sample rows for a demo student graduation map.
2. Add support ticket sequence generation using a database-backed counter.
3. Add official transcript request events.
4. Add internship letter request endpoint.
5. Add file upload integration for support evidence and internship acceptance letters.
