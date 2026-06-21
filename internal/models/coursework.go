package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LecturerAssignment struct {
	ID                    uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CourseID              uuid.UUID      `json:"course_id" gorm:"type:uuid;not null;index"`
	Course                Course         `json:"course" gorm:"foreignKey:CourseID"`
	CreatedByID           uuid.UUID      `json:"created_by_id" gorm:"type:uuid;not null;index"`
	Title                 string         `json:"title" gorm:"size:240;not null"`
	Instructions          string         `json:"instructions"`
	AssignmentType        string         `json:"assignment_type" gorm:"size:40;default:assignment"`
	TotalMarks            float64        `json:"total_marks" gorm:"default:0"`
	DueAt                 *time.Time     `json:"due_at"`
	AllowLateSubmission   bool           `json:"allow_late_submission" gorm:"default:false"`
	AllowFileUpload       bool           `json:"allow_file_upload" gorm:"default:true"`
	AllowTextSubmission   bool           `json:"allow_text_submission" gorm:"default:true"`
	FeedbackEnabled       bool           `json:"feedback_enabled" gorm:"default:true"`
	FeedbackReleasePolicy string         `json:"feedback_release_policy" gorm:"size:40;default:after_marking"`
	Rubric                datatypes.JSON `json:"rubric"`
	Status                string         `json:"status" gorm:"size:30;default:draft;index"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

func (a *LecturerAssignment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.AssignmentType == "" {
		a.AssignmentType = "assignment"
	}
	if a.FeedbackReleasePolicy == "" {
		a.FeedbackReleasePolicy = "after_marking"
	}
	if a.Status == "" {
		a.Status = "draft"
	}
	return nil
}

type StudentAssignmentSubmission struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	AssignmentID       uuid.UUID      `json:"assignment_id" gorm:"type:uuid;not null;index"`
	Assignment         LecturerAssignment `json:"assignment" gorm:"foreignKey:AssignmentID"`
	StudentID          uuid.UUID      `json:"student_id" gorm:"type:uuid;not null;index"`
	TextSubmission     string         `json:"text_submission"`
	FileURL            string         `json:"file_url"`
	SubmittedAt        *time.Time     `json:"submitted_at"`
	Score              *float64       `json:"score"`
	Feedback           string         `json:"feedback"`
	FeedbackReleasedAt *time.Time     `json:"feedback_released_at"`
	MarkedByID         *uuid.UUID     `json:"marked_by_id" gorm:"type:uuid"`
	MarkedAt           *time.Time     `json:"marked_at"`
	Status             string         `json:"status" gorm:"size:30;default:submitted;index"`
	Metadata           datatypes.JSON `json:"metadata"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

func (s *StudentAssignmentSubmission) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = "submitted"
	}
	if s.SubmittedAt == nil {
		now := time.Now()
		s.SubmittedAt = &now
	}
	return nil
}

type CASubmission struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	CourseID        uuid.UUID  `json:"course_id" gorm:"type:uuid;not null;index"`
	Course          Course     `json:"course" gorm:"foreignKey:CourseID"`
	LecturerID      uuid.UUID  `json:"lecturer_id" gorm:"type:uuid;not null;index"`
	ExamOfficerID   *uuid.UUID `json:"exam_officer_id" gorm:"type:uuid;index"`
	AcademicSession string     `json:"academic_session" gorm:"size:20;index"`
	Semester        string     `json:"semester" gorm:"size:30"`
	TotalStudents   int        `json:"total_students" gorm:"default:0"`
	TotalMarks      float64    `json:"total_marks" gorm:"default:40"`
	FileURL         string     `json:"file_url"`
	Summary         string     `json:"summary"`
	Status          string     `json:"status" gorm:"size:40;default:submitted_to_exam_officer;index"`
	SubmittedAt     *time.Time `json:"submitted_at"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (c *CASubmission) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.Status == "" {
		c.Status = "submitted_to_exam_officer"
	}
	if c.SubmittedAt == nil {
		now := time.Now()
		c.SubmittedAt = &now
	}
	return nil
}

type MarkedExamScriptSubmission struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	AssessmentID    uuid.UUID  `json:"assessment_id" gorm:"type:uuid;not null;index"`
	Assessment      Assessment `json:"assessment" gorm:"foreignKey:AssessmentID"`
	CourseID        uuid.UUID  `json:"course_id" gorm:"type:uuid;not null;index"`
	LecturerID      uuid.UUID  `json:"lecturer_id" gorm:"type:uuid;not null;index"`
	ExamOfficerID   *uuid.UUID `json:"exam_officer_id" gorm:"type:uuid;index"`
	AcademicSession string     `json:"academic_session" gorm:"size:20;index"`
	Semester        string     `json:"semester" gorm:"size:30"`
	MarkedCount     int        `json:"marked_count" gorm:"default:0"`
	FileURL         string     `json:"file_url"`
	Summary         string     `json:"summary"`
	Status          string     `json:"status" gorm:"size:40;default:submitted_to_exam_officer;index"`
	SubmittedAt     *time.Time `json:"submitted_at"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (s *MarkedExamScriptSubmission) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = "submitted_to_exam_officer"
	}
	if s.SubmittedAt == nil {
		now := time.Now()
		s.SubmittedAt = &now
	}
	return nil
}
