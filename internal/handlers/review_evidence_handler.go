package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"kslasbackend/internal/models"
)

type reviewEvidenceActionPayload struct {
	ActorID   *uuid.UUID `json:"actor_id"`
	ActorRole string     `json:"actor_role"`
	Comment   string     `json:"comment"`
}

func (h *AssessmentHandler) listReviewEvidenceCases(w http.ResponseWriter, r *http.Request) {
	query := h.db.
		Preload("EvidenceFiles").
		Preload("Timeline", func(db any) any { return db }).
		Preload("Actions").
		Order("updated_at desc")

	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if attentionLevel := strings.TrimSpace(r.URL.Query().Get("attention_level")); attentionLevel != "" {
		query = query.Where("attention_level = ?", attentionLevel)
	}
	if role := strings.TrimSpace(r.URL.Query().Get("assigned_role")); role != "" {
		query = query.Where("assigned_role = ?", role)
	}
	if courseCode := strings.TrimSpace(r.URL.Query().Get("course_code")); courseCode != "" {
		query = query.Where("course_code = ?", courseCode)
	}
	if attemptID := strings.TrimSpace(r.URL.Query().Get("attempt_id")); attemptID != "" {
		query = query.Where("attempt_id = ?", attemptID)
	}

	var cases []models.ReviewEvidenceCase
	if err := query.Find(&cases).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cases)
}

func (h *AssessmentHandler) createReviewEvidenceCase(w http.ResponseWriter, r *http.Request) {
	var reviewCase models.ReviewEvidenceCase
	if err := decodeJSON(w, r, &reviewCase); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if reviewCase.StudentName == "" {
		writeError(w, http.StatusBadRequest, "student_name is required")
		return
	}
	if reviewCase.ReviewSummary == "" {
		reviewCase.ReviewSummary = "Activity records are available for human review."
	}
	if err := h.db.Create(&reviewCase).Error; err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.db.Preload("EvidenceFiles").Preload("Timeline").Preload("Actions").First(&reviewCase, "id = ?", reviewCase.ID)
	writeJSON(w, http.StatusCreated, reviewCase)
}

func (h *AssessmentHandler) reviewEvidenceCaseAction(w http.ResponseWriter, r *http.Request) {
	id, action, ok := splitIDAction(r.URL.Path, "/api/review/evidence-cases/")
	if !ok {
		writeError(w, http.StatusNotFound, "invalid action")
		return
	}

	var payload reviewEvidenceActionPayload
	if err := decodeOptionalJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var reviewCase models.ReviewEvidenceCase
	if err := h.db.First(&reviewCase, "id = ?", id).Error; err != nil {
		writeError(w, http.StatusNotFound, "review case not found")
		return
	}

	fromStatus := reviewCase.Status
	toStatus, ok := reviewEvidenceNextStatus(action)
	if !ok {
		writeError(w, http.StatusNotFound, "unknown action")
		return
	}
	reviewCase.Status = toStatus
	reviewCase.UpdatedAt = time.Now()

	actionRow := models.ReviewEvidenceAction{
		CaseID:     reviewCase.ID,
		ActorID:    payload.ActorID,
		ActorRole:  payload.ActorRole,
		Action:     action,
		Comment:    payload.Comment,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
	}

	if err := h.db.Transaction(func(tx any) error {
		db := tx.(interface {
			Save(value any) *gorm.DB
			Create(value any) *gorm.DB
		})
		if err := db.Save(&reviewCase).Error; err != nil {
			return err
		}
		if err := db.Create(&actionRow).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.db.Preload("EvidenceFiles").Preload("Timeline").Preload("Actions").First(&reviewCase, "id = ?", reviewCase.ID)
	writeJSON(w, http.StatusOK, reviewCase)
}

func reviewEvidenceNextStatus(action string) (string, bool) {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "mark-incorrect-alert":
		return "marked_incorrect_alert", true
	case "clear":
		return "cleared", true
	case "send-to-hod":
		return "sent_to_hod", true
	case "request-student-explanation":
		return "requires_student_explanation", true
	case "finalize":
		return "finalized", true
	default:
		return "", false
	}
}

func metadataJSON(value map[string]any) datatypes.JSON {
	if value == nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON([]byte("{}"))
}
