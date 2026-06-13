package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	Code        string    `gorm:"size:100;uniqueIndex;not null"`
	Description string    `gorm:"type:text"`
	IsSystem    bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	RolePermissions []RolePermission `gorm:"constraint:OnDelete:CASCADE"`
	UserRoles       []UserRole       `gorm:"constraint:OnDelete:CASCADE"`
}

type Permission struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:120;not null"`
	Code        string    `gorm:"size:120;uniqueIndex;not null"`
	Module      string    `gorm:"size:80;index;not null"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	RolePermissions []RolePermission `gorm:"constraint:OnDelete:CASCADE"`
}

type RolePermission struct {
	ID           uint      `gorm:"primaryKey"`
	RoleID       uint      `gorm:"not null;uniqueIndex:idx_role_permission"`
	PermissionID uint      `gorm:"not null;uniqueIndex:idx_role_permission"`
	CreatedAt    time.Time

	Role       Role       `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE"`
	Permission Permission `gorm:"foreignKey:PermissionID;constraint:OnDelete:CASCADE"`
}

type UserRole struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"not null;uniqueIndex:idx_user_role_scope"`
	RoleID     uint      `gorm:"not null;uniqueIndex:idx_user_role_scope"`
	ScopeType  ScopeType `gorm:"size:30;not null;index"`
	ScopeID    *uint     `gorm:"index"`
	ScopeKey   string    `gorm:"size:64;not null;uniqueIndex:idx_user_role_scope"`
	IsPrimary  bool      `gorm:"not null;default:false"`
	AssignedBy *uint     `gorm:"index"`
	StartsAt   *time.Time `gorm:"index"`
	EndsAt     *time.Time `gorm:"index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	User           User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Role           Role  `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE"`
	AssignedByUser *User `gorm:"foreignKey:AssignedBy;constraint:OnDelete:SET NULL"`
}

func (u *UserRole) BeforeSave(_ *gorm.DB) error {
	if !u.ScopeType.Valid() {
		return errors.New("invalid scope_type")
	}

	if u.ScopeType == ScopeSchool {
		u.ScopeID = nil
	}

	scopeKey, err := BuildScopeKey(u.ScopeType, u.ScopeID)
	if err != nil {
		return err
	}

	u.ScopeKey = scopeKey

	if u.StartsAt != nil && u.EndsAt != nil && u.EndsAt.Before(*u.StartsAt) {
		return errors.New("ends_at must be after starts_at")
	}

	return nil
}

func (u UserRole) IsActiveAt(at time.Time) bool {
	if u.StartsAt != nil && at.Before(*u.StartsAt) {
		return false
	}

	if u.EndsAt != nil && at.After(*u.EndsAt) {
		return false
	}

	return true
}
