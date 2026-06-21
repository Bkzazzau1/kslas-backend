package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"kslasbackend/internal/models"
)

type moderationPayload struct {
	ActorID  *uuid.UUID `json:"actor_id"`
	Comment  string     `json:"comment"`
	Feedback string     `json:"feedback"`
}

func (h *AssessmentHandler) listModeratorAssessments(w http.ResponseWriter, r *http.Request) {
	statuses := []string{"submitted_to_moderator", "returned_for_correction"}
	if status := r.URL.Query().Get("status"); status != "" {
		statuses = []string{status}
	}

	var assessments []models.Assessment
	if err := h.db.Preload("Course").Preload("Questions.Options").Where("status IN ?", statuses).Order("updated_at desc").Find(&assessments).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, assessments)
}

func (h *AssessmentHandler) moderatorAssessmentAction(w http.ResponseWriter, r *http.Request) {
	id, action, ok := splitIDAction(r.URL.Path, "/api/moderator/assessments/")
	if !ok {
		writeError(w, http.StatusNotFound, "invalid moderator assessment action")
		return
	}

	var payload moderationPayload
	if err := decodeOptionalJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var assessment models.Assessment
	if err := h.db.First(&assessment, "id = ?", id).Error; err != nil {
		writeError(w, http.StatusNotFound, "assessment not found")
		return
	}

	switch action {
	case "approve":
		if assessment.Status != "submitted_to_moderator" {
			writeError(w, http.StatusBadRequest, "only assessments submitted to moderator can be approved")
			return
		}
		fromStatus := assessment.Status
		assessment.Status = "approved_by_moderator"
		assessment.ModerationStatus = assessment.Status
		assessment.ModeratorID = payload.ActorID
		assessment.ModerationFeedback = firstNonEmpty(payload.Feedback, payload.Comment)
		assessment.ModeratedAt = nowPtr()
		if err := h.saveAssessmentWithAction(&assessment, payload.ActorID, "moderator_approved", fromStatus, assessment.Status, assessment.ModerationFeedback); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	case "return":
		if assessment.Status != "submitted_to_moderator" {
			writeError(w, http.StatusBadRequest, "only assessments submitted to moderator can be returned")
			return
		}
		fromStatus := assessment.Status
		assessment.Status = "returned_for_correction"
		assessment.ModerationStatus = assessment.Status
		assessment.ModeratorID = payload.ActorID
		assessment.ModerationFeedback = firstNonEmpty(payload.Feedback, payload.Comment)
		assessment.ModeratedAt = nowPtr()
		if err := h.saveAssessmentWithAction(&assessment, payload.ActorID, "moderator_returned", fromStatus, assessment.Status, assessment.ModerationFeedback); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	default:
		writeError(w, http.StatusNotFound, "unknown moderator action")
		return
	}

	writeJSON(w, http.StatusOK, assessment)
}

func (h *AssessmentHandler) listExamOfficerAssessments(w http.ResponseWriter, r *http.Request) {
	statuses := []string{"submitted_to_exam_officer", "approved_for_exam"}
	if status := r.URL.Query().Get("status"); status != "" {
		statuses = []string{status}
	}

	var assessments []models.Assessment
	if err := h.db.Preload("Course").Preload("Questions.Options").Where("status IN ?", statuses).Order("updated_at desc").Find(&assessments).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, assessments)
}

func (h *AssessmentHandler) examOfficerAssessmentAction(w http.ResponseWriter, r *http.Request) {
	id, action, ok := splitIDAction(r.URL.Path, "/api/exam-officer/assessments/")
	if !ok {
		writeError(w, http.StatusNotFound, "invalid exam officer assessment action")
		return
	}

	var payload moderationPayload
	if err := decodeOptionalJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var assessment models.Assessment
	if err := h.db.First(&assessment, "id = ?", id).Error; err != nil {
		writeError(w, http.StatusNotFound, "assessment not found")
		return
	}

	switch action {
	case "approve":
		if assessment.Status != "submitted_to_exam_officer" {
			writeError(w, http.StatusBadRequest, "only assessments submitted to exam officer can be approved")
			return
		}
		fromStatus := assessment.Status
		assessment.Status = "approved_for_exam"
		assessment.ModerationStatus = assessment.Status
		assessment.ExamOfficerID = payload.ActorID
		assessment.ExamOfficerFeedback = firstNonEmpty(payload.Feedback, payload.Comment)
		assessment.ExamOfficerApprovedAt = nowPtr()
		if err := h.saveAssessmentWithAction(&assessment, payload.ActorID, "exam_officer_approved", fromStatus, assessment.Status, assessment.ExamOfficerFeedback); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	case "return":
		if assessment.Status != "submitted_to_exam_officer" {
			writeError(w, http.StatusBadRequest, "only assessments submitted to exam officer can be returned")
			return
		}
		fromStatus := assessment.Status
		assessment.Status = "returned_for_correction"
		assessment.ModerationStatus = assessment.Status
		assessment.ExamOfficerID = payload.ActorID
		assessment.ExamOfficerFeedback = firstNonEmpty(payload.Feedback, payload.Comment)
		if err := h.saveAssessmentWithAction(&assessment, payload.ActorID, "exam_officer_returned", fromStatus, assessment.Status, assessment.ExamOfficerFeedback); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	default:
		writeError(w, http.StatusNotFound, "unknown exam officer action")
		return
	}

	writeJSON(w, http.StatusOK, assessment)
}

func (h *AssessmentHandler) saveAssessmentWithAction(assessment *models.Assessment, actorID *uuid.UUID, action string, fromStatus string, toStatus string, comment string) error {
	return h.db.Transaction(func(tx interface{ Error() error }) error {
		return nil
	})
}

func (h *AssessmentHandler) saveAssessmentAction(assessment *models.Assessment, actorID *uuid.UUID, action string, fromStatus string, toStatus string, comment string) error {
	if err := h.db.Save(assessment).Error; err != nil {
		return err
	}
	moderationAction := models.AssessmentModerationAction{
		AssessmentID: assessment.ID,
		ActorID:      actorID,
		Action:       action,
		FromStatus:   fromStatus,
		ToStatus:     toStatus,
		Comment:      comment,
	}
	return h.db.Create(&moderationAction).Error
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
