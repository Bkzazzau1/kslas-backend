package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"kslasbackend/internal/models"
)

type createExamActivityPayload struct {
	StudentID       string                   `json:"student_id"`
	StudentName     string                   `json:"student_name"`
	MatricNumber    string                   `json:"matric_number"`
	CourseCode      string                   `json:"course_code"`
	CourseTitle     string                   `json:"course_title"`
	AssessmentTitle string                   `json:"assessment_title"`
	SessionID       string                   `json:"session_id"`
	AttemptID       string                   `json:"attempt_id"`
	EventType       string                   `json:"event_type"`
	Severity        string                   `json:"severity"`
	AlertLevel      string                   `json:"alert_level"`
	RiskLevel       string                   `json:"risk_level"`
	RiskPoints      int                      `json:"risk_points"`
	AttentionLevel  string                   `json:"attention_level"`
	Message         string                   `json:"message"`
	Metadata        map[string]any           `json:"metadata"`
	EvidenceFiles   []activityEvidencePayload `json:"evidence_files"`
}

type activityEvidencePayload struct {
	Title        string         `json:"title"`
	EvidenceType string         `json:"evidence_type"`
	SourceKey    string         `json:"source_key"`
	FileURL      string         `json:"file_url"`
	Status       string         `json:"status"`
	Detail       string         `json:"detail"`
	Metadata     map[string]any `json:"metadata"`
}

func (h *AssessmentHandler) createExamActivityRecord(w http.ResponseWriter, r *http.Request) {
	var payload createExamActivityPayload
	if err := decodeJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	payload.normalize()
	if payload.EventType == "" {
		writeError(w, http.StatusBadRequest, "event_type is required")
		return
	}
	if payload.AttemptID == "" && payload.SessionID == "" {
		writeError(w, http.StatusBadRequest, "attempt_id or session_id is required")
		return
	}

	record := models.ExamActivityRecord{
		StudentID:       payload.StudentID,
		StudentName:     payload.StudentName,
		MatricNumber:    payload.MatricNumber,
		CourseCode:      payload.CourseCode,
		CourseTitle:     payload.CourseTitle,
		AssessmentTitle: payload.AssessmentTitle,
		SessionID:       payload.SessionID,
		AttemptID:       payload.AttemptID,
		EventType:       payload.EventType,
		AlertLevel:      payload.AlertLevel,
		RiskLevel:       payload.RiskLevel,
		RiskPoints:      payload.RiskPoints,
		AttentionLevel:  payload.AttentionLevel,
		Message:         payload.Message,
		Metadata:        jsonFromAny(payload.Metadata),
	}

	evidenceFiles := evidenceFilesFromActivity(payload)
	var reviewCase *models.ReviewEvidenceCase

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		if shouldCreateReviewCase(payload, evidenceFiles) {
			createdCase, err := upsertReviewCaseFromActivity(tx, payload)
			if err != nil {
				return err
			}
			reviewCase = createdCase
			timeline := models.ReviewEvidenceTimelineItem{
				CaseID:     createdCase.ID,
				EventType:  payload.EventType,
				Message:    payload.Message,
				AlertLevel: payload.AlertLevel,
				Metadata:   jsonFromAny(payload.Metadata),
			}
			if err := tx.Create(&timeline).Error; err != nil {
				return err
			}
			for _, file := range evidenceFiles {
				file.CaseID = createdCase.ID
				if err := tx.Create(&file).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]any{
		"activity_record": record,
		"review_case_created": reviewCase != nil,
	}
	if reviewCase != nil {
		h.db.
			Preload("EvidenceFiles").
			Preload("Timeline", func(db *gorm.DB) *gorm.DB { return db.Order("event_time asc") }).
			Preload("Actions", func(db *gorm.DB) *gorm.DB { return db.Order("created_at asc") }).
			First(reviewCase, "id = ?", reviewCase.ID)
		response["review_case"] = reviewCase
	}
	writeJSON(w, http.StatusCreated, response)
}

func (p *createExamActivityPayload) normalize() {
	p.StudentID = strings.TrimSpace(p.StudentID)
	p.StudentName = strings.TrimSpace(p.StudentName)
	p.MatricNumber = strings.TrimSpace(p.MatricNumber)
	p.CourseCode = strings.TrimSpace(p.CourseCode)
	p.CourseTitle = strings.TrimSpace(p.CourseTitle)
	p.AssessmentTitle = strings.TrimSpace(p.AssessmentTitle)
	p.SessionID = strings.TrimSpace(p.SessionID)
	p.AttemptID = strings.TrimSpace(p.AttemptID)
	p.EventType = strings.TrimSpace(p.EventType)
	p.AlertLevel = firstNonEmpty(strings.TrimSpace(p.AlertLevel), strings.TrimSpace(p.Severity), "medium")
	p.RiskLevel = strings.TrimSpace(p.RiskLevel)
	p.AttentionLevel = firstNonEmpty(strings.TrimSpace(p.AttentionLevel), attentionLevelForRecord(p.RiskPoints, p.AlertLevel))
	p.Message = firstNonEmpty(strings.TrimSpace(p.Message), "Activity record saved for review.")
	if p.Metadata == nil {
		p.Metadata = map[string]any{}
	}
}

func shouldCreateReviewCase(payload createExamActivityPayload, evidenceFiles []models.ReviewEvidenceFile) bool {
	level := strings.ToLower(payload.AlertLevel)
	return payload.RiskPoints > 0 || len(evidenceFiles) > 0 || level == "medium" || level == "warning" || level == "high" || level == "critical" || level == "urgent"
}

func upsertReviewCaseFromActivity(tx *gorm.DB, payload createExamActivityPayload) (*models.ReviewEvidenceCase, error) {
	var reviewCase models.ReviewEvidenceCase
	query := tx.Where("status NOT IN ?", []string{"finalized"}).Order("updated_at desc")
	if payload.AttemptID != "" {
		query = query.Where("attempt_id = ?", payload.AttemptID)
	} else {
		query = query.Where("session_id = ?", payload.SessionID)
	}

	err := query.First(&reviewCase).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		reviewCase = models.ReviewEvidenceCase{
			StudentID:       payload.StudentID,
			StudentName:     firstNonEmpty(payload.StudentName, "Student"),
			MatricNumber:    payload.MatricNumber,
			CourseCode:      payload.CourseCode,
			CourseTitle:     payload.CourseTitle,
			AssessmentTitle: payload.AssessmentTitle,
			SessionID:       payload.SessionID,
			AttemptID:       payload.AttemptID,
			AttentionLevel:  payload.AttentionLevel,
			Status:          "awaiting_review",
			RiskLevel:       payload.RiskLevel,
			RiskPoints:      payload.RiskPoints,
			ReviewSummary:   reviewSummaryForActivity(payload),
			AssignedRole:    "exam_officer",
			Metadata:        jsonFromAny(payload.Metadata),
		}
		if err := tx.Create(&reviewCase).Error; err != nil {
			return nil, err
		}
		return &reviewCase, nil
	}

	if payload.RiskPoints > reviewCase.RiskPoints {
		reviewCase.RiskPoints = payload.RiskPoints
		reviewCase.RiskLevel = payload.RiskLevel
		reviewCase.AttentionLevel = payload.AttentionLevel
	}
	if reviewCase.Status == "cleared" || reviewCase.Status == "marked_incorrect_alert" {
		reviewCase.Status = "awaiting_review"
	}
	if reviewCase.ReviewSummary == "" {
		reviewCase.ReviewSummary = reviewSummaryForActivity(payload)
	}
	if err := tx.Save(&reviewCase).Error; err != nil {
		return nil, err
	}
	return &reviewCase, nil
}

