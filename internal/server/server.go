package server

import (
	"encoding/json"
	"net/http"

	"kslasbackend/internal/handlers"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/services"
)

type Dependencies struct {
	AuthHandler                *handlers.AuthHandler
	AcademicHandler            *handlers.AcademicHandler
	AdminHandler               *handlers.AdministrationHandler
	MaterialHandler            *handlers.MaterialHandler
	AssignmentHandler          *handlers.AssignmentHandler
	ForumHandler               *handlers.ForumHandler
	MessageHandler             *handlers.DirectMessageHandler
	ContentHandler             *handlers.TeachingContentHandler
	ExamHandler                *handlers.ExamHandler
	InvigilatorEvidenceHandler *handlers.InvigilatorEvidenceHandler
	ProctoringReviewHandler    *handlers.ProctoringReviewHandler
	IdentityHandler            *handlers.IdentityHandler
	ResultHandler              *handlers.ResultHandler
	ReportHandler              *handlers.ReportHandler
	JWTService                 *services.JWTService
	PermissionService          *services.PermissionService
}

func NewRouter(dep *Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/health", method(http.MethodGet, http.HandlerFunc(handlers.HealthHandler)))
	mux.Handle("/healthz", method(http.MethodGet, http.HandlerFunc(handlers.HealthHandler)))
	mux.Handle("/api/auth/login", method(http.MethodPost, http.HandlerFunc(dep.AuthHandler.Login)))
	mux.Handle(
		"/api/auth/me",
		chain(
			method(http.MethodGet, http.HandlerFunc(dep.AuthHandler.Me)),
			middleware.AuthMiddleware(dep.JWTService),
		),
	)
	mux.Handle(
		"/api/auth/change-password",
		chain(
			method(http.MethodPost, http.HandlerFunc(dep.AuthHandler.ChangePassword)),
			middleware.AuthMiddleware(dep.JWTService),
		),
	)
	mux.Handle(
		"/api/academic/ping",
		chain(
			method(http.MethodGet, http.HandlerFunc(academicPingHandler)),
			middleware.AuthMiddleware(dep.JWTService),
			middleware.RequirePermission(dep.PermissionService, "faculty.view", nil),
		),
	)
	mux.Handle("/api/faculties", chain(http.HandlerFunc(dep.AcademicHandler.Faculties), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/faculties/{facultyID}", chain(http.HandlerFunc(dep.AcademicHandler.FacultyByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/departments", chain(http.HandlerFunc(dep.AcademicHandler.Departments), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/departments/{departmentID}", chain(http.HandlerFunc(dep.AcademicHandler.DepartmentByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/programmes", chain(http.HandlerFunc(dep.AcademicHandler.Programmes), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/programmes/{programmeID}", chain(http.HandlerFunc(dep.AcademicHandler.ProgrammeByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/courses", chain(http.HandlerFunc(dep.AcademicHandler.Courses), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/courses/{courseID}", chain(http.HandlerFunc(dep.AcademicHandler.CourseByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/staff", chain(http.HandlerFunc(dep.AdminHandler.Staff), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/staff/{staffID}/reset-password", chain(http.HandlerFunc(dep.AdminHandler.StaffPasswordReset), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/staff/{staffID}/status", chain(http.HandlerFunc(dep.AdminHandler.StaffStatus), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/admin/staff-roles", chain(http.HandlerFunc(dep.AdminHandler.StaffRoleAssignment), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/students", chain(http.HandlerFunc(dep.AdminHandler.Students), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/courses/{courseID}/lecturers", chain(http.HandlerFunc(dep.AdminHandler.CourseLecturers), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/courses/{courseID}/forum/posts", chain(http.HandlerFunc(dep.ForumHandler.CourseForumPosts), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/courses/{courseID}/forum/posts/{postID}", chain(http.HandlerFunc(dep.ForumHandler.CourseForumPostByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/courses/{courseID}/messages", chain(http.HandlerFunc(dep.MessageHandler.CourseMessages), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/my/eligible-courses", chain(http.HandlerFunc(dep.AdminHandler.EligibleCourses), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/my/course-registrations", chain(http.HandlerFunc(dep.AdminHandler.MyCourseRegistrations), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/materials", chain(http.HandlerFunc(dep.MaterialHandler.Materials), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/materials/{materialID}", chain(http.HandlerFunc(dep.MaterialHandler.MaterialByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/materials/{materialID}/publish", chain(http.HandlerFunc(dep.MaterialHandler.MaterialPublish), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/video-lectures", chain(http.HandlerFunc(dep.ContentHandler.VideoLectures), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/video-lectures/{lectureID}", chain(http.HandlerFunc(dep.ContentHandler.VideoLectureByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/video-lectures/{lectureID}/watch", chain(http.HandlerFunc(dep.ContentHandler.VideoLectureWatch), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/live-sessions", chain(http.HandlerFunc(dep.ContentHandler.LiveSessions), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/live-sessions/{sessionID}", chain(http.HandlerFunc(dep.ContentHandler.LiveSessionByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/assignments", chain(http.HandlerFunc(dep.AssignmentHandler.Assignments), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/assignments/{assignmentID}", chain(http.HandlerFunc(dep.AssignmentHandler.AssignmentByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/assignments/{assignmentID}/submissions", chain(http.HandlerFunc(dep.AssignmentHandler.AssignmentSubmissions), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/assignments/{assignmentID}/grades", chain(http.HandlerFunc(dep.AssignmentHandler.AssignmentGrades), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/assignments/{assignmentID}/peer-reviews/{reviewID}", chain(http.HandlerFunc(dep.AssignmentHandler.AssignmentPeerReview), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exams", chain(http.HandlerFunc(dep.ExamHandler.Exams), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exam-venues", chain(http.HandlerFunc(dep.ExamHandler.Venues), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exams/{examID}", chain(http.HandlerFunc(dep.ExamHandler.ExamByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exams/{examID}/release", chain(http.HandlerFunc(dep.ExamHandler.ExamRelease), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exams/{examID}/attempts", chain(http.HandlerFunc(dep.ExamHandler.ExamAttempts), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exams/{examID}/students", chain(http.HandlerFunc(dep.ExamHandler.ExamStudents), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exams/{examID}/venue-allocations", chain(http.HandlerFunc(dep.ExamHandler.ExamVenueAllocations), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exam-attempts/{attemptID}/submit", chain(http.HandlerFunc(dep.ExamHandler.AttemptSubmit), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exam-attempts/{attemptID}/share-with-lecturer", chain(http.HandlerFunc(dep.ExamHandler.AttemptShareWithLecturer), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exam-attempts/{attemptID}/mark", chain(http.HandlerFunc(dep.ExamHandler.AttemptMark), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exam-attempts/{attemptID}/moderate", chain(http.HandlerFunc(dep.ExamHandler.AttemptModerate), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exam-attempts/{attemptID}/script.pdf", chain(http.HandlerFunc(dep.ExamHandler.AttemptScriptPDF), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exam-attempts/{attemptID}/annotations", chain(http.HandlerFunc(dep.ExamHandler.AttemptAnnotations), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/exam-attempts/{attemptID}/proctoring-alerts", chain(http.HandlerFunc(dep.ExamHandler.AttemptAlerts), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/invigilator/alerts", chain(http.HandlerFunc(dep.ExamHandler.InvigilatorAlerts), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/invigilator/alerts/{alertID}/acknowledge", chain(http.HandlerFunc(dep.ExamHandler.AlertAcknowledge), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/invigilator/evidence", chain(http.HandlerFunc(dep.InvigilatorEvidenceHandler.Queue), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/invigilator/evidence/{caseID}/decision", chain(http.HandlerFunc(dep.InvigilatorEvidenceHandler.Decision), middleware.AuthMiddleware(dep.JWTService)))

	mux.Handle("/api/proctoring/pre-exam-review", method(http.MethodPost, http.HandlerFunc(dep.ProctoringReviewHandler.PreExamReview)))
	mux.Handle("/api/proctoring/start-approval", method(http.MethodPost, http.HandlerFunc(dep.ProctoringReviewHandler.StartApproval)))
	mux.Handle("/api/proctoring/live-events", method(http.MethodPost, http.HandlerFunc(dep.ProctoringReviewHandler.LiveEvent)))
	mux.Handle("/api/proctoring/random-video-samples", method(http.MethodPost, http.HandlerFunc(dep.ProctoringReviewHandler.RandomVideoSample)))
	mux.Handle("/api/proctoring/random-video-samples/list", method(http.MethodGet, http.HandlerFunc(dep.ProctoringReviewHandler.RandomVideoSamples)))
	mux.Handle("/api/proctoring/random-video-samples/{sampleID}", method(http.MethodGet, http.HandlerFunc(dep.ProctoringReviewHandler.RandomVideoSampleByID)))
	mux.Handle("/api/proctoring/random-video-samples/{sampleID}/file", method(http.MethodGet, http.HandlerFunc(dep.ProctoringReviewHandler.RandomVideoSampleFile)))

	mux.Handle("/api/identity/face-enrollment", method(http.MethodPost, http.HandlerFunc(dep.IdentityHandler.FaceEnrollment)))
	mux.Handle("/api/identity/face-enrollments/latest", method(http.MethodGet, http.HandlerFunc(dep.IdentityHandler.FaceEnrollmentLatest)))
	mux.Handle("/api/identity/face-enrollments", method(http.MethodGet, http.HandlerFunc(dep.IdentityHandler.FaceEnrollments)))
	mux.Handle("/api/identity/face-enrollments/{enrollmentID}", method(http.MethodGet, http.HandlerFunc(dep.IdentityHandler.FaceEnrollmentByID)))
	mux.Handle("/api/identity/face-enrollments/{enrollmentID}/images/{fileName}", method(http.MethodGet, http.HandlerFunc(dep.IdentityHandler.FaceEnrollmentImage)))

	mux.Handle("/api/results", chain(http.HandlerFunc(dep.ResultHandler.Results), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/results/{resultID}", chain(http.HandlerFunc(dep.ResultHandler.ResultByID), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/results/{resultID}/approve", chain(http.HandlerFunc(dep.ResultHandler.ResultApprove), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/results/{resultID}/publish", chain(http.HandlerFunc(dep.ResultHandler.ResultPublish), middleware.AuthMiddleware(dep.JWTService)))
	mux.Handle("/api/hod/lecturer-reports", chain(http.HandlerFunc(dep.ReportHandler.LecturerReports), middleware.AuthMiddleware(dep.JWTService)))

	return mux
}

func chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

func method(expected string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expected {
			w.Header().Set("Allow", expected)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
				"message": "method not allowed",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func academicPingHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "academic route protected by permission",
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
