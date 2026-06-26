package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ExamActivityRecord struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	StudentID       string         `json:"student_id" gorm:"size:80;index"`
	StudentName     string         `json:"student_name" gorm:"size:180"`
	MatricNumber    string         `json:"matric_number" gorm:"size:80;index"`
	CourseCode      string         `json:"course_code" gorm:"size:40;index"`
	CourseTitle     string         `json:"course_title" gorm:"size:220"`
	AssessmentTitle string         `json:"assessment_title" gorm:"size:240"`
	SessionID       string         `json:"session_id" gorm:"size:120;index"`
	AttemptID       string         `json:"attempt_id" gorm:"size:120;index"`
	EventType       string         `json:"event_type" gorm:"size:100;index"`
	AlertLevel      string         `json:"alert_level" gorm:"size:60;index"`
	RiskLevel       string         `json:"risk_level" gorm:"size:60;index"`
	RiskPoints      int            `json:"risk_points" gorm:"default:0"`
	AttentionLevel  string         `json:"attention_level" gorm:"size:80;index"`
	Message         string         `json:"message"`
	Metadata        datatypes.JSON `json:"metadata"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (r *ExamActivityRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.AlertLevel == "" {
		r.AlertLevel = "medium"
	}
	if r.AttentionLevel == "" {
		r.AttentionLevel = attentionLevelForPoints(r.RiskPoints, r.AlertLevel)
	}
	return nil
}

func attentionLevelForPoints(points int, alertLevel string) string {
	if points >= 81 || alertLevel == "critical" || alertLevel == "urgent" {
		return "urgent_review_required"
	}
	if points >= 51 || alertLevel == "high" {
		return "high_attention_required"
	}
	return "medium_attention_required"
}
