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

var ErrPermissionDenied = errors.New("permission denied")

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type AcademicService struct {
	repo              *repository.AcademicRepository
	permissionService *PermissionService
}

func NewAcademicService(repo *repository.AcademicRepository, permissionService *PermissionService) *AcademicService {
	return &AcademicService{
		repo:              repo,
		permissionService: permissionService,
	}
}

func (s *AcademicService) ListFaculties(ctx context.Context, userID uint) ([]dto.FacultyResponse, error) {
	if err := s.ensurePermission(ctx, userID, "faculty.view", scopePtr(rbac.SchoolScope())); err != nil {
		return nil, err
	}

	items, err := s.repo.ListFaculties(ctx)
	if err != nil {
		return nil, err
	}

	return mapFaculties(items), nil
}

func (s *AcademicService) GetFaculty(ctx context.Context, userID, facultyID uint) (*dto.FacultyResponse, error) {
	if err := s.ensurePermission(ctx, userID, "faculty.view", scopePtr(rbac.FacultyScope(facultyID))); err != nil {
		return nil, err
	}

	item, err := s.repo.GetFaculty(ctx, facultyID)
	if err != nil {
		return nil, err
	}

	response := mapFaculty(item)
	return &response, nil
}

func (s *AcademicService) CreateFaculty(ctx context.Context, userID uint, req dto.FacultyCreateRequest) (*dto.FacultyResponse, error) {
	if err := s.ensurePermission(ctx, userID, "faculty.create", scopePtr(rbac.SchoolScope())); err != nil {
		return nil, err
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.TrimSpace(req.Code)
	if req.Name == "" || req.Code == "" {
		return nil, ValidationError{Message: "name and code are required"}
	}

	item := &models.Faculty{
		Name: req.Name,
		Code: req.Code,
	}
	if err := s.repo.CreateFaculty(ctx, item); err != nil {
		return nil, err
	}

	response := mapFaculty(item)
	return &response, nil
}

func (s *AcademicService) UpdateFaculty(ctx context.Context, userID, facultyID uint, req dto.FacultyUpdateRequest) (*dto.FacultyResponse, error) {
	if err := s.ensurePermission(ctx, userID, "faculty.update", scopePtr(rbac.FacultyScope(facultyID))); err != nil {
		return nil, err
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.TrimSpace(req.Code)
	if req.Name == "" || req.Code == "" {
		return nil, ValidationError{Message: "name and code are required"}
	}

	item, err := s.repo.GetFaculty(ctx, facultyID)
	if err != nil {
		return nil, err
	}

	item.Name = req.Name
	item.Code = req.Code
	if err := s.repo.UpdateFaculty(ctx, item); err != nil {
		return nil, err
	}

	response := mapFaculty(item)
	return &response, nil
}

func (s *AcademicService) DeleteFaculty(ctx context.Context, userID, facultyID uint) error {
	if err := s.ensurePermission(ctx, userID, "faculty.delete", scopePtr(rbac.FacultyScope(facultyID))); err != nil {
		return err
	}

	if _, err := s.repo.GetFaculty(ctx, facultyID); err != nil {
		return err
	}

	return s.repo.DeleteFaculty(ctx, facultyID)
}

func (s *AcademicService) ListDepartments(ctx context.Context, userID uint, facultyID *uint) ([]dto.DepartmentResponse, error) {
	target := scopePtr(rbac.SchoolScope())
	if facultyID != nil {
		target = scopePtr(rbac.FacultyScope(*facultyID))
	}

	if err := s.ensurePermission(ctx, userID, "department.view", target); err != nil {
		return nil, err
	}

	items, err := s.repo.ListDepartments(ctx, repository.DepartmentListFilter{FacultyID: facultyID})
	if err != nil {
		return nil, err
	}

	return mapDepartments(items), nil
}

func (s *AcademicService) GetDepartment(ctx context.Context, userID, departmentID uint) (*dto.DepartmentResponse, error) {
	if err := s.ensurePermission(ctx, userID, "department.view", scopePtr(rbac.DepartmentScope(departmentID))); err != nil {
		return nil, err
	}

	item, err := s.repo.GetDepartment(ctx, departmentID)
	if err != nil {
		return nil, err
	}

	response := mapDepartment(item)
	return &response, nil
}

func (s *AcademicService) CreateDepartment(ctx context.Context, userID uint, req dto.DepartmentCreateRequest) (*dto.DepartmentResponse, error) {
	if req.FacultyID == 0 {
		return nil, ValidationError{Message: "faculty_id is required"}
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.TrimSpace(req.Code)
	if req.Name == "" || req.Code == "" {
		return nil, ValidationError{Message: "name and code are required"}
	}

	if err := s.ensurePermission(ctx, userID, "department.create", scopePtr(rbac.FacultyScope(req.FacultyID))); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetFaculty(ctx, req.FacultyID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ValidationError{Message: "faculty not found"}
		}
		return nil, err
	}

	item := &models.Department{
		FacultyID: req.FacultyID,
		Name:      req.Name,
		Code:      req.Code,
	}
	if err := s.repo.CreateDepartment(ctx, item); err != nil {
		return nil, err
	}

	response := mapDepartment(item)
	return &response, nil
}

func (s *AcademicService) UpdateDepartment(ctx context.Context, userID, departmentID uint, req dto.DepartmentUpdateRequest) (*dto.DepartmentResponse, error) {
	if req.FacultyID == 0 {
		return nil, ValidationError{Message: "faculty_id is required"}
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.TrimSpace(req.Code)
	if req.Name == "" || req.Code == "" {
		return nil, ValidationError{Message: "name and code are required"}
	}

	item, err := s.repo.GetDepartment(ctx, departmentID)
	if err != nil {
		return nil, err
	}

	if err := s.ensurePermission(ctx, userID, "department.update", scopePtr(rbac.DepartmentScope(departmentID))); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetFaculty(ctx, req.FacultyID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ValidationError{Message: "faculty not found"}
		}
		return nil, err
	}

	item.FacultyID = req.FacultyID
	item.Name = req.Name
	item.Code = req.Code
	if err := s.repo.UpdateDepartment(ctx, item); err != nil {
		return nil, err
	}

	response := mapDepartment(item)
	return &response, nil
}

func (s *AcademicService) DeleteDepartment(ctx context.Context, userID, departmentID uint) error {
	if err := s.ensurePermission(ctx, userID, "department.delete", scopePtr(rbac.DepartmentScope(departmentID))); err != nil {
		return err
	}

	if _, err := s.repo.GetDepartment(ctx, departmentID); err != nil {
		return err
	}

	return s.repo.DeleteDepartment(ctx, departmentID)
}

func (s *AcademicService) ListProgrammes(ctx context.Context, userID uint, departmentID *uint) ([]dto.ProgrammeResponse, error) {
	target := scopePtr(rbac.SchoolScope())
	if departmentID != nil {
		target = scopePtr(rbac.DepartmentScope(*departmentID))
	}

	if err := s.ensurePermission(ctx, userID, "programme.view", target); err != nil {
		return nil, err
	}

	items, err := s.repo.ListProgrammes(ctx, repository.ProgrammeListFilter{DepartmentID: departmentID})
	if err != nil {
		return nil, err
	}

	return mapProgrammes(items), nil
}

func (s *AcademicService) GetProgramme(ctx context.Context, userID, programmeID uint) (*dto.ProgrammeResponse, error) {
	if err := s.ensurePermission(ctx, userID, "programme.view", scopePtr(rbac.ProgrammeScope(programmeID))); err != nil {
		return nil, err
	}

	item, err := s.repo.GetProgramme(ctx, programmeID)
	if err != nil {
		return nil, err
	}

	response := mapProgramme(item)
	return &response, nil
}

func (s *AcademicService) CreateProgramme(ctx context.Context, userID uint, req dto.ProgrammeCreateRequest) (*dto.ProgrammeResponse, error) {
	if req.DepartmentID == 0 {
		return nil, ValidationError{Message: "department_id is required"}
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.TrimSpace(req.Code)
	req.LevelType = strings.TrimSpace(req.LevelType)
	if req.Name == "" || req.Code == "" || req.LevelType == "" {
		return nil, ValidationError{Message: "name, code, and level_type are required"}
	}

	if err := s.ensurePermission(ctx, userID, "programme.create", scopePtr(rbac.DepartmentScope(req.DepartmentID))); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetDepartment(ctx, req.DepartmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ValidationError{Message: "department not found"}
		}
		return nil, err
	}

	item := &models.Programme{
		DepartmentID: req.DepartmentID,
		Name:         req.Name,
		Code:         req.Code,
		LevelType:    req.LevelType,
	}
	if err := s.repo.CreateProgramme(ctx, item); err != nil {
		return nil, err
	}

	response := mapProgramme(item)
	return &response, nil
}

func (s *AcademicService) UpdateProgramme(ctx context.Context, userID, programmeID uint, req dto.ProgrammeUpdateRequest) (*dto.ProgrammeResponse, error) {
	if req.DepartmentID == 0 {
		return nil, ValidationError{Message: "department_id is required"}
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.TrimSpace(req.Code)
	req.LevelType = strings.TrimSpace(req.LevelType)
	if req.Name == "" || req.Code == "" || req.LevelType == "" {
		return nil, ValidationError{Message: "name, code, and level_type are required"}
	}

	item, err := s.repo.GetProgramme(ctx, programmeID)
	if err != nil {
		return nil, err
	}

	if err := s.ensurePermission(ctx, userID, "programme.update", scopePtr(rbac.ProgrammeScope(programmeID))); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetDepartment(ctx, req.DepartmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ValidationError{Message: "department not found"}
		}
		return nil, err
	}

	item.DepartmentID = req.DepartmentID
	item.Name = req.Name
	item.Code = req.Code
	item.LevelType = req.LevelType
	if err := s.repo.UpdateProgramme(ctx, item); err != nil {
		return nil, err
	}

	response := mapProgramme(item)
	return &response, nil
}

func (s *AcademicService) DeleteProgramme(ctx context.Context, userID, programmeID uint) error {
	if err := s.ensurePermission(ctx, userID, "programme.delete", scopePtr(rbac.ProgrammeScope(programmeID))); err != nil {
		return err
	}

	if _, err := s.repo.GetProgramme(ctx, programmeID); err != nil {
		return err
	}

	return s.repo.DeleteProgramme(ctx, programmeID)
}

func (s *AcademicService) ListCourses(ctx context.Context, userID uint, filter repository.CourseListFilter) ([]dto.CourseResponse, error) {
	target := scopePtr(rbac.SchoolScope())
	switch {
	case filter.ProgrammeID != nil:
		target = scopePtr(rbac.ProgrammeScope(*filter.ProgrammeID))
	case filter.DepartmentID != nil:
		target = scopePtr(rbac.DepartmentScope(*filter.DepartmentID))
	}

	if err := s.ensurePermission(ctx, userID, "course.view", target); err != nil {
		return nil, err
	}

	items, err := s.repo.ListCourses(ctx, filter)
	if err != nil {
		return nil, err
	}

	return mapCourses(items), nil
}

func (s *AcademicService) GetCourse(ctx context.Context, userID, courseID uint) (*dto.CourseResponse, error) {
	if err := s.ensurePermission(ctx, userID, "course.view", scopePtr(rbac.CourseScope(courseID))); err != nil {
		return nil, err
	}

	item, err := s.repo.GetCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	response := mapCourse(item)
	return &response, nil
}

func (s *AcademicService) CreateCourse(ctx context.Context, userID uint, req dto.CourseCreateRequest) (*dto.CourseResponse, error) {
	if err := validateCourseRequest(req); err != nil {
		return nil, err
	}

	if err := s.ensurePermission(ctx, userID, "course.create", scopePtr(rbac.DepartmentScope(req.DepartmentID))); err != nil {
		return nil, err
	}

	if err := s.validateCourseRelations(ctx, req.DepartmentID, req.ProgrammeID); err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	item := &models.Course{
		DepartmentID: req.DepartmentID,
		ProgrammeID:  req.ProgrammeID,
		Title:        strings.TrimSpace(req.Title),
		Code:         strings.TrimSpace(req.Code),
		Unit:         req.Unit,
		Semester:     strings.TrimSpace(req.Semester),
		Level:        strings.TrimSpace(req.Level),
		IsActive:     isActive,
	}
	if err := s.repo.CreateCourse(ctx, item); err != nil {
		return nil, err
	}

	response := mapCourse(item)
	return &response, nil
}

func (s *AcademicService) UpdateCourse(ctx context.Context, userID, courseID uint, req dto.CourseUpdateRequest) (*dto.CourseResponse, error) {
	if err := validateCourseRequest(dto.CourseCreateRequest(req)); err != nil {
		return nil, err
	}

	item, err := s.repo.GetCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if err := s.ensurePermission(ctx, userID, "course.update", scopePtr(rbac.CourseScope(courseID))); err != nil {
		return nil, err
	}

	if err := s.validateCourseRelations(ctx, req.DepartmentID, req.ProgrammeID); err != nil {
		return nil, err
	}

	isActive := item.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	item.DepartmentID = req.DepartmentID
	item.ProgrammeID = req.ProgrammeID
	item.Title = strings.TrimSpace(req.Title)
	item.Code = strings.TrimSpace(req.Code)
	item.Unit = req.Unit
	item.Semester = strings.TrimSpace(req.Semester)
	item.Level = strings.TrimSpace(req.Level)
	item.IsActive = isActive
	if err := s.repo.UpdateCourse(ctx, item); err != nil {
		return nil, err
	}

	response := mapCourse(item)
	return &response, nil
}

func (s *AcademicService) DeleteCourse(ctx context.Context, userID, courseID uint) error {
	if err := s.ensurePermission(ctx, userID, "course.delete", scopePtr(rbac.CourseScope(courseID))); err != nil {
		return err
	}

	if _, err := s.repo.GetCourse(ctx, courseID); err != nil {
		return err
	}

	return s.repo.DeleteCourse(ctx, courseID)
}

func (s *AcademicService) ensurePermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) error {
	allowed, err := s.permissionService.UserHasPermission(ctx, userID, permissionCode, target)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}

	return nil
}

