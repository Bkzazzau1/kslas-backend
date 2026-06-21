package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"kslasbackend/internal/middleware"
	"kslasbackend/internal/models"
)

type LecturerAnalyticsOverview struct {
	LecturerID uuid.UUID                    `json:"lecturer_id"`
	Summary    LecturerAnalyticsSummary     `json:"summary"`
	Courses    []LecturerCourseAnalytics    `json:"courses"`
	Assessments LecturerAssessmentAnalytics `json:"assessments"`
	Assignments LecturerAssignmentAnalytics `json:"assignments"`
	Marking    LecturerMarkingAnalytics     `json:"marking"`
	CA         LecturerSubmissionAnalytics  `json:"ca"`
	ExamScripts LecturerSubmissionAnalytics `json:"exam_scripts"`
}

type LecturerAnalyticsSummary struct {
	AssignedCourses       int     `json:"assigned_courses"`
	TeachingHoursPerWeek  float64 `json:"teaching_hours_per_week"`
	AssessmentsCreated    int64   `json:"assessments_created"`
	AssignmentsCreated    int64   `json:"assignments_created"`
	PendingMarking        int64   `json:"pending_marking"`
	SubmissionsReceived   int64   `json:"submissions_received"`
	CASubmitted           int64   `json:"ca_submitted"`
	MarkedScriptsSubmitted int64  `json:"marked_scripts_submitted"`
}

type LecturerCourseAnalytics struct {
	CourseID             uuid.UUID `json:"course_id"`
	CourseCode           string    `json:"course_code"`
	CourseTitle          string    `json:"course_title"`
	AcademicSession      string    `json:"academic_session"`
	Semester             string    `json:"semester"`
	Level                string    `json:"level"`
	TeachingHoursPerWeek float64   `json:"teaching_hours_per_week"`
	AssessmentsCreated   int64     `json:"assessments_created"`
	AssignmentsCreated   int64     `json:"assignments_created"`
	SubmissionsReceived  int64     `json:"submissions_received"`
	CASubmitted          int64     `json:"ca_submitted"`
	MarkedScriptsSubmitted int64   `json:"marked_scripts_submitted"`
}

type LecturerAssessmentAnalytics struct {
	Total       int64 `json:"total"`
	Draft       int64 `json:"draft"`
	Submitted   int64 `json:"submitted_to_exam_officer"`
	Approved    int64 `json:"approved_for_exam"`
	Published   int64 `json:"published"`
	Closed      int64 `json:"closed"`
	Returned    int64 `json:"returned_to_lecturer"`
	Questions   int64 `json:"questions"`
}

type LecturerAssignmentAnalytics struct {
	Total       int64 `json:"total"`
	Draft       int64 `json:"draft"`
	Published   int64 `json:"published"`
	Closed      int64 `json:"closed"`
	Submissions int64 `json:"submissions"`
	Marked      int64 `json:"marked"`
}

type LecturerMarkingAnalytics struct {
	ExamSubmissions     int64   `json:"exam_submissions"`
	StudentAnswers      int64   `json:"student_answers"`
	PendingManualMarking int64  `json:"pending_manual_marking"`
	MarkedAnswers       int64   `json:"marked_answers"`
	AverageScore        float64 `json:"average_score"`
}

type LecturerSubmissionAnalytics struct {
	Total     int64 `json:"total"`
	Submitted int64 `json:"submitted_to_exam_officer"`
	Accepted  int64 `json:"accepted"`
	Returned  int64 `json:"returned"`
}

func (h *AssessmentHandler) lecturerAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "lecturer", "hod", "dlc_director", "admin", "exam_officer") {
		return
	}
	lecturerID, ok := h.resolveLecturerIDForAnalytics(w, r)
	if !ok {
		return
	}

	overview := LecturerAnalyticsOverview{LecturerID: lecturerID}
	overview.Courses = h.buildLecturerCourseAnalytics(lecturerID)
	overview.Assessments = h.buildLecturerAssessmentAnalytics(lecturerID)
	overview.Assignments = h.buildLecturerAssignmentAnalytics(lecturerID)
	overview.Marking = h.buildLecturerMarkingAnalytics(lecturerID)
	overview.CA = h.buildLecturerCASummary(lecturerID)
	overview.ExamScripts = h.buildLecturerExamScriptSummary(lecturerID)

	overview.Summary.AssignedCourses = len(overview.Courses)
	for _, course := range overview.Courses {
		overview.Summary.TeachingHoursPerWeek += course.TeachingHoursPerWeek
	}
	overview.Summary.AssessmentsCreated = overview.Assessments.Total
	overview.Summary.AssignmentsCreated = overview.Assignments.Total
	overview.Summary.PendingMarking = overview.Marking.PendingManualMarking
	overview.Summary.SubmissionsReceived = overview.Marking.ExamSubmissions + overview.Assignments.Submissions
	overview.Summary.CASubmitted = overview.CA.Total
	overview.Summary.MarkedScriptsSubmitted = overview.ExamScripts.Total

	writeJSON(w, http.StatusOK, overview)
}

