package services

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"kslasbackend/internal/database/models"
	"kslasbackend/internal/dto"
	"kslasbackend/internal/rbac"
	"kslasbackend/internal/repository"
)

type AdministrationService struct {
	repo              *repository.TeachingRepository
	passwordService   *PasswordService
	permissionService *PermissionService
}

func NewAdministrationService(repo *repository.TeachingRepository, passwordService *PasswordService, permissionService *PermissionService) *AdministrationService {
	return &AdministrationService{
		repo:              repo,
		passwordService:   passwordService,
		permissionService: permissionService,
	}
}

func (s *AdministrationService) CreateStaff(ctx context.Context, userID uint, req dto.StaffCreateRequest) (*dto.UserResponse, error) {
	roleCode := strings.ToLower(strings.TrimSpace(req.RoleCode))
	if !allowedStaffRole(roleCode) {
		return nil, ValidationError{Message: "unsupported staff role"}
	}
	roleCode = normalizeStaffRole(roleCode)
	if req.DepartmentID == 0 {
		return nil, ValidationError{Message: "department_id is required"}
	}
	if err := s.ensurePermission(ctx, userID, "user.create", scopePtr(rbac.DepartmentScope(req.DepartmentID))); err != nil {
		return nil, err
	}
	hash, err := s.passwordService.Hash(req.Password)
	if err != nil {
		return nil, ValidationError{Message: err.Error()}
	}
	staffID := strings.TrimSpace(req.StaffID)
	user := &models.User{
		UUID:         uuid.NewString(),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		MiddleName:   req.MiddleName,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hash,
		Gender:       req.Gender,
		MatricNo:     "STAFF-" + staffID,
		StaffID:      staffID,
		UserType:     models.UserTypeStaff,
		Status:       models.UserStatusActive,
	}
	if err := s.repo.CreateUserWithRole(ctx, user, roleCode, models.ScopeDepartment, &req.DepartmentID, userID); err != nil {
		return nil, err
	}
	response := mapUserResponse(user)
	return &response, nil
}

func (s *AdministrationService) CreateStudent(ctx context.Context, userID uint, req dto.StudentCreateRequest) (*dto.StudentRegistrationResponse, error) {
	if req.DepartmentID == 0 || req.ProgrammeID == 0 {
		return nil, ValidationError{Message: "department_id and programme_id are required"}
	}
	if err := s.ensurePermission(ctx, userID, "user.create", scopePtr(rbac.ProgrammeScope(req.ProgrammeID))); err != nil {
		return nil, err
	}
	hash, err := s.passwordService.Hash(req.Password)
	if err != nil {
		return nil, ValidationError{Message: err.Error()}
	}
	user := &models.User{
		UUID:         uuid.NewString(),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		MiddleName:   req.MiddleName,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hash,
		Gender:       req.Gender,
		MatricNo:     req.MatricNo,
		StaffID:      "STUDENT-" + strings.TrimSpace(req.MatricNo),
		UserType:     models.UserTypeStudent,
		Status:       models.UserStatusActive,
	}
	profile := &models.StudentAcademicProfile{
		DepartmentID:    req.DepartmentID,
		ProgrammeID:     req.ProgrammeID,
		Level:           req.Level,
		Semester:        req.Semester,
		AcademicSession: req.AcademicSession,
		RegisteredBy:    userID,
	}
	if err := s.repo.CreateStudentWithProfile(ctx, user, profile, userID); err != nil {
		return nil, err
	}
	return &dto.StudentRegistrationResponse{
		User:    mapUserResponse(user),
		Profile: mapStudentProfile(profile),
	}, nil
}

func (s *AdministrationService) AssignLecturerToCourse(ctx context.Context, userID, courseID, lecturerID uint) error {
	if err := s.ensurePermission(ctx, userID, "course.assign_lecturer", scopePtr(rbac.CourseScope(courseID))); err != nil {
		return err
	}
	return s.repo.AssignLecturerToCourse(ctx, courseID, lecturerID, userID)
}

