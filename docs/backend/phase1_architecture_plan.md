# K-SLAS Backend Phase 1 Architecture Plan

This repository is the backend for the K-SLAS ecosystem.

Frontend repositories:

- `Bkzazzau1/k-slas-my-course`: student-facing app only.
- `Bkzazzau1/kslas-admin-ui`: staff/admin operations only.

Finance and billing are excluded until the school contract approves that scope.

## Current Backend Position

The backend already has a Go module and a modular service structure. The existing code includes:

- Authentication and JWT service wiring.
- PostgreSQL/GORM dependency setup.
- Academic repositories and handlers.
- Administration, teaching, material, assignment, forum, message, content, exam, result and report services.
- A `net/http` router.
- RBAC and permission middleware.

The next work should not restart from zero. We should strengthen the existing structure and align the routes and database model with the two frontend apps.

## Target API Boundary

The backend should move toward two clean API namespaces:

```text
/api/v1/student/*
/api/v1/admin/*
```

### Student namespace

Student routes must derive the student identity from the JWT token. The frontend must not be trusted to send another `student_id` for student-owned data.

Student routes should cover:

- Dashboard summary.
- Courses and materials.
- Course registration.
- Academic record.
- Graduation mapping.
- Unofficial transcript preview and official transcript request.
- Internship/SIWES profile, placement letter request, acceptance upload and logbook.
- Assignments and submissions.
- Exams and student attempts.
- Results visible to the student.
- Notices visible to the student.
- Live classes and replays.
- Student support tickets.

### Admin namespace

Admin routes must derive staff identity, role and scope from JWT and permission records.

Admin routes should cover:

- Academic structure: faculties, departments, programmes, courses and cohorts.
- Staff and role management.
- Notice publishing.
- Course registration approval.
- Lecturer assignment and marking.
- Question moderation.
- Exam setup and readiness.
- Invigilation operations.
- Exam session overview and sign-off.
- Results approval and records reconciliation.
- Student support/helpdesk operations.
- Reports and analytics.

## Phase 1 Scope

Phase 1 should focus on foundation and must be completed before adding more advanced workflows.

### 1. Auth and identity

Required backend pieces:

- User login.
- JWT access token.
- Refresh token/session table.
- `me` endpoint with roles and scopes.
- Student/staff user type handling.
- Password hashing and reset flow later.

Suggested tables:

```text
users
roles
user_roles
permissions
role_permissions
permission_scopes
sessions
```

### 2. RBAC and scoped permissions

The backend must support permissions such as:

```text
student.self.read
student.support.create
student.transcript.request
student.graduation.read
academic.structure.manage
course.registration.approve
assignment.grade
question.moderate
exam.manage
invigilation.operate
records.manage
support.manage
reports.view
```

Scope types:

```text
school
faculty
department
programme
course
cohort
exam
room
```

### 3. Audit logs

Every sensitive action should write an audit log.

Suggested table:

```text
audit_logs: id, actor_user_id, actor_role, action, entity_type, entity_id, before_json, after_json, ip_address, user_agent, created_at
```

Audit required for:

- Role and permission changes.
- Academic structure changes.
- Course registration approval/rejection.
- Result review/edit/release.
- Transcript request review/dispatch.
- Exam setup, open, close and incident handling.
- Support ticket assignment, escalation and resolution.

### 4. Academic structure

Minimum tables:

```text
faculties
departments
programmes
courses
programme_courses
academic_sessions
semesters
cohorts
cohort_students
```

The current `Course` model can be extended into a more complete curriculum system using `programme_courses` rather than relying only on `programme_id` inside the course table.

### 5. Student academic services

After foundation, add:

```text
course_registration_windows
course_registrations
course_registration_items
carryover_confirmations
academic_records
cgpa_snapshots
graduation_maps
graduation_map_items
transcript_requests
transcript_request_events
```

The student app must be able to show:

- Personal academic record.
- Course registration status.
- Graduation mapping.
- Remaining courses.
- Pending/carryover alerts.
- Unofficial transcript preview.
- Official transcript request status.

## Existing Route Alignment Plan

The current router already exposes routes such as `/api/faculties`, `/api/courses`, `/api/my/course-registrations`, `/api/exams`, `/api/results` and invigilator alert routes.

Do not delete these immediately. Use a staged route migration:

1. Keep existing routes working.
2. Add new `/api/v1/student/*` and `/api/v1/admin/*` routes beside them.
3. Point the Flutter apps to the new v1 routes.
4. Mark older routes as legacy.
5. Remove legacy routes only after both apps are stable.

## Recommended Package Structure

Continue with the current style, but make modules more explicit:

```text
cmd/kslasbackend/main.go
internal/config
internal/database
internal/database/models
internal/dto
internal/handlers
internal/middleware
internal/rbac
internal/repository
internal/server
internal/services
```

Suggested additions:

```text
internal/dto/student_services_dto.go
internal/handlers/student_services_handler.go
internal/repository/student_services_repository.go
internal/services/student_services_service.go
internal/dto/audit_dto.go
internal/services/audit_service.go
internal/repository/audit_repository.go
```

## First Backend Coding Order

1. Confirm project builds locally.
2. Add missing scope types: cohort, exam and room.
3. Add sessions/refresh-token table if not already present.
4. Add audit log model and repository.
5. Add student services models for support, internship, transcript and graduation mapping.
6. Add `/api/v1/student/me/summary`.
7. Add `/api/v1/student/graduation-map`.
8. Add `/api/v1/student/support/tickets`.
9. Add `/api/v1/student/transcripts/unofficial-preview` and official request creation.
10. Add `/api/v1/student/internship/profile`.

## Non-Negotiable Rules

- No Finance/Billing endpoints yet.
- Student app reads and writes only authenticated student-owned resources.
- Admin app performs staff workflows only through scoped permissions.
- All admin changes must be audit logged.
- Official transcript approval and dispatch are staff workflows, not student workflows.
- Graduation mapping must be derived from academic records, curriculum and registration/result history.
- Invigilation and exam session sign-off must remain admin-only.
