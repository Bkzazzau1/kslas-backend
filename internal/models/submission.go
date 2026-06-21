package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
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

type StudentAnswer struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	SubmissionID       uuid.UUID      `json:"submission_id" gorm:"type:uuid;not null;index:idx_answer_unique,unique"`
	QuestionID         uuid.UUID      `json:"question_id" gorm:"type:uuid;not null;index:idx_answer_unique,unique"`
	SelectedOptionID   *uuid.UUID     `json:"selected_option_id" gorm:"type:uuid"`
	SelectedOptionIDs  datatypes.JSON `json:"selected_option_ids"`
	TextAnswer         string         `json:"text_answer"`
	BlankAnswers       datatypes.JSON `json:"blank_answers"`
	DragDropAnswer     datatypes.JSON `json:"drag_drop_answer"`
	ImageAnswerURL     string         `json:"image_answer_url"`
	AnswerFileURL      string         `json:"answer_file_url"`
	WhiteboardImageURL string         `json:"whiteboard_image_url"`
	WhiteboardData     datatypes.JSON `json:"whiteboard_data"`
	IsAutoMarked       bool           `json:"is_auto_marked" gorm:"default:false"`
	AutoScore          float64        `json:"auto_score" gorm:"default:0"`
	ManualScore        *float64       `json:"manual_score"`
	FinalScore         float64        `json:"final_score" gorm:"default:0"`
	MarkingStatus      string         `json:"marking_status" gorm:"size:30;default:pending"`
	LecturerFeedback   string         `json:"lecturer_feedback"`
	MarkedByID         *uuid.UUID     `json:"marked_by_id" gorm:"type:uuid"`
	MarkedAt           *time.Time     `json:"marked_at"`
	SubmittedAt        time.Time      `json:"submitted_at"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

func (a *StudentAnswer) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.MarkingStatus == "" {
		a.MarkingStatus = "pending"
	}
	if a.SubmittedAt.IsZero() {
		a.SubmittedAt = time.Now()
	}
	return nil
}
