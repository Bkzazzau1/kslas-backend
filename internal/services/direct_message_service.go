package services

import (
	"context"
	"strings"

	"kslasbackend/internal/database/models"
	"kslasbackend/internal/dto"
	"kslasbackend/internal/rbac"
	"kslasbackend/internal/repository"
)

type DirectMessageService struct {
	repo              *repository.TeachingRepository
	permissionService *PermissionService
}

func NewDirectMessageService(repo *repository.TeachingRepository, permissionService *PermissionService) *DirectMessageService {
	return &DirectMessageService{repo: repo, permissionService: permissionService}
}

func (s *DirectMessageService) ListCourseMessages(ctx context.Context, userID, courseID uint, withUserID *uint) ([]dto.CourseDirectMessageResponse, error) {
	if err := s.ensureCourseParticipant(ctx, userID, courseID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListCourseDirectMessages(ctx, courseID, userID, withUserID)
	if err != nil {
		return nil, err
	}
	return mapCourseDirectMessages(items), nil
}

func (s *DirectMessageService) CreateCourseMessage(ctx context.Context, userID, courseID uint, req dto.CourseDirectMessageCreateRequest) (*dto.CourseDirectMessageResponse, error) {
	if strings.TrimSpace(req.Body) == "" {
		return nil, ValidationError{Message: "body is required"}
	}
	if err := s.ensureCourseParticipant(ctx, userID, courseID); err != nil {
		return nil, err
	}
	if err := s.ensureAllowedPair(ctx, userID, req.RecipientID, courseID); err != nil {
		return nil, err
	}
	item := &models.CourseDirectMessage{
		CourseID:    courseID,
		SenderID:    userID,
		RecipientID: req.RecipientID,
		Body:        req.Body,
	}
	if err := s.repo.CreateCourseDirectMessage(ctx, item); err != nil {
		return nil, err
	}
	items, err := s.repo.ListCourseDirectMessages(ctx, courseID, userID, &req.RecipientID)
	if err != nil {
		return nil, err
	}
	response := mapCourseDirectMessage(&items[len(items)-1])
	return &response, nil
}

func (s *DirectMessageService) ensureCourseParticipant(ctx context.Context, userID, courseID uint) error {
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
	isLecturer, err := s.repo.UserHasRole(ctx, userID, "lecturer")
	if err != nil {
		return err
	}
	if isLecturer {
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

func (s *DirectMessageService) ensureAllowedPair(ctx context.Context, senderID, recipientID, courseID uint) error {
	if recipientID == 0 || senderID == recipientID {
		return ValidationError{Message: "valid recipient_id is required"}
	}
	senderIsStudent, err := s.repo.UserHasRole(ctx, senderID, "student")
	if err != nil {
		return err
	}
	recipientIsLecturer, err := s.repo.UserHasRole(ctx, recipientID, "lecturer")
	if err != nil {
		return err
	}
	if senderIsStudent {
		if !recipientIsLecturer {
			return ErrPermissionDenied
		}
		assigned, err := s.repo.LecturerAssignedToCourse(ctx, recipientID, courseID)
		if err != nil {
			return err
		}
		if !assigned {
			return ErrPermissionDenied
		}
		return nil
	}
	senderIsLecturer, err := s.repo.UserHasRole(ctx, senderID, "lecturer")
	if err != nil {
		return err
	}
	if senderIsLecturer {
		registered, err := s.repo.StudentRegisteredForCourse(ctx, recipientID, courseID)
		if err != nil {
			return err
		}
		if !registered {
			return ErrPermissionDenied
		}
	}
	return nil
}

func (s *DirectMessageService) ensurePermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) error {
	allowed, err := s.permissionService.UserHasPermission(ctx, userID, permissionCode, target)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func mapCourseDirectMessages(items []models.CourseDirectMessage) []dto.CourseDirectMessageResponse {
	out := make([]dto.CourseDirectMessageResponse, 0, len(items))
	for i := range items {
		out = append(out, mapCourseDirectMessage(&items[i]))
	}
	return out
}

func mapCourseDirectMessage(item *models.CourseDirectMessage) dto.CourseDirectMessageResponse {
	return dto.CourseDirectMessageResponse{
		ID:                   item.ID,
		UUID:                 item.UUID,
		CourseID:             item.CourseID,
		CourseCode:           item.Course.Code,
		CourseTitle:          item.Course.Title,
		SenderID:             item.SenderID,
		SenderDisplayName:    displayName(item.Sender),
		RecipientID:          item.RecipientID,
		RecipientDisplayName: displayName(item.Recipient),
		Body:                 item.Body,
		ReadAt:               item.ReadAt,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}
}