func (s *AcademicService) validateCourseRelations(ctx context.Context, departmentID uint, programmeID *uint) error {
	if departmentID == 0 {
		return ValidationError{Message: "department_id is required"}
	}

	if _, err := s.repo.GetDepartment(ctx, departmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ValidationError{Message: "department not found"}
		}
		return err
	}

	if programmeID == nil {
		return nil
	}

	programme, err := s.repo.GetProgramme(ctx, *programmeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ValidationError{Message: "programme not found"}
		}
		return err
	}

	if programme.DepartmentID != departmentID {
		return ValidationError{Message: "programme does not belong to the selected department"}
	}

	return nil
}

func validateCourseRequest(req dto.CourseCreateRequest) error {
	if req.DepartmentID == 0 {
		return ValidationError{Message: "department_id is required"}
	}

	if strings.TrimSpace(req.Title) == "" ||
		strings.TrimSpace(req.Code) == "" ||
		strings.TrimSpace(req.Semester) == "" ||
		strings.TrimSpace(req.Level) == "" {
		return ValidationError{Message: "title, code, semester, and level are required"}
	}

	if req.Unit == 0 {
		return ValidationError{Message: "unit must be greater than zero"}
	}

	return nil
}

func scopePtr(scope rbac.Scope) *rbac.Scope {
	return &scope
}

