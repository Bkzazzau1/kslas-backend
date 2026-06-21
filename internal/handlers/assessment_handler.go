package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"kslasbackend/internal/models"
)

type AssessmentHandler struct { db *gorm.DB }

func NewAssessmentHandler(db *gorm.DB) *AssessmentHandler { return &AssessmentHandler{db: db} }

func (h *AssessmentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /api/lecturer/assessments", h.listAssessments)
	mux.HandleFunc("POST /api/lecturer/assessments", h.createAssessment)
	mux.HandleFunc("POST /api/lecturer/assessments/", h.assessmentAction)
	mux.HandleFunc("GET /api/lecturer/questions", h.listQuestions)
	mux.HandleFunc("POST /api/lecturer/questions", h.createQuestion)
	mux.HandleFunc("POST /api/lecturer/options", h.createOption)
	mux.HandleFunc("POST /api/lecturer/assets", h.createAsset)
	mux.HandleFunc("GET /api/lecturer/submissions", h.listSubmissions)
	mux.HandleFunc("PATCH /api/lecturer/answers/", h.markAnswer)
	mux.HandleFunc("GET /api/lecturer/course-assignments", h.listLecturerCourseAssignments)
	mux.HandleFunc("GET /api/lecturer/courses", h.listAssignedCoursesForLecturer)
	mux.HandleFunc("GET /api/admin/lecturer-course-assignments", h.listLecturerCourseAssignments)
	mux.HandleFunc("POST /api/admin/lecturer-course-assignments", h.createLecturerCourseAssignment)
	mux.HandleFunc("GET /api/moderator/assessments", h.listModeratorAssessments)
	mux.HandleFunc("POST /api/moderator/assessments/", h.moderatorAssessmentAction)
	mux.HandleFunc("GET /api/exam-officer/assessments", h.listExamOfficerAssessments)
	mux.HandleFunc("POST /api/exam-officer/assessments/", h.examOfficerAssessmentAction)
	mux.HandleFunc("GET /api/student/assessments", h.listPublishedAssessments)
	mux.HandleFunc("POST /api/student/assessments/", h.studentAssessmentAction)
	mux.HandleFunc("POST /api/student/answers", h.submitAnswer)
}

func (h *AssessmentHandler) health(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status": "ok"}) }

func (h *AssessmentHandler) listAssessments(w http.ResponseWriter, r *http.Request) {
	query := h.db.Preload("Course").Preload("Questions.Options").Preload("Questions.Assets").Order("updated_at desc")
	if lecturerID := r.URL.Query().Get("lecturer_id"); lecturerID != "" { query = query.Where("created_by_id = ?", lecturerID) }
	if courseID := r.URL.Query().Get("course_id"); courseID != "" { query = query.Where("course_id = ?", courseID) }
	if status := r.URL.Query().Get("status"); status != "" { query = query.Where("status = ?", status) }
	var assessments []models.Assessment
	if err := query.Find(&assessments).Error; err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	writeJSON(w, http.StatusOK, assessments)
}

func (h *AssessmentHandler) createAssessment(w http.ResponseWriter, r *http.Request) {
	var assessment models.Assessment
	if err := decodeJSON(w, r, &assessment); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	if assessment.CreatedByID != nil && !h.lecturerHasActiveCourseAssignment(*assessment.CreatedByID, assessment.CourseID) { writeError(w, http.StatusForbidden, "not assigned to this course"); return }
	if err := h.db.Create(&assessment).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusCreated, assessment)
}

func (h *AssessmentHandler) assessmentAction(w http.ResponseWriter, r *http.Request) {
	id, action, ok := splitIDAction(r.URL.Path, "/api/lecturer/assessments/")
	if !ok { writeError(w, http.StatusNotFound, "invalid action"); return }
	var payload moderationPayload
	if err := decodeOptionalJSON(w, r, &payload); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	var assessment models.Assessment
	if err := h.db.First(&assessment, "id = ?", id).Error; err != nil { writeError(w, http.StatusNotFound, "assessment not found"); return }
	fromStatus := assessment.Status
	switch action {
	case "submit-to-exam-officer":
		if assessment.Status != "draft" && assessment.Status != "returned_to_lecturer" { writeError(w, http.StatusBadRequest, "not ready for submission"); return }
		assessment.Status = "submitted_to_exam_officer"
		assessment.ModerationStatus = "submitted_to_exam_officer"
		assessment.SubmittedForReviewAt = nowPtr()
	case "publish":
		if assessment.Status != "approved_for_exam" { writeError(w, http.StatusBadRequest, "not approved for publishing"); return }
		assessment.Status = "published"
		assessment.ModerationStatus = "published"
	case "close":
		assessment.Status = "closed"
		assessment.ModerationStatus = "closed"
	default:
		writeError(w, http.StatusNotFound, "unknown action"); return
	}
	if err := h.saveAssessmentWithAction(&assessment, payload.ActorID, action, fromStatus, assessment.Status, firstNonEmpty(payload.Feedback, payload.Comment)); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, assessment)
}

func splitIDAction(path string, prefix string) (uuid.UUID, string, bool) {
	trimmed := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 { return uuid.Nil, "", false }
	id, err := uuid.Parse(parts[0])
	if err != nil { return uuid.Nil, "", false }
	return id, parts[1], true
}

func nowPtr() *time.Time { now := time.Now(); return &now }
