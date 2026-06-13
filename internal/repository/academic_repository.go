package repository

import (
	"context"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
)

type AcademicRepository struct {
	db *gorm.DB
}

type DepartmentListFilter struct {
	FacultyID *uint
}

type ProgrammeListFilter struct {
	DepartmentID *uint
}

type CourseListFilter struct {
	DepartmentID *uint
	ProgrammeID  *uint
	Semester     string
	Level        string
	IsActive     *bool
}

func NewAcademicRepository(db *gorm.DB) *AcademicRepository {
	return &AcademicRepository{db: db}
}

func (r *AcademicRepository) ListFaculties(ctx context.Context) ([]models.Faculty, error) {
	var items []models.Faculty
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&items).Error
	return items, err
}

func (r *AcademicRepository) GetFaculty(ctx context.Context, facultyID uint) (*models.Faculty, error) {
	var item models.Faculty
	err := r.db.WithContext(ctx).First(&item, facultyID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *AcademicRepository) CreateFaculty(ctx context.Context, faculty *models.Faculty) error {
	return r.db.WithContext(ctx).Create(faculty).Error
}

func (r *AcademicRepository) UpdateFaculty(ctx context.Context, faculty *models.Faculty) error {
	return r.db.WithContext(ctx).Save(faculty).Error
}

func (r *AcademicRepository) DeleteFaculty(ctx context.Context, facultyID uint) error {
	return r.db.WithContext(ctx).Delete(&models.Faculty{}, facultyID).Error
}

func (r *AcademicRepository) ListDepartments(ctx context.Context, filter DepartmentListFilter) ([]models.Department, error) {
	query := r.db.WithContext(ctx).Model(&models.Department{})
	if filter.FacultyID != nil {
		query = query.Where("faculty_id = ?", *filter.FacultyID)
	}

	var items []models.Department
	err := query.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *AcademicRepository) GetDepartment(ctx context.Context, departmentID uint) (*models.Department, error) {
	var item models.Department
	err := r.db.WithContext(ctx).First(&item, departmentID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *AcademicRepository) CreateDepartment(ctx context.Context, department *models.Department) error {
	return r.db.WithContext(ctx).Create(department).Error
}

func (r *AcademicRepository) UpdateDepartment(ctx context.Context, department *models.Department) error {
	return r.db.WithContext(ctx).Save(department).Error
}

func (r *AcademicRepository) DeleteDepartment(ctx context.Context, departmentID uint) error {
	return r.db.WithContext(ctx).Delete(&models.Department{}, departmentID).Error
}

func (r *AcademicRepository) ListProgrammes(ctx context.Context, filter ProgrammeListFilter) ([]models.Programme, error) {
	query := r.db.WithContext(ctx).Model(&models.Programme{})
	if filter.DepartmentID != nil {
		query = query.Where("department_id = ?", *filter.DepartmentID)
	}

	var items []models.Programme
	err := query.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *AcademicRepository) GetProgramme(ctx context.Context, programmeID uint) (*models.Programme, error) {
	var item models.Programme
	err := r.db.WithContext(ctx).First(&item, programmeID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *AcademicRepository) CreateProgramme(ctx context.Context, programme *models.Programme) error {
	return r.db.WithContext(ctx).Create(programme).Error
}

func (r *AcademicRepository) UpdateProgramme(ctx context.Context, programme *models.Programme) error {
	return r.db.WithContext(ctx).Save(programme).Error
}

func (r *AcademicRepository) DeleteProgramme(ctx context.Context, programmeID uint) error {
	return r.db.WithContext(ctx).Delete(&models.Programme{}, programmeID).Error
}

func (r *AcademicRepository) ListCourses(ctx context.Context, filter CourseListFilter) ([]models.Course, error) {
	query := r.db.WithContext(ctx).Model(&models.Course{})
	if filter.DepartmentID != nil {
		query = query.Where("department_id = ?", *filter.DepartmentID)
	}
	if filter.ProgrammeID != nil {
		query = query.Where("programme_id = ?", *filter.ProgrammeID)
	}
	if filter.Semester != "" {
		query = query.Where("semester = ?", filter.Semester)
	}
	if filter.Level != "" {
		query = query.Where("level = ?", filter.Level)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	var items []models.Course
	err := query.Order("code ASC").Find(&items).Error
	return items, err
}

func (r *AcademicRepository) GetCourse(ctx context.Context, courseID uint) (*models.Course, error) {
	var item models.Course
	err := r.db.WithContext(ctx).First(&item, courseID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *AcademicRepository) CreateCourse(ctx context.Context, course *models.Course) error {
	return r.db.WithContext(ctx).Create(course).Error
}

func (r *AcademicRepository) UpdateCourse(ctx context.Context, course *models.Course) error {
	return r.db.WithContext(ctx).Save(course).Error
}

func (r *AcademicRepository) DeleteCourse(ctx context.Context, courseID uint) error {
	return r.db.WithContext(ctx).Delete(&models.Course{}, courseID).Error
}