func mapFaculties(items []models.Faculty) []dto.FacultyResponse {
	out := make([]dto.FacultyResponse, 0, len(items))
	for i := range items {
		out = append(out, mapFaculty(&items[i]))
	}
	return out
}

func mapFaculty(item *models.Faculty) dto.FacultyResponse {
	return dto.FacultyResponse{
		ID:        item.ID,
		Name:      item.Name,
		Code:      item.Code,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapDepartments(items []models.Department) []dto.DepartmentResponse {
	out := make([]dto.DepartmentResponse, 0, len(items))
	for i := range items {
		out = append(out, mapDepartment(&items[i]))
	}
	return out
}

func mapDepartment(item *models.Department) dto.DepartmentResponse {
	return dto.DepartmentResponse{
		ID:        item.ID,
		FacultyID: item.FacultyID,
		Name:      item.Name,
		Code:      item.Code,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapProgrammes(items []models.Programme) []dto.ProgrammeResponse {
	out := make([]dto.ProgrammeResponse, 0, len(items))
	for i := range items {
		out = append(out, mapProgramme(&items[i]))
	}
	return out
}

func mapProgramme(item *models.Programme) dto.ProgrammeResponse {
	return dto.ProgrammeResponse{
		ID:           item.ID,
		DepartmentID: item.DepartmentID,
		Name:         item.Name,
		Code:         item.Code,
		LevelType:    item.LevelType,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func mapCourses(items []models.Course) []dto.CourseResponse {
	out := make([]dto.CourseResponse, 0, len(items))
	for i := range items {
		out = append(out, mapCourse(&items[i]))
	}
	return out
}

func mapCourse(item *models.Course) dto.CourseResponse {
	return dto.CourseResponse{
		ID:           item.ID,
		DepartmentID: item.DepartmentID,
		ProgrammeID:  item.ProgrammeID,
		Title:        item.Title,
		Code:         item.Code,
		Unit:         item.Unit,
		Semester:     item.Semester,
		Level:        item.Level,
		IsActive:     item.IsActive,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}
