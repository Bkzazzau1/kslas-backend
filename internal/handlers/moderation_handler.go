package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"kslasbackend/internal/models"
)

type moderationPayload struct {
	ActorID  *uuid.UUID `json:"actor_id"`
	Comment  string     `json:"comment"`
	Feedback string     `json:"feedback"`
}

func (h *AssessmentHandler) listModeratorAssessments(w http.ResponseWriter, r *http.Request) {
	statuses := []string{"sent_to_moderator"}
	if status := r.URL.Query().Get("moderation_status"); status != "" { statuses = []string{status} }
	var assessments []models.Assessment
	if err := h.db.Preload("Course").Preload("Questions.Options").Where("moderation_status IN ?", statuses).Order("updated_at desc").Find(&assessments).Error; err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	writeJSON(w, http.StatusOK, assessments)
}

func (h *AssessmentHandler) moderatorAssessmentAction(w http.ResponseWriter, r *http.Request) {
	id, action, ok := splitIDAction(r.URL.Path, "/api/moderator/assessments/")
	if !ok { writeError(w, http.StatusNotFound, "invalid moderator assessment action"); return }
	var payload moderationPayload
	if err := decodeOptionalJSON(w, r, &payload); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	var assessment models.Assessment
	if err := h.db.First(&assessment, "id = ?", id).Error; err != nil { writeError(w, http.StatusNotFound, "assessment not found"); return }
	if assessment.ModerationStatus != "sent_to_moderator" { writeError(w, http.StatusBadRequest, "assessment is not currently with moderator"); return }
	fromStatus := assessment.ModerationStatus
	switch action {
	case "approve":
		assessment.Status = "submitted_to_exam_officer"
		assessment.ModerationStatus = "moderator_approved_to_exam_officer"
		assessment.ModeratorID = payload.ActorID
		assessment.ModerationFeedback = firstNonEmpty(payload.Feedback, payload.Comment)
		assessment.ModeratedAt = nowPtr()
		if err := h.saveAssessmentWithAction(&assessment, payload.ActorID, "moderator_approved_to_exam_officer", fromStatus, assessment.ModerationStatus, assessment.ModerationFeedback); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
		h.notifyStaffRole("exam_officer", h.courseDepartmentID(assessment.CourseID), payload.ActorID, "Question moderation completed", "Moderator approved submitted questions and returned them to Exam Officer.", "exam_moderation", "high", "/exam-officer/assessments", "assessment", &assessment.ID)
	case "return":
		assessment.Status = "submitted_to_exam_officer"
		assessment.ModerationStatus = "moderator_returned_to_exam_officer"
		assessment.ModeratorID = payload.ActorID
		assessment.ModerationFeedback = firstNonEmpty(payload.Feedback, payload.Comment)
		assessment.ModeratedAt = nowPtr()
		if err := h.saveAssessmentWithAction(&assessment, payload.ActorID, "moderator_returned_to_exam_officer", fromStatus, assessment.ModerationStatus, assessment.ModerationFeedback); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
		h.notifyStaffRole("exam_officer", h.courseDepartmentID(assessment.CourseID), payload.ActorID, "Moderator returned questions", "Moderator returned submitted questions to Exam Officer for review.", "exam_moderation", "high", "/exam-officer/assessments", "assessment", &assessment.ID)
	default:
		writeError(w, http.StatusNotFound, "unknown moderator action"); return
	}
	writeJSON(w, http.StatusOK, assessment)
}

func (h *AssessmentHandler) listExamOfficerAssessments(w http.ResponseWriter, r *http.Request) {
	query := h.db.Preload("Course").Preload("Questions.Options").Order("updated_at desc")
	if status := r.URL.Query().Get("status"); status != "" { query = query.Where("status = ?", status) } else { query = query.Where("status IN ?", []string{"submitted_to_exam_officer", "approved_for_exam"}) }
	if moderationStatus := r.URL.Query().Get("moderation_status"); moderationStatus != "" { query = query.Where("moderation_status = ?", moderationStatus) }
	var assessments []models.Assessment
	if err := query.Find(&assessments).Error; err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	writeJSON(w, http.StatusOK, assessments)
}

