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

type ResultService struct {
	repo              *repository.TeachingRepository
	permissionService *PermissionService
}

func NewResultService(repo *repository.TeachingRepository, permissionService *PermissionService) *ResultService {
	return &ResultService{repo: repo, permissionService: permissionService}
}

func (s *ResultService) ListResults(ctx context.Context, userID uint, filter repository.ResultListFilter) ([]dto.ResultResponse, error) {
	target := scopePtr(rbac.SchoolScope())
	if filter.CourseID != nil {
		target = scopePtr(rbac.CourseScope(*filter.CourseID))
	}
	if err := s.ensurePermission(ctx, userID, "result.view", target); err != nil {
		return nil, err
	}

	items, err := s.repo.ListResults(ctx, filter)
	if err != nil {
		return nil, err
	}
	return mapResults(items), nil
}

func (s *ResultService) GetResult(ctx context.Context, userID, resultID uint) (*dto.ResultResponse, error) {
	item, err := s.repo.GetResult(ctx, resultID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "result.view", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	response := mapResult(item)
	return &response, nil
}

func (s *ResultService) CreateResult(ctx context.Context, userID uint, req dto.ResultCreateRequest) (*dto.ResultResponse, error) {
	courseID, err := s.resolveCourseID(ctx, req.CourseID, req.CourseCode)
	if err != nil {
		return nil, err
	}
	if err := validateResultRequest(courseID, req.StudentID, req.AssessmentType); err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "result.mark", scopePtr(rbac.CourseScope(courseID))); err != nil {
		return nil, err
	}
	if err := s.ensureMarkingOwner(ctx, userID, courseID, req.AssessmentType); err != nil {
		return nil, err
	}

	status := resultStatusOrDefault(req.Status)
	item := &models.Result{
		CourseID:              courseID,
		StudentID:             req.StudentID,
		AssessmentType:        models.ResultReferenceType(strings.ToLower(strings.TrimSpace(req.AssessmentType))),
		ReferenceID:           req.ReferenceID,
		Score:                 req.Score,
		GradedAssessmentScore: req.GradedAssessmentScore,
		AssignmentScore:       req.AssignmentScore,
		GroupAssignmentScore:  req.GroupAssignmentScore,
		PeerReviewScore:       req.PeerReviewScore,
		ExaminationScore:      req.ExaminationScore,
		TotalScore:            req.TotalScore,
		Grade:                 req.Grade,
		Remark:                req.Remark,
		Status:                status,
		MarkedBy:              userID,
		PublishedAt:           req.PublishedAt,
	}
	if item.PublishedAt != nil {
		item.Status = models.ResultStatusPublished
	}
	if err := s.repo.CreateResult(ctx, item); err != nil {
		return nil, err
	}
	item, err = s.repo.GetResult(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	response := mapResult(item)
	return &response, nil
}

func (s *ResultService) UpdateResult(ctx context.Context, userID, resultID uint, req dto.ResultUpdateRequest) (*dto.ResultResponse, error) {
	item, err := s.repo.GetResult(ctx, resultID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "result.mark", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if err := s.ensureMarkingOwner(ctx, userID, item.CourseID, string(item.AssessmentType)); err != nil {
		return nil, err
	}
	if item.Status == models.ResultStatusPublished {
		return nil, ValidationError{Message: "published results cannot be edited"}
	}

	item.Score = req.Score
	item.GradedAssessmentScore = req.GradedAssessmentScore
	item.AssignmentScore = req.AssignmentScore
	item.GroupAssignmentScore = req.GroupAssignmentScore
	item.PeerReviewScore = req.PeerReviewScore
	item.ExaminationScore = req.ExaminationScore
	item.TotalScore = req.TotalScore
	item.Grade = req.Grade
	item.Remark = req.Remark
	if strings.TrimSpace(req.Status) != "" {
		item.Status = resultStatusOrDefault(req.Status)
	}
	item.PublishedAt = req.PublishedAt
	if item.PublishedAt != nil {
		item.Status = models.ResultStatusPublished
	}
	if err := s.repo.UpdateResult(ctx, item); err != nil {
		return nil, err
	}
	response := mapResult(item)
	return &response, nil
}

func (s *ResultService) ApproveResult(ctx context.Context, userID, resultID uint) (*dto.ResultResponse, error) {
	item, err := s.repo.GetResult(ctx, resultID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "result.approve", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	item.Status = models.ResultStatusApproved
	item.ApprovedBy = &userID
	item.ApprovedAt = &now
	if err := s.repo.UpdateResult(ctx, item); err != nil {
		return nil, err
	}
	response := mapResult(item)
	return &response, nil
}

func (s *ResultService) PublishResult(ctx context.Context, userID, resultID uint) (*dto.ResultResponse, error) {
	item, err := s.repo.GetResult(ctx, resultID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "result.publish", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if item.ApprovedBy == nil {
		item.ApprovedBy = &userID
		item.ApprovedAt = &now
	}
	item.Status = models.ResultStatusPublished
	item.PublishedAt = &now
	if err := s.repo.UpdateResult(ctx, item); err != nil {
		return nil, err
	}
	response := mapResult(item)
	return &response, nil
}

func (s *ResultService) resolveCourseID(ctx context.Context, courseID uint, courseCode string) (uint, error) {
	if courseID != 0 {
		return courseID, nil
	}
	if strings.TrimSpace(courseCode) == "" {
		return 0, ValidationError{Message: "course_id or course_code is required"}
	}
	course, err := s.repo.GetCourseByCode(ctx, strings.TrimSpace(courseCode))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ValidationError{Message: "course not found"}
		}
		return 0, err
	}
	return course.ID, nil
}

func (s *ResultService) ensurePermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) error {
	allowed, err := s.permissionService.UserHasPermission(ctx, userID, permissionCode, target)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func (s *ResultService) ensureMarkingOwner(ctx context.Context, userID, courseID uint, assessmentType string) error {
	switch models.ResultReferenceType(strings.ToLower(strings.TrimSpace(assessmentType))) {
	case models.ResultReferenceExam:
		ok, err := s.repo.UserHasRole(ctx, userID, "exam_officer")
		if err != nil {
			return err
		}
		if !ok {
			return ErrPermissionDenied
		}
	case models.ResultReferenceAssignment, models.ResultReferenceQuiz:
		ok, err := s.repo.UserHasRole(ctx, userID, "lecturer")
		if err != nil {
			return err
		}
		if !ok {
			return ErrPermissionDenied
		}
		assigned, err := s.repo.LecturerAssignedToCourse(ctx, userID, courseID)
		if err != nil {
			return err
		}
		if !assigned {
			return ErrPermissionDenied
		}
	}
	return nil
}

func validateResultRequest(courseID, studentID uint, assessmentType string) error {
	if courseID == 0 {
		return ValidationError{Message: "course_id is required"}
	}
	if studentID == 0 {
		return ValidationError{Message: "student_id is required"}
	}
	ref := models.ResultReferenceType(strings.ToLower(strings.TrimSpace(assessmentType)))
	if !ref.Valid() {
		return ValidationError{Message: "invalid assessment_type"}
	}
	return nil
}

func resultStatusOrDefault(raw string) models.ResultStatus {
	status := models.ResultStatus(strings.ToLower(strings.TrimSpace(raw)))
	if status == "" {
		return models.ResultStatusSubmitted
	}
	if !status.Valid() {
		return models.ResultStatusSubmitted
	}
	return status
}

func mapResults(items []models.Result) []dto.ResultResponse {
	out := make([]dto.ResultResponse, 0, len(items))
	for i := range items {
		out = append(out, mapResult(&items[i]))
	}
	return out
}

func mapResult(item *models.Result) dto.ResultResponse {
	return dto.ResultResponse{
		ID:                    item.ID,
		UUID:                  item.UUID,
		CourseID:              item.CourseID,
		CourseCode:            item.Course.Code,
		CourseTitle:           item.Course.Title,
		StudentID:             item.StudentID,
		AssessmentType:        string(item.AssessmentType),
		ReferenceID:           item.ReferenceID,
		Score:                 item.Score,
		GradedAssessmentScore: item.GradedAssessmentScore,
		AssignmentScore:       item.AssignmentScore,
		GroupAssignmentScore:  item.GroupAssignmentScore,
		PeerReviewScore:       item.PeerReviewScore,
		ExaminationScore:      item.ExaminationScore,
		TotalScore:            item.TotalScore,
		Grade:                 item.Grade,
		Remark:                item.Remark,
		Status:                string(item.Status),
		MarkedBy:              item.MarkedBy,
		ApprovedBy:            item.ApprovedBy,
		ApprovedAt:            item.ApprovedAt,
		PublishedAt:           item.PublishedAt,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
	}
}
