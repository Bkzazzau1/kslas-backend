package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
	"kslasbackend/internal/dto"
	"kslasbackend/internal/rbac"
	"kslasbackend/internal/repository"
)

type ExamService struct {
	repo              *repository.TeachingRepository
	permissionService *PermissionService
}

func (s *ExamService) ListVenues(ctx context.Context, userID uint, activeOnly bool) ([]dto.ExamVenueResponse, error) {
	if err := s.ensurePermission(ctx, userID, "exam.schedule", scopePtr(rbac.SchoolScope())); err != nil {
		return nil, err
	}
	items, err := s.repo.ListExamVenues(ctx, activeOnly)
	if err != nil {
		return nil, err
	}
	return mapExamVenues(items), nil
}

func (s *ExamService) CreateVenue(ctx context.Context, userID uint, req dto.ExamVenueCreateRequest) (*dto.ExamVenueResponse, error) {
	if err := s.ensurePermission(ctx, userID, "exam.schedule", scopePtr(rbac.SchoolScope())); err != nil {
		return nil, err
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	item := &models.ExamVenue{
		Name:      req.Name,
		Address:   req.Address,
		City:      req.City,
		Country:   req.Country,
		Capacity:  req.Capacity,
		IsActive:  isActive,
		CreatedBy: userID,
	}
	if err := s.repo.CreateExamVenue(ctx, item); err != nil {
		return nil, err
	}
	response := mapExamVenue(item)
	return &response, nil
}

func NewExamService(repo *repository.TeachingRepository, permissionService *PermissionService) *ExamService {
	return &ExamService{repo: repo, permissionService: permissionService}
}

func (s *ExamService) ListExams(ctx context.Context, userID uint, filter repository.ExamListFilter) ([]dto.ExamResponse, error) {
	target := scopePtr(rbac.SchoolScope())
	if filter.CourseID != nil {
		target = scopePtr(rbac.CourseScope(*filter.CourseID))
	}
	if err := s.ensurePermission(ctx, userID, "course.view", target); err != nil {
		return nil, err
	}
	items, err := s.repo.ListExams(ctx, filter)
	if err != nil {
		return nil, err
	}
	return mapExams(items), nil
}

func (s *ExamService) CreateExam(ctx context.Context, userID uint, req dto.ExamCreateRequest) (*dto.ExamResponse, error) {
	courseID, err := s.resolveCourseID(ctx, req.CourseID, req.CourseCode)
	if err != nil {
		return nil, err
	}
	if err := validateExamRequest(courseID, req.Title, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}
	lecturerID := nonZero(req.LecturerID, userID)
	assignedLecturer, err := s.repo.LecturerAssignedToCourse(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	if !assignedLecturer {
		if err := s.ensurePermission(ctx, userID, "exam.create", scopePtr(rbac.CourseScope(courseID))); err != nil {
			return nil, err
		}
	}

	payload, err := encodeJSONBytes(req.QuestionPayload)
	if err != nil {
		return nil, ValidationError{Message: "question_payload must be valid json"}
	}
	if err := validateQuestionPayload(req.QuestionPayload); err != nil {
		return nil, err
	}
	status := examStatusOrDefault(req.Status)
	if status == models.ExamStatusReleased {
		status = models.ExamStatusScheduled
	}
	item := &models.Exam{
		CourseID:        courseID,
		Title:           req.Title,
		Description:     req.Description,
		Instructions:    req.Instructions,
		Venue:           req.Venue,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		DurationMinutes: req.DurationMinutes,
		DeliveryMode:    examDeliveryModeOrDefault(req.DeliveryMode),
		QuestionPayload: payload,
		LecturerID:      lecturerID,
		ExamOfficerID:   req.ExamOfficerID,
		Status:          status,
		CreatedBy:       userID,
	}
	if err := s.repo.CreateExam(ctx, item); err != nil {
		return nil, err
	}
	if len(req.InvigilatorIDs) > 0 {
		if _, err := s.repo.ReplaceExamInvigilators(ctx, item.ID, userID, req.InvigilatorIDs); err != nil {
			return nil, err
		}
	}
	item, err = s.repo.GetExam(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	response := mapExam(item)
	return &response, nil
}

func (s *ExamService) UpdateExam(ctx context.Context, userID, examID uint, req dto.ExamUpdateRequest) (*dto.ExamResponse, error) {
	item, err := s.repo.GetExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.schedule", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if item.Status == models.ExamStatusCompleted {
		return nil, ValidationError{Message: "completed exams cannot be edited"}
	}
	if err := validateExamRequest(item.CourseID, req.Title, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}
	payload, err := encodeJSONBytes(req.QuestionPayload)
	if err != nil {
		return nil, ValidationError{Message: "question_payload must be valid json"}
	}
	if err := validateQuestionPayload(req.QuestionPayload); err != nil {
		return nil, err
	}
	item.Title = req.Title
	item.Description = req.Description
	item.Instructions = req.Instructions
	item.Venue = req.Venue
	item.StartTime = req.StartTime
	item.EndTime = req.EndTime
	item.DurationMinutes = req.DurationMinutes
	item.DeliveryMode = examDeliveryModeOrDefault(req.DeliveryMode)
	item.QuestionPayload = payload
	item.ExamOfficerID = req.ExamOfficerID
	item.Status = examStatusOrDefault(req.Status)
	if err := s.repo.UpdateExam(ctx, item); err != nil {
		return nil, err
	}
	if req.InvigilatorIDs != nil {
		if _, err := s.repo.ReplaceExamInvigilators(ctx, item.ID, userID, req.InvigilatorIDs); err != nil {
			return nil, err
		}
	}
	item, err = s.repo.GetExam(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	response := mapExam(item)
	return &response, nil
}

func (s *ExamService) ReleaseExam(ctx context.Context, userID, examID uint) (*dto.ExamResponse, error) {
	item, err := s.repo.GetExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.schedule", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if len(item.Invigilators) == 0 {
		return nil, ValidationError{Message: "assign at least one invigilator before release"}
	}
	now := time.Now().UTC()
	item.Status = models.ExamStatusReleased
	item.ReleasedBy = userID
	item.ReleasedAt = &now
	if err := s.repo.UpdateExam(ctx, item); err != nil {
		return nil, err
	}
	response := mapExam(item)
	return &response, nil
}

func (s *ExamService) StartAttempt(ctx context.Context, userID, examID uint, req dto.ExamAttemptStartRequest) (*dto.ExamAttemptResponse, error) {
	exam, err := s.repo.GetExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.start", scopePtr(rbac.CourseScope(exam.CourseID))); err != nil {
		return nil, err
	}
	if exam.Status != models.ExamStatusReleased && exam.Status != models.ExamStatusScheduled {
		return nil, ValidationError{Message: "exam is not released"}
	}
	attempt, err := s.repo.GetExamAttemptForStudent(ctx, examID, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	now := time.Now().UTC()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		questionPayload, err := randomizedQuestionPayload(exam.QuestionPayload, examID, userID)
		if err != nil {
			return nil, err
		}
		attempt = &models.ExamAttempt{
			ExamID:          examID,
			StudentID:       userID,
			Status:          models.ExamAttemptInProgress,
			QuestionPayload: questionPayload,
			IntegrityScore:  100,
		}
	}
	if req.EnvironmentConfirmed {
		attempt.EnvironmentConfirmedAt = &now
		attempt.MonitoringArmedAt = &now
	}
	if err := s.repo.UpsertExamAttempt(ctx, attempt); err != nil {
		return nil, err
	}
	response := mapExamAttempt(attempt)
	return &response, nil
}

func (s *ExamService) SubmitAttempt(ctx context.Context, userID, attemptID uint, req dto.ExamAttemptSubmitRequest) (*dto.ExamAttemptResponse, error) {
	attempt, err := s.repo.GetExamAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if attempt.StudentID != userID {
		return nil, ErrPermissionDenied
	}
	if attempt.Status != models.ExamAttemptInProgress {
		return nil, ValidationError{Message: "submitted exams cannot be edited"}
	}
	payload, err := encodeJSONBytes(req.AnswerPayload)
	if err != nil {
		return nil, ValidationError{Message: "answer_payload must be valid json"}
	}
	now := time.Now().UTC()
	attempt.AnswerPayload = payload
	attempt.Status = models.ExamAttemptSubmitted
	attempt.SubmittedAt = &now
	attempt.SubmittedToOfficerAt = &now
	attempt.IntegrityScore = req.IntegrityScore
	attempt.TerminationReason = req.TerminationReason
	if attempt.IntegrityScore == 0 {
		attempt.IntegrityScore = 100
	}
	if err := s.repo.UpsertExamAttempt(ctx, attempt); err != nil {
		return nil, err
	}
	response := mapExamAttempt(attempt)
	return &response, nil
}

func (s *ExamService) ShareAttemptWithLecturer(ctx context.Context, userID, attemptID uint) (*dto.ExamAttemptResponse, error) {
	attempt, err := s.repo.GetExamAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.submit_review", scopePtr(rbac.CourseScope(attempt.Exam.CourseID))); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	attempt.Status = models.ExamAttemptSubmittedForMarking
	attempt.SharedWithLecturerAt = &now
	if err := s.repo.UpsertExamAttempt(ctx, attempt); err != nil {
		return nil, err
	}
	response := mapExamAttempt(attempt)
	return &response, nil
}

func (s *ExamService) MarkAttempt(ctx context.Context, userID, attemptID uint, req dto.ExamAttemptMarkRequest) (*dto.ExamAttemptResponse, error) {
	attempt, err := s.repo.GetExamAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if userID != attempt.Exam.LecturerID {
		if err := s.ensurePermission(ctx, userID, "result.mark", scopePtr(rbac.CourseScope(attempt.Exam.CourseID))); err != nil {
			return nil, err
		}
		if err := s.ensureExamOfficer(ctx, userID); err != nil {
			return nil, err
		}
	}
	if userID == attempt.Exam.LecturerID && attempt.Status != models.ExamAttemptSubmittedForMarking && attempt.Status != models.ExamAttemptMarked {
		return nil, ValidationError{Message: "script must be shared with lecturer before marking"}
	}
	if attempt.Status != models.ExamAttemptSubmitted && attempt.Status != models.ExamAttemptSubmittedForMarking && attempt.Status != models.ExamAttemptMarked {
		return nil, ValidationError{Message: "exam must be submitted before marking"}
	}
	now := time.Now().UTC()
	attempt.Status = models.ExamAttemptMarked
	attempt.Score = req.Score
	if attempt.LecturerScore == 0 {
		attempt.LecturerScore = req.Score
	}
	attempt.Feedback = req.Feedback
	attempt.MarkedBy = userID
	attempt.MarkedAt = &now
	if err := s.repo.UpsertExamAttempt(ctx, attempt); err != nil {
		return nil, err
	}
	response := mapExamAttempt(attempt)
	return &response, nil
}

func (s *ExamService) ModerateAttempt(ctx context.Context, userID, attemptID uint, req dto.ExamAttemptModerateRequest) (*dto.ExamAttemptResponse, error) {
	attempt, err := s.repo.GetExamAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "result.mark", scopePtr(rbac.CourseScope(attempt.Exam.CourseID))); err != nil {
		return nil, err
	}
	if err := s.ensureExamOfficer(ctx, userID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if attempt.LecturerScore == 0 {
		attempt.LecturerScore = attempt.Score
	}
	attempt.Score = req.Score
	attempt.ModeratedScore = req.Score
	attempt.ModerationComment = req.Comment
	attempt.ModeratedBy = userID
	attempt.ModeratedAt = &now
	attempt.MarkedBy = userID
	attempt.MarkedAt = &now
	attempt.Status = models.ExamAttemptMarked
	if err := s.repo.UpsertExamAttempt(ctx, attempt); err != nil {
		return nil, err
	}
	response := mapExamAttempt(attempt)
	return &response, nil
}

func (s *ExamService) ListRegisteredStudents(ctx context.Context, userID, examID uint) ([]dto.ExamStudentResponse, error) {
	exam, err := s.repo.GetExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.schedule", scopePtr(rbac.CourseScope(exam.CourseID))); err != nil {
		return nil, err
	}
	items, err := s.repo.ListCourseRegistrations(ctx, exam.CourseID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ExamStudentResponse, 0, len(items))
	for _, item := range items {
		profile, _ := s.repo.GetStudentAcademicProfile(ctx, item.StudentID)
		country := ""
		isInternational := false
		if profile != nil {
			country = profile.CountryOfResidence
			isInternational = profile.IsInternational || (country != "" && !strings.EqualFold(country, "Nigeria"))
		}
		mode := string(exam.DeliveryMode)
		if isInternational {
			mode = string(models.ExamDeliveryRemoteProctored)
		}
		out = append(out, dto.ExamStudentResponse{
			StudentID:               item.StudentID,
			Name:                    displayName(item.Student),
			MatricNo:                item.Student.MatricNo,
			Email:                   item.Student.Email,
			Level:                   item.Level,
			AcademicSession:         item.AcademicSession,
			CountryOfResidence:      country,
			IsInternational:         isInternational,
			RecommendedDeliveryMode: mode,
		})
	}
	return out, nil
}

