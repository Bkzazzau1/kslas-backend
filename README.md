# K-SLAS Backend

Go backend for K-SLAS.

## Lecturer Assessment Engine

This update adds the lecturer assessment backend in Go.

Lecturers can create and manage:

- Exams
- Graded assessments
- Ungraded assessments
- Practice question sets

Supported question formats:

- Objective / single answer: `single_choice`
- Multiple answers: `multiple_choice`
- Essay: `essay`
- Fill in the blank: `fill_blank`
- Drag and drop: `drag_drop`
- Image-based question: `image_question`

Supported answer tools:

- Typed answer
- Image upload URL
- File upload URL
- Whiteboard/sketch image URL
- Whiteboard JSON stroke data

## Lecturer course assignment

Lecturer assessment creation is now linked to course assignment. If `created_by_id` is sent when creating an assessment, the lecturer must have an active assignment for the selected course.

Assignment endpoints:

- `GET /api/admin/lecturer-course-assignments`
- `POST /api/admin/lecturer-course-assignments`
- `GET /api/lecturer/course-assignments?lecturer_id={id}`
- `GET /api/lecturer/courses?lecturer_id={id}`

Example assignment payload:

```json
{
  "lecturer_id": "00000000-0000-0000-0000-000000000000",
  "course_id": "00000000-0000-0000-0000-000000000000",
  "academic_session": "2025/2026",
  "semester": "first",
  "level": "200",
  "teaching_hours_per_week": 4,
  "assigned_by_id": "00000000-0000-0000-0000-000000000000"
}
```

## Exam moderation workflow

Assessments now follow this moderation path:

```txt
draft
submitted_to_moderator
approved_by_moderator
submitted_to_exam_officer
approved_for_exam
published
closed
```

Returned questions move to:

```txt
returned_for_correction
```

Moderation endpoints:

- `POST /api/lecturer/assessments/{id}/submit-for-moderation`
- `GET /api/moderator/assessments`
- `POST /api/moderator/assessments/{id}/approve`
- `POST /api/moderator/assessments/{id}/return`
- `POST /api/lecturer/assessments/{id}/submit-to-exam-officer`
- `GET /api/exam-officer/assessments`
- `POST /api/exam-officer/assessments/{id}/approve`
- `POST /api/exam-officer/assessments/{id}/return`
- `POST /api/lecturer/assessments/{id}/publish`
- `POST /api/lecturer/assessments/{id}/close`

Each moderation action is saved in `assessment_moderation_actions` for audit trail.

## Main endpoints

- `GET /api/lecturer/assessments`
- `POST /api/lecturer/assessments`
- `POST /api/lecturer/assessments/{id}/submit-for-moderation`
- `POST /api/lecturer/assessments/{id}/submit-to-exam-officer`
- `POST /api/lecturer/assessments/{id}/publish`
- `POST /api/lecturer/assessments/{id}/close`
- `GET /api/lecturer/questions?assessment_id={id}`
- `POST /api/lecturer/questions`
- `POST /api/lecturer/options`
- `POST /api/lecturer/assets`
- `GET /api/lecturer/submissions?assessment_id={id}`
- `PATCH /api/lecturer/answers/{id}/mark`
- `GET /api/student/assessments`
- `POST /api/student/assessments/{id}/start?student_id={id}`
- `POST /api/student/assessments/{id}/submit?student_id={id}`
- `POST /api/student/answers`

## Run

Set your Postgres connection string first:

```bash
export DATABASE_DSN="host=localhost user=postgres password=postgres dbname=kslas port=5432 sslmode=disable"
export ALLOWED_ORIGINS="https://YOUR-ADMIN-UI-DOMAIN"
go mod tidy
go run ./cmd/api
```

On Windows PowerShell:

```powershell
$env:DATABASE_DSN="host=localhost user=postgres password=postgres dbname=kslas port=5432 sslmode=disable"
$env:ALLOWED_ORIGINS="https://YOUR-ADMIN-UI-DOMAIN"
go mod tidy
go run ./cmd/api
```

The server starts on port `8080` unless `PORT` is set.

## VPS note

When the admin UI and backend are on different domains or ports, set `ALLOWED_ORIGINS` to the admin UI origin. For multiple origins, separate them with commas:

```bash
export ALLOWED_ORIGINS="https://admin.example.com,https://www.admin.example.com"
```

If Nginx serves the admin UI and proxies `/api` to the Go backend from the same origin, the Flutter app can use the same origin automatically.
