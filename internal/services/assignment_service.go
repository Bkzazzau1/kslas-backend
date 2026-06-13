package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
	"kslasbackend/internal/dto"
	"kslasbackend/internal/rbac"
	"kslasbackend/internal/repository"
)

type AssignmentService struct {
	repo              *repository.TeachingRepository
	permissionService *PermissionService
}

func NewAssignmentService(repo *repository.TeachingRepository, permissionService *PermissionService) *AssignmentService {
	return &AssignmentService{repo: repo, permissionService: permissionService}
}

func (s *AssignmentService) ListAssignments(ctx context.Context, userID uint, filter repository.AssignmentListFilter) ([]dto.AssignmentResponse, error) {
	target := scopePtr(rbac.SchoolScope())
	if filter.CourseID != nil {
		target = scopePtr(rbac.CourseScope(*filter.CourseID))
	}
	if err := s.ensurePermission(ctx, userID, "course.view", target); err != nil {
		return nil, err
	}

	items, err := s.repo.ListAssignments(ctx, filter)
	if err != nil {
		return nil, err
	}
	return s.mapAssignmentsForUser(ctx, userID, items), nil
}

func (s *AssignmentService) GetAssignment(ctx context.Context, userID, assignmentID uint) (*dto.AssignmentResponse, error) {
	item, err := s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "course.view", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}

	response := s.mapAssignmentForUser(ctx, userID, item)
	return &response, nil
}

func (s *AssignmentService) CreateAssignment(ctx context.Context, userID uint, req dto.AssignmentCreateRequest) (*dto.AssignmentResponse, error) {
	courseID := req.CourseID
	if courseID == 0 && strings.TrimSpace(req.CourseCode) != "" {
		course, err := s.repo.GetCourseByCode(ctx, strings.TrimSpace(req.CourseCode))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ValidationError{Message: "course not found"}
			}
			return nil, err
		}
		courseID = course.ID
	}

	if err := validateAssignmentRequest(courseID, req.Title, req.MaxScore); err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "assignment.create", scopePtr(rbac.CourseScope(courseID))); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetCourse(ctx, courseID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ValidationError{Message: "course not found"}
		}
		return nil, err
	}

	item := &models.Assignment{
		CourseID:           courseID,
		Title:              req.Title,
		Description:        req.Description,
		Instructions:       req.Instructions,
		DueAt:              req.DueAt,
		MaxScore:           req.MaxScore,
		AssignmentType:     assignmentTypeOrDefault(req.AssignmentType),
		SubmissionMode:     submissionModeOrDefault(req.SubmissionMode),
		AllowedExtensions:  cleanStrings(req.AllowedExtensions),
		WhiteboardEnabled:  req.WhiteboardEnabled,
		WhiteboardRequired: req.WhiteboardRequired,
		WhiteboardPrompt:   req.WhiteboardPrompt,
		PeerReviewEnabled:  req.PeerReviewEnabled,
		PeerReviewRubric:   cleanStrings(req.PeerReviewRubric),
		GroupSource:        assignmentGroupSource(req.GroupSource),
		Status:             assignmentStatusOrDefault(req.Status),
		CreatedBy:          userID,
	}
	if err := s.repo.CreateAssignment(ctx, item); err != nil {
		return nil, err
	}

	if len(req.Groups) > 0 {
		if _, err := s.repo.ReplaceAssignmentGroups(ctx, item.ID, userID, groupsFromRequest(req.Groups)); err != nil {
			return nil, err
		}
	}
	if len(req.PeerReviews) > 0 {
		if _, err := s.repo.ReplaceAssignmentPeerReviews(ctx, item.ID, userID, peerReviewsFromRequest(req.PeerReviews)); err != nil {
			return nil, err
		}
	}

	item, err := s.repo.GetAssignment(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	response := mapAssignment(item)
	return &response, nil
}

