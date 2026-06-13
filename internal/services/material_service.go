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

type MaterialService struct {
	repo              *repository.TeachingRepository
	permissionService *PermissionService
}

func NewMaterialService(repo *repository.TeachingRepository, permissionService *PermissionService) *MaterialService {
	return &MaterialService{repo: repo, permissionService: permissionService}
}

func (s *MaterialService) ListMaterials(ctx context.Context, userID uint, filter repository.MaterialListFilter) ([]dto.CourseMaterialResponse, error) {
	if filter.CourseID == nil && strings.TrimSpace(filter.CourseCode) != "" {
		course, err := s.repo.GetCourseByCode(ctx, strings.TrimSpace(filter.CourseCode))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ValidationError{Message: "course not found"}
			}
			return nil, err
		}
		filter.CourseID = &course.ID
	}

	target := scopePtr(rbac.SchoolScope())
	if filter.CourseID != nil {
		target = scopePtr(rbac.CourseScope(*filter.CourseID))
	}
	if err := s.ensurePermission(ctx, userID, "material.view", target); err != nil {
		return nil, err
	}
	if err := s.ensureStudentCourseRegistration(ctx, userID, filter.CourseID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListMaterials(ctx, filter)
	if err != nil {
		return nil, err
	}
	return mapCourseMaterials(items), nil
}

func (s *MaterialService) CreateMaterial(ctx context.Context, userID uint, req dto.CourseMaterialCreateRequest) (*dto.CourseMaterialResponse, error) {
	courseID, err := s.resolveCourseID(ctx, req.CourseID, req.CourseCode)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "material.upload", scopePtr(rbac.CourseScope(courseID))); err != nil {
		return nil, err
	}
	if err := s.ensureCourseLecturer(ctx, userID, courseID); err != nil {
		return nil, err
	}
	if err := validateMaterialRequest(req.Title, req.MaterialType, req.ExternalURL); err != nil {
		return nil, err
	}
	allowDownload := true
	if req.AllowDownload != nil {
		allowDownload = *req.AllowDownload
	}
	var publishedAt *time.Time
	if req.Publish == nil || *req.Publish {
		now := time.Now().UTC()
		publishedAt = &now
	}
	item := &models.CourseMaterial{
		CourseID:      courseID,
		Title:         req.Title,
		Description:   req.Description,
		MaterialType:  materialTypeOrDefault(req.MaterialType),
		ExternalURL:   strings.TrimSpace(req.ExternalURL),
		AllowDownload: allowDownload,
		UploadedBy:    userID,
		PublishedAt:   publishedAt,
	}
	if err := s.repo.CreateMaterial(ctx, item); err != nil {
		return nil, err
	}
	item, err = s.repo.GetMaterial(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	response := mapCourseMaterial(item)
	return &response, nil
}