func (h *AssessmentHandler) resolveLecturerIDForAnalytics(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	claims, ok := middleware.StaffClaimsFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "staff authentication is required")
		return uuid.Nil, false
	}
	lecturerIDText := r.URL.Query().Get("lecturer_id")
	if lecturerIDText == "" {
		return claims.ID, true
	}
	lecturerID, err := uuid.Parse(lecturerIDText)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid lecturer_id")
		return uuid.Nil, false
	}
	if lecturerID != claims.ID && claims.HasAnyRole(map[string]bool{"hod": true, "dlc_director": true, "admin": true, "exam_officer": true}) == false {
		writeError(w, http.StatusForbidden, "not allowed to view another lecturer analytics")
		return uuid.Nil, false
	}
	return lecturerID, true
}

func (h *AssessmentHandler) buildLecturerCourseAnalytics(lecturerID uuid.UUID) []LecturerCourseAnalytics {
	var assignments []models.LecturerCourseAssignment
	h.db.Preload("Course").Where("lecturer_id = ? AND status = ?", lecturerID, "active").Order("created_at desc").Find(&assignments)

	courses := make([]LecturerCourseAnalytics, 0, len(assignments))
	for _, assignment := range assignments {
		item := LecturerCourseAnalytics{
			CourseID:             assignment.CourseID,
			CourseCode:           assignment.Course.Code,
			CourseTitle:          assignment.Course.Title,
			AcademicSession:      assignment.AcademicSession,
			Semester:             assignment.Semester,
			Level:                assignment.Level,
			TeachingHoursPerWeek: assignment.TeachingHoursPerWeek,
		}
		h.db.Model(&models.Assessment{}).Where("created_by_id = ? AND course_id = ?", lecturerID, assignment.CourseID).Count(&item.AssessmentsCreated)
		h.db.Model(&models.LecturerAssignment{}).Where("created_by_id = ? AND course_id = ?", lecturerID, assignment.CourseID).Count(&item.AssignmentsCreated)
		h.db.Table("student_submissions").Joins("JOIN assessments ON assessments.id = student_submissions.assessment_id").Where("assessments.created_by_id = ? AND assessments.course_id = ?", lecturerID, assignment.CourseID).Count(&item.SubmissionsReceived)
		h.db.Model(&models.CASubmission{}).Where("lecturer_id = ? AND course_id = ?", lecturerID, assignment.CourseID).Count(&item.CASubmitted)
		h.db.Model(&models.MarkedExamScriptSubmission{}).Where("lecturer_id = ? AND course_id = ?", lecturerID, assignment.CourseID).Count(&item.MarkedScriptsSubmitted)
		courses = append(courses, item)
	}
	return courses
}

func (h *AssessmentHandler) buildLecturerAssessmentAnalytics(lecturerID uuid.UUID) LecturerAssessmentAnalytics {
	var item LecturerAssessmentAnalytics
	h.db.Model(&models.Assessment{}).Where("created_by_id = ?", lecturerID).Count(&item.Total)
	h.db.Model(&models.Assessment{}).Where("created_by_id = ? AND status = ?", lecturerID, "draft").Count(&item.Draft)
	h.db.Model(&models.Assessment{}).Where("created_by_id = ? AND status = ?", lecturerID, "submitted_to_exam_officer").Count(&item.Submitted)
	h.db.Model(&models.Assessment{}).Where("created_by_id = ? AND status = ?", lecturerID, "approved_for_exam").Count(&item.Approved)
	h.db.Model(&models.Assessment{}).Where("created_by_id = ? AND status = ?", lecturerID, "published").Count(&item.Published)
	h.db.Model(&models.Assessment{}).Where("created_by_id = ? AND status = ?", lecturerID, "closed").Count(&item.Closed)
	h.db.Model(&models.Assessment{}).Where("created_by_id = ? AND status = ?", lecturerID, "returned_to_lecturer").Count(&item.Returned)
	h.db.Table("questions").Joins("JOIN assessments ON assessments.id = questions.assessment_id").Where("assessments.created_by_id = ?", lecturerID).Count(&item.Questions)
	return item
}