func (s *AssignmentService) UpdateAssignment(ctx context.Context, userID, assignmentID uint, req dto.AssignmentUpdateRequest) (*dto.AssignmentResponse, error) {
	item, err := s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "assignment.create", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if err := validateAssignmentRequest(item.CourseID, req.Title, req.MaxScore); err != nil {
		return nil, err
	}

	item.Title = req.Title
	item.Description = req.Description
	item.Instructions = req.Instructions
	item.DueAt = req.DueAt
	item.MaxScore = req.MaxScore
	item.AssignmentType = assignmentTypeOrDefault(req.AssignmentType)
	item.SubmissionMode = submissionModeOrDefault(req.SubmissionMode)
	item.AllowedExtensions = cleanStrings(req.AllowedExtensions)
	item.WhiteboardEnabled = req.WhiteboardEnabled
	item.WhiteboardRequired = req.WhiteboardRequired
	item.WhiteboardPrompt = req.WhiteboardPrompt
	item.PeerReviewEnabled = req.PeerReviewEnabled
	item.PeerReviewRubric = cleanStrings(req.PeerReviewRubric)
	item.GroupSource = assignmentGroupSource(req.GroupSource)
	item.Status = assignmentStatusOrDefault(req.Status)

	if err := s.repo.UpdateAssignment(ctx, item); err != nil {
		return nil, err
	}

	item, err = s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	response := mapAssignment(item)
	return &response, nil
}

func (s *AssignmentService) DeleteAssignment(ctx context.Context, userID, assignmentID uint) error {
	item, err := s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return err
	}
	if err := s.ensurePermission(ctx, userID, "assignment.create", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return err
	}
	return s.repo.DeleteAssignment(ctx, assignmentID)
}

func (s *AssignmentService) SubmitAssignment(ctx context.Context, userID, assignmentID uint, req dto.AssignmentSubmissionCreateRequest) (*dto.AssignmentSubmissionResponse, error) {
	assignment, err := s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "assignment.submit", scopePtr(rbac.CourseScope(assignment.CourseID))); err != nil {
		return nil, err
	}
	if assignment.Status != models.AssignmentStatusPublished {
		return nil, ValidationError{Message: "assignment is not open for submission"}
	}
	if assignment.DueAt != nil && time.Now().UTC().After(*assignment.DueAt) {
		return nil, ValidationError{Message: "assignment deadline has passed"}
	}
	if strings.TrimSpace(req.TextAnswer) == "" && strings.TrimSpace(req.WhiteboardData) == "" && len(req.Files) == 0 {
		return nil, ValidationError{Message: "submission must include text, whiteboard data, or files"}
	}

	submission := &models.AssignmentSubmission{
		AssignmentID:   assignmentID,
		StudentID:      userID,
		GroupID:        req.GroupID,
		TextAnswer:     req.TextAnswer,
		WhiteboardData: []byte(req.WhiteboardData),
		Status:         models.AssignmentSubmissionStatusSubmitted,
		SubmittedAt:    time.Now().UTC(),
		Files:          submissionFilesFromRequest(req.Files),
	}
	item, err := s.repo.UpsertAssignmentSubmission(ctx, submission)
	if err != nil {
		return nil, err
	}
	response := mapAssignmentSubmission(item)
	return &response, nil
}

func (s *AssignmentService) ListSubmissions(ctx context.Context, userID, assignmentID uint) ([]dto.AssignmentSubmissionResponse, error) {
	assignment, err := s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "assignment.mark", scopePtr(rbac.CourseScope(assignment.CourseID))); err != nil {
		return nil, err
	}
	if err := s.ensureCourseLecturer(ctx, userID, assignment.CourseID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListAssignmentSubmissions(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	return mapAssignmentSubmissions(items), nil
}

func (s *AssignmentService) SubmitPeerReview(ctx context.Context, userID, assignmentID, reviewID uint, req dto.AssignmentPeerReviewSubmitRequest) (*dto.AssignmentPeerReviewResponse, error) {
	assignment, err := s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "assignment.submit", scopePtr(rbac.CourseScope(assignment.CourseID))); err != nil {
		return nil, err
	}

	review, err := s.repo.GetAssignmentPeerReview(ctx, assignmentID, reviewID)
	if err != nil {
		return nil, err
	}
	if review.ReviewerID != userID {
		return nil, ErrPermissionDenied
	}
	if req.Score < 0 || req.Score > assignment.MaxScore {
		return nil, ValidationError{Message: "peer review score is outside assignment score range"}
	}

	now := time.Now().UTC()
	review.Score = req.Score
	review.Feedback = req.Feedback
	review.RubricChecks = req.RubricChecks
	review.SubmittedAt = &now
	if err := s.repo.UpdateAssignmentPeerReview(ctx, review); err != nil {
		return nil, err
	}

	response := mapAssignmentPeerReviewForStudent(review, userID)
	return &response, nil
}

