package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentSubmission struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	AssessmentID uuid.UUID `json:"assessment_id" gorm:"type:uuid;not null;index:idx_submission_unique,unique"`
	StudentID    uuid.UUID `json:"student_id" gorm:"type:uuid;not null;index:idx_submission_unique,unique"`
	Status       string    `json:"status" gorm:"size:30;default:in_progress"`
	StartedAt    time.Time `json:"started_at"`
	SubmittedAt  *time.Time `json:"submitted_at"`
	TotalScore   float64   `json:"total_score" gorm:"default:0"`
	ReleasedAt   *time.Time `json:"released_at"`
	Answers      []StudentAnswer `json:"answers,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *StudentSubmission) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = "in_progress"
	}
	if s.StartedAt.IsZero() {
		s.StartedAt = time.Now()
	}
	return nil
}
