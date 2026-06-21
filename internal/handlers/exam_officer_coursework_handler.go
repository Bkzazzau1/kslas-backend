package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"kslasbackend/internal/middleware"
	"kslasbackend/internal/models"
)

func (h *AssessmentHandler) reviewCASubmission(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "exam_officer") { return }
	idText := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/exam-officer/ca-submissions/"), "/")
	idText = strings.TrimSuffix(idText, "/review")
	idText = strings.Trim(idText, "/")
	id, err := uuid.Parse(idText)
	if err != nil { writeError(w, http.StatusBadRequest, "invalid CA submission id"); return }
	var payload struct { Status string `json:"status"`; Feedback string `json:"feedback"` }
	if err := decodeJSON(w, r, &payload); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	if payload.Status == "" { payload.Status = "accepted" }
	claims, _ := middleware.StaffClaimsFromRequest(r)
	var item models.CASubmission
	if err := h.db.First(&item, "id = ?", id).Error; err != nil { writeError(w, http.StatusNotFound, "CA submission not found"); return }
	item.Status = payload.Status
	item.ExamOfficerID = &claims.ID
	item.ExamOfficerFeedback = payload.Feedback
	item.ReviewedAt = nowPtr()
	if err := h.db.Save(&item).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	h.createNotification(item.LecturerID, &claims.ID, "CA reviewed", "Exam Officer reviewed your CA submission.", "ca", "normal", "/lecturer/ca-submissions", "ca_submission", &item.ID)
	writeJSON(w, http.StatusOK, item)
}

func (h *AssessmentHandler) reviewMarkedExamScripts(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "exam_officer") { return }
	idText := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/exam-officer/marked-exam-scripts/"), "/")
	idText = strings.TrimSuffix(idText, "/review")
	idText = strings.Trim(idText, "/")
	id, err := uuid.Parse(idText)
	if err != nil { writeError(w, http.StatusBadRequest, "invalid marked script submission id"); return }
	var payload struct { Status string `json:"status"`; Feedback string `json:"feedback"` }
	if err := decodeJSON(w, r, &payload); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	if payload.Status == "" { payload.Status = "accepted" }
	claims, _ := middleware.StaffClaimsFromRequest(r)
	var item models.MarkedExamScriptSubmission
	if err := h.db.First(&item, "id = ?", id).Error; err != nil { writeError(w, http.StatusNotFound, "marked exam script submission not found"); return }
	item.Status = payload.Status
	item.ExamOfficerID = &claims.ID
	item.ExamOfficerFeedback = payload.Feedback
	item.ReviewedAt = nowPtr()
	if err := h.db.Save(&item).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	h.createNotification(item.LecturerID, &claims.ID, "Marked scripts reviewed", "Exam Officer reviewed your marked exam script submission.", "exam_scripts", "normal", "/lecturer/marked-exam-scripts", "marked_exam_script", &item.ID)
	writeJSON(w, http.StatusOK, item)
}