func (s *ExamService) AllocateVenues(ctx context.Context, userID, examID uint, req dto.ExamVenueAllocationRequest) ([]dto.ExamStudentAllocationResponse, error) {
	exam, err := s.repo.GetExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.schedule", scopePtr(rbac.CourseScope(exam.CourseID))); err != nil {
		return nil, err
	}
	registrations, err := s.repo.ListCourseRegistrations(ctx, exam.CourseID)
	if err != nil {
		return nil, err
	}
	registered := map[uint]models.CourseRegistration{}
	for _, item := range registrations {
		registered[item.StudentID] = item
	}
	items := make([]models.ExamStudentAllocation, 0, len(req.Items))
	for _, input := range req.Items {
		reg, ok := registered[input.StudentID]
		if !ok {
			return nil, ValidationError{Message: "student is not registered for this course"}
		}
		profile, _ := s.repo.GetStudentAcademicProfile(ctx, input.StudentID)
		country := ""
		isInternational := false
		if profile != nil {
			country = profile.CountryOfResidence
			isInternational = profile.IsInternational || (country != "" && !strings.EqualFold(country, "Nigeria"))
		}
		mode := examDeliveryModeOrDefault(input.DeliveryMode)
		if strings.TrimSpace(input.DeliveryMode) == "" {
			mode = exam.DeliveryMode
		}
		if isInternational {
			mode = models.ExamDeliveryRemoteProctored
		}
		allocation := models.ExamStudentAllocation{
			StudentID:          reg.StudentID,
			DeliveryMode:       mode,
			VenueID:            input.VenueID,
			IsInternational:    isInternational,
			CountryOfResidence: country,
		}
		if input.VenueID != nil {
			venue, err := s.repo.GetExamVenue(ctx, *input.VenueID)
			if err != nil {
				return nil, err
			}
			allocation.VenueName = venue.Name
			allocation.VenueAddress = venue.Address
		}
		items = append(items, allocation)
	}
	allocations, err := s.repo.UpsertExamStudentAllocations(ctx, examID, userID, items)
	if err != nil {
		return nil, err
	}
	return mapExamStudentAllocations(allocations), nil
}