func (s *MaterialService) GetMaterial(ctx context.Context, userID, materialID uint) (*dto.CourseMaterialResponse, error) {
	item, err := s.repo.GetMaterial(ctx, materialID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "material.view", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if err := s.ensureStudentCourseRegistration(ctx, userID, &item.CourseID); err != nil {
		return nil, err
	}
	response := mapCourseMaterial(item)
	return &response, nil
}

func (s *MaterialService) UpdateMaterial(ctx context.Context, userID, materialID uint, req dto.CourseMaterialUpdateRequest) (*dto.CourseMaterialResponse, error) {
	item, err := s.repo.GetMaterial(ctx, materialID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "course.manage_content", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if err := s.ensureCourseLecturer(ctx, userID, item.CourseID); err != nil {
		return nil, err
	}
	if err := validateMaterialRequest(req.Title, req.MaterialType, req.ExternalURL); err != nil {
		return nil, err
	}
	item.Title = req.Title
	item.Description = req.Description
	item.MaterialType = materialTypeOrDefault(req.MaterialType)
	item.ExternalURL = strings.TrimSpace(req.ExternalURL)
	if req.AllowDownload != nil {
		item.AllowDownload = *req.AllowDownload
	}
	if req.Publish != nil {
		if *req.Publish {
			now := time.Now().UTC()
			item.PublishedAt = &now
		} else {
			item.PublishedAt = nil
		}
	}
	if err := s.repo.UpdateMaterial(ctx, item); err != nil {
		return nil, err
	}
	response := mapCourseMaterial(item)
	return &response, nil
}

func (s *MaterialService) PublishMaterial(ctx context.Context, userID, materialID uint) (*dto.CourseMaterialResponse, error) {
	item, err := s.repo.GetMaterial(ctx, materialID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "course.manage_content", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if err := s.ensureCourseLecturer(ctx, userID, item.CourseID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	item.PublishedAt = &now
	if err := s.repo.UpdateMaterial(ctx, item); err != nil {
		return nil, err
	}
	response := mapCourseMaterial(item)
	return &response, nil
}

func (s *MaterialService) DeleteMaterial(ctx context.Context, userID, materialID uint) error {
	item, err := s.repo.GetMaterial(ctx, materialID)
	if err != nil {
		return err
	}
	if err := s.ensurePermission(ctx, userID, "course.manage_content", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return err
	}
	if err := s.ensureCourseLecturer(ctx, userID, item.CourseID); err != nil {
		return err
	}
	return s.repo.DeleteMaterial(ctx, materialID)
}

func (s *MaterialService) resolveCourseID(ctx context.Context, courseID uint, courseCode string) (uint, error) {
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

func (s *MaterialService) ensurePermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) error {
	allowed, err := s.permissionService.UserHasPermission(ctx, userID, permissionCode, target)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func (s *MaterialService) ensureStudentCourseRegistration(ctx context.Context, userID uint, courseID *uint) error {
	if courseID == nil || *courseID == 0 {
		return nil
	}
	isStudent, err := s.repo.UserHasRole(ctx, userID, "student")
	if err != nil {
		return err
	}
	if !isStudent {
		return nil
	}
	registered, err := s.repo.StudentRegisteredForCourse(ctx, userID, *courseID)
	if err != nil {
		return err
	}
	if !registered {
		return ErrPermissionDenied
	}
	return nil
}

func (s *MaterialService) ensureCourseLecturer(ctx context.Context, userID, courseID uint) error {
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

func validateMaterialRequest(title, materialType, externalURL string) error {
	if strings.TrimSpace(title) == "" {
		return ValidationError{Message: "title is required"}
	}
	if !materialTypeOrDefault(materialType).Valid() {
		return ValidationError{Message: "invalid material_type"}
	}
	if strings.TrimSpace(externalURL) == "" {
		return ValidationError{Message: "external_url is required until file storage is connected"}
	}
	return nil
}

func materialTypeOrDefault(raw string) models.CourseMaterialType {
	item := models.CourseMaterialType(strings.ToLower(strings.TrimSpace(raw)))
	if item == "" {
		return models.CourseMaterialTypeDocument
	}
	if !item.Valid() {
		return models.CourseMaterialTypeDocument
	}
	return item
}

func mapCourseMaterials(items []models.CourseMaterial) []dto.CourseMaterialResponse {
	out := make([]dto.CourseMaterialResponse, 0, len(items))
	for i := range items {
		out = append(out, mapCourseMaterial(&items[i]))
	}
	return out
}

func mapCourseMaterial(item *models.CourseMaterial) dto.CourseMaterialResponse {
	return dto.CourseMaterialResponse{
		ID:               item.ID,
		UUID:             item.UUID,
		CourseID:         item.CourseID,
		CourseCode:       item.Course.Code,
		CourseTitle:      item.Course.Title,
		Title:            item.Title,
		Description:      item.Description,
		MaterialType:     string(item.MaterialType),
		OriginalFileName: item.OriginalFileName,
		ExternalURL:      item.ExternalURL,
		MimeType:         item.MimeType,
		SizeBytes:        item.SizeBytes,
		AllowDownload:    item.AllowDownload,
		UploadedBy:       item.UploadedBy,
		PublishedAt:      item.PublishedAt,
		FileURL:          item.ExternalURL,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}