func (s *AdministrationService) ListEligibleCourses(ctx context.Context, studentID uint, semester string) ([]dto.CourseResponse, error) {
	profile, err := s.repo.GetStudentAcademicProfile(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, studentID, "course.register", scopePtr(rbac.ProgrammeScope(profile.ProgrammeID))); err != nil {
		return nil, err
	}
	items, err := s.repo.ListEligibleCoursesForStudent(ctx, *profile, semester)
	if err != nil {
		return nil, err
	}
	return mapCourses(items), nil
}

func (s *AdministrationService) RegisterCourses(ctx context.Context, studentID uint, req dto.CourseRegistrationCreateRequest) ([]dto.CourseRegistrationResponse, error) {
	if len(req.CourseIDs) == 0 {
		return nil, ValidationError{Message: "course_ids is required"}
	}
	profile, err := s.repo.GetStudentAcademicProfile(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, studentID, "course.register", scopePtr(rbac.ProgrammeScope(profile.ProgrammeID))); err != nil {
		return nil, err
	}
	academicSession := strings.TrimSpace(req.AcademicSession)
	if academicSession == "" {
		academicSession = profile.AcademicSession
	}
	semester := strings.TrimSpace(req.Semester)
	if semester == "" {
		semester = profile.Semester
	}
	items, err := s.repo.RegisterStudentCourses(ctx, *profile, req.CourseIDs, academicSession, semester)
	if err != nil {
		return nil, err
	}
	return mapCourseRegistrations(items), nil
}

func (s *AdministrationService) ListMyCourseRegistrations(ctx context.Context, studentID uint) ([]dto.CourseRegistrationResponse, error) {
	profile, err := s.repo.GetStudentAcademicProfile(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, studentID, "course.view", scopePtr(rbac.ProgrammeScope(profile.ProgrammeID))); err != nil {
		return nil, err
	}
	items, err := s.repo.ListStudentCourseRegistrations(ctx, studentID)
	if err != nil {
		return nil, err
	}
	return mapCourseRegistrations(items), nil
}

func (s *AdministrationService) ensurePermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) error {
	allowed, err := s.permissionService.UserHasPermission(ctx, userID, permissionCode, target)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func mapUserResponse(user *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:         user.ID,
		UUID:       user.UUID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Email:      user.Email,
		Phone:      user.Phone,
		Gender:     user.Gender,
		MatricNo:   user.MatricNo,
		StaffID:    user.StaffID,
		UserType:   string(user.UserType),
		Status:     string(user.Status),
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

func mapStudentProfile(profile *models.StudentAcademicProfile) dto.StudentAcademicProfileResponse {
	return dto.StudentAcademicProfileResponse{
		ID:                 profile.ID,
		StudentID:          profile.StudentID,
		DepartmentID:       profile.DepartmentID,
		ProgrammeID:        profile.ProgrammeID,
		Level:              profile.Level,
		Semester:           profile.Semester,
		AcademicSession:    profile.AcademicSession,
		CountryOfResidence: profile.CountryOfResidence,
		IsInternational:    profile.IsInternational,
		RegisteredBy:       profile.RegisteredBy,
		CreatedAt:          profile.CreatedAt,
		UpdatedAt:          profile.UpdatedAt,
	}
}

func mapCourseRegistrations(items []models.CourseRegistration) []dto.CourseRegistrationResponse {
	out := make([]dto.CourseRegistrationResponse, 0, len(items))
	for i := range items {
		out = append(out, dto.CourseRegistrationResponse{
			ID:              items[i].ID,
			StudentID:       items[i].StudentID,
			CourseID:        items[i].CourseID,
			CourseCode:      items[i].Course.Code,
			CourseTitle:     items[i].Course.Title,
			AcademicSession: items[i].AcademicSession,
			Semester:        items[i].Semester,
			Level:           items[i].Level,
			Status:          string(items[i].Status),
			RegisteredAt:    items[i].RegisteredAt,
			Course:          mapCourse(&items[i].Course),
		})
	}
	return out
}