func (s *AssignmentService) ListGrades(ctx context.Context, userID, assignmentID uint) ([]dto.AssignmentGradeResponse, error) {
	assignment, err := s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "assignment.mark", scopePtr(rbac.CourseScope(assignment.CourseID))); err != nil {
		return nil, err
	}
	items, err := s.repo.ListAssignmentGrades(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	return mapAssignmentGrades(items), nil
}

func (s *AssignmentService) UpsertGrades(ctx context.Context, userID, assignmentID uint, req dto.AssignmentGradeBulkUpsertRequest) ([]dto.AssignmentGradeResponse, error) {
	assignment, err := s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "assignment.mark", scopePtr(rbac.CourseScope(assignment.CourseID))); err != nil {
		return nil, err
	}

	items := make([]models.AssignmentGrade, 0, len(req.Items))
	now := time.Now().UTC()
	for _, input := range req.Items {
		if input.StudentID == 0 {
			return nil, ValidationError{Message: "student_id is required"}
		}
		if input.Score < 0 || input.Score > assignment.MaxScore {
			return nil, ValidationError{Message: "grade score is outside assignment score range"}
		}
		status := strings.ToLower(strings.TrimSpace(input.Status))
		if status == "" {
			status = "marked"
		}
		items = append(items, models.AssignmentGrade{
			StudentID: input.StudentID,
			Score:     input.Score,
			Feedback:  input.Feedback,
			Status:    status,
			MarkedAt:  &now,
		})
	}

	grades, err := s.repo.UpsertAssignmentGrades(ctx, assignmentID, userID, items)
	if err != nil {
		return nil, err
	}
	return mapAssignmentGrades(grades), nil
}

func (s *AssignmentService) ensurePermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) error {
	allowed, err := s.permissionService.UserHasPermission(ctx, userID, permissionCode, target)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func (s *AssignmentService) ensureCourseLecturer(ctx context.Context, userID, courseID uint) error {
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

func validateAssignmentRequest(courseID uint, title string, maxScore float64) error {
	if courseID == 0 {
		return ValidationError{Message: "course_id is required"}
	}
	if strings.TrimSpace(title) == "" {
		return ValidationError{Message: "title is required"}
	}
	if maxScore <= 0 {
		return ValidationError{Message: "max_score must be greater than zero"}
	}
	return nil
}

func assignmentTypeOrDefault(value string) models.AssignmentType {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return models.AssignmentTypeIndividual
	}
	return models.AssignmentType(value)
}

func submissionModeOrDefault(value string) models.AssignmentSubmissionMode {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return models.AssignmentSubmissionModeMixed
	}
	return models.AssignmentSubmissionMode(value)
}

func assignmentStatusOrDefault(value string) models.AssignmentStatus {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return models.AssignmentStatusDraft
	}
	return models.AssignmentStatus(value)
}

func assignmentGroupSource(value string) models.AssignmentGroupSource {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	return models.AssignmentGroupSource(value)
}

func cleanStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func groupsFromRequest(inputs []dto.AssignmentGroupCreateInput) []models.AssignmentGroup {
	out := make([]models.AssignmentGroup, 0, len(inputs))
	for _, input := range inputs {
		group := models.AssignmentGroup{
			Name:   input.Name,
			Source: assignmentGroupSource(input.Source),
		}
		if group.Source == "" {
			group.Source = models.AssignmentGroupSourceLecturerSetBackend
		}
		for _, studentID := range input.StudentIDs {
			if studentID == 0 {
				continue
			}
			group.Members = append(group.Members, models.AssignmentGroupMember{
				StudentID: studentID,
			})
		}
		out = append(out, group)
	}
	return out
}

func peerReviewsFromRequest(inputs []dto.AssignmentPeerReviewAssignInput) []models.AssignmentPeerReview {
	out := make([]models.AssignmentPeerReview, 0, len(inputs))
	for _, input := range inputs {
		if input.ReviewerID == 0 || input.TargetStudentID == 0 {
			continue
		}
		out = append(out, models.AssignmentPeerReview{
			ReviewerID:         input.ReviewerID,
			TargetStudentID:    input.TargetStudentID,
			TargetSubmissionID: input.TargetSubmissionID,
			RubricChecks:       map[string]bool{},
		})
	}
	return out
}

