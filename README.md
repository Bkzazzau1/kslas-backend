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

## Main endpoints

- `GET /api/lecturer/assessments`
- `POST /api/lecturer/assessments`
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