func evidenceFilesFromActivity(payload createExamActivityPayload) []models.ReviewEvidenceFile {
	files := make([]models.ReviewEvidenceFile, 0, len(payload.EvidenceFiles)+3)
	for _, file := range payload.EvidenceFiles {
		files = append(files, models.ReviewEvidenceFile{
			Title:        firstNonEmpty(file.Title, reviewEvidenceTitle(file.SourceKey)),
			EvidenceType: firstNonEmpty(file.EvidenceType, evidenceTypeForSource(file.SourceKey)),
			SourceKey:    firstNonEmpty(file.SourceKey, "review_record"),
			FileURL:      file.FileURL,
			Status:       firstNonEmpty(file.Status, "available"),
			Detail:       firstNonEmpty(file.Detail, "Evidence record saved for review."),
			Metadata:     jsonFromAny(file.Metadata),
		})
	}

	for _, sourceKey := range []string{"local_audio_record", "local_camera_record", "local_record"} {
		value, exists := payload.Metadata[sourceKey]
		if !exists || value == nil {
			continue
		}
		files = append(files, models.ReviewEvidenceFile{
			Title:        reviewEvidenceTitle(sourceKey),
			EvidenceType: evidenceTypeForSource(sourceKey),
			SourceKey:    sourceKey,
			Status:       "available",
			Detail:       "Evidence record saved by the secure exam app.",
			Metadata:     jsonFromAny(value),
		})
	}
	return files
}

func reviewEvidenceTitle(sourceKey string) string {
	switch sourceKey {
	case "local_audio_record":
		return "Room sound record"
	case "local_camera_record":
		return "Camera review record"
	case "local_record":
		return "Activity review record"
	default:
		return "Evidence record"
	}
}

func evidenceTypeForSource(sourceKey string) string {
	switch sourceKey {
	case "local_audio_record":
		return "sound_evidence"
	case "local_camera_record":
		return "camera_evidence"
	default:
		return "activity_record"
	}
}

func reviewSummaryForActivity(payload createExamActivityPayload) string {
	return fmt.Sprintf("%s Activity records are available for human review.", payload.Message)
}

func attentionLevelForRecord(points int, alertLevel string) string {
	level := strings.ToLower(alertLevel)
	if points >= 81 || level == "critical" || level == "urgent" {
		return "urgent_review_required"
	}
	if points >= 51 || level == "high" {
		return "high_attention_required"
	}
	return "medium_attention_required"
}

func jsonFromAny(value any) datatypes.JSON {
	if value == nil {
		return datatypes.JSON([]byte("{}"))
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(bytes)
}
