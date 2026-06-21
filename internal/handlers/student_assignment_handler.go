package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"kslasbackend/internal/models"
)

func (h *AssessmentHandler) listPublishedAssignmentsForStudents(w http.ResponseWriter, r *http.Request) {
	query := h.db.Preload("Course").Where("status = ?", "published").Order("created_at desc")
	if courseID := r.URL.Query().Get("course_id"); courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}
	var assignments []models.LecturerAssignment
	if err := query.Find(&assignments).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, assignments)
}

func (h *AssessmentHandler) submitStudentAssignment(w http.ResponseWriter, r *http.Request) {
	assignmentID, action, ok := splitIDAction(r.URL.Path, "/api/student/assignments/")
	if !ok || action != "submit" {
		writeError(w, http.StatusNotFound, "invalid student assignment action")
		return
	}

	studentID, err := uuid.Parse(r.URL.Query().Get("student_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "student_id query parameter is required")
		return
	}

	var assignment models.LecturerAssignment
	if err := h.db.First(&assignment, "id = ?", assignmentID).Error; err != nil {
		writeError(w, http.StatusNotFound, "assignment not found")
		return
	}
	if assignment.Status != "published" {
		writeError(w, http.StatusBadRequest, "assignment is not open for submission")
		return
	}

	var payload struct {
		TextSubmission string `json:"text_submission"`
		FileURL        string `json:"file_url"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.TextSubmission) == "" && strings.TrimSpace(payload.FileURL) == "" {
		writeError(w, http.StatusBadRequest, "text_submission or file_url is required")
		return
	}

	submission := models.StudentAssignmentSubmission{
		AssignmentID:   assignmentID,
		StudentID:      studentID,
		TextSubmission: payload.TextSubmission,
		FileURL:        payload.FileURL,
	}
	if err := h.db.Create(&submission).Error; err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, submission)
}