func submissionFilesFromRequest(inputs []dto.AssignmentSubmissionFileInput) []models.AssignmentSubmissionFile {
	out := make([]models.AssignmentSubmissionFile, 0, len(inputs))
	for _, input := range inputs {
		out = append(out, models.AssignmentSubmissionFile{
			StoragePath:      input.StoragePath,
			OriginalFileName: input.OriginalFileName,
			StoredFileName:   input.StoredFileName,
			MimeType:         input.MimeType,
			SizeBytes:        input.SizeBytes,
		})
	}
	return out
}

func mapAssignments(items []models.Assignment) []dto.AssignmentResponse {
	out := make([]dto.AssignmentResponse, 0, len(items))
	for i := range items {
		out = append(out, mapAssignment(&items[i]))
	}
	return out
}

func (s *AssignmentService) mapAssignmentsForUser(ctx context.Context, userID uint, items []models.Assignment) []dto.AssignmentResponse {
	out := make([]dto.AssignmentResponse, 0, len(items))
	for i := range items {
		out = append(out, s.mapAssignmentForUser(ctx, userID, &items[i]))
	}
	return out
}

func (s *AssignmentService) mapAssignmentForUser(ctx context.Context, userID uint, item *models.Assignment) dto.AssignmentResponse {
	response := mapAssignment(item)
	isStudent, err := s.repo.UserHasRole(ctx, userID, "student")
	if err != nil || !isStudent {
		return response
	}
	response.PeerReviews = mapAssignmentPeerReviewsForStudent(item.PeerReviews, userID)
	return response
}