func (s *ExamService) ExamScriptPDF(ctx context.Context, userID, attemptID uint) (string, []byte, error) {
	attempt, err := s.repo.GetExamAttempt(ctx, attemptID)
	if err != nil {
		return "", nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.submit_review", scopePtr(rbac.CourseScope(attempt.Exam.CourseID))); err != nil {
		return "", nil, err
	}
	body := fmt.Sprintf("K-SLAS EXAM SCRIPT\nWATERMARK: EXAM OFFICER AUTHENTICATED COPY\n\nCourse: %s\nExam: %s\nStudent ID: %d\nScore: %.2f\nIntegrity score: %d\n\nSignature space: ______________________________\nExam officer: _________________________________\nDate: _________________________________________\n\nAnswers payload:\n%s\n",
		attempt.Exam.Course.Code,
		attempt.Exam.Title,
		attempt.StudentID,
		attempt.Score,
		attempt.IntegrityScore,
		string(attempt.AnswerPayload),
	)
	fileName := fmt.Sprintf("exam-script-%d-%d.pdf", attempt.ExamID, attempt.StudentID)
	return fileName, minimalPDF(body), nil
}

func (s *ExamService) AddScriptAnnotation(ctx context.Context, userID, attemptID uint, req dto.ExamScriptAnnotationCreateRequest) (*dto.ExamScriptAnnotationResponse, error) {
	attempt, err := s.repo.GetExamAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if attempt.Status != models.ExamAttemptSubmittedForMarking && attempt.Status != models.ExamAttemptMarked {
		return nil, ValidationError{Message: "exam script must be shared with lecturer before annotation"}
	}
	if err := s.ensureCourseLecturer(ctx, userID, attempt.Exam.CourseID); err != nil {
		return nil, err
	}
	ink, err := encodeJSONBytes(req.InkPayload)
	if err != nil {
		return nil, ValidationError{Message: "ink_payload must be valid json"}
	}
	item := &models.ExamScriptAnnotation{
		AttemptID:      attemptID,
		LecturerID:     userID,
		QuestionID:     req.QuestionID,
		Comment:        req.Comment,
		HighlightColor: req.HighlightColor,
		InkPayload:     ink,
		Score:          req.Score,
	}
	if err := s.repo.CreateExamScriptAnnotation(ctx, item); err != nil {
		return nil, err
	}
	response := mapExamScriptAnnotation(item)
	return &response, nil
}

