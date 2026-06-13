package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Faculty struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:120;not null"`
	Code      string `gorm:"size:30;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Departments []Department
}

func (f *Faculty) BeforeSave(_ *gorm.DB) error {
	f.Name = strings.TrimSpace(f.Name)
	f.Code = strings.ToUpper(strings.TrimSpace(f.Code))
	if f.Name == "" || f.Code == "" {
		return errors.New("faculty name and code are required")
	}
	return nil
}

type Department struct {
	ID        uint   `gorm:"primaryKey"`
	FacultyID uint   `gorm:"not null;index"`
	Name      string `gorm:"size:120;not null"`
	Code      string `gorm:"size:30;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Faculty    Faculty     `gorm:"foreignKey:FacultyID;constraint:OnDelete:CASCADE"`
	Programmes []Programme `gorm:"constraint:OnDelete:CASCADE"`
	Courses    []Course    `gorm:"constraint:OnDelete:CASCADE"`
}

func (d *Department) BeforeSave(_ *gorm.DB) error {
	d.Name = strings.TrimSpace(d.Name)
	d.Code = strings.ToUpper(strings.TrimSpace(d.Code))
	if d.FacultyID == 0 || d.Name == "" || d.Code == "" {
		return errors.New("department faculty_id, name, and code are required")
	}
	return nil
}

type Programme struct {
	ID           uint   `gorm:"primaryKey"`
	DepartmentID uint   `gorm:"not null;index"`
	Name         string `gorm:"size:120;not null"`
	Code         string `gorm:"size:30;uniqueIndex;not null"`
	LevelType    string `gorm:"size:30;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Department Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:CASCADE"`
	Courses    []Course   `gorm:"constraint:OnDelete:SET NULL"`
}

func (p *Programme) BeforeSave(_ *gorm.DB) error {
	p.Name = strings.TrimSpace(p.Name)
	p.Code = strings.ToUpper(strings.TrimSpace(p.Code))
	p.LevelType = strings.ToLower(strings.TrimSpace(p.LevelType))
	if p.DepartmentID == 0 || p.Name == "" || p.Code == "" || p.LevelType == "" {
		return errors.New("programme department_id, name, code, and level_type are required")
	}
	return nil
}

type Course struct {
	ID           uint   `gorm:"primaryKey"`
	DepartmentID uint   `gorm:"not null;index"`
	ProgrammeID  *uint  `gorm:"index"`
	Title        string `gorm:"size:150;not null"`
	Code         string `gorm:"size:30;uniqueIndex;not null"`
	Unit         uint   `gorm:"not null"`
	Semester     string `gorm:"size:20;not null"`
	Level        string `gorm:"size:20"`
	IsActive     bool   `gorm:"not null;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Department Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:CASCADE"`
	Programme  *Programme `gorm:"foreignKey:ProgrammeID;constraint:OnDelete:SET NULL"`
}

func (c *Course) BeforeSave(_ *gorm.DB) error {
	c.Title = strings.TrimSpace(c.Title)
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
	c.Semester = strings.TrimSpace(c.Semester)
	c.Level = strings.TrimSpace(c.Level)
	if c.DepartmentID == 0 || c.Title == "" || c.Code == "" || c.Semester == "" || c.Level == "" || c.Unit == 0 {
		return errors.New("course department_id, title, code, unit, semester, and level are required")
	}
	return nil
}

type CourseLecturerAssignment struct {
	ID         uint `gorm:"primaryKey"`
	CourseID   uint `gorm:"not null;index;uniqueIndex:idx_course_lecturer"`
	LecturerID uint `gorm:"not null;index;uniqueIndex:idx_course_lecturer"`
	AssignedBy uint `gorm:"not null;index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Course   Course `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
	Lecturer User   `gorm:"foreignKey:LecturerID;constraint:OnDelete:CASCADE"`
}

func (a *CourseLecturerAssignment) BeforeSave(_ *gorm.DB) error {
	if a.CourseID == 0 || a.LecturerID == 0 || a.AssignedBy == 0 {
		return errors.New("course_id, lecturer_id, and assigned_by are required")
	}
	return nil
}

type StudentAcademicProfile struct {
	ID                 uint   `gorm:"primaryKey"`
	StudentID          uint   `gorm:"not null;uniqueIndex"`
	DepartmentID       uint   `gorm:"not null;index"`
	ProgrammeID        uint   `gorm:"not null;index"`
	Level              string `gorm:"size:20;not null;index"`
	Semester           string `gorm:"size:20;not null;index"`
	AcademicSession    string `gorm:"size:20;not null;index"`
	CountryOfResidence string `gorm:"size:80"`
	IsInternational    bool   `gorm:"not null;default:false;index"`
	RegisteredBy       uint   `gorm:"not null;index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Student    User       `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Department Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:CASCADE"`
	Programme  Programme  `gorm:"foreignKey:ProgrammeID;constraint:OnDelete:CASCADE"`
}

func (p *StudentAcademicProfile) BeforeSave(_ *gorm.DB) error {
	p.Level = strings.TrimSpace(p.Level)
	p.Semester = strings.TrimSpace(p.Semester)
	p.AcademicSession = strings.TrimSpace(p.AcademicSession)
	p.CountryOfResidence = strings.TrimSpace(p.CountryOfResidence)
	if p.StudentID == 0 || p.DepartmentID == 0 || p.ProgrammeID == 0 || p.Level == "" || p.Semester == "" || p.AcademicSession == "" || p.RegisteredBy == 0 {
		return errors.New("student academic profile requires student, department, programme, level, semester, academic_session, and registered_by")
	}
	return nil
}

type CourseRegistrationStatus string

const (
	CourseRegistrationPending  CourseRegistrationStatus = "pending"
	CourseRegistrationApproved CourseRegistrationStatus = "approved"
)

func (s CourseRegistrationStatus) Valid() bool {
	switch s {
	case CourseRegistrationPending, CourseRegistrationApproved:
		return true
	default:
		return false
	}
}

type CourseRegistration struct {
	ID              uint                     `gorm:"primaryKey"`
	StudentID       uint                     `gorm:"not null;index;uniqueIndex:idx_student_course_session"`
	CourseID        uint                     `gorm:"not null;index;uniqueIndex:idx_student_course_session"`
	AcademicSession string                   `gorm:"size:20;not null;uniqueIndex:idx_student_course_session"`
	Semester        string                   `gorm:"size:20;not null;index"`
	Level           string                   `gorm:"size:20;not null;index"`
	Status          CourseRegistrationStatus `gorm:"size:20;not null;default:pending;index"`
	RegisteredAt    time.Time                `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Student User   `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Course  Course `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
}

func (r *CourseRegistration) BeforeSave(_ *gorm.DB) error {
	r.Level = strings.TrimSpace(r.Level)
	r.Semester = strings.TrimSpace(r.Semester)
	r.AcademicSession = strings.TrimSpace(r.AcademicSession)
	if r.StudentID == 0 || r.CourseID == 0 || r.Level == "" || r.Semester == "" || r.AcademicSession == "" {
		return errors.New("course registration requires student, course, level, semester, and academic_session")
	}
	if r.Status == "" {
		r.Status = CourseRegistrationPending
	}
	if !r.Status.Valid() {
		return errors.New("invalid course registration status")
	}
	if r.RegisteredAt.IsZero() {
		r.RegisteredAt = time.Now().UTC()
	}
	return nil
}
