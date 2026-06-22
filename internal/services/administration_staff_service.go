package services

import (
	"context"
	"strings"
	"time"

	"kslasbackend/internal/database/models"
	"kslasbackend/internal/dto"
)

func allowedStaffRole(roleCode string) bool {
	switch strings.ToLower(strings.TrimSpace(roleCode)) {
	case "system_admin", "academic_admin", "dean", "hod", "programme_coordinator", "exam_officer", "lecturer", "moderator", "proctor", "content_manager", "registry_officer", "student_affairs", "marker", "teaching_assistant":
		return true
	case "dlc_director", "level_adviser", "academic_records", "invigilator":
		return true
	default:
		return false
	}
}

func normalizeStaffRole(roleCode string) string {
	roleCode = strings.ToLower(strings.TrimSpace(roleCode))
	switch roleCode {
	case "dlc_director":
		return "programme_coordinator"
	case "academic_records":
		return "registry_officer"
	case "invigilator":
		return "proctor"
	case "level_adviser":
		return "programme_coordinator"
	default:
		return roleCode
	}
}

func (s *AdministrationService) ListStaff(ctx context.Context, userID uint) ([]dto.StaffResponse, error) {
	items, err := s.repo.ListStaff(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.StaffResponse, 0, len(items))
	for i := range items {
		out = append(out, mapStaffResponse(&items[i]))
	}
	return out, nil
}

func (s *AdministrationService) ResetStaffPassword(ctx context.Context, userID, staffUserID uint, temporaryPassword string) error {
	if strings.TrimSpace(temporaryPassword) == "" {
		return ValidationError{Message: "password is required"}
	}
	hash, err := s.passwordService.Hash(temporaryPassword)
	if err != nil {
		return ValidationError{Message: err.Error()}
	}
	return s.repo.SaveStaffCredentialHash(ctx, staffUserID, hash)
}

func (s *AdministrationService) UpdateStaffStatus(ctx context.Context, userID, staffUserID uint, status string) (*dto.StaffResponse, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "active" && status != "inactive" && status != "suspended" {
		return nil, ValidationError{Message: "status must be active, inactive, or suspended"}
	}
	if err := s.repo.UpdateStaffStatus(ctx, staffUserID, models.UserStatus(status)); err != nil {
		return nil, err
	}
	user, err := s.repo.GetStaffUser(ctx, staffUserID)
	if err != nil {
		return nil, err
	}
	response := mapStaffResponse(user)
	return &response, nil
}

func mapStaffResponse(user *models.User) dto.StaffResponse {
	now := time.Now().UTC()
	roles := make([]dto.UserRolePayload, 0, len(user.UserRoles))
	primaryRole := ""
	for _, assignment := range user.UserRoles {
		if !assignment.IsActiveAt(now) {
			continue
		}
		role := dto.UserRolePayload{
			Code:      assignment.Role.Code,
			Name:      assignment.Role.Name,
			ScopeType: string(assignment.ScopeType),
			ScopeID:   assignment.ScopeID,
			IsPrimary: assignment.IsPrimary,
		}
		roles = append(roles, role)
		if assignment.IsPrimary && primaryRole == "" {
			primaryRole = assignment.Role.Code
		}
	}
	return dto.StaffResponse{
		ID:              user.ID,
		UUID:            user.UUID,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		MiddleName:      user.MiddleName,
		Email:           user.Email,
		Phone:           user.Phone,
		Gender:          user.Gender,
		MatricNo:        user.MatricNo,
		StaffID:         user.StaffID,
		UserType:        string(user.UserType),
		Status:          string(user.Status),
		PrimaryRoleCode: primaryRole,
		Roles:           roles,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}
