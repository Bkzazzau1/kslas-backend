package rbac

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
)

type Scope struct {
	Type models.ScopeType
	ID   *uint
}

type Authorizer struct {
	db *gorm.DB
}

type permissionGrant struct {
	RoleCode  string
	ScopeType models.ScopeType
	ScopeID   *uint
	StartsAt  *time.Time
	EndsAt    *time.Time
}

func NewAuthorizer(db *gorm.DB) *Authorizer {
	return &Authorizer{db: db}
}

func SchoolScope() Scope {
	return Scope{Type: models.ScopeSchool}
}

func FacultyScope(id uint) Scope {
	return Scope{Type: models.ScopeFaculty, ID: &id}
}

func DepartmentScope(id uint) Scope {
	return Scope{Type: models.ScopeDepartment, ID: &id}
}

func ProgrammeScope(id uint) Scope {
	return Scope{Type: models.ScopeProgramme, ID: &id}
}

func CourseScope(id uint) Scope {
	return Scope{Type: models.ScopeCourse, ID: &id}
}

func (a *Authorizer) HasPermission(ctx context.Context, userID uint, permissionCode string, target *Scope) (bool, error) {
	var grants []permissionGrant

	err := a.db.WithContext(ctx).
		Table("user_roles").
		Select("roles.code AS role_code, user_roles.scope_type, user_roles.scope_id, user_roles.starts_at, user_roles.ends_at").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Joins("JOIN role_permissions ON role_permissions.role_id = roles.id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("user_roles.user_id = ?", userID).
		Where("permissions.code = ?", permissionCode).
		Scan(&grants).Error
	if err != nil {
		return false, fmt.Errorf("query permission grants: %w", err)
	}

	if len(grants) == 0 {
		return false, nil
	}

	now := time.Now().UTC()

	if target == nil || target.Type == "" {
		for _, grant := range grants {
			if isGrantActive(grant, now) {
				return true, nil
			}
		}

		return false, nil
	}

	allowedScopes, err := a.expandScope(ctx, *target)
	if err != nil {
		return false, err
	}

	for _, grant := range grants {
		if !isGrantActive(grant, now) {
			continue
		}

		if scopeMatches(grant.ScopeType, grant.ScopeID, allowedScopes) {
			return true, nil
		}
	}

	return false, nil
}

func (a *Authorizer) expandScope(ctx context.Context, target Scope) ([]Scope, error) {
	if !target.Type.Valid() {
		return nil, fmt.Errorf("invalid target scope type %q", target.Type)
	}

	if target.Type == models.ScopeSchool {
		return []Scope{SchoolScope()}, nil
	}

	if target.ID == nil || *target.ID == 0 {
		return nil, fmt.Errorf("scope id is required for %s scope", target.Type)
	}

	scopes := []Scope{SchoolScope()}

	switch target.Type {
	case models.ScopeFaculty:
		scopes = append(scopes, FacultyScope(*target.ID))
	case models.ScopeDepartment:
		var department models.Department
		if err := a.db.WithContext(ctx).Select("id", "faculty_id").First(&department, *target.ID).Error; err != nil {
			return nil, fmt.Errorf("load department %d: %w", *target.ID, err)
		}

		scopes = append(scopes, FacultyScope(department.FacultyID), DepartmentScope(department.ID))
	case models.ScopeProgramme:
		var programme models.Programme
		if err := a.db.WithContext(ctx).Select("id", "department_id").First(&programme, *target.ID).Error; err != nil {
			return nil, fmt.Errorf("load programme %d: %w", *target.ID, err)
		}

		var department models.Department
		if err := a.db.WithContext(ctx).Select("id", "faculty_id").First(&department, programme.DepartmentID).Error; err != nil {
			return nil, fmt.Errorf("load department %d: %w", programme.DepartmentID, err)
		}

		scopes = append(scopes,
			FacultyScope(department.FacultyID),
			DepartmentScope(department.ID),
			ProgrammeScope(programme.ID),
		)
	case models.ScopeCourse:
		var course models.Course
		if err := a.db.WithContext(ctx).Select("id", "department_id", "programme_id").First(&course, *target.ID).Error; err != nil {
			return nil, fmt.Errorf("load course %d: %w", *target.ID, err)
		}

		var department models.Department
		if err := a.db.WithContext(ctx).Select("id", "faculty_id").First(&department, course.DepartmentID).Error; err != nil {
			return nil, fmt.Errorf("load department %d: %w", course.DepartmentID, err)
		}

		scopes = append(scopes, FacultyScope(department.FacultyID), DepartmentScope(department.ID))
		if course.ProgrammeID != nil && *course.ProgrammeID != 0 {
			scopes = append(scopes, ProgrammeScope(*course.ProgrammeID))
		}
		scopes = append(scopes, CourseScope(course.ID))
	default:
		return nil, fmt.Errorf("unsupported scope type %q", target.Type)
	}

	return scopes, nil
}

func isGrantActive(grant permissionGrant, at time.Time) bool {
	if grant.StartsAt != nil && at.Before(*grant.StartsAt) {
		return false
	}

	if grant.EndsAt != nil && at.After(*grant.EndsAt) {
		return false
	}

	return true
}

func scopeMatches(grantType models.ScopeType, grantID *uint, allowedScopes []Scope) bool {
	for _, scope := range allowedScopes {
		if scope.Type != grantType {
			continue
		}

		if scope.Type == models.ScopeSchool {
			return true
		}

		if scope.ID == nil || grantID == nil {
			continue
		}

		if *scope.ID == *grantID {
			return true
		}
	}

	return false
}
