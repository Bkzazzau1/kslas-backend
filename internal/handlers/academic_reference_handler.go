package handlers

import (
	"net/http"

	"kslasbackend/internal/models"
)

func (h *AssessmentHandler) listDepartments(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "admin", "dlc_director", "hod", "exam_officer", "lecturer", "moderator", "academic_records") {
		return
	}
	query := h.db.Where("is_active = ?", true).Order("name asc")
	var departments []models.Department
	if err := query.Find(&departments).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, departments)
}

func (h *AssessmentHandler) listCourses(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "admin", "dlc_director", "hod", "exam_officer", "lecturer", "moderator", "academic_records") {
		return
	}
	query := h.db.Preload("Department").Where("courses.is_active = ?", true).Order("code asc")
	if departmentID := r.URL.Query().Get("department_id"); departmentID != "" {
		query = query.Where("department_id = ?", departmentID)
	}
	var courses []models.Course
	if err := query.Find(&courses).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, courses)
}