func (s *ExamService) ListScriptAnnotations(ctx context.Context, userID, attemptID uint) ([]dto.ExamScriptAnnotationResponse, error) {
	attempt, err := s.repo.GetExamAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if attempt.StudentID != userID {
		if err := s.ensurePermission(ctx, userID, "result.mark", scopePtr(rbac.CourseScope(attempt.Exam.CourseID))); err != nil {
			return nil, err
		}
	}
	items, err := s.repo.ListExamScriptAnnotations(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	return mapExamScriptAnnotations(items), nil
}

func (s *ExamService) RecordProctoringAlert(ctx context.Context, userID, attemptID uint, req dto.ProctoringAlertCreateRequest) (*dto.ProctoringAlertResponse, error) {
	attempt, err := s.repo.GetExamAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if attempt.StudentID != userID {
		return nil, ErrPermissionDenied
	}
	alert := &models.ProctoringAlert{
		ExamID:         attempt.ExamID,
		AttemptID:      attempt.ID,
		StudentID:      attempt.StudentID,
		InvigilatorID:  firstInvigilatorID(attempt.Exam.Invigilators),
		Severity:       proctoringSeverityOrDefault(req.Severity),
		EventType:      req.EventType,
		Message:        req.Message,
		IntegrityScore: req.IntegrityScore,
		Evidence:       req.Evidence,
	}
	if alert.IntegrityScore == 0 {
		alert.IntegrityScore = attempt.IntegrityScore
	}
	if err := s.repo.CreateProctoringAlert(ctx, alert); err != nil {
		return nil, err
	}
	response := mapProctoringAlert(alert)
	return &response, nil
}

func (s *ExamService) ListInvigilatorAlerts(ctx context.Context, userID uint, acknowledged *bool) ([]dto.ProctoringAlertResponse, error) {
	if err := s.ensurePermission(ctx, userID, "exam.monitor", scopePtr(rbac.SchoolScope())); err != nil {
		return nil, err
	}
	items, err := s.repo.ListProctoringAlerts(ctx, &userID, acknowledged)
	if err != nil {
		return nil, err
	}
	return mapProctoringAlerts(items), nil
}

func (s *ExamService) AcknowledgeAlert(ctx context.Context, userID, alertID uint) (*dto.ProctoringAlertResponse, error) {
	if err := s.ensurePermission(ctx, userID, "exam.monitor", scopePtr(rbac.SchoolScope())); err != nil {
		return nil, err
	}
	item, err := s.repo.AcknowledgeProctoringAlert(ctx, alertID, userID)
	if err != nil {
		return nil, err
	}
	response := mapProctoringAlert(item)
	return &response, nil
}

func (s *ExamService) ensurePermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) error {
	allowed, err := s.permissionService.UserHasPermission(ctx, userID, permissionCode, target)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func (s *ExamService) ensureExamOfficer(ctx context.Context, userID uint) error {
	ok, err := s.repo.UserHasRole(ctx, userID, "exam_officer")
	if err != nil {
		return err
	}
	if !ok {
		return ErrPermissionDenied
	}
	return nil
}

func (s *ExamService) ensureCourseLecturer(ctx context.Context, userID, courseID uint) error {
	isLecturer, err := s.repo.UserHasRole(ctx, userID, "lecturer")
	if err != nil {
		return err
	}
	if !isLecturer {
		return ErrPermissionDenied
	}
	assigned, err := s.repo.LecturerAssignedToCourse(ctx, userID, courseID)
	if err != nil {
		return err
	}
	if !assigned {
		return ErrPermissionDenied
	}
	return nil
}

func (s *ExamService) resolveCourseID(ctx context.Context, courseID uint, courseCode string) (uint, error) {
	if courseID != 0 {
		return courseID, nil
	}
	code := strings.TrimSpace(courseCode)
	if code == "" {
		return 0, ValidationError{Message: "course_id or course_code is required"}
	}
	course, err := s.repo.GetCourseByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ValidationError{Message: "course not found"}
		}
		return 0, err
	}
	return course.ID, nil
}

