package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Assessment struct {
	ID                    uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CourseID              uuid.UUID      `json:"course_id" gorm:"type:uuid;not null"`
	Course                Course         `json:"course" gorm:"foreignKey:CourseID"`
	Title                 string         `json:"title" gorm:"size:240;not null"`
	Description           string         `json:"description"`
	AssessmentType        string         `json:"assessment_type" gorm:"size:40;default:graded_assessment"`
	DurationMinutes       int            `json:"duration_minutes" gorm:"default:30"`
	TotalMarks            float64        `json:"total_marks" gorm:"default:0"`
	StartTime             *time.Time     `json:"start_time"`
	EndTime               *time.Time     `json:"end_time"`
	CreatedByID           *uuid.UUID     `json:"created_by_id" gorm:"type:uuid"`
	Status                string         `json:"status" gorm:"size:30;default:draft"`
	ProctoringLevel       string         `json:"proctoring_level" gorm:"size:20;default:none"`
	AllowMobile           bool           `json:"allow_mobile" gorm:"default:true"`
	ShuffleQuestions      bool           `json:"shuffle_questions" gorm:"default:false"`
	ShuffleOptions        bool           `json:"shuffle_options" gorm:"default:false"`
	ShowResultImmediately bool           `json:"show_result_immediately" gorm:"default:false"`
	Rules                 datatypes.JSON `json:"rules"`
	Questions             []Question     `json:"questions,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

func (a *Assessment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Status == "" {
		a.Status = "draft"
	}
	if a.AssessmentType == "" {
		a.AssessmentType = "graded_assessment"
	}
	if a.ProctoringLevel == "" {
		a.ProctoringLevel = "none"
	}
	return nil
}

type Question struct {
	ID                    uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	AssessmentID          uuid.UUID      `json:"assessment_id" gorm:"type:uuid;not null;index"`
	Assessment            Assessment     `json:"-" gorm:"foreignKey:AssessmentID"`
	CourseID              uuid.UUID      `json:"course_id" gorm:"type:uuid;not null"`
	QuestionType          string         `json:"question_type" gorm:"size:40;not null"`
	QuestionText          string         `json:"question_text" gorm:"not null"`
	Instruction           string         `json:"instruction"`
	Marks                 float64        `json:"marks" gorm:"default:1"`
	OrderNumber           int            `json:"order_number" gorm:"default:1"`
	Difficulty            string         `json:"difficulty" gorm:"size:20;default:medium"`
	AllowWhiteboard       bool           `json:"allow_whiteboard" gorm:"default:false"`
	AllowImageUpload      bool           `json:"allow_image_upload" gorm:"default:false"`
	AllowFileUpload       bool           `json:"allow_file_upload" gorm:"default:false"`
	RequiresManualMarking bool           `json:"requires_manual_marking" gorm:"default:false"`
	AutoMarkingEnabled    bool           `json:"auto_marking_enabled" gorm:"default:true"`
	Metadata              datatypes.JSON `json:"metadata"`
	IsActive              bool           `json:"is_active" gorm:"default:true"`
	Options               []QuestionOption `json:"options,omitempty"`
	Assets                []QuestionAsset  `json:"assets,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

func (q *Question) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	q.ApplyMarkingPolicy()
	return nil
}

func (q *Question) BeforeSave(tx *gorm.DB) error {
	q.ApplyMarkingPolicy()
	return nil
}

func (q *Question) ApplyMarkingPolicy() {
	if q.QuestionType == "essay" || q.QuestionType == "image_question" || q.AllowWhiteboard || q.AllowImageUpload || q.AllowFileUpload {
		q.RequiresManualMarking = true
		q.AutoMarkingEnabled = false
	}
}

type QuestionOption struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	QuestionID  uuid.UUID `json:"question_id" gorm:"type:uuid;not null;index"`
	OptionText  string    `json:"option_text"`
	OptionImage string    `json:"option_image"`
	IsCorrect   bool      `json:"is_correct" gorm:"default:false"`
	OrderNumber int       `json:"order_number" gorm:"default:1"`
	Feedback    string    `json:"feedback"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (o *QuestionOption) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

type QuestionAsset struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	QuestionID uuid.UUID `json:"question_id" gorm:"type:uuid;not null;index"`
	AssetType string    `json:"asset_type" gorm:"size:20;default:image"`
	FileURL   string    `json:"file_url" gorm:"not null"`
	Caption   string    `json:"caption" gorm:"size:255"`
	AltText   string    `json:"alt_text" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *QuestionAsset) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
