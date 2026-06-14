package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserType string

const (
	UserTypeStudent UserType = "student"
	UserTypeStaff   UserType = "staff"
)

func (t UserType) Valid() bool {
	switch t {
	case UserTypeStudent, UserTypeStaff:
		return true
	default:
		return false
	}
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

func (s UserStatus) Valid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusSuspended:
		return true
	default:
		return false
	}
}

type ScopeType string

const (
	ScopeSchool     ScopeType = "school"
	ScopeFaculty    ScopeType = "faculty"
	ScopeDepartment ScopeType = "department"
	ScopeProgramme  ScopeType = "programme"
	ScopeCourse     ScopeType = "course"
	ScopeCohort     ScopeType = "cohort"
	ScopeExam       ScopeType = "exam"
	ScopeRoom       ScopeType = "room"
)

func (s ScopeType) Valid() bool {
	switch s {
	case ScopeSchool, ScopeFaculty, ScopeDepartment, ScopeProgramme, ScopeCourse, ScopeCohort, ScopeExam, ScopeRoom:
		return true
	default:
		return false
	}
}

func BuildScopeKey(scopeType ScopeType, scopeID *uint) (string, error) {
	if !scopeType.Valid() {
		return "", fmt.Errorf("invalid scope type %q", scopeType)
	}

	if scopeType == ScopeSchool {
		return "school:*", nil
	}

	if scopeID == nil || *scopeID == 0 {
		return "", errors.New("scope id is required for non-school scope")
	}

	return fmt.Sprintf("%s:%d", scopeType, *scopeID), nil
}

type User struct {
	ID           uint           `gorm:"primaryKey"`
	UUID         string         `gorm:"size:36;uniqueIndex;not null"`
	FirstName    string         `gorm:"size:100;not null"`
	LastName     string         `gorm:"size:100;not null"`
	MiddleName   string         `gorm:"size:100"`
	Email        string         `gorm:"size:120;uniqueIndex"`
	Phone        string         `gorm:"size:20;uniqueIndex"`
	PasswordHash string         `gorm:"size:255;not null"`
	Gender       string         `gorm:"size:20"`
	Status       UserStatus     `gorm:"size:20;not null;default:active"`
	MatricNo     string         `gorm:"size:50;uniqueIndex"`
	StaffID      string         `gorm:"size:50;uniqueIndex"`
	UserType     UserType       `gorm:"size:20;not null"`
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	UserRoles []UserRole `gorm:"constraint:OnDelete:CASCADE"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(u.UUID) == "" {
		u.UUID = uuid.NewString()
	}

	return nil
}

func (u *User) BeforeSave(_ *gorm.DB) error {
	u.FirstName = strings.TrimSpace(u.FirstName)
	u.LastName = strings.TrimSpace(u.LastName)
	u.MiddleName = strings.TrimSpace(u.MiddleName)
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Phone = strings.TrimSpace(u.Phone)
	u.MatricNo = strings.TrimSpace(u.MatricNo)
	u.StaffID = strings.TrimSpace(u.StaffID)

	if u.FirstName == "" || u.LastName == "" {
		return errors.New("first_name and last_name are required")
	}

	if !u.UserType.Valid() {
		return fmt.Errorf("invalid user type %q", u.UserType)
	}

	if u.Status == "" {
		u.Status = UserStatusActive
	}

	if !u.Status.Valid() {
		return fmt.Errorf("invalid user status %q", u.Status)
	}

	return nil
}