func mapAssignment(item *models.Assignment) dto.AssignmentResponse {
	return dto.AssignmentResponse{
		ID:                 item.ID,
		UUID:               item.UUID,
		CourseID:           item.CourseID,
		CourseCode:         item.Course.Code,
		CourseTitle:        item.Course.Title,
		Title:              item.Title,
		Description:        item.Description,
		Instructions:       item.Instructions,
		DueAt:              item.DueAt,
		MaxScore:           item.MaxScore,
		AssignmentType:     string(item.AssignmentType),
		SubmissionMode:     string(item.SubmissionMode),
		AllowedExtensions:  item.AllowedExtensions,
		WhiteboardEnabled:  item.WhiteboardEnabled,
		WhiteboardRequired: item.WhiteboardRequired,
		WhiteboardPrompt:   item.WhiteboardPrompt,
		PeerReviewEnabled:  item.PeerReviewEnabled,
		PeerReviewRubric:   item.PeerReviewRubric,
		GroupSource:        string(item.GroupSource),
		Groups:             mapAssignmentGroups(item.Groups),
		PeerReviews:        mapAssignmentPeerReviews(item.PeerReviews),
		SubmissionCount:    len(item.Submissions),
		GradeCount:         len(item.Grades),
		Status:             string(item.Status),
		CreatedBy:          item.CreatedBy,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}

func mapAssignmentGroups(items []models.AssignmentGroup) []dto.AssignmentGroupResponse {
	out := make([]dto.AssignmentGroupResponse, 0, len(items))
	for i := range items {
		out = append(out, dto.AssignmentGroupResponse{
			ID:           items[i].ID,
			UUID:         items[i].UUID,
			AssignmentID: items[i].AssignmentID,
			Name:         items[i].Name,
			Source:       string(items[i].Source),
			Members:      mapAssignmentGroupMembers(items[i].Members),
			CreatedBy:    items[i].CreatedBy,
			CreatedAt:    items[i].CreatedAt,
			UpdatedAt:    items[i].UpdatedAt,
		})
	}
	return out
}

func mapAssignmentGroupMembers(items []models.AssignmentGroupMember) []dto.AssignmentGroupMemberResponse {
	out := make([]dto.AssignmentGroupMemberResponse, 0, len(items))
	for i := range items {
		out = append(out, dto.AssignmentGroupMemberResponse{
			ID:        items[i].ID,
			GroupID:   items[i].GroupID,
			StudentID: items[i].StudentID,
			AddedBy:   items[i].AddedBy,
			CreatedAt: items[i].CreatedAt,
		})
	}
	return out
}

func mapAssignmentSubmissions(items []models.AssignmentSubmission) []dto.AssignmentSubmissionResponse {
	out := make([]dto.AssignmentSubmissionResponse, 0, len(items))
	for i := range items {
		out = append(out, mapAssignmentSubmission(&items[i]))
	}
	return out
}

func mapAssignmentSubmission(item *models.AssignmentSubmission) dto.AssignmentSubmissionResponse {
	return dto.AssignmentSubmissionResponse{
		ID:             item.ID,
		UUID:           item.UUID,
		AssignmentID:   item.AssignmentID,
		StudentID:      item.StudentID,
		GroupID:        item.GroupID,
		TextAnswer:     item.TextAnswer,
		WhiteboardData: string(item.WhiteboardData),
		Status:         string(item.Status),
		SubmittedAt:    item.SubmittedAt,
		Files:          mapAssignmentSubmissionFiles(item.Files),
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func mapAssignmentSubmissionFiles(items []models.AssignmentSubmissionFile) []dto.AssignmentSubmissionFileResponse {
	out := make([]dto.AssignmentSubmissionFileResponse, 0, len(items))
	for i := range items {
		out = append(out, dto.AssignmentSubmissionFileResponse{
			ID:               items[i].ID,
			SubmissionID:     items[i].SubmissionID,
			StoragePath:      items[i].StoragePath,
			OriginalFileName: items[i].OriginalFileName,
			StoredFileName:   items[i].StoredFileName,
			MimeType:         items[i].MimeType,
			SizeBytes:        items[i].SizeBytes,
			CreatedAt:        items[i].CreatedAt,
		})
	}
	return out
}

func mapAssignmentPeerReviews(items []models.AssignmentPeerReview) []dto.AssignmentPeerReviewResponse {
	out := make([]dto.AssignmentPeerReviewResponse, 0, len(items))
	for i := range items {
		out = append(out, mapAssignmentPeerReview(&items[i]))
	}
	return out
}

func mapAssignmentPeerReviewsForStudent(items []models.AssignmentPeerReview, reviewerID uint) []dto.AssignmentPeerReviewResponse {
	out := make([]dto.AssignmentPeerReviewResponse, 0, len(items))
	for i := range items {
		if items[i].ReviewerID == reviewerID {
			out = append(out, mapAssignmentPeerReviewForStudent(&items[i], reviewerID))
		}
	}
	return out
}

func mapAssignmentPeerReview(item *models.AssignmentPeerReview) dto.AssignmentPeerReviewResponse {
	return dto.AssignmentPeerReviewResponse{
		ID:                 item.ID,
		UUID:               item.UUID,
		AssignmentID:       item.AssignmentID,
		ReviewerID:         item.ReviewerID,
		TargetStudentID:    item.TargetStudentID,
		TargetSubmissionID: item.TargetSubmissionID,
		Score:              item.Score,
		Feedback:           item.Feedback,
		RubricChecks:       item.RubricChecks,
		AssignedBy:         item.AssignedBy,
		SubmittedAt:        item.SubmittedAt,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}

func mapAssignmentPeerReviewForStudent(item *models.AssignmentPeerReview, reviewerID uint) dto.AssignmentPeerReviewResponse {
	response := mapAssignmentPeerReview(item)
	if item.ReviewerID == reviewerID {
		response.ReviewerID = 0
		response.TargetStudentID = 0
	}
	return response
}

func mapAssignmentGrades(items []models.AssignmentGrade) []dto.AssignmentGradeResponse {
	out := make([]dto.AssignmentGradeResponse, 0, len(items))
	for i := range items {
		out = append(out, dto.AssignmentGradeResponse{
			ID:           items[i].ID,
			AssignmentID: items[i].AssignmentID,
			StudentID:    items[i].StudentID,
			MarkerID:     items[i].MarkerID,
			Score:        items[i].Score,
			Feedback:     items[i].Feedback,
			Status:       items[i].Status,
			MarkedAt:     items[i].MarkedAt,
			CreatedAt:    items[i].CreatedAt,
			UpdatedAt:    items[i].UpdatedAt,
		})
	}
	return out
}
