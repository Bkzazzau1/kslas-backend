package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Staff struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	StaffNumber  string     `json:"staff_number" gorm:"size:60;uniqueIndex;not null"`
	Title        string     `json:"title" gorm:"size:30"`
	FirstName    string     `json:"first_name" gorm:"size:100;not null"`
	LastName     string     `json:"last_name" gorm:"size:100;not null"`
	OtherNames   string     `json:"other_names" gorm:"size:120"`
	Email        string     `json:"email" gorm:"size:180;uniqueIndex;not null"`
	Phone        string     `json:"phone" gorm:"size:40"`
	DepartmentID *uuid.UUID `json:"department_id" gorm:"type:uuid;index"`
	Department   Department `json:"department" gorm:"foreignKey:DepartmentID"`
	PrimaryRole  string     `json:"primary_role" gorm:"size:40;index;default:lecturer"`
	Rank         string     `json:"rank" gorm:"size:80"`
	Specialty    string     `json:"specialty" gorm:"size:180"`
	EmploymentType string   `json:"employment_type" gorm:"size:60"`
	IsActive     bool       `json:"is_active" gorm:"default:true;index"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (s *Staff) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.PrimaryRole == "" {
		s.PrimaryRole = "lecturer"
	}
	return nil
}

type StaffRoleAssignment struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	StaffID      uuid.UUID  `json:"staff_id" gorm:"type:uuid;not null;index"`
	Staff        Staff      `json:"staff" gorm:"foreignKey:StaffID"`
	Role         string     `json:"role" gorm:"size:40;not null;index"`
	DepartmentID *uuid.UUID `json:"department_id" gorm:"type:uuid;index"`
	CourseID     *uuid.UUID `json:"course_id" gorm:"type:uuid;index"`
	Scope        string     `json:"scope" gorm:"size:60;default:department"`
	IsActive     bool       `json:"is_active" gorm:"default:true;index"`
	AssignedByID *uuid.UUID `json:"assigned_by_id" gorm:"type:uuid"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (r *StaffRoleAssignment) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Scope == "" {
		r.Scope = "department"
	}
	return nil
}
