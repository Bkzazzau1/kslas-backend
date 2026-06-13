package services

import (
	"context"

	"kslasbackend/internal/rbac"
)

type PermissionService struct {
	authorizer *rbac.Authorizer
}

func NewPermissionService(authorizer *rbac.Authorizer) *PermissionService {
	return &PermissionService{authorizer: authorizer}
}

func (s *PermissionService) UserHasPermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) (bool, error) {
	return s.authorizer.HasPermission(ctx, userID, permissionCode, target)
}
