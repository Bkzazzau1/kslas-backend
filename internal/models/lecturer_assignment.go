package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LecturerCourseAssignment struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	LecturerID           uuid.UUID  `json:"lecturer_id" gorm:"type:uuid;not null;index"`
	CourseID             uuid.UUID  `json:"course_id" gorm:"type:uuid;not null;index"`
	Course               Course     `json:"course" gorm:"foreignKey:CourseID"`
	DepartmentID         *uuid.UUID `json:"department_id" gorm:"type:uuid;index"`
	AcademicSession      string     `json:"academic_session" gorm:"size:20;default:2025/2026;index"`
	Semester             string     `json:"semester" gorm:"size:30;default:first"`
	Level                string     `json:"level" gorm:"size:20"`
	Role                 string     `json:"role" gorm:"size:30;default:lecturer"`
	TeachingHoursPerWeek float64    `json:"teaching_hours_per_week" gorm:"default:0"`
	AssignedByID         *uuid.UUID `json:"assigned_by_id" gorm:"type:uuid"`
	StartsAt             *time.Time `json:"starts_at"`
	EndsAt               *time.Time `json:"ends_at"`
	Status               string     `json:"status" gorm:"size:30;default:active;index"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (a *LecturerCourseAssignment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Status == "" {
		a.Status = "active"
	}
	if a.Role == "" {
		a.Role = "lecturer"
	}
	if a.AcademicSession == "" {
		a.AcademicSession = "2025/2026"
	}
	if a.Semester == "" {
		a.Semester = "first"
	}
	return nil
}

type AssessmentModerationAction struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	AssessmentID uuid.UUID  `json:"assessment_id" gorm:"type:uuid;not null;index"`
	Assessment   Assessment `json:"-" gorm:"foreignKey:AssessmentID"`
	ActorID      *uuid.UUID `json:"actor_id" gorm:"type:uuid;index"`
	Action       string     `json:"action" gorm:"size:60;not null;index"`
	FromStatus   string     `json:"from_status" gorm:"size:60"`
	ToStatus     string     `json:"to_status" gorm:"size:60"`
	Comment      string     `json:"comment"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (a *AssessmentModerationAction) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
