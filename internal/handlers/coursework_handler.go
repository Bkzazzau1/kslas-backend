package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"kslasbackend/internal/middleware"
	"kslasbackend/internal/models"
)

func (h *AssessmentHandler) listLecturerAssignments(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "lecturer", "exam_officer", "hod", "dlc_director", "admin") { return }
	query := h.db.Preload("Course").Order("created_at desc")
	if lecturerID := r.URL.Query().Get("lecturer_id"); lecturerID != "" { query = query.Where("created_by_id = ?", lecturerID) }
	if courseID := r.URL.Query().Get("course_id"); courseID != "" { query = query.Where("course_id = ?", courseID) }
	if status := r.URL.Query().Get("status"); status != "" { query = query.Where("status = ?", status) }
	var assignments []models.LecturerAssignment
	if err := query.Find(&assignments).Error; err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	writeJSON(w, http.StatusOK, assignments)
}

func (h *AssessmentHandler) createLecturerAssignment(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "lecturer") { return }
	claims, _ := middleware.StaffClaimsFromRequest(r)
	var assignment models.LecturerAssignment
	if err := decodeJSON(w, r, &assignment); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	assignment.CreatedByID = claims.ID
	if !h.lecturerHasActiveCourseAssignment(claims.ID, assignment.CourseID) { writeError(w, http.StatusForbidden, "lecturer is not assigned to this course"); return }
	if err := h.db.Create(&assignment).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusCreated, assignment)
}

func (h *AssessmentHandler) assignmentAction(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "lecturer") { return }
	id, action, ok := splitIDAction(r.URL.Path, "/api/lecturer/assignments/")
	if !ok { writeError(w, http.StatusNotFound, "invalid assignment action"); return }
	var assignment models.LecturerAssignment
	if err := h.db.First(&assignment, "id = ?", id).Error; err != nil { writeError(w, http.StatusNotFound, "assignment not found"); return }
	switch action {
	case "publish": assignment.Status = "published"
	case "close": assignment.Status = "closed"
	default: writeError(w, http.StatusNotFound, "unknown assignment action"); return
	}
	if err := h.db.Save(&assignment).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, assignment)
}

func (h *AssessmentHandler) listAssignmentSubmissions(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "lecturer", "hod", "dlc_director", "admin") { return }
	query := h.db.Preload("Assignment").Order("created_at desc")
	if assignmentID := r.URL.Query().Get("assignment_id"); assignmentID != "" { query = query.Where("assignment_id = ?", assignmentID) }
	if studentID := r.URL.Query().Get("student_id"); studentID != "" { query = query.Where("student_id = ?", studentID) }
	var submissions []models.StudentAssignmentSubmission
	if err := query.Find(&submissions).Error; err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	writeJSON(w, http.StatusOK, submissions)
}

func (h *AssessmentHandler) markAssignmentSubmission(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "lecturer") { return }
	idText := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/lecturer/assignment-submissions/"), "/")
	idText = strings.TrimSuffix(idText, "/mark")
	idText = strings.Trim(idText, "/")
	id, err := uuid.Parse(idText)
	if err != nil { writeError(w, http.StatusBadRequest, "invalid submission id"); return }
	var payload struct { Score float64 `json:"score"`; Feedback string `json:"feedback"` }
	if err := decodeJSON(w, r, &payload); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	claims, _ := middleware.StaffClaimsFromRequest(r)
	var submission models.StudentAssignmentSubmission
	if err := h.db.Preload("Assignment").First(&submission, "id = ?", id).Error; err != nil { writeError(w, http.StatusNotFound, "submission not found"); return }
	submission.Score = &payload.Score
	submission.Feedback = payload.Feedback
	submission.MarkedByID = &claims.ID
	submission.MarkedAt = nowPtr()
	submission.Status = "marked"
	if submission.Assignment.FeedbackEnabled { submission.FeedbackReleasedAt = nowPtr() }
	if err := h.db.Save(&submission).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, submission)
}

func (h *AssessmentHandler) submitCA(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "lecturer") { return }
	claims, _ := middleware.StaffClaimsFromRequest(r)
	var ca models.CASubmission
	if err := decodeJSON(w, r, &ca); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	ca.LecturerID = claims.ID
	if !h.lecturerHasActiveCourseAssignment(claims.ID, ca.CourseID) { writeError(w, http.StatusForbidden, "lecturer is not assigned to this course"); return }
	if err := h.db.Create(&ca).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	h.notifyStaffRole("exam_officer", h.courseDepartmentID(ca.CourseID), &claims.ID, "CA submitted", "A lecturer submitted CA records to Exam Officer.", "ca", "high", "/exam-officer/ca-submissions", "ca_submission", &ca.ID)
	writeJSON(w, http.StatusCreated, ca)
}

func (h *AssessmentHandler) listCASubmissions(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "lecturer", "exam_officer", "hod", "dlc_director", "admin") { return }
	query := h.db.Preload("Course").Order("created_at desc")
	if lecturerID := r.URL.Query().Get("lecturer_id"); lecturerID != "" { query = query.Where("lecturer_id = ?", lecturerID) }
	if courseID := r.URL.Query().Get("course_id"); courseID != "" { query = query.Where("course_id = ?", courseID) }
	var items []models.CASubmission
	if err := query.Find(&items).Error; err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	writeJSON(w, http.StatusOK, items)
}

func (h *AssessmentHandler) submitMarkedExamScripts(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "lecturer") { return }
	claims, _ := middleware.StaffClaimsFromRequest(r)
	var item models.MarkedExamScriptSubmission
	if err := decodeJSON(w, r, &item); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	item.LecturerID = claims.ID
	if !h.lecturerHasActiveCourseAssignment(claims.ID, item.CourseID) { writeError(w, http.StatusForbidden, "lecturer is not assigned to this course"); return }
	if err := h.db.Create(&item).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	h.notifyStaffRole("exam_officer", h.courseDepartmentID(item.CourseID), &claims.ID, "Marked exam scripts submitted", "A lecturer submitted marked exam scripts to Exam Officer.", "exam_scripts", "high", "/exam-officer/marked-exam-scripts", "marked_exam_script", &item.ID)
	writeJSON(w, http.StatusCreated, item)
}

func (h *AssessmentHandler) listMarkedExamScripts(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "lecturer", "exam_officer", "hod", "dlc_director", "admin") { return }
	query := h.db.Preload("Assessment").Order("created_at desc")
	if lecturerID := r.URL.Query().Get("lecturer_id"); lecturerID != "" { query = query.Where("lecturer_id = ?", lecturerID) }
	if courseID := r.URL.Query().Get("course_id"); courseID != "" { query = query.Where("course_id = ?", courseID) }
	var items []models.MarkedExamScriptSubmission
	if err := query.Find(&items).Error; err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	writeJSON(w, http.StatusOK, items)
}
