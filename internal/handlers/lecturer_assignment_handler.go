package handlers

import (
	"net/http"

	"github.com/google/uuid"

	"kslasbackend/internal/models"
)

func (h *AssessmentHandler) listLecturerCourseAssignments(w http.ResponseWriter, r *http.Request) {
	query := h.db.Preload("Course").Preload("Course.Department").Order("created_at desc")

	if lecturerID := r.URL.Query().Get("lecturer_id"); lecturerID != "" {
		query = query.Where("lecturer_id = ?", lecturerID)
	}
	if courseID := r.URL.Query().Get("course_id"); courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var assignments []models.LecturerCourseAssignment
	if err := query.Find(&assignments).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, assignments)
}

func (h *AssessmentHandler) createLecturerCourseAssignment(w http.ResponseWriter, r *http.Request) {
	var assignment models.LecturerCourseAssignment
	if err := decodeJSON(w, r, &assignment); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if assignment.LecturerID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "lecturer_id is required")
		return
	}
	if assignment.CourseID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "course_id is required")
		return
	}

	var course models.Course
	if err := h.db.First(&course, "id = ?", assignment.CourseID).Error; err != nil {
		writeError(w, http.StatusBadRequest, "course not found")
		return
	}
	if assignment.DepartmentID == nil && course.DepartmentID != uuid.Nil {
		assignment.DepartmentID = &course.DepartmentID
	}

	if err := h.db.Create(&assignment).Error; err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.db.Preload("Course").Preload("Course.Department").First(&assignment, "id = ?", assignment.ID).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, assignment)
}

func (h *AssessmentHandler) listAssignedCoursesForLecturer(w http.ResponseWriter, r *http.Request) {
	lecturerID := r.URL.Query().Get("lecturer_id")
	if lecturerID == "" {
		writeError(w, http.StatusBadRequest, "lecturer_id query parameter is required")
		return
	}

	var assignments []models.LecturerCourseAssignment
	if err := h.db.Preload("Course").Preload("Course.Department").Where(
		"lecturer_id = ? AND status = ?",
		lecturerID,
		"active",
	).Order("created_at desc").Find(&assignments).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	courses := make([]models.Course, 0, len(assignments))
	for _, assignment := range assignments {
		courses = append(courses, assignment.Course)
	}
	writeJSON(w, http.StatusOK, courses)
}

func (h *AssessmentHandler) lecturerHasActiveCourseAssignment(lecturerID uuid.UUID, courseID uuid.UUID) bool {
	var count int64
	h.db.Model(&models.LecturerCourseAssignment{}).Where(
		"lecturer_id = ? AND course_id = ? AND status = ?",
		lecturerID,
		courseID,
		"active",
	).Count(&count)
	return count > 0
}
