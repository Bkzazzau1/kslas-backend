package handlers

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"kslasbackend/internal/models"
)

type LecturerAssessmentView struct {
	ID                    uuid.UUID      `json:"id"`
	CourseID              uuid.UUID      `json:"course_id"`
	Course                models.Course  `json:"course"`
	Title                 string         `json:"title"`
	Description           string         `json:"description"`
	AssessmentType        string         `json:"assessment_type"`
	DurationMinutes       int            `json:"duration_minutes"`
	TotalMarks            float64        `json:"total_marks"`
	StartTime             *time.Time     `json:"start_time"`
	EndTime               *time.Time     `json:"end_time"`
	CreatedByID           *uuid.UUID     `json:"created_by_id"`
	Status                string         `json:"status"`
	ModerationStatus      string         `json:"moderation_status"`
	ExamOfficerFeedback   string         `json:"exam_officer_feedback,omitempty"`
	SubmittedForReviewAt  *time.Time     `json:"submitted_for_review_at,omitempty"`
	ExamOfficerApprovedAt *time.Time     `json:"exam_officer_approved_at,omitempty"`
	ProctoringLevel       string         `json:"proctoring_level"`
	AllowMobile           bool           `json:"allow_mobile"`
	ShuffleQuestions      bool           `json:"shuffle_questions"`
	ShuffleOptions        bool           `json:"shuffle_options"`
	ShowResultImmediately bool           `json:"show_result_immediately"`
	Rules                 datatypes.JSON `json:"rules"`
	Questions             []models.Question `json:"questions,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

func toLecturerAssessmentViews(assessments []models.Assessment) []LecturerAssessmentView {
	views := make([]LecturerAssessmentView, 0, len(assessments))
	for _, assessment := range assessments {
		views = append(views, LecturerAssessmentView{
			ID:                    assessment.ID,
			CourseID:              assessment.CourseID,
			Course:                assessment.Course,
			Title:                 assessment.Title,
			Description:           assessment.Description,
			AssessmentType:        assessment.AssessmentType,
			DurationMinutes:       assessment.DurationMinutes,
			TotalMarks:            assessment.TotalMarks,
			StartTime:             assessment.StartTime,
			EndTime:               assessment.EndTime,
			CreatedByID:           assessment.CreatedByID,
			Status:                lecturerVisibleStatus(assessment),
			ModerationStatus:      lecturerVisibleStatus(assessment),
			ExamOfficerFeedback:   lecturerVisibleFeedback(assessment),
			SubmittedForReviewAt:  assessment.SubmittedForReviewAt,
			ExamOfficerApprovedAt: assessment.ExamOfficerApprovedAt,
			ProctoringLevel:       assessment.ProctoringLevel,
			AllowMobile:           assessment.AllowMobile,
			ShuffleQuestions:      assessment.ShuffleQuestions,
			ShuffleOptions:        assessment.ShuffleOptions,
			ShowResultImmediately: assessment.ShowResultImmediately,
			Rules:                 assessment.Rules,
			Questions:             assessment.Questions,
			CreatedAt:             assessment.CreatedAt,
			UpdatedAt:             assessment.UpdatedAt,
		})
	}
	return views
}

func lecturerVisibleStatus(assessment models.Assessment) string {
	switch assessment.Status {
	case "draft", "returned_to_lecturer", "approved_for_exam", "published", "closed":
		return assessment.Status
	default:
		return "submitted_to_exam_officer"
	}
}

func lecturerVisibleFeedback(assessment models.Assessment) string {
	if assessment.Status == "returned_to_lecturer" {
		return assessment.ExamOfficerFeedback
	}
	return ""
}