func validateExamRequest(courseID uint, title string, startTime, endTime time.Time) error {
	if courseID == 0 {
		return ValidationError{Message: "course_id is required"}
	}
	if strings.TrimSpace(title) == "" {
		return ValidationError{Message: "title is required"}
	}
	if startTime.IsZero() || endTime.IsZero() || !endTime.After(startTime) {
		return ValidationError{Message: "start_time and end_time must be valid"}
	}
	return nil
}

func examStatusOrDefault(raw string) models.ExamStatus {
	status := models.ExamStatus(strings.ToLower(strings.TrimSpace(raw)))
	if status == "" {
		return models.ExamStatusOfficerReview
	}
	if !status.Valid() {
		return models.ExamStatusOfficerReview
	}
	return status
}

func examDeliveryModeOrDefault(raw string) models.ExamDeliveryMode {
	mode := models.ExamDeliveryMode(strings.ToLower(strings.TrimSpace(raw)))
	if mode == "" {
		return models.ExamDeliveryRemoteProctored
	}
	if !mode.Valid() {
		return models.ExamDeliveryRemoteProctored
	}
	return mode
}

func proctoringSeverityOrDefault(raw string) models.ProctoringAlertSeverity {
	severity := models.ProctoringAlertSeverity(strings.ToLower(strings.TrimSpace(raw)))
	if severity == "" {
		return models.ProctoringAlertWarning
	}
	if !severity.Valid() {
		return models.ProctoringAlertWarning
	}
	return severity
}

func nonZero(value, fallback uint) uint {
	if value != 0 {
		return value
	}
	return fallback
}

func encodeJSONBytes(value map[string]any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}

func decodeJSONBytes(value []byte) map[string]any {
	if len(value) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(value, &out); err != nil {
		return nil
	}
	return out
}

func validateQuestionPayload(payload map[string]any) error {
	if payload == nil {
		return nil
	}

	questions := collectQuestionPayloadItems(payload)
	for _, question := range questions {
		qType := normalizeQuestionType(fmt.Sprint(question["type"]))
		switch qType {
		case "objective", "fill_blank", "drag_drop", "essay", "image_question", "file_upload":
		default:
			return ValidationError{Message: "question type must be objective, fill_blank, drag_drop, essay, image_question, or file_upload"}
		}

		if strings.TrimSpace(fmt.Sprint(question["id"])) == "" {
			return ValidationError{Message: "each question must have an id"}
		}
		if strings.TrimSpace(fmt.Sprint(question["prompt"])) == "" {
			return ValidationError{Message: "each question must have a prompt"}
		}
	}

	return nil
}

func collectQuestionPayloadItems(payload map[string]any) []map[string]any {
	out := make([]map[string]any, 0)

	if rawQuestions, ok := payload["questions"]; ok {
		if questions, ok := rawQuestions.([]any); ok {
			out = append(out, questionMapsFromList(questions)...)
		}
	}

	if rawSections, ok := payload["sections"]; ok {
		if sections, ok := rawSections.([]any); ok {
			for _, rawSection := range sections {
				section, ok := rawSection.(map[string]any)
				if !ok {
					continue
				}
				rawQuestions, ok := section["questions"].([]any)
				if !ok {
					continue
				}
				out = append(out, questionMapsFromList(rawQuestions)...)
			}
		}
	}

	return out
}

func questionMapsFromList(items []any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, raw := range items {
		question, ok := raw.(map[string]any)
		if ok {
			out = append(out, question)
		}
	}
	return out
}

func normalizeQuestionType(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")

	switch value {
	case "obj", "objective_question", "multiple_choice", "mcq":
		return "objective"
	case "fill_in_blank", "fill_the_blank", "fillinblank", "fill_blank":
		return "fill_blank"
	case "dragdrop", "drag_and_drop", "drag_drop", "matching":
		return "drag_drop"
	case "theory", "long_answer", "short_answer", "essay":
		return "essay"
	case "image", "picture", "picture_upload", "image_upload", "image_question":
		return "image_question"
	case "file", "file_upload", "practical", "attachment":
		return "file_upload"
	default:
		return value
	}
}

func randomizedQuestionPayload(payload []byte, examID, studentID uint) ([]byte, error) {
	if len(payload) == 0 {
		return nil, nil
	}
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, ValidationError{Message: "question_payload must be valid json"}
	}
	rawQuestions, ok := data["questions"].([]any)
	if !ok || len(rawQuestions) < 2 {
		return payload, nil
	}
	rng := rand.New(rand.NewSource(int64(examID)*1_000_003 + int64(studentID)*97))
	rng.Shuffle(len(rawQuestions), func(i, j int) {
		rawQuestions[i], rawQuestions[j] = rawQuestions[j], rawQuestions[i]
	})
	data["questions"] = rawQuestions
	return json.Marshal(data)
}

func displayName(user models.User) string {
	name := strings.TrimSpace(strings.TrimSpace(user.FirstName) + " " + strings.TrimSpace(user.LastName))
	if name == "" {
		return fmt.Sprintf("Student %d", user.ID)
	}
	return name
}

