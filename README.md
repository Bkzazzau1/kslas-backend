# K-SLAS Backend

This branch adds the first production-ready foundation for the Lecturer Assessment Engine.

## What this module supports

Lecturers can create and manage:

- Exams
- Graded assessments
- Ungraded assessments
- Practice question sets

Supported question formats:

- Objective / single answer
- Multiple answers
- Essay
- Fill in the blank
- Drag and drop
- Image-based question

Supported answer tools:

- Typed answer
- Student image upload
- File upload
- Whiteboard / sketch / diagram snapshot
- Whiteboard JSON stroke data for later review

## Key API areas

- `GET/POST /api/lecturer/assessments/`
- `POST /api/lecturer/assessments/{id}/publish/`
- `POST /api/lecturer/assessments/{id}/close/`
- `GET/POST /api/lecturer/questions/`
- `GET/POST /api/lecturer/options/`
- `GET/POST /api/lecturer/assets/`
- `GET /api/lecturer/submissions/`
- `PATCH /api/lecturer/answers/{id}/mark/`
- `GET /api/student/assessments/`
- `POST /api/student/assessments/{id}/start/`
- `POST /api/student/assessments/{id}/submit/`

## Run locally

```bash
python -m venv .venv
. .venv/Scripts/activate  # Windows PowerShell users can activate the equivalent script
pip install -r requirements.txt
python manage.py makemigrations
python manage.py migrate
python manage.py createsuperuser
python manage.py runserver
```

## Notes

The lecturer module uses Django groups or staff access through the standard Django user model for now. A dedicated role/permission service can be added next when the wider HoD, exam officer, invigilator, and academic records modules are connected.
