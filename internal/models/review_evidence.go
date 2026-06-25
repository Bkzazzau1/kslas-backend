package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ReviewEvidenceCase struct {
	ID              uuid.UUID                    `json:"id" gorm:"type:uuid;primaryKey"`
	StudentID       string                       `json:"student_id" gorm:"size:80;index"`
	StudentName     string                       `json:"student_name" gorm:"size:180;not null"`
	MatricNumber    string                       `json:"matric_number" gorm:"size:80;index"`
	CourseCode      string                       `json:"course_code" gorm:"size:40;index"`
	CourseTitle     string                       `json:"course_title" gorm:"size:220"`
	AssessmentTitle string                       `json:"assessment_title" gorm:"size:240"`
	SessionID       string                       `json:"session_id" gorm:"size:120;index"`
	AttemptID       string                       `json:"attempt_id" gorm:"size:120;index"`
	AttentionLevel  string                       `json:"attention_level" gorm:"size:80;default:medium_attention_required;index"`
	Status          string                       `json:"status" gorm:"size:80;default:awaiting_review;index"`
	RiskLevel       string                       `json:"risk_level" gorm:"size:40;index"`
	RiskPoints      int                          `json:"risk_points" gorm:"default:0"`
	ReviewSummary   string                       `json:"review_summary"`
	AssignedRole    string                       `json:"assigned_role" gorm:"size:80;index"`
	Metadata        datatypes.JSON               `json:"metadata"`
	EvidenceFiles   []ReviewEvidenceFile         `json:"evidence_files,omitempty" gorm:"foreignKey:CaseID"`
	Timeline        []ReviewEvidenceTimelineItem `json:"timeline,omitempty" gorm:"foreignKey:CaseID"`
	Actions         []ReviewEvidenceAction       `json:"actions,omitempty" gorm:"foreignKey:CaseID"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
}

func (c *ReviewEvidenceCase) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.AttentionLevel == "" {
		c.AttentionLevel = "medium_attention_required"
	}
	if c.Status == "" {
		c.Status = "awaiting_review"
	}
	if c.AssignedRole == "" {
		c.AssignedRole = "exam_officer"
	}
	return nil
}

type ReviewEvidenceFile struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CaseID       uuid.UUID      `json:"case_id" gorm:"type:uuid;not null;index"`
	Title        string         `json:"title" gorm:"size:180;not null"`
	EvidenceType string         `json:"evidence_type" gorm:"size:80;index"`
	SourceKey    string         `json:"source_key" gorm:"size:80;index"`
	FileURL      string         `json:"file_url"`
	Status       string         `json:"status" gorm:"size:80;default:available"`
	Detail       string         `json:"detail"`
	Metadata     datatypes.JSON `json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func (f *ReviewEvidenceFile) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	if f.Status == "" {
		f.Status = "available"
	}
	return nil
}

type ReviewEvidenceTimelineItem struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CaseID     uuid.UUID      `json:"case_id" gorm:"type:uuid;not null;index"`
	EventType  string         `json:"event_type" gorm:"size:100;index"`
	Message    string         `json:"message"`
	AlertLevel string         `json:"alert_level" gorm:"size:60;index"`
	EventTime  time.Time      `json:"event_time" gorm:"index"`
	Metadata   datatypes.JSON `json:"metadata"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func (t *ReviewEvidenceTimelineItem) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.EventTime.IsZero() {
		t.EventTime = time.Now()
	}
	return nil
}

type ReviewEvidenceAction struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	CaseID     uuid.UUID  `json:"case_id" gorm:"type:uuid;not null;index"`
	ActorID    *uuid.UUID `json:"actor_id" gorm:"type:uuid;index"`
	ActorRole  string     `json:"actor_role" gorm:"size:80;index"`
	Action     string     `json:"action" gorm:"size:80;index"`
	Comment    string     `json:"comment"`
	FromStatus string     `json:"from_status" gorm:"size:80"`
	ToStatus   string     `json:"to_status" gorm:"size:80"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (a *ReviewEvidenceAction) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