func minimalPDF(text string) []byte {
	escaped := strings.NewReplacer(`\`, `\\`, `(`, `\(`, `)`, `\)`, "\r", "", "\n", `\n`).Replace(text)
	stream := fmt.Sprintf("BT /F1 10 Tf 50 780 Td (%s) Tj ET", escaped)
	var buf bytes.Buffer
	offsets := []int{}
	writeObj := func(id int, body string) {
		offsets = append(offsets, buf.Len())
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", id, body)
	}
	buf.WriteString("%PDF-1.4\n")
	writeObj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>")
	writeObj(4, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	writeObj(5, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream))
	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n0000000000 65535 f \n", len(offsets)+1)
	for _, offset := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&buf, "trailer << /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(offsets)+1, xref)
	return buf.Bytes()
}

func firstInvigilatorID(items []models.ExamInvigilator) *uint {
	if len(items) == 0 || items[0].InvigilatorID == 0 {
		return nil
	}
	id := items[0].InvigilatorID
	return &id
}

func mapExamVenues(items []models.ExamVenue) []dto.ExamVenueResponse {
	out := make([]dto.ExamVenueResponse, 0, len(items))
	for i := range items {
		out = append(out, mapExamVenue(&items[i]))
	}
	return out
}

func mapExamVenue(item *models.ExamVenue) dto.ExamVenueResponse {
	return dto.ExamVenueResponse{
		ID:        item.ID,
		UUID:      item.UUID,
		Name:      item.Name,
		Address:   item.Address,
		City:      item.City,
		Country:   item.Country,
		Capacity:  item.Capacity,
		IsActive:  item.IsActive,
		CreatedBy: item.CreatedBy,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapExamStudentAllocations(items []models.ExamStudentAllocation) []dto.ExamStudentAllocationResponse {
	out := make([]dto.ExamStudentAllocationResponse, 0, len(items))
	for i := range items {
		out = append(out, dto.ExamStudentAllocationResponse{
			ID:                 items[i].ID,
			UUID:               items[i].UUID,
			ExamID:             items[i].ExamID,
			StudentID:          items[i].StudentID,
			DeliveryMode:       string(items[i].DeliveryMode),
			VenueID:            items[i].VenueID,
			VenueName:          items[i].VenueName,
			VenueAddress:       items[i].VenueAddress,
			IsInternational:    items[i].IsInternational,
			CountryOfResidence: items[i].CountryOfResidence,
			AllocatedBy:        items[i].AllocatedBy,
			AllocatedAt:        items[i].AllocatedAt,
		})
	}
	return out
}

func mapExams(items []models.Exam) []dto.ExamResponse {
	out := make([]dto.ExamResponse, 0, len(items))
	for i := range items {
		out = append(out, mapExam(&items[i]))
	}
	return out
}

func mapExam(item *models.Exam) dto.ExamResponse {
	return dto.ExamResponse{
		ID:              item.ID,
		UUID:            item.UUID,
		CourseID:        item.CourseID,
		CourseCode:      item.Course.Code,
		CourseTitle:     item.Course.Title,
		Title:           item.Title,
		Description:     item.Description,
		Instructions:    item.Instructions,
		Venue:           item.Venue,
		StartTime:       item.StartTime,
		EndTime:         item.EndTime,
		DurationMinutes: item.DurationMinutes,
		DeliveryMode:    string(item.DeliveryMode),
		QuestionPayload: decodeJSONBytes(item.QuestionPayload),
		LecturerID:      item.LecturerID,
		ExamOfficerID:   item.ExamOfficerID,
		ReleasedBy:      item.ReleasedBy,
		ReleasedAt:      item.ReleasedAt,
		Invigilators:    mapExamInvigilators(item.Invigilators),
		Status:          string(item.Status),
		CreatedBy:       item.CreatedBy,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func mapExamInvigilators(items []models.ExamInvigilator) []dto.ExamInvigilatorResponse {
	out := make([]dto.ExamInvigilatorResponse, 0, len(items))
	for i := range items {
		out = append(out, dto.ExamInvigilatorResponse{
			ID:            items[i].ID,
			ExamID:        items[i].ExamID,
			InvigilatorID: items[i].InvigilatorID,
			AssignedBy:    items[i].AssignedBy,
			AssignedAt:    items[i].AssignedAt,
		})
	}
	return out
}

func mapExamAttempt(item *models.ExamAttempt) dto.ExamAttemptResponse {
	return dto.ExamAttemptResponse{
		ID:                     item.ID,
		UUID:                   item.UUID,
		ExamID:                 item.ExamID,
		StudentID:              item.StudentID,
		Status:                 string(item.Status),
		QuestionPayload:        decodeJSONBytes(item.QuestionPayload),
		AnswerPayload:          decodeJSONBytes(item.AnswerPayload),
		IntegrityScore:         item.IntegrityScore,
		EnvironmentConfirmedAt: item.EnvironmentConfirmedAt,
		MonitoringArmedAt:      item.MonitoringArmedAt,
		SubmittedAt:            item.SubmittedAt,
		SubmittedToOfficerAt:   item.SubmittedToOfficerAt,
		SharedWithLecturerAt:   item.SharedWithLecturerAt,
		MarkedBy:               item.MarkedBy,
		Score:                  item.Score,
		LecturerScore:          item.LecturerScore,
		ModeratedScore:         item.ModeratedScore,
		ModerationComment:      item.ModerationComment,
		ModeratedBy:            item.ModeratedBy,
		ModeratedAt:            item.ModeratedAt,
		Feedback:               item.Feedback,
		MarkedAt:               item.MarkedAt,
		TerminationReason:      item.TerminationReason,
		CreatedAt:              item.CreatedAt,
		UpdatedAt:              item.UpdatedAt,
	}
}

func mapExamScriptAnnotations(items []models.ExamScriptAnnotation) []dto.ExamScriptAnnotationResponse {
	out := make([]dto.ExamScriptAnnotationResponse, 0, len(items))
	for i := range items {
		out = append(out, mapExamScriptAnnotation(&items[i]))
	}
	return out
}

func mapExamScriptAnnotation(item *models.ExamScriptAnnotation) dto.ExamScriptAnnotationResponse {
	return dto.ExamScriptAnnotationResponse{
		ID:             item.ID,
		UUID:           item.UUID,
		AttemptID:      item.AttemptID,
		LecturerID:     item.LecturerID,
		QuestionID:     item.QuestionID,
		Comment:        item.Comment,
		HighlightColor: item.HighlightColor,
		InkPayload:     decodeJSONBytes(item.InkPayload),
		Score:          item.Score,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func mapProctoringAlerts(items []models.ProctoringAlert) []dto.ProctoringAlertResponse {
	out := make([]dto.ProctoringAlertResponse, 0, len(items))
	for i := range items {
		out = append(out, mapProctoringAlert(&items[i]))
	}
	return out
}

func mapProctoringAlert(item *models.ProctoringAlert) dto.ProctoringAlertResponse {
	return dto.ProctoringAlertResponse{
		ID:             item.ID,
		UUID:           item.UUID,
		ExamID:         item.ExamID,
		AttemptID:      item.AttemptID,
		StudentID:      item.StudentID,
		InvigilatorID:  item.InvigilatorID,
		EventType:      item.EventType,
		Message:        item.Message,
		Severity:       string(item.Severity),
		IntegrityScore: item.IntegrityScore,
		Evidence:       item.Evidence,
		AcknowledgedBy: item.AcknowledgedBy,
		AcknowledgedAt: item.AcknowledgedAt,
		CreatedAt:      item.CreatedAt,
	}
}

func (s *ExamService) SubmitExamToOfficer(ctx context.Context, userID, examID uint, req dto.ExamWorkflowActionRequest) (*dto.ExamResponse, error) {
	item, err := s.repo.GetExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	if userID != item.LecturerID {
		if err := s.ensurePermission(ctx, userID, "exam.create", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
			return nil, err
		}
	}
	if item.Status == models.ExamStatusCompleted || item.Status == models.ExamStatusCancelled {
		return nil, ValidationError{Message: "this exam can no longer be submitted"}
	}
	item.Status = models.ExamStatusOfficerReview
	if err := appendExamWorkflowNote(item, "submitted_to_exam_officer", req.Comment, userID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateExam(ctx, item); err != nil {
		return nil, err
	}
	response := mapExam(item)
	return &response, nil
}

func (s *ExamService) SendExamToModerator(ctx context.Context, userID, examID uint, req dto.ExamWorkflowActionRequest) (*dto.ExamResponse, error) {
	item, err := s.repo.GetExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.schedule", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if item.Status != models.ExamStatusOfficerReview && item.Status != models.ExamStatusModerated {
		return nil, ValidationError{Message: "exam must be with exam officer before sending to moderator"}
	}
	item.Status = models.ExamStatusModeratorReview
	if err := appendExamWorkflowNote(item, "sent_to_moderator", req.Comment, userID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateExam(ctx, item); err != nil {
		return nil, err
	}
	response := mapExam(item)
	return &response, nil
}

func (s *ExamService) ModeratorReturnExam(ctx context.Context, userID, examID uint, req dto.ExamWorkflowActionRequest) (*dto.ExamResponse, error) {
	item, err := s.repo.GetExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.submit_review", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if item.Status != models.ExamStatusModeratorReview {
		return nil, ValidationError{Message: "exam must be with moderator before review return"}
	}
	item.Status = models.ExamStatusModerated
	if err := appendExamWorkflowNote(item, "moderator_returned", req.Comment, userID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateExam(ctx, item); err != nil {
		return nil, err
	}
	response := mapExam(item)
	return &response, nil
}

func (s *ExamService) SendExamBackToLecturer(ctx context.Context, userID, examID uint, req dto.ExamWorkflowActionRequest) (*dto.ExamResponse, error) {
	item, err := s.repo.GetExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.schedule", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if item.Status != models.ExamStatusModerated && item.Status != models.ExamStatusOfficerReview {
		return nil, ValidationError{Message: "exam must be under review before sending correction to lecturer"}
	}
	item.Status = models.ExamStatusLecturerCorrection
	if err := appendExamWorkflowNote(item, "sent_back_to_lecturer", req.Comment, userID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateExam(ctx, item); err != nil {
		return nil, err
	}
	response := mapExam(item)
	return &response, nil
}

func (s *ExamService) ScheduleExam(ctx context.Context, userID, examID uint, req dto.ExamScheduleRequest) (*dto.ExamResponse, error) {
	item, err := s.repo.GetExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "exam.schedule", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if err := validateExamRequest(item.CourseID, item.Title, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}
	if item.Status != models.ExamStatusModerated && item.Status != models.ExamStatusOfficerReview && item.Status != models.ExamStatusLecturerCorrection {
		return nil, ValidationError{Message: "exam must complete review before scheduling"}
	}
	item.StartTime = req.StartTime
	item.EndTime = req.EndTime
	item.DurationMinutes = req.DurationMinutes
	item.Venue = req.Venue
	item.Status = models.ExamStatusScheduled
	if err := appendExamWorkflowNote(item, "scheduled_by_exam_officer", req.Comment, userID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateExam(ctx, item); err != nil {
		return nil, err
	}
	if req.InvigilatorIDs != nil {
		if _, err := s.repo.ReplaceExamInvigilators(ctx, item.ID, userID, req.InvigilatorIDs); err != nil {
			return nil, err
		}
	}
	item, err = s.repo.GetExam(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	response := mapExam(item)
	return &response, nil
}

func appendExamWorkflowNote(item *models.Exam, action, comment string, userID uint) error {
	data := decodeJSONBytes(item.QuestionPayload)
	if data == nil {
		data = map[string]any{}
	}

	rawNotes, _ := data["workflow_notes"].([]any)
	rawNotes = append(rawNotes, map[string]any{
		"action":  action,
		"comment": strings.TrimSpace(comment),
		"user_id": userID,
		"at":      time.Now().UTC().Format(time.RFC3339),
	})

	data["workflow_notes"] = rawNotes

	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	item.QuestionPayload = encoded
	return nil
}

func (s *ExamService) ListLecturerExamScripts(ctx context.Context, userID uint, lecturerID *uint) ([]dto.LecturerExamScriptResponse, error) {
	targetLecturerID := userID
	if lecturerID != nil && *lecturerID != 0 {
		targetLecturerID = *lecturerID
	}

	items, err := s.repo.ListLecturerExamScripts(ctx, targetLecturerID)
	if err != nil {
		return nil, err
	}

	out := make([]dto.LecturerExamScriptResponse, 0, len(items))
	for _, item := range items {
		questionPayload := decodeJSONBytes(item.QuestionPayload)
		answerPayload := decodeJSONBytes(item.AnswerPayload)
		questions := collectQuestionPayloadItems(questionPayload)
		questionTypes := examQuestionTypes(questions)
		maxScore := examMaxScore(questionPayload, questions)

		out = append(out, dto.LecturerExamScriptResponse{
			AttemptID:            item.AttemptID,
			AttemptUUID:          item.AttemptUUID,
			ExamID:               item.ExamID,
			ExamTitle:            item.ExamTitle,
			CourseID:             item.CourseID,
			CourseCode:           item.CourseCode,
			CourseTitle:          item.CourseTitle,
			StudentID:            item.StudentID,
			StudentName:          cleanName(item.StudentFirstName, item.StudentLastName, fmt.Sprintf("Student %d", item.StudentID)),
			CandidateNo:          nonEmpty(item.MatricNo, fmt.Sprintf("DLC/EXAM/%03d", item.StudentID)),
			Email:                item.Email,
			Status:               string(item.Status),
			QuestionPayload:      questionPayload,
			AnswerPayload:        answerPayload,
			QuestionTypes:        questionTypes,
			QuestionCount:        len(questions),
			ObjectiveScore:       examScorePart(answerPayload, "objective_score", "objective", "obj"),
			TheoryScore:          examScorePart(answerPayload, "theory_score", "essay_score", "theory", "essay"),
			PracticalScore:       examScorePart(answerPayload, "practical_score", "file_upload_score", "image_score", "practical"),
			Score:                item.Score,
			LecturerScore:        item.LecturerScore,
			ModeratedScore:       item.ModeratedScore,
			MaxScore:             maxScore,
			IntegrityScore:       item.IntegrityScore,
			Feedback:             item.Feedback,
			TerminationReason:    item.TerminationReason,
			SubmittedAt:          item.SubmittedAt,
			SharedWithLecturerAt: item.SharedWithLecturerAt,
		})
	}

	return out, nil
}

func examQuestionTypes(questions []map[string]any) []string {
	seen := map[string]bool{}
	out := make([]string, 0)

	for _, question := range questions {
		qType := normalizeQuestionType(fmt.Sprint(question["type"]))
		if qType == "" || seen[qType] {
			continue
		}
		seen[qType] = true
		out = append(out, qType)
	}

	return out
}

func examMaxScore(payload map[string]any, questions []map[string]any) float64 {
	total := 0.0

	for _, question := range questions {
		mark := examNumber(question["marks"])
		if mark == 0 {
			mark = examNumber(question["max_marks"])
		}
		total += mark
	}

	if total == 0 {
		total = examNumber(payload["total_marks"])
	}
	if total == 0 {
		total = examNumber(payload["max_score"])
	}

	return total
}

func examScorePart(payload map[string]any, keys ...string) float64 {
	for _, key := range keys {
		if value := examNumber(payload[key]); value != 0 {
			return value
		}
	}

	rawScores, ok := payload["scores"].(map[string]any)
	if ok {
		for _, key := range keys {
			if value := examNumber(rawScores[key]); value != 0 {
				return value
			}
		}
	}

	return 0
}

func examNumber(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	case json.Number:
		out, _ := v.Float64()
		return out
	case string:
		var out float64
		_, _ = fmt.Sscan(strings.TrimSpace(v), &out)
		return out
	default:
		return 0
	}
}

func cleanName(firstName, lastName, fallback string) string {
	name := strings.TrimSpace(strings.TrimSpace(firstName) + " " + strings.TrimSpace(lastName))
	if name == "" {
		return fallback
	}
	return name
}

func nonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