func (h *AssessmentHandler) buildLecturerAssignmentAnalytics(lecturerID uuid.UUID) LecturerAssignmentAnalytics {
	var item LecturerAssignmentAnalytics
	h.db.Model(&models.LecturerAssignment{}).Where("created_by_id = ?", lecturerID).Count(&item.Total)
	h.db.Model(&models.LecturerAssignment{}).Where("created_by_id = ? AND status = ?", lecturerID, "draft").Count(&item.Draft)
	h.db.Model(&models.LecturerAssignment{}).Where("created_by_id = ? AND status = ?", lecturerID, "published").Count(&item.Published)
	h.db.Model(&models.LecturerAssignment{}).Where("created_by_id = ? AND status = ?", lecturerID, "closed").Count(&item.Closed)
	h.db.Table("student_assignment_submissions").Joins("JOIN lecturer_assignments ON lecturer_assignments.id = student_assignment_submissions.assignment_id").Where("lecturer_assignments.created_by_id = ?", lecturerID).Count(&item.Submissions)
	h.db.Table("student_assignment_submissions").Joins("JOIN lecturer_assignments ON lecturer_assignments.id = student_assignment_submissions.assignment_id").Where("lecturer_assignments.created_by_id = ? AND student_assignment_submissions.status = ?", lecturerID, "marked").Count(&item.Marked)
	return item
}

func (h *AssessmentHandler) buildLecturerMarkingAnalytics(lecturerID uuid.UUID) LecturerMarkingAnalytics {
	var item LecturerMarkingAnalytics
	h.db.Table("student_submissions").Joins("JOIN assessments ON assessments.id = student_submissions.assessment_id").Where("assessments.created_by_id = ?", lecturerID).Count(&item.ExamSubmissions)
	h.db.Table("student_answers").Joins("JOIN questions ON questions.id = student_answers.question_id").Joins("JOIN assessments ON assessments.id = questions.assessment_id").Where("assessments.created_by_id = ?", lecturerID).Count(&item.StudentAnswers)
	h.db.Table("student_answers").Joins("JOIN questions ON questions.id = student_answers.question_id").Joins("JOIN assessments ON assessments.id = questions.assessment_id").Where("assessments.created_by_id = ? AND student_answers.marking_status IN ?", lecturerID, []string{"pending_manual", "pending", "needs_review"}).Count(&item.PendingManualMarking)
	h.db.Table("student_answers").Joins("JOIN questions ON questions.id = student_answers.question_id").Joins("JOIN assessments ON assessments.id = questions.assessment_id").Where("assessments.created_by_id = ? AND student_answers.marking_status = ?", lecturerID, "marked").Count(&item.MarkedAnswers)

type avgResult struct { AverageScore float64 }
	var avg avgResult
	h.db.Table("student_answers").Select("COALESCE(AVG(final_score), 0) AS average_score").Joins("JOIN questions ON questions.id = student_answers.question_id").Joins("JOIN assessments ON assessments.id = questions.assessment_id").Where("assessments.created_by_id = ?", lecturerID).Scan(&avg)
	item.AverageScore, _ = strconv.ParseFloat(strconv.FormatFloat(avg.AverageScore, 'f', 2, 64), 64)
	return item
}

func (h *AssessmentHandler) buildLecturerCASummary(lecturerID uuid.UUID) LecturerSubmissionAnalytics {
	var item LecturerSubmissionAnalytics
	h.db.Model(&models.CASubmission{}).Where("lecturer_id = ?", lecturerID).Count(&item.Total)
	h.db.Model(&models.CASubmission{}).Where("lecturer_id = ? AND status = ?", lecturerID, "submitted_to_exam_officer").Count(&item.Submitted)
	h.db.Model(&models.CASubmission{}).Where("lecturer_id = ? AND status = ?", lecturerID, "accepted").Count(&item.Accepted)
	h.db.Model(&models.CASubmission{}).Where("lecturer_id = ? AND status IN ?", lecturerID, []string{"returned", "returned_to_lecturer"}).Count(&item.Returned)
	return item
}

func (h *AssessmentHandler) buildLecturerExamScriptSummary(lecturerID uuid.UUID) LecturerSubmissionAnalytics {
	var item LecturerSubmissionAnalytics
	h.db.Model(&models.MarkedExamScriptSubmission{}).Where("lecturer_id = ?", lecturerID).Count(&item.Total)
	h.db.Model(&models.MarkedExamScriptSubmission{}).Where("lecturer_id = ? AND status = ?", lecturerID, "submitted_to_exam_officer").Count(&item.Submitted)
	h.db.Model(&models.MarkedExamScriptSubmission{}).Where("lecturer_id = ? AND status = ?", lecturerID, "accepted").Count(&item.Accepted)
	h.db.Model(&models.MarkedExamScriptSubmission{}).Where("lecturer_id = ? AND status IN ?", lecturerID, []string{"returned", "returned_to_lecturer"}).Count(&item.Returned)
	return item
}