func (h *AssessmentHandler) examOfficerAssessmentAction(w http.ResponseWriter, r *http.Request) {
	id, action, ok := splitIDAction(r.URL.Path, "/api/exam-officer/assessments/")
	if !ok { writeError(w, http.StatusNotFound, "invalid exam officer assessment action"); return }
	var payload moderationPayload
	if err := decodeOptionalJSON(w, r, &payload); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	var assessment models.Assessment
	if err := h.db.First(&assessment, "id = ?", id).Error; err != nil { writeError(w, http.StatusNotFound, "assessment not found"); return }
	fromStatus := assessment.ModerationStatus
	switch action {
	case "send-to-moderator":
		if assessment.Status != "submitted_to_exam_officer" { writeError(w, http.StatusBadRequest, "assessment must be submitted to exam officer first"); return }
		assessment.ModerationStatus = "sent_to_moderator"
		assessment.ExamOfficerID = payload.ActorID
		if err := h.saveAssessmentWithAction(&assessment, payload.ActorID, "exam_officer_sent_to_moderator", fromStatus, assessment.ModerationStatus, firstNonEmpty(payload.Feedback, payload.Comment)); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
		h.notifyStaffRole("moderator", h.courseDepartmentID(assessment.CourseID), payload.ActorID, "Questions sent for moderation", "Exam Officer sent questions for moderation.", "exam_moderation", "high", "/moderator/assessments", "assessment", &assessment.ID)
	case "approve":
		if assessment.ModerationStatus != "moderator_approved_to_exam_officer" { writeError(w, http.StatusBadRequest, "assessment must return approved from moderator first"); return }
		assessment.Status = "approved_for_exam"
		assessment.ModerationStatus = "approved_for_exam"
		assessment.ExamOfficerID = payload.ActorID
		assessment.ExamOfficerFeedback = firstNonEmpty(payload.Feedback, payload.Comment)
		assessment.ExamOfficerApprovedAt = nowPtr()
		if err := h.saveAssessmentWithAction(&assessment, payload.ActorID, "exam_officer_approved", fromStatus, assessment.ModerationStatus, assessment.ExamOfficerFeedback); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
		if assessment.CreatedByID != nil { h.createNotification(*assessment.CreatedByID, payload.ActorID, "Questions approved for exam", "Exam Officer approved the submitted questions for exam use.", "exam_moderation", "high", "/lecturer/assessments", "assessment", &assessment.ID) }
	case "return-to-lecturer":
		if assessment.Status != "submitted_to_exam_officer" && assessment.ModerationStatus != "moderator_returned_to_exam_officer" { writeError(w, http.StatusBadRequest, "only exam officer can return submitted work to lecturer"); return }
		assessment.Status = "returned_to_lecturer"
		assessment.ModerationStatus = "returned_to_lecturer"
		assessment.ExamOfficerID = payload.ActorID
		assessment.ExamOfficerFeedback = firstNonEmpty(payload.Feedback, payload.Comment)
		if err := h.saveAssessmentWithAction(&assessment, payload.ActorID, "exam_officer_returned_to_lecturer", fromStatus, assessment.ModerationStatus, assessment.ExamOfficerFeedback); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
		if assessment.CreatedByID != nil { h.createNotification(*assessment.CreatedByID, payload.ActorID, "Questions returned for correction", "Exam Officer returned the questions for correction.", "exam_moderation", "high", "/lecturer/assessments", "assessment", &assessment.ID) }
	default:
		writeError(w, http.StatusNotFound, "unknown exam officer action"); return
	}
	writeJSON(w, http.StatusOK, assessment)
}

func (h *AssessmentHandler) saveAssessmentWithAction(assessment *models.Assessment, actorID *uuid.UUID, action string, fromStatus string, toStatus string, comment string) error {
	return h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(assessment).Error; err != nil { return err }
		moderationAction := models.AssessmentModerationAction{AssessmentID: assessment.ID, ActorID: actorID, Action: action, FromStatus: fromStatus, ToStatus: toStatus, Comment: comment}
		return tx.Create(&moderationAction).Error
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values { if strings.TrimSpace(value) != "" { return value } }
	return ""
}
