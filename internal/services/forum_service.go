package services

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
	"kslasbackend/internal/dto"
	"kslasbackend/internal/rbac"
	"kslasbackend/internal/repository"
)

type ForumService struct {
	repo              *repository.TeachingRepository
	permissionService *PermissionService
}

func NewForumService(repo *repository.TeachingRepository, permissionService *PermissionService) *ForumService {
	return &ForumService{repo: repo, permissionService: permissionService}
}

func (s *ForumService) ListCoursePosts(ctx context.Context, userID, courseID uint) ([]dto.CourseForumPostResponse, error) {
	if err := s.ensureCourseAccess(ctx, userID, courseID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListCourseForumPosts(ctx, courseID)
	if err != nil {
		return nil, err
	}
	return mapCourseForumPosts(items), nil
}

func (s *ForumService) CreateCoursePost(ctx context.Context, userID, courseID uint, req dto.CourseForumPostCreateRequest) (*dto.CourseForumPostResponse, error) {
	if err := s.ensureCourseAccess(ctx, userID, courseID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Body) == "" {
		return nil, ValidationError{Message: "body is required"}
	}
	if req.ParentID != nil {
		parent, err := s.repo.GetCourseForumPost(ctx, *req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ValidationError{Message: "parent forum post not found"}
			}
			return nil, err
		}
		if parent.CourseID != courseID {
			return nil, ValidationError{Message: "parent forum post does not belong to this course"}
		}
		if parent.IsLocked {
			return nil, ValidationError{Message: "this discussion is locked by the lecturer"}
		}
	}

	item := &models.CourseForumPost{
		CourseID: courseID,
		AuthorID: userID,
		Title:    req.Title,
		Body:     req.Body,
		ParentID: req.ParentID,
	}
	if err := s.repo.CreateCourseForumPost(ctx, item); err != nil {
		return nil, err
	}
	item, err := s.repo.GetCourseForumPost(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	response := mapCourseForumPost(item)
	return &response, nil
}

func (s *ForumService) ModerateCoursePost(ctx context.Context, userID, courseID, postID uint, req dto.CourseForumPostModerationRequest) (*dto.CourseForumPostResponse, error) {
	if err := s.ensureCourseLecturer(ctx, userID, courseID); err != nil {
		return nil, err
	}
	item, err := s.repo.GetCourseForumPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	if item.CourseID != courseID {
		return nil, ValidationError{Message: "forum post does not belong to this course"}
	}
	if req.IsPinned != nil {
		item.IsPinned = *req.IsPinned
	}
	if req.IsLocked != nil {
		item.IsLocked = *req.IsLocked
	}
	if err := s.repo.UpdateCourseForumPost(ctx, item); err != nil {
		return nil, err
	}
	response := mapCourseForumPost(item)
	return &response, nil
}

func (s *ForumService) ensureCourseAccess(ctx context.Context, userID, courseID uint) error {
	if courseID == 0 {
		return ValidationError{Message: "course_id is required"}
	}
	if err := s.ensurePermission(ctx, userID, "course.view", scopePtr(rbac.CourseScope(courseID))); err != nil {
		return err
	}
	isStudent, err := s.repo.UserHasRole(ctx, userID, "student")
	if err != nil {
		return err
	}
	if isStudent {
		registered, err := s.repo.StudentRegisteredForCourse(ctx, userID, courseID)
		if err != nil {
			return err
		}
		if !registered {
			return ErrPermissionDenied
		}
		return nil
	}
	return nil
}

func (s *ForumService) ensureCourseLecturer(ctx context.Context, userID, courseID uint) error {
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
	if err := s.ensurePermission(ctx, userID, "course.manage_content", scopePtr(rbac.CourseScope(courseID))); err != nil {
		return err
	}
	return nil
}

func (s *ForumService) ensurePermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) error {
	allowed, err := s.permissionService.UserHasPermission(ctx, userID, permissionCode, target)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func mapCourseForumPosts(items []models.CourseForumPost) []dto.CourseForumPostResponse {
	out := make([]dto.CourseForumPostResponse, 0, len(items))
	for i := range items {
		out = append(out, mapCourseForumPost(&items[i]))
	}
	return out
}

func mapCourseForumPost(item *models.CourseForumPost) dto.CourseForumPostResponse {
	return dto.CourseForumPostResponse{
		ID:                item.ID,
		UUID:              item.UUID,
		CourseID:          item.CourseID,
		CourseCode:        item.Course.Code,
		CourseTitle:       item.Course.Title,
		AuthorID:          item.AuthorID,
		AuthorDisplayName: forumAuthorName(item.Author),
		AuthorRole:        forumAuthorRole(item.Author),
		Title:             item.Title,
		Body:              item.Body,
		IsPinned:          item.IsPinned,
		IsLocked:          item.IsLocked,
		ParentID:          item.ParentID,
		Replies:           mapCourseForumPosts(item.Replies),
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func forumAuthorName(user models.User) string {
	name := strings.TrimSpace(strings.TrimSpace(user.FirstName) + " " + strings.TrimSpace(user.LastName))
	if name == "" {
		return "Course member"
	}
	return name
}

func forumAuthorRole(user models.User) string {
	for _, assignment := range user.UserRoles {
		code := strings.TrimSpace(assignment.Role.Code)
		if code == "lecturer" {
			return "lecturer"
		}
	}
	return "student"
}
